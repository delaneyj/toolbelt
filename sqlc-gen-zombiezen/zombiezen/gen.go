package zombiezen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/delaneyj/toolbelt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"golang.org/x/tools/imports"
)

type Options struct {
	DisableCRUD           bool `json:"disable_crud"`
	DisableTimeConversion bool `json:"disable_time_conversion"`
}

type generationConfig struct {
	packageName toolbelt.CasedString
}

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	options, err := parseOptions(req)
	if err != nil {
		return nil, fmt.Errorf("parsing options: %w", err)
	}

	cfg, err := buildGenerationConfig(req, options)
	if err != nil {
		return nil, fmt.Errorf("configuring generation: %w", err)
	}

	if err := cleanupOutputDirectory(req); err != nil {
		return nil, fmt.Errorf("cleaning output directory: %w", err)
	}

	res := &plugin.GenerateResponse{}

	queryFiles, err := generateQueries(req, options, cfg.packageName)
	if err != nil {
		return nil, fmt.Errorf("generating queries: %w", err)
	}
	res.Files = append(res.Files, queryFiles...)

	if !options.DisableCRUD {
		crudFiles, err := generateCRUD(req, options, cfg.packageName)
		if err != nil {
			return nil, fmt.Errorf("generating crud: %w", err)
		}
		res.Files = append(res.Files, crudFiles...)
	}

	if err := formatGeneratedFiles(res.Files); err != nil {
		return nil, fmt.Errorf("formatting generated files: %w", err)
	}

	return res, nil
}

func buildGenerationConfig(req *plugin.GenerateRequest, opts *Options) (generationConfig, error) {
	settings := req.GetSettings()
	var outPath, pluginName string
	if settings != nil && settings.GetCodegen() != nil {
		outPath = strings.TrimSpace(settings.GetCodegen().GetOut())
		pluginName = strings.TrimSpace(settings.GetCodegen().GetPlugin())
	}

	packageCandidate := ""
	if outPath != "" {
		packageCandidate = lastPathComponent(outPath)
	}
	if packageCandidate == "" && pluginName != "" {
		packageCandidate = pluginName
	}
	if packageCandidate == "" {
		packageCandidate = "zz"
	}

	return generationConfig{
		packageName: toolbelt.ToCasedString(packageCandidate),
	}, nil
}

func cleanupOutputDirectory(req *plugin.GenerateRequest) error {
	settings := req.GetSettings()
	if settings == nil || settings.GetCodegen() == nil {
		return fmt.Errorf("missing codegen settings")
	}

	baseOut := strings.TrimSpace(settings.GetCodegen().GetOut())
	if baseOut == "" {
		return nil
	}

	target := filepath.Clean(baseOut)
	if strings.HasPrefix(target, "..") {
		return fmt.Errorf("refusing to clean output outside workspace: %q", target)
	}

	if target == "." || target == "" {
		return fmt.Errorf("refusing to remove unsafe output directory %q", target)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("determining absolute output path: %w", err)
	}
	if filepath.Dir(absTarget) == absTarget {
		return fmt.Errorf("refusing to remove root directory %q", absTarget)
	}

	if err := os.RemoveAll(absTarget); err != nil {
		return fmt.Errorf("removing output directory %s: %w", absTarget, err)
	}
	if err := os.MkdirAll(absTarget, 0o755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", absTarget, err)
	}
	return nil
}

func lastPathComponent(p string) string {
	trimmed := strings.TrimSpace(p)
	if trimmed == "" {
		return ""
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	base := path.Base(normalized)
	if base == "." || base == "/" {
		return ""
	}
	return base
}

func parseOptions(req *plugin.GenerateRequest) (*Options, error) {
	opts := &Options{}
	if len(req.PluginOptions) == 0 {
		return opts, nil
	}

	dec := json.NewDecoder(bytes.NewReader(req.PluginOptions))
	dec.DisallowUnknownFields()
	if err := dec.Decode(opts); err != nil {
		return nil, fmt.Errorf("unmarshalling options: %w", err)
	}

	return opts, nil
}

func formatGeneratedFiles(files []*plugin.File) error {
	opts := &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	}

	for _, f := range files {
		if filepath.Ext(f.Name) != ".go" {
			continue
		}

		formatted, err := imports.Process(f.Name, f.Contents, opts)
		if err == nil {
			f.Contents = formatted
			continue
		}

		formatted, err = format.Source(f.Contents)
		if err != nil {
			return fmt.Errorf("formatting %s: %w", f.Name, err)
		}
		f.Contents = formatted
	}

	return nil
}
