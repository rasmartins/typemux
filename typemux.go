// Package typemux provides a public API for using TypeMUX as a Go library.
//
// TypeMUX is an Interface Definition Language (IDL) and code generator that converts
// a single schema definition into multiple output formats: GraphQL schemas, Protocol
// Buffers (proto3), OpenAPI 3.0 specifications, and Go code.
//
// Basic usage:
//
//	// Parse a TypeMUX schema
//	schema, err := typemux.ParseSchema(idlContent)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Generate outputs
//	factory := typemux.NewGeneratorFactory()
//	graphql, _ := factory.Generate("graphql", schema)
//	protobuf, _ := factory.Generate("protobuf", schema)
//	openapi, _ := factory.Generate("openapi", schema)
//	goCode, _ := factory.Generate("go", schema)
//
// With annotations:
//
//	schema, err := typemux.ParseWithAnnotations(idlContent, yamlAnnotation1, yamlAnnotation2)
//
// Using configuration:
//
//	config := typemux.NewConfigBuilder().
//	    WithSchema(schemaContent).
//	    WithFormats("graphql", "protobuf").
//	    WithOutputDir("./generated").
//	    Build()
//
//	factory := typemux.NewGeneratorFactory()
//	outputs, err := factory.GenerateWithConfig(config)
package typemux

import (
	"fmt"

	"github.com/rasmartins/typemux/internal/annotations"
	"github.com/rasmartins/typemux/internal/ast"
	"github.com/rasmartins/typemux/internal/lexer"
	"github.com/rasmartins/typemux/internal/parser"
)

// Schema represents a parsed TypeMUX schema.
// It contains all types, enums, unions, and services defined in the schema.
type Schema = ast.Schema

// ParseSchema parses a TypeMUX IDL schema from a string.
// It returns the parsed schema or an error if parsing fails.
//
// Example:
//
//	idl := `
//	  namespace myapi
//	  type User {
//	    id: string @required
//	    email: string @required
//	  }
//	`
//	schema, err := typemux.ParseSchema(idl)
func ParseSchema(content string) (*Schema, error) {
	l := lexer.New(content)
	p := parser.New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors:\n%s", p.PrintErrors())
	}

	return schema, nil
}

// ParseWithAnnotations parses a TypeMUX IDL schema and merges YAML annotations.
// Multiple annotation strings can be provided; they are merged in order (later
// annotations override earlier ones).
//
// Example:
//
//	schema, err := typemux.ParseWithAnnotations(idl, yamlAnnotations1, yamlAnnotations2)
func ParseWithAnnotations(schemaContent string, yamlAnnotations ...string) (*Schema, error) {
	// Parse the schema
	schema, err := ParseSchema(schemaContent)
	if err != nil {
		return nil, err
	}

	// If no annotations provided, return schema as-is
	if len(yamlAnnotations) == 0 {
		return schema, nil
	}

	// Merge YAML annotations
	mergedAnnotations, err := annotations.MergeYAMLAnnotationsFromContent(yamlAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to merge annotations: %w", err)
	}

	// Validate annotations against schema
	validator := annotations.NewValidator(schema)
	validationErrors := validator.Validate(mergedAnnotations)
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("annotation validation failed:\n%s", validator.FormatErrors())
	}

	// Merge annotations into schema
	merger := annotations.NewMerger(mergedAnnotations)
	merger.Merge(schema)

	return schema, nil
}

// ParseOptions provides options for parsing schemas.
type ParseOptions struct {
	// Schema is the TypeMUX IDL content
	Schema string

	// Annotations are optional YAML annotation strings
	Annotations []string

	// BaseDir is the base directory for resolving relative imports
	// Currently not used but reserved for future import resolution
	BaseDir string
}

// Parse parses a TypeMUX schema with the given options.
//
// Example:
//
//	schema, err := typemux.Parse(typemux.ParseOptions{
//	    Schema:      idlContent,
//	    Annotations: []string{yaml1, yaml2},
//	})
func Parse(opts ParseOptions) (*Schema, error) {
	return ParseWithAnnotations(opts.Schema, opts.Annotations...)
}

// Version returns the TypeMUX version supported by this library.
const Version = "1.0.0"
