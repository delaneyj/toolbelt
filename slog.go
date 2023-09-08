package toolbelt

import (
	"context"
	"log/slog"
)

type CtxKey string

const CtxSlogKey CtxKey = "slog"

func CtxWithSlog(ctx context.Context, slog *slog.Logger) context.Context {
	return context.WithValue(ctx, CtxSlogKey, slog)
}

func CtxSlog(ctx context.Context) (logger *slog.Logger, ok bool) {
	logger, ok = ctx.Value(CtxSlogKey).(*slog.Logger)
	return logger, ok
}
