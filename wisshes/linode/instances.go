package linode

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/delaneyj/toolbelt/wisshes"
	"github.com/goccy/go-json"
	"github.com/linode/linodego"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/sets"
)

const ctxLinodeKeyInstances = ctxLinodeKeyPrefix + "instances"

func CtxLinodeInstances(ctx context.Context) []linodego.Instance {
	return ctx.Value(ctxLinodeKeyInstances).([]linodego.Instance)
}

func CtxWithLinodeInstances(ctx context.Context, instances []linodego.Instance) context.Context {
	return context.WithValue(ctx, ctxLinodeKeyInstances, instances)
}

func CurrentInstances(includes ...linodego.InstanceStatus) wisshes.Step {
	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		includedStrs := make([]string, len(includes))
		for i, include := range includes {
			includedStrs[i] = string(include)
		}
		name := "instances-" + strings.Join(includedStrs, "-")

		linodeClient := CtxLinodeClient(ctx)
		if linodeClient == nil {
			return ctx, name, wisshes.StepStatusFailed, errors.New("linode client not found")
		}

		instances, err := linodeClient.ListInstances(ctx, nil)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list instances: %w", err)
		}
		toInclude := sets.New(includes...)
		instances = lo.Filter(instances, func(instance linodego.Instance, i int) bool {
			return toInclude.Has(instance.Status)
		})

		b, err := json.MarshalIndent(instances, "", "  ")
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("marshal: %w", err)
		}

		fp := filepath.Join(wisshes.ArtifactsDir(), name+".json")
		if previous, err := os.ReadFile(fp); err == nil {
			if bytes.Equal(previous, b) {
				ctx = CtxWithLinodeInstances(ctx, instances)
				return ctx, name, wisshes.StepStatusUnchanged, nil
			}
		}
		if err := os.WriteFile(fp, b, 0644); err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("write checksum: %w", err)
		}

		ctx = CtxWithLinodeInstances(ctx, instances)
		return ctx, name, wisshes.StepStatusUnchanged, nil
	}
}

func RemoveAllInstances(prefixes ...string) wisshes.Step {
	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		name := "remove-all-instances-" + strings.Join(prefixes, "-")

		linodeClient := CtxLinodeClient(ctx)
		if linodeClient == nil {
			return ctx, name, wisshes.StepStatusFailed, errors.New("linode client not found")
		}

		instances, err := linodeClient.ListInstances(ctx, nil)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list instances: %w", err)
		}
		instances = lo.Filter(instances, func(instance linodego.Instance, i int) bool {
			for _, prefix := range prefixes {
				if strings.HasPrefix(instance.Label, prefix) {
					return true
				}
			}
			return false
		})

		if len(instances) == 0 {
			return ctx, name, wisshes.StepStatusUnchanged, nil
		}

		// Delete all instances
		wgDeletes := &sync.WaitGroup{}
		wgDeletes.Add(len(instances))
		for _, instance := range instances {
			go func(instance linodego.Instance) {
				defer wgDeletes.Done()
				linodeClient.DeleteInstance(ctx, instance.ID)
			}(instance)
		}
		wgDeletes.Wait()

		return ctx, name, wisshes.StepStatusChanged, nil
	}
}

type DesiredInstancesArgs struct {
	RootPrefix              string
	LabelPrefix             string
	RootPassword            string
	Regions                 []string
	InstancesPerRegionCount int
	TargetMonthlyBudget     float32
	Tags                    []string
}

func DesiredInstances(instancesArgs []DesiredInstancesArgs, spinupSteps ...wisshes.Step) wisshes.Step {
	allDesiredInstances := []linodego.Instance{}

	steps := lo.Map(instancesArgs, func(args DesiredInstancesArgs, desiredArgsIdx int) wisshes.Step {
		return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
			name := "desired-instances-" + args.LabelPrefix

			linodeClient := CtxLinodeClient(ctx)
			if linodeClient == nil {
				return ctx, name, wisshes.StepStatusFailed, errors.New("linode client not found")
			}

			instanceTypes := CtxLinodeInstanceTypes(ctx)
			if len(instanceTypes) == 0 {
				return ctx, name, wisshes.StepStatusFailed, errors.New("no instances found")
			}

			allAvailableRegions := CtxLinodeRegion(ctx)
			if len(allAvailableRegions) == 0 {
				return ctx, name, wisshes.StepStatusFailed, errors.New("no regions found")
			}

			regionCount := len(args.Regions)
			chosenRegions := make([]linodego.Region, 0, regionCount)
			for _, region := range allAvailableRegions {
				for _, desiredRegion := range args.Regions {
					if region.ID == desiredRegion {
						chosenRegions = append(chosenRegions, region)
					}
				}
			}

			if len(chosenRegions) != regionCount {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("could not find all regions")
			}

			allInstances, err := linodeClient.ListInstances(ctx, nil)
			if err != nil {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list instances: %w", err)
			}
			allInstances = lo.Filter(allInstances, func(instance linodego.Instance, i int) bool {
				return strings.HasPrefix(instance.Label, args.LabelPrefix)
			})

			var changed int32

			totalInstanceCount := args.InstancesPerRegionCount * regionCount
			if len(allInstances) != totalInstanceCount {

				perInstanceBudget := args.TargetMonthlyBudget / float32(totalInstanceCount)

				instanceTypesInBudget := lo.Filter(instanceTypes, func(instanceType linodego.LinodeType, i int) bool {
					return instanceType.Price.Monthly <= perInstanceBudget
				})
				if len(instanceTypesInBudget) == 0 {
					return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("no instance types found within budget %f", perInstanceBudget)
				}

				slices.SortFunc(instanceTypesInBudget, func(a, b linodego.LinodeType) int {
					return cmp.Compare(b.Price.Monthly, a.Price.Monthly)
				})

				instanceTypeToUse := instanceTypesInBudget[0]
				allInstances, err = linodeClient.ListInstances(ctx, nil)
				if err != nil {
					return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list instances: %w", err)
				}

				totalMonthlyCost := instanceTypeToUse.Price.Monthly * float32(totalInstanceCount)
				log.Printf(
					"Using a total of %d %s across %d regions with total monthly cost %f",
					totalInstanceCount,
					instanceTypeToUse.Label,
					len(chosenRegions),
					totalMonthlyCost,
				)

				existingInstances := lo.Filter(allInstances, func(instance linodego.Instance, i int) bool {
					hasPrefix := strings.HasPrefix(instance.Label, args.LabelPrefix)
					rightType := instance.Type == instanceTypeToUse.ID
					withinRightRegion := true
					for _, region := range chosenRegions {
						if instance.Region == region.ID {
							withinRightRegion = true
							break
						}
					}
					return hasPrefix && rightType && withinRightRegion
				})
				existingInstancesByRegionId := lo.GroupBy(existingInstances, func(instance linodego.Instance) string {
					return instance.Region
				})

				images, err := linodeClient.ListImages(ctx, nil)
				if err != nil {
					return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list images: %w", err)
				}

				var latestImage linodego.Image
				for _, image := range images {
					if strings.Contains(image.ID, "ubuntu") {
						if latestImage.ID == "" || latestImage.ID < image.ID {
							latestImage = image
						}
					}
				}
				log.Printf("Using image %s", latestImage.Label)

				errCh := make(chan error, totalInstanceCount)
				wgRegions := &sync.WaitGroup{}
				wgRegions.Add(len(chosenRegions))

				for _, region := range chosenRegions {
					go func(region linodego.Region) {
						defer wgRegions.Done()
						instances := existingInstancesByRegionId[region.ID]

						currentInstanceCount := len(instances)
						switch {
						case currentInstanceCount > args.InstancesPerRegionCount:
							// delete instances
							instancesToDelete := instances[args.InstancesPerRegionCount:]
							for _, instance := range instancesToDelete {
								atomic.AddInt32(&changed, 1)
								if err := linodeClient.DeleteInstance(ctx, instance.ID); err != nil {
									errCh <- fmt.Errorf("delete instance %s: %w", instance.Label, err)
									return
								}
							}

						case currentInstanceCount < args.InstancesPerRegionCount:
							delta := args.InstancesPerRegionCount - currentInstanceCount

							newInstances := make([]linodego.Instance, 0, delta)
							// create instances
							for i := 0; i < delta; i++ {
								label := fmt.Sprintf("%s-%s-%d", args.LabelPrefix, region.ID, currentInstanceCount+i+1)
								instance, err := linodeClient.CreateInstance(ctx, linodego.InstanceCreateOptions{
									Region:   region.ID,
									Type:     instanceTypeToUse.ID,
									Label:    label,
									Group:    args.LabelPrefix,
									RootPass: args.RootPassword,
									Image:    latestImage.ID,
									Tags:     args.Tags,
								})
								if err != nil {
									errCh <- fmt.Errorf("create instance: %w", err)
									return
								}
								allInstances = append(allInstances, *instance)
								newInstances = append(newInstances, *instance)
							}

							wgInstances := &sync.WaitGroup{}
							wgInstances.Add(len(newInstances))
							for _, instance := range newInstances {
								go func(instance linodego.Instance) {
									defer wgInstances.Done()
									for {
										// log.Printf("Instance '%s' is %s", instance.Label, instance.Status)
										possibleInstance, err := linodeClient.WaitForInstanceStatus(ctx, instance.ID, linodego.InstanceRunning, 5*60)
										if err == nil && possibleInstance != nil && possibleInstance.Status == linodego.InstanceRunning {
											log.Printf("Instance '%s' is %s", instance.Label, possibleInstance.Status)
											break
										}
									}
									atomic.AddInt32(&changed, 1)
								}(instance)
							}
							wgInstances.Wait()
							// do nothing
							return
						}
					}(region)
				}
				wgRegions.Wait()
				close(errCh)

				if len(errCh) > 0 {
					erss := make([]error, 0, len(errCh))
					for err := range errCh {
						erss = append(erss, err)
					}
					return ctx, name, wisshes.StepStatusFailed, errors.Join(erss...)
				}
			}

			allDesiredInstances = append(allDesiredInstances, allInstances...)

			inv, err := instanceToInventory(args.RootPassword, allDesiredInstances...)
			if err != nil {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("inventory: %w", err)
			}

			if desiredArgsIdx == len(instancesArgs)-1 {
				ctx = wisshes.CtxWithInventory(ctx, inv)
				ctx = CtxWithLinodeInstances(ctx, allDesiredInstances)
			}

			status := wisshes.StepStatusUnchanged
			if changed > 0 {
				status = wisshes.StepStatusChanged
			}
			ctx = wisshes.CtxWithPreviousStep(ctx, status)

			lastStatus, err := inv.Run(ctx, spinupSteps...)
			if err != nil {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("run: %w", err)
			}

			wait := 5 * time.Second
			if changed > 0 {
				wait *= 2
			}

			log.Printf("Setup %d instances, waiting %s for them to be ready", len(allInstances), wait)
			time.Sleep(wait)

			return ctx, name, lastStatus, nil
		}
	})

	return wisshes.RunAll(steps...)
}

func instanceToInventory(rootPassword string, instances ...linodego.Instance) (*wisshes.Inventory, error) {
	namesAndIPS := []string{}
	for _, instance := range instances {
		name := instance.Label
		ip := instance.IPv4[0].String()
		namesAndIPS = append(namesAndIPS, name, ip)
	}

	inv, err := wisshes.NewInventory(rootPassword, namesAndIPS...)
	if err != nil {
		return nil, fmt.Errorf("inventory: %w", err)
	}

	return inv, nil
}

func ForEachInstance(
	rootPassword string,
	cb func(ctx context.Context, instance linodego.Instance) ([]wisshes.Step, error),
) wisshes.Step {
	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		name := "for-each-instance"

		instances := CtxLinodeInstances(ctx)
		if len(instances) == 0 {
			return ctx, name, wisshes.StepStatusFailed, errors.New("no instances found")
		}

		status := wisshes.StepStatusUnchanged
		ctx = wisshes.CtxWithPreviousStep(ctx, status)

		for _, instance := range instances {
			inv, err := instanceToInventory(rootPassword, instance)
			if err != nil {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("inventory: %w", err)
			}

			steps, err := cb(ctx, instance)
			if err != nil {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("cb: %w", err)
			}

			lastStatus, err := inv.Run(ctx, steps...)
			if err != nil {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("run: %w", err)
			}

			if lastStatus == wisshes.StepStatusFailed {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("run: %w", err)
			}

			if lastStatus == wisshes.StepStatusChanged {
				status = wisshes.StepStatusChanged
			}

		}

		return ctx, name, status, nil
	}
}

func InstanceToIP4(instance linodego.Instance) string {
	return instance.IPv4[0].String()
}
