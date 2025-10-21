package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestGraphQLGenerator_Generate(t *testing.T) {
	schema := &ast.Schema{
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
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "id",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
						Required: true,
					},
					{
						Name: "name",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
						Required: true,
					},
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
						OutputType: "GetUserResponse",
					},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Check for enum
	if !strings.Contains(output, "enum UserRole") {
		t.Error("Expected enum UserRole in output")
	}

	if !strings.Contains(output, "ADMIN") {
		t.Error("Expected ADMIN value in enum")
	}

	// Check for type
	if !strings.Contains(output, "type User") {
		t.Error("Expected type User in output")
	}

	if !strings.Contains(output, "id: String!") {
		t.Error("Expected id field to be String! (required)")
	}

	// Check for Query
	if !strings.Contains(output, "type Query") {
		t.Error("Expected type Query in output")
	}

	if !strings.Contains(output, "getUser") {
		t.Error("Expected getUser method in Query")
	}
}

func TestGraphQLGenerator_GenerateEnum(t *testing.T) {
	gen := NewGraphQLGenerator()
	enum := &ast.Enum{
		Name: "Status",
		Values: []*ast.EnumValue{
			{Name: "ACTIVE"},
			{Name: "INACTIVE"},
			{Name: "PENDING"},
		},
	}

	output := gen.generateEnum(enum)

	if !strings.Contains(output, "enum Status") {
		t.Error("Expected enum Status in output")
	}

	for _, value := range enum.Values {
		if !strings.Contains(output, value.Name) {
			t.Errorf("Expected value %q in enum output", value.Name)
		}
	}
}

func TestGraphQLGenerator_GenerateType(t *testing.T) {
	gen := NewGraphQLGenerator()
	typ := &ast.Type{
		Name: "Post",
		Fields: []*ast.Field{
			{
				Name: "id",
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
				Required: true,
			},
			{
				Name: "title",
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
				Required: false,
			},
			{
				Name: "count",
				Type: &ast.FieldType{
					Name:      "int32",
					IsBuiltin: true,
				},
				Required: false,
			},
		},
	}

	output := gen.generateType(typ, false, false, make(map[string]bool), make(map[string]string), make(map[string]string))

	if !strings.Contains(output, "type Post") {
		t.Error("Expected type Post in output")
	}

	if !strings.Contains(output, "id: String!") {
		t.Error("Expected id to be required (String!)")
	}

	if !strings.Contains(output, "title: String") && !strings.Contains(output, "title: String!") {
		t.Error("Expected title field")
	}

	if !strings.Contains(output, "count: Int") {
		t.Error("Expected count field to be Int")
	}
}

func TestGraphQLGenerator_ConvertFieldType(t *testing.T) {
	gen := NewGraphQLGenerator()

	tests := []struct {
		name     string
		field    *ast.Field
		expected string
	}{
		{
			name: "required string",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
				Required: true,
			},
			expected: "String!",
		},
		{
			name: "optional int",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "int32",
					IsBuiltin: true,
				},
				Required: false,
			},
			expected: "Int",
		},
		{
			name: "array of strings",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
					IsArray:   true,
				},
				Required: false,
			},
			expected: "[String]",
		},
		{
			name: "required array",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
					IsArray:   true,
				},
				Required: true,
			},
			expected: "[String]!",
		},
		{
			name: "boolean field",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "bool",
					IsBuiltin: true,
				},
				Required: false,
			},
			expected: "Boolean",
		},
		{
			name: "custom type",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "User",
					IsBuiltin: false,
				},
				Required: true,
			},
			expected: "User!",
		},
		{
			name: "map type",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:     "map",
					IsMap:    true,
					MapKey:   "string",
					MapValue: "string",
				},
				Required: false,
			},
			expected: "JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.convertFieldType(tt.field, false, make(map[string]string), make(map[string]string))
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGraphQLGenerator_MapTypeToGraphQL(t *testing.T) {
	gen := NewGraphQLGenerator()

	tests := []struct {
		fieldType *ast.FieldType
		expected  string
	}{
		{&ast.FieldType{Name: "string", IsBuiltin: true}, "String"},
		{&ast.FieldType{Name: "int32", IsBuiltin: true}, "Int"},
		{&ast.FieldType{Name: "int64", IsBuiltin: true}, "Int"},
		{&ast.FieldType{Name: "float32", IsBuiltin: true}, "Float"},
		{&ast.FieldType{Name: "float64", IsBuiltin: true}, "Float"},
		{&ast.FieldType{Name: "bool", IsBuiltin: true}, "Boolean"},
		{&ast.FieldType{Name: "timestamp", IsBuiltin: true}, "String"},
		{&ast.FieldType{Name: "bytes", IsBuiltin: true}, "String"},
		{&ast.FieldType{Name: "User", IsBuiltin: false}, "User"},
		{&ast.FieldType{Name: "map", IsMap: true}, "JSON"},
	}

	for _, tt := range tests {
		result := gen.mapTypeToGraphQL(tt.fieldType)
		if result != tt.expected {
			t.Errorf("mapTypeToGraphQL(%q) = %q, want %q", tt.fieldType.Name, result, tt.expected)
		}
	}
}

func TestGraphQLGenerator_GenerateServiceMethod(t *testing.T) {
	gen := NewGraphQLGenerator()

	method := &ast.Method{
		Name:       "CreateUser",
		InputType:  "CreateUserRequest",
		OutputType: "CreateUserResponse",
	}

	typeUsage := map[string]string{
		"CreateUserRequest":  "input",
		"CreateUserResponse": "output",
	}
	result := gen.generateServiceMethod(method, typeUsage)

	if !strings.Contains(result, "createUser") {
		t.Error("Expected camelCase method name 'createUser'")
	}

	if !strings.Contains(result, "CreateUserRequest") {
		t.Error("Expected input type in method signature")
	}

	if !strings.Contains(result, "CreateUserResponse") {
		t.Error("Expected output type in method signature")
	}
}

func TestGraphQLGenerator_AnalyzeTypeUsage(t *testing.T) {
	gen := NewGraphQLGenerator()

	schema := &ast.Schema{
		Services: []*ast.Service{
			{
				Name: "UserService",
				Methods: []*ast.Method{
					{
						Name:       "CreateUser",
						InputType:  "CreateUserRequest",
						OutputType: "CreateUserResponse",
					},
					{
						Name:       "GetUser",
						InputType:  "GetUserRequest",
						OutputType: "User",
					},
				},
			},
			{
				Name: "PostService",
				Methods: []*ast.Method{
					{
						Name:       "CreatePost",
						InputType:  "Post",
						OutputType: "Post",
					},
				},
			},
		},
	}

	typeUsage := gen.analyzeTypeUsage(schema)

	// CreateUserRequest is only used as input
	if typeUsage["CreateUserRequest"] != "input" {
		t.Errorf("Expected CreateUserRequest to be 'input', got '%s'", typeUsage["CreateUserRequest"])
	}

	// CreateUserResponse is only used as output
	if typeUsage["CreateUserResponse"] != "output" {
		t.Errorf("Expected CreateUserResponse to be 'output', got '%s'", typeUsage["CreateUserResponse"])
	}

	// Post is used as both input and output
	if typeUsage["Post"] != "both" {
		t.Errorf("Expected Post to be 'both', got '%s'", typeUsage["Post"])
	}

	// User is only used as output
	if typeUsage["User"] != "output" {
		t.Errorf("Expected User to be 'output', got '%s'", typeUsage["User"])
	}
}

func TestGraphQLGenerator_GenerateType_InputSuffix(t *testing.T) {
	gen := NewGraphQLGenerator()

	typ := &ast.Type{
		Name: "Post",
		Fields: []*ast.Field{
			{
				Name: "id",
				Type: &ast.FieldType{
					Name:    "string",
					IsArray: false,
				},
				Required: true,
			},
			{
				Name: "title",
				Type: &ast.FieldType{
					Name:    "string",
					IsArray: false,
				},
				Required: true,
			},
		},
	}

	// Test generating as input with suffix
	inputOutput := gen.generateType(typ, true, true, make(map[string]bool), make(map[string]string), make(map[string]string))
	if !strings.Contains(inputOutput, "input PostInput") {
		t.Error("Expected 'input PostInput' when isInput=true and addInputSuffix=true")
	}

	// Test generating as input without suffix
	inputNoSuffix := gen.generateType(typ, true, false, make(map[string]bool), make(map[string]string), make(map[string]string))
	if !strings.Contains(inputNoSuffix, "input Post {") {
		t.Error("Expected 'input Post' when isInput=true and addInputSuffix=false")
	}

	// Test generating as output type
	outputType := gen.generateType(typ, false, false, make(map[string]bool), make(map[string]string), make(map[string]string))
	if !strings.Contains(outputType, "type Post") {
		t.Error("Expected 'type Post' when isInput=false")
	}
}

func TestGraphQLGenerator_ServiceMethod_WithBothType(t *testing.T) {
	gen := NewGraphQLGenerator()

	method := &ast.Method{
		Name:       "CreatePost",
		InputType:  "Post",
		OutputType: "Post",
	}

	// When Post is used as both input and output, should use PostInput for input
	typeUsage := map[string]string{
		"Post": "both",
	}
	result := gen.generateServiceMethod(method, typeUsage)

	if !strings.Contains(result, "createPost(input: PostInput)") {
		t.Error("Expected method to use PostInput when type is used as both")
	}

	if !strings.Contains(result, "): Post") {
		t.Error("Expected method to return Post")
	}
}

func TestGraphQLGenerator_ServiceMethod_WithInputOnly(t *testing.T) {
	gen := NewGraphQLGenerator()

	method := &ast.Method{
		Name:       "CreateUser",
		InputType:  "CreateUserRequest",
		OutputType: "CreateUserResponse",
	}

	// When types are only used as input or output, use them as-is
	typeUsage := map[string]string{
		"CreateUserRequest":  "input",
		"CreateUserResponse": "output",
	}
	result := gen.generateServiceMethod(method, typeUsage)

	if !strings.Contains(result, "createUser(input: CreateUserRequest)") {
		t.Error("Expected method to use CreateUserRequest without suffix")
	}

	if !strings.Contains(result, "): CreateUserResponse") {
		t.Error("Expected method to return CreateUserResponse")
	}
}

func TestGraphQLGenerator_QueryAndMutationSeparation(t *testing.T) {
	schema := &ast.Schema{
		Services: []*ast.Service{
			{
				Name: "UserService",
				Methods: []*ast.Method{
					{Name: "GetUser", InputType: "Req", OutputType: "Res"},
					{Name: "ListUsers", InputType: "Req", OutputType: "Res"},
					{Name: "CreateUser", InputType: "Req", OutputType: "Res"},
					{Name: "UpdateUser", InputType: "Req", OutputType: "Res"},
					{Name: "DeleteUser", InputType: "Req", OutputType: "Res"},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Check Query type contains Get and List methods
	if !strings.Contains(output, "type Query") {
		t.Error("Expected type Query in output")
	}

	querySection := output[strings.Index(output, "type Query"):]
	if !strings.Contains(querySection, "getUser") {
		t.Error("Expected getUser in Query type")
	}

	if !strings.Contains(querySection, "listUsers") {
		t.Error("Expected listUsers in Query type")
	}

	// Check Mutation type contains Create, Update, Delete methods
	if !strings.Contains(output, "type Mutation") {
		t.Error("Expected type Mutation in output")
	}

	mutationSection := output[strings.Index(output, "type Mutation"):]
	if !strings.Contains(mutationSection, "createUser") {
		t.Error("Expected createUser in Mutation type")
	}

	if !strings.Contains(mutationSection, "updateUser") {
		t.Error("Expected updateUser in Mutation type")
	}

	if !strings.Contains(mutationSection, "deleteUser") {
		t.Error("Expected deleteUser in Mutation type")
	}
}

func TestGraphQLGenerator_EmptySchema(t *testing.T) {
	schema := &ast.Schema{
		Enums:    []*ast.Enum{},
		Types:    []*ast.Type{},
		Services: []*ast.Service{},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	if !strings.Contains(output, "# Generated GraphQL Schema") {
		t.Error("Expected header comment in output")
	}

	// Should not generate Query or Mutation types if there are no services
	if strings.Contains(output, "type Query {") {
		t.Error("Should not generate Query type for empty schema")
	}

	if strings.Contains(output, "type Mutation {") {
		t.Error("Should not generate Mutation type for empty schema")
	}
}

func TestGraphQLGenerator_ArrayOfCustomTypes(t *testing.T) {
	gen := NewGraphQLGenerator()
	field := &ast.Field{
		Name: "users",
		Type: &ast.FieldType{
			Name:      "User",
			IsBuiltin: false,
			IsArray:   true,
		},
		Required: true,
	}

	result := gen.convertFieldType(field, false, make(map[string]string), make(map[string]string))
	expected := "[User]!"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGraphQLGenerator_TimestampType(t *testing.T) {
	gen := NewGraphQLGenerator()
	field := &ast.Field{
		Type: &ast.FieldType{
			Name:      "timestamp",
			IsBuiltin: true,
		},
		Required: true,
	}

	result := gen.convertFieldType(field, false, make(map[string]string), make(map[string]string))
	expected := "String!"

	if result != expected {
		t.Errorf("Expected timestamp to map to %q, got %q", expected, result)
	}
}

func TestGraphQLGenerator_Namespace(t *testing.T) {
	tests := []struct {
		name              string
		namespace         string
		expectedInOutput  string
	}{
		{
			name:             "namespace in comment",
			namespace:        "com.example.api",
			expectedInOutput: "# Namespace: com.example.api",
		},
		{
			name:             "no namespace comment when empty",
			namespace:        "",
			expectedInOutput: "# Generated GraphQL Schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGraphQLGenerator()
			schema := &ast.Schema{
				Namespace: tt.namespace,
				Enums:     []*ast.Enum{},
				Types:     []*ast.Type{},
				Services:  []*ast.Service{},
			}

			output := gen.Generate(schema)

			if !strings.Contains(output, tt.expectedInOutput) {
				t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", tt.expectedInOutput, output)
			}
		})
	}
}

func TestGraphQLGenerator_CheckForDuplicates(t *testing.T) {
	gen := NewGraphQLGenerator()

	tests := []struct {
		name        string
		schema      *ast.Schema
		expectError bool
		errorMsg    string
	}{
		{
			name: "no duplicates",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{Name: "User", Namespace: "com.example.users"},
					{Name: "Order", Namespace: "com.example.orders"},
				},
			},
			expectError: false,
		},
		{
			name: "duplicate type names in different namespaces",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{Name: "User", Namespace: "com.example.users"},
					{Name: "User", Namespace: "com.example.orders"},
				},
			},
			expectError: true,
			errorMsg:    "duplicate type name 'User'",
		},
		{
			name: "duplicate enum names in different namespaces",
			schema: &ast.Schema{
				Enums: []*ast.Enum{
					{Name: "Status", Namespace: "com.example.users"},
					{Name: "Status", Namespace: "com.example.orders"},
				},
			},
			expectError: true,
			errorMsg:    "duplicate enum name 'Status'",
		},
		{
			name: "duplicate union names in different namespaces",
			schema: &ast.Schema{
				Unions: []*ast.Union{
					{Name: "Result", Namespace: "com.example.users"},
					{Name: "Result", Namespace: "com.example.orders"},
				},
			},
			expectError: true,
			errorMsg:    "duplicate union name 'Result'",
		},
		{
			name: "same type name in same namespace multiple times",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{Name: "User", Namespace: "com.example.api"},
					{Name: "User", Namespace: "com.example.api"},
				},
			},
			expectError: false, // Same namespace, so it's OK (duplicates are filtered)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.checkForDuplicates(tt.schema)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGraphQLGenerator_Generate_WithDuplicates(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{Name: "User", Namespace: "com.example.users"},
			{Name: "User", Namespace: "com.example.orders"},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Should contain error message
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected output to contain ERROR message")
	}

	if !strings.Contains(output, "duplicate type name") {
		t.Error("Expected output to contain duplicate type name message")
	}

	if !strings.Contains(output, "GraphQL does not support multiple types with the same name") {
		t.Error("Expected output to contain explanation about GraphQL limitation")
	}
}

func TestGraphQLGenerator_NameAnnotation(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "api",
				Annotations: &ast.FormatAnnotations{
					GraphQLName: "UserAccount",
				},
				Fields: []*ast.Field{
					{
						Name: "id",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
						Required: true,
					},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	if !strings.Contains(output, "type UserAccount {") {
		t.Error("Expected output to contain 'type UserAccount {', but it didn't")
	}

	if strings.Contains(output, "type User {") {
		t.Error("Expected output NOT to contain 'type User {', but it did")
	}
}
