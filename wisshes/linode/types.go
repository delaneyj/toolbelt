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

const ctxLinodeKeyInstanceTypes wisshes.CtxKey = "linode-instance-types"

func CtxLinodeInstanceTypes(ctx context.Context) []linodego.LinodeType {
	return ctx.Value(ctxLinodeKeyInstanceTypes).([]linodego.LinodeType)
}

func CtxWithLinodeInstanceTypes(ctx context.Context, instanceTypes []linodego.LinodeType) context.Context {
	return context.WithValue(ctx, ctxLinodeKeyInstanceTypes, instanceTypes)
}

func InstanceTypes() wisshes.Step {
	return func(ctx context.Context) (context.Context, string, wisshes.StepStatus, error) {
		name := "linode_instance_types"

		linodeClient := CtxLinodeClient(ctx)
		if linodeClient == nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("linode client not found")
		}

		linodeTypes, err := linodeClient.ListTypes(ctx, nil)
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("list types: %w", err)
		}

		b, err := json.MarshalIndent(linodeTypes, "", "  ")
		if err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("marshal: %w", err)
		}

		fp := filepath.Join(wisshes.ArtifactsDir(), name+".json")
		if previous, err := os.ReadFile(fp); err == nil {
			if bytes.Equal(previous, b) {
				ctx = CtxWithLinodeInstanceTypes(ctx, linodeTypes)
				return ctx, name, wisshes.StepStatusUnchanged, nil
			}
		}

		if err := os.WriteFile(fp, b, 0644); err != nil {
			return ctx, name, wisshes.StepStatusFailed, fmt.Errorf("write checksum: %w", err)
		}

		ctx = CtxWithLinodeInstanceTypes(ctx, linodeTypes)
		return ctx, name, wisshes.StepStatusChanged, nil
	}
}
