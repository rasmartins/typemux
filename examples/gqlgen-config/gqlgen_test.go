package gqlgen_test

import (
	"os"
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/generator"
	"github.com/rasmartins/typemux/internal/lexer"
	"github.com/rasmartins/typemux/internal/parser"
	"gopkg.in/yaml.v3"
)

func TestGqlgenConfigExample(t *testing.T) {
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

	// Generate gqlgen config
	gen := generator.NewGqlgenConfigGenerator()
	opts := &generator.GqlgenOptions{
		SchemaFiles:   []string{"schema.graphql"},
		ModelsPackage: "github.com/example/userservice/models",
		ExecPackage:   "github.com/example/userservice/generated",
		ResolverType:  "Resolver",
		GenerateStubs: true,
	}

	output, err := gen.Generate(schema, opts)
	if err != nil {
		t.Fatalf("failed to generate gqlgen config: %v", err)
	}

	// Write to file
	if err := os.WriteFile("gqlgen.yml", []byte(output), 0644); err != nil {
		t.Fatalf("failed to write gqlgen.yml: %v", err)
	}

	// Verify the output is valid YAML
	var config generator.GqlgenConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify schema files
	if len(config.Schema) != 1 || config.Schema[0] != "schema.graphql" {
		t.Errorf("expected schema [schema.graphql], got %v", config.Schema)
	}

	// Verify exec config
	if config.Exec.Package != "github.com/example/userservice/generated" {
		t.Errorf("expected exec package github.com/example/userservice/generated, got %q", config.Exec.Package)
	}

	// Verify model config
	if config.Model.Filename != "models_gen.go" {
		t.Errorf("expected model filename models_gen.go, got %q", config.Model.Filename)
	}

	// Verify resolver config
	if config.Resolver.Type != "Resolver" {
		t.Errorf("expected resolver type Resolver, got %q", config.Resolver.Type)
	}

	// Verify autobind
	if len(config.Autobind) != 1 || config.Autobind[0] != "github.com/example/userservice/models" {
		t.Errorf("expected autobind [github.com/example/userservice/models], got %v", config.Autobind)
	}

	// Verify User type mapping
	userModel, ok := config.Models["User"]
	if !ok {
		t.Fatal("expected User model mapping")
	}
	if len(userModel.Model) != 1 || userModel.Model[0] != "github.com/example/userservice/models.User" {
		t.Errorf("expected User model github.com/example/userservice/models.User, got %v", userModel.Model)
	}

	// Verify UserRole enum mapping
	roleModel, ok := config.Models["UserRole"]
	if !ok {
		t.Fatal("expected UserRole model mapping")
	}
	if len(roleModel.Model) != 1 || roleModel.Model[0] != "github.com/example/userservice/models.UserRole" {
		t.Errorf("expected UserRole model github.com/example/userservice/models.UserRole, got %v", roleModel.Model)
	}

	// Verify Timestamp scalar mapping
	timestampModel, ok := config.Models["Timestamp"]
	if !ok {
		t.Fatal("expected Timestamp scalar mapping")
	}
	if len(timestampModel.Model) != 1 || timestampModel.Model[0] != "time.Time" {
		t.Errorf("expected Timestamp mapped to time.Time, got %v", timestampModel.Model)
	}

	// Verify input types are in structs
	createReqStruct, ok := config.Structs["CreateUserRequest"]
	if !ok {
		t.Fatal("expected CreateUserRequest in structs")
	}
	if len(createReqStruct.Model) != 1 || createReqStruct.Model[0] != "github.com/example/userservice/models.CreateUserRequest" {
		t.Errorf("expected CreateUserRequest model github.com/example/userservice/models.CreateUserRequest, got %v", createReqStruct.Model)
	}
}

func TestGqlgenConfigContent(t *testing.T) {
	// Read the generated gqlgen.yml
	content, err := os.ReadFile("gqlgen.yml")
	if err != nil {
		t.Skipf("gqlgen.yml not found (run TestGqlgenConfigExample first): %v", err)
	}

	output := string(content)

	// Verify key sections exist
	if !strings.Contains(output, "schema:") {
		t.Error("expected schema: section")
	}

	if !strings.Contains(output, "exec:") {
		t.Error("expected exec: section")
	}

	if !strings.Contains(output, "model:") {
		t.Error("expected model: section")
	}

	if !strings.Contains(output, "resolver:") {
		t.Error("expected resolver: section")
	}

	if !strings.Contains(output, "autobind:") {
		t.Error("expected autobind: section")
	}

	if !strings.Contains(output, "models:") {
		t.Error("expected models: section")
	}

	// Verify model mappings
	if !strings.Contains(output, "User:") {
		t.Error("expected User model mapping")
	}

	if !strings.Contains(output, "UserRole:") {
		t.Error("expected UserRole model mapping")
	}

	if !strings.Contains(output, "Timestamp:") {
		t.Error("expected Timestamp scalar mapping")
	}

	if !strings.Contains(output, "time.Time") {
		t.Error("expected time.Time in Timestamp mapping")
	}
}

func TestGenerateGraphQLSchema(t *testing.T) {
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

	// Generate GraphQL schema
	gen := generator.NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Write to file
	if err := os.WriteFile("schema.graphql", []byte(output), 0644); err != nil {
		t.Fatalf("failed to write schema.graphql: %v", err)
	}

	// Verify GraphQL types are generated
	if !strings.Contains(output, "type User {") {
		t.Error("expected User type in GraphQL schema")
	}

	if !strings.Contains(output, "enum UserRole {") {
		t.Error("expected UserRole enum in GraphQL schema")
	}

	if !strings.Contains(output, "type Query {") {
		t.Error("expected Query type in GraphQL schema")
	}

	if !strings.Contains(output, "type Mutation {") {
		t.Error("expected Mutation type in GraphQL schema")
	}

	// Verify service methods are in correct operation types
	if !strings.Contains(output, "getUser") {
		t.Error("expected getUser query")
	}

	if !strings.Contains(output, "createUser") {
		t.Error("expected createUser mutation")
	}
}
