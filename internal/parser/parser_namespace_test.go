package parser

import (
	"testing"

	"github.com/rasmartins/typemux/internal/lexer"
)

func TestNamespaceAnnotations_Inline(t *testing.T) {
	input := `
@proto.option(go_package="github.com/example/proto")
@proto.option(java_package="com.example.proto")
namespace com.example.users

type User {
	id: string = 1
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser had errors: %v", p.Errors())
	}

	if schema.Namespace != "com.example.users" {
		t.Errorf("expected namespace 'com.example.users', got '%s'", schema.Namespace)
	}

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set, got nil")
	}

	if len(schema.NamespaceAnnotations.Proto) != 2 {
		t.Errorf("expected 2 proto annotations, got %d", len(schema.NamespaceAnnotations.Proto))
	}

	// Check proto options (stored with spaces around =)
	expectedProto := map[string]bool{
		"go_package = \"github.com/example/proto\"": true,
		"java_package = \"com.example.proto\"":      true,
	}

	for _, annotation := range schema.NamespaceAnnotations.Proto {
		if !expectedProto[annotation] {
			t.Errorf("unexpected proto annotation: %s", annotation)
		}
	}
}

func TestNamespaceAnnotations_GraphQL(t *testing.T) {
	input := `
@graphql.directive("@link(url: \"https://specs.apollo.dev/federation/v2.0\")")
namespace com.example.products

type Product {
	id: string = 1
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser had errors: %v", p.Errors())
	}

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set, got nil")
	}

	if len(schema.NamespaceAnnotations.GraphQL) != 1 {
		t.Fatalf("expected 1 graphql annotation, got %d", len(schema.NamespaceAnnotations.GraphQL))
	}

	expected := "\"@link(url: \\\"https://specs.apollo.dev/federation/v2.0\\\")\""
	if schema.NamespaceAnnotations.GraphQL[0] != expected {
		t.Errorf("expected graphql annotation '%s', got '%s'", expected, schema.NamespaceAnnotations.GraphQL[0])
	}
}

func TestNamespaceAnnotations_Mixed(t *testing.T) {
	input := `
@proto.option(go_package="github.com/example/proto")
@graphql.directive("@link(url: \"test\")")
@openapi.extension("x-api-id: \"my-api\"")
namespace com.example.mixed

type Item {
	id: string = 1
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser had errors: %v", p.Errors())
	}

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set, got nil")
	}

	// Check all three annotation types
	if len(schema.NamespaceAnnotations.Proto) != 1 {
		t.Errorf("expected 1 proto annotation, got %d", len(schema.NamespaceAnnotations.Proto))
	}

	if len(schema.NamespaceAnnotations.GraphQL) != 1 {
		t.Errorf("expected 1 graphql annotation, got %d", len(schema.NamespaceAnnotations.GraphQL))
	}

	if len(schema.NamespaceAnnotations.OpenAPI) != 1 {
		t.Errorf("expected 1 openapi annotation, got %d", len(schema.NamespaceAnnotations.OpenAPI))
	}
}

func TestNamespaceAnnotations_NoAnnotations(t *testing.T) {
	input := `
namespace com.example.plain

type Simple {
	id: string = 1
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser had errors: %v", p.Errors())
	}

	if schema.Namespace != "com.example.plain" {
		t.Errorf("expected namespace 'com.example.plain', got '%s'", schema.Namespace)
	}

	// Should be nil or empty when no annotations are present
	if schema.NamespaceAnnotations != nil {
		if len(schema.NamespaceAnnotations.Proto) > 0 ||
			len(schema.NamespaceAnnotations.GraphQL) > 0 ||
			len(schema.NamespaceAnnotations.OpenAPI) > 0 {
			t.Error("expected no namespace annotations, but some were found")
		}
	}
}

func TestNamespaceAnnotations_MultipleProtoOptions(t *testing.T) {
	input := `
@proto.option(go_package="github.com/example/proto")
@proto.option(java_package="com.example.proto")
@proto.option(java_multiple_files=true)
@proto.option(csharp_namespace="Example.Proto")
namespace com.example.multi

type Data {
	value: string = 1
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser had errors: %v", p.Errors())
	}

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set, got nil")
	}

	if len(schema.NamespaceAnnotations.Proto) != 4 {
		t.Errorf("expected 4 proto annotations, got %d", len(schema.NamespaceAnnotations.Proto))
	}

	// Verify all options are present
	optionsFound := make(map[string]bool)
	for _, annotation := range schema.NamespaceAnnotations.Proto {
		optionsFound[annotation] = true
	}

	expectedOptions := []string{
		"go_package = \"github.com/example/proto\"",
		"java_package = \"com.example.proto\"",
		"java_multiple_files = true",
		"csharp_namespace = \"Example.Proto\"",
	}

	for _, expected := range expectedOptions {
		if !optionsFound[expected] {
			t.Errorf("expected proto option '%s' not found", expected)
		}
	}
}

func TestNamespaceAnnotations_Go(t *testing.T) {
	input := `
@go.package("mypackage")
namespace com.example.api

type User {
	id: string
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser had errors: %v", p.Errors())
	}

	if schema.Namespace != "com.example.api" {
		t.Errorf("expected namespace 'com.example.api', got '%s'", schema.Namespace)
	}

	if schema.NamespaceAnnotations == nil {
		t.Fatal("expected NamespaceAnnotations to be set, got nil")
	}

	if len(schema.NamespaceAnnotations.Go) != 1 {
		t.Fatalf("expected 1 go annotation, got %d", len(schema.NamespaceAnnotations.Go))
	}

	expected := `package = "mypackage"`
	if schema.NamespaceAnnotations.Go[0] != expected {
		t.Errorf("expected go annotation '%s', got '%s'", expected, schema.NamespaceAnnotations.Go[0])
	}
}
