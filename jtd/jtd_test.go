package jtd

import (
	"testing"
)

func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name: "valid type schema",
			schema: Schema{
				Type: TypeString,
			},
			wantErr: false,
		},
		{
			name: "valid enum schema",
			schema: Schema{
				Enum: []string{"foo", "bar", "baz"},
			},
			wantErr: false,
		},
		{
			name: "invalid empty enum",
			schema: Schema{
				Enum: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid duplicate enum",
			schema: Schema{
				Enum: []string{"foo", "bar", "foo"},
			},
			wantErr: true,
		},
		{
			name: "valid properties schema",
			schema: Schema{
				Properties: map[string]*Schema{
					"name": {Type: TypeString},
					"age":  {Type: TypeInt32},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid overlapping properties",
			schema: Schema{
				Properties: map[string]*Schema{
					"name": {Type: TypeString},
				},
				OptionalProperties: map[string]*Schema{
					"name": {Type: TypeString},
				},
			},
			wantErr: true,
		},
		{
			name: "valid discriminator schema",
			schema: Schema{
				Discriminator: "type",
				Mapping: map[string]*Schema{
					"user": {
						Properties: map[string]*Schema{
							"name": {Type: TypeString},
						},
					},
					"admin": {
						Properties: map[string]*Schema{
							"name":  {Type: TypeString},
							"level": {Type: TypeInt32},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid discriminator with non-properties mapping",
			schema: Schema{
				Discriminator: "type",
				Mapping: map[string]*Schema{
					"user": {Type: TypeString},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid multiple forms",
			schema: Schema{
				Type: TypeString,
				Enum: []string{"foo", "bar"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Schema.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseSchema(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "simple type schema",
			json: `{"type": "string"}`,
		},
		{
			name: "nullable type",
			json: `{"type": "string", "nullable": true}`,
		},
		{
			name: "enum schema",
			json: `{"enum": ["red", "green", "blue"]}`,
		},
		{
			name: "array schema",
			json: `{"elements": {"type": "string"}}`,
		},
		{
			name: "object schema",
			json: `{
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "int32"}
				},
				"optionalProperties": {
					"bio": {"type": "string"}
				}
			}`,
		},
		{
			name: "map schema",
			json: `{"values": {"type": "float64"}}`,
		},
		{
			name: "discriminator schema",
			json: `{
				"discriminator": "type",
				"mapping": {
					"email": {
						"properties": {
							"to": {"type": "string"}
						}
					}
				}
			}`,
		},
		{
			name: "schema with definitions",
			json: `{
				"definitions": {
					"user": {
						"properties": {
							"id": {"type": "string"}
						}
					}
				},
				"ref": "user"
			}`,
		},
		{
			name:    "invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "invalid schema",
			json:    `{"type": "invalid"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSchema([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeToGoType(t *testing.T) {
	tests := []struct {
		typ  Type
		want string
	}{
		{TypeBoolean, "bool"},
		{TypeString, "string"},
		{TypeTimestamp, "time.Time"},
		{TypeFloat32, "float32"},
		{TypeFloat64, "float64"},
		{TypeInt8, "int8"},
		{TypeInt16, "int16"},
		{TypeInt32, "int32"},
		{TypeUint8, "uint8"},
		{TypeUint16, "uint16"},
		{TypeUint32, "uint32"},
		{Type("unknown"), "any"},
	}

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			if got := tt.typ.ToGoType(); got != tt.want {
				t.Errorf("Type.ToGoType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONValidation(t *testing.T) {
	parser := NewParser()

	schemaJSON := `{
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "int32"}
		}
	}`

	schema, err := parser.Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid object",
			json:    `{"name": "John", "age": 30}`,
			wantErr: false,
		},
		{
			name:    "missing required property",
			json:    `{"name": "John"}`,
			wantErr: true,
		},
		{
			name:    "wrong type",
			json:    `{"name": "John", "age": "thirty"}`,
			wantErr: true,
		},
		{
			name:    "extra property",
			json:    `{"name": "John", "age": 30, "extra": true}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateJSON([]byte(tt.json), schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaForm(t *testing.T) {
	tests := []struct {
		name   string
		schema Schema
		want   string
	}{
		{
			name:   "empty schema",
			schema: Schema{},
			want:   "empty",
		},
		{
			name:   "type form",
			schema: Schema{Type: TypeString},
			want:   "type",
		},
		{
			name:   "enum form",
			schema: Schema{Enum: []string{"a", "b"}},
			want:   "enum",
		},
		{
			name:   "elements form",
			schema: Schema{Elements: &Schema{Type: TypeString}},
			want:   "elements",
		},
		{
			name:   "properties form",
			schema: Schema{Properties: map[string]*Schema{"x": {Type: TypeString}}},
			want:   "properties",
		},
		{
			name:   "values form",
			schema: Schema{Values: &Schema{Type: TypeString}},
			want:   "values",
		},
		{
			name:   "discriminator form",
			schema: Schema{Discriminator: "type", Mapping: map[string]*Schema{}},
			want:   "discriminator",
		},
		{
			name:   "ref form",
			schema: Schema{Ref: "foo"},
			want:   "ref",
		},
		{
			name:   "invalid multiple forms",
			schema: Schema{Type: TypeString, Ref: "foo"},
			want:   "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.schema.Form(); got != tt.want {
				t.Errorf("Schema.Form() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratorIntegration(t *testing.T) {
	schemaJSON := `{
		"definitions": {
			"user": {
				"properties": {
					"id": {"type": "string"},
					"name": {"type": "string"},
					"age": {"type": "int32", "nullable": true}
				},
				"optionalProperties": {
					"bio": {"type": "string"}
				}
			},
			"role": {
				"enum": ["admin", "user", "guest"]
			}
		}
	}`

	parser := NewParser()
	_, err := parser.Parse([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	generator := NewGenerator(parser, GeneratorOptions{
		PackageName: "test",
	})

	code, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Basic smoke test - check that code was generated
	if len(code) == 0 {
		t.Error("Generated code is empty")
	}

	// Check that it contains expected content
	codeStr := string(code)
	expectedContent := []string{
		"package test",
		"type User struct",
		"type Role string",
		"RoleAdmin",
		"RoleUser",
		"RoleGuest",
	}

	for _, expected := range expectedContent {
		if !contains(codeStr, expected) {
			t.Errorf("Generated code does not contain expected content: %s", expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}