package zombiezen

import (
	"github.com/delaneyj/toolbelt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func generateUtil(req *plugin.GenerateRequest) (files []*plugin.File, err error) {
	queryContents := GenerateUtil(&GenerateUtilContext{
		PackageName: toolbelt.ToCasedString(req.Settings.Codegen.Out),
	})

	f := &plugin.File{
		Name:     "util.go",
		Contents: []byte(queryContents),
	}

	return []*plugin.File{f}, nil
}

type GenerateUtilContext struct {
	PackageName toolbelt.CasedString
}
