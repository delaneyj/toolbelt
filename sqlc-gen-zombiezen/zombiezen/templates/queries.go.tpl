package {{.PackageName.Lower}}

import (
    "fmt"
    "zombiezen.com/go/sqlite"

    {{- if .NeedsTimePackage}}
    "time"
    "github.com/delaneyj/toolbelt"
    {{- end}}
)

{{- define "fillResponse"}}
{{- if .ResponseIsSingularField}}
    {{(index .ResponseFields 0).Name.Camel}} = stmt.Column{{(index .ResponseFields 0).SQLType.Pascal}}(0)
{{- else}}
row := {{.ResponseType.Pascal}}{
{{- range .ResponseFields}}
    {{.Name.Pascal}} : stmt.Column{{.SQLType.Pascal}}({{.Column}}),
{{- end}}
}
{{- end}}
{{- end}}

{{- range .Queries}}
    {{- if and .HasResponse (not .ResponseIsSingularField)}}
type {{.Name.Pascal}}Res struct {
        {{- range .ResponseFields}}
    {{.Name.Pascal}} {{.GoType.Original}} `json:"{{.Name.Lower}}"`
        {{- end}}
}
    {{- end}}

    {{- if and .HasParams (not .ParamsIsSingularField)}}
type {{.Name.Pascal}}Params struct{
        {{- range .Params}}
    {{.Name.Pascal}} {{.GoType.Original}} `json:"{{.Name.Lower}}"`
        {{- end}}
}
    {{- end}}

func {{.Name.Pascal}}(
    tx *sqlite.Conn,
    {{- if .HasParams}}
    {{- if .ParamsIsSingularField}}
    {{(index .Params 0).Name.Lower}} {{(index .Params 0).GoType.Original}},
    {{- else}}
    params {{.Name.Pascal}}Params,
    {{- end}}
    {{- end}}
) (

    {{- if .HasResponse}}
    {{- if .ResponseIsSingularField}}
    {{(index .ResponseFields 0).Name.Camel}} {{(index .ResponseFields 0).GoType.Original}},
    {{- else}}
    res {{if .ResponseHasMultiple}}[]{{else}}*{{end}}{{.Name.Pascal}}Res,
    {{- end}}
    {{- end}}
    err error,
) {
    // Prepare statement into cache
    stmt := tx.Prep(`{{.SQL}}`)
    defer stmt.Reset()

    {{ if len .Params -}}
    // Bind parameters
        {{$singular := .ParamsIsSingularField}}
        {{- range .Params}}
            {{- if eq .GoType.Original "time.Time"}}
    stmt.Bind{{.SQLType.Pascal}}({{.Column}}, toolbelt.TimeToJulianDay({{- if not $singular}}params.{{end}}{{.Name.Pascal}}))
            {{- else if eq .GoType.Original "time.Duration"}}
    stmt.Bind{{.SQLType.Pascal}}({{.Column}}, toolbelt.DurationToMilliseconds({{- if not $singular}}params.{{end}}{{.Name.Pascal}}))
            {{- else }}
    stmt.Bind{{.SQLType.Pascal}}({{.Column}}, {{- if $singular}}{{.Name.Camel}}{{else}}params.{{.Name.Pascal}}{{end}})
            {{- end}}
        {{- end}}
    {{- end}}

    // Execute query
    {{- if .HasResponse}}
        {{- if .ResponseHasMultiple}}
    for {
        if hasRow, err := stmt.Step(); err != nil {
            return res, fmt.Errorf("failed to execute {{.Name.Lower}} SQL: %w", err)
        } else if !hasRow {
            break
        }
            {{template "fillResponse" .}}

        res = append(res, row)
    }
        {{- else}}
    if hasRow, err := stmt.Step(); err != nil {
        {{- if .ResponseIsSingularField}}
        return {{(index .ResponseFields 0).Name.Camel}}, fmt.Errorf("failed to execute {{.Name.Lower}} SQL: %w", err)
        {{- else}}
        return res, err
        {{- end}}
    } else if hasRow {
            {{template "fillResponse" .}}
        {{- if not .ResponseIsSingularField}}
        res = &row
        {{- end}}
    }
        {{- end}}
    {{- else}}
    if _, err := stmt.Step(); err != nil {
        return fmt.Errorf("failed to execute {{.Name.Lower}} SQL: %w", err)
    }
    {{- end}}

    {{- if .HasResponse -}}
        {{- if .ResponseIsSingularField}}
    return {{(index .ResponseFields 0).Name.Camel}}, nil
        {{- else}}
    return res, nil
        {{- end}}
    {{- else }}
    return nil
    {{- end }}
}
{{- end}}
