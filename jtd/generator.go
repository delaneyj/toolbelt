package jtd

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"
)

// Generator generates Go code from JTD schemas
type Generator struct {
	parser       *Parser
	packageName  string
	imports      map[string]bool
	typeNames    map[string]string
	rootTypeName string
}

// GeneratorOptions configures the code generator
type GeneratorOptions struct {
	PackageName      string
	GenerateComments bool
	GenerateValidate bool
	TypeNamePrefix   string
	TypeNameSuffix   string
	RootTypeName     string // Name for the root schema type (defaults to package name if empty)
}

// NewGenerator creates a new code generator
func NewGenerator(parser *Parser, opts GeneratorOptions) *Generator {
	if opts.PackageName == "" {
		opts.PackageName = "types"
	}

	// Default root type name to capitalized package name if not specified
	rootTypeName := opts.RootTypeName
	if rootTypeName == "" {
		// Capitalize first letter of package name
		if len(opts.PackageName) > 0 {
			rootTypeName = strings.ToUpper(opts.PackageName[:1]) + opts.PackageName[1:]
		} else {
			rootTypeName = "Root"
		}
	}

	return &Generator{
		parser:       parser,
		packageName:  opts.PackageName,
		imports:      make(map[string]bool),
		typeNames:    make(map[string]string),
		rootTypeName: rootTypeName,
	}
}

// Generate generates Go code for all schemas
func (g *Generator) Generate() ([]byte, error) {
	var buf bytes.Buffer

	// Write package declaration
	fmt.Fprintf(&buf, "package %s\n\n", g.packageName)

	// Extract all schemas
	schemas := g.parser.ExtractSchemas(g.rootTypeName)

	// Generate type names first (for references)
	for _, info := range schemas {
		if info.Schema.Form() == "ref" {
			continue
		}
		g.typeNames[info.Name] = info.GoName
	}

	// Generate types
	var typeDecls []string
	for _, info := range schemas {
		if info.Schema.Form() == "ref" {
			continue
		}

		decl, err := g.generateType(info)
		if err != nil {
			return nil, fmt.Errorf("failed to generate type %s: %w", info.Name, err)
		}
		if decl != "" {
			typeDecls = append(typeDecls, decl)
		}
	}

	// Write imports if needed
	if len(g.imports) > 0 {
		buf.WriteString("import (\n")
		imports := make([]string, 0, len(g.imports))
		for imp := range g.imports {
			imports = append(imports, imp)
		}
		sort.Strings(imports)
		for _, imp := range imports {
			fmt.Fprintf(&buf, "\t%q\n", imp)
		}
		buf.WriteString(")\n\n")
	}

	// Write type declarations
	for i, decl := range typeDecls {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(decl)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code for debugging
		return buf.Bytes(), fmt.Errorf("failed to format code: %w", err)
	}

	return formatted, nil
}

// generateType generates a Go type declaration for a schema
func (g *Generator) generateType(info SchemaInfo) (string, error) {
	var buf bytes.Buffer

	// Add comment if description exists
	if info.Description != "" {
		fmt.Fprintf(&buf, "// %s %s\n", info.GoName, info.Description)
	}

	switch info.Schema.Form() {
	case "type", "empty":
		// For simple types, generate a type alias
		goType := g.getGoType(info.Schema, false)
		fmt.Fprintf(&buf, "type %s %s\n", info.GoName, goType)

	case "enum":
		// Generate enum type and constants
		fmt.Fprintf(&buf, "type %s string\n\n", info.GoName)
		fmt.Fprintf(&buf, "const (\n")
		for _, value := range info.Schema.Enum {
			constName := fmt.Sprintf("%s%s", info.GoName, toGoName(value))
			fmt.Fprintf(&buf, "\t%s %s = %q\n", constName, info.GoName, value)
		}
		fmt.Fprintf(&buf, ")\n")

	case "elements":
		// Array type
		elemType := g.getGoType(info.Schema.Elements, false)
		fmt.Fprintf(&buf, "type %s []%s\n", info.GoName, elemType)

	case "properties":
		// Struct type
		fmt.Fprintf(&buf, "type %s struct {\n", info.GoName)

		// Generate fields for required properties
		var fields []fieldInfo
		for name, prop := range info.Schema.Properties {
			fields = append(fields, fieldInfo{
				name:     name,
				goName:   toGoName(name),
				schema:   prop,
				required: true,
			})
		}

		// Generate fields for optional properties
		for name, prop := range info.Schema.OptionalProperties {
			fields = append(fields, fieldInfo{
				name:     name,
				goName:   toGoName(name),
				schema:   prop,
				required: false,
			})
		}

		// Sort fields for consistent output
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].name < fields[j].name
		})

		// Generate field declarations
		for _, field := range fields {
			g.generateStructField(&buf, field)
		}

		// Add additional properties field if allowed
		if info.Schema.AdditionalProperties {
			g.imports["encoding/json"] = true
			fmt.Fprintf(&buf, "\t// AdditionalProperties captures any extra properties\n")
			fmt.Fprintf(&buf, "\tAdditionalProperties map[string]json.RawMessage `json:\"-\"`\n")
		}

		fmt.Fprintf(&buf, "}\n")

	case "values":
		// Map type
		valueType := g.getGoType(info.Schema.Values, false)
		fmt.Fprintf(&buf, "type %s map[string]%s\n", info.GoName, valueType)

	case "discriminator":
		// Tagged union - generate interface and concrete types
		fmt.Fprintf(&buf, "// %s is a discriminated union\n", info.GoName)
		fmt.Fprintf(&buf, "type %s interface {\n", info.GoName)
		fmt.Fprintf(&buf, "\tis%s()\n", info.GoName)
		fmt.Fprintf(&buf, "\t%s() string\n", toGoName(info.Schema.Discriminator))
		fmt.Fprintf(&buf, "}\n")

		// Generate concrete types for each mapping
		for tag, schema := range info.Schema.Mapping {
			typeName := fmt.Sprintf("%s%s", info.GoName, toGoName(tag))

			// Generate the concrete type
			fmt.Fprintf(&buf, "\n// %s is the %q variant of %s\n", typeName, tag, info.GoName)
			fmt.Fprintf(&buf, "type %s struct {\n", typeName)

			// Add discriminator field with underscore prefix to avoid conflicts
			discriminatorFieldName := toGoName(info.Schema.Discriminator)
			if discriminatorFieldName == "Type" {
				discriminatorFieldName = "Type_"
			}
			fmt.Fprintf(&buf, "\t%s string `json:%q`\n", discriminatorFieldName, info.Schema.Discriminator)

			// Add fields from the mapping schema
			var fields []fieldInfo
			for name, prop := range schema.Properties {
				fields = append(fields, fieldInfo{
					name:     name,
					goName:   toGoName(name),
					schema:   prop,
					required: true,
				})
			}
			for name, prop := range schema.OptionalProperties {
				fields = append(fields, fieldInfo{
					name:     name,
					goName:   toGoName(name),
					schema:   prop,
					required: false,
				})
			}

			sort.Slice(fields, func(i, j int) bool {
				return fields[i].name < fields[j].name
			})

			for _, field := range fields {
				g.generateStructField(&buf, field)
			}

			fmt.Fprintf(&buf, "}\n")

			// Implement interface methods
			fmt.Fprintf(&buf, "\nfunc (%s) is%s() {}\n", typeName, info.GoName)
			fmt.Fprintf(&buf, "func (v %s) %s() string { return v.%s }\n", typeName, toGoName(info.Schema.Discriminator), discriminatorFieldName)
		}
	}

	return buf.String(), nil
}

// fieldInfo contains information about a struct field
type fieldInfo struct {
	name     string
	goName   string
	schema   *Schema
	required bool
}

// generateStructField generates a struct field declaration
func (g *Generator) generateStructField(buf *bytes.Buffer, field fieldInfo) {
	// Add field comment if description exists
	if desc, ok := field.schema.Metadata["description"].(string); ok && desc != "" {
		fmt.Fprintf(buf, "\t// %s %s\n", field.goName, desc)
	}

	// Get field type
	fieldType := g.getGoType(field.schema, !field.required)

	// Generate JSON tag
	jsonTag := field.name
	if !field.required {
		jsonTag += ",omitempty"
	}

	fmt.Fprintf(buf, "\t%s %s `json:%q`\n", field.goName, fieldType, jsonTag)
}

// getGoType returns the Go type for a schema
func (g *Generator) getGoType(schema *Schema, forcePointer bool) string {
	if schema == nil {
		return "any"
	}

	var baseType string

	switch schema.Form() {
	case "type":
		baseType = schema.Type.ToGoType()
		if baseType == "time.Time" {
			g.imports["time"] = true
		}

	case "enum":
		// Find the enum type name
		for name, s := range g.parser.definitions {
			if s == schema {
				baseType = g.typeNames[name]
				break
			}
		}
		if baseType == "" {
			baseType = "string" // Fallback
		}

	case "elements":
		elemType := g.getGoType(schema.Elements, false)
		baseType = "[]" + elemType

	case "properties":
		// Find the struct type name
		for name, s := range g.parser.definitions {
			if s == schema {
				baseType = g.typeNames[name]
				break
			}
		}
		if baseType == "" {
			baseType = "struct{}" // Anonymous struct
		}

	case "values":
		valueType := g.getGoType(schema.Values, false)
		baseType = "map[string]" + valueType

	case "discriminator":
		// Find the interface type name
		for name, s := range g.parser.definitions {
			if s == schema {
				baseType = g.typeNames[name]
				break
			}
		}
		if baseType == "" {
			baseType = "any"
		}

	case "ref":
		// Resolve reference
		ref := g.parser.GetDefinition(schema.Ref)
		if ref != nil {
			return g.getGoType(ref, forcePointer)
		}
		baseType = g.typeNames[schema.Ref]
		if baseType == "" {
			baseType = "any"
		}

	case "empty":
		baseType = "any"

	default:
		baseType = "any"
	}

	// Add pointer if nullable or optional
	if schema.Nullable || forcePointer {
		// Don't add pointer to slices, maps, or interfaces
		if !strings.HasPrefix(baseType, "[]") &&
			!strings.HasPrefix(baseType, "map[") &&
			baseType != "any" {
			baseType = "*" + baseType
		}
	}

	return baseType
}

// GenerateValidation generates validation methods for types
func (g *Generator) GenerateValidation() ([]byte, error) {
	var buf bytes.Buffer

	schemas := g.parser.ExtractSchemas("")

	for _, info := range schemas {
		if info.Schema.Form() == "ref" {
			continue
		}

		switch info.Schema.Form() {
		case "enum":
			// Generate validation for enum
			fmt.Fprintf(&buf, "\n// IsValid returns true if the value is a valid %s\n", info.GoName)
			fmt.Fprintf(&buf, "func (v %s) IsValid() bool {\n", info.GoName)
			fmt.Fprintf(&buf, "\tswitch v {\n")
			fmt.Fprintf(&buf, "\tcase ")
			for i, value := range info.Schema.Enum {
				if i > 0 {
					fmt.Fprintf(&buf, ", ")
				}
				constName := fmt.Sprintf("%s%s", info.GoName, toGoName(value))
				fmt.Fprintf(&buf, "%s", constName)
			}
			fmt.Fprintf(&buf, ":\n\t\treturn true\n")
			fmt.Fprintf(&buf, "\t}\n\treturn false\n}\n")

		case "properties":
			// Could generate validation methods for required fields
			// This is left as an exercise
		}
	}

	return buf.Bytes(), nil
}

// fileTemplate is the template for the generated file
const fileTemplate = `// Code generated by jtd2go. DO NOT EDIT.

package {{.Package}}

{{if .Imports}}
import (
{{range .Imports}}	"{{.}}"
{{end}})
{{end}}

{{range .Types}}
{{.}}
{{end}}
`

// templateData holds data for the file template
type templateData struct {
	Package string
	Imports []string
	Types   []string
}

// GenerateWithTemplate generates code using a template
func (g *Generator) GenerateWithTemplate() ([]byte, error) {
	tmpl, err := template.New("file").Parse(fileTemplate)
	if err != nil {
		return nil, err
	}

	// Extract schemas and generate types
	schemas := g.parser.ExtractSchemas("")
	var types []string

	for _, info := range schemas {
		if info.Schema.Form() == "ref" {
			continue
		}

		decl, err := g.generateType(info)
		if err != nil {
			return nil, err
		}
		if decl != "" {
			types = append(types, decl)
		}
	}

	// Prepare template data
	imports := make([]string, 0, len(g.imports))
	for imp := range g.imports {
		imports = append(imports, imp)
	}
	sort.Strings(imports)

	data := templateData{
		Package: g.packageName,
		Imports: imports,
		Types:   types,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	// Format the code
	return format.Source(buf.Bytes())
}
