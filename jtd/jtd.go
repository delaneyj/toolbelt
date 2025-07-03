package jtd

import (
	"encoding/json"
	"fmt"
)

// Schema represents a JSON Type Definition schema according to RFC 8927
type Schema struct {
	// Metadata
	Metadata map[string]any `json:"metadata,omitempty"`

	// Nullable indicates whether the value can be null
	Nullable bool `json:"nullable,omitempty"`

	// Type forms
	Type Type `json:"type,omitempty"`

	// Enum form
	Enum []string `json:"enum,omitempty"`

	// Elements form (for arrays)
	Elements *Schema `json:"elements,omitempty"`

	// Properties form (for objects)
	Properties           map[string]*Schema `json:"properties,omitempty"`
	OptionalProperties   map[string]*Schema `json:"optionalProperties,omitempty"`
	AdditionalProperties bool               `json:"additionalProperties,omitempty"`

	// Values form (for maps)
	Values *Schema `json:"values,omitempty"`

	// Discriminator form (for tagged unions)
	Discriminator string              `json:"discriminator,omitempty"`
	Mapping       map[string]*Schema `json:"mapping,omitempty"`

	// Ref form
	Ref string `json:"ref,omitempty"`

	// Definitions (for root schema only)
	Definitions map[string]*Schema `json:"definitions,omitempty"`
}

// Type represents the primitive types in JTD
type Type string

const (
	TypeBoolean   Type = "boolean"
	TypeString    Type = "string"
	TypeTimestamp Type = "timestamp"
	TypeFloat32   Type = "float32"
	TypeFloat64   Type = "float64"
	TypeInt8      Type = "int8"
	TypeInt16     Type = "int16"
	TypeInt32     Type = "int32"
	TypeUint8     Type = "uint8"
	TypeUint16    Type = "uint16"
	TypeUint32    Type = "uint32"
)

// IsValid checks if the type is a valid JTD type
func (t Type) IsValid() bool {
	switch t {
	case TypeBoolean, TypeString, TypeTimestamp,
		TypeFloat32, TypeFloat64,
		TypeInt8, TypeInt16, TypeInt32,
		TypeUint8, TypeUint16, TypeUint32:
		return true
	}
	return false
}

// ToGoType converts JTD type to Go type
func (t Type) ToGoType() string {
	switch t {
	case TypeBoolean:
		return "bool"
	case TypeString:
		return "string"
	case TypeTimestamp:
		return "time.Time"
	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"
	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	default:
		return "any"
	}
}

// Form returns the form of the schema
func (s *Schema) Form() string {
	if s == nil {
		return "empty"
	}

	formCount := 0
	form := ""

	if s.Ref != "" {
		formCount++
		form = "ref"
	}
	if s.Type != "" {
		formCount++
		form = "type"
	}
	if s.Enum != nil {
		formCount++
		form = "enum"
	}
	if s.Elements != nil {
		formCount++
		form = "elements"
	}
	if len(s.Properties) > 0 || len(s.OptionalProperties) > 0 {
		formCount++
		form = "properties"
	}
	if s.Values != nil {
		formCount++
		form = "values"
	}
	if s.Discriminator != "" {
		formCount++
		form = "discriminator"
	}

	if formCount == 0 {
		return "empty"
	}
	if formCount > 1 {
		return "invalid"
	}

	return form
}

// Validate checks if the schema is valid according to RFC 8927
func (s *Schema) Validate() error {
	form := s.Form()
	if form == "invalid" {
		return fmt.Errorf("schema has multiple forms")
	}

	switch form {
	case "type":
		if !s.Type.IsValid() {
			return fmt.Errorf("invalid type: %s", s.Type)
		}
	case "enum":
		if len(s.Enum) == 0 {
			return fmt.Errorf("enum cannot be empty")
		}
		// Check for duplicates
		seen := make(map[string]bool)
		for _, v := range s.Enum {
			if seen[v] {
				return fmt.Errorf("duplicate enum value: %s", v)
			}
			seen[v] = true
		}
	case "elements":
		if err := s.Elements.Validate(); err != nil {
			return fmt.Errorf("invalid elements schema: %w", err)
		}
	case "properties":
		// Check for overlapping property names
		for name := range s.Properties {
			if _, exists := s.OptionalProperties[name]; exists {
				return fmt.Errorf("property %s appears in both properties and optionalProperties", name)
			}
		}
		// Validate all property schemas
		for name, prop := range s.Properties {
			if err := prop.Validate(); err != nil {
				return fmt.Errorf("invalid property %s: %w", name, err)
			}
		}
		for name, prop := range s.OptionalProperties {
			if err := prop.Validate(); err != nil {
				return fmt.Errorf("invalid optional property %s: %w", name, err)
			}
		}
	case "values":
		if err := s.Values.Validate(); err != nil {
			return fmt.Errorf("invalid values schema: %w", err)
		}
	case "discriminator":
		if s.Discriminator == "" {
			return fmt.Errorf("discriminator cannot be empty")
		}
		if len(s.Mapping) == 0 {
			return fmt.Errorf("discriminator mapping cannot be empty")
		}
		// Validate all mapping schemas
		for tag, schema := range s.Mapping {
			if err := schema.Validate(); err != nil {
				return fmt.Errorf("invalid discriminator mapping %s: %w", tag, err)
			}
			// Ensure mapping schemas are properties form
			if schema.Form() != "properties" {
				return fmt.Errorf("discriminator mapping %s must be properties form", tag)
			}
			// Ensure discriminator property doesn't exist in mapping schemas
			if _, exists := schema.Properties[s.Discriminator]; exists {
				return fmt.Errorf("discriminator property %s cannot exist in mapping %s", s.Discriminator, tag)
			}
			if _, exists := schema.OptionalProperties[s.Discriminator]; exists {
				return fmt.Errorf("discriminator property %s cannot exist in mapping %s optional properties", s.Discriminator, tag)
			}
		}
	case "ref":
		if s.Ref == "" {
			return fmt.Errorf("ref cannot be empty")
		}
	}

	return nil
}

// ParseSchema parses a JSON Type Definition schema from JSON
func ParseSchema(data []byte) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Validate root schema
	if err := schema.Validate(); err != nil {
		return nil, err
	}

	// Validate definitions if present
	for name, def := range schema.Definitions {
		if err := def.Validate(); err != nil {
			return nil, fmt.Errorf("invalid definition %s: %w", name, err)
		}
	}

	return &schema, nil
}