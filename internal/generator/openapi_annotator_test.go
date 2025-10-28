package generator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
	"gopkg.in/yaml.v3"
)

func TestOpenAPIAnnotator_Basic(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "user",
		Version:   "1.0.0",
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
					{Name: "ADMIN"},
					{Name: "USER"},
					{Name: "GUEST"},
				},
			},
		},
	}

	annotator := NewOpenAPIAnnotator()
	opts := &OpenAPIAnnotatorOptions{
		ModelsPackage: "github.com/example/project/models",
	}

	output, err := annotator.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the YAML to verify structure
	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
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

	expectedType := "github.com/example/project/models.User"
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

	if importMap["path"] != "github.com/example/project/models" {
		t.Errorf("expected import path github.com/example/project/models, got %v", importMap["path"])
	}

	// Verify UserRole enum has x-go-type annotation
	roleSchema, ok := spec.Components.Schemas["UserRole"]
	if !ok {
		t.Fatal("expected UserRole schema")
	}

	if roleSchema.Extensions == nil {
		t.Fatal("expected UserRole schema to have extensions")
	}

	roleGoType, ok := roleSchema.Extensions["x-go-type"]
	if !ok {
		t.Fatal("expected UserRole schema to have x-go-type extension")
	}

	expectedRoleType := "github.com/example/project/models.UserRole"
	if roleGoType != expectedRoleType {
		t.Errorf("expected x-go-type %q, got %v", expectedRoleType, roleGoType)
	}
}

func TestOpenAPIAnnotator_WithService(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "user",
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
				},
			},
			{
				Name: "GetUserRequest",
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
						Name:         "GetUser",
						InputType:    "GetUserRequest",
						OutputType:   "User",
						HTTPMethod:   "GET",
						PathTemplate: "/users/{id}",
					},
				},
			},
		},
	}

	annotator := NewOpenAPIAnnotator()
	opts := &OpenAPIAnnotatorOptions{
		ModelsPackage: "github.com/example/project/models",
	}

	output, err := annotator.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify both types have annotations
	userSchema := spec.Components.Schemas["User"]
	if userSchema.Extensions["x-go-type"] == nil {
		t.Error("expected User to have x-go-type")
	}

	requestSchema := spec.Components.Schemas["GetUserRequest"]
	if requestSchema.Extensions["x-go-type"] == nil {
		t.Error("expected GetUserRequest to have x-go-type")
	}

	// Verify paths exist (from base generator)
	if len(spec.Paths) == 0 {
		t.Error("expected paths to be generated")
	}
}

func TestOpenAPIAnnotator_WithUnions(t *testing.T) {
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

	annotator := NewOpenAPIAnnotator()
	opts := &OpenAPIAnnotatorOptions{
		ModelsPackage: "github.com/example/project/models",
	}

	output, err := annotator.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify Pet union has x-go-type
	petSchema, ok := spec.Components.Schemas["Pet"]
	if !ok {
		t.Fatal("expected Pet schema")
	}

	if petSchema.Extensions["x-go-type"] == nil {
		t.Error("expected Pet union to have x-go-type")
	}

	expectedPetType := "github.com/example/project/models.Pet"
	if petSchema.Extensions["x-go-type"] != expectedPetType {
		t.Errorf("expected Pet x-go-type %q, got %v", expectedPetType, petSchema.Extensions["x-go-type"])
	}

	// Verify Dog and Cat also have annotations
	dogSchema := spec.Components.Schemas["Dog"]
	if dogSchema.Extensions["x-go-type"] == nil {
		t.Error("expected Dog to have x-go-type")
	}

	catSchema := spec.Components.Schemas["Cat"]
	if catSchema.Extensions["x-go-type"] == nil {
		t.Error("expected Cat to have x-go-type")
	}
}

func TestOpenAPIAnnotator_DefaultOptions(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
				},
			},
		},
	}

	annotator := NewOpenAPIAnnotator()

	// Call with nil options (should use defaults)
	output, err := annotator.Generate(schema, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify default package is used
	userSchema := spec.Components.Schemas["User"]
	xGoType := userSchema.Extensions["x-go-type"]

	expectedType := "models.User"
	if xGoType != expectedType {
		t.Errorf("expected default x-go-type %q, got %v", expectedType, xGoType)
	}
}

func TestOpenAPIAnnotator_ComplexTypes(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
					{Name: "tags", Type: &ast.FieldType{Name: "string", IsArray: true}, Number: 2},
					{
						Name: "metadata",
						Type: &ast.FieldType{
							Name:         "map",
							IsMap:        true,
							MapKey:       "string",
							MapValueType: &ast.FieldType{Name: "string"},
						},
						Number: 3,
					},
				},
			},
			{
				Name: "Post",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
					{Name: "author", Type: &ast.FieldType{Name: "User"}, Number: 2},
					{Name: "comments", Type: &ast.FieldType{Name: "Comment", IsArray: true}, Number: 3},
				},
			},
			{
				Name: "Comment",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
					{Name: "text", Type: &ast.FieldType{Name: "string"}, Number: 2},
				},
			},
		},
	}

	annotator := NewOpenAPIAnnotator()
	opts := &OpenAPIAnnotatorOptions{
		ModelsPackage: "github.com/example/blog/models",
	}

	output, err := annotator.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify all types have annotations
	for _, typeName := range []string{"User", "Post", "Comment"} {
		schema := spec.Components.Schemas[typeName]
		if schema.Extensions["x-go-type"] == nil {
			t.Errorf("expected %s to have x-go-type", typeName)
		}

		expectedType := fmt.Sprintf("github.com/example/blog/models.%s", typeName)
		if schema.Extensions["x-go-type"] != expectedType {
			t.Errorf("expected %s x-go-type %q, got %v", typeName, expectedType, schema.Extensions["x-go-type"])
		}
	}
}

func TestOpenAPIAnnotator_YAMLOutput(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
				},
			},
		},
	}

	annotator := NewOpenAPIAnnotator()
	opts := &OpenAPIAnnotatorOptions{
		ModelsPackage: "models",
	}

	output, err := annotator.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify YAML structure
	if !strings.Contains(output, "x-go-type:") {
		t.Error("expected x-go-type: in YAML output")
	}

	if !strings.Contains(output, "x-go-type-import:") {
		t.Error("expected x-go-type-import: in YAML output")
	}

	if !strings.Contains(output, "path:") {
		t.Error("expected path: field in x-go-type-import")
	}

	// Verify no tabs (YAML should use spaces)
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "\t") {
			t.Errorf("line %d contains tabs: %q", i+1, line)
		}
	}
}

func TestOpenAPIAnnotator_JSONOutput(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}, Number: 1},
				},
			},
		},
	}

	annotator := NewOpenAPIAnnotator()
	opts := &OpenAPIAnnotatorOptions{
		ModelsPackage: "github.com/example/project/models",
	}

	output, err := annotator.GenerateJSON(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify JSON structure
	if !strings.Contains(output, `"x-go-type"`) {
		t.Error("expected x-go-type in JSON output")
	}

	if !strings.Contains(output, `"x-go-type-import"`) {
		t.Error("expected x-go-type-import in JSON output")
	}

	if !strings.Contains(output, `"path"`) {
		t.Error("expected path field in x-go-type-import")
	}

	// Verify it's valid JSON structure
	if !strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Error("expected JSON output to start with {")
	}

	if !strings.HasSuffix(strings.TrimSpace(output), "}") {
		t.Error("expected JSON output to end with }")
	}
}

func TestOpenAPIAnnotator_MultipleEnums(t *testing.T) {
	schema := &ast.Schema{
		Enums: []*ast.Enum{
			{
				Name: "UserRole",
				Values: []*ast.EnumValue{
					{Name: "ADMIN"},
					{Name: "USER"},
				},
			},
			{
				Name: "Status",
				Values: []*ast.EnumValue{
					{Name: "ACTIVE"},
					{Name: "INACTIVE"},
				},
			},
		},
	}

	annotator := NewOpenAPIAnnotator()
	opts := &OpenAPIAnnotatorOptions{
		ModelsPackage: "enums",
	}

	output, err := annotator.Generate(schema, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	// Verify both enums have annotations
	for _, enumName := range []string{"UserRole", "Status"} {
		schema := spec.Components.Schemas[enumName]
		if schema.Extensions["x-go-type"] == nil {
			t.Errorf("expected %s to have x-go-type", enumName)
		}

		expectedType := fmt.Sprintf("enums.%s", enumName)
		if schema.Extensions["x-go-type"] != expectedType {
			t.Errorf("expected %s x-go-type %q, got %v", enumName, expectedType, schema.Extensions["x-go-type"])
		}
	}
}
