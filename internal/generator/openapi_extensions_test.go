package generator

import (
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestOpenAPIGenerator_ParseExtensions_Simple(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "single string property",
			input: `{"x-custom": "value"}`,
			expected: map[string]interface{}{
				"x-custom": "value",
			},
		},
		{
			name:  "multiple properties",
			input: `{"x-internal-id": "prod-v1", "x-category": "commerce"}`,
			expected: map[string]interface{}{
				"x-internal-id": "prod-v1",
				"x-category":    "commerce",
			},
		},
		{
			name:  "number property",
			input: `{"x-precision": 2}`,
			expected: map[string]interface{}{
				"x-precision": float64(2),
			},
		},
		{
			name:  "boolean property",
			input: `{"x-internal": true}`,
			expected: map[string]interface{}{
				"x-internal": true,
			},
		},
		{
			name:  "mixed types",
			input: `{"x-format": "currency", "x-precision": 2, "x-required": false}`,
			expected: map[string]interface{}{
				"x-format":    "currency",
				"x-precision": float64(2),
				"x-required":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.parseExtensions(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d properties, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				actualValue, ok := result[key]
				if !ok {
					t.Errorf("Expected key %s not found in result", key)
					continue
				}

				// Compare values based on type
				switch expected := expectedValue.(type) {
				case string:
					if actual, ok := actualValue.(string); !ok || actual != expected {
						t.Errorf("For key %s: expected %v, got %v", key, expected, actualValue)
					}
				case float64:
					if actual, ok := actualValue.(float64); !ok || actual != expected {
						t.Errorf("For key %s: expected %v, got %v", key, expected, actualValue)
					}
				case bool:
					if actual, ok := actualValue.(bool); !ok || actual != expected {
						t.Errorf("For key %s: expected %v, got %v", key, expected, actualValue)
					}
				}
			}
		})
	}
}

func TestOpenAPIGenerator_ParseExtensions_Nested(t *testing.T) {
	gen := NewOpenAPIGenerator()

	input := `{"x-metadata": {"version": "v2", "internal": true}}`
	result := gen.parseExtensions(input)

	if len(result) != 1 {
		t.Fatalf("Expected 1 top-level property, got %d", len(result))
	}

	metadata, ok := result["x-metadata"]
	if !ok {
		t.Fatal("Expected x-metadata property")
	}

	metadataMap, ok := metadata.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected x-metadata to be a map, got %T", metadata)
	}

	if version, ok := metadataMap["version"].(string); !ok || version != "v2" {
		t.Errorf("Expected version to be 'v2', got %v", metadataMap["version"])
	}

	if internal, ok := metadataMap["internal"].(bool); !ok || !internal {
		t.Errorf("Expected internal to be true, got %v", metadataMap["internal"])
	}
}

func TestOpenAPIGenerator_ParseExtensions_Array(t *testing.T) {
	gen := NewOpenAPIGenerator()

	input := `{"x-features": ["auth", "caching", "logging"]}`
	result := gen.parseExtensions(input)

	if len(result) != 1 {
		t.Fatalf("Expected 1 property, got %d", len(result))
	}

	features, ok := result["x-features"]
	if !ok {
		t.Fatal("Expected x-features property")
	}

	featuresArray, ok := features.([]interface{})
	if !ok {
		t.Fatalf("Expected x-features to be an array, got %T", features)
	}

	if len(featuresArray) != 3 {
		t.Errorf("Expected 3 features, got %d", len(featuresArray))
	}

	expectedFeatures := []string{"auth", "caching", "logging"}
	for i, expected := range expectedFeatures {
		if actual, ok := featuresArray[i].(string); !ok || actual != expected {
			t.Errorf("Expected feature[%d] to be '%s', got %v", i, expected, featuresArray[i])
		}
	}
}

func TestOpenAPIGenerator_ParseExtensions_DeeplyNested(t *testing.T) {
	gen := NewOpenAPIGenerator()

	input := `{"x-validation": {"min": 1, "max": 3600, "default": 30}}`
	result := gen.parseExtensions(input)

	validation, ok := result["x-validation"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected x-validation to be a map")
	}

	if min, ok := validation["min"].(float64); !ok || min != 1 {
		t.Errorf("Expected min to be 1, got %v", validation["min"])
	}

	if max, ok := validation["max"].(float64); !ok || max != 3600 {
		t.Errorf("Expected max to be 3600, got %v", validation["max"])
	}

	if defaultVal, ok := validation["default"].(float64); !ok || defaultVal != 30 {
		t.Errorf("Expected default to be 30, got %v", validation["default"])
	}
}

func TestOpenAPIGenerator_ParseExtensions_InvalidJSON(t *testing.T) {
	gen := NewOpenAPIGenerator()

	tests := []struct {
		name  string
		input string
	}{
		{"incomplete json", `{"x-custom": "value"`},
		{"invalid syntax", `{x-custom: value}`},
		{"empty string", ``},
		{"not json object", `"just a string"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.parseExtensions(tt.input)
			if len(result) != 0 {
				t.Errorf("Expected empty map for invalid JSON, got %d properties", len(result))
			}
		})
	}
}

func TestOpenAPIGenerator_GenerateSchema_WithExtensions(t *testing.T) {
	gen := NewOpenAPIGenerator()

	typ := &ast.Type{
		Name: "Product",
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
		Annotations: &ast.FormatAnnotations{
			OpenAPI: []string{
				`{"x-internal-id": "prod-v1", "x-category": "commerce"}`,
			},
		},
	}

	schema := gen.generateSchema(typ, make(map[string]string))

	if len(schema.Extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(schema.Extensions))
	}

	if schema.Extensions["x-internal-id"] != "prod-v1" {
		t.Errorf("Expected x-internal-id to be 'prod-v1', got %v", schema.Extensions["x-internal-id"])
	}

	if schema.Extensions["x-category"] != "commerce" {
		t.Errorf("Expected x-category to be 'commerce', got %v", schema.Extensions["x-category"])
	}
}

func TestOpenAPIGenerator_ConvertFieldToProperty_WithExtensions(t *testing.T) {
	gen := NewOpenAPIGenerator()

	field := &ast.Field{
		Name: "price",
		Type: &ast.FieldType{
			Name:      "float64",
			IsBuiltin: true,
		},
		Required: true,
		Annotations: &ast.FormatAnnotations{
			OpenAPI: []string{
				`{"x-format": "currency", "x-precision": 2}`,
			},
		},
	}

	property := gen.convertFieldToProperty(field, make(map[string]string))

	if len(property.Extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(property.Extensions))
	}

	if property.Extensions["x-format"] != "currency" {
		t.Errorf("Expected x-format to be 'currency', got %v", property.Extensions["x-format"])
	}

	if precision, ok := property.Extensions["x-precision"].(float64); !ok || precision != 2 {
		t.Errorf("Expected x-precision to be 2, got %v", property.Extensions["x-precision"])
	}
}

func TestOpenAPIGenerator_MultipleExtensionAnnotations(t *testing.T) {
	gen := NewOpenAPIGenerator()

	typ := &ast.Type{
		Name:   "Config",
		Fields: []*ast.Field{},
		Annotations: &ast.FormatAnnotations{
			OpenAPI: []string{
				`{"x-internal": true}`,
				`{"x-version": "v2"}`,
				`{"x-features": ["auth", "cache"]}`,
			},
		},
	}

	schema := gen.generateSchema(typ, make(map[string]string))

	// Should merge all extensions
	if len(schema.Extensions) != 3 {
		t.Errorf("Expected 3 extensions from multiple annotations, got %d", len(schema.Extensions))
	}

	if schema.Extensions["x-internal"] != true {
		t.Errorf("Expected x-internal to be true, got %v", schema.Extensions["x-internal"])
	}

	if schema.Extensions["x-version"] != "v2" {
		t.Errorf("Expected x-version to be 'v2', got %v", schema.Extensions["x-version"])
	}

	if features, ok := schema.Extensions["x-features"].([]interface{}); !ok {
		t.Errorf("Expected x-features to be an array, got %T", schema.Extensions["x-features"])
	} else if len(features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(features))
	}
}

func TestOpenAPIGenerator_Generate_WithExtensions(t *testing.T) {
	gen := NewOpenAPIGenerator()

	schema := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "Product",
				Fields: []*ast.Field{
					{
						Name: "price",
						Type: &ast.FieldType{
							Name:      "float64",
							IsBuiltin: true,
						},
						Required: true,
						Annotations: &ast.FormatAnnotations{
							OpenAPI: []string{
								`{"x-format": "currency"}`,
							},
						},
					},
				},
				Annotations: &ast.FormatAnnotations{
					OpenAPI: []string{
						`{"x-category": "commerce"}`,
					},
				},
			},
		},
	}

	output := gen.Generate(schema)

	// Check that extensions appear in YAML output
	if !strings.Contains(output, "x-category: commerce") {
		t.Error("Expected x-category extension in output")
	}

	if !strings.Contains(output, "x-format: currency") {
		t.Error("Expected x-format extension in output")
	}
}
