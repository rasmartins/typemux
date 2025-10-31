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

func TestConvertEnum(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"Status": {
					Type:        "string",
					Description: "Status enum",
					Enum:        []interface{}{"active", "inactive", "pending"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Should generate enum type
	if !strings.Contains(result, "enum Status {") {
		t.Errorf("expected Status enum declaration, got:\n%s", result)
	}

	// Should have all enum values
	if !strings.Contains(result, "ACTIVE = 0") {
		t.Error("expected ACTIVE enum value")
	}

	if !strings.Contains(result, "INACTIVE = 1") {
		t.Error("expected INACTIVE enum value")
	}

	if !strings.Contains(result, "PENDING = 2") {
		t.Error("expected PENDING enum value")
	}

	// Description should be converted to comment
	if !strings.Contains(result, "// Status enum") {
		t.Error("expected enum description as comment")
	}
}

func TestConvertService(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "User API",
			Version: "1.0.0",
		},
		Paths: map[string]*PathItem{
			"/users": {
				Get: &Operation{
					OperationID: "listUsers",
					Summary:     "List all users",
					Responses: map[string]*Response{
						"200": {
							Description: "Success",
						},
					},
				},
				Post: &Operation{
					OperationID: "createUser",
					Summary:     "Create a new user",
					Responses: map[string]*Response{
						"201": {
							Description: "Created",
						},
					},
				},
			},
			"/users/{id}": {
				Get: &Operation{
					OperationID: "getUser",
					Summary:     "Get user by ID",
					Responses: map[string]*Response{
						"200": {
							Description: "Success",
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Should generate service
	if !strings.Contains(result, "service UserAPIService {") {
		t.Errorf("expected service declaration, got:\n%s", result)
	}

	// Should have all RPC methods
	if !strings.Contains(result, "rpc ListUsers") {
		t.Error("expected ListUsers method")
	}

	if !strings.Contains(result, "rpc CreateUser") {
		t.Error("expected CreateUser method")
	}

	if !strings.Contains(result, "rpc GetUser") {
		t.Error("expected GetUser method")
	}

	// Should have HTTP method and path comments
	if !strings.Contains(result, "// GET /users") {
		t.Error("expected GET /users comment")
	}

	if !strings.Contains(result, "// POST /users") {
		t.Error("expected POST /users comment")
	}
}

func TestConvertMethodWithRequestBody(t *testing.T) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   "API",
			Version: "1.0.0",
		},
		Components: &Components{
			Schemas: map[string]*Schema{
				"User": {
					Type: "object",
					Properties: map[string]*Schema{
						"name": {Type: "string"},
					},
				},
			},
		},
		Paths: map[string]*PathItem{
			"/users": {
				Post: &Operation{
					OperationID: "createUser",
					Description: "Creates a new user in the system",
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]*MediaType{
							"application/json": {
								Schema: &Schema{
									Ref: "#/components/schemas/User",
								},
							},
						},
					},
					Responses: map[string]*Response{
						"201": {
							Description: "Created",
							Content: map[string]*MediaType{
								"application/json": {
									Schema: &Schema{
										Ref: "#/components/schemas/User",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(spec)

	// Should use request/response types
	if !strings.Contains(result, "rpc CreateUser(CreateUserRequest) returns (User)") {
		t.Errorf("expected CreateUser method with request body, got:\n%s", result)
	}

	// Should have description comment
	if !strings.Contains(result, "// Creates a new user in the system") {
		t.Error("expected method description comment")
	}
}

func TestGenerateMethodName(t *testing.T) {
	tests := []struct {
		path     string
		method   string
		expected string
	}{
		{"/users", "GET", "getUsers"},
		{"/users", "POST", "postUsers"},
		{"/users/{id}", "GET", "getUsers"},
		{"/users/{id}", "DELETE", "deleteUsers"},
		{"/api/v1/products", "GET", "getApiV1Products"},
		{"/orders/{orderId}/items", "POST", "postOrdersItems"},
	}

	for _, tt := range tests {
		result := generateMethodName(tt.path, tt.method)
		if result != tt.expected {
			t.Errorf("generateMethodName(%q, %q) = %q, expected %q", tt.path, tt.method, result, tt.expected)
		}
	}
}

func TestEscapeFieldName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"id", "id"},
		{"name", "name"},
		{"namespace", "namespace_"},
		{"import", "import_"},
		{"type", "type_"},
		{"enum", "enum_"},
		{"service", "service_"},
		{"rpc", "rpc_"},
	}

	for _, tt := range tests {
		result := escapeFieldName(tt.name)
		if result != tt.expected {
			t.Errorf("escapeFieldName(%q) = %q, expected %q", tt.name, result, tt.expected)
		}
	}
}

func TestConvertSchemaTypeEdgeCases(t *testing.T) {
	converter := NewConverter()
	converter.spec = &OpenAPISpec{}

	tests := []struct {
		name     string
		schema   *Schema
		expected string
	}{
		{
			name:     "array without items",
			schema:   &Schema{Type: "array"},
			expected: "[]string",
		},
		{
			name:     "generic object without properties",
			schema:   &Schema{Type: "object"},
			expected: "map<string, string>",
		},
		{
			name:     "string with date-time format",
			schema:   &Schema{Type: "string", Format: "date-time"},
			expected: "timestamp",
		},
		{
			name:     "string with date format",
			schema:   &Schema{Type: "string", Format: "date"},
			expected: "timestamp",
		},
		{
			name:     "integer with int64 format",
			schema:   &Schema{Type: "integer", Format: "int64"},
			expected: "int64",
		},
		{
			name:     "number with double format",
			schema:   &Schema{Type: "number", Format: "double"},
			expected: "double",
		},
		{
			name:     "number with float format",
			schema:   &Schema{Type: "number"},
			expected: "float",
		},
		{
			name:     "boolean type",
			schema:   &Schema{Type: "boolean"},
			expected: "bool",
		},
		{
			name:     "unknown type defaults to string",
			schema:   &Schema{Type: "unknown"},
			expected: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.convertSchemaType(tt.schema)
			if result != tt.expected {
				t.Errorf("convertSchemaType() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
