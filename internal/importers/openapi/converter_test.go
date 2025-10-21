package openapi

import (
	"strings"
	"testing"
)

func TestConvertBasicSchema(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"User": {
					Type: "object",
					Properties: map[string]*Schema{
						"id":   {Type: "string"},
						"name": {Type: "string"},
						"age":  {Type: "integer"},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	if !strings.Contains(result, "@typemux") {
		t.Error("expected typemux annotation")
	}

	if !strings.Contains(result, "type User {") {
		t.Error("expected User type declaration")
	}

	if !strings.Contains(result, "id: string") {
		t.Error("expected id field")
	}

	if !strings.Contains(result, "name: string") {
		t.Error("expected name field")
	}

	if !strings.Contains(result, "age: int32") {
		t.Error("expected age field")
	}
}

func TestConvertArrayType(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"UserList": {
					Type: "object",
					Properties: map[string]*Schema{
						"users": {
							Type: "array",
							Items: &Schema{
								Type: "string",
							},
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	if !strings.Contains(result, "users: []string") {
		t.Error("expected users as array of string")
	}
}

func TestConvertRequiredFields(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"User": {
					Type: "object",
					Properties: map[string]*Schema{
						"id":   {Type: "string"},
						"name": {Type: "string"},
					},
					Required: []string{"id", "name"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Both fields should be present (required fields don't use ? in TypeMux)
	if !strings.Contains(result, "id: string") {
		t.Error("expected id field")
	}

	if !strings.Contains(result, "name: string") {
		t.Error("expected name field")
	}
}

func TestConvertDescription(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:       "Test API",
			Version:     "1.0.0",
			Description: "API description",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"User": {
					Type:        "object",
					Description: "User object",
					Properties: map[string]*Schema{
						"id": {
							Type:        "string",
							Description: "User ID",
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Descriptions should be converted to comments
	if !strings.Contains(result, "//") {
		t.Error("expected descriptions to be converted to comments")
	}
}

func TestConvertEmptySpec(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Empty API",
			Version: "1.0.0",
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	if !strings.Contains(result, "@typemux") {
		t.Errorf("expected typemux annotation, got:\n%s", result)
	}

	if !strings.Contains(result, "namespace") {
		t.Errorf("expected namespace declaration, got:\n%s", result)
	}
}

func TestConvertRef(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"Pet": {
					Type: "object",
					Properties: map[string]*Schema{
						"owner": {
							Ref: "#/components/schemas/User",
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Ref should be resolved to User type
	if !strings.Contains(result, "owner: User") {
		t.Error("expected owner field with User type")
	}
}
