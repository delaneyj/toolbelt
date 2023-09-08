package wisshes

import (
	"context"

	"github.com/melbahja/goph"
)

type CtxKey string

const (
	ctxKeySSHClient    CtxKey = "ssh-client"
	ctxKeyPreviousStep CtxKey = "previous-step"
	ctxKeyInventory    CtxKey = "inventory"
)

func CtxSSHClient(ctx context.Context) *goph.Client {
	return ctx.Value(ctxKeySSHClient).(*goph.Client)
}

func CtxWithSSHClient(ctx context.Context, client *goph.Client) context.Context {
	return context.WithValue(ctx, ctxKeySSHClient, client)
}

func CtxPreviousStep(ctx context.Context) StepStatus {
	return ctx.Value(ctxKeyPreviousStep).(StepStatus)
}

func CtxWithPreviousStep(ctx context.Context, step StepStatus) context.Context {
	return context.WithValue(ctx, ctxKeyPreviousStep, step)
}

func CtxInventory(ctx context.Context) *Inventory {
	return ctx.Value(ctxKeyInventory).(*Inventory)
}

func CtxWithInventory(ctx context.Context, inv *Inventory) context.Context {
	return context.WithValue(ctx, ctxKeyInventory, inv)
}
