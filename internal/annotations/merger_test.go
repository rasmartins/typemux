package annotations

import (
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func createTestSchemaForMerger() *ast.Schema {
	return &ast.Schema{
		Types: []*ast.Type{
			{
				Name:        "User",
				Namespace:   "com.example.api",
				Annotations: ast.NewFormatAnnotations(),
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}, Required: false, Annotations: ast.NewFormatAnnotations()},
					{Name: "username", Type: &ast.FieldType{Name: "string", IsBuiltin: true}, Required: false, Annotations: ast.NewFormatAnnotations()},
					{Name: "email", Type: &ast.FieldType{Name: "string", IsBuiltin: true}, Required: false, Annotations: ast.NewFormatAnnotations()},
				},
			},
			{
				Name:        "User",
				Namespace:   "com.example.admin",
				Annotations: ast.NewFormatAnnotations(),
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string", IsBuiltin: true}, Required: false, Annotations: ast.NewFormatAnnotations()},
					{Name: "adminLevel", Type: &ast.FieldType{Name: "int32", IsBuiltin: true}, Required: false, Annotations: ast.NewFormatAnnotations()},
				},
			},
		},
		Enums: []*ast.Enum{
			{
				Name:        "UserStatus",
				Namespace:   "com.example.api",
				Annotations: ast.NewFormatAnnotations(),
				Values: []*ast.EnumValue{
					{Name: "ACTIVE"},
					{Name: "INACTIVE"},
				},
			},
		},
		Unions: []*ast.Union{
			{
				Name:        "SearchResult",
				Namespace:   "com.example.api",
				Annotations: ast.NewFormatAnnotations(),
				Options:     []string{"User", "Product"},
			},
		},
		Services: []*ast.Service{
			{
				Name:      "UserService",
				Namespace: "com.example.api",
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
}

func TestMerger_TypeAnnotations(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Proto:   &FormatSpecificAnnotations{Name: "UserV2"},
				GraphQL: &FormatSpecificAnnotations{Name: "UserAccount"},
				OpenAPI: &FormatSpecificAnnotations{Name: "UserProfile"},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	// Find the first User type (should match by simple name)
	userType := schema.Types[0]

	if userType.Annotations.ProtoName != "UserV2" {
		t.Errorf("Expected ProtoName 'UserV2', got '%s'", userType.Annotations.ProtoName)
	}
	if userType.Annotations.GraphQLName != "UserAccount" {
		t.Errorf("Expected GraphQLName 'UserAccount', got '%s'", userType.Annotations.GraphQLName)
	}
	if userType.Annotations.OpenAPIName != "UserProfile" {
		t.Errorf("Expected OpenAPIName 'UserProfile', got '%s'", userType.Annotations.OpenAPIName)
	}
}

func TestMerger_QualifiedTypeName(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"com.example.api.User": {
				Proto:   &FormatSpecificAnnotations{Name: "ApiUser"},
				GraphQL: &FormatSpecificAnnotations{Name: "ApiUserAccount"},
			},
			"com.example.admin.User": {
				Proto:   &FormatSpecificAnnotations{Name: "AdminUser"},
				GraphQL: &FormatSpecificAnnotations{Name: "AdminUserAccount"},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	// Check first User type (com.example.api)
	apiUser := schema.Types[0]
	if apiUser.Annotations.ProtoName != "ApiUser" {
		t.Errorf("Expected ApiUser ProtoName 'ApiUser', got '%s'", apiUser.Annotations.ProtoName)
	}
	if apiUser.Annotations.GraphQLName != "ApiUserAccount" {
		t.Errorf("Expected ApiUser GraphQLName 'ApiUserAccount', got '%s'", apiUser.Annotations.GraphQLName)
	}

	// Check second User type (com.example.admin)
	adminUser := schema.Types[1]
	if adminUser.Annotations.ProtoName != "AdminUser" {
		t.Errorf("Expected AdminUser ProtoName 'AdminUser', got '%s'", adminUser.Annotations.ProtoName)
	}
	if adminUser.Annotations.GraphQLName != "AdminUserAccount" {
		t.Errorf("Expected AdminUser GraphQLName 'AdminUserAccount', got '%s'", adminUser.Annotations.GraphQLName)
	}
}

func TestMerger_FieldAnnotations(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Fields: map[string]*FieldAnnotations{
					"email": {
						Required: true,
						Proto:    &FormatSpecificAnnotations{Name: "email_address"},
						OpenAPI:  &FormatSpecificAnnotations{Extension: `{"x-format": "email"}`},
					},
				},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	// Find email field
	userType := schema.Types[0]
	var emailField *ast.Field
	for _, field := range userType.Fields {
		if field.Name == "email" {
			emailField = field
			break
		}
	}

	if emailField == nil {
		t.Fatal("email field not found")
	}

	if !emailField.Required {
		t.Error("Expected email field to be required")
	}

	if emailField.Annotations.ProtoName != "email_address" {
		t.Errorf("Expected ProtoName 'email_address', got '%s'", emailField.Annotations.ProtoName)
	}

	if len(emailField.Annotations.OpenAPI) != 1 {
		t.Fatalf("Expected 1 OpenAPI annotation, got %d", len(emailField.Annotations.OpenAPI))
	}

	if emailField.Annotations.OpenAPI[0] != `{"x-format": "email"}` {
		t.Errorf("Expected OpenAPI extension, got '%s'", emailField.Annotations.OpenAPI[0])
	}
}

func TestMerger_FieldExcludeAndOnly(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Fields: map[string]*FieldAnnotations{
					"username": {
						Exclude: []string{"proto", "graphql"},
					},
					"email": {
						Only: []string{"openapi"},
					},
				},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	userType := schema.Types[0]

	// Check username exclude
	var usernameField *ast.Field
	for _, field := range userType.Fields {
		if field.Name == "username" {
			usernameField = field
			break
		}
	}

	if usernameField == nil {
		t.Fatal("username field not found")
	}

	if len(usernameField.ExcludeFrom) != 2 {
		t.Errorf("Expected 2 exclude items, got %d", len(usernameField.ExcludeFrom))
	}

	// Check email only
	var emailField *ast.Field
	for _, field := range userType.Fields {
		if field.Name == "email" {
			emailField = field
			break
		}
	}

	if emailField == nil {
		t.Fatal("email field not found")
	}

	if len(emailField.OnlyFor) != 1 {
		t.Errorf("Expected 1 only item, got %d", len(emailField.OnlyFor))
	}
}

func TestMerger_EnumAnnotations(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Enums: map[string]*EnumAnnotations{
			"UserStatus": {
				Proto:   &FormatSpecificAnnotations{Name: "UserStatusEnum"},
				GraphQL: &FormatSpecificAnnotations{Name: "UserStatusType"},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	enumType := schema.Enums[0]

	if enumType.Annotations.ProtoName != "UserStatusEnum" {
		t.Errorf("Expected ProtoName 'UserStatusEnum', got '%s'", enumType.Annotations.ProtoName)
	}
	if enumType.Annotations.GraphQLName != "UserStatusType" {
		t.Errorf("Expected GraphQLName 'UserStatusType', got '%s'", enumType.Annotations.GraphQLName)
	}
}

func TestMerger_QualifiedEnumName(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Enums: map[string]*EnumAnnotations{
			"com.example.api.UserStatus": {
				Proto: &FormatSpecificAnnotations{Name: "ApiUserStatus"},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	enumType := schema.Enums[0]

	if enumType.Annotations.ProtoName != "ApiUserStatus" {
		t.Errorf("Expected ProtoName 'ApiUserStatus', got '%s'", enumType.Annotations.ProtoName)
	}
}

func TestMerger_UnionAnnotations(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Unions: map[string]*UnionAnnotations{
			"SearchResult": {
				Proto:   &FormatSpecificAnnotations{Name: "SearchResultUnion"},
				GraphQL: &FormatSpecificAnnotations{Name: "SearchResultType"},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	unionType := schema.Unions[0]

	if unionType.Annotations.ProtoName != "SearchResultUnion" {
		t.Errorf("Expected ProtoName 'SearchResultUnion', got '%s'", unionType.Annotations.ProtoName)
	}
	if unionType.Annotations.GraphQLName != "SearchResultType" {
		t.Errorf("Expected GraphQLName 'SearchResultType', got '%s'", unionType.Annotations.GraphQLName)
	}
}

func TestMerger_ServiceMethodAnnotations(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Services: map[string]*ServiceAnnotations{
			"UserService": {
				Methods: map[string]*MethodAnnotations{
					"GetUser": {
						HTTP:    "GET",
						Path:    "/api/v1/users/{id}",
						GraphQL: "query",
						Success: []int{200},
						Errors:  []int{404, 500},
					},
				},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	service := schema.Services[0]
	method := service.Methods[0]

	if method.HTTPMethod != "GET" {
		t.Errorf("Expected HTTPMethod 'GET', got '%s'", method.HTTPMethod)
	}
	if method.PathTemplate != "/api/v1/users/{id}" {
		t.Errorf("Expected PathTemplate '/api/v1/users/{id}', got '%s'", method.PathTemplate)
	}
	if method.GraphQLType != "query" {
		t.Errorf("Expected GraphQLType 'query', got '%s'", method.GraphQLType)
	}

	if len(method.SuccessCodes) != 1 {
		t.Errorf("Expected 1 success code, got %d", len(method.SuccessCodes))
	} else if method.SuccessCodes[0] != "200" {
		t.Errorf("Expected success code '200', got '%s'", method.SuccessCodes[0])
	}

	if len(method.ErrorCodes) != 2 {
		t.Errorf("Expected 2 error codes, got %d", len(method.ErrorCodes))
	}
}

func TestMerger_QualifiedServiceName(t *testing.T) {
	schema := createTestSchemaForMerger()

	annotations := &YAMLAnnotations{
		Services: map[string]*ServiceAnnotations{
			"com.example.api.UserService": {
				Methods: map[string]*MethodAnnotations{
					"GetUser": {
						HTTP: "GET",
						Path: "/api/users/{id}",
					},
				},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	service := schema.Services[0]
	method := service.Methods[0]

	if method.HTTPMethod != "GET" {
		t.Errorf("Expected HTTPMethod 'GET', got '%s'", method.HTTPMethod)
	}
	if method.PathTemplate != "/api/users/{id}" {
		t.Errorf("Expected PathTemplate '/api/users/{id}', got '%s'", method.PathTemplate)
	}
}

func TestMerger_NoAnnotations(t *testing.T) {
	schema := createTestSchemaForMerger()

	// Test with empty annotations
	annotations := &YAMLAnnotations{
		Types:    make(map[string]*TypeAnnotations),
		Enums:    make(map[string]*EnumAnnotations),
		Unions:   make(map[string]*UnionAnnotations),
		Services: make(map[string]*ServiceAnnotations),
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	// Schema should remain unchanged
	userType := schema.Types[0]
	if userType.Annotations.ProtoName != "" {
		t.Errorf("Expected no ProtoName annotation, got '%s'", userType.Annotations.ProtoName)
	}
}

func TestMerger_ListMerging(t *testing.T) {
	schema := createTestSchemaForMerger()

	// Pre-populate some exclude values
	schema.Types[0].Fields[0].ExcludeFrom = []string{"proto"}

	annotations := &YAMLAnnotations{
		Types: map[string]*TypeAnnotations{
			"User": {
				Fields: map[string]*FieldAnnotations{
					"id": {
						Exclude: []string{"graphql"},
					},
				},
			},
		},
	}

	merger := NewMerger(annotations)
	merger.Merge(schema)

	idField := schema.Types[0].Fields[0]

	// Should have both proto and graphql
	if len(idField.ExcludeFrom) != 2 {
		t.Errorf("Expected 2 exclude items, got %d", len(idField.ExcludeFrom))
	}

	excludeMap := make(map[string]bool)
	for _, item := range idField.ExcludeFrom {
		excludeMap[item] = true
	}

	if !excludeMap["proto"] {
		t.Error("Expected 'proto' in exclude list")
	}
	if !excludeMap["graphql"] {
		t.Error("Expected 'graphql' in exclude list")
	}
}

func TestMergeLists(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
		{
			name:     "a empty",
			a:        []string{},
			b:        []string{"proto", "graphql"},
			expected: []string{"proto", "graphql"},
		},
		{
			name:     "b empty",
			a:        []string{"proto", "graphql"},
			b:        []string{},
			expected: []string{"proto", "graphql"},
		},
		{
			name:     "no overlap",
			a:        []string{"proto"},
			b:        []string{"graphql"},
			expected: []string{"proto", "graphql"},
		},
		{
			name:     "with duplicates",
			a:        []string{"proto", "graphql"},
			b:        []string{"graphql", "openapi"},
			expected: []string{"proto", "graphql", "openapi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeLists(tt.a, tt.b)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
			}

			resultMap := make(map[string]bool)
			for _, item := range result {
				resultMap[item] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Expected '%s' in result", expected)
				}
			}
		})
	}
}
