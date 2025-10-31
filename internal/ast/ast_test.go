package ast

import (
	"testing"
)

func TestIsBuiltinType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"string is builtin", "string", true},
		{"int32 is builtin", "int32", true},
		{"int64 is builtin", "int64", true},
		{"float32 is builtin", "float32", true},
		{"float64 is builtin", "float64", true},
		{"bool is builtin", "bool", true},
		{"timestamp is builtin", "timestamp", true},
		{"bytes is builtin", "bytes", true},
		{"User is not builtin", "User", false},
		{"CustomType is not builtin", "CustomType", false},
		{"empty string is not builtin", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBuiltinType(tt.typeName); got != tt.want {
				t.Errorf("IsBuiltinType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestSchema(t *testing.T) {
	schema := &Schema{
		Enums:    []*Enum{},
		Types:    []*Type{},
		Services: []*Service{},
	}

	if schema.Enums == nil {
		t.Error("Expected Enums to be initialized")
	}
	if schema.Types == nil {
		t.Error("Expected Types to be initialized")
	}
	if schema.Services == nil {
		t.Error("Expected Services to be initialized")
	}
}

func TestEnum(t *testing.T) {
	enum := &Enum{
		Name: "UserRole",
		Values: []*EnumValue{
			{Name: "ADMIN"},
			{Name: "USER"},
			{Name: "GUEST"},
		},
	}

	if enum.Name != "UserRole" {
		t.Errorf("Expected Name to be 'UserRole', got %q", enum.Name)
	}
	if len(enum.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(enum.Values))
	}
	if enum.Values[0].Name != "ADMIN" {
		t.Errorf("Expected first value to be 'ADMIN', got %q", enum.Values[0].Name)
	}
}

func TestType(t *testing.T) {
	typ := &Type{
		Name: "User",
		Fields: []*Field{
			{
				Name: "id",
				Type: &FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
				Required: true,
			},
		},
	}

	if typ.Name != "User" {
		t.Errorf("Expected Name to be 'User', got %q", typ.Name)
	}
	if len(typ.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(typ.Fields))
	}
	if typ.Fields[0].Name != "id" {
		t.Errorf("Expected field name to be 'id', got %q", typ.Fields[0].Name)
	}
}

func TestField(t *testing.T) {
	field := &Field{
		Name: "email",
		Type: &FieldType{
			Name:      "string",
			IsBuiltin: true,
		},
		Required: true,
		Default:  "",
		Attributes: map[string]string{
			"required": "",
		},
	}

	if field.Name != "email" {
		t.Errorf("Expected Name to be 'email', got %q", field.Name)
	}
	if !field.Required {
		t.Error("Expected Required to be true")
	}
	if field.Type.Name != "string" {
		t.Errorf("Expected Type.Name to be 'string', got %q", field.Type.Name)
	}
	if _, ok := field.Attributes["required"]; !ok {
		t.Error("Expected 'required' attribute to exist")
	}
}

func TestFieldType(t *testing.T) {
	t.Run("simple type", func(t *testing.T) {
		ft := &FieldType{
			Name:      "string",
			IsBuiltin: true,
			IsArray:   false,
			IsMap:     false,
		}

		if ft.Name != "string" {
			t.Errorf("Expected Name to be 'string', got %q", ft.Name)
		}
		if !ft.IsBuiltin {
			t.Error("Expected IsBuiltin to be true")
		}
		if ft.IsArray {
			t.Error("Expected IsArray to be false")
		}
		if ft.IsMap {
			t.Error("Expected IsMap to be false")
		}
	})

	t.Run("array type", func(t *testing.T) {
		ft := &FieldType{
			Name:      "string",
			IsBuiltin: true,
			IsArray:   true,
			IsMap:     false,
		}

		if !ft.IsArray {
			t.Error("Expected IsArray to be true")
		}
	})

	t.Run("map type", func(t *testing.T) {
		ft := &FieldType{
			Name:     "map",
			IsMap:    true,
			MapKey:   "string",
			MapValue: "int32",
		}

		if !ft.IsMap {
			t.Error("Expected IsMap to be true")
		}
		if ft.MapKey != "string" {
			t.Errorf("Expected MapKey to be 'string', got %q", ft.MapKey)
		}
		if ft.MapValue != "int32" {
			t.Errorf("Expected MapValue to be 'int32', got %q", ft.MapValue)
		}
	})

	t.Run("custom type", func(t *testing.T) {
		ft := &FieldType{
			Name:      "User",
			IsBuiltin: false,
		}

		if ft.IsBuiltin {
			t.Error("Expected IsBuiltin to be false")
		}
	})
}

func TestService(t *testing.T) {
	service := &Service{
		Name: "UserService",
		Methods: []*Method{
			{
				Name:       "GetUser",
				InputType:  "GetUserRequest",
				OutputType: "GetUserResponse",
			},
		},
	}

	if service.Name != "UserService" {
		t.Errorf("Expected Name to be 'UserService', got %q", service.Name)
	}
	if len(service.Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(service.Methods))
	}
	if service.Methods[0].Name != "GetUser" {
		t.Errorf("Expected method name to be 'GetUser', got %q", service.Methods[0].Name)
	}
}

func TestMethod(t *testing.T) {
	method := &Method{
		Name:       "CreateUser",
		InputType:  "CreateUserRequest",
		OutputType: "CreateUserResponse",
	}

	if method.Name != "CreateUser" {
		t.Errorf("Expected Name to be 'CreateUser', got %q", method.Name)
	}
	if method.InputType != "CreateUserRequest" {
		t.Errorf("Expected InputType to be 'CreateUserRequest', got %q", method.InputType)
	}
	if method.OutputType != "CreateUserResponse" {
		t.Errorf("Expected OutputType to be 'CreateUserResponse', got %q", method.OutputType)
	}
}

func TestBuiltinTypes(t *testing.T) {
	expectedTypes := []string{
		"string", "int32", "int64", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "timestamp", "bytes",
	}

	for _, typeName := range expectedTypes {
		if !BuiltinTypes[typeName] {
			t.Errorf("Expected %q to be in BuiltinTypes", typeName)
		}
	}

	if len(BuiltinTypes) != len(expectedTypes) {
		t.Errorf("Expected %d builtin types, got %d", len(expectedTypes), len(BuiltinTypes))
	}
}

func TestDocumentation_GetDoc(t *testing.T) {
	tests := []struct {
		name     string
		doc      *Documentation
		lang     string
		expected string
	}{
		{
			name:     "nil documentation",
			doc:      nil,
			lang:     "proto",
			expected: "",
		},
		{
			name: "general documentation only",
			doc: &Documentation{
				General: "General description",
			},
			lang:     "proto",
			expected: "General description",
		},
		{
			name: "specific documentation for proto",
			doc: &Documentation{
				General: "General description",
				Specific: map[string]string{
					"proto": "Proto-specific description",
				},
			},
			lang:     "proto",
			expected: "Proto-specific description",
		},
		{
			name: "fallback to general when specific not found",
			doc: &Documentation{
				General: "General description",
				Specific: map[string]string{
					"graphql": "GraphQL-specific description",
				},
			},
			lang:     "proto",
			expected: "General description",
		},
		{
			name: "multiple specific languages",
			doc: &Documentation{
				General: "General description",
				Specific: map[string]string{
					"proto":   "Proto-specific",
					"graphql": "GraphQL-specific",
					"openapi": "OpenAPI-specific",
				},
			},
			lang:     "graphql",
			expected: "GraphQL-specific",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.doc.GetDoc(tt.lang)
			if result != tt.expected {
				t.Errorf("GetDoc(%q) = %q, want %q", tt.lang, result, tt.expected)
			}
		})
	}
}

func TestField_ShouldIncludeInGenerator(t *testing.T) {
	tests := []struct {
		name      string
		field     *Field
		generator string
		expected  bool
	}{
		{
			name: "no restrictions",
			field: &Field{
				Name: "field1",
			},
			generator: "proto",
			expected:  true,
		},
		{
			name: "excluded from proto",
			field: &Field{
				Name:        "field1",
				ExcludeFrom: []string{"proto"},
			},
			generator: "proto",
			expected:  false,
		},
		{
			name: "excluded from proto but checking graphql",
			field: &Field{
				Name:        "field1",
				ExcludeFrom: []string{"proto"},
			},
			generator: "graphql",
			expected:  true,
		},
		{
			name: "excluded from multiple generators",
			field: &Field{
				Name:        "field1",
				ExcludeFrom: []string{"proto", "graphql"},
			},
			generator: "graphql",
			expected:  false,
		},
		{
			name: "only for proto",
			field: &Field{
				Name:    "field1",
				OnlyFor: []string{"proto"},
			},
			generator: "proto",
			expected:  true,
		},
		{
			name: "only for proto but checking graphql",
			field: &Field{
				Name:    "field1",
				OnlyFor: []string{"proto"},
			},
			generator: "graphql",
			expected:  false,
		},
		{
			name: "only for multiple generators",
			field: &Field{
				Name:    "field1",
				OnlyFor: []string{"proto", "openapi"},
			},
			generator: "openapi",
			expected:  true,
		},
		{
			name: "only for multiple generators but not this one",
			field: &Field{
				Name:    "field1",
				OnlyFor: []string{"proto", "openapi"},
			},
			generator: "graphql",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.ShouldIncludeInGenerator(tt.generator)
			if result != tt.expected {
				t.Errorf("ShouldIncludeInGenerator(%q) = %v, want %v", tt.generator, result, tt.expected)
			}
		})
	}
}

func TestMethod_GetHTTPMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   *Method
		expected string
	}{
		{
			name: "explicit POST",
			method: &Method{
				Name:       "CreateUser",
				HTTPMethod: "POST",
			},
			expected: "post",
		},
		{
			name: "explicit GET",
			method: &Method{
				Name:       "GetUser",
				HTTPMethod: "GET",
			},
			expected: "get",
		},
		{
			name: "explicit DELETE",
			method: &Method{
				Name:       "RemoveUser",
				HTTPMethod: "DELETE",
			},
			expected: "delete",
		},
		{
			name: "heuristic Get prefix",
			method: &Method{
				Name: "GetUser",
			},
			expected: "get",
		},
		{
			name: "heuristic List prefix",
			method: &Method{
				Name: "ListUsers",
			},
			expected: "get",
		},
		{
			name: "heuristic default to POST",
			method: &Method{
				Name: "CreateUser",
			},
			expected: "post",
		},
		{
			name: "heuristic default to POST for Update",
			method: &Method{
				Name: "UpdateUser",
			},
			expected: "post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method.GetHTTPMethod()
			if result != tt.expected {
				t.Errorf("GetHTTPMethod() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMethod_GetGraphQLType(t *testing.T) {
	tests := []struct {
		name     string
		method   *Method
		expected string
	}{
		{
			name: "explicit query",
			method: &Method{
				Name:        "CreateUser",
				GraphQLType: "query",
			},
			expected: "query",
		},
		{
			name: "explicit mutation",
			method: &Method{
				Name:        "GetUser",
				GraphQLType: "mutation",
			},
			expected: "mutation",
		},
		{
			name: "explicit subscription",
			method: &Method{
				Name:        "CreateUser",
				GraphQLType: "subscription",
			},
			expected: "subscription",
		},
		{
			name: "heuristic Get prefix",
			method: &Method{
				Name: "GetUser",
			},
			expected: "query",
		},
		{
			name: "heuristic List prefix",
			method: &Method{
				Name: "ListUsers",
			},
			expected: "query",
		},
		{
			name: "heuristic default to mutation",
			method: &Method{
				Name: "CreateUser",
			},
			expected: "mutation",
		},
		{
			name: "heuristic default to mutation for Update",
			method: &Method{
				Name: "UpdateUser",
			},
			expected: "mutation",
		},
		{
			name: "heuristic stream output is subscription",
			method: &Method{
				Name:         "WatchMessages",
				OutputStream: true,
			},
			expected: "subscription",
		},
		{
			name: "heuristic stream output overrides Get prefix",
			method: &Method{
				Name:         "GetUpdates",
				OutputStream: true,
			},
			expected: "subscription",
		},
		{
			name: "heuristic stream input does not affect type",
			method: &Method{
				Name:        "UploadData",
				InputStream: true,
			},
			expected: "mutation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method.GetGraphQLType()
			if result != tt.expected {
				t.Errorf("GetGraphQLType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTypeRegistry_NewTypeRegistry(t *testing.T) {
	registry := NewTypeRegistry()

	if registry == nil {
		t.Fatal("NewTypeRegistry should return non-nil registry")
	}

	if registry.Types == nil {
		t.Error("TypeRegistry.Types should be initialized")
	}

	if registry.Enums == nil {
		t.Error("TypeRegistry.Enums should be initialized")
	}

	if registry.Unions == nil {
		t.Error("TypeRegistry.Unions should be initialized")
	}
}

func TestTypeRegistry_RegisterType(t *testing.T) {
	registry := NewTypeRegistry()

	typ := &Type{
		Name:      "User",
		Namespace: "com.example.api",
	}

	registry.RegisterType(typ)

	qualName := "com.example.api.User"
	if registry.Types[qualName] != typ {
		t.Errorf("Type not registered with qualified name %q", qualName)
	}
}

func TestTypeRegistry_RegisterEnum(t *testing.T) {
	registry := NewTypeRegistry()

	enum := &Enum{
		Name:      "Status",
		Namespace: "com.example.api",
	}

	registry.RegisterEnum(enum)

	qualName := "com.example.api.Status"
	if registry.Enums[qualName] != enum {
		t.Errorf("Enum not registered with qualified name %q", qualName)
	}
}

func TestTypeRegistry_RegisterUnion(t *testing.T) {
	registry := NewTypeRegistry()

	union := &Union{
		Name:      "Result",
		Namespace: "com.example.api",
	}

	registry.RegisterUnion(union)

	qualName := "com.example.api.Result"
	if registry.Unions[qualName] != union {
		t.Errorf("Union not registered with qualified name %q", qualName)
	}
}

func TestTypeRegistry_ResolveType(t *testing.T) {
	registry := NewTypeRegistry()

	// Register types in different namespaces
	userType := &Type{
		Name:      "User",
		Namespace: "com.example.users",
	}
	orderType := &Type{
		Name:      "Order",
		Namespace: "com.example.orders",
	}

	registry.RegisterType(userType)
	registry.RegisterType(orderType)

	tests := []struct {
		name             string
		typeName         string
		currentNamespace string
		expectedResolved string
		shouldFind       bool
	}{
		{
			name:             "qualified name - exact match",
			typeName:         "com.example.users.User",
			currentNamespace: "com.example.orders",
			expectedResolved: "com.example.users.User",
			shouldFind:       true,
		},
		{
			name:             "unqualified name - current namespace",
			typeName:         "Order",
			currentNamespace: "com.example.orders",
			expectedResolved: "com.example.orders.Order",
			shouldFind:       true,
		},
		{
			name:             "unqualified name - search other namespace",
			typeName:         "User",
			currentNamespace: "com.example.orders",
			expectedResolved: "com.example.users.User",
			shouldFind:       true,
		},
		{
			name:             "non-existent type",
			typeName:         "Product",
			currentNamespace: "com.example.orders",
			expectedResolved: "Product",
			shouldFind:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, found := registry.ResolveType(tt.typeName, tt.currentNamespace)
			if found != tt.shouldFind {
				t.Errorf("Expected found=%v, got %v", tt.shouldFind, found)
			}
			if resolved != tt.expectedResolved {
				t.Errorf("Expected resolved=%q, got %q", tt.expectedResolved, resolved)
			}
		})
	}
}

func TestGetUnqualifiedName(t *testing.T) {
	tests := []struct {
		name          string
		qualifiedName string
		expected      string
	}{
		{
			name:          "fully qualified name",
			qualifiedName: "com.example.users.User",
			expected:      "User",
		},
		{
			name:          "simple qualified name",
			qualifiedName: "api.User",
			expected:      "User",
		},
		{
			name:          "unqualified name",
			qualifiedName: "User",
			expected:      "User",
		},
		{
			name:          "empty string",
			qualifiedName: "",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUnqualifiedName(tt.qualifiedName)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
