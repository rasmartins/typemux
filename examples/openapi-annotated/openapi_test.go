package openapi_test

import (
	"os"
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/generator"
	"github.com/rasmartins/typemux/internal/lexer"
	"github.com/rasmartins/typemux/internal/parser"
	"gopkg.in/yaml.v3"
)

func TestOpenAPIAnnotatedExample(t *testing.T) {
	// Read the TypeMUX schema
	content, err := os.ReadFile("user.typemux")
	if err != nil {
		t.Fatalf("failed to read user.typemux: %v", err)
	}

	// Parse the schema
	l := lexer.New(string(content))
	p := parser.New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Generate annotated OpenAPI spec
	annotator := generator.NewOpenAPIAnnotator()
	opts := &generator.OpenAPIAnnotatorOptions{
		ModelsPackage: "github.com/example/userservice/models",
	}

	output, err := annotator.Generate(schema, opts)
	if err != nil {
		t.Fatalf("failed to generate annotated OpenAPI: %v", err)
	}

	// Write to file
	if err := os.WriteFile("openapi-annotated.yaml", []byte(output), 0644); err != nil {
		t.Fatalf("failed to write openapi-annotated.yaml: %v", err)
	}

	// Verify the output is valid YAML
	var spec generator.OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify OpenAPI version
	if spec.OpenAPI != "3.0.0" {
		t.Errorf("expected OpenAPI version 3.0.0, got %q", spec.OpenAPI)
	}

	// Verify info
	if spec.Info.Title != "user API" {
		t.Errorf("expected title 'user API', got %q", spec.Info.Title)
	}

	// Verify User type has x-go-type annotation
	userSchema, ok := spec.Components.Schemas["User"]
	if !ok {
		t.Fatal("expected User schema")
	}

	if userSchema.Extensions == nil {
		t.Fatal("expected User schema to have extensions")
	}

	xGoType, ok := userSchema.Extensions["x-go-type"]
	if !ok {
		t.Fatal("expected User schema to have x-go-type extension")
	}

	expectedType := "github.com/example/userservice/models.User"
	if xGoType != expectedType {
		t.Errorf("expected x-go-type %q, got %v", expectedType, xGoType)
	}

	// Verify x-go-type-import
	xGoTypeImport, ok := userSchema.Extensions["x-go-type-import"]
	if !ok {
		t.Fatal("expected User schema to have x-go-type-import extension")
	}

	importMap, ok := xGoTypeImport.(map[string]interface{})
	if !ok {
		t.Fatalf("expected x-go-type-import to be a map, got %T", xGoTypeImport)
	}

	if importMap["path"] != "github.com/example/userservice/models" {
		t.Errorf("expected import path github.com/example/userservice/models, got %v", importMap["path"])
	}

	// Verify UserRole enum has x-go-type annotation
	roleSchema, ok := spec.Components.Schemas["UserRole"]
	if !ok {
		t.Fatal("expected UserRole schema")
	}

	roleGoType := roleSchema.Extensions["x-go-type"]
	expectedRoleType := "github.com/example/userservice/models.UserRole"
	if roleGoType != expectedRoleType {
		t.Errorf("expected UserRole x-go-type %q, got %v", expectedRoleType, roleGoType)
	}

	// Verify all request types have annotations
	requestTypes := []string{"GetUserRequest", "CreateUserRequest", "UpdateUserRequest", "ListUsersRequest"}
	for _, typeName := range requestTypes {
		schema, ok := spec.Components.Schemas[typeName]
		if !ok {
			t.Errorf("expected %s schema", typeName)
			continue
		}

		if schema.Extensions["x-go-type"] == nil {
			t.Errorf("expected %s to have x-go-type extension", typeName)
		}
	}

	// Verify paths exist
	if len(spec.Paths) == 0 {
		t.Error("expected paths to be generated")
	}

	// Verify specific path exists
	usersPath, ok := spec.Paths["/api/v1/users"]
	if !ok {
		t.Error("expected /api/v1/users path")
	}

	// Verify GET operation (ListUsers)
	if usersPath["get"].OperationID != "ListUsers" {
		t.Errorf("expected ListUsers operation ID, got %q", usersPath["get"].OperationID)
	}

	// Verify POST operation (CreateUser)
	if usersPath["post"].OperationID != "CreateUser" {
		t.Errorf("expected CreateUser operation ID, got %q", usersPath["post"].OperationID)
	}
}

func TestOpenAPIAnnotatedContent(t *testing.T) {
	// Read the generated openapi-annotated.yaml
	content, err := os.ReadFile("openapi-annotated.yaml")
	if err != nil {
		t.Skipf("openapi-annotated.yaml not found (run TestOpenAPIAnnotatedExample first): %v", err)
	}

	output := string(content)

	// Verify key sections exist
	if !strings.Contains(output, "openapi:") {
		t.Error("expected openapi: section")
	}

	if !strings.Contains(output, "info:") {
		t.Error("expected info: section")
	}

	if !strings.Contains(output, "paths:") {
		t.Error("expected paths: section")
	}

	if !strings.Contains(output, "components:") {
		t.Error("expected components: section")
	}

	// Verify x-go-type annotations
	if !strings.Contains(output, "x-go-type:") {
		t.Error("expected x-go-type: annotations")
	}

	if !strings.Contains(output, "x-go-type-import:") {
		t.Error("expected x-go-type-import: annotations")
	}

	// Verify model references
	if !strings.Contains(output, "github.com/example/userservice/models.User") {
		t.Error("expected User model reference")
	}

	if !strings.Contains(output, "github.com/example/userservice/models.UserRole") {
		t.Error("expected UserRole model reference")
	}
}

func TestGenerateJSONFormat(t *testing.T) {
	// Read the TypeMUX schema
	content, err := os.ReadFile("user.typemux")
	if err != nil {
		t.Fatalf("failed to read user.typemux: %v", err)
	}

	// Parse the schema
	l := lexer.New(string(content))
	p := parser.New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Generate annotated OpenAPI spec in JSON format
	annotator := generator.NewOpenAPIAnnotator()
	opts := &generator.OpenAPIAnnotatorOptions{
		ModelsPackage: "github.com/example/userservice/models",
	}

	output, err := annotator.GenerateJSON(schema, opts)
	if err != nil {
		t.Fatalf("failed to generate JSON OpenAPI: %v", err)
	}

	// Write to file
	if err := os.WriteFile("openapi-annotated.json", []byte(output), 0644); err != nil {
		t.Fatalf("failed to write openapi-annotated.json: %v", err)
	}

	// Verify JSON structure
	if !strings.Contains(output, `"x-go-type"`) {
		t.Error("expected x-go-type in JSON output")
	}

	if !strings.Contains(output, `"x-go-type-import"`) {
		t.Error("expected x-go-type-import in JSON output")
	}

	// Verify it's valid JSON
	if !strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Error("expected JSON to start with {")
	}

	if !strings.HasSuffix(strings.TrimSpace(output), "}") {
		t.Error("expected JSON to end with }")
	}
}
