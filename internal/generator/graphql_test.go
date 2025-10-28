package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

// Helper function to create a wrapper registry for testing
func newWrapperRegistry() *wrapperRegistry {
	return &wrapperRegistry{
		fieldToName: make(map[string]string),
	}
}

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

	registry := newWrapperRegistry()
	output := gen.generateType(typ, false, false, make(map[string]bool), make(map[string]string), make(map[string]string), registry)

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
			expected: "[StringStringEntry!]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := newWrapperRegistry()
			result := gen.convertFieldType(tt.field, false, make(map[string]string), make(map[string]string), registry)
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
		{&ast.FieldType{Name: "map", IsMap: true, MapKey: "string", MapValue: "string"}, "StringStringEntry"},
		{&ast.FieldType{Name: "map", IsMap: true, MapKey: "string", MapValue: "int32"}, "StringIntEntry"},
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

	registry := newWrapperRegistry()

	// Test generating as input with suffix
	inputOutput := gen.generateType(typ, true, true, make(map[string]bool), make(map[string]string), make(map[string]string), registry)
	if !strings.Contains(inputOutput, "input PostInput") {
		t.Error("Expected 'input PostInput' when isInput=true and addInputSuffix=true")
	}

	// Test generating as input without suffix
	inputNoSuffix := gen.generateType(typ, true, false, make(map[string]bool), make(map[string]string), make(map[string]string), registry)
	if !strings.Contains(inputNoSuffix, "input Post {") {
		t.Error("Expected 'input Post' when isInput=true and addInputSuffix=false")
	}

	// Test generating as output type
	outputType := gen.generateType(typ, false, false, make(map[string]bool), make(map[string]string), make(map[string]string), registry)
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

	registry := newWrapperRegistry()
	result := gen.convertFieldType(field, false, make(map[string]string), make(map[string]string), registry)
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

	registry := newWrapperRegistry()
	result := gen.convertFieldType(field, false, make(map[string]string), make(map[string]string), registry)
	expected := "String!"

	if result != expected {
		t.Errorf("Expected timestamp to map to %q, got %q", expected, result)
	}
}

func TestGraphQLGenerator_Namespace(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		expectedInOutput string
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

func TestGenerateOptionalFieldsGraphQL(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name:     "id",
						Required: true,
						Type: &ast.FieldType{
							Name:     "string",
							Optional: false,
						},
					},
					{
						Name:     "name",
						Required: false,
						Type: &ast.FieldType{
							Name:     "string",
							Optional: true,
						},
					},
					{
						Name:     "email",
						Required: true,
						Type: &ast.FieldType{
							Name:     "string",
							Optional: true, // Explicitly optional overrides @required
						},
					},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Required field without optional should have !
	if !strings.Contains(output, "id: String!") {
		t.Error("Expected 'id: String!' for required field")
	}

	// Optional field should not have !
	if !strings.Contains(output, "name: String\n") || strings.Contains(output, "name: String!") {
		t.Error("Expected 'name: String' (without !) for optional field")
	}

	// Explicitly optional should not have ! even if marked required
	if strings.Contains(output, "email: String!") {
		t.Error("Optional marker should override @required annotation")
	}
}

func TestGraphQLGenerator_CollectMapTypes(t *testing.T) {
	gen := NewGraphQLGenerator()

	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "metadata",
						Type: &ast.FieldType{
							IsMap:    true,
							MapKey:   "string",
							MapValue: "string",
						},
					},
					{
						Name: "scores",
						Type: &ast.FieldType{
							IsMap:    true,
							MapKey:   "string",
							MapValue: "int32",
						},
					},
				},
			},
			{
				Name: "Config",
				Fields: []*ast.Field{
					{
						Name: "settings",
						Type: &ast.FieldType{
							IsMap:    true,
							MapKey:   "string",
							MapValue: "string",
						},
					},
				},
			},
		},
	}

	registry := newWrapperRegistry()
	mapTypes, _ := gen.collectMapTypesWithRegistry(schema, registry)

	// Should collect unique map types
	if len(mapTypes) != 2 {
		t.Errorf("Expected 2 unique map types, got %d", len(mapTypes))
	}

	// Check that we have the expected map types
	foundStringString := false
	foundStringInt := false
	for _, mt := range mapTypes {
		if mt.KeyType == "string" && mt.ValueType == "string" {
			foundStringString = true
		}
		if mt.KeyType == "string" && mt.ValueType == "int32" {
			foundStringInt = true
		}
	}

	if !foundStringString {
		t.Error("Expected to find map<string, string> type")
	}
	if !foundStringInt {
		t.Error("Expected to find map<string, int32> type")
	}
}

func TestGraphQLGenerator_NestedMaps(t *testing.T) {
	gen := NewGraphQLGenerator()

	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "NestedMapTest",
				Fields: []*ast.Field{
					{
						Name: "simpleMap",
						Type: &ast.FieldType{
							IsMap:  true,
							MapKey: "string",
							MapValueType: &ast.FieldType{
								Name:      "string",
								IsBuiltin: true,
							},
						},
						Required: false,
					},
					{
						Name: "nestedMap",
						Type: &ast.FieldType{
							IsMap:  true,
							MapKey: "string",
							MapValueType: &ast.FieldType{
								IsMap:  true,
								MapKey: "string",
								MapValueType: &ast.FieldType{
									Name:      "int32",
									IsBuiltin: true,
								},
							},
						},
						Required: false,
					},
					{
						Name: "tripleNestedMap",
						Type: &ast.FieldType{
							IsMap:  true,
							MapKey: "string",
							MapValueType: &ast.FieldType{
								IsMap:  true,
								MapKey: "string",
								MapValueType: &ast.FieldType{
									IsMap:  true,
									MapKey: "string",
									MapValueType: &ast.FieldType{
										Name:      "bool",
										IsBuiltin: true,
									},
								},
							},
						},
						Required: false,
					},
				},
			},
		},
	}

	output := gen.Generate(schema)

	// Check that wrapper types are generated
	if !strings.Contains(output, "type MapWrapper") {
		t.Error("Expected MapWrapper types to be generated for nested maps")
	}

	// Check that the main type uses the wrapper types
	if !strings.Contains(output, "type NestedMapTest") {
		t.Error("Expected NestedMapTest type in output")
	}

	// Check simple map uses KeyValue entry
	if !strings.Contains(output, "StringStringEntry") {
		t.Error("Expected StringStringEntry for simple map")
	}

	// Check nested map uses wrapper entry
	if !strings.Contains(output, "MapWrapper") {
		t.Error("Expected MapWrapper for nested maps")
	}
}

func TestGraphQLGenerator_GetMapValueType(t *testing.T) {
	tests := []struct {
		name       string
		field      *ast.FieldType
		expectNil  bool
		expectName string
	}{
		{
			name: "simple map with MapValueType",
			field: &ast.FieldType{
				IsMap:  true,
				MapKey: "string",
				MapValueType: &ast.FieldType{
					Name:      "int32",
					IsBuiltin: true,
				},
			},
			expectNil:  false,
			expectName: "int32",
		},
		{
			name: "nested map",
			field: &ast.FieldType{
				IsMap:  true,
				MapKey: "string",
				MapValueType: &ast.FieldType{
					IsMap:  true,
					MapKey: "string",
					MapValueType: &ast.FieldType{
						Name:      "string",
						IsBuiltin: true,
					},
				},
			},
			expectNil: false,
		},
		{
			name: "backward compatibility with MapValue string",
			field: &ast.FieldType{
				IsMap:    true,
				MapKey:   "string",
				MapValue: "string",
			},
			expectNil:  false,
			expectName: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueType := tt.field.GetMapValueType()
			if tt.expectNil {
				if valueType != nil {
					t.Error("Expected GetMapValueType to return nil")
				}
			} else {
				if valueType == nil {
					t.Error("Expected GetMapValueType to return non-nil")
				} else if tt.expectName != "" && valueType.Name != tt.expectName {
					t.Errorf("Expected name %q, got %q", tt.expectName, valueType.Name)
				}
			}
		})
	}
}

func TestGraphQLGenerator_GetKeyValueTypeName(t *testing.T) {
	gen := NewGraphQLGenerator()

	tests := []struct {
		keyType   string
		valueType string
		expected  string
	}{
		{"string", "string", "StringStringEntry"},
		{"string", "int32", "StringIntEntry"},
		{"string", "int64", "StringIntEntry"},
		{"int32", "string", "IntStringEntry"},
		{"string", "User", "StringUserEntry"},
	}

	for _, tt := range tests {
		result := gen.getKeyValueTypeName(tt.keyType, tt.valueType)
		if result != tt.expected {
			t.Errorf("getKeyValueTypeName(%q, %q) = %q, want %q", tt.keyType, tt.valueType, result, tt.expected)
		}
	}
}

func TestGraphQLGenerator_CapitalizeTypeName(t *testing.T) {
	gen := NewGraphQLGenerator()

	tests := []struct {
		input    string
		expected string
	}{
		{"string", "String"},
		{"int", "Int"},
		{"user", "User"},
		{"", ""},
		{"a", "A"},
	}

	for _, tt := range tests {
		result := gen.capitalizeTypeName(tt.input)
		if result != tt.expected {
			t.Errorf("capitalizeTypeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGraphQLGenerator_MapScalarToGraphQLType(t *testing.T) {
	gen := NewGraphQLGenerator()

	tests := []struct {
		input    string
		expected string
	}{
		{"string", "String"},
		{"int32", "Int"},
		{"int64", "Int"},
		{"float32", "Float"},
		{"float64", "Float"},
		{"bool", "Boolean"},
		{"timestamp", "String"},
		{"bytes", "String"},
		{"User", "User"},
		{"com.example.User", "User"},
	}

	for _, tt := range tests {
		result := gen.mapScalarToGraphQLType(tt.input)
		if result != tt.expected {
			t.Errorf("mapScalarToGraphQLType(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGraphQLGenerator_GenerateKeyValueType(t *testing.T) {
	gen := NewGraphQLGenerator()

	mapType := MapTypeKey{
		KeyType:   "string",
		ValueType: "int32",
	}

	// Test output type
	output := gen.generateKeyValueType(mapType, false)
	if !strings.Contains(output, "type StringIntEntry") {
		t.Error("Expected type StringIntEntry in output")
	}
	if !strings.Contains(output, "key: String!") {
		t.Error("Expected key field with type String!")
	}
	if !strings.Contains(output, "value: Int!") {
		t.Error("Expected value field with type Int!")
	}
	if !strings.Contains(output, "map<string, int32>") {
		t.Error("Expected documentation mentioning original map type")
	}

	// Test input type
	inputOutput := gen.generateKeyValueType(mapType, true)
	if !strings.Contains(inputOutput, "input StringIntEntryInput") {
		t.Error("Expected input StringIntEntryInput in output")
	}
}

func TestGraphQLGenerator_MapTypes(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "Configuration",
				Fields: []*ast.Field{
					{
						Name: "metadata",
						Type: &ast.FieldType{
							Name:     "map",
							IsMap:    true,
							MapKey:   "string",
							MapValue: "string",
						},
						Required: true,
					},
					{
						Name: "scores",
						Type: &ast.FieldType{
							Name:     "map",
							IsMap:    true,
							MapKey:   "string",
							MapValue: "int32",
						},
						Required: false,
					},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Check that KeyValue types are generated for output types
	if !strings.Contains(output, "type StringStringEntry") {
		t.Error("Expected StringStringEntry type to be generated")
	}

	if !strings.Contains(output, "type StringIntEntry") {
		t.Error("Expected StringIntEntry type to be generated")
	}

	// Check that KeyValue input types are generated
	if !strings.Contains(output, "input StringStringEntryInput") {
		t.Error("Expected StringStringEntryInput type to be generated")
	}

	if !strings.Contains(output, "input StringIntEntryInput") {
		t.Error("Expected StringIntEntryInput type to be generated")
	}

	// Check that key and value fields are present
	if !strings.Contains(output, "key: String!") {
		t.Error("Expected key field with type String!")
	}

	if !strings.Contains(output, "value: String!") {
		t.Error("Expected value field with type String! in StringStringEntry")
	}

	if !strings.Contains(output, "value: Int!") {
		t.Error("Expected value field with type Int! in StringIntEntry")
	}

	// Check that the Configuration type uses the KeyValue types as arrays
	if !strings.Contains(output, "metadata: [StringStringEntry!]!") {
		t.Error("Expected metadata field to be [StringStringEntry!]! (required array of non-null entries)")
	}

	if !strings.Contains(output, "scores: [StringIntEntry!]") {
		t.Error("Expected scores field to be [StringIntEntry!] (optional array of non-null entries)")
	}

	// Check that JSON scalar is no longer generated
	if strings.Contains(output, "scalar JSON") {
		t.Error("Should not generate scalar JSON when using KeyValue types")
	}

	// Check for documentation
	if !strings.Contains(output, "map<string, string>") {
		t.Error("Expected documentation mentioning the original map type")
	}
}

func TestGraphQLGenerator_Subscriptions(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "Message",
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
						Name: "content",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
						Required: true,
					},
				},
			},
			{
				Name:   "Empty",
				Fields: []*ast.Field{},
			},
		},
		Services: []*ast.Service{
			{
				Name: "ChatService",
				Methods: []*ast.Method{
					{
						Name:         "GetMessage",
						InputType:    "Empty",
						OutputType:   "Message",
						InputStream:  false,
						OutputStream: false,
					},
					{
						Name:         "SendMessage",
						InputType:    "Message",
						OutputType:   "Empty",
						InputStream:  false,
						OutputStream: false,
					},
					{
						Name:         "WatchMessages",
						InputType:    "Empty",
						OutputType:   "Message",
						InputStream:  false,
						OutputStream: true, // Stream indicates subscription
					},
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Check that Query type contains GetMessage
	if !strings.Contains(output, "type Query {") {
		t.Error("Expected Query type to be generated")
	}
	if !strings.Contains(output, "getMessage") {
		t.Error("Expected getMessage in Query type")
	}

	// Check that Mutation type contains SendMessage
	if !strings.Contains(output, "type Mutation {") {
		t.Error("Expected Mutation type to be generated")
	}
	if !strings.Contains(output, "sendMessage") {
		t.Error("Expected sendMessage in Mutation type")
	}

	// Check that Subscription type contains WatchMessages
	if !strings.Contains(output, "type Subscription {") {
		t.Error("Expected Subscription type to be generated")
	}
	if !strings.Contains(output, "watchMessages") {
		t.Error("Expected watchMessages in Subscription type")
	}

	// Verify watchMessages is not in Query or Mutation
	lines := strings.Split(output, "\n")
	inQuery := false
	inMutation := false
	inSubscription := false

	for _, line := range lines {
		if strings.Contains(line, "type Query {") {
			inQuery = true
			inMutation = false
			inSubscription = false
		} else if strings.Contains(line, "type Mutation {") {
			inQuery = false
			inMutation = true
			inSubscription = false
		} else if strings.Contains(line, "type Subscription {") {
			inQuery = false
			inMutation = false
			inSubscription = true
		} else if strings.Contains(line, "}") {
			inQuery = false
			inMutation = false
			inSubscription = false
		}

		if (inQuery || inMutation) && strings.Contains(line, "watchMessages") {
			t.Error("watchMessages should only be in Subscription type, not Query or Mutation")
		}

		if inSubscription && (strings.Contains(line, "getMessage") || strings.Contains(line, "sendMessage")) {
			t.Error("Non-streaming methods should not be in Subscription type")
		}
	}
}
