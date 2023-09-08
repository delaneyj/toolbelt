package linode

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/delaneyj/toolbelt/wisshes"
	"github.com/go-ping/ping"
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

func RemoveAllInstances(prefix string) wisshes.Step {
	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		name := "remove-all-instances-" + prefix

		linodeClient := CtxLinodeClient(ctx)
		if linodeClient == nil {
			return ctx, name, wisshes.StepStatusFailed, errors.New("linode client not found")
		}

		instances, err := linodeClient.ListInstances(ctx, nil)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list instances: %w", err)
		}
		instances = lo.Filter(instances, func(instance linodego.Instance, i int) bool {
			return strings.HasPrefix(instance.Label, prefix)
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

type PingMode string

const (
	PingModeShortest    PingMode = "shortest"
	PingModeDistributed PingMode = "distributed"
)

type DesiredInstancesArgs struct {
	RootPassword            string
	LabelPrefix             string
	InstancesPerRegionCount int
	CountrySelection        PingMode
	RegionCount             int
	RegionSelection         PingMode
	TargetMonthlyBudget     float32
}

func DesiredInstances(args DesiredInstancesArgs, spinupSteps ...wisshes.Step) wisshes.Step {
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

		regions := CtxLinodeRegion(ctx)
		if len(regions) == 0 {
			return ctx, name, wisshes.StepStatusFailed, errors.New("no regions found")
		}

		type regionPing struct {
			region linodego.Region
			ping   time.Duration
		}
		regionPings := make([]regionPing, len(regions))

		allInstances, err := linodeClient.ListInstances(ctx, nil)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list instances: %w", err)
		}
		allInstances = lo.Filter(allInstances, func(instance linodego.Instance, i int) bool {
			return strings.HasPrefix(instance.Label, args.LabelPrefix)
		})

		var changed int32

		totalInstanceCount := args.InstancesPerRegionCount * args.RegionCount
		if len(allInstances) != totalInstanceCount {
			wg := &sync.WaitGroup{}
			wg.Add(len(regions))
			for i, region := range regions {
				go func(i int, region linodego.Region) {
					defer wg.Done()
					regionCaps := sets.NewString(region.Capabilities...)
					if !regionCaps.HasAll("Linodes") {
						return
					}

					addrs := strings.Split(region.Resolvers.IPv4, ", ")
					for _, addr := range addrs {
						pinger, err := ping.NewPinger(addr)
						if err != nil {
							continue
						}
						pinger.Count = 1
						pinger.Interval = 10 * time.Millisecond
						pinger.Run()
						stats := pinger.Statistics()

						ping := stats.AvgRtt
						//log.Printf("Region %s: %s", region.ID, ping)

						regionPings[i].region = region
						regionPings[i].ping = ping
						break
					}
				}(i, region)
			}
			wg.Wait()

			slices.SortFunc(regionPings, func(a, b regionPing) int {
				if a.ping == 0 {
					return 1
				}
				return int(a.ping - b.ping)
			})

			if args.RegionCount > len(regionPings) {
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("region count %d is greater than available regions %d", args.RegionCount, len(regionPings))
			}

			chosenRegions := make([]linodego.Region, args.RegionCount)
			switch args.RegionSelection {
			case PingModeShortest:
				for i := range chosenRegions {
					chosenRegions[i] = regionPings[i].region
				}
			case PingModeDistributed:
				regionCountPerPingF := float64(len(regionPings)) / float64(args.RegionCount)
				for i := 0; i < args.RegionCount-1; i++ {
					chosenIdx := int(math.Round(float64(i) * regionCountPerPingF))
					chosenRegions[i] = regionPings[chosenIdx].region
					log.Printf("Chose region %s with ping %s", chosenRegions[i].ID, regionPings[chosenIdx].ping)
				}
				lastIdx := len(regionPings) - 1
				lastRegionPing := regionPings[lastIdx]
				chosenRegions[args.RegionCount-1] = lastRegionPing.region
				log.Printf("Chose region %s with ping %s", lastRegionPing.region.ID, lastRegionPing.ping)
			default:
				return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("unknown region selection mode %s", args.RegionSelection)
			}

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

		inv, err := instanceToInventory(args.RootPassword, allInstances...)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("inventory: %w", err)
		}

		ctx = wisshes.CtxWithInventory(ctx, inv)
		ctx = CtxWithLinodeInstances(ctx, allInstances)

		status := wisshes.StepStatusUnchanged
		if changed > 0 {
			status = wisshes.StepStatusChanged
		}
		ctx = wisshes.CtxWithPreviousStep(ctx, status)

		lastStatus, err := inv.Run(ctx, spinupSteps...)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("run: %w", err)
		}

		wait := 10 * time.Second
		log.Printf("Setup %d instances, waiting %s for them to be ready", len(allInstances), wait)
		time.Sleep(wait)

		return ctx, name, lastStatus, nil
	}
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

		wg := &sync.WaitGroup{}
		wg.Add(len(instances))
		errs := make([]error, len(instances))

		for i, instance := range instances {
			go func(instance linodego.Instance, i int) {
				defer wg.Done()
				inv, err := instanceToInventory(rootPassword, instance)
				if err != nil {
					errs[i] = fmt.Errorf("inventory: %w", err)
					return
				}

				steps, err := cb(ctx, instance)
				if err != nil {
					errs[i] = fmt.Errorf("callback: %w", err)
					return
				}

				if len(steps) == 0 {
					return
				}

				lastStatus, err := inv.Run(ctx, steps...)
				if err != nil {
					errs[i] = fmt.Errorf("run: %w", err)
					return
				}

				if lastStatus == wisshes.StepStatusFailed {
					errs[i] = fmt.Errorf("last status failed")
					return
				}

				if lastStatus == wisshes.StepStatusChanged {
					status = wisshes.StepStatusChanged
				}
			}(instance, i)
		}
		wg.Wait()

		if err := errors.Join(errs...); err != nil {
			return ctx, name, wisshes.StepStatusFailed, err
		}

		return ctx, name, status, nil
	}
}

func InstanceToIP4(instance linodego.Instance) string {
	return instance.IPv4[0].String()
}
