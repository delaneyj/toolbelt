package wisshes

import (
	"context"
)

type StepStatus string

const (
	StepStatusUnchanged StepStatus = "unchanged"
	StepStatusChanged   StepStatus = "changed"
	StepStatusFailed    StepStatus = "failed"
)

type Step func(ctx context.Context) (revisedCtx context.Context, name string, status StepStatus, err error)
