package zombiezen

import (
	"fmt"

	"strings"

	"github.com/delaneyj/toolbelt"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func generateCRUD(req *plugin.GenerateRequest) (files []*plugin.File, err error) {
	pluralClient := pluralize.NewClient()

	packageName := toolbelt.ToCasedString(req.Settings.Codegen.Out)
	for _, schema := range req.Catalog.Schemas {
		schemaName := toolbelt.ToCasedString(schema.Name)

		for _, table := range schema.Tables {
			tbl := &GenerateCRUDTable{
				PackageName: packageName,
				Schema:      schemaName,
				Name:        toolbelt.ToCasedString(table.Rel.Name),
				SingleName:  toolbelt.ToCasedString(pluralClient.Singular(table.Rel.Name)),
			}

			if strings.HasSuffix(tbl.Name.Snake, "_fts") {
				continue
			}
			for i, column := range table.Columns {
				if column.Name == "id" {
					tbl.HasID = true
				}
				columnName := toolbelt.ToCasedString(column.Name)

				goType, needsTime := toGoType(column)
				if needsTime {
					tbl.NeedsTimePackage = true
				}
				f := GenerateField{
					Column:     i + 1,
					Offset:     i,
					Name:       columnName,
					SQLType:    toolbelt.ToCasedString(toSQLType(column)),
					GoType:     toolbelt.ToCasedString(goType),
					IsNullable: !column.NotNull,
				}
				tbl.Fields = append(tbl.Fields, f)
			}

			contents := GenerateCRUD(tbl)
			filename := fmt.Sprintf("crud_%s_%s.go", schemaName.Snake, tbl.Name.Snake)

			files = append(files, &plugin.File{
				Name:     filename,
				Contents: []byte(contents),
			})
		}
	}
	return files, nil
}

type GenerateCRUDTable struct {
	PackageName      toolbelt.CasedString
	NeedsTimePackage bool
	Schema           toolbelt.CasedString
	Name             toolbelt.CasedString
	SingleName       toolbelt.CasedString
	Fields           []GenerateField
	HasID            bool
}
