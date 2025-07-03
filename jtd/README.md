# JTD to Go Generator

A JSON Type Definition (RFC 8927) to Go code generator.

## Installation

```bash
go install github.com/delaneyj/toolbelt/jtd/cmd/jtd2go@latest
```

## Usage

```bash
jtd2go -input schema.json -output types.go -package myapp
```

### Options

- `-input`: Input JTD schema file (required)
- `-output`: Output Go file (default: stdout)
- `-package`: Go package name (default: "types")
- `-comments`: Generate comments from descriptions (default: true)
- `-validate`: Generate validation methods (default: false)
- `-help`: Show help

## Features

### Supported JTD Forms

- **Type forms**: All RFC 8927 primitive types
  - `boolean` → `bool`
  - `string` → `string`
  - `timestamp` → `time.Time`
  - `float32`, `float64` → `float32`, `float64`
  - `int8`, `int16`, `int32` → `int8`, `int16`, `int32`
  - `uint8`, `uint16`, `uint32` → `uint8`, `uint16`, `uint32`

- **Enum form**: String enumerations → Go constants
- **Elements form**: Arrays → Go slices
- **Properties form**: Objects → Go structs
- **Values form**: Maps → Go maps
- **Discriminator form**: Tagged unions → Go interfaces
- **Ref form**: Schema references
- **Empty form**: Any type → `any`

### Additional Features

- Nullable types using pointers
- Optional struct fields with `omitempty` tags
- Metadata descriptions as Go comments
- Validation method generation
- Proper handling of circular references

## Examples

### Simple Schema

```json
{
  "properties": {
    "name": { "type": "string" },
    "age": { "type": "int32" }
  }
}
```

Generates:

```go
type Root struct {
    Name string `json:"name"`
    Age  int32  `json:"age"`
}
```

### Enum Schema

```json
{
  "definitions": {
    "status": {
      "enum": ["active", "inactive", "pending"]
    }
  }
}
```

Generates:

```go
type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
    StatusPending  Status = "pending"
)
```

### Discriminated Union

```json
{
  "definitions": {
    "shape": {
      "discriminator": "type",
      "mapping": {
        "circle": {
          "properties": {
            "radius": { "type": "float64" }
          }
        },
        "rectangle": {
          "properties": {
            "width": { "type": "float64" },
            "height": { "type": "float64" }
          }
        }
      }
    }
  }
}
```

Generates:

```go
type Shape interface {
    isShape()
    Type() string
}

type ShapeCircle struct {
    Type   string  `json:"type"`
    Radius float64 `json:"radius"`
}

func (ShapeCircle) isShape() {}
func (v ShapeCircle) Type() string { return v.Type }

type ShapeRectangle struct {
    Type   string  `json:"type"`
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

func (ShapeRectangle) isShape() {}
func (v ShapeRectangle) Type() string { return v.Type }
```

## Testing

```bash
go test ./jtd
```

## License

Same as the parent toolbelt project.