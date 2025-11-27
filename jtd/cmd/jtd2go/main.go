package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/delaneyj/toolbelt/jtd"
)

func main() {
	var (
		input       = flag.String("input", "", "Input JTD schema file (required)")
		output      = flag.String("output", "", "Output Go file (default: stdout)")
		packageName = flag.String("package", "types", "Go package name")
		comments    = flag.Bool("comments", true, "Generate comments from descriptions")
		validate    = flag.Bool("validate", false, "Generate validation methods")
		help        = flag.Bool("help", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "jtd2go - JSON Type Definition to Go code generator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input schema.json -output types.go\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input schema.json -package myapp > types.go\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input schema.json -validate -output types.go\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *input == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Read input file
	data, err := os.ReadFile(*input)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// Parse the schema
	parser := jtd.NewParser()
	_, err = parser.Parse(data)
	if err != nil {
		log.Fatalf("Failed to parse schema: %v", err)
	}

	// Create generator
	opts := jtd.GeneratorOptions{
		PackageName:      *packageName,
		GenerateComments: *comments,
		GenerateValidate: *validate,
	}
	generator := jtd.NewGenerator(parser, opts)

	// Generate code
	code, err := generator.Generate()
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}

	// Generate validation methods if requested
	if *validate {
		validationCode, err := generator.GenerateValidation()
		if err != nil {
			log.Fatalf("Failed to generate validation code: %v", err)
		}
		code = append(code, validationCode...)
	}

	// Write output
	if *output != "" {
		// Ensure output directory exists
		dir := filepath.Dir(*output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		if err := os.WriteFile(*output, code, 0644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Generated %s\n", *output)
	} else {
		// Write to stdout
		fmt.Print(string(code))
	}
}

// Example schema for testing:
/*
{
  "metadata": {
    "description": "User profile schema"
  },
  "definitions": {
    "user": {
      "properties": {
        "id": { "type": "string" },
        "name": { "type": "string" },
        "email": { "type": "string" },
        "age": { "type": "int32" },
        "created": { "type": "timestamp" }
      },
      "optionalProperties": {
        "bio": { "type": "string" },
        "avatar": { "type": "string" }
      }
    },
    "userRole": {
      "enum": ["admin", "user", "guest"]
    },
    "userList": {
      "elements": { "ref": "user" }
    },
    "userMap": {
      "values": { "ref": "user" }
    },
    "notification": {
      "discriminator": "type",
      "mapping": {
        "email": {
          "properties": {
            "to": { "type": "string" },
            "subject": { "type": "string" },
            "body": { "type": "string" }
          }
        },
        "sms": {
          "properties": {
            "to": { "type": "string" },
            "message": { "type": "string" }
          }
        }
      }
    }
  }
}
*/
