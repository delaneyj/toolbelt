package jtd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Parser handles parsing JTD schemas and resolving references
type Parser struct {
	rootSchema  *Schema
	definitions map[string]*Schema
	resolved    map[string]*Schema
}

// NewParser creates a new JTD parser
func NewParser() *Parser {
	return &Parser{
		resolved: make(map[string]*Schema),
	}
}

// ParseFile parses a JTD schema from a file
func (p *Parser) ParseFile(filename string) (*Schema, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses a JTD schema from JSON data
func (p *Parser) Parse(data []byte) (*Schema, error) {
	schema, err := ParseSchema(data)
	if err != nil {
		return nil, err
	}

	p.rootSchema = schema
	p.definitions = schema.Definitions
	if p.definitions == nil {
		p.definitions = make(map[string]*Schema)
	}

	// Resolve all references
	if err := p.resolveRefs(schema); err != nil {
		return nil, fmt.Errorf("failed to resolve references: %w", err)
	}

	return schema, nil
}

// resolveRefs recursively resolves all references in a schema
func (p *Parser) resolveRefs(schema *Schema) error {
	if schema == nil {
		return nil
	}

	// Handle ref form
	if schema.Ref != "" {
		// Check for circular references
		if _, exists := p.resolved[schema.Ref]; exists {
			// Already resolved or in process
			return nil
		}

		// Mark as being resolved to detect cycles
		p.resolved[schema.Ref] = nil

		// Find the referenced schema
		ref, exists := p.definitions[schema.Ref]
		if !exists {
			return fmt.Errorf("undefined reference: %s", schema.Ref)
		}

		// Resolve references in the referenced schema
		if err := p.resolveRefs(ref); err != nil {
			return err
		}

		// Store resolved reference
		p.resolved[schema.Ref] = ref
		return nil
	}

	// Handle other forms recursively
	switch schema.Form() {
	case "elements":
		return p.resolveRefs(schema.Elements)
	case "properties":
		for _, prop := range schema.Properties {
			if err := p.resolveRefs(prop); err != nil {
				return err
			}
		}
		for _, prop := range schema.OptionalProperties {
			if err := p.resolveRefs(prop); err != nil {
				return err
			}
		}
	case "values":
		return p.resolveRefs(schema.Values)
	case "discriminator":
		for _, mapping := range schema.Mapping {
			if err := p.resolveRefs(mapping); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetResolvedSchema returns a resolved schema for a given reference
func (p *Parser) GetResolvedSchema(ref string) *Schema {
	return p.resolved[ref]
}

// GetDefinition returns a definition by name
func (p *Parser) GetDefinition(name string) *Schema {
	return p.definitions[name]
}

// SchemaInfo contains information about a schema for code generation
type SchemaInfo struct {
	Name        string
	GoName      string
	Schema      *Schema
	IsRoot      bool
	IsNullable  bool
	Description string
}

// ExtractSchemas extracts all schemas that need code generation
func (p *Parser) ExtractSchemas(rootTypeName string) []SchemaInfo {
	var schemas []SchemaInfo

	// Add root schema if it needs generation
	if p.rootSchema != nil && p.rootSchema.Form() != "empty" {
		// Use provided rootTypeName or default to "Root"
		if rootTypeName == "" {
			rootTypeName = "Root"
		}
		info := SchemaInfo{
			Name:       rootTypeName,
			GoName:     rootTypeName,
			Schema:     p.rootSchema,
			IsRoot:     true,
			IsNullable: p.rootSchema.Nullable,
		}
		if desc, ok := p.getDescription(p.rootSchema); ok {
			info.Description = desc
		}
		schemas = append(schemas, info)
	}

	// Add all definitions
	for name, schema := range p.definitions {
		info := SchemaInfo{
			Name:       name,
			GoName:     toGoName(name),
			Schema:     schema,
			IsRoot:     false,
			IsNullable: schema.Nullable,
		}
		if desc, ok := p.getDescription(schema); ok {
			info.Description = desc
		}
		schemas = append(schemas, info)
	}

	return schemas
}

// getDescription extracts description from metadata
func (p *Parser) getDescription(schema *Schema) (string, bool) {
	if schema.Metadata == nil {
		return "", false
	}

	desc, ok := schema.Metadata["description"]
	if !ok {
		return "", false
	}

	descStr, ok := desc.(string)
	return descStr, ok
}

// toGoName converts a string to a valid Go identifier
func toGoName(s string) string {
	// Replace common separators with underscores
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ".", "_")

	// Split by underscores and capitalize each word
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}

// ValidateJSON validates JSON data against a schema
func (p *Parser) ValidateJSON(data []byte, schema *Schema) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return p.validate(v, schema, "")
}

// validate recursively validates a value against a schema
func (p *Parser) validate(value interface{}, schema *Schema, path string) error {
	if schema == nil {
		return nil // Empty schema accepts anything
	}

	// Handle nullable
	if value == nil {
		if schema.Nullable {
			return nil
		}
		return fmt.Errorf("null value at %s not allowed", path)
	}

	// Handle different forms
	switch schema.Form() {
	case "type":
		return p.validateType(value, schema.Type, path)
	case "enum":
		return p.validateEnum(value, schema.Enum, path)
	case "elements":
		return p.validateElements(value, schema.Elements, path)
	case "properties":
		return p.validateProperties(value, schema, path)
	case "values":
		return p.validateValues(value, schema.Values, path)
	case "discriminator":
		return p.validateDiscriminator(value, schema, path)
	case "ref":
		ref := p.GetDefinition(schema.Ref)
		if ref == nil {
			return fmt.Errorf("undefined reference: %s", schema.Ref)
		}
		return p.validate(value, ref, path)
	case "empty":
		return nil // Empty schema accepts anything
	}

	return fmt.Errorf("invalid schema form at %s", path)
}

// validateType validates a value against a type constraint
func (p *Parser) validateType(value interface{}, typ Type, path string) error {
	switch typ {
	case TypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean at %s, got %T", path, value)
		}
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string at %s, got %T", path, value)
		}
	case TypeTimestamp:
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string timestamp at %s, got %T", path, value)
		}
		// RFC3339 validation would go here
		_ = str
	case TypeFloat32, TypeFloat64:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected number at %s, got %T", path, value)
		}
	case TypeInt8, TypeInt16, TypeInt32, TypeUint8, TypeUint16, TypeUint32:
		// JSON numbers can be float64
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected number at %s, got %T", path, value)
		}
		// Check if it's an integer
		if num != float64(int64(num)) {
			return fmt.Errorf("expected integer at %s, got float", path)
		}
		// Range validation would go here based on type
	}
	return nil
}

// validateEnum validates a value against enum constraint
func (p *Parser) validateEnum(value interface{}, enum []string, path string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string for enum at %s, got %T", path, value)
	}

	for _, e := range enum {
		if str == e {
			return nil
		}
	}

	return fmt.Errorf("value %q not in enum at %s", str, path)
}

// validateElements validates an array against elements constraint
func (p *Parser) validateElements(value interface{}, elements *Schema, path string) error {
	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected array at %s, got %T", path, value)
	}

	for i, elem := range arr {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		if err := p.validate(elem, elements, elemPath); err != nil {
			return err
		}
	}

	return nil
}

// validateProperties validates an object against properties constraints
func (p *Parser) validateProperties(value interface{}, schema *Schema, path string) error {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object at %s, got %T", path, value)
	}

	// Check required properties
	for name, prop := range schema.Properties {
		val, exists := obj[name]
		if !exists {
			return fmt.Errorf("missing required property %s at %s", name, path)
		}
		propPath := fmt.Sprintf("%s.%s", path, name)
		if err := p.validate(val, prop, propPath); err != nil {
			return err
		}
	}

	// Check optional properties
	for name, prop := range schema.OptionalProperties {
		if val, exists := obj[name]; exists {
			propPath := fmt.Sprintf("%s.%s", path, name)
			if err := p.validate(val, prop, propPath); err != nil {
				return err
			}
		}
	}

	// Check for additional properties
	if !schema.AdditionalProperties {
		for name := range obj {
			if _, required := schema.Properties[name]; !required {
				if _, optional := schema.OptionalProperties[name]; !optional {
					return fmt.Errorf("unexpected property %s at %s", name, path)
				}
			}
		}
	}

	return nil
}

// validateValues validates a map against values constraint
func (p *Parser) validateValues(value interface{}, values *Schema, path string) error {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object at %s, got %T", path, value)
	}

	for key, val := range obj {
		valPath := fmt.Sprintf("%s[%q]", path, key)
		if err := p.validate(val, values, valPath); err != nil {
			return err
		}
	}

	return nil
}

// validateDiscriminator validates a discriminated union
func (p *Parser) validateDiscriminator(value interface{}, schema *Schema, path string) error {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object at %s, got %T", path, value)
	}

	// Get discriminator value
	discValue, exists := obj[schema.Discriminator]
	if !exists {
		return fmt.Errorf("missing discriminator %s at %s", schema.Discriminator, path)
	}

	discStr, ok := discValue.(string)
	if !ok {
		return fmt.Errorf("discriminator must be string at %s", path)
	}

	// Find matching schema
	mappingSchema, exists := schema.Mapping[discStr]
	if !exists {
		return fmt.Errorf("unknown discriminator value %q at %s", discStr, path)
	}

	// Validate against mapping schema
	return p.validateProperties(value, mappingSchema, path)
}
