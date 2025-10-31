package typemux_test

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux"
)

func TestParseSchema(t *testing.T) {
	idl := `
namespace myapi

type User {
  id: string @required
  email: string @required
  age: int32
}
`

	schema, err := typemux.ParseSchema(idl)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema to be non-nil")
	}

	if schema.Namespace != "myapi" {
		t.Errorf("Expected namespace 'myapi', got %q", schema.Namespace)
	}

	if len(schema.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(schema.Types))
	}

	if schema.Types[0].Name != "User" {
		t.Errorf("Expected type 'User', got %q", schema.Types[0].Name)
	}
}

func TestParseWithAnnotations(t *testing.T) {
	idl := `
namespace myapi

type User {
  id: string
  email: string
}
`

	yamlAnnotations := `
types:
  User:
    fields:
      id:
        annotations:
          - name: "@required"
      email:
        annotations:
          - name: "@required"
`

	schema, err := typemux.ParseWithAnnotations(idl, yamlAnnotations)
	if err != nil {
		t.Fatalf("ParseWithAnnotations failed: %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema to be non-nil")
	}
}

func TestGeneratorFactory(t *testing.T) {
	idl := `
namespace myapi

type User {
  id: string @required
  email: string @required
}
`

	schema, err := typemux.ParseSchema(idl)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	factory := typemux.NewGeneratorFactory()

	// Test GraphQL generation
	graphql, err := factory.Generate("graphql", schema)
	if err != nil {
		t.Fatalf("GraphQL generation failed: %v", err)
	}
	if !strings.Contains(graphql, "type User") {
		t.Error("Expected GraphQL to contain 'type User'")
	}

	// Test Protobuf generation
	proto, err := factory.Generate("protobuf", schema)
	if err != nil {
		t.Fatalf("Protobuf generation failed: %v", err)
	}
	if !strings.Contains(proto, "message User") {
		t.Error("Expected Protobuf to contain 'message User'")
	}

	// Test OpenAPI generation
	openapi, err := factory.Generate("openapi", schema)
	if err != nil {
		t.Fatalf("OpenAPI generation failed: %v", err)
	}
	if !strings.Contains(openapi, "openapi:") {
		t.Error("Expected OpenAPI to contain 'openapi:'")
	}

	// Test Go generation
	goCode, err := factory.Generate("go", schema)
	if err != nil {
		t.Fatalf("Go generation failed: %v", err)
	}
	if !strings.Contains(goCode, "type User struct") {
		t.Error("Expected Go code to contain 'type User struct'")
	}
}

func TestGenerateAll(t *testing.T) {
	idl := `
namespace myapi

type User {
  id: string @required
}
`

	schema, err := typemux.ParseSchema(idl)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	factory := typemux.NewGeneratorFactory()
	outputs, err := factory.GenerateAll(schema)
	if err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	expectedFormats := []string{"graphql", "protobuf", "openapi", "go"}
	for _, format := range expectedFormats {
		if _, ok := outputs[format]; !ok {
			t.Errorf("Expected output for format %q", format)
		}
	}
}

func TestConfigBuilder(t *testing.T) {
	idl := `namespace myapi
type User { id: string @required }`

	config, err := typemux.NewConfigBuilder().
		WithSchema(idl).
		WithFormats("graphql", "protobuf").
		WithOutputDir("./gen").
		Build()

	if err != nil {
		t.Fatalf("Build config failed: %v", err)
	}

	if config.Input.Schema != idl {
		t.Error("Schema not set correctly")
	}

	if len(config.Output.Formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(config.Output.Formats))
	}

	if config.Output.Directory != "./gen" {
		t.Errorf("Expected output dir './gen', got %q", config.Output.Directory)
	}
}

func TestGenerateWithConfig(t *testing.T) {
	idl := `namespace myapi
type User { id: string @required }`

	config, err := typemux.NewConfigBuilder().
		WithSchema(idl).
		WithFormats("graphql", "go").
		Build()

	if err != nil {
		t.Fatalf("Build config failed: %v", err)
	}

	factory := typemux.NewGeneratorFactory()
	outputs, err := factory.GenerateWithConfig(config)
	if err != nil {
		t.Fatalf("GenerateWithConfig failed: %v", err)
	}

	if len(outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(outputs))
	}

	if _, ok := outputs["graphql"]; !ok {
		t.Error("Expected graphql output")
	}

	if _, ok := outputs["go"]; !ok {
		t.Error("Expected go output")
	}
}

func TestImporterFactory(t *testing.T) {
	factory := typemux.NewImporterFactory()

	// Test GraphQL import
	graphqlSchema := `
type User {
  id: ID!
  email: String!
}
`

	typemuxIDL, err := factory.ImportGraphQL(graphqlSchema)
	if err != nil {
		t.Fatalf("ImportGraphQL failed: %v", err)
	}

	if !strings.Contains(typemuxIDL, "type User") {
		t.Error("Expected TypeMUX IDL to contain 'type User'")
	}
}

func TestDiff(t *testing.T) {
	baseIDL := `namespace myapi
type User {
  id: string @required
  email: string @required
}`

	headIDL := `namespace myapi
type User {
  id: string @required
  email: string @required
  name: string
}`

	baseSchema, err := typemux.ParseSchema(baseIDL)
	if err != nil {
		t.Fatalf("ParseSchema base failed: %v", err)
	}

	headSchema, err := typemux.ParseSchema(headIDL)
	if err != nil {
		t.Fatalf("ParseSchema head failed: %v", err)
	}

	result, err := typemux.Diff(baseSchema, headSchema)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !result.HasChanges() {
		t.Error("Expected changes to be detected")
	}

	if result.HasBreakingChanges() {
		t.Error("Expected no breaking changes for field addition")
	}

	report := result.CompactReport()
	if report == "" {
		t.Error("Expected non-empty compact report")
	}
}

func TestGetBuiltinAnnotations(t *testing.T) {
	annotations := typemux.GetBuiltinAnnotations()

	if len(annotations) == 0 {
		t.Error("Expected at least one built-in annotation")
	}

	// Check for common annotations
	found := false
	for _, ann := range annotations {
		if ann.Name == "@required" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected @required annotation to be in built-in list")
	}
}

func TestFilterAnnotationsByScope(t *testing.T) {
	fieldAnnotations := typemux.FilterAnnotationsByScope("field")

	if len(fieldAnnotations) == 0 {
		t.Error("Expected at least one field-scoped annotation")
	}

	// All returned annotations should have "field" in their scope
	for _, ann := range fieldAnnotations {
		hasFieldScope := false
		for _, scope := range ann.Scope {
			if scope == "field" {
				hasFieldScope = true
				break
			}
		}
		if !hasFieldScope {
			t.Errorf("Annotation %s should have field scope", ann.Name)
		}
	}
}

func TestGetAnnotation(t *testing.T) {
	ann, found := typemux.GetAnnotation("@required")
	if !found {
		t.Error("Expected @required annotation to be found")
	}

	if ann.Name != "@required" {
		t.Errorf("Expected annotation name '@required', got %q", ann.Name)
	}

	if ann.Description == "" {
		t.Error("Expected annotation to have a description")
	}
}

// customGenerator is a test implementation of the Generator interface
type customGenerator struct{}

func TestGeneratorRegistration(t *testing.T) {
	// Create a custom generator
	customGen := &customGenerator{}

	factory := typemux.NewGeneratorFactory()

	// Initially shouldn't have custom format
	if factory.HasFormat("custom") {
		t.Error("Should not have custom format initially")
	}

	// Register custom generator
	factory.Register(customGen)

	// Now should have custom format
	if !factory.HasFormat("custom") {
		t.Error("Should have custom format after registration")
	}

	// Unregister
	factory.Unregister("custom")

	if factory.HasFormat("custom") {
		t.Error("Should not have custom format after unregistration")
	}
}

// Implement Generator interface for test
func (g *customGenerator) Generate(schema *typemux.Schema) (string, error) {
	return "custom output", nil
}

func (g *customGenerator) Format() string {
	return "custom"
}

func (g *customGenerator) FileExtension() string {
	return ".custom"
}
