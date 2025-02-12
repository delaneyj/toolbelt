package wisshes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

func RunAll(steps ...Step) Step {
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		names := make([]string, len(steps))
		statuses := make([]StepStatus, len(steps))
		errs := make([]error, len(steps))

		for i, step := range steps {
			var (
				name   string
				status StepStatus
				err    error
			)
			ctx, name, status, err = step(ctx)
			if ctx == nil {
				panic("ctx is nil")
			}

			log.Printf("step %s: %s", name, status)

			names[i] = name
			statuses[i] = status
			errs[i] = err

			ctx = CtxWithPreviousStep(ctx, status)

			if err != nil {
				break
			}
		}

		name := fmt.Sprintf("run-all-%s", strings.Join(names, "-"))

		if err := errors.Join(errs...); err != nil {
			return ctx, name, StepStatusFailed, err
		}

		for _, status := range statuses {
			if status == StepStatusChanged {
				return ctx, name, StepStatusChanged, nil
			}
		}

		return ctx, name, StepStatusUnchanged, nil
	}
}

func IfPreviousChanged(steps ...Step) Step {
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		prevStep := CtxPreviousStep(ctx)
		if prevStep != StepStatusChanged {
			return ctx, "if-prev-changed", StepStatusUnchanged, nil
		}

		ctx, n, s, err := RunAll(steps...)(ctx)
		name := fmt.Sprintf("if-prev-changed-%s", n)
		return ctx, name, s, err
	}
}

func IfCond(cond bool, steps ...Step) Step {
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		if !cond {
			return nil, "if-cond", StepStatusUnchanged, nil
		}

		ctx, n, s, err := RunAll(steps...)(ctx)
		name := fmt.Sprintf("if-cond-%s", n)
		return ctx, name, s, err
	}
}

func Ternary(cond bool, ifTrue, ifFalse Step) Step {
	return func(ctx context.Context) (context.Context, string, StepStatus, error) {
		if cond {
			return ifTrue(ctx)
		}
		return ifFalse(ctx)
	}
}
