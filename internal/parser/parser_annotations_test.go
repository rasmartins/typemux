package parser

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/lexer"
)

func TestParseProtoOption(t *testing.T) {
	input := `
type User {
	tags: []string @proto.option([packed = false])
	metadata: bytes @proto.option([retention = RETENTION_SOURCE])
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]
	if typ.Name != "User" {
		t.Errorf("Expected type name 'User', got %s", typ.Name)
	}

	// Check tags field
	tagsField := typ.Fields[0]
	if tagsField.Annotations == nil {
		t.Fatal("Expected annotations on tags field")
	}
	if len(tagsField.Annotations.Proto) != 1 {
		t.Fatalf("Expected 1 proto annotation, got %d", len(tagsField.Annotations.Proto))
	}
	if tagsField.Annotations.Proto[0] != "[packed = false]" {
		t.Errorf("Expected '[packed = false]', got %s", tagsField.Annotations.Proto[0])
	}

	// Check metadata field
	metadataField := typ.Fields[1]
	if metadataField.Annotations == nil {
		t.Fatal("Expected annotations on metadata field")
	}
	if len(metadataField.Annotations.Proto) != 1 {
		t.Fatalf("Expected 1 proto annotation, got %d", len(metadataField.Annotations.Proto))
	}
	if metadataField.Annotations.Proto[0] != "[retention = RETENTION_SOURCE]" {
		t.Errorf("Expected '[retention = RETENTION_SOURCE]', got %s", metadataField.Annotations.Proto[0])
	}
}

func TestParseGraphQLDirective(t *testing.T) {
	input := `
type User @graphql.directive(@key(fields: "id")) {
	id: string @required @graphql.directive(@external)
	email: string @required
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]

	// Check type-level annotation
	if typ.Annotations == nil {
		t.Fatal("Expected annotations on type")
	}
	if len(typ.Annotations.GraphQL) != 1 {
		t.Fatalf("Expected 1 graphql annotation on type, got %d", len(typ.Annotations.GraphQL))
	}
	expected := `@key(fields:"id")`
	if typ.Annotations.GraphQL[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, typ.Annotations.GraphQL[0])
	}

	// Check field-level annotation
	idField := typ.Fields[0]
	if idField.Annotations == nil {
		t.Fatal("Expected annotations on id field")
	}
	if len(idField.Annotations.GraphQL) != 1 {
		t.Fatalf("Expected 1 graphql annotation on field, got %d", len(idField.Annotations.GraphQL))
	}
	if idField.Annotations.GraphQL[0] != "@external" {
		t.Errorf("Expected '@external', got '%s'", idField.Annotations.GraphQL[0])
	}
}

func TestParseOpenAPIExtension(t *testing.T) {
	input := `
type Product @openapi.extension({"x-internal-id": "prod-v1", "x-category": "commerce"}) {
	id: string @required
	price: float64 @required @openapi.extension({"x-format": "currency", "x-precision": 2})
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]

	// Check type-level annotation
	if typ.Annotations == nil {
		t.Fatal("Expected annotations on type")
	}
	if len(typ.Annotations.OpenAPI) != 1 {
		t.Fatalf("Expected 1 openapi annotation on type, got %d", len(typ.Annotations.OpenAPI))
	}

	typeAnnotation := typ.Annotations.OpenAPI[0]
	if !strings.Contains(typeAnnotation, `"x-internal-id"`) {
		t.Errorf("Expected x-internal-id in annotation, got: %s", typeAnnotation)
	}
	if !strings.Contains(typeAnnotation, `"prod-v1"`) {
		t.Errorf("Expected prod-v1 in annotation, got: %s", typeAnnotation)
	}

	// Check field-level annotation
	priceField := typ.Fields[1]
	if priceField.Annotations == nil {
		t.Fatal("Expected annotations on price field")
	}
	if len(priceField.Annotations.OpenAPI) != 1 {
		t.Fatalf("Expected 1 openapi annotation on field, got %d", len(priceField.Annotations.OpenAPI))
	}

	fieldAnnotation := priceField.Annotations.OpenAPI[0]
	if !strings.Contains(fieldAnnotation, `"x-format"`) {
		t.Errorf("Expected x-format in annotation, got: %s", fieldAnnotation)
	}
	if !strings.Contains(fieldAnnotation, `"currency"`) {
		t.Errorf("Expected currency in annotation, got: %s", fieldAnnotation)
	}
}

func TestParseOpenAPIExtensionNested(t *testing.T) {
	input := `
type Config @openapi.extension({"x-metadata": {"version": "v2", "internal": true, "features": ["auth", "caching"]}}) {
	timeout: int32 @default(30) @openapi.extension({"x-validation": {"min": 1, "max": 3600, "default": 30}})
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]

	// Check type-level nested annotation
	if typ.Annotations == nil {
		t.Fatal("Expected annotations on type")
	}
	if len(typ.Annotations.OpenAPI) != 1 {
		t.Fatalf("Expected 1 openapi annotation on type, got %d", len(typ.Annotations.OpenAPI))
	}

	typeAnnotation := typ.Annotations.OpenAPI[0]
	if !strings.Contains(typeAnnotation, `"x-metadata"`) {
		t.Errorf("Expected x-metadata in annotation, got: %s", typeAnnotation)
	}
	if !strings.Contains(typeAnnotation, `"version"`) {
		t.Errorf("Expected version in nested annotation, got: %s", typeAnnotation)
	}
	if !strings.Contains(typeAnnotation, `"v2"`) {
		t.Errorf("Expected v2 in nested annotation, got: %s", typeAnnotation)
	}
	if !strings.Contains(typeAnnotation, `"features"`) {
		t.Errorf("Expected features array in annotation, got: %s", typeAnnotation)
	}

	// Check field-level nested annotation
	timeoutField := typ.Fields[0]
	if timeoutField.Annotations == nil {
		t.Fatal("Expected annotations on timeout field")
	}
	if len(timeoutField.Annotations.OpenAPI) != 1 {
		t.Fatalf("Expected 1 openapi annotation on field, got %d", len(timeoutField.Annotations.OpenAPI))
	}

	fieldAnnotation := timeoutField.Annotations.OpenAPI[0]
	if !strings.Contains(fieldAnnotation, `"x-validation"`) {
		t.Errorf("Expected x-validation in annotation, got: %s", fieldAnnotation)
	}
	if !strings.Contains(fieldAnnotation, `"min"`) {
		t.Errorf("Expected min in nested annotation, got: %s", fieldAnnotation)
	}
	if !strings.Contains(fieldAnnotation, `"max"`) {
		t.Errorf("Expected max in nested annotation, got: %s", fieldAnnotation)
	}
}

func TestParseMultipleAnnotations(t *testing.T) {
	input := `
type User {
	id: string @required @proto.option([jstype = JS_STRING]) @graphql.directive(@external)
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	typ := schema.Types[0]
	field := typ.Fields[0]

	// Should be required
	if !field.Required {
		t.Error("Expected field to be required")
	}

	// Should have proto annotation
	if field.Annotations == nil || len(field.Annotations.Proto) == 0 {
		t.Error("Expected proto annotation")
	} else if field.Annotations.Proto[0] != "[jstype = JS_STRING]" {
		t.Errorf("Expected '[jstype = JS_STRING]', got '%s'", field.Annotations.Proto[0])
	}

	// Should have graphql annotation
	if field.Annotations == nil || len(field.Annotations.GraphQL) == 0 {
		t.Error("Expected graphql annotation")
	} else if field.Annotations.GraphQL[0] != "@external" {
		t.Errorf("Expected '@external', got '%s'", field.Annotations.GraphQL[0])
	}
}

func TestParseAnnotationContentWithParentheses(t *testing.T) {
	input := `
type User @graphql.directive(@key(fields: "id", select: "user { id email }")) {
	id: string
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	typ := schema.Types[0]
	if typ.Annotations == nil || len(typ.Annotations.GraphQL) == 0 {
		t.Fatal("Expected graphql annotation")
	}

	annotation := typ.Annotations.GraphQL[0]
	// Should preserve nested parentheses and braces
	if !strings.Contains(annotation, "user { id email }") {
		t.Errorf("Expected nested braces to be preserved, got: %s", annotation)
	}
}

func TestParseAnnotationWithBrackets(t *testing.T) {
	input := `
type User {
	tags: []string @proto.option([packed = false, deprecated = true])
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	field := schema.Types[0].Fields[0]
	if field.Annotations == nil || len(field.Annotations.Proto) == 0 {
		t.Fatal("Expected proto annotation")
	}

	annotation := field.Annotations.Proto[0]
	expected := "[packed = false, deprecated = true]"
	if annotation != expected {
		t.Errorf("Expected '%s', got '%s'", expected, annotation)
	}
}

func TestParseNameAnnotations(t *testing.T) {
	input := `
type User @proto.name("UserV2") @graphql.name("UserAccount") @openapi.name("UserProfile") {
	id: string @required
}
`
	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]
	if typ.Annotations == nil {
		t.Fatal("Expected type to have annotations")
	}

	if typ.Annotations.ProtoName != "UserV2" {
		t.Errorf("Expected ProtoName to be 'UserV2', got '%s'", typ.Annotations.ProtoName)
	}

	if typ.Annotations.GraphQLName != "UserAccount" {
		t.Errorf("Expected GraphQLName to be 'UserAccount', got '%s'", typ.Annotations.GraphQLName)
	}

	if typ.Annotations.OpenAPIName != "UserProfile" {
		t.Errorf("Expected OpenAPIName to be 'UserProfile', got '%s'", typ.Annotations.OpenAPIName)
	}
}
