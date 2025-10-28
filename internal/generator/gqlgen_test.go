package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
	"gopkg.in/yaml.v3"
)

func TestGqlgenConfigGenerator_Basic(t *testing.T) {
	schema := &ast.Schema{
		Version: "1.0.0",
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
					{Name: "name", Type: &ast.FieldType{Name: "string"}, Number: 2},
					{Name: "email", Type: &ast.FieldType{Name: "string"}, Number: 3},
				},
			},
		},
		Enums: []*ast.Enum{
			{
				Name: "UserRole",
				Values: []*ast.EnumValue{
					{Name: "ADMIN", Number: 0},
					{Name: "USER", Number: 1},
					{Name: "GUEST", Number: 2},
				},
			},
		},
		Services: []*ast.Service{
			{
				Name: "UserService",
				Methods: []*ast.Method{
					{
						Name:       "GetUser",
						InputType:  "GetUserRequest",
						OutputType: "User",
					},
				},
			},
		},
	}

	gen := NewGqlgenConfigGenerator()
	opts := &GqlgenOptions{
		SchemaFiles:   []string{"schema.graphql"},
		ModelsPackage: "github.com/example/project/models",
		ExecPackage:   "generated",
	}

	output, err := gen.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify YAML is valid
	var config GqlgenConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify schema files
	if len(config.Schema) != 1 || config.Schema[0] != "schema.graphql" {
		t.Errorf("expected schema files [schema.graphql], got %v", config.Schema)
	}

	// Verify exec config
	if config.Exec.Filename != "generated/exec.go" {
		t.Errorf("expected exec filename %q, got %q", "generated/exec.go", config.Exec.Filename)
	}

	if config.Exec.Package != "generated" {
		t.Errorf("expected exec package %q, got %q", "generated", config.Exec.Package)
	}

	// Verify autobind
	if len(config.Autobind) != 1 || config.Autobind[0] != "github.com/example/project/models" {
		t.Errorf("expected autobind [github.com/example/project/models], got %v", config.Autobind)
	}

	// Verify User model mapping
	userModel, ok := config.Models["User"]
	if !ok {
		t.Fatal("expected User model mapping")
	}

	expectedUserModel := "github.com/example/project/models.User"
	if len(userModel.Model) != 1 || userModel.Model[0] != expectedUserModel {
		t.Errorf("expected User model %q, got %v", expectedUserModel, userModel.Model)
	}

	// Verify UserRole enum mapping
	roleModel, ok := config.Models["UserRole"]
	if !ok {
		t.Fatal("expected UserRole model mapping")
	}

	expectedRoleModel := "github.com/example/project/models.UserRole"
	if len(roleModel.Model) != 1 || roleModel.Model[0] != expectedRoleModel {
		t.Errorf("expected UserRole model %q, got %v", expectedRoleModel, roleModel.Model)
	}

	// Verify Timestamp scalar mapping
	timestampModel, ok := config.Models["Timestamp"]
	if !ok {
		t.Fatal("expected Timestamp scalar mapping")
	}

	if len(timestampModel.Model) != 1 || timestampModel.Model[0] != "time.Time" {
		t.Errorf("expected Timestamp mapped to time.Time, got %v", timestampModel.Model)
	}

	// Verify Bytes scalar mapping
	bytesModel, ok := config.Models["Bytes"]
	if !ok {
		t.Fatal("expected Bytes scalar mapping")
	}

	if len(bytesModel.Model) != 1 || bytesModel.Model[0] != "[]byte" {
		t.Errorf("expected Bytes mapped to []byte, got %v", bytesModel.Model)
	}
}

func TestGqlgenConfigGenerator_InputTypes(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
				},
			},
			{
				Name: "CreateUserRequest",
				Fields: []*ast.Field{
					{Name: "name", Type: &ast.FieldType{Name: "string"}, Number: 1},
					{Name: "email", Type: &ast.FieldType{Name: "string"}, Number: 2},
				},
			},
		},
		Services: []*ast.Service{
			{
				Name: "UserService",
				Methods: []*ast.Method{
					{
						Name:       "CreateUser",
						InputType:  "CreateUserRequest",
						OutputType: "User",
					},
				},
			},
		},
	}

	gen := NewGqlgenConfigGenerator()
	opts := &GqlgenOptions{
		SchemaFiles:   []string{"schema.graphql"},
		ModelsPackage: "github.com/example/project/models",
	}

	output, err := gen.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config GqlgenConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify CreateUserRequest is in structs
	createReqStruct, ok := config.Structs["CreateUserRequest"]
	if !ok {
		t.Fatal("expected CreateUserRequest in structs mapping")
	}

	expectedModel := "github.com/example/project/models.CreateUserRequest"
	if len(createReqStruct.Model) != 1 || createReqStruct.Model[0] != expectedModel {
		t.Errorf("expected CreateUserRequest model %q, got %v", expectedModel, createReqStruct.Model)
	}
}

func TestGqlgenConfigGenerator_WithResolver(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
				},
			},
		},
		Services: []*ast.Service{
			{
				Name: "UserService",
				Methods: []*ast.Method{
					{
						Name:       "GetUser",
						InputType:  "string",
						OutputType: "User",
					},
				},
			},
		},
	}

	gen := NewGqlgenConfigGenerator()
	opts := &GqlgenOptions{
		SchemaFiles:   []string{"schema.graphql"},
		ModelsPackage: "github.com/example/project/models",
		ResolverType:  "Resolver",
		GenerateStubs: true,
	}

	output, err := gen.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config GqlgenConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify resolver config
	if config.Resolver.Filename != "resolver.go" {
		t.Errorf("expected resolver filename %q, got %q", "resolver.go", config.Resolver.Filename)
	}

	if config.Resolver.Type != "Resolver" {
		t.Errorf("expected resolver type %q, got %q", "Resolver", config.Resolver.Type)
	}
}

func TestGqlgenConfigGenerator_MultipleSchemas(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{Name: "User", Fields: []*ast.Field{{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1}}},
		},
	}

	gen := NewGqlgenConfigGenerator()
	opts := &GqlgenOptions{
		SchemaFiles:   []string{"schema1.graphql", "schema2.graphql"},
		ModelsPackage: "models",
	}

	output, err := gen.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config GqlgenConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify multiple schema files
	if len(config.Schema) != 2 {
		t.Fatalf("expected 2 schema files, got %d", len(config.Schema))
	}

	if config.Schema[0] != "schema1.graphql" || config.Schema[1] != "schema2.graphql" {
		t.Errorf("unexpected schema files: %v", config.Schema)
	}
}

func TestGqlgenConfigGenerator_UnionTypes(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "Dog",
				Fields: []*ast.Field{
					{Name: "breed", Type: &ast.FieldType{Name: "string"}, Number: 1},
				},
			},
			{
				Name: "Cat",
				Fields: []*ast.Field{
					{Name: "meow", Type: &ast.FieldType{Name: "bool"}, Number: 1},
				},
			},
		},
		Unions: []*ast.Union{
			{
				Name:    "Pet",
				Options: []string{"Dog", "Cat"},
			},
		},
	}

	gen := NewGqlgenConfigGenerator()
	opts := &GqlgenOptions{
		SchemaFiles:   []string{"schema.graphql"},
		ModelsPackage: "github.com/example/project/models",
	}

	output, err := gen.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config GqlgenConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify Dog and Cat are mapped
	if _, ok := config.Models["Dog"]; !ok {
		t.Error("expected Dog model mapping")
	}

	if _, ok := config.Models["Cat"]; !ok {
		t.Error("expected Cat model mapping")
	}

	// Union types (Pet) should NOT be in models - let gqlgen generate them
	if _, ok := config.Models["Pet"]; ok {
		t.Error("union type Pet should not be in model mappings")
	}
}

func TestGqlgenConfigGenerator_DefaultOptions(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{Name: "User", Fields: []*ast.Field{{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1}}},
		},
	}

	gen := NewGqlgenConfigGenerator()

	// Call Generate with nil options (should use defaults)
	output, err := gen.Generate(schema, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var config GqlgenConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify defaults
	if len(config.Schema) != 1 || config.Schema[0] != "schema.graphql" {
		t.Errorf("expected default schema [schema.graphql], got %v", config.Schema)
	}

	if config.Exec.Package != "generated" {
		t.Errorf("expected default exec package %q, got %q", "generated", config.Exec.Package)
	}

	if len(config.Autobind) != 1 || config.Autobind[0] != "models" {
		t.Errorf("expected default autobind [models], got %v", config.Autobind)
	}
}

func TestGqlgenConfigGenerator_YAMLFormat(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{Name: "User", Fields: []*ast.Field{{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1}}},
		},
	}

	gen := NewGqlgenConfigGenerator()
	opts := &GqlgenOptions{
		SchemaFiles:   []string{"schema.graphql"},
		ModelsPackage: "models",
	}

	output, err := gen.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify YAML structure
	if !strings.Contains(output, "schema:") {
		t.Error("expected schema: field in YAML")
	}

	if !strings.Contains(output, "exec:") {
		t.Error("expected exec: field in YAML")
	}

	if !strings.Contains(output, "model:") {
		t.Error("expected model: field in YAML")
	}

	if !strings.Contains(output, "autobind:") {
		t.Error("expected autobind: field in YAML")
	}

	if !strings.Contains(output, "models:") {
		t.Error("expected models: field in YAML")
	}

	// Verify proper YAML indentation
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "\t") {
			t.Error("YAML should use spaces, not tabs")
			break
		}
	}
}
