package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestProtobufGenerator_GenerateUnion(t *testing.T) {
	gen := NewProtobufGenerator()

	union := &ast.Union{
		Name: "Payment",
		Options: []string{
			"CreditCard",
			"BankTransfer",
			"PayPal",
		},
	}

	output := gen.generateUnion(union)

	// Should create a message with oneof
	if !strings.Contains(output, "message Payment") {
		t.Error("Expected 'message Payment' in output")
	}

	if !strings.Contains(output, "oneof value") {
		t.Error("Expected 'oneof value' in output")
	}

	// Check all options are included
	if !strings.Contains(output, "CreditCard creditCard = 1") {
		t.Error("Expected CreditCard option")
	}

	if !strings.Contains(output, "BankTransfer bankTransfer = 2") {
		t.Error("Expected BankTransfer option")
	}

	if !strings.Contains(output, "PayPal payPal = 3") {
		t.Error("Expected PayPal option")
	}
}

func TestProtobufGenerator_GenerateUnion_WithDocumentation(t *testing.T) {
	gen := NewProtobufGenerator()

	union := &ast.Union{
		Name: "Status",
		Options: []string{
			"Active",
			"Inactive",
		},
		Doc: &ast.Documentation{
			General: "Payment status options",
		},
	}

	output := gen.generateUnion(union)

	// Should include documentation
	if !strings.Contains(output, "// Payment status options") {
		t.Error("Expected documentation comment in output")
	}
}

func TestGraphQLGenerator_GenerateUnion(t *testing.T) {
	gen := NewGraphQLGenerator()

	union := &ast.Union{
		Name: "SearchResult",
		Options: []string{
			"User",
			"Post",
			"Comment",
		},
	}

	output := gen.generateUnion(union)

	// Should create a union type
	if !strings.Contains(output, "union SearchResult") {
		t.Error("Expected 'union SearchResult' in output")
	}

	// Check all options with pipe separator
	if !strings.Contains(output, "User | Post | Comment") {
		t.Error("Expected union options with pipe separator")
	}
}

func TestGraphQLGenerator_GenerateUnion_WithDocumentation(t *testing.T) {
	gen := NewGraphQLGenerator()

	union := &ast.Union{
		Name: "Result",
		Options: []string{
			"Success",
			"Error",
		},
		Doc: &ast.Documentation{
			General: "Operation result",
		},
	}

	output := gen.generateUnion(union)

	// Should include documentation as string
	if !strings.Contains(output, "Operation result") {
		t.Error("Expected documentation in output")
	}
}

func TestGraphQLGenerator_GenerateUnionInput(t *testing.T) {
	gen := NewGraphQLGenerator()

	union := &ast.Union{
		Name: "MediaInput",
		Options: []string{
			"Image",
			"Video",
			"Audio",
		},
	}

	output := gen.generateUnionInput(union)

	// Should create input type with @oneOf directive
	if !strings.Contains(output, "input MediaInputInput @oneOf") {
		t.Error("Expected 'input MediaInputInput @oneOf' in output")
	}

	// Check all options as camelCase fields
	if !strings.Contains(output, "image: ImageInput") {
		t.Error("Expected 'image: ImageInput' field")
	}

	if !strings.Contains(output, "video: VideoInput") {
		t.Error("Expected 'video: VideoInput' field")
	}

	if !strings.Contains(output, "audio: AudioInput") {
		t.Error("Expected 'audio: AudioInput' field")
	}
}

func TestGraphQLGenerator_Generate_WithUnion(t *testing.T) {
	gen := NewGraphQLGenerator()

	schema := &ast.Schema{
		Unions: []*ast.Union{
			{
				Name: "SearchResult",
				Options: []string{
					"User",
					"Post",
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
			{
				Name: "Post",
				Fields: []*ast.Field{
					{
						Name: "title",
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

	output := gen.Generate(schema)

	// Should generate both union and input version
	if !strings.Contains(output, "union SearchResult = User | Post") {
		t.Error("Expected union definition in output")
	}

	if !strings.Contains(output, "input SearchResultInput @oneOf") {
		t.Error("Expected union input definition in output")
	}
}

func TestOpenAPIGenerator_GenerateUnionSchema(t *testing.T) {
	gen := NewOpenAPIGenerator()

	union := &ast.Union{
		Name: "Pet",
		Options: []string{
			"Cat",
			"Dog",
			"Bird",
		},
	}

	schema := gen.generateUnionSchema(union)

	// Should have oneOf with references
	if len(schema.OneOf) != 3 {
		t.Errorf("Expected 3 oneOf options, got %d", len(schema.OneOf))
	}

	// Check references
	expectedRefs := []string{
		"#/components/schemas/Cat",
		"#/components/schemas/Dog",
		"#/components/schemas/Bird",
	}

	for i, expectedRef := range expectedRefs {
		if schema.OneOf[i].Ref != expectedRef {
			t.Errorf("Expected ref %s, got %s", expectedRef, schema.OneOf[i].Ref)
		}
	}
}

func TestOpenAPIGenerator_GenerateUnionSchema_WithDocumentation(t *testing.T) {
	gen := NewOpenAPIGenerator()

	union := &ast.Union{
		Name: "Animal",
		Options: []string{
			"Cat",
			"Dog",
		},
		Doc: &ast.Documentation{
			General: "Supported animal types",
		},
	}

	schema := gen.generateUnionSchema(union)

	// Should include documentation
	if schema.Description != "Supported animal types" {
		t.Errorf("Expected documentation, got: %s", schema.Description)
	}
}

func TestOpenAPIGenerator_Generate_WithUnion(t *testing.T) {
	gen := NewOpenAPIGenerator()

	schema := &ast.Schema{
		Unions: []*ast.Union{
			{
				Name: "Result",
				Options: []string{
					"Success",
					"Error",
				},
			},
		},
		Types: []*ast.Type{
			{
				Name: "Success",
				Fields: []*ast.Field{
					{
						Name: "data",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
				},
			},
			{
				Name: "Error",
				Fields: []*ast.Field{
					{
						Name: "message",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
				},
			},
		},
	}

	output := gen.Generate(schema)

	// Check for Result schema with oneOf
	if !strings.Contains(output, "Result:") {
		t.Error("Expected Result schema in output")
	}

	if !strings.Contains(output, "oneOf:") {
		t.Error("Expected oneOf in output")
	}

	if !strings.Contains(output, "#/components/schemas/Success") {
		t.Error("Expected Success reference in oneOf")
	}

	if !strings.Contains(output, "#/components/schemas/Error") {
		t.Error("Expected Error reference in oneOf")
	}
}

func TestGraphQLGenerator_UnionFieldInType(t *testing.T) {
	gen := NewGraphQLGenerator()

	schema := &ast.Schema{
		Unions: []*ast.Union{
			{
				Name:    "Result",
				Options: []string{"Success", "Error"},
			},
		},
		Types: []*ast.Type{
			{
				Name: "Response",
				Fields: []*ast.Field{
					{
						Name: "result",
						Type: &ast.FieldType{
							Name:      "Result",
							IsBuiltin: false,
						},
						Required: true,
					},
				},
			},
		},
	}

	output := gen.Generate(schema)

	// Union field in output type should use the union name directly
	if !strings.Contains(output, "result: Result!") {
		t.Error("Expected 'result: Result!' in output type")
	}
}

func TestGraphQLGenerator_UnionAsInputType(t *testing.T) {
	gen := NewGraphQLGenerator()

	// Test that union types that are part of input structures get Input versions
	schema := &ast.Schema{
		Unions: []*ast.Union{
			{
				Name:    "Filter",
				Options: []string{"TextFilter", "DateFilter"},
			},
		},
		Types: []*ast.Type{
			{
				Name: "TextFilter",
				Fields: []*ast.Field{
					{
						Name: "text",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
				},
			},
			{
				Name: "DateFilter",
				Fields: []*ast.Field{
					{
						Name: "date",
						Type: &ast.FieldType{
							Name:      "string",
							IsBuiltin: true,
						},
					},
				},
			},
		},
	}

	output := gen.Generate(schema)

	// Should have the @oneOf input type for the union
	if !strings.Contains(output, "input FilterInput @oneOf") {
		t.Error("Expected FilterInput with @oneOf")
	}

	// Should have TextFilterInput and DateFilterInput since they're union options
	if !strings.Contains(output, "input TextFilterInput") {
		t.Error("Expected TextFilterInput")
	}

	if !strings.Contains(output, "input DateFilterInput") {
		t.Error("Expected DateFilterInput")
	}
}
