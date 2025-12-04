package zombiezen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/delaneyj/toolbelt"
	"github.com/delaneyj/toolbelt/bytebufferpool"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/samber/lo"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func generateQueries(req *plugin.GenerateRequest, opts *Options, packageName toolbelt.CasedString) (files []*plugin.File, err error) {
	pluralClient := pluralize.NewClient()
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
			goType, _ := toGoType(p.Column, opts)

			isSlice := p.Column.GetIsSqlcSlice()
			fieldGoType := toolbelt.ToCasedString(goType)
			if isSlice {
				fieldGoType = toolbelt.ToCasedString("[]" + goType)
			}

			param := GenerateField{
				Column:           int(p.Number),
				Offset:           int(p.Number) - 1,
				Name:             toolbelt.ToCasedString(toFieldName(p.Column)),
				SQLType:          toolbelt.ToCasedString(toSQLType(p.Column)),
				GoType:           fieldGoType,
				BindGoType:       toolbelt.ToCasedString(goType),
				OriginalName:     p.Column.Name,
				IsNullable:       !p.Column.NotNull,
				IsSlice:          isSlice,
				DurationFromText: isDurationFromText(p.Column),
			}
			if isSlice {
				queryCtx.HasSliceParams = true
			}
			if usesToolbeltParam(param) {
				queryCtx.NeedsToolbelt = true
			}
			return param
		})
		queryCtx.HasParams = len(q.Params) > 0
		queryCtx.ParamsIsSingularField = len(q.Params) == 1

		if len(q.Columns) > 0 {
			queryCtx.HasResponse = true
			queryCtx.ResponseFields = lo.Map(q.Columns, func(c *plugin.Column, ci int) GenerateField {
				goType, _ := toGoType(c, opts)

				col := GenerateField{
					Column:           ci + 1,
					Offset:           ci,
					Name:             toolbelt.ToCasedString(toFieldName(c)),
					SQLType:          toolbelt.ToCasedString(toSQLType(c)),
					GoType:           toolbelt.ToCasedString(goType),
					BindGoType:       toolbelt.ToCasedString(goType),
					OriginalName:     c.Name,
					IsNullable:       !c.NotNull,
					DurationFromText: isDurationFromText(c),
				}
				if usesToolbeltResponse(col) {
					queryCtx.NeedsToolbelt = true
				}
				return col
			})
			queryCtx.ResponseHasMultiple = q.Cmd == ":many"
			queryCtx.ResponseIsSingularField = len(q.Columns) == 1

			if modelName, ok := findModelReturn(pluralClient, req, q.Columns); ok {
				queryCtx.ResponseModelName = modelName
				queryCtx.ResponseHasModel = true
			}
		}
		queryCtx.NeedsTimePackage = queryNeedsTimeImport(queryCtx)

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

func toGoType(c *plugin.Column, opts *Options) (val string, needsTime bool) {
	typ := toolbelt.Lower(c.Type.Name)
	disableTime := opts != nil && opts.DisableTimeConversion

	if strings.HasSuffix(c.Name, "ms") {
		return "time.Duration", true
	}
	if strings.HasSuffix(c.Name, "_interval") && typ == "text" {
		return "time.Duration", true
	}

	if !disableTime && (c.Name == "at" || strings.HasSuffix(c.Name, "_at") || typ == "datetime") {
		return "time.Time", true
	}

	switch typ {
	case "text":
		return "string", false
	case "integer", "int":
		return "int64", false
	case "real":
		return "float64", false
	case "datetime":
		return "string", false
	case "boolean", "bool":
		return "bool", false
	case "blob":
		return "[]byte", false
	default:
		panic(fmt.Sprintf("toGoType unhandled type '%s' for column '%s'", c.Type.Name, c.Name))
	}
}

type GenerateField struct {
	Column           int // 1-indexed
	Offset           int // 0-indexed
	Name             toolbelt.CasedString
	SQLType          toolbelt.CasedString
	GoType           toolbelt.CasedString
	BindGoType       toolbelt.CasedString
	OriginalName     string
	IsNullable       bool
	IsSlice          bool
	DurationFromText bool
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

	ResponseHasModel  bool
	ResponseModelName toolbelt.CasedString

	NeedsTimePackage bool
	NeedsToolbelt    bool
	HasSliceParams   bool
}

func queryNeedsTimeImport(q *GenerateQueryContext) bool {
	usesDurationParse := func(fields []GenerateField) bool {
		for _, f := range fields {
			if f.GoType.Original == "time.Duration" && f.DurationFromText {
				return true
			}
		}
		return false
	}
	usesTimeType := func(fields []GenerateField) bool {
		for _, f := range fields {
			if strings.Contains(f.GoType.Original, "time.") {
				return true
			}
		}
		return false
	}

	if usesDurationParse(q.Params) || usesDurationParse(q.ResponseFields) {
		return true
	}

	if q.HasParams {
		if q.ParamsIsSingularField {
			if strings.Contains(q.Params[0].GoType.Original, "time.") {
				return true
			}
		} else if usesTimeType(q.Params) {
			return true
		}
	}

	if q.HasResponse {
		if q.ResponseIsSingularField {
			if strings.Contains(q.ResponseFields[0].GoType.Original, "time.") {
				return true
			}
		} else if !q.ResponseHasModel && usesTimeType(q.ResponseFields) {
			return true
		}
	}

	return false
}

func findModelReturn(pluralClient *pluralize.Client, req *plugin.GenerateRequest, cols []*plugin.Column) (toolbelt.CasedString, bool) {
	if len(cols) == 0 {
		return toolbelt.CasedString{}, false
	}

	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}

	var hint *plugin.Identifier
	for i, c := range cols {
		if c.Table == nil || c.Table.Name == "" {
			break
		}
		if i == 0 {
			hint = c.Table
			continue
		}
		if c.Table.GetName() != hint.GetName() || c.Table.GetSchema() != hint.GetSchema() {
			break
		}
	}

	candidates := make([]*plugin.Table, 0, 1)
	for _, schema := range req.Catalog.Schemas {
		if hint != nil && hint.GetSchema() != "" && schema.Name != hint.GetSchema() {
			continue
		}
		for _, t := range schema.Tables {
			if hint != nil && t.Rel != nil && t.Rel.GetName() != hint.GetName() {
				continue
			}
			if len(t.Columns) != len(names) {
				continue
			}
			match := true
			for i := range t.Columns {
				if t.Columns[i].Name != names[i] {
					match = false
					break
				}
			}
			if match {
				candidates = append(candidates, t)
			}
		}
	}

	if len(candidates) != 1 {
		return toolbelt.CasedString{}, false
	}

	return toolbelt.ToCasedString(pluralClient.Singular(candidates[0].Rel.GetName())), true
}

func isDurationFromText(c *plugin.Column) bool {
	typ := toolbelt.Lower(c.Type.Name)
	return strings.HasSuffix(c.Name, "_interval") && typ == "text"
}

func usesToolbeltResponse(f GenerateField) bool {
	switch f.GoType.Original {
	case "time.Time":
		return true
	case "time.Duration":
		return !f.DurationFromText
	case "[]byte":
		return true
	default:
		return false
	}
}

func usesToolbeltParam(f GenerateField) bool {
	switch f.BindGoType.Original {
	case "time.Time":
		return true
	case "time.Duration":
		return !f.DurationFromText
	default:
		return false
	}
}
