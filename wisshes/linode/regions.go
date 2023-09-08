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

const ctxLinodeKeyRegions = ctxLinodeKeyPrefix + "regions"

func CtxLinodeRegion(ctx context.Context) []linodego.Region {
	return ctx.Value(ctxLinodeKeyRegions).([]linodego.Region)
}

func CtxWithLinodeRegion(ctx context.Context, regions []linodego.Region) context.Context {
	return context.WithValue(ctx, ctxLinodeKeyRegions, regions)
}

func Regions() wisshes.Step {
	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		name := "regions"

		linodeClient := CtxLinodeClient(ctx)

		regions, err := linodeClient.ListRegions(ctx, nil)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, err
		}

		b, err := json.MarshalIndent(regions, "", "  ")
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, err
		}

		fp := filepath.Join(wisshes.ArtifactsDir(), name+".json")
		if previous, err := os.ReadFile(fp); err == nil {
			if bytes.Equal(previous, b) {
				ctx = CtxWithLinodeRegion(ctx, regions)
				return ctx, name, wisshes.StepStatusUnchanged, nil
			}
		}
		if err := os.WriteFile(fp, b, 0644); err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("write checksum: %w", err)
		}

		ctx = CtxWithLinodeRegion(ctx, regions)
		return ctx, name, wisshes.StepStatusUnchanged, nil
	}
}
