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

	// Both fields should be present (required fields don't use ? in TypeMUX)
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

func TestConvertOneOf(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"Dog": {
					Type: "object",
					Properties: map[string]*Schema{
						"breed": {Type: "string"},
					},
				},
				"Cat": {
					Type: "object",
					Properties: map[string]*Schema{
						"meow": {Type: "boolean"},
					},
				},
				"Pet": {
					OneOf: []*Schema{
						{Ref: "#/components/schemas/Dog"},
						{Ref: "#/components/schemas/Cat"},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Should generate a union type
	if !strings.Contains(result, "union Pet {") {
		t.Errorf("expected Pet union declaration, got:\n%s", result)
	}

	// Should have Dog and Cat as variants
	if !strings.Contains(result, "Dog: Dog") {
		t.Errorf("expected Dog variant, got:\n%s", result)
	}

	if !strings.Contains(result, "Cat: Cat") {
		t.Errorf("expected Cat variant, got:\n%s", result)
	}

	// Should also have Dog and Cat type definitions
	if !strings.Contains(result, "type Dog {") {
		t.Error("expected Dog type declaration")
	}

	if !strings.Contains(result, "type Cat {") {
		t.Error("expected Cat type declaration")
	}
}

func TestConvertAnyOf(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"StringValue": {
					Type: "object",
					Properties: map[string]*Schema{
						"value": {Type: "string"},
					},
				},
				"NumberValue": {
					Type: "object",
					Properties: map[string]*Schema{
						"value": {Type: "integer"},
					},
				},
				"FlexibleValue": {
					AnyOf: []*Schema{
						{Ref: "#/components/schemas/StringValue"},
						{Ref: "#/components/schemas/NumberValue"},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Should generate a union type for anyOf
	if !strings.Contains(result, "union FlexibleValue {") {
		t.Errorf("expected FlexibleValue union declaration, got:\n%s", result)
	}

	// Should have StringValue and NumberValue as variants
	if !strings.Contains(result, "StringValue: StringValue") {
		t.Errorf("expected StringValue variant, got:\n%s", result)
	}

	if !strings.Contains(result, "NumberValue: NumberValue") {
		t.Errorf("expected NumberValue variant, got:\n%s", result)
	}
}

func TestConvertOneOfPrimitives(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"StringOrNumber": {
					OneOf: []*Schema{
						{Type: "string"},
						{Type: "integer"},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Should generate a union type with primitive variants
	if !strings.Contains(result, "union StringOrNumber {") {
		t.Errorf("expected StringOrNumber union declaration, got:\n%s", result)
	}

	// Should have variants for string and int
	if !strings.Contains(result, "StringValue: string") {
		t.Errorf("expected StringValue variant, got:\n%s", result)
	}

	if !strings.Contains(result, "Int32Value: int32") {
		t.Errorf("expected Int32Value variant, got:\n%s", result)
	}
}

func TestConvertComplexOneOf(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Payment API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"CreditCard": {
					Type: "object",
					Properties: map[string]*Schema{
						"cardNumber": {Type: "string"},
						"cvv":        {Type: "string"},
					},
				},
				"BankTransfer": {
					Type: "object",
					Properties: map[string]*Schema{
						"accountNumber": {Type: "string"},
						"routingNumber": {Type: "string"},
					},
				},
				"PayPal": {
					Type: "object",
					Properties: map[string]*Schema{
						"email": {Type: "string"},
					},
				},
				"PaymentMethod": {
					Description: "Payment method union type",
					OneOf: []*Schema{
						{Ref: "#/components/schemas/CreditCard"},
						{Ref: "#/components/schemas/BankTransfer"},
						{Ref: "#/components/schemas/PayPal"},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Should generate union type
	if !strings.Contains(result, "union PaymentMethod {") {
		t.Errorf("expected PaymentMethod union, got:\n%s", result)
	}

	// Should have all three payment methods as variants
	if !strings.Contains(result, "CreditCard: CreditCard") {
		t.Error("expected CreditCard variant")
	}

	if !strings.Contains(result, "BankTransfer: BankTransfer") {
		t.Error("expected BankTransfer variant")
	}

	if !strings.Contains(result, "PayPal: PayPal") {
		t.Error("expected PayPal variant")
	}

	// Description should be converted to comment
	if !strings.Contains(result, "// Payment method union type") {
		t.Error("expected union description as comment")
	}
}
