package toolbelt

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type ErrGroupSharedCtx struct {
	eg  *errgroup.Group
	ctx context.Context
}

type CtxErrFunc func(ctx context.Context) error

func NewErrGroupSharedCtx(ctx context.Context, funcs ...CtxErrFunc) *ErrGroupSharedCtx {
	eg, ctx := errgroup.WithContext(ctx)

	egCtx := &ErrGroupSharedCtx{
		eg:  eg,
		ctx: ctx,
	}

	egCtx.Go(funcs...)

	return egCtx
}

func (egc *ErrGroupSharedCtx) Go(funcs ...CtxErrFunc) {
	for _, f := range funcs {
		fn := f
		egc.eg.Go(func() error {
			return fn(egc.ctx)
		})
	}
}

func (egc *ErrGroupSharedCtx) Wait() error {
	return egc.eg.Wait()
}

type ErrGroupSeparateCtx struct {
	eg *errgroup.Group
}

func NewErrGroupSeparateCtx() *ErrGroupSeparateCtx {
	eg := &errgroup.Group{}

	egCtx := &ErrGroupSeparateCtx{
		eg: eg,
	}

	return egCtx
}

func (egc *ErrGroupSeparateCtx) Go(ctx context.Context, funcs ...CtxErrFunc) {
	for _, f := range funcs {
		fn := f
		egc.eg.Go(func() error {
			return fn(ctx)
		})
	}
}

func (egc *ErrGroupSeparateCtx) Wait() error {
	return egc.eg.Wait()
}
