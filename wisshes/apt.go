package wisshes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
)

type AptitudeStatus string

const (
	AptitudeStatusUninstalled AptitudeStatus = "uninstalled"
	AptitudeStatusInstalled   AptitudeStatus = "installed"
)

func Aptitude(desiredStatus AptitudeStatus, packageNames ...string) Step {
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		client := CtxSSHClient(ctx)

		name := fmt.Sprintf("aptitude-%s-%s", desiredStatus, strcase.ToKebab(strings.Join(packageNames, "-")))

		if _, err := RunFn(client, "apt-get update"); err != nil {
			return ctx, name, StepStatusFailed, fmt.Errorf("apt update: %w", err)
		}
		// log.Print(out)

		results := make([]StepStatus, len(packageNames))
		errs := make([]error, len(packageNames))
		for i, packageName := range packageNames {
			query, err := RunFn(client, "dpkg -l %s", packageName)

			isNotInstalled := strings.Contains(query, "no packages found matching")
			shouldInstall := err != nil || (desiredStatus == AptitudeStatusInstalled && isNotInstalled)
			shouldUninstall := desiredStatus == AptitudeStatusUninstalled && !isNotInstalled

			if !shouldInstall && !shouldUninstall {
				results[i] = StepStatusUnchanged
				continue
			}

			switch desiredStatus {
			case AptitudeStatusInstalled:
				out, err := RunFn(client, "apt-get install -y %s", packageName)
				if err != nil {
					log.Print(out)
					results[i] = StepStatusFailed
					errs[i] = fmt.Errorf("apt-get install: %w", err)
					continue
				}
			case AptitudeStatusUninstalled:
				out, err := RunFn(client, "apt remove -y %s", packageName)
				if err != nil {
					log.Print(out)
					results[i] = StepStatusFailed
					errs[i] = fmt.Errorf("apt remove: %w", err)
					continue
				}
			default:
				panic("unreachable")
			}
		}

		if err := errors.Join(errs...); err != nil {
			return ctx, name, StepStatusFailed, err
		}

		for _, result := range results {
			if result == StepStatusChanged {
				return ctx, name, result, nil
			}
		}

		return ctx, name, StepStatusUnchanged, nil
	}
}
