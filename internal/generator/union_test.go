package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestGenerateUnionProtobuf(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "TextMessage",
				Fields: []*ast.Field{
					{
						Name:      "content",
						Type:      &ast.FieldType{Name: "string"},
						Number:    1,
						HasNumber: true,
					},
				},
			},
			{
				Name: "ImageMessage",
				Fields: []*ast.Field{
					{
						Name:      "imageUrl",
						Type:      &ast.FieldType{Name: "string"},
						Number:    1,
						HasNumber: true,
					},
				},
			},
		},
		Unions: []*ast.Union{
			{
				Name:    "Message",
				Options: []string{"TextMessage", "ImageMessage"},
				Doc: &ast.Documentation{
					General: "A message union",
				},
			},
		},
	}

	gen := NewProtobufGenerator()
	output := gen.Generate(schema)

	// Check for oneof structure
	if !strings.Contains(output, "oneof value {") {
		t.Error("Expected 'oneof value {' in Protobuf output")
	}

	// Check for union options
	if !strings.Contains(output, "TextMessage textMessage = 1;") {
		t.Error("Expected 'TextMessage textMessage = 1;' in oneof")
	}
	if !strings.Contains(output, "ImageMessage imageMessage = 2;") {
		t.Error("Expected 'ImageMessage imageMessage = 2;' in oneof")
	}

	// Check for union message
	if !strings.Contains(output, "message Message {") {
		t.Error("Expected 'message Message {' in Protobuf output")
	}

	// Check documentation
	if !strings.Contains(output, "// A message union") {
		t.Error("Expected union documentation in output")
	}
}

func TestGenerateUnionGraphQL(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "TextMessage",
				Fields: []*ast.Field{
					{
						Name:     "content",
						Type:     &ast.FieldType{Name: "string"},
						Required: true,
					},
				},
			},
			{
				Name: "ImageMessage",
				Fields: []*ast.Field{
					{
						Name:     "imageUrl",
						Type:     &ast.FieldType{Name: "string"},
						Required: true,
					},
				},
			},
		},
		Unions: []*ast.Union{
			{
				Name:    "Message",
				Options: []string{"TextMessage", "ImageMessage"},
				Doc: &ast.Documentation{
					General: "A message union",
				},
			},
		},
	}

	gen := NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Check for union declaration
	if !strings.Contains(output, "union Message = TextMessage | ImageMessage") {
		t.Error("Expected 'union Message = TextMessage | ImageMessage' in GraphQL output")
	}

	// Check for input variant with @oneOf directive
	if !strings.Contains(output, "input MessageInput @oneOf {") {
		t.Error("Expected 'input MessageInput @oneOf {' in GraphQL output")
	}

	// Check for camelCase field names in input
	if !strings.Contains(output, "textMessage: TextMessageInput") {
		t.Error("Expected 'textMessage: TextMessageInput' in MessageInput")
	}
	if !strings.Contains(output, "imageMessage: ImageMessageInput") {
		t.Error("Expected 'imageMessage: ImageMessageInput' in MessageInput")
	}

	// Check documentation
	if !strings.Contains(output, "A message union") {
		t.Error("Expected union documentation in output")
	}
}

func TestGenerateUnionOpenAPI(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "TextMessage",
				Fields: []*ast.Field{
					{
						Name:     "content",
						Type:     &ast.FieldType{Name: "string"},
						Required: true,
					},
				},
			},
			{
				Name: "ImageMessage",
				Fields: []*ast.Field{
					{
						Name:     "imageUrl",
						Type:     &ast.FieldType{Name: "string"},
						Required: true,
					},
				},
			},
		},
		Unions: []*ast.Union{
			{
				Name:    "Message",
				Options: []string{"TextMessage", "ImageMessage"},
				Doc: &ast.Documentation{
					General: "A message union",
				},
			},
		},
	}

	gen := NewOpenAPIGenerator()
	output := gen.Generate(schema)

	// Check for oneOf structure
	if !strings.Contains(output, "oneOf:") {
		t.Error("Expected 'oneOf:' in OpenAPI output")
	}

	// Check for discriminator
	if !strings.Contains(output, "discriminator:") {
		t.Error("Expected 'discriminator:' in OpenAPI output")
	}
	if !strings.Contains(output, "propertyName: type") {
		t.Error("Expected 'propertyName: type' in discriminator")
	}

	// Check for mapping
	if !strings.Contains(output, "mapping:") {
		t.Error("Expected 'mapping:' in discriminator")
	}
	if !strings.Contains(output, "TextMessage: '#/components/schemas/TextMessage'") {
		t.Error("Expected TextMessage mapping in discriminator")
	}
	if !strings.Contains(output, "ImageMessage: '#/components/schemas/ImageMessage'") {
		t.Error("Expected ImageMessage mapping in discriminator")
	}

	// Check for schema references
	if !strings.Contains(output, "$ref: '#/components/schemas/TextMessage'") {
		t.Error("Expected reference to TextMessage schema")
	}
	if !strings.Contains(output, "$ref: '#/components/schemas/ImageMessage'") {
		t.Error("Expected reference to ImageMessage schema")
	}

	// Check documentation
	if !strings.Contains(output, "A message union") {
		t.Error("Expected union documentation in output")
	}
}

func TestUnionWithManyOptions(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Unions: []*ast.Union{
			{
				Name:    "MultiUnion",
				Options: []string{"Type1", "Type2", "Type3", "Type4", "Type5"},
			},
		},
	}

	// Test Protobuf
	genProto := NewProtobufGenerator()
	outputProto := genProto.Generate(schema)
	for i := 1; i <= 5; i++ {
		expected := "Type" + string(rune('0'+i))
		if !strings.Contains(outputProto, expected) {
			t.Errorf("Expected '%s' in Protobuf union", expected)
		}
	}

	// Test GraphQL
	genGraphQL := NewGraphQLGenerator()
	outputGraphQL := genGraphQL.Generate(schema)
	if !strings.Contains(outputGraphQL, "union MultiUnion = Type1 | Type2 | Type3 | Type4 | Type5") {
		t.Error("Expected all types in GraphQL union separated by |")
	}

	// Test OpenAPI
	genOpenAPI := NewOpenAPIGenerator()
	outputOpenAPI := genOpenAPI.Generate(schema)
	for i := 1; i <= 5; i++ {
		expected := "Type" + string(rune('0'+i))
		if !strings.Contains(outputOpenAPI, expected) {
			t.Errorf("Expected '%s' in OpenAPI union", expected)
		}
	}
}

func TestUnionInServiceMethod(t *testing.T) {
	schema := &ast.Schema{
		Namespace: "test",
		Types: []*ast.Type{
			{
				Name: "TextMessage",
				Fields: []*ast.Field{
					{Name: "content", Type: &ast.FieldType{Name: "string"}},
				},
			},
		},
		Unions: []*ast.Union{
			{
				Name:    "Message",
				Options: []string{"TextMessage"},
			},
		},
		Services: []*ast.Service{
			{
				Name: "MessageService",
				Methods: []*ast.Method{
					{
						Name:       "SendMessage",
						InputType:  "Message",
						OutputType: "Message",
					},
				},
			},
		},
	}

	// Test Protobuf - unions should work as message types
	genProto := NewProtobufGenerator()
	outputProto := genProto.Generate(schema)
	if !strings.Contains(outputProto, "rpc SendMessage(Message) returns (Message)") {
		t.Error("Expected union type used in RPC method")
	}

	// Test GraphQL - unions should work with Input variant
	genGraphQL := NewGraphQLGenerator()
	outputGraphQL := genGraphQL.Generate(schema)
	if !strings.Contains(outputGraphQL, "input: MessageInput") {
		t.Error("Expected MessageInput in GraphQL method input")
	}
}
