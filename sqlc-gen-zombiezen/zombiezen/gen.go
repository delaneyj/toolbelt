package zombiezen

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
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

	queries := lo.Map(req.Queries, func(q *plugin.Query, qi int) *GenerateQueryContext {
		queryCtx := &GenerateQueryContext{
			PackageName: toolbelt.ToCasedString(req.Settings.Codegen.Out),
			Name:        toolbelt.ToCasedString(q.Name),
		}
		queryCtx.Params = lo.Map(q.Params, func(p *plugin.Parameter, pi int) GenerateField {
			param := GenerateField{
				Column:  int(p.Number),
				Name:    toolbelt.ToCasedString(toFieldName(p.Column)),
				SQLType: toolbelt.ToCasedString(toSQLType(p.Column)),
				GoType:  toolbelt.ToCasedString(toGoType(queryCtx, p.Column)),
			}
			return param
		})
		queryCtx.HasParams = len(q.Params) > 0
		queryCtx.ParamsIsSingularField = len(q.Params) == 1

		if len(q.Columns) > 0 {
			queryCtx.HasResponse = true
			queryCtx.ResponseFields = lo.Map(q.Columns, func(c *plugin.Column, ci int) GenerateField {
				col := GenerateField{
					Column:  ci + 1,
					Name:    toolbelt.ToCasedString(toFieldName(c)),
					SQLType: toolbelt.ToCasedString(toSQLType(c)),
					GoType:  toolbelt.ToCasedString(toGoType(queryCtx, c)),
				}
				return col
			})
			queryCtx.ResponseHasMultiple = q.Cmd == ":many"
			queryCtx.SQL = q.Text
			queryCtx.ResponseIsSingularField = len(q.Columns) == 1
		}
		return queryCtx
	})

	files := make([]*plugin.File, len(queries))
	for i, q := range queries {

		buf := &bytes.Buffer{}
		if err := tmpls.ExecuteTemplate(buf, "queries.go.tpl", q); err != nil {
			return nil, fmt.Errorf("executing template: %w", err)
		}

		files[i] = &plugin.File{
			Name:     fmt.Sprintf("%s.go", q.Name.Snake),
			Contents: buf.Bytes(),
		}
	}

	return &plugin.GenerateResponse{
		Files: files,
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

func toFieldName(c *plugin.Column) string {
	n := c.Name
	log.Printf("toFieldName %s", n)
	if strings.HasSuffix(n, "_ms") {
		return n[:len(n)-3]
	}
	return n
}

func toGoType(queryCtx *GenerateQueryContext, c *plugin.Column) string {
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
	PackageName                      toolbelt.CasedString
	Name                             toolbelt.CasedString
	HasParams, ParamsIsSingularField bool
	Params                           []GenerateField
	SQL                              string
	HasResponse                      bool
	ResponseIsSingularField          bool
	ResponseFields                   []GenerateField
	ResponseHasMultiple              bool

	NeedsTimePackage bool
}
