package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestGoGenerator_Generate(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name: "id",
						Type: &ast.FieldType{Name: "string"},
					},
					{
						Name: "email",
						Type: &ast.FieldType{Name: "string"},
					},
					{
						Name: "age",
						Type: &ast.FieldType{Name: "int32"},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check package declaration
	if !strings.Contains(output, "package api") {
		t.Errorf("Expected package declaration, got: %s", output)
	}

	// Check struct definition
	if !strings.Contains(output, "type User struct {") {
		t.Errorf("Expected User struct definition")
	}

	// Check fields
	if !strings.Contains(output, "Id string") {
		t.Errorf("Expected Id field")
	}
	if !strings.Contains(output, "Email string") {
		t.Errorf("Expected Email field")
	}
	if !strings.Contains(output, "Age int32") {
		t.Errorf("Expected Age field")
	}

	// Check JSON tags
	if !strings.Contains(output, "`json:\"id\"`") {
		t.Errorf("Expected JSON tag for id field")
	}
}

func TestGoGenerator_GenerateEnum(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Enums: []*ast.Enum{
			{
				Name:      "Status",
				Namespace: "api",
				Values: []*ast.EnumValue{
					{Name: "Active"},
					{Name: "Inactive"},
					{Name: "Pending"},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check enum type
	if !strings.Contains(output, "type Status int") {
		t.Errorf("Expected Status enum type definition")
	}

	// Check enum values
	if !strings.Contains(output, "StatusActive Status = iota") {
		t.Errorf("Expected StatusActive enum value")
	}
	if !strings.Contains(output, "StatusInactive") {
		t.Errorf("Expected StatusInactive enum value")
	}
	if !strings.Contains(output, "StatusPending") {
		t.Errorf("Expected StatusPending enum value")
	}
}

func TestGoGenerator_GenerateWithTimestamp(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "Event",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name: "timestamp",
						Type: &ast.FieldType{Name: "timestamp"},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check time import
	if !strings.Contains(output, "import (\n\t\"time\"\n)") {
		t.Errorf("Expected time import")
	}

	// Check time.Time field
	if !strings.Contains(output, "Timestamp time.Time") {
		t.Errorf("Expected Timestamp field of type time.Time")
	}
}

func TestGoGenerator_GenerateArrayField(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "Post",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name: "tags",
						Type: &ast.FieldType{
							Name:    "string",
							IsArray: true,
						},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check array field
	if !strings.Contains(output, "Tags []string") {
		t.Errorf("Expected Tags field of type []string, got: %s", output)
	}
}

func TestGoGenerator_GenerateMapField(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "Config",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name: "settings",
						Type: &ast.FieldType{
							MapKey:   "string",
							MapValue: "string",
						},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check map field
	if !strings.Contains(output, "Settings map[string]string") {
		t.Errorf("Expected Settings field of type map[string]string")
	}
}

func TestGoGenerator_GenerateOptionalField(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "Profile",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name: "bio",
						Type: &ast.FieldType{
							Name:     "string",
							Optional: true,
						},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check optional field with omitempty tag
	if !strings.Contains(output, "`json:\"bio,omitempty\"`") {
		t.Errorf("Expected omitempty JSON tag for optional field")
	}
}

func TestGoGenerator_GenerateUnion(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Unions: []*ast.Union{
			{
				Name:      "PaymentMethod",
				Namespace: "api",
				Options:   []string{"CreditCard", "PayPal", "BankTransfer"},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check union interface
	if !strings.Contains(output, "type PaymentMethod interface") {
		t.Errorf("Expected PaymentMethod interface")
	}
	if !strings.Contains(output, "isPaymentMethod()") {
		t.Errorf("Expected isPaymentMethod marker method")
	}

	// Check concrete types
	if !strings.Contains(output, "type PaymentMethodCreditCard struct") {
		t.Errorf("Expected PaymentMethodCreditCard type")
	}
	if !strings.Contains(output, "type PaymentMethodPayPal struct") {
		t.Errorf("Expected PaymentMethodPayPal type")
	}
}

func TestGoGenerator_GenerateService(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Services: []*ast.Service{
			{
				Name:      "UserService",
				Namespace: "api",
				Methods: []*ast.Method{
					{
						Name:       "GetUser",
						InputType:  "GetUserRequest",
						OutputType: "GetUserResponse",
					},
					{
						Name:         "ListUsers",
						InputType:    "ListUsersRequest",
						OutputType:   "ListUsersResponse",
						OutputStream: true,
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check service interface
	if !strings.Contains(output, "type UserService interface") {
		t.Errorf("Expected UserService interface")
	}

	// Check regular method
	if !strings.Contains(output, "GetUser(input *GetUserRequest) (*GetUserResponse, error)") {
		t.Errorf("Expected GetUser method signature")
	}

	// Check streaming method
	if !strings.Contains(output, "ListUsers(input *ListUsersRequest, stream chan *ListUsersResponse) error") {
		t.Errorf("Expected ListUsers streaming method signature")
	}
}

func TestGoGenerator_GenerateWithDocumentation(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "api",
				Doc: &ast.Documentation{
					General: "User represents a system user",
				},
				Fields: []*ast.Field{
					{
						Name: "id",
						Type: &ast.FieldType{Name: "string"},
						Doc: &ast.Documentation{
							General: "Unique identifier",
						},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check type documentation
	if !strings.Contains(output, "// User represents a system user") {
		t.Errorf("Expected type documentation")
	}

	// Check field documentation
	if !strings.Contains(output, "// Unique identifier") {
		t.Errorf("Expected field documentation")
	}
}

func TestGoGenerator_GetPackageName(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		expected  string
	}{
		{
			name:      "simple namespace",
			namespace: "users",
			expected:  "users",
		},
		{
			name:      "dotted namespace",
			namespace: "com.example.users",
			expected:  "users",
		},
		{
			name:      "empty namespace",
			namespace: "",
			expected:  "api",
		},
		{
			name:      "namespace with hyphens",
			namespace: "user-service",
			expected:  "userservice",
		},
	}

	gen := NewGoGenerator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.getPackageName(tt.namespace)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGoGenerator_ExportFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "name",
			expected: "Name",
		},
		{
			name:     "snake_case",
			input:    "user_id",
			expected: "UserId",
		},
		{
			name:     "multiple underscores",
			input:    "created_at_timestamp",
			expected: "CreatedAtTimestamp",
		},
	}

	gen := NewGoGenerator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.exportFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGoGenerator_PackageAnnotation(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.api",
		NamespaceAnnotations: &ast.FormatAnnotations{
			Go: []string{`package = "custompackage"`},
		},
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "id",
						Type: &ast.FieldType{Name: "string", IsBuiltin: true},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	if !strings.Contains(output, "package custompackage") {
		t.Errorf("Expected package custompackage, got:\n%s", output)
	}
}

func TestGoGenerator_PackageAnnotationDefault(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "com.example.api",
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "id",
						Type: &ast.FieldType{Name: "string", IsBuiltin: true},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Should use default package name derived from namespace
	if !strings.Contains(output, "package api") {
		t.Errorf("Expected package api, got:\n%s", output)
	}
}

func TestGoGenerator_JSONNameAnnotation(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name:     "userId",
						JSONName: "user_id",
						Type:     &ast.FieldType{Name: "string"},
					},
					{
						Name:     "createdAt",
						JSONName: "created_at",
						Type:     &ast.FieldType{Name: "timestamp"},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check that JSON tag uses JSONName
	if !strings.Contains(output, "`json:\"user_id\"`") {
		t.Errorf("Expected json:\"user_id\" tag, got: %s", output)
	}
	if !strings.Contains(output, "`json:\"created_at\"`") {
		t.Errorf("Expected json:\"created_at\" tag, got: %s", output)
	}
}

func TestGoGenerator_JSONNullableAnnotation(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name:         "middleName",
						JSONNullable: true,
						Type:         &ast.FieldType{Name: "string"},
					},
					{
						Name:         "age",
						JSONNullable: true,
						Type:         &ast.FieldType{Name: "int32"},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check that nullable fields are pointer types
	if !strings.Contains(output, "MiddleName *string") {
		t.Errorf("Expected MiddleName *string for nullable field, got: %s", output)
	}
	if !strings.Contains(output, "Age *int32") {
		t.Errorf("Expected Age *int32 for nullable field, got: %s", output)
	}
}

func TestGoGenerator_JSONOmitEmptyAnnotation(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "User",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name:          "description",
						JSONOmitEmpty: true,
						Type:          &ast.FieldType{Name: "string"},
					},
					{
						Name:          "metadata",
						JSONOmitEmpty: true,
						Type: &ast.FieldType{
							MapKey:   "string",
							MapValue: "string",
						},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check that omitempty is added to JSON tags
	if !strings.Contains(output, "`json:\"description,omitempty\"`") {
		t.Errorf("Expected json:\"description,omitempty\" tag, got: %s", output)
	}
	if !strings.Contains(output, "`json:\"metadata,omitempty\"`") {
		t.Errorf("Expected json:\"metadata,omitempty\" tag, got: %s", output)
	}
}

func TestGoGenerator_CombinedJSONAnnotations(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "api",
		Types: []*ast.Type{
			{
				Name:      "Profile",
				Namespace: "api",
				Fields: []*ast.Field{
					{
						Name:          "phoneNumber",
						JSONName:      "phone_number",
						JSONNullable:  true,
						JSONOmitEmpty: true,
						Type:          &ast.FieldType{Name: "string"},
					},
				},
			},
		},
	}

	gen := NewGoGenerator()
	output := gen.Generate(schema)

	// Check that all annotations work together
	if !strings.Contains(output, "PhoneNumber *string") {
		t.Errorf("Expected PhoneNumber *string, got: %s", output)
	}
	if !strings.Contains(output, "`json:\"phone_number,omitempty\"`") {
		t.Errorf("Expected json:\"phone_number,omitempty\" tag, got: %s", output)
	}
}
