package parser

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
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

// TestParser_MergeAnnotations tests the mergeAnnotations function
func TestParser_MergeAnnotations(t *testing.T) {
	p := &Parser{}

	t.Run("both nil", func(t *testing.T) {
		result := p.mergeAnnotations(nil, nil)
		if result != nil {
			t.Error("Expected nil when both annotations are nil")
		}
	})

	t.Run("leading nil", func(t *testing.T) {
		trailing := ast.NewFormatAnnotations()
		trailing.Proto = []string{"option1"}
		result := p.mergeAnnotations(nil, trailing)
		if result != trailing {
			t.Error("Expected trailing when leading is nil")
		}
	})

	t.Run("trailing nil", func(t *testing.T) {
		leading := ast.NewFormatAnnotations()
		leading.Proto = []string{"option1"}
		result := p.mergeAnnotations(leading, nil)
		if result != leading {
			t.Error("Expected leading when trailing is nil")
		}
	})

	t.Run("merge proto annotations", func(t *testing.T) {
		leading := ast.NewFormatAnnotations()
		leading.Proto = []string{"option1", "option2"}

		trailing := ast.NewFormatAnnotations()
		trailing.Proto = []string{"option3"}

		result := p.mergeAnnotations(leading, trailing)
		if len(result.Proto) != 3 {
			t.Errorf("Expected 3 proto annotations, got %d", len(result.Proto))
		}
		if result.Proto[0] != "option1" || result.Proto[1] != "option2" || result.Proto[2] != "option3" {
			t.Errorf("Proto annotations not merged correctly: %v", result.Proto)
		}
	})

	t.Run("name override - proto", func(t *testing.T) {
		leading := ast.NewFormatAnnotations()
		leading.ProtoName = "LeadingName"

		trailing := ast.NewFormatAnnotations()
		trailing.ProtoName = "TrailingName"

		result := p.mergeAnnotations(leading, trailing)
		if result.ProtoName != "TrailingName" {
			t.Errorf("Expected trailing proto name to override, got %q", result.ProtoName)
		}
	})

	t.Run("name override - graphql empty trailing", func(t *testing.T) {
		leading := ast.NewFormatAnnotations()
		leading.GraphQLName = "LeadingName"

		trailing := ast.NewFormatAnnotations()
		trailing.GraphQLName = ""

		result := p.mergeAnnotations(leading, trailing)
		if result.GraphQLName != "LeadingName" {
			t.Errorf("Expected leading name when trailing is empty, got %q", result.GraphQLName)
		}
	})

	t.Run("complex merge - all annotations", func(t *testing.T) {
		leading := ast.NewFormatAnnotations()
		leading.Proto = []string{"proto1"}
		leading.GraphQL = []string{"graphql1"}
		leading.OpenAPI = []string{"openapi1"}
		leading.Go = []string{"go1"}
		leading.ProtoName = "ProtoName1"
		leading.GraphQLName = "GraphQLName1"

		trailing := ast.NewFormatAnnotations()
		trailing.Proto = []string{"proto2"}
		trailing.GraphQL = []string{"graphql2"}
		trailing.OpenAPI = []string{"openapi2"}
		trailing.Go = []string{"go2"}
		trailing.GraphQLName = "GraphQLName2"
		trailing.OpenAPIName = "OpenAPIName2"

		result := p.mergeAnnotations(leading, trailing)

		// Check all arrays are merged
		if len(result.Proto) != 2 {
			t.Errorf("Expected 2 proto annotations, got %d", len(result.Proto))
		}
		if len(result.GraphQL) != 2 {
			t.Errorf("Expected 2 graphql annotations, got %d", len(result.GraphQL))
		}
		if len(result.OpenAPI) != 2 {
			t.Errorf("Expected 2 openapi annotations, got %d", len(result.OpenAPI))
		}
		if len(result.Go) != 2 {
			t.Errorf("Expected 2 go annotations, got %d", len(result.Go))
		}

		// Check name precedence
		if result.ProtoName != "ProtoName1" {
			t.Errorf("Expected proto name from leading (trailing empty), got %q", result.ProtoName)
		}
		if result.GraphQLName != "GraphQLName2" {
			t.Errorf("Expected graphql name from trailing, got %q", result.GraphQLName)
		}
		if result.OpenAPIName != "OpenAPIName2" {
			t.Errorf("Expected openapi name from trailing, got %q", result.OpenAPIName)
		}
	})
}

// TestParser_ParseValidationRules_Comprehensive tests comprehensive validation rule parsing
func TestParser_ParseValidationRules_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		check    func(*testing.T, *ast.ValidationRules)
		hasError bool
	}{
		{
			name: "format validation",
			input: `
namespace test
type User {
  email: string @validate(format="email")
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.Format != "email" {
					t.Errorf("Expected format 'email', got %q", rules.Format)
				}
			},
		},
		{
			name: "pattern validation",
			input: `
namespace test
type User {
  username: string @validate(pattern="^[a-z0-9_]+$")
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.Pattern != "^[a-z0-9_]+$" {
					t.Errorf("Expected pattern '^[a-z0-9_]+$', got %q", rules.Pattern)
				}
			},
		},
		{
			name: "string length validation",
			input: `
namespace test
type User {
  name: string @validate(minLength=3, maxLength=50)
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.MinLength == nil || *rules.MinLength != 3 {
					t.Errorf("Expected minLength 3, got %v", ptrIntValue(rules.MinLength))
				}
				if rules.MaxLength == nil || *rules.MaxLength != 50 {
					t.Errorf("Expected maxLength 50, got %v", ptrIntValue(rules.MaxLength))
				}
			},
		},
		{
			name: "numeric range validation",
			input: `
namespace test
type Product {
  price: float64 @validate(min=0, max=1000000)
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.Min == nil || *rules.Min != 0 {
					t.Errorf("Expected min 0.0, got %v", ptrFloatValue(rules.Min))
				}
				if rules.Max == nil || *rules.Max != 1000000 {
					t.Errorf("Expected max 1000000.0, got %v", ptrFloatValue(rules.Max))
				}
			},
		},
		{
			name: "exclusive range validation",
			input: `
namespace test
type Percentage {
  value: float32 @validate(exclusiveMin=0, exclusiveMax=100)
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.ExclusiveMin == nil || *rules.ExclusiveMin != 0 {
					t.Errorf("Expected exclusiveMin 0, got %v", ptrFloatValue(rules.ExclusiveMin))
				}
				if rules.ExclusiveMax == nil || *rules.ExclusiveMax != 100 {
					t.Errorf("Expected exclusiveMax 100, got %v", ptrFloatValue(rules.ExclusiveMax))
				}
			},
		},
		{
			name: "multipleOf validation",
			input: `
namespace test
type Config {
  port: int32 @validate(multipleOf=10)
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.MultipleOf == nil || *rules.MultipleOf != 10 {
					t.Errorf("Expected multipleOf 10, got %v", ptrFloatValue(rules.MultipleOf))
				}
			},
		},
		{
			name: "array validation",
			input: `
namespace test
type Config {
  tags: []string @validate(minItems=1, maxItems=10, uniqueItems=true)
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.MinItems == nil || *rules.MinItems != 1 {
					t.Errorf("Expected minItems 1, got %v", ptrIntValue(rules.MinItems))
				}
				if rules.MaxItems == nil || *rules.MaxItems != 10 {
					t.Errorf("Expected maxItems 10, got %v", ptrIntValue(rules.MaxItems))
				}
				if !rules.UniqueItems {
					t.Error("Expected uniqueItems to be true")
				}
			},
		},
		{
			name: "complex validation - multiple rules",
			input: `
namespace test
type User {
  email: string @validate(format="email", minLength=5, maxLength=100, pattern=".*@.*")
}`,
			check: func(t *testing.T, rules *ast.ValidationRules) {
				if rules.Format != "email" {
					t.Errorf("Expected format 'email', got %q", rules.Format)
				}
				if rules.Pattern != ".*@.*" {
					t.Errorf("Expected pattern '.*@.*', got %q", rules.Pattern)
				}
				if rules.MinLength == nil || *rules.MinLength != 5 {
					t.Errorf("Expected minLength 5, got %v", ptrIntValue(rules.MinLength))
				}
				if rules.MaxLength == nil || *rules.MaxLength != 100 {
					t.Errorf("Expected maxLength 100, got %v", ptrIntValue(rules.MaxLength))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.hasError {
				if len(p.Errors()) == 0 {
					t.Error("Expected parser errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected parser errors: %v", p.Errors())
			}

			if len(schema.Types) == 0 || len(schema.Types[0].Fields) == 0 {
				t.Fatal("No types or fields parsed")
			}

			field := schema.Types[0].Fields[0]
			if field.Validation == nil {
				t.Fatal("Expected field to have validation rules")
			}

			tt.check(t, field.Validation)
		})
	}
}

// Helper functions
func ptrIntValue(ptr *int) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func ptrFloatValue(ptr *float64) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}
