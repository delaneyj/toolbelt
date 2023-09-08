package wisshes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/melbahja/goph"
	"github.com/zeebo/xxh3"
)

func RunFn(client *goph.Client, format string, args ...any) (string, error) {
	cmd := fmt.Sprintf(format, args...)
	// log.Printf("Running %s", cmd)
	out, err := client.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("run %s: %w", cmd, err)
	}
	return string(out), nil
}

func Commands(cmds ...string) Step {
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		client := CtxSSHClient(ctx)
		name := fmt.Sprintf("commands-%d", xxh3.HashString(strings.Join(cmds, "\n")))

		results := make([]StepStatus, len(cmds))
		errs := make([]error, len(cmds))
		for i, cmd := range cmds {
			out, err := RunFn(client, cmd)
			if err != nil {
				log.Print(out)
				results[i] = StepStatusFailed
				errs[i] = fmt.Errorf("run: '%s' %w", cmd, err)
				break
			}

			results[i] = StepStatusChanged
		}

		if err := errors.Join(errs...); err != nil {
			return ctx, name, StepStatusFailed, err
		}

		for _, result := range results {
			if result == StepStatusChanged {
				return ctx, name, StepStatusChanged, nil
			}
		}

		return ctx, name, StepStatusUnchanged, nil
	}
}
