package zombiezen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/delaneyj/toolbelt"
	"github.com/samber/lo"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/valyala/bytebufferpool"
)

func generateQueries(req *plugin.GenerateRequest) (files []*plugin.File, err error) {
	packageName := toolbelt.ToCasedString(req.Settings.Codegen.Out)
	queries := make([]*GenerateQueryContext, len(req.Queries))
	for i, q := range req.Queries {
		queryCtx := &GenerateQueryContext{
			PackageName: packageName,
			Name:        toolbelt.ToCasedString(q.Name),
			SQL:         strings.TrimSpace(q.Text),
		}
		if queryCtx.SQL == "" {
			return nil, fmt.Errorf("query %s has no SQL", q.Name)
		}

		queryCtx.Params = lo.Map(q.Params, func(p *plugin.Parameter, pi int) GenerateField {
			goType, needsTime := toGoType(p.Column)
			if needsTime {
				queryCtx.NeedsTimePackage = true
			}

			param := GenerateField{
				Column:     int(p.Number),
				Offset:     int(p.Number) - 1,
				Name:       toolbelt.ToCasedString(toFieldName(p.Column)),
				SQLType:    toolbelt.ToCasedString(toSQLType(p.Column)),
				GoType:     toolbelt.ToCasedString(goType),
				IsNullable: !p.Column.NotNull,
			}
			return param
		})
		queryCtx.HasParams = len(q.Params) > 0
		queryCtx.ParamsIsSingularField = len(q.Params) == 1

		if len(q.Columns) > 0 {
			queryCtx.HasResponse = true
			queryCtx.ResponseFields = lo.Map(q.Columns, func(c *plugin.Column, ci int) GenerateField {
				goType, needsTime := toGoType(c)
				if needsTime {
					queryCtx.NeedsTimePackage = true
				}

				col := GenerateField{
					Column:     ci + 1,
					Offset:     ci,
					Name:       toolbelt.ToCasedString(toFieldName(c)),
					SQLType:    toolbelt.ToCasedString(toSQLType(c)),
					GoType:     toolbelt.ToCasedString(goType),
					IsNullable: !c.NotNull,
				}
				return col
			})
			queryCtx.ResponseHasMultiple = q.Cmd == ":many"
			queryCtx.ResponseIsSingularField = len(q.Columns) == 1
		}

		queries[i] = queryCtx
	}

	for _, q := range queries {
		buf := bytebufferpool.Get()
		defer bytebufferpool.Put(buf)
		queryContents := GenerateQuery(q)

		f := &plugin.File{
			Name:     fmt.Sprintf("%s.go", q.Name.Snake),
			Contents: []byte(queryContents),
		}
		files = append(files, f)
	}

	return files, nil
}

func toSQLType(c *plugin.Column) string {
	switch toolbelt.Lower(c.Type.Name) {
	case "text":
		return "text"
	case "integer", "int":
		return "int64"
	case "datetime", "real":
		return "float"
	case "boolean":
		return "bool"
	case "blob":
		return "bytes"
	case "bool":
		return "bool"
	default:
		panic(fmt.Sprintf("toSQLType unhandled type %s", c.Type.Name))
	}

}

func toFieldName(c *plugin.Column) string {
	n := c.Name
	if strings.HasSuffix(n, "_ms") {
		return n[:len(n)-3]
	}
	return n
}

func toGoType(c *plugin.Column) (val string, needsTime bool) {
	typ := toolbelt.Lower(c.Type.Name)

	if strings.HasSuffix(c.Name, "ms") {
		return "time.Duration", true
	} else if c.Name == "at" || strings.HasSuffix(c.Name, "_at") || typ == "datetime" {
		return "time.Time", true
	} else {
		switch typ {
		case "text":
			return "string", false
		case "integer", "int":
			return "int64", false
		case "real":
			return "float64", false
		case "boolean", "bool":
			return "bool", false
		case "blob":
			return "[]byte", false
		default:
			panic(fmt.Sprintf("toGoType unhandled type '%s' for column '%s'", c.Type.Name, c.Name))
		}
	}
}

type GenerateField struct {
	Column     int // 1-indexed
	Offset     int // 0-indexed
	Name       toolbelt.CasedString
	SQLType    toolbelt.CasedString
	GoType     toolbelt.CasedString
	IsNullable bool
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
