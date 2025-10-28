package subscriptions

import (
	"os"
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/generator"
	"github.com/rasmartins/typemux/internal/lexer"
	"github.com/rasmartins/typemux/internal/parser"
)

func TestSubscriptionsExample(t *testing.T) {
	// Read the TypeMUX schema
	content, err := os.ReadFile("chat.typemux")
	if err != nil {
		t.Fatalf("Failed to read chat.typemux: %v", err)
	}

	// Parse the schema
	l := lexer.New(string(content))
	p := parser.New(l)
	schema := p.Parse()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse chat.typemux: %v", p.Errors())
	}

	// Generate GraphQL schema
	gen := generator.NewGraphQLGenerator()
	output := gen.Generate(schema)

	// Verify Query type exists with correct methods
	if !strings.Contains(output, "type Query {") {
		t.Error("Expected Query type to be generated")
	}
	if !strings.Contains(output, "getMessage") {
		t.Error("Expected getMessage in Query type")
	}
	if !strings.Contains(output, "listMessages") {
		t.Error("Expected listMessages in Query type")
	}

	// Verify Mutation type exists with correct methods
	if !strings.Contains(output, "type Mutation {") {
		t.Error("Expected Mutation type to be generated")
	}
	if !strings.Contains(output, "sendMessage") {
		t.Error("Expected sendMessage in Mutation type")
	}
	if !strings.Contains(output, "deleteMessage") {
		t.Error("Expected deleteMessage in Mutation type")
	}

	// Verify Subscription type exists with correct methods
	if !strings.Contains(output, "type Subscription {") {
		t.Error("Expected Subscription type to be generated")
	}
	if !strings.Contains(output, "watchMessages") {
		t.Error("Expected watchMessages in Subscription type")
	}
	if !strings.Contains(output, "watchMessagesBySender") {
		t.Error("Expected watchMessagesBySender in Subscription type")
	}

	// Verify streaming methods are NOT in Query or Mutation
	lines := strings.Split(output, "\n")
	inQuery := false
	inMutation := false
	inSubscription := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "type Query {") {
			inQuery = true
			inMutation = false
			inSubscription = false
			continue
		} else if strings.HasPrefix(trimmed, "type Mutation {") {
			inQuery = false
			inMutation = true
			inSubscription = false
			continue
		} else if strings.HasPrefix(trimmed, "type Subscription {") {
			inQuery = false
			inMutation = false
			inSubscription = true
			continue
		} else if trimmed == "}" {
			inQuery = false
			inMutation = false
			inSubscription = false
			continue
		}

		// Check that streaming methods are only in Subscription
		if (inQuery || inMutation) && (strings.Contains(trimmed, "watchMessages") || strings.Contains(trimmed, "watchMessagesBySender")) {
			t.Errorf("Streaming method found in %s: %s", map[bool]string{true: "Query", false: "Mutation"}[inQuery], trimmed)
		}

		// Check that non-streaming methods are not in Subscription
		if inSubscription {
			if strings.Contains(trimmed, "getMessage") || strings.Contains(trimmed, "listMessages") ||
				strings.Contains(trimmed, "sendMessage") || strings.Contains(trimmed, "deleteMessage") {
				t.Errorf("Non-streaming method found in Subscription: %s", trimmed)
			}
		}
	}

	// Verify Message type is generated
	if !strings.Contains(output, "type Message {") {
		t.Error("Expected Message type to be generated")
	}

	// Save the generated GraphQL schema for manual inspection
	err = os.WriteFile("chat.graphql", []byte(output), 0644)
	if err != nil {
		t.Logf("Warning: Could not write chat.graphql: %v", err)
	}
}

func TestProtobufGeneration(t *testing.T) {
	// Read the TypeMUX schema
	content, err := os.ReadFile("chat.typemux")
	if err != nil {
		t.Fatalf("Failed to read chat.typemux: %v", err)
	}

	// Parse the schema
	l := lexer.New(string(content))
	p := parser.New(l)
	schema := p.Parse()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse chat.typemux: %v", p.Errors())
	}

	// Generate Protobuf schema
	gen := generator.NewProtobufGenerator()
	output := gen.Generate(schema)

	// Verify service is generated
	if !strings.Contains(output, "service ChatService {") {
		t.Error("Expected ChatService to be generated")
	}

	// Verify streaming methods have 'stream' keyword in proto
	if !strings.Contains(output, "returns (stream Message)") {
		t.Error("Expected stream keyword in Protobuf for streaming methods")
	}

	// Verify both streaming methods exist
	if !strings.Contains(output, "rpc WatchMessages") {
		t.Error("Expected WatchMessages method")
	}
	if !strings.Contains(output, "rpc WatchMessagesBySender") {
		t.Error("Expected WatchMessagesBySender method")
	}

	// Save the generated Protobuf schema for manual inspection
	err = os.WriteFile("chat.proto", []byte(output), 0644)
	if err != nil {
		t.Logf("Warning: Could not write chat.proto: %v", err)
	}
}

func TestOpenAPIGeneration(t *testing.T) {
	// Read the TypeMUX schema
	content, err := os.ReadFile("chat.typemux")
	if err != nil {
		t.Fatalf("Failed to read chat.typemux: %v", err)
	}

	// Parse the schema
	l := lexer.New(string(content))
	p := parser.New(l)
	schema := p.Parse()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse chat.typemux: %v", p.Errors())
	}

	// Generate OpenAPI schema
	gen := generator.NewOpenAPIGenerator()
	output := gen.Generate(schema)

	// OpenAPI doesn't support subscriptions natively, but all methods
	// should be generated as regular endpoints
	if !strings.Contains(output, "/getmessage") && !strings.Contains(output, "/getMessage") && !strings.Contains(output, "/GetMessage") {
		t.Error("Expected getMessage endpoint in OpenAPI")
	}

	if !strings.Contains(output, "post") || !strings.Contains(output, "get") {
		t.Error("Expected HTTP methods in OpenAPI")
	}

	// Verify streaming methods are included (OpenAPI treats them as regular endpoints)
	if !strings.Contains(output, "watchmessages") && !strings.Contains(output, "watchMessages") && !strings.Contains(output, "WatchMessages") {
		t.Error("Expected watchMessages endpoint in OpenAPI")
	}

	// Save the generated OpenAPI schema for manual inspection
	err = os.WriteFile("chat.openapi.yaml", []byte(output), 0644)
	if err != nil {
		t.Logf("Warning: Could not write chat.openapi.yaml: %v", err)
	}
}
