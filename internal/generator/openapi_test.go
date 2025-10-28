package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
	"gopkg.in/yaml.v3"
)

func TestOpenAPIGenerator_Generate(t *testing.T) {
	schema := &ast.Schema{
		Enums: []*ast.Enum{
			{
				Name: "UserRole",
				Values: []*ast.EnumValue{
					{Name: "ADMIN"},
					{Name: "USER"},
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

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	// Verify it's valid YAML
	var spec OpenAPISpec
	err := yaml.Unmarshal([]byte(output), &spec)
	if err != nil {
		t.Fatalf("Failed to parse OpenAPI YAML: %v", err)
	}

	// Check OpenAPI version
	if spec.OpenAPI != "3.0.0" {
		t.Errorf("Expected OpenAPI version 3.0.0, got %s", spec.OpenAPI)
	}

	// Check schema components
	if _, ok := spec.Components.Schemas["UserRole"]; !ok {
		t.Error("Expected UserRole in schemas")
	}

	if _, ok := spec.Components.Schemas["User"]; !ok {
		t.Error("Expected User in schemas")
	}

	// Check paths
	if len(spec.Paths) == 0 {
		t.Error("Expected paths to be generated")
	}
}

func TestOpenAPIGenerator_GenerateSchema(t *testing.T) {
	gen := NewOpenAPIGenerator()
	typ := &ast.Type{
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
			{
				Name: "age",
				Type: &ast.FieldType{
					Name:      "int32",
					IsBuiltin: true,
				},
				Required: false,
			},
		},
	}

	schema := gen.generateSchema(typ, make(map[string]string))

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %q", schema.Type)
	}

	if len(schema.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(schema.Properties))
	}

	if len(schema.Required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(schema.Required))
	}

	// Check properties
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("Expected 'id' property")
	}

	if _, ok := schema.Properties["name"]; !ok {
		t.Error("Expected 'name' property")
	}

	if _, ok := schema.Properties["age"]; !ok {
		t.Error("Expected 'age' property")
	}
}

func TestOpenAPIGenerator_ConvertFieldToProperty(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		name          string
		field         *ast.Field
		expectedType  string
		expectedRef   string
		expectedArray bool
	}{
		{
			name: "string field",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
			},
			expectedType: "string",
		},
		{
			name: "int32 field",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "int32",
					IsBuiltin: true,
				},
			},
			expectedType: "integer",
		},
		{
			name: "bool field",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "bool",
					IsBuiltin: true,
				},
			},
			expectedType: "boolean",
		},
		{
			name: "array field",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
					IsArray:   true,
				},
			},
			expectedType:  "array",
			expectedArray: true,
		},
		{
			name: "custom type field",
			field: &ast.Field{
				Type: &ast.FieldType{
					Name:      "User",
					IsBuiltin: false,
				},
			},
			expectedRef: "#/components/schemas/User",
		},
		{
			name: "map field with string values",
			field: &ast.Field{
				Type: &ast.FieldType{
					IsMap:    true,
					MapKey:   "string",
					MapValue: "string",
				},
			},
			expectedType: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.convertFieldToProperty(tt.field, make(map[string]string))

			if tt.expectedType != "" && result.Type != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, result.Type)
			}

			if tt.expectedRef != "" && result.Ref != tt.expectedRef {
				t.Errorf("Expected ref %q, got %q", tt.expectedRef, result.Ref)
			}

			if tt.expectedArray && result.Items == nil {
				t.Error("Expected items to be set for array type")
			}
		})
	}
}

func TestOpenAPIGenerator_MapTypeToOpenAPI(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		typeName string
		expected string
	}{
		{"string", "string"},
		{"int32", "integer"},
		{"int64", "integer"},
		{"float32", "number"},
		{"float64", "number"},
		{"bool", "boolean"},
		{"timestamp", "string"},
		{"bytes", "string"},
		{"CustomType", "object"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := gen.mapTypeToOpenAPI(tt.typeName)
			if result != tt.expected {
				t.Errorf("mapTypeToOpenAPI(%q) = %q, want %q", tt.typeName, result, tt.expected)
			}
		})
	}
}

func TestOpenAPIGenerator_GetFormatForType(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		typeName string
		expected string
	}{
		{"int32", "int32"},
		{"int64", "int64"},
		{"float32", "float"},
		{"float64", "double"},
		{"timestamp", "date-time"},
		{"bytes", "byte"},
		{"string", ""},
		{"bool", ""},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := gen.getFormatForType(tt.typeName)
			if result != tt.expected {
				t.Errorf("getFormatForType(%q) = %q, want %q", tt.typeName, result, tt.expected)
			}
		})
	}
}

func TestOpenAPIGenerator_AddServiceMethod(t *testing.T) {
	gen := NewOpenAPIGenerator()
	spec := &OpenAPISpec{
		Paths: make(map[string]map[string]OpenAPIOperation),
		Components: OpenAPIComponents{
			Schemas: make(map[string]OpenAPISchema),
		},
	}

	service := &ast.Service{
		Name: "UserService",
	}

	method := &ast.Method{
		Name:       "GetUser",
		InputType:  "GetUserRequest",
		OutputType: "GetUserResponse",
	}

	gen.addServiceMethod(spec, service, method, make(map[string]string))

	// Check path was created
	path := "/userservice/getuser"
	if _, ok := spec.Paths[path]; !ok {
		t.Errorf("Expected path %q to be created", path)
	}

	// Check GET method (since it's a Get* method)
	if _, ok := spec.Paths[path]["get"]; !ok {
		t.Error("Expected GET method for GetUser")
	}
}

func TestOpenAPIGenerator_HTTPMethodSelection(t *testing.T) {
	gen := NewOpenAPIGenerator()
	spec := &OpenAPISpec{
		Paths: make(map[string]map[string]OpenAPIOperation),
		Components: OpenAPIComponents{
			Schemas: make(map[string]OpenAPISchema),
		},
	}

	service := &ast.Service{Name: "TestService"}

	tests := []struct {
		methodName   string
		expectedHTTP string
	}{
		{"GetUser", "get"},
		{"ListUsers", "get"},
		{"CreateUser", "post"},
		{"UpdateUser", "post"},
		{"DeleteUser", "post"},
	}

	for _, tt := range tests {
		t.Run(tt.methodName, func(t *testing.T) {
			method := &ast.Method{
				Name:       tt.methodName,
				InputType:  "Request",
				OutputType: "Response",
			}

			gen.addServiceMethod(spec, service, method, make(map[string]string))

			path := "/testservice/" + strings.ToLower(tt.methodName)
			if methods, ok := spec.Paths[path]; ok {
				if _, ok := methods[tt.expectedHTTP]; !ok {
					t.Errorf("Expected HTTP method %q for %q", tt.expectedHTTP, tt.methodName)
				}
			} else {
				t.Errorf("Path %q not found", path)
			}
		})
	}
}

func TestOpenAPIGenerator_EnumSchema(t *testing.T) {
	schema := &ast.Schema{
		Enums: []*ast.Enum{
			{
				Name: "Status",
				Values: []*ast.EnumValue{
					{Name: "ACTIVE"},
					{Name: "INACTIVE"},
					{Name: "PENDING"},
				},
			},
		},
		Types:    []*ast.Type{},
		Services: []*ast.Service{},
	}

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	var spec OpenAPISpec
	err := yaml.Unmarshal([]byte(output), &spec)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	statusSchema, ok := spec.Components.Schemas["Status"]
	if !ok {
		t.Fatal("Expected Status enum in schemas")
	}

	if statusSchema.Type != "string" {
		t.Errorf("Expected enum type to be 'string', got %q", statusSchema.Type)
	}

	if len(statusSchema.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(statusSchema.Enum))
	}

	expectedValues := []string{"ACTIVE", "INACTIVE", "PENDING"}
	for i, expected := range expectedValues {
		if statusSchema.Enum[i] != expected {
			t.Errorf("Enum value %d: expected %q, got %q", i, expected, statusSchema.Enum[i])
		}
	}
}

func TestOpenAPIGenerator_ArrayOfCustomTypes(t *testing.T) {
	gen := NewOpenAPIGenerator()
	field := &ast.Field{
		Name: "users",
		Type: &ast.FieldType{
			Name:      "User",
			IsBuiltin: false,
			IsArray:   true,
		},
	}

	property := gen.convertFieldToProperty(field, make(map[string]string))

	if property.Type != "array" {
		t.Errorf("Expected type 'array', got %q", property.Type)
	}

	if property.Items == nil {
		t.Fatal("Expected items to be set")
	}

	if property.Items.Ref != "#/components/schemas/User" {
		t.Errorf("Expected items ref to be '#/components/schemas/User', got %q", property.Items.Ref)
	}
}

func TestOpenAPIGenerator_DefaultValue(t *testing.T) {
	gen := NewOpenAPIGenerator()
	field := &ast.Field{
		Name: "isActive",
		Type: &ast.FieldType{
			Name:      "bool",
			IsBuiltin: true,
		},
		Default: "true",
	}

	property := gen.convertFieldToProperty(field, make(map[string]string))

	if property.Default == nil {
		t.Fatal("Expected default value to be set")
	}

	// Default should be converted to boolean true, not string "true"
	if boolVal, ok := property.Default.(bool); !ok {
		t.Errorf("Expected default to be bool type, got %T", property.Default)
	} else if boolVal != true {
		t.Errorf("Expected default true, got %v", boolVal)
	}
}

func TestOpenAPIGenerator_EmptySchema(t *testing.T) {
	schema := &ast.Schema{
		Enums:    []*ast.Enum{},
		Types:    []*ast.Type{},
		Services: []*ast.Service{},
	}

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	var spec OpenAPISpec
	err := yaml.Unmarshal([]byte(output), &spec)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if spec.OpenAPI != "3.0.0" {
		t.Error("Expected OpenAPI version even for empty schema")
	}

	if spec.Info.Title == "" {
		t.Error("Expected info title even for empty schema")
	}

	if len(spec.Paths) != 0 {
		t.Error("Expected no paths for empty schema")
	}
}

func TestOpenAPIGenerator_RequestBody(t *testing.T) {
	gen := NewOpenAPIGenerator()
	spec := &OpenAPISpec{
		Paths: make(map[string]map[string]OpenAPIOperation),
		Components: OpenAPIComponents{
			Schemas: make(map[string]OpenAPISchema),
		},
	}

	service := &ast.Service{Name: "UserService"}
	method := &ast.Method{
		Name:       "CreateUser",
		InputType:  "CreateUserRequest",
		OutputType: "CreateUserResponse",
	}

	gen.addServiceMethod(spec, service, method, make(map[string]string))

	path := "/userservice/createuser"
	operation := spec.Paths[path]["post"]

	if operation.RequestBody == nil {
		t.Fatal("Expected request body for POST method")
	}

	if !operation.RequestBody.Required {
		t.Error("Expected request body to be required")
	}

	content, ok := operation.RequestBody.Content["application/json"]
	if !ok {
		t.Fatal("Expected application/json content type")
	}

	expectedRef := "#/components/schemas/CreateUserRequest"
	if content.Schema.Ref != expectedRef {
		t.Errorf("Expected schema ref %q, got %q", expectedRef, content.Schema.Ref)
	}
}

func TestOpenAPIGenerator_ResponseSchema(t *testing.T) {
	gen := NewOpenAPIGenerator()
	spec := &OpenAPISpec{
		Paths: make(map[string]map[string]OpenAPIOperation),
		Components: OpenAPIComponents{
			Schemas: make(map[string]OpenAPISchema),
		},
	}

	service := &ast.Service{Name: "UserService"}
	method := &ast.Method{
		Name:       "GetUser",
		InputType:  "GetUserRequest",
		OutputType: "GetUserResponse",
	}

	gen.addServiceMethod(spec, service, method, make(map[string]string))

	path := "/userservice/getuser"
	operation := spec.Paths[path]["get"]

	response, ok := operation.Responses["200"]
	if !ok {
		t.Fatal("Expected 200 response")
	}

	if response.Description == "" {
		t.Error("Expected response description")
	}

	content, ok := response.Content["application/json"]
	if !ok {
		t.Fatal("Expected application/json content type")
	}

	expectedRef := "#/components/schemas/GetUserResponse"
	if content.Schema.Ref != expectedRef {
		t.Errorf("Expected schema ref %q, got %q", expectedRef, content.Schema.Ref)
	}
}

func TestOpenAPIGenerator_TimestampFormat(t *testing.T) {
	gen := NewOpenAPIGenerator()
	field := &ast.Field{
		Name: "createdAt",
		Type: &ast.FieldType{
			Name:      "timestamp",
			IsBuiltin: true,
		},
	}

	property := gen.convertFieldToProperty(field, make(map[string]string))

	if property.Type != "string" {
		t.Errorf("Expected timestamp type to be 'string', got %q", property.Type)
	}

	if property.Format != "date-time" {
		t.Errorf("Expected timestamp format to be 'date-time', got %q", property.Format)
	}
}

func TestOpenAPIGenerator_IntegerFormats(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		typeName       string
		expectedFormat string
	}{
		{"int32", "int32"},
		{"int64", "int64"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			field := &ast.Field{
				Type: &ast.FieldType{
					Name:      tt.typeName,
					IsBuiltin: true,
				},
			}

			property := gen.convertFieldToProperty(field, make(map[string]string))

			if property.Type != "integer" {
				t.Errorf("Expected type 'integer', got %q", property.Type)
			}

			if property.Format != tt.expectedFormat {
				t.Errorf("Expected format %q, got %q", tt.expectedFormat, property.Format)
			}
		})
	}
}

func TestOpenAPIGenerator_NumberFormats(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		typeName       string
		expectedFormat string
	}{
		{"float32", "float"},
		{"float64", "double"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			field := &ast.Field{
				Type: &ast.FieldType{
					Name:      tt.typeName,
					IsBuiltin: true,
				},
			}

			property := gen.convertFieldToProperty(field, make(map[string]string))

			if property.Type != "number" {
				t.Errorf("Expected type 'number', got %q", property.Type)
			}

			if property.Format != tt.expectedFormat {
				t.Errorf("Expected format %q, got %q", tt.expectedFormat, property.Format)
			}
		})
	}
}
func TestOpenAPIGenerator_PathTemplates(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		name             string
		schema           *ast.Schema
		expectedPath     string
		shouldHaveParams bool
		paramNames       []string
	}{
		{
			name: "custom path template with single parameter",
			schema: &ast.Schema{
				Services: []*ast.Service{
					{
						Name: "UserService",
						Methods: []*ast.Method{
							{
								Name:         "GetUser",
								InputType:    "GetUserRequest",
								OutputType:   "GetUserResponse",
								PathTemplate: "/users/{id}",
								HTTPMethod:   "GET",
							},
						},
					},
				},
			},
			expectedPath:     "/users/{id}",
			shouldHaveParams: true,
			paramNames:       []string{"id"},
		},
		{
			name: "custom path with multiple parameters",
			schema: &ast.Schema{
				Services: []*ast.Service{
					{
						Name: "BlogService",
						Methods: []*ast.Method{
							{
								Name:         "GetPost",
								InputType:    "GetPostRequest",
								OutputType:   "GetPostResponse",
								PathTemplate: "/users/{userId}/posts/{postId}",
							},
						},
					},
				},
			},
			expectedPath:     "/users/{userId}/posts/{postId}",
			shouldHaveParams: true,
			paramNames:       []string{"userId", "postId"},
		},
		{
			name: "path without parameters",
			schema: &ast.Schema{
				Services: []*ast.Service{
					{
						Name: "UserService",
						Methods: []*ast.Method{
							{
								Name:         "ListUsers",
								InputType:    "ListUsersRequest",
								OutputType:   "ListUsersResponse",
								PathTemplate: "/api/v1/users",
							},
						},
					},
				},
			},
			expectedPath:     "/api/v1/users",
			shouldHaveParams: false,
		},
		{
			name: "default path generation when no template",
			schema: &ast.Schema{
				Services: []*ast.Service{
					{
						Name: "UserService",
						Methods: []*ast.Method{
							{
								Name:       "CreateUser",
								InputType:  "CreateUserRequest",
								OutputType: "CreateUserResponse",
							},
						},
					},
				},
			},
			expectedPath:     "/userservice/createuser",
			shouldHaveParams: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := gen.Generate(tt.schema)

			// Parse YAML to check structure
			var spec map[string]interface{}
			if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			paths, ok := spec["paths"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected paths in OpenAPI spec")
			}

			if _, exists := paths[tt.expectedPath]; !exists {
				t.Errorf("Expected path %q not found in spec. Available paths: %v", tt.expectedPath, paths)
			}

			// Check parameters if expected
			if tt.shouldHaveParams {
				pathData := paths[tt.expectedPath].(map[string]interface{})
				// Get the first HTTP method (get, post, etc.)
				var operation map[string]interface{}
				for _, v := range pathData {
					operation = v.(map[string]interface{})
					break
				}

				params, hasParams := operation["parameters"]
				if !hasParams {
					t.Error("Expected parameters in operation")
				} else {
					paramList := params.([]interface{})
					if len(paramList) != len(tt.paramNames) {
						t.Errorf("Expected %d parameters, got %d", len(tt.paramNames), len(paramList))
					}

					for i, expectedName := range tt.paramNames {
						if i >= len(paramList) {
							break
						}
						param := paramList[i].(map[string]interface{})
						if param["name"] != expectedName {
							t.Errorf("Parameter %d: expected name %q, got %q", i, expectedName, param["name"])
						}
						if param["in"] != "path" {
							t.Errorf("Parameter %d: expected in=path, got %q", i, param["in"])
						}
						if param["required"] != true {
							t.Errorf("Parameter %d: expected required=true", i)
						}
					}
				}
			}
		})
	}
}

func TestOpenAPIGenerator_ConvertDefaultValue(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		name         string
		defaultStr   string
		typeName     string
		expectedType string
		expectedVal  interface{}
	}{
		{
			name:         "int32 value",
			defaultStr:   "42",
			typeName:     "int32",
			expectedType: "int64",
			expectedVal:  int64(42),
		},
		{
			name:         "int64 value",
			defaultStr:   "1000",
			typeName:     "int64",
			expectedType: "int64",
			expectedVal:  int64(1000),
		},
		{
			name:         "float32 value",
			defaultStr:   "3.14",
			typeName:     "float32",
			expectedType: "float64",
			expectedVal:  float64(3.14),
		},
		{
			name:         "float64 value",
			defaultStr:   "2.71828",
			typeName:     "float64",
			expectedType: "float64",
			expectedVal:  float64(2.71828),
		},
		{
			name:         "bool true",
			defaultStr:   "true",
			typeName:     "bool",
			expectedType: "bool",
			expectedVal:  true,
		},
		{
			name:         "bool false",
			defaultStr:   "false",
			typeName:     "bool",
			expectedType: "bool",
			expectedVal:  false,
		},
		{
			name:         "string value",
			defaultStr:   "hello",
			typeName:     "string",
			expectedType: "string",
			expectedVal:  "hello",
		},
		{
			name:         "zero int",
			defaultStr:   "0",
			typeName:     "int32",
			expectedType: "int64",
			expectedVal:  int64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.convertDefaultValue(tt.defaultStr, tt.typeName)

			// Check type
			switch tt.expectedType {
			case "int64":
				if _, ok := result.(int64); !ok {
					t.Errorf("Expected int64 type, got %T", result)
				}
			case "float64":
				if _, ok := result.(float64); !ok {
					t.Errorf("Expected float64 type, got %T", result)
				}
			case "bool":
				if _, ok := result.(bool); !ok {
					t.Errorf("Expected bool type, got %T", result)
				}
			case "string":
				if _, ok := result.(string); !ok {
					t.Errorf("Expected string type, got %T", result)
				}
			}

			// Check value
			if result != tt.expectedVal {
				t.Errorf("Expected value %v, got %v", tt.expectedVal, result)
			}
		})
	}
}

func TestOpenAPIGenerator_DefaultValues_InSchema(t *testing.T) {
	gen := NewOpenAPIGenerator()

	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "Config",
				Fields: []*ast.Field{
					{
						Name: "port",
						Type: &ast.FieldType{
							Name: "int32",
						},
						Required: false,
						Default:  "8080",
					},
					{
						Name: "timeout",
						Type: &ast.FieldType{
							Name: "int64",
						},
						Required: false,
						Default:  "30",
					},
					{
						Name: "enabled",
						Type: &ast.FieldType{
							Name: "bool",
						},
						Required: false,
						Default:  "true",
					},
					{
						Name: "name",
						Type: &ast.FieldType{
							Name: "string",
						},
						Required: false,
						Default:  "default",
					},
				},
			},
		},
	}

	output := gen.Generate(schema)

	// Parse YAML to check default value types
	var spec OpenAPISpec
	err := yaml.Unmarshal([]byte(output), &spec)
	if err != nil {
		t.Fatalf("Failed to parse OpenAPI YAML: %v", err)
	}

	configSchema := spec.Components.Schemas["Config"]

	// Check port default is integer
	portProp := configSchema.Properties["port"]
	if portInt, ok := portProp.Default.(int); !ok {
		t.Errorf("Expected port default to be int, got %T: %v", portProp.Default, portProp.Default)
	} else if portInt != 8080 {
		t.Errorf("Expected port default to be 8080, got %d", portInt)
	}

	// Check timeout default is integer
	timeoutProp := configSchema.Properties["timeout"]
	if timeoutInt, ok := timeoutProp.Default.(int); !ok {
		t.Errorf("Expected timeout default to be int, got %T: %v", timeoutProp.Default, timeoutProp.Default)
	} else if timeoutInt != 30 {
		t.Errorf("Expected timeout default to be 30, got %d", timeoutInt)
	}

	// Check enabled default is boolean
	enabledProp := configSchema.Properties["enabled"]
	if enabledBool, ok := enabledProp.Default.(bool); !ok {
		t.Errorf("Expected enabled default to be bool, got %T: %v", enabledProp.Default, enabledProp.Default)
	} else if enabledBool != true {
		t.Errorf("Expected enabled default to be true, got %v", enabledBool)
	}

	// Check name default is string
	nameProp := configSchema.Properties["name"]
	if nameStr, ok := nameProp.Default.(string); !ok {
		t.Errorf("Expected name default to be string, got %T: %v", nameProp.Default, nameProp.Default)
	} else if nameStr != "default" {
		t.Errorf("Expected name default to be 'default', got %s", nameStr)
	}
}

func TestOpenAPIGenerator_Namespace(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		expectedTitle string
	}{
		{
			name:          "namespace in title",
			namespace:     "com.example.api",
			expectedTitle: "com.example.api API",
		},
		{
			name:          "default title when no namespace",
			namespace:     "",
			expectedTitle: "Generated API",
		},
		{
			name:          "simple namespace",
			namespace:     "users",
			expectedTitle: "users API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewOpenAPIGenerator()
			schema := &ast.Schema{
				Namespace: tt.namespace,
				Enums:     []*ast.Enum{},
				Types:     []*ast.Type{},
				Services:  []*ast.Service{},
			}

			output := gen.Generate(schema)

			// Parse the YAML output to check the title
			var spec map[string]interface{}
			err := yaml.Unmarshal([]byte(output), &spec)
			if err != nil {
				t.Fatalf("Failed to parse OpenAPI YAML: %v", err)
			}

			info, ok := spec["info"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected 'info' field in OpenAPI spec")
			}

			title, ok := info["title"].(string)
			if !ok {
				t.Fatal("Expected 'title' field in info")
			}

			if title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
			}
		})
	}
}

func TestOpenAPIGenerator_NameAnnotation_References(t *testing.T) {
	// Test that references to types with custom names use the custom names
	schema := &ast.Schema{
		Namespace: "com.example.api",
		Types: []*ast.Type{
			{
				Name: "User",
				Annotations: &ast.FormatAnnotations{
					OpenAPIName: "UserProfile",
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
			{
				Name: "Product",
				Annotations: &ast.FormatAnnotations{
					OpenAPIName: "ProductItem",
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
			{
				Name: "GetUserResponse",
				Fields: []*ast.Field{
					{
						Name: "user",
						Type: &ast.FieldType{
							Name:      "User",
							IsBuiltin: false,
						},
						Required: true,
					},
				},
			},
			{
				Name: "Order",
				Fields: []*ast.Field{
					{
						Name: "products",
						Type: &ast.FieldType{
							Name:      "Product",
							IsBuiltin: false,
							IsArray:   true,
						},
						Required: true,
					},
				},
			},
		},
	}

	g := NewOpenAPIGenerator()
	output := g.Generate(schema)

	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		t.Fatalf("Failed to parse OpenAPI output: %v", err)
	}

	// Check that User was renamed to UserProfile in schemas
	if _, ok := spec.Components.Schemas["UserProfile"]; !ok {
		t.Error("Expected 'UserProfile' schema to exist")
	}
	if _, ok := spec.Components.Schemas["User"]; ok {
		t.Error("Expected 'User' schema to NOT exist (should be UserProfile)")
	}

	// Check that Product was renamed to ProductItem in schemas
	if _, ok := spec.Components.Schemas["ProductItem"]; !ok {
		t.Error("Expected 'ProductItem' schema to exist")
	}
	if _, ok := spec.Components.Schemas["Product"]; ok {
		t.Error("Expected 'Product' schema to NOT exist (should be ProductItem)")
	}

	// Check that GetUserResponse references UserProfile, not User
	getUserResponse, ok := spec.Components.Schemas["GetUserResponse"]
	if !ok {
		t.Fatal("Expected 'GetUserResponse' schema to exist")
	}

	userField, ok := getUserResponse.Properties["user"]
	if !ok {
		t.Fatal("Expected 'user' field in GetUserResponse")
	}

	expectedRef := "#/components/schemas/UserProfile"
	if userField.Ref != expectedRef {
		t.Errorf("Expected user field to reference %q, got %q", expectedRef, userField.Ref)
	}

	// Check that Order references ProductItem in array, not Product
	order, ok := spec.Components.Schemas["Order"]
	if !ok {
		t.Fatal("Expected 'Order' schema to exist")
	}

	productsField, ok := order.Properties["products"]
	if !ok {
		t.Fatal("Expected 'products' field in Order")
	}

	if productsField.Type != "array" {
		t.Errorf("Expected products field to be array type, got %q", productsField.Type)
	}

	if productsField.Items == nil {
		t.Fatal("Expected products field to have items")
	}

	expectedItemRef := "#/components/schemas/ProductItem"
	if productsField.Items.Ref != expectedItemRef {
		t.Errorf("Expected products items to reference %q, got %q", expectedItemRef, productsField.Items.Ref)
	}
}

func TestGenerateOptionalFieldsOpenAPI(t *testing.T) {
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

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	// Check that required field is in required array
	if !strings.Contains(output, "required:") {
		t.Error("Expected 'required:' section in OpenAPI output")
	}
	if !strings.Contains(output, "- id") {
		t.Error("Expected 'id' to be in required array")
	}

	// Optional fields should not be in required array
	if strings.Contains(output, "- name") {
		t.Error("Optional field 'name' should not be in required array")
	}

	// Explicitly optional should override @required
	if strings.Contains(output, "- email") {
		t.Error("Explicitly optional field 'email' should not be in required array even with @required")
	}
}

func TestOpenAPIGenerator_MapTypes(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		name                          string
		field                         *ast.Field
		expectedType                  string
		checkAdditionalProps          bool
		expectedAdditionalPropsType   string
		expectedAdditionalPropsFormat string
		expectedAdditionalPropsRef    string
	}{
		{
			name: "map with string values",
			field: &ast.Field{
				Name: "metadata",
				Type: &ast.FieldType{
					IsMap:    true,
					MapKey:   "string",
					MapValue: "string",
				},
			},
			expectedType:                "object",
			checkAdditionalProps:        true,
			expectedAdditionalPropsType: "string",
		},
		{
			name: "map with int32 values",
			field: &ast.Field{
				Name: "scores",
				Type: &ast.FieldType{
					IsMap:    true,
					MapKey:   "string",
					MapValue: "int32",
				},
			},
			expectedType:                  "object",
			checkAdditionalProps:          true,
			expectedAdditionalPropsType:   "integer",
			expectedAdditionalPropsFormat: "int32",
		},
		{
			name: "map with int64 values",
			field: &ast.Field{
				Name: "counters",
				Type: &ast.FieldType{
					IsMap:    true,
					MapKey:   "string",
					MapValue: "int64",
				},
			},
			expectedType:                  "object",
			checkAdditionalProps:          true,
			expectedAdditionalPropsType:   "integer",
			expectedAdditionalPropsFormat: "int64",
		},
		{
			name: "map with custom type values",
			field: &ast.Field{
				Name: "users",
				Type: &ast.FieldType{
					IsMap:    true,
					MapKey:   "string",
					MapValue: "User",
				},
			},
			expectedType:               "object",
			checkAdditionalProps:       true,
			expectedAdditionalPropsRef: "#/components/schemas/User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.convertFieldToProperty(tt.field, make(map[string]string))

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %q, got %q", tt.expectedType, result.Type)
			}

			if !strings.Contains(result.Description, "Map of") {
				t.Errorf("Expected description to mention 'Map of', got %q", result.Description)
			}

			if tt.checkAdditionalProps {
				if result.AdditionalProperties == nil {
					t.Fatal("Expected additionalProperties to be set")
				}

				if tt.expectedAdditionalPropsType != "" && result.AdditionalProperties.Type != tt.expectedAdditionalPropsType {
					t.Errorf("Expected additionalProperties type %q, got %q", tt.expectedAdditionalPropsType, result.AdditionalProperties.Type)
				}

				if tt.expectedAdditionalPropsFormat != "" && result.AdditionalProperties.Format != tt.expectedAdditionalPropsFormat {
					t.Errorf("Expected additionalProperties format %q, got %q", tt.expectedAdditionalPropsFormat, result.AdditionalProperties.Format)
				}

				if tt.expectedAdditionalPropsRef != "" && result.AdditionalProperties.Ref != tt.expectedAdditionalPropsRef {
					t.Errorf("Expected additionalProperties ref %q, got %q", tt.expectedAdditionalPropsRef, result.AdditionalProperties.Ref)
				}

				// When using ref, type and format should be empty
				if tt.expectedAdditionalPropsRef != "" {
					if result.AdditionalProperties.Type != "" {
						t.Error("Expected type to be empty when using ref")
					}
					if result.AdditionalProperties.Format != "" {
						t.Error("Expected format to be empty when using ref")
					}
				}
			}
		})
	}
}

func TestOpenAPIGenerator_MapTypes_Integration(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "name",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
				},
			},
			{
				Name: "Department",
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
							MapValue: "int64",
						},
					},
					{
						Name: "users",
						Type: &ast.FieldType{
							IsMap:    true,
							MapKey:   "string",
							MapValue: "User",
						},
					},
				},
			},
		},
	}

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	// Check for proper map descriptions
	if !strings.Contains(output, "Map of string to string") {
		t.Error("Expected 'Map of string to string' in output")
	}

	if !strings.Contains(output, "Map of string to int64") {
		t.Error("Expected 'Map of string to int64' in output")
	}

	if !strings.Contains(output, "Map of string to User") {
		t.Error("Expected 'Map of string to User' in output")
	}

	// Check for additionalProperties with correct types
	if !strings.Contains(output, "additionalProperties:") {
		t.Error("Expected additionalProperties in output")
	}

	// Verify the YAML structure can be parsed
	var spec OpenAPISpec
	err := yaml.Unmarshal([]byte(output), &spec)
	if err != nil {
		t.Fatalf("Failed to parse generated YAML: %v", err)
	}

	// Check Department schema
	dept, ok := spec.Components.Schemas["Department"]
	if !ok {
		t.Fatal("Expected Department schema to be present")
	}

	// Check metadata field
	metadata := dept.Properties["metadata"]
	if metadata.Type != "object" {
		t.Errorf("Expected metadata type to be 'object', got %q", metadata.Type)
	}
	if metadata.AdditionalProperties == nil {
		t.Fatal("Expected metadata to have additionalProperties")
	}
	if metadata.AdditionalProperties.Type != "string" {
		t.Errorf("Expected metadata additionalProperties type to be 'string', got %q", metadata.AdditionalProperties.Type)
	}

	// Check scores field
	scores := dept.Properties["scores"]
	if scores.AdditionalProperties == nil {
		t.Fatal("Expected scores to have additionalProperties")
	}
	if scores.AdditionalProperties.Type != "integer" {
		t.Errorf("Expected scores additionalProperties type to be 'integer', got %q", scores.AdditionalProperties.Type)
	}
	if scores.AdditionalProperties.Format != "int64" {
		t.Errorf("Expected scores additionalProperties format to be 'int64', got %q", scores.AdditionalProperties.Format)
	}

	// Check users field
	users := dept.Properties["users"]
	if users.AdditionalProperties == nil {
		t.Fatal("Expected users to have additionalProperties")
	}
	if users.AdditionalProperties.Ref != "#/components/schemas/User" {
		t.Errorf("Expected users additionalProperties ref to be '#/components/schemas/User', got %q", users.AdditionalProperties.Ref)
	}
}

func TestOpenAPIGenerator_NestedMaps(t *testing.T) {
	gen := NewOpenAPIGenerator()

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

	// Parse the YAML output
	var spec OpenAPISpec
	err := yaml.Unmarshal([]byte(output), &spec)
	if err != nil {
		t.Fatalf("Failed to parse OpenAPI YAML: %v", err)
	}

	// Check NestedMapTest schema exists
	nestedMapTest, exists := spec.Components.Schemas["NestedMapTest"]
	if !exists {
		t.Fatal("Expected NestedMapTest schema to exist")
	}

	// Check simpleMap field - should have additionalProperties with type: string
	simpleMap := nestedMapTest.Properties["simpleMap"]
	if simpleMap.Type != "object" {
		t.Errorf("Expected simpleMap type to be 'object', got %q", simpleMap.Type)
	}
	if simpleMap.AdditionalProperties == nil {
		t.Fatal("Expected simpleMap to have additionalProperties")
	}
	if simpleMap.AdditionalProperties.Type != "string" {
		t.Errorf("Expected simpleMap additionalProperties type to be 'string', got %q", simpleMap.AdditionalProperties.Type)
	}

	// Check nestedMap field - should have nested additionalProperties
	nestedMap := nestedMapTest.Properties["nestedMap"]
	if nestedMap.Type != "object" {
		t.Errorf("Expected nestedMap type to be 'object', got %q", nestedMap.Type)
	}
	if nestedMap.AdditionalProperties == nil {
		t.Fatal("Expected nestedMap to have additionalProperties")
	}
	if nestedMap.AdditionalProperties.Type != "object" {
		t.Errorf("Expected nestedMap additionalProperties type to be 'object', got %q", nestedMap.AdditionalProperties.Type)
	}
	// Check inner additionalProperties
	if nestedMap.AdditionalProperties.AdditionalProperties == nil {
		t.Fatal("Expected nestedMap to have nested additionalProperties")
	}
	if nestedMap.AdditionalProperties.AdditionalProperties.Type != "integer" {
		t.Errorf("Expected nested additionalProperties type to be 'integer', got %q", nestedMap.AdditionalProperties.AdditionalProperties.Type)
	}
	if nestedMap.AdditionalProperties.AdditionalProperties.Format != "int32" {
		t.Errorf("Expected nested additionalProperties format to be 'int32', got %q", nestedMap.AdditionalProperties.AdditionalProperties.Format)
	}

	// Check tripleNestedMap field - should have three levels of nested additionalProperties
	tripleNestedMap := nestedMapTest.Properties["tripleNestedMap"]
	if tripleNestedMap.Type != "object" {
		t.Errorf("Expected tripleNestedMap type to be 'object', got %q", tripleNestedMap.Type)
	}
	if tripleNestedMap.AdditionalProperties == nil {
		t.Fatal("Expected tripleNestedMap to have additionalProperties")
	}
	// Level 1
	if tripleNestedMap.AdditionalProperties.Type != "object" {
		t.Errorf("Expected level 1 additionalProperties type to be 'object', got %q", tripleNestedMap.AdditionalProperties.Type)
	}
	if tripleNestedMap.AdditionalProperties.AdditionalProperties == nil {
		t.Fatal("Expected level 2 additionalProperties")
	}
	// Level 2
	if tripleNestedMap.AdditionalProperties.AdditionalProperties.Type != "object" {
		t.Errorf("Expected level 2 additionalProperties type to be 'object', got %q", tripleNestedMap.AdditionalProperties.AdditionalProperties.Type)
	}
	if tripleNestedMap.AdditionalProperties.AdditionalProperties.AdditionalProperties == nil {
		t.Fatal("Expected level 3 additionalProperties")
	}
	// Level 3
	if tripleNestedMap.AdditionalProperties.AdditionalProperties.AdditionalProperties.Type != "boolean" {
		t.Errorf("Expected level 3 additionalProperties type to be 'boolean', got %q", tripleNestedMap.AdditionalProperties.AdditionalProperties.AdditionalProperties.Type)
	}

	// Check descriptions
	if !strings.Contains(nestedMap.Description, "Map of string to") {
		t.Errorf("Expected nestedMap description to contain 'Map of string to', got: %s", nestedMap.Description)
	}
}
