package docgen

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestGenerateBasicMarkdown(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "User",
				Doc: &ast.Documentation{
					General: "User entity",
				},
				Fields: []*ast.Field{
					{
						Name:     "id",
						Required: true,
						Type: &ast.FieldType{
							Name: "string",
						},
						Doc: &ast.Documentation{
							General: "User ID",
						},
					},
					{
						Name: "name",
						Type: &ast.FieldType{
							Name:     "string",
							Optional: true,
						},
					},
				},
			},
		},
	}

	gen := NewMarkdownGenerator()
	output := gen.Generate(schema)

	// Check title
	if !strings.Contains(output, "# test API Documentation") {
		t.Error("Expected title with namespace")
	}

	// Check table of contents
	if !strings.Contains(output, "## Table of Contents") {
		t.Error("Expected table of contents")
	}

	// Check type name
	if !strings.Contains(output, "### User") {
		t.Error("Expected User type heading")
	}

	// Check documentation
	if !strings.Contains(output, "User entity") {
		t.Error("Expected type documentation")
	}

	// Check fields table
	if !strings.Contains(output, "| Field | Type | Required | Description |") {
		t.Error("Expected fields table header")
	}

	// Check required field
	if !strings.Contains(output, "| `id` | `string` | Yes | User ID |") {
		t.Error("Expected id field with 'Yes' for required")
	}

	// Check optional field
	if !strings.Contains(output, "| `name` | `string?` | No |") {
		t.Error("Expected name field with 'No' for optional and ? in type")
	}
}

func TestGenerateEnumMarkdown(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Enums: []*ast.Enum{
			{
				Name: "Status",
				Doc: &ast.Documentation{
					General: "Status enumeration",
				},
				Values: []*ast.EnumValue{
					{
						Name:   "ACTIVE",
						Number: 0,
						Doc: &ast.Documentation{
							General: "Active status",
						},
					},
					{
						Name:   "INACTIVE",
						Number: 1,
					},
				},
			},
		},
	}

	gen := NewMarkdownGenerator()
	output := gen.Generate(schema)

	// Check enum name
	if !strings.Contains(output, "### Status") {
		t.Error("Expected Status enum heading")
	}

	// Check values table
	if !strings.Contains(output, "| Value | Number | Description |") {
		t.Error("Expected enum values table header")
	}

	// Check enum value with documentation
	if !strings.Contains(output, "| `ACTIVE` | 0 | Active status |") {
		t.Error("Expected ACTIVE enum value with description")
	}

	// Check enum value without documentation
	if !strings.Contains(output, "| `INACTIVE` | 1 |") {
		t.Error("Expected INACTIVE enum value")
	}
}

func TestGenerateServiceMarkdown(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Services: []*ast.Service{
			{
				Name: "UserService",
				Doc: &ast.Documentation{
					General: "User management service",
				},
				Methods: []*ast.Method{
					{
						Name:         "GetUser",
						InputType:    "GetUserRequest",
						OutputType:   "User",
						HTTPMethod:   "GET",
						PathTemplate: "/users/{id}",
						Doc: &ast.Documentation{
							General: "Retrieves a user by ID",
						},
					},
					{
						Name:         "StreamUsers",
						InputType:    "Empty",
						OutputType:   "User",
						OutputStream: true,
					},
				},
			},
		},
	}

	gen := NewMarkdownGenerator()
	output := gen.Generate(schema)

	// Check service name
	if !strings.Contains(output, "### UserService") {
		t.Error("Expected UserService heading")
	}

	// Check method name
	if !strings.Contains(output, "##### GetUser") {
		t.Error("Expected GetUser method heading")
	}

	// Check method documentation
	if !strings.Contains(output, "Retrieves a user by ID") {
		t.Error("Expected method documentation")
	}

	// Check request/response
	if !strings.Contains(output, "**Request:** `GetUserRequest`") {
		t.Error("Expected request type")
	}
	if !strings.Contains(output, "**Response:** `User`") {
		t.Error("Expected response type")
	}

	// Check HTTP mapping
	if !strings.Contains(output, "**HTTP:** `GET /users/{id}`") {
		t.Error("Expected HTTP mapping")
	}

	// Check streaming indicator
	if !strings.Contains(output, "StreamUsers (server streaming)") {
		t.Error("Expected server streaming indicator")
	}
}

func TestGenerateMapTypeMarkdown(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "Config",
				Fields: []*ast.Field{
					{
						Name: "metadata",
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

	gen := NewMarkdownGenerator()
	output := gen.Generate(schema)

	// Check map type formatting
	if !strings.Contains(output, "`map<string, string>`") {
		t.Error("Expected map type to be formatted as 'map<string, string>'")
	}
}

func TestGenerateArrayTypeMarkdown(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "tags",
						Type: &ast.FieldType{
							IsArray: true,
							Name:    "string",
						},
					},
				},
			},
		},
	}

	gen := NewMarkdownGenerator()
	output := gen.Generate(schema)

	// Check array type formatting
	if !strings.Contains(output, "`[]string`") {
		t.Error("Expected array type to be formatted as '[]string'")
	}
}

func TestGenerateDeprecatedFieldMarkdown(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name: "oldField",
						Type: &ast.FieldType{
							Name: "string",
						},
						Deprecated: &ast.DeprecationInfo{
							Reason: "Use newField instead",
							Since:  "2.0.0",
						},
					},
				},
			},
		},
	}

	gen := NewMarkdownGenerator()
	output := gen.Generate(schema)

	// Check deprecation warning
	if !strings.Contains(output, "⚠️ **DEPRECATED**") {
		t.Error("Expected deprecation warning")
	}
	if !strings.Contains(output, "Use newField instead") {
		t.Error("Expected deprecation reason")
	}
}
