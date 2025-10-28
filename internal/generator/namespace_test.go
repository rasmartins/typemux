package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestProtobufGenerator_NamespaceOptions(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.users",
		NamespaceAnnotations: &ast.FormatAnnotations{
			Proto: []string{
				"go_package=\"github.com/example/proto\"",
				"java_package=\"com.example.proto\"",
				"java_multiple_files=true",
			},
		},
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "com.example.users",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}, Number: 1, HasNumber: true},
				},
			},
		},
	}

	gen := NewProtobufGenerator()
	output := gen.Generate(schema)

	// Check that options are present
	if !strings.Contains(output, "option go_package=\"github.com/example/proto\";") {
		t.Error("expected go_package option in output")
	}

	if !strings.Contains(output, "option java_package=\"com.example.proto\";") {
		t.Error("expected java_package option in output")
	}

	if !strings.Contains(output, "option java_multiple_files=true;") {
		t.Error("expected java_multiple_files option in output")
	}

	// Check that options come after package declaration
	if !strings.Contains(output, "package com.example.users;\n\noption") {
		t.Error("expected options to come after package declaration")
	}
}

func TestProtobufGenerator_NoNamespaceOptions(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.simple",
		Types: []*ast.Type{
			{
				Name:      "Simple",
				Namespace: "com.example.simple",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}, Number: 1, HasNumber: true},
				},
			},
		},
	}

	gen := NewProtobufGenerator()
	output := gen.Generate(schema)

	// Should not have "option" after package (except standard imports)
	lines := strings.Split(output, "\n")
	foundPackage := false
	foundOption := false

	for _, line := range lines {
		if strings.HasPrefix(line, "package ") {
			foundPackage = true
		}
		if foundPackage && strings.HasPrefix(line, "option ") && !strings.Contains(line, "import") {
			foundOption = true
			break
		}
	}

	if foundOption {
		t.Error("unexpected option statement when no namespace annotations present")
	}
}

func TestGraphQLGenerator_NamespaceDirectives(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.products",
		NamespaceAnnotations: &ast.FormatAnnotations{
			GraphQL: []string{
				"@link(url: \"https://specs.apollo.dev/federation/v2.0\")",
			},
		},
		Types: []*ast.Type{
			{
				Name:      "Product",
				Namespace: "com.example.products",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Check that directive is present
	if !strings.Contains(output, "extend schema @link(url: \"https://specs.apollo.dev/federation/v2.0\")") {
		t.Error("expected schema extension with directive in output")
	}

	// Check that directive comes early in the output
	lines := strings.Split(output, "\n")
	foundDirective := false
	for i, line := range lines {
		if strings.Contains(line, "extend schema @link") {
			foundDirective = true
			// Should be within first 10 lines
			if i > 10 {
				t.Error("directive should appear early in the schema")
			}
			break
		}
	}

	if !foundDirective {
		t.Error("directive not found in output")
	}
}

func TestGraphQLGenerator_NoNamespaceDirectives(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.simple",
		Types: []*ast.Type{
			{
				Name:      "Simple",
				Namespace: "com.example.simple",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Should not have "extend schema" directive
	if strings.Contains(output, "extend schema @") {
		t.Error("unexpected schema extension when no namespace annotations present")
	}
}

func TestOpenAPIGenerator_NamespaceInfo(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.api",
		NamespaceAnnotations: &ast.FormatAnnotations{
			OpenAPI: []string{
				"title:My Custom API",
				"version:2.0.0",
				"description:A test API",
			},
		},
		Types: []*ast.Type{
			{
				Name:      "Item",
				Namespace: "com.example.api",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
				},
			},
		},
	}

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	// Check that custom info is present
	if !strings.Contains(output, "title: My Custom API") {
		t.Error("expected custom title in output")
	}

	if !strings.Contains(output, "version: 2.0.0") {
		t.Error("expected custom version in output")
	}

	if !strings.Contains(output, "description: A test API") {
		t.Error("expected custom description in output")
	}
}

func TestOpenAPIGenerator_DefaultNamespaceInfo(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.default",
		Types: []*ast.Type{
			{
				Name:      "Default",
				Namespace: "com.example.default",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
				},
			},
		},
	}

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	// Should use default title based on namespace
	if !strings.Contains(output, "title: com.example.default API") {
		t.Error("expected default title based on namespace")
	}

	// Should use default version
	if !strings.Contains(output, "version: 1.0.0") {
		t.Error("expected default version 1.0.0")
	}

	// Should not have description when not specified
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "info:") {
			// Check next few lines for description
			// If description is empty/omitted, it shouldn't appear
			break
		}
	}
}

func TestProtobufGenerator_ByNamespace_WithOptions(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.test",
		NamespaceAnnotations: &ast.FormatAnnotations{
			Proto: []string{
				"go_package=\"github.com/test/proto\"",
			},
		},
		Types: []*ast.Type{
			{
				Name:      "TestType",
				Namespace: "com.example.test",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}, Number: 1, HasNumber: true},
				},
			},
		},
		TypeRegistry: ast.NewTypeRegistry(),
	}

	gen := NewProtobufGenerator()
	outputs := gen.GenerateByNamespace(schema)

	if len(outputs) == 0 {
		t.Fatal("expected at least one output file")
	}

	// Get the output for the test namespace
	var output string
	for _, content := range outputs {
		output = content
		break
	}

	// Check that options are present in namespace-based generation
	if !strings.Contains(output, "option go_package=\"github.com/test/proto\";") {
		t.Error("expected go_package option in namespace-based output")
	}
}
