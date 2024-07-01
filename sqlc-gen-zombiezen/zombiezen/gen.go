package zombiezen

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/delaneyj/toolbelt"
	"github.com/delaneyj/toolbelt/sqlc-gen-zombiezen/pb/plugin"
	"github.com/samber/lo"
)

//go:embed templates/*.go.tpl
var templates embed.FS

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {

	tmpls, err := template.New("queries").Funcs(template.FuncMap{}).ParseFS(templates, "templates/*.go.tpl")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	queriesCtx := &GenerateQueriesContext{
		PackageName: toolbelt.ToCasedString(req.Settings.Codegen.Out),
	}
	queriesCtx.Queries = lo.Map(req.Queries, func(q *plugin.Query, qi int) GenerateQueryContext {
		queryCtx := GenerateQueryContext{
			Name: toolbelt.ToCasedString(q.Name),
			Params: lo.Map(q.Params, func(p *plugin.Parameter, pi int) GenerateField {
				param := GenerateField{
					Column:  int(p.Number),
					Name:    toolbelt.ToCasedString(p.Column.Name),
					SQLType: toolbelt.ToCasedString(toSQLType(p.Column)),
					GoType:  toolbelt.ToCasedString(toGoType(queriesCtx, p.Column)),
				}
				return param
			}),
		}
		queryCtx.HasParams = len(q.Params) > 0
		queryCtx.ParamsIsSingularField = len(q.Params) == 1

		if len(q.Columns) > 0 {
			queryCtx.HasResponse = true
			queryCtx.ResponseFields = lo.Map(q.Columns, func(c *plugin.Column, ci int) GenerateField {
				col := GenerateField{
					Column:  ci + 1,
					Name:    toolbelt.ToCasedString(c.Name),
					SQLType: toolbelt.ToCasedString(toSQLType(c)),
					GoType:  toolbelt.ToCasedString(toGoType(queriesCtx, c)),
				}
				return col
			})
			queryCtx.ResponseHasMultiple = q.Cmd == ":many"
			queryCtx.SQL = q.Text
			queryCtx.ResponseIsSingularField = len(q.Columns) == 1
		}
		return queryCtx
	})

	buf := &bytes.Buffer{}
	if err := tmpls.ExecuteTemplate(buf, "queries.go.tpl", queriesCtx); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &plugin.GenerateResponse{
		Files: []*plugin.File{
			{
				Name:     "queries.go",
				Contents: buf.Bytes(),
			},
		},
	}, nil
}

func toSQLType(c *plugin.Column) string {
	switch toolbelt.Lower(c.Type.Name) {
	case "text":
		return "text"
	case "integer":
		return "int64"
	case "datetime", "real":
		return "float"
	case "boolean":
		return "bool"
	default:
		panic(fmt.Sprintf("toSQLType unhandled type %s", c.Type.Name))
	}
}

func toGoType(queryCtx *GenerateQueriesContext, c *plugin.Column) string {
	typ := toolbelt.Lower(c.Type.Name)

	if strings.HasSuffix(c.Name, "ms") {
		queryCtx.NeedsTimePackage = true
		return "time.Duration"
	} else if strings.HasSuffix(c.Name, "at") || typ == "datetime" {
		queryCtx.NeedsTimePackage = true
		return "time.Time"
	} else {
		switch typ {
		case "text":
			return "string"
		case "integer":
			return "int64"
		case "real":
			return "float64"
		case "boolean":
			return "bool"
		default:
			panic(fmt.Sprintf("toGoType unhandled type %s for column %s ", c.Type.Name, c.Name))
		}
	}
}

type GenerateField struct {
	Column  int
	Name    toolbelt.CasedString
	SQLType toolbelt.CasedString
	GoType  toolbelt.CasedString
}

type GenerateQueryContext struct {
	Name                             toolbelt.CasedString
	HasParams, ParamsIsSingularField bool
	Params                           []GenerateField
	SQL                              string
	HasResponse                      bool
	ResponseIsSingularField          bool
	ResponseFields                   []GenerateField
	ResponseHasMultiple              bool
}

type GenerateQueriesContext struct {
	NeedsTimePackage bool
	PackageName      toolbelt.CasedString
	Queries          []GenerateQueryContext
}
