package linode

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/delaneyj/toolbelt/wisshes"
	"github.com/goccy/go-json"
	"github.com/linode/linodego"
)

const ctxLinodeKeyDomains = ctxLinodeKeyPrefix + "domains"

func CtxLinodeDomains(ctx context.Context) []linodego.Domain {
	return ctx.Value(ctxLinodeKeyDomains).([]linodego.Domain)
}

func CtxWithLinodeDomains(ctx context.Context, domains []linodego.Domain) context.Context {
	return context.WithValue(ctx, ctxLinodeKeyDomains, domains)
}

func Domains() wisshes.Step {
	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		name := "domains"

		linodeClient := CtxLinodeClient(ctx)

		domains, err := linodeClient.ListDomains(ctx, nil)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, err
		}

		b, err := json.MarshalIndent(domains, "", "  ")
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, err
		}

		fp := filepath.Join(wisshes.ArtifactsDir(), name+".json")
		if previous, err := os.ReadFile(fp); err == nil {
			if bytes.Equal(previous, b) {
				ctx = CtxWithLinodeDomains(ctx, domains)
				return ctx, name, wisshes.StepStatusUnchanged, nil
			}
		}
		if err := os.WriteFile(fp, b, 0644); err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("write checksum: %w", err)
		}

		ctx = CtxWithLinodeDomains(ctx, domains)
		return ctx, name, wisshes.StepStatusUnchanged, nil
	}
}
