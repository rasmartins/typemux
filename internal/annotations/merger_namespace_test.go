package annotations

import (
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestMerger_NamespaceAnnotations_Proto(t *testing.T) {
	yamlAnnotations := &YAMLAnnotations{
		Namespaces: map[string]*NamespaceAnnotations{
			"com.example.users": {
				Proto: &NamespaceProtoAnnotations{
					Options: map[string]string{
						"go_package":          "github.com/example/proto",
						"java_package":        "com.example.proto",
						"java_multiple_files": "true",
					},
				},
			},
		},
	}

	schema := &ast.Schema{
		Namespace: "com.example.users",
		Types:     []*ast.Type{},
	}

	merger := NewMerger(yamlAnnotations)
	merger.Merge(schema)

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set")
	}

	if len(schema.NamespaceAnnotations.Proto) != 3 {
		t.Errorf("expected 3 proto options, got %d", len(schema.NamespaceAnnotations.Proto))
	}

	// Check that all options are formatted correctly
	optionsFound := make(map[string]bool)
	for _, opt := range schema.NamespaceAnnotations.Proto {
		optionsFound[opt] = true
	}

	expectedOptions := map[string]bool{
		"go_package=\"github.com/example/proto\"": true,
		"java_package=\"com.example.proto\"":      true,
		"java_multiple_files=\"true\"":            true,
	}

	for expected := range expectedOptions {
		if !optionsFound[expected] {
			t.Errorf("expected option '%s' not found", expected)
		}
	}
}

func TestMerger_NamespaceAnnotations_GraphQL(t *testing.T) {
	yamlAnnotations := &YAMLAnnotations{
		Namespaces: map[string]*NamespaceAnnotations{
			"com.example.products": {
				GraphQL: &FormatSpecificAnnotations{
					Directive: "@link(url: \"https://specs.apollo.dev/federation/v2.0\")",
				},
			},
		},
	}

	schema := &ast.Schema{
		Namespace: "com.example.products",
		Types:     []*ast.Type{},
	}

	merger := NewMerger(yamlAnnotations)
	merger.Merge(schema)

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set")
	}

	if len(schema.NamespaceAnnotations.GraphQL) != 1 {
		t.Fatalf("expected 1 graphql directive, got %d", len(schema.NamespaceAnnotations.GraphQL))
	}

	expected := "@link(url: \"https://specs.apollo.dev/federation/v2.0\")"
	if schema.NamespaceAnnotations.GraphQL[0] != expected {
		t.Errorf("expected directive '%s', got '%s'", expected, schema.NamespaceAnnotations.GraphQL[0])
	}
}

func TestMerger_NamespaceAnnotations_OpenAPI(t *testing.T) {
	yamlAnnotations := &YAMLAnnotations{
		Namespaces: map[string]*NamespaceAnnotations{
			"com.example.api": {
				OpenAPI: &NamespaceOpenAPIAnnotations{
					Info: map[string]string{
						"title":       "My API",
						"version":     "2.0.0",
						"description": "Test API",
					},
					Extensions: map[string]string{
						"x-api-id":      "my-api",
						"x-internal-id": "prod-001",
					},
				},
			},
		},
	}

	schema := &ast.Schema{
		Namespace: "com.example.api",
		Types:     []*ast.Type{},
	}

	merger := NewMerger(yamlAnnotations)
	merger.Merge(schema)

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set")
	}

	// Should have 3 info items + 2 extensions = 5 total
	if len(schema.NamespaceAnnotations.OpenAPI) != 5 {
		t.Errorf("expected 5 openapi annotations, got %d", len(schema.NamespaceAnnotations.OpenAPI))
	}

	// Check that info and extensions are present
	annotationsFound := make(map[string]bool)
	for _, annotation := range schema.NamespaceAnnotations.OpenAPI {
		annotationsFound[annotation] = true
	}

	expectedAnnotations := []string{
		"title:My API",
		"version:2.0.0",
		"description:Test API",
		"x-api-id:my-api",
		"x-internal-id:prod-001",
	}

	for _, expected := range expectedAnnotations {
		if !annotationsFound[expected] {
			t.Errorf("expected annotation '%s' not found", expected)
		}
	}
}

func TestMerger_NamespaceAnnotations_AllFormats(t *testing.T) {
	yamlAnnotations := &YAMLAnnotations{
		Namespaces: map[string]*NamespaceAnnotations{
			"com.example.mixed": {
				Proto: &NamespaceProtoAnnotations{
					Options: map[string]string{
						"go_package": "github.com/example/proto",
					},
				},
				GraphQL: &FormatSpecificAnnotations{
					Directive: "@link(url: \"test\")",
				},
				OpenAPI: &NamespaceOpenAPIAnnotations{
					Info: map[string]string{
						"title": "Mixed API",
					},
				},
			},
		},
	}

	schema := &ast.Schema{
		Namespace: "com.example.mixed",
		Types:     []*ast.Type{},
	}

	merger := NewMerger(yamlAnnotations)
	merger.Merge(schema)

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set")
	}

	// Check all three formats
	if len(schema.NamespaceAnnotations.Proto) != 1 {
		t.Errorf("expected 1 proto option, got %d", len(schema.NamespaceAnnotations.Proto))
	}

	if len(schema.NamespaceAnnotations.GraphQL) != 1 {
		t.Errorf("expected 1 graphql directive, got %d", len(schema.NamespaceAnnotations.GraphQL))
	}

	if len(schema.NamespaceAnnotations.OpenAPI) != 1 {
		t.Errorf("expected 1 openapi annotation, got %d", len(schema.NamespaceAnnotations.OpenAPI))
	}
}

func TestMerger_NamespaceAnnotations_NoMatch(t *testing.T) {
	yamlAnnotations := &YAMLAnnotations{
		Namespaces: map[string]*NamespaceAnnotations{
			"com.example.other": {
				Proto: &NamespaceProtoAnnotations{
					Options: map[string]string{
						"go_package": "github.com/other/proto",
					},
				},
			},
		},
	}

	schema := &ast.Schema{
		Namespace: "com.example.users", // Different namespace
		Types:     []*ast.Type{},
	}

	merger := NewMerger(yamlAnnotations)
	merger.Merge(schema)

	// Should not have namespace annotations since namespace doesn't match
	if schema.NamespaceAnnotations != nil {
		if len(schema.NamespaceAnnotations.Proto) > 0 ||
			len(schema.NamespaceAnnotations.GraphQL) > 0 ||
			len(schema.NamespaceAnnotations.OpenAPI) > 0 {
			t.Error("expected no namespace annotations for non-matching namespace")
		}
	}
}

func TestMerger_NamespaceAnnotations_EmptyOptions(t *testing.T) {
	yamlAnnotations := &YAMLAnnotations{
		Namespaces: map[string]*NamespaceAnnotations{
			"com.example.empty": {
				Proto: &NamespaceProtoAnnotations{
					Options: map[string]string{},
				},
			},
		},
	}

	schema := &ast.Schema{
		Namespace: "com.example.empty",
		Types:     []*ast.Type{},
	}

	merger := NewMerger(yamlAnnotations)
	merger.Merge(schema)

	// Should initialize but have no options
	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be initialized")
	}

	if len(schema.NamespaceAnnotations.Proto) != 0 {
		t.Error("expected no proto options when options map is empty")
	}
}

func TestMerger_NamespaceAnnotations_WithTypeAnnotations(t *testing.T) {
	yamlAnnotations := &YAMLAnnotations{
		Namespaces: map[string]*NamespaceAnnotations{
			"com.example.users": {
				Proto: &NamespaceProtoAnnotations{
					Options: map[string]string{
						"go_package": "github.com/example/proto",
					},
				},
			},
		},
		Types: map[string]*TypeAnnotations{
			"User": {
				Fields: map[string]*FieldAnnotations{
					"id": {
						Required: true,
					},
				},
			},
		},
	}

	schema := &ast.Schema{
		Namespace: "com.example.users",
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "com.example.users",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}},
				},
			},
		},
	}

	merger := NewMerger(yamlAnnotations)
	merger.Merge(schema)

	// Check namespace annotations
	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set")
	}

	if len(schema.NamespaceAnnotations.Proto) != 1 {
		t.Error("expected 1 proto option")
	}

	// Check type field annotations still work
	if !schema.Types[0].Fields[0].Required {
		t.Error("expected field 'id' to be required")
	}
}
