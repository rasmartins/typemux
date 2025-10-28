package annotations

import (
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func createTestSchema() *ast.Schema {
	return &ast.Schema{
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "com.example.api",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
					{Name: "username", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
					{Name: "email", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
				},
			},
			{
				Name:      "Product",
				Namespace: "com.example.api",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
					{Name: "name", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
				},
			},
			{
				Name:      "User",
				Namespace: "com.example.admin",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
					{Name: "adminLevel", Type: &ast.FieldType{Name: "int32", IsBuiltin: true}},
				},
			},
		},
		Enums: []*ast.Enum{
			{
				Name:      "UserStatus",
				Namespace: "com.example.api",
				Values: []*ast.EnumValue{
					{Name: "ACTIVE"},
					{Name: "INACTIVE"},
				},
			},
		},
		Unions: []*ast.Union{
			{
				Name:      "SearchResult",
				Namespace: "com.example.api",
				Options:   []string{"User", "Product"},
			},
		},
		Services: []*ast.Service{
			{
				Name:      "UserService",
				Namespace: "com.example.api",
				Methods: []*ast.Method{
					{Name: "GetUser", InputType: "GetUserRequest", OutputType: "GetUserResponse"},
					{Name: "CreateUser", InputType: "CreateUserRequest", OutputType: "CreateUserResponse"},
				},
			},
		},
	}
}

func TestValidator_ValidAnnotations(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Proto:   &FormatSpecificAnnotations{Name: "UserV2"},
				GraphQL: &FormatSpecificAnnotations{Name: "UserAccount"},
				Fields: map[string]*FieldAnnotations{
					"email": {
						Required: true,
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}
}

func TestValidator_NonExistentType(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"NonExistentType": {
				Proto: &FormatSpecificAnnotations{Name: "SomeName"},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Path != "types.NonExistentType" {
		t.Errorf("Expected error path 'types.NonExistentType', got '%s'", errors[0].Path)
	}
}

func TestValidator_NonExistentField(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Fields: map[string]*FieldAnnotations{
					"nonexistentField": {
						Required: true,
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Path != "types.User.fields.nonexistentField" {
		t.Errorf("Expected error path 'types.User.fields.nonexistentField', got '%s'", errors[0].Path)
	}
}

func TestValidator_ConflictingFieldAnnotations(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Fields: map[string]*FieldAnnotations{
					"email": {
						Exclude: []string{"proto"},
						Only:    []string{"graphql"},
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Message != "cannot specify both 'exclude' and 'only' annotations" {
		t.Errorf("Unexpected error message: %s", errors[0].Message)
	}
}

func TestValidator_InvalidGeneratorName(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Fields: map[string]*FieldAnnotations{
					"email": {
						Exclude: []string{"invalid_generator"},
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Message != "invalid generator name in exclude: 'invalid_generator'" {
		t.Errorf("Unexpected error message: %s", errors[0].Message)
	}
}

func TestValidator_InvalidHTTPMethod(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Services: map[string]*ServiceAnnotations{
			"UserService": {
				Methods: map[string]*MethodAnnotations{
					"GetUser": {
						HTTP: "INVALID",
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Message != "invalid HTTP method: 'INVALID'" {
		t.Errorf("Unexpected error message: %s", errors[0].Message)
	}
}

func TestValidator_InvalidGraphQLOperationType(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Services: map[string]*ServiceAnnotations{
			"UserService": {
				Methods: map[string]*MethodAnnotations{
					"GetUser": {
						GraphQL: "invalid",
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Message != "invalid GraphQL operation type: 'invalid'" {
		t.Errorf("Unexpected error message: %s", errors[0].Message)
	}
}

func TestValidator_InvalidStatusCode(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Services: map[string]*ServiceAnnotations{
			"UserService": {
				Methods: map[string]*MethodAnnotations{
					"GetUser": {
						Success: []int{999},
						Errors:  []int{50},
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 2 {
		t.Fatalf("Expected 2 validation errors, got %d", len(errors))
	}
}

func TestValidator_QualifiedTypeName(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	// Use qualified name to target specific User type
	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"com.example.api.User": {
				Proto: &FormatSpecificAnnotations{Name: "ApiUser"},
				Fields: map[string]*FieldAnnotations{
					"email": {
						Required: true,
					},
				},
			},
			"com.example.admin.User": {
				Proto: &FormatSpecificAnnotations{Name: "AdminUser"},
				Fields: map[string]*FieldAnnotations{
					"adminLevel": {
						Required: true,
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}
}

func TestValidator_QualifiedTypeName_NonExistent(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"com.example.wrong.User": {
				Proto: &FormatSpecificAnnotations{Name: "WrongUser"},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Path != "types.com.example.wrong.User" {
		t.Errorf("Expected error path 'types.com.example.wrong.User', got '%s'", errors[0].Path)
	}
}

func TestValidator_QualifiedEnumName(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Enums: map[string]*EnumAnnotations{
			"com.example.api.UserStatus": {
				Proto: &FormatSpecificAnnotations{Name: "ApiUserStatus"},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}
}

func TestValidator_QualifiedServiceName(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Services: map[string]*ServiceAnnotations{
			"com.example.api.UserService": {
				Methods: map[string]*MethodAnnotations{
					"GetUser": {
						HTTP: "GET",
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}
}

func TestValidator_NonExistentMethod(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Services: map[string]*ServiceAnnotations{
			"UserService": {
				Methods: map[string]*MethodAnnotations{
					"NonExistentMethod": {
						HTTP: "GET",
					},
				},
			},
		},
	}

	errors := validator.Validate(annotations)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(errors))
	}

	if errors[0].Path != "services.UserService.methods.NonExistentMethod" {
		t.Errorf("Expected error path 'services.UserService.methods.NonExistentMethod', got '%s'", errors[0].Path)
	}
}

func TestValidator_FormatErrors(t *testing.T) {
	schema := createTestSchema()
	validator := NewValidator(schema)

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"NonExistent": {
				Proto: &FormatSpecificAnnotations{Name: "Test"},
			},
		},
	}

	validator.Validate(annotations)

	if !validator.HasErrors() {
		t.Error("Expected validator to have errors")
	}

	formatted := validator.FormatErrors()
	if formatted == "" {
		t.Error("Expected non-empty formatted error string")
	}

	if len(formatted) < 10 {
		t.Errorf("Expected formatted errors to be substantial, got: %s", formatted)
	}
}
