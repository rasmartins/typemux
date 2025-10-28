package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestProtobufGenerator_Generate(t *testing.T) {
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

	gen := NewProtobufGenerator()
	output := gen.Generate(schema)

	// Check syntax
	if !strings.Contains(output, `syntax = "proto3"`) {
		t.Error("Expected proto3 syntax declaration")
	}

	// Check package
	if !strings.Contains(output, "package api") {
		t.Error("Expected package declaration")
	}

	// Check timestamp import
	if !strings.Contains(output, `import "google/protobuf/timestamp.proto"`) {
		t.Error("Expected timestamp import")
	}

	// Check enum
	if !strings.Contains(output, "enum UserRole") {
		t.Error("Expected enum UserRole")
	}

	// Check message
	if !strings.Contains(output, "message User") {
		t.Error("Expected message User")
	}

	// Check service
	if !strings.Contains(output, "service UserService") {
		t.Error("Expected service UserService")
	}
}

func TestProtobufGenerator_GenerateEnum(t *testing.T) {
	gen := NewProtobufGenerator()
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
		t.Error("Expected enum Status")
	}

	// Check for UNSPECIFIED value
	if !strings.Contains(output, "STATUS_UNSPECIFIED = 0") {
		t.Error("Expected UNSPECIFIED value at 0")
	}

	// Check for enum values
	if !strings.Contains(output, "ACTIVE = 1") {
		t.Error("Expected ACTIVE = 1")
	}

	if !strings.Contains(output, "INACTIVE = 2") {
		t.Error("Expected INACTIVE = 2")
	}

	if !strings.Contains(output, "PENDING = 3") {
		t.Error("Expected PENDING = 3")
	}
}

func TestProtobufGenerator_GenerateEnumWithCustomNumbers(t *testing.T) {
	gen := NewProtobufGenerator()

	tests := []struct {
		name     string
		enum     *ast.Enum
		expected map[string]int // value name -> expected number
	}{
		{
			name: "all custom numbers",
			enum: &ast.Enum{
				Name: "UserRole",
				Values: []*ast.EnumValue{
					{Name: "ADMIN", Number: 10, HasNumber: true},
					{Name: "USER", Number: 20, HasNumber: true},
					{Name: "GUEST", Number: 30, HasNumber: true},
				},
			},
			expected: map[string]int{
				"ADMIN": 10,
				"USER":  20,
				"GUEST": 30,
			},
		},
		{
			name: "mixed auto and custom numbers",
			enum: &ast.Enum{
				Name: "Status",
				Values: []*ast.EnumValue{
					{Name: "ACTIVE", Number: 1, HasNumber: true},
					{Name: "INACTIVE"}, // Should get 2 (next available)
					{Name: "PENDING", Number: 5, HasNumber: true},
					{Name: "ARCHIVED"}, // Should get 6 (next after 5)
				},
			},
			expected: map[string]int{
				"ACTIVE":   1,
				"INACTIVE": 2,
				"PENDING":  5,
				"ARCHIVED": 6,
			},
		},
		{
			name: "sparse numbering",
			enum: &ast.Enum{
				Name: "Priority",
				Values: []*ast.EnumValue{
					{Name: "LOW", Number: 100, HasNumber: true},
					{Name: "MEDIUM", Number: 200, HasNumber: true},
					{Name: "HIGH", Number: 300, HasNumber: true},
				},
			},
			expected: map[string]int{
				"LOW":    100,
				"MEDIUM": 200,
				"HIGH":   300,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := gen.generateEnum(tt.enum)

			for valueName, expectedNum := range tt.expected {
				expectedStr := valueName + " = " + string(rune('0'+expectedNum))
				if expectedNum >= 10 {
					expectedStr = valueName + " = "
					// For multi-digit numbers, check differently
					if !strings.Contains(output, expectedStr) {
						t.Errorf("Expected to find %q with number %d in output", valueName, expectedNum)
					}
				} else if !strings.Contains(output, expectedStr) {
					t.Errorf("Expected %q in output, got:\n%s", expectedStr, output)
				}
			}
		})
	}
}

func TestProtobufGenerator_GenerateMessage(t *testing.T) {
	gen := NewProtobufGenerator()
	typ := &ast.Type{
		Name: "User",
		Fields: []*ast.Field{
			{
				Name: "id",
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
			},
			{
				Name: "name",
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
			},
			{
				Name: "age",
				Type: &ast.FieldType{
					Name:      "int32",
					IsBuiltin: true,
				},
			},
		},
	}

	output := gen.generateMessage(typ)

	if !strings.Contains(output, "message User") {
		t.Error("Expected message User")
	}

	if !strings.Contains(output, "string id = 1") {
		t.Error("Expected id field at position 1")
	}

	if !strings.Contains(output, "string name = 2") {
		t.Error("Expected name field at position 2")
	}

	if !strings.Contains(output, "int32 age = 3") {
		t.Error("Expected age field at position 3")
	}
}

func TestProtobufGenerator_GenerateMessageField(t *testing.T) {
	gen := NewProtobufGenerator()

	tests := []struct {
		name     string
		field    *ast.Field
		fieldNum int
		expected string
	}{
		{
			name: "simple string field",
			field: &ast.Field{
				Name: "name",
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
				},
			},
			fieldNum: 1,
			expected: "string name = 1;",
		},
		{
			name: "int32 field",
			field: &ast.Field{
				Name: "count",
				Type: &ast.FieldType{
					Name:      "int32",
					IsBuiltin: true,
				},
			},
			fieldNum: 2,
			expected: "int32 count = 2;",
		},
		{
			name: "array field",
			field: &ast.Field{
				Name: "tags",
				Type: &ast.FieldType{
					Name:      "string",
					IsBuiltin: true,
					IsArray:   true,
				},
			},
			fieldNum: 3,
			expected: "repeated string tags = 3;",
		},
		{
			name: "map field",
			field: &ast.Field{
				Name: "metadata",
				Type: &ast.FieldType{
					IsMap:    true,
					MapKey:   "string",
					MapValue: "string",
				},
			},
			fieldNum: 4,
			expected: "map<string, string> metadata = 4;",
		},
		{
			name: "custom type field",
			field: &ast.Field{
				Name: "user",
				Type: &ast.FieldType{
					Name:      "User",
					IsBuiltin: false,
				},
			},
			fieldNum: 5,
			expected: "User user = 5;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.generateMessageField(tt.field, tt.fieldNum)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProtobufGenerator_MapTypeToProtobuf(t *testing.T) {
	gen := NewProtobufGenerator()

	tests := []struct {
		typeName string
		expected string
	}{
		{"string", "string"},
		{"int32", "int32"},
		{"int64", "int64"},
		{"float32", "float"},
		{"float64", "double"},
		{"bool", "bool"},
		{"timestamp", "google.protobuf.Timestamp"},
		{"bytes", "bytes"},
		{"User", "User"},
		{"CustomType", "CustomType"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := gen.mapScalarType(tt.typeName)
			if result != tt.expected {
				t.Errorf("mapScalarType(%q) = %q, want %q", tt.typeName, result, tt.expected)
			}
		})
	}
}

func TestProtobufGenerator_GenerateService(t *testing.T) {
	gen := NewProtobufGenerator()
	service := &ast.Service{
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
				OutputType: "GetUserResponse",
			},
		},
	}

	output := gen.generateService(service)

	if !strings.Contains(output, "service UserService") {
		t.Error("Expected service UserService")
	}

	if !strings.Contains(output, "rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)") {
		t.Error("Expected CreateUser RPC method")
	}

	if !strings.Contains(output, "rpc GetUser(GetUserRequest) returns (GetUserResponse)") {
		t.Error("Expected GetUser RPC method")
	}
}

func TestProtobufGenerator_MapField(t *testing.T) {
	gen := NewProtobufGenerator()
	field := &ast.Field{
		Name: "settings",
		Type: &ast.FieldType{
			IsMap:    true,
			MapKey:   "string",
			MapValue: "int32",
		},
	}

	result := gen.generateMessageField(field, 1)
	expected := "map<string, int32> settings = 1;"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProtobufGenerator_RepeatedCustomType(t *testing.T) {
	gen := NewProtobufGenerator()
	field := &ast.Field{
		Name: "users",
		Type: &ast.FieldType{
			Name:      "User",
			IsBuiltin: false,
			IsArray:   true,
		},
	}

	result := gen.generateMessageField(field, 1)
	expected := "repeated User users = 1;"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProtobufGenerator_TimestampField(t *testing.T) {
	gen := NewProtobufGenerator()
	field := &ast.Field{
		Name: "createdAt",
		Type: &ast.FieldType{
			Name:      "timestamp",
			IsBuiltin: true,
		},
	}

	result := gen.generateMessageField(field, 1)
	expected := "google.protobuf.Timestamp createdAt = 1;"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProtobufGenerator_EmptySchema(t *testing.T) {
	schema := &ast.Schema{
		Enums:    []*ast.Enum{},
		Types:    []*ast.Type{},
		Services: []*ast.Service{},
	}

	gen := NewProtobufGenerator()
	output := gen.Generate(schema)

	if !strings.Contains(output, `syntax = "proto3"`) {
		t.Error("Expected proto3 syntax even for empty schema")
	}

	if !strings.Contains(output, "package api") {
		t.Error("Expected package declaration even for empty schema")
	}
}

func TestProtobufGenerator_EnumUnspecifiedPrefix(t *testing.T) {
	gen := NewProtobufGenerator()

	tests := []struct {
		enumName       string
		expectedPrefix string
	}{
		{"UserRole", "USERROLE_UNSPECIFIED"},
		{"Status", "STATUS_UNSPECIFIED"},
		{"OrderType", "ORDERTYPE_UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.enumName, func(t *testing.T) {
			enum := &ast.Enum{
				Name: tt.enumName,
				Values: []*ast.EnumValue{
					{Name: "VALUE1"},
				},
			}

			output := gen.generateEnum(enum)

			if !strings.Contains(output, tt.expectedPrefix) {
				t.Errorf("Expected %q in enum output", tt.expectedPrefix)
			}
		})
	}
}

func TestProtobufGenerator_FieldNumbering(t *testing.T) {
	gen := NewProtobufGenerator()
	typ := &ast.Type{
		Name: "Test",
		Fields: []*ast.Field{
			{Name: "field1", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
			{Name: "field2", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
			{Name: "field3", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
			{Name: "field4", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
			{Name: "field5", Type: &ast.FieldType{Name: "string", IsBuiltin: true}},
		},
	}

	output := gen.generateMessage(typ)

	for i := 1; i <= 5; i++ {
		expected := " = " + string(rune('0'+i)) + ";"
		if !strings.Contains(output, expected) {
			t.Errorf("Expected field number %d in output", i)
		}
	}
}

func TestProtobufGenerator_CustomFieldNumbers(t *testing.T) {
	gen := NewProtobufGenerator()
	typ := &ast.Type{
		Name: "Test",
		Fields: []*ast.Field{
			{
				Name:      "field1",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    1,
				HasNumber: true,
			},
			{
				Name:      "field2",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    5,
				HasNumber: true,
			},
			{
				Name:      "field3",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    10,
				HasNumber: true,
			},
		},
	}

	output := gen.generateMessage(typ)

	if !strings.Contains(output, "field1 = 1;") {
		t.Error("Expected field1 to have number 1")
	}
	if !strings.Contains(output, "field2 = 5;") {
		t.Error("Expected field2 to have number 5")
	}
	if !strings.Contains(output, "field3 = 10;") {
		t.Error("Expected field3 to have number 10")
	}
}

func TestProtobufGenerator_MixedAutoAndCustomFieldNumbers(t *testing.T) {
	gen := NewProtobufGenerator()
	typ := &ast.Type{
		Name: "Test",
		Fields: []*ast.Field{
			{
				Name:      "field1",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    1,
				HasNumber: true,
			},
			{
				Name: "field2",
				Type: &ast.FieldType{Name: "string", IsBuiltin: true},
				// No custom number - should auto-assign 2
			},
			{
				Name:      "field3",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    5,
				HasNumber: true,
			},
			{
				Name: "field4",
				Type: &ast.FieldType{Name: "string", IsBuiltin: true},
				// No custom number - should auto-assign 6 (next after 5)
			},
			{
				Name: "field5",
				Type: &ast.FieldType{Name: "string", IsBuiltin: true},
				// No custom number - should auto-assign 7
			},
		},
	}

	output := gen.generateMessage(typ)

	if !strings.Contains(output, "field1 = 1;") {
		t.Error("Expected field1 to have number 1")
	}
	if !strings.Contains(output, "field2 = 2;") {
		t.Error("Expected field2 to have auto-assigned number 2")
	}
	if !strings.Contains(output, "field3 = 5;") {
		t.Error("Expected field3 to have number 5")
	}
	if !strings.Contains(output, "field4 = 6;") {
		t.Error("Expected field4 to have auto-assigned number 6")
	}
	if !strings.Contains(output, "field5 = 7;") {
		t.Error("Expected field5 to have auto-assigned number 7")
	}
}

func TestProtobufGenerator_CustomFieldNumbers_SparseNumbering(t *testing.T) {
	gen := NewProtobufGenerator()
	typ := &ast.Type{
		Name: "Test",
		Fields: []*ast.Field{
			{
				Name:      "field1",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    100,
				HasNumber: true,
			},
			{
				Name:      "field2",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    200,
				HasNumber: true,
			},
			{
				Name: "field3",
				Type: &ast.FieldType{Name: "string", IsBuiltin: true},
				// Should auto-assign 201 (next after 200)
			},
		},
	}

	output := gen.generateMessage(typ)

	if !strings.Contains(output, "field1 = 100;") {
		t.Error("Expected field1 to have number 100")
	}
	if !strings.Contains(output, "field2 = 200;") {
		t.Error("Expected field2 to have number 200")
	}
	if !strings.Contains(output, "field3 = 201;") {
		t.Error("Expected field3 to have auto-assigned number 201")
	}
}

func TestProtobufGenerator_CustomFieldNumbers_WithExclusion(t *testing.T) {
	gen := NewProtobufGenerator()
	typ := &ast.Type{
		Name: "Test",
		Fields: []*ast.Field{
			{
				Name:      "field1",
				Type:      &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:    1,
				HasNumber: true,
			},
			{
				Name:        "excludedField",
				Type:        &ast.FieldType{Name: "string", IsBuiltin: true},
				Number:      2,
				HasNumber:   true,
				ExcludeFrom: []string{"proto"},
			},
			{
				Name: "field2",
				Type: &ast.FieldType{Name: "string", IsBuiltin: true},
				// Should auto-assign 2 (field 2 was excluded, so its number is available)
			},
		},
	}

	output := gen.generateMessage(typ)

	if !strings.Contains(output, "field1 = 1;") {
		t.Error("Expected field1 to have number 1")
	}
	if strings.Contains(output, "excludedField") {
		t.Error("Expected excludedField to be excluded from output")
	}
	if !strings.Contains(output, "field2 = 2;") {
		t.Error("Expected field2 to have auto-assigned number 2")
	}
}

func TestProtobufGenerator_BoolType(t *testing.T) {
	gen := NewProtobufGenerator()
	field := &ast.Field{
		Name: "isActive",
		Type: &ast.FieldType{
			Name:      "bool",
			IsBuiltin: true,
		},
	}

	result := gen.generateMessageField(field, 1)
	expected := "bool isActive = 1;"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProtobufGenerator_Float32AndFloat64(t *testing.T) {
	gen := NewProtobufGenerator()

	field32 := &ast.Field{
		Name: "value32",
		Type: &ast.FieldType{
			Name:      "float32",
			IsBuiltin: true,
		},
	}

	result32 := gen.generateMessageField(field32, 1)
	if !strings.Contains(result32, "float value32") {
		t.Errorf("Expected float32 to map to 'float', got %q", result32)
	}

	field64 := &ast.Field{
		Name: "value64",
		Type: &ast.FieldType{
			Name:      "float64",
			IsBuiltin: true,
		},
	}

	result64 := gen.generateMessageField(field64, 2)
	if !strings.Contains(result64, "double value64") {
		t.Errorf("Expected float64 to map to 'double', got %q", result64)
	}
}

func TestProtobufGenerator_Documentation(t *testing.T) {
	gen := NewProtobufGenerator()

	tests := []struct {
		name     string
		schema   *ast.Schema
		contains []string
	}{
		{
			name: "enum with general documentation",
			schema: &ast.Schema{
				Enums: []*ast.Enum{
					{
						Name: "UserRole",
						Doc: &ast.Documentation{
							General: "User role enumeration",
						},
						Values: []*ast.EnumValue{
							{Name: "ADMIN"},
						},
					},
				},
			},
			contains: []string{
				"// User role enumeration",
				"enum UserRole",
			},
		},
		{
			name: "enum with proto-specific documentation",
			schema: &ast.Schema{
				Enums: []*ast.Enum{
					{
						Name: "Status",
						Doc: &ast.Documentation{
							General: "General status",
							Specific: map[string]string{
								"proto": "Proto-specific status description",
							},
						},
						Values: []*ast.EnumValue{
							{Name: "ACTIVE"},
						},
					},
				},
			},
			contains: []string{
				"// Proto-specific status description",
				"enum Status",
			},
		},
		{
			name: "enum value with documentation",
			schema: &ast.Schema{
				Enums: []*ast.Enum{
					{
						Name: "UserRole",
						Values: []*ast.EnumValue{
							{
								Name: "ADMIN",
								Doc: &ast.Documentation{
									General: "Administrator with full access",
								},
							},
						},
					},
				},
			},
			contains: []string{
				"// Administrator with full access",
				"ADMIN =",
			},
		},
		{
			name: "type with documentation",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{
						Name: "User",
						Doc: &ast.Documentation{
							General: "User entity",
						},
						Fields: []*ast.Field{
							{
								Name: "id",
								Type: &ast.FieldType{Name: "string", IsBuiltin: true},
							},
						},
					},
				},
			},
			contains: []string{
				"// User entity",
				"message User",
			},
		},
		{
			name: "field with documentation",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{
						Name: "User",
						Fields: []*ast.Field{
							{
								Name: "id",
								Type: &ast.FieldType{Name: "string", IsBuiltin: true},
								Doc: &ast.Documentation{
									General: "Unique identifier",
								},
							},
						},
					},
				},
			},
			contains: []string{
				"// Unique identifier",
				"string id = 1",
			},
		},
		{
			name: "service with documentation",
			schema: &ast.Schema{
				Services: []*ast.Service{
					{
						Name: "UserService",
						Doc: &ast.Documentation{
							General: "User management service",
						},
						Methods: []*ast.Method{
							{
								Name:       "GetUser",
								InputType:  "Req",
								OutputType: "Res",
							},
						},
					},
				},
			},
			contains: []string{
				"// User management service",
				"service UserService",
			},
		},
		{
			name: "method with documentation",
			schema: &ast.Schema{
				Services: []*ast.Service{
					{
						Name: "UserService",
						Methods: []*ast.Method{
							{
								Name:       "CreateUser",
								InputType:  "Req",
								OutputType: "Res",
								Doc: &ast.Documentation{
									General: "Create a new user",
								},
							},
						},
					},
				},
			},
			contains: []string{
				"// Create a new user",
				"rpc CreateUser",
			},
		},
		{
			name: "multiline documentation",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{
						Name: "User",
						Doc: &ast.Documentation{
							General: "User entity\nContains user information\nStored in database",
						},
						Fields: []*ast.Field{
							{
								Name: "id",
								Type: &ast.FieldType{Name: "string", IsBuiltin: true},
							},
						},
					},
				},
			},
			contains: []string{
				"// User entity",
				"// Contains user information",
				"// Stored in database",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := gen.Generate(tt.schema)

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

func TestProtobufGenerator_FieldExclusion(t *testing.T) {
	gen := NewProtobufGenerator()

	tests := []struct {
		name             string
		schema           *ast.Schema
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "exclude field from proto",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{
						Name: "User",
						Fields: []*ast.Field{
							{
								Name: "id",
								Type: &ast.FieldType{Name: "string", IsBuiltin: true},
							},
							{
								Name:        "internalField",
								Type:        &ast.FieldType{Name: "string", IsBuiltin: true},
								ExcludeFrom: []string{"proto"},
							},
						},
					},
				},
			},
			shouldContain: []string{
				"string id = 1",
			},
			shouldNotContain: []string{
				"internalField",
			},
		},
		{
			name: "only include field in proto",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{
						Name: "User",
						Fields: []*ast.Field{
							{
								Name:    "passwordHash",
								Type:    &ast.FieldType{Name: "string", IsBuiltin: true},
								OnlyFor: []string{"proto"},
							},
							{
								Name:    "publicField",
								Type:    &ast.FieldType{Name: "string", IsBuiltin: true},
								OnlyFor: []string{"graphql", "openapi"},
							},
						},
					},
				},
			},
			shouldContain: []string{
				"string passwordHash = 1",
			},
			shouldNotContain: []string{
				"publicField",
			},
		},
		{
			name: "field numbering with exclusions",
			schema: &ast.Schema{
				Types: []*ast.Type{
					{
						Name: "User",
						Fields: []*ast.Field{
							{
								Name: "field1",
								Type: &ast.FieldType{Name: "string", IsBuiltin: true},
							},
							{
								Name:        "field2",
								Type:        &ast.FieldType{Name: "string", IsBuiltin: true},
								ExcludeFrom: []string{"proto"},
							},
							{
								Name: "field3",
								Type: &ast.FieldType{Name: "string", IsBuiltin: true},
							},
						},
					},
				},
			},
			shouldContain: []string{
				"string field1 = 1",
				"string field3 = 2", // Should be 2, not 3, since field2 is excluded
			},
			shouldNotContain: []string{
				"field2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := gen.Generate(tt.schema)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q\nGot:\n%s", expected, output)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output NOT to contain %q\nGot:\n%s", notExpected, output)
				}
			}
		})
	}
}

func TestProtobufGenerator_Namespace(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		expectedInOutput string
	}{
		{
			name:             "simple namespace",
			namespace:        "myapi",
			expectedInOutput: "package myapi;",
		},
		{
			name:             "dotted namespace",
			namespace:        "com.example.api",
			expectedInOutput: "package com.example.api;",
		},
		{
			name:             "default namespace",
			namespace:        "",
			expectedInOutput: "package api;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewProtobufGenerator()
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

func TestProtobufGenerator_GenerateByNamespace(t *testing.T) {
	schema := &ast.Schema{
		Enums: []*ast.Enum{
			{
				Name:      "UserStatus",
				Namespace: "com.example.users",
				Values: []*ast.EnumValue{
					{Name: "ACTIVE", Number: 1, HasNumber: true},
					{Name: "INACTIVE", Number: 2, HasNumber: true},
				},
			},
			{
				Name:      "OrderStatus",
				Namespace: "com.example.orders",
				Values: []*ast.EnumValue{
					{Name: "PENDING", Number: 1, HasNumber: true},
					{Name: "CONFIRMED", Number: 2, HasNumber: true},
				},
			},
		},
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "com.example.users",
				Fields: []*ast.Field{
					{
						Name:      "id",
						Number:    1,
						HasNumber: true,
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
				},
			},
			{
				Name:      "Order",
				Namespace: "com.example.orders",
				Fields: []*ast.Field{
					{
						Name:      "id",
						Number:    1,
						HasNumber: true,
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
					{
						Name:      "customer",
						Number:    2,
						HasNumber: true,
						Type: &ast.FieldType{
							Name:      "com.example.users.User",
							IsBuiltin: false,
						},
					},
				},
			},
		},
	}

	gen := NewProtobufGenerator()
	protoFiles := gen.GenerateByNamespace(schema)

	// Should generate two separate proto files
	if len(protoFiles) != 2 {
		t.Fatalf("Expected 2 proto files, got %d", len(protoFiles))
	}

	// Check users namespace file
	usersProto, ok := protoFiles["com.example.users"]
	if !ok {
		t.Fatal("Expected com.example.users namespace proto file")
	}

	if !strings.Contains(usersProto, "package com.example.users;") {
		t.Error("Users proto should contain correct package declaration")
	}

	if !strings.Contains(usersProto, "enum UserStatus") {
		t.Error("Users proto should contain UserStatus enum")
	}

	if !strings.Contains(usersProto, "message User") {
		t.Error("Users proto should contain User message")
	}

	// Check orders namespace file
	ordersProto, ok := protoFiles["com.example.orders"]
	if !ok {
		t.Fatal("Expected com.example.orders namespace proto file")
	}

	if !strings.Contains(ordersProto, "package com.example.orders;") {
		t.Error("Orders proto should contain correct package declaration")
	}

	if !strings.Contains(ordersProto, "enum OrderStatus") {
		t.Error("Orders proto should contain OrderStatus enum")
	}

	if !strings.Contains(ordersProto, "message Order") {
		t.Error("Orders proto should contain Order message")
	}

	// Check for import of users namespace in orders proto
	if !strings.Contains(ordersProto, "import \"com/example/users.proto\";") {
		t.Error("Orders proto should import users proto file")
	}

	// Check for fully qualified type reference
	if !strings.Contains(ordersProto, "com.example.users.User customer") {
		t.Error("Orders proto should use fully qualified type name for cross-namespace reference")
	}
}

func TestProtobufGenerator_MapScalarTypeWithPackage(t *testing.T) {
	gen := NewProtobufGenerator()

	tests := []struct {
		name             string
		typeName         string
		currentNamespace string
		expected         string
	}{
		{
			name:             "builtin type",
			typeName:         "string",
			currentNamespace: "com.example.api",
			expected:         "string",
		},
		{
			name:             "same namespace type",
			typeName:         "com.example.api.User",
			currentNamespace: "com.example.api",
			expected:         "User",
		},
		{
			name:             "different namespace type",
			typeName:         "com.example.users.User",
			currentNamespace: "com.example.orders",
			expected:         "com.example.users.User",
		},
		{
			name:             "unqualified local type",
			typeName:         "User",
			currentNamespace: "com.example.api",
			expected:         "User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.mapScalarTypeWithPackage(tt.typeName, tt.currentNamespace)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProtobufGenerator_FindRequiredNamespaces(t *testing.T) {
	gen := NewProtobufGenerator()

	nsSchema := &ast.Schema{
		Namespace: "com.example.orders",
		Types: []*ast.Type{
			{
				Name:      "Order",
				Namespace: "com.example.orders",
				Fields: []*ast.Field{
					{
						Name: "customer",
						Type: &ast.FieldType{
							Name: "com.example.users.User",
						},
					},
					{
						Name: "status",
						Type: &ast.FieldType{
							Name: "OrderStatus",
						},
					},
				},
			},
		},
		Services: []*ast.Service{
			{
				Name:      "OrderService",
				Namespace: "com.example.orders",
				Methods: []*ast.Method{
					{
						Name:       "GetUser",
						InputType:  "com.example.users.GetUserRequest",
						OutputType: "GetUserResponse",
					},
				},
			},
		},
	}

	required := gen.findRequiredNamespaces(nsSchema)

	// Should find com.example.users namespace
	found := false
	for _, ns := range required {
		if ns == "com.example.users" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find com.example.users in required namespaces, got %v", required)
	}
}

func TestProtobufGenerator_NameAnnotation(t *testing.T) {
	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "api",
				Annotations: &ast.FormatAnnotations{
					ProtoName: "UserV2",
				},
				Fields: []*ast.Field{
					{
						Name:      "id",
						Number:    1,
						HasNumber: true,
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
				},
			},
		},
	}

	gen := NewProtobufGenerator()
	output := gen.Generate(schema)

	if !strings.Contains(output, "message UserV2 {") {
		t.Error("Expected output to contain 'message UserV2 {', but it didn't")
	}

	if strings.Contains(output, "message User {") {
		t.Error("Expected output NOT to contain 'message User {', but it did")
	}
}

func TestGenerateOptionalFields(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "id",
						Type: &ast.FieldType{
							Name:     "string",
							Optional: false,
						},
						Number:    1,
						HasNumber: true,
					},
					{
						Name: "name",
						Type: &ast.FieldType{
							Name:     "string",
							Optional: true,
						},
						Number:    2,
						HasNumber: true,
					},
					{
						Name: "tags",
						Type: &ast.FieldType{
							Name:     "string",
							IsArray:  true,
							Optional: true,
						},
						Number:    3,
						HasNumber: true,
					},
				},
			},
		},
	}

	gen := NewProtobufGenerator()
	output := gen.Generate(schema)

	// Check for optional keyword
	if !strings.Contains(output, "optional string name = 2;") {
		t.Error("Expected 'optional string name = 2;' in output")
	}

	// Non-optional field should not have optional keyword
	if strings.Contains(output, "optional string id") {
		t.Error("Non-optional field 'id' should not have 'optional' keyword")
	}

	// Arrays should use repeated, not optional
	if !strings.Contains(output, "repeated string tags = 3;") {
		t.Error("Expected 'repeated string tags = 3;' in output")
	}
	if strings.Contains(output, "optional repeated") {
		t.Error("Should not have 'optional repeated' for array fields")
	}
}

func TestProtobufGenerator_NestedMaps(t *testing.T) {
	gen := NewProtobufGenerator()

	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "NestedMapTest",
				Fields: []*ast.Field{
					{
						Name: "simple_map",
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
						Name: "nested_map",
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
						Name: "triple_nested_map",
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

	// Check simple map
	if !strings.Contains(output, "map<string, string> simple_map = 1;") {
		t.Error("Expected 'map<string, string> simple_map = 1;' in output")
	}

	// Check nested map - should have map<string, map<string, int32>>
	if !strings.Contains(output, "map<string, map<string, int32>> nested_map = 2;") {
		t.Error("Expected 'map<string, map<string, int32>> nested_map = 2;' in output")
	}

	// Check triple nested map
	if !strings.Contains(output, "map<string, map<string, map<string, bool>>> triple_nested_map = 3;") {
		t.Error("Expected 'map<string, map<string, map<string, bool>>> triple_nested_map = 3;' in output")
	}
}
