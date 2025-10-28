package parser

import (
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
	"github.com/rasmartins/typemux/internal/lexer"
)

func TestParseEnum(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectedValues []string
		expectErrors   bool
	}{
		{
			name:           "simple enum",
			input:          "enum UserRole { ADMIN USER GUEST }",
			expectedName:   "UserRole",
			expectedValues: []string{"ADMIN", "USER", "GUEST"},
			expectErrors:   false,
		},
		{
			name:           "enum with newlines",
			input:          "enum Status {\n  ACTIVE\n  INACTIVE\n}",
			expectedName:   "Status",
			expectedValues: []string{"ACTIVE", "INACTIVE"},
			expectErrors:   false,
		},
		{
			name:           "single value enum",
			input:          "enum Single { VALUE }",
			expectedName:   "Single",
			expectedValues: []string{"VALUE"},
			expectErrors:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.expectErrors {
				if len(p.Errors()) == 0 {
					t.Error("Expected errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
				return
			}

			if len(schema.Enums) != 1 {
				t.Fatalf("Expected 1 enum, got %d", len(schema.Enums))
			}

			enum := schema.Enums[0]
			if enum.Name != tt.expectedName {
				t.Errorf("Expected enum name %q, got %q", tt.expectedName, enum.Name)
			}

			if len(enum.Values) != len(tt.expectedValues) {
				t.Fatalf("Expected %d values, got %d", len(tt.expectedValues), len(enum.Values))
			}

			for i, expectedValue := range tt.expectedValues {
				if enum.Values[i].Name != expectedValue {
					t.Errorf("Value %d: expected %q, got %q", i, expectedValue, enum.Values[i].Name)
				}
			}
		})
	}
}

func TestParseEnumWithNumbers(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectedValues []struct {
			name      string
			number    int
			hasNumber bool
		}
		expectErrors bool
	}{
		{
			name:         "enum with all custom numbers",
			input:        "enum UserRole { ADMIN = 10 USER = 20 GUEST = 30 }",
			expectedName: "UserRole",
			expectedValues: []struct {
				name      string
				number    int
				hasNumber bool
			}{
				{"ADMIN", 10, true},
				{"USER", 20, true},
				{"GUEST", 30, true},
			},
			expectErrors: false,
		},
		{
			name:         "enum with mixed numbering",
			input:        "enum Status { ACTIVE = 1 INACTIVE PENDING = 5 }",
			expectedName: "Status",
			expectedValues: []struct {
				name      string
				number    int
				hasNumber bool
			}{
				{"ACTIVE", 1, true},
				{"INACTIVE", 0, false},
				{"PENDING", 5, true},
			},
			expectErrors: false,
		},
		{
			name:         "enum with sparse numbering",
			input:        "enum Priority { LOW = 100 MEDIUM = 200 HIGH = 300 }",
			expectedName: "Priority",
			expectedValues: []struct {
				name      string
				number    int
				hasNumber bool
			}{
				{"LOW", 100, true},
				{"MEDIUM", 200, true},
				{"HIGH", 300, true},
			},
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.expectErrors {
				if len(p.Errors()) == 0 {
					t.Error("Expected errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
				return
			}

			if len(schema.Enums) != 1 {
				t.Fatalf("Expected 1 enum, got %d", len(schema.Enums))
			}

			enum := schema.Enums[0]
			if enum.Name != tt.expectedName {
				t.Errorf("Expected enum name %q, got %q", tt.expectedName, enum.Name)
			}

			if len(enum.Values) != len(tt.expectedValues) {
				t.Fatalf("Expected %d values, got %d", len(tt.expectedValues), len(enum.Values))
			}

			for i, expected := range tt.expectedValues {
				actual := enum.Values[i]
				if actual.Name != expected.name {
					t.Errorf("Value %d: expected name %q, got %q", i, expected.name, actual.Name)
				}
				if actual.HasNumber != expected.hasNumber {
					t.Errorf("Value %d (%s): expected hasNumber=%v, got %v", i, expected.name, expected.hasNumber, actual.HasNumber)
				}
				if expected.hasNumber && actual.Number != expected.number {
					t.Errorf("Value %d (%s): expected number=%d, got %d", i, expected.name, expected.number, actual.Number)
				}
			}
		})
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		fieldCount   int
		expectErrors bool
	}{
		{
			name: "simple type",
			input: `type User {
				id: string @required
				name: string
			}`,
			expectedName: "User",
			fieldCount:   2,
			expectErrors: false,
		},
		{
			name: "type with various fields",
			input: `type Post {
				id: string @required
				title: string @required
				count: int32
				isPublished: bool
			}`,
			expectedName: "Post",
			fieldCount:   4,
			expectErrors: false,
		},
		{
			name: "type with array field",
			input: `type Container {
				items: []string
			}`,
			expectedName: "Container",
			fieldCount:   1,
			expectErrors: false,
		},
		{
			name: "type with map field",
			input: `type Config {
				settings: map<string, string>
			}`,
			expectedName: "Config",
			fieldCount:   1,
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.expectErrors {
				if len(p.Errors()) == 0 {
					t.Error("Expected errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
				return
			}

			if len(schema.Types) != 1 {
				t.Fatalf("Expected 1 type, got %d", len(schema.Types))
			}

			typ := schema.Types[0]
			if typ.Name != tt.expectedName {
				t.Errorf("Expected type name %q, got %q", tt.expectedName, typ.Name)
			}

			if len(typ.Fields) != tt.fieldCount {
				t.Errorf("Expected %d fields, got %d", tt.fieldCount, len(typ.Fields))
			}
		})
	}
}

func TestParseField(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		fieldName    string
		fieldType    string
		required     bool
		isArray      bool
		isMap        bool
		defaultValue string
	}{
		{
			name:      "required string field",
			input:     "type T { name: string @required }",
			fieldName: "name",
			fieldType: "string",
			required:  true,
		},
		{
			name:      "optional int field",
			input:     "type T { count: int32 }",
			fieldName: "count",
			fieldType: "int32",
			required:  false,
		},
		{
			name:      "array field",
			input:     "type T { tags: []string }",
			fieldName: "tags",
			fieldType: "string",
			isArray:   true,
		},
		{
			name:      "map field",
			input:     "type T { metadata: map<string, string> }",
			fieldName: "metadata",
			fieldType: "map",
			isMap:     true,
		},
		{
			name:         "field with default",
			input:        "type T { active: bool @default(true) }",
			fieldName:    "active",
			fieldType:    "bool",
			defaultValue: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
				return
			}

			if len(schema.Types) != 1 {
				t.Fatalf("Expected 1 type, got %d", len(schema.Types))
			}

			if len(schema.Types[0].Fields) != 1 {
				t.Fatalf("Expected 1 field, got %d", len(schema.Types[0].Fields))
			}

			field := schema.Types[0].Fields[0]

			if field.Name != tt.fieldName {
				t.Errorf("Expected field name %q, got %q", tt.fieldName, field.Name)
			}

			if field.Type.Name != tt.fieldType {
				t.Errorf("Expected field type %q, got %q", tt.fieldType, field.Type.Name)
			}

			if field.Required != tt.required {
				t.Errorf("Expected required=%v, got %v", tt.required, field.Required)
			}

			if field.Type.IsArray != tt.isArray {
				t.Errorf("Expected isArray=%v, got %v", tt.isArray, field.Type.IsArray)
			}

			if field.Type.IsMap != tt.isMap {
				t.Errorf("Expected isMap=%v, got %v", tt.isMap, field.Type.IsMap)
			}

			if tt.defaultValue != "" && field.Default != tt.defaultValue {
				t.Errorf("Expected default=%q, got %q", tt.defaultValue, field.Default)
			}
		})
	}
}

func TestParseService(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		serviceName  string
		methodCount  int
		expectErrors bool
	}{
		{
			name: "simple service",
			input: `service UserService {
				rpc GetUser(GetUserRequest) returns (GetUserResponse)
			}`,
			serviceName: "UserService",
			methodCount: 1,
		},
		{
			name: "service with multiple methods",
			input: `service UserService {
				rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
				rpc GetUser(GetUserRequest) returns (GetUserResponse)
				rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse)
			}`,
			serviceName: "UserService",
			methodCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.expectErrors {
				if len(p.Errors()) == 0 {
					t.Error("Expected errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
				return
			}

			if len(schema.Services) != 1 {
				t.Fatalf("Expected 1 service, got %d", len(schema.Services))
			}

			service := schema.Services[0]
			if service.Name != tt.serviceName {
				t.Errorf("Expected service name %q, got %q", tt.serviceName, service.Name)
			}

			if len(service.Methods) != tt.methodCount {
				t.Errorf("Expected %d methods, got %d", tt.methodCount, len(service.Methods))
			}
		})
	}
}

func TestParseMethod(t *testing.T) {
	input := `service TestService {
		rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
	}`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(schema.Services))
	}

	if len(schema.Services[0].Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(schema.Services[0].Methods))
	}

	method := schema.Services[0].Methods[0]

	if method.Name != "CreateUser" {
		t.Errorf("Expected method name 'CreateUser', got %q", method.Name)
	}

	if method.InputType != "CreateUserRequest" {
		t.Errorf("Expected input type 'CreateUserRequest', got %q", method.InputType)
	}

	if method.OutputType != "CreateUserResponse" {
		t.Errorf("Expected output type 'CreateUserResponse', got %q", method.OutputType)
	}
}

func TestParseCompleteSchema(t *testing.T) {
	input := `
// User roles
enum UserRole {
  ADMIN
  USER
}

// User type
type User {
  id: string @required
  name: string @required
  role: UserRole @required
  tags: []string
  metadata: map<string, string>
}

type GetUserRequest {
  id: string @required
}

type GetUserResponse {
  user: User
}

// User service
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse)
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	// Check enums
	if len(schema.Enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(schema.Enums))
	}

	// Check types
	if len(schema.Types) != 3 {
		t.Errorf("Expected 3 types, got %d", len(schema.Types))
	}

	// Check services
	if len(schema.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(schema.Services))
	}

	// Verify User type structure
	var userType *ast.Type
	for _, typ := range schema.Types {
		if typ.Name == "User" {
			userType = typ
			break
		}
	}

	if userType == nil {
		t.Fatal("User type not found")
	}

	if len(userType.Fields) != 5 {
		t.Errorf("Expected User type to have 5 fields, got %d", len(userType.Fields))
	}

	// Check array field
	var tagsField *ast.Field
	for _, field := range userType.Fields {
		if field.Name == "tags" {
			tagsField = field
			break
		}
	}

	if tagsField == nil {
		t.Fatal("tags field not found")
	}

	if !tagsField.Type.IsArray {
		t.Error("Expected tags field to be an array")
	}

	// Check map field
	var metadataField *ast.Field
	for _, field := range userType.Fields {
		if field.Name == "metadata" {
			metadataField = field
			break
		}
	}

	if metadataField == nil {
		t.Fatal("metadata field not found")
	}

	if !metadataField.Type.IsMap {
		t.Error("Expected metadata field to be a map")
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing enum brace",
			input: "enum UserRole ADMIN",
		},
		{
			name:  "missing type brace",
			input: "type User id: string",
		},
		{
			name:  "missing field colon",
			input: "type User { id string }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.Parse()

			if len(p.Errors()) == 0 {
				t.Error("Expected errors but got none")
			}
		})
	}
}

func TestParseEmptyInput(t *testing.T) {
	l := lexer.New("")
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Enums) != 0 {
		t.Errorf("Expected 0 enums, got %d", len(schema.Enums))
	}

	if len(schema.Types) != 0 {
		t.Errorf("Expected 0 types, got %d", len(schema.Types))
	}

	if len(schema.Services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(schema.Services))
	}
}

func TestParseMultipleEnums(t *testing.T) {
	input := `
enum UserRole { ADMIN USER }
enum Status { ACTIVE INACTIVE }
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Enums) != 2 {
		t.Errorf("Expected 2 enums, got %d", len(schema.Enums))
	}
}

func TestParseMultipleServices(t *testing.T) {
	input := `
service UserService {
  rpc GetUser(Req) returns (Res)
}

service PostService {
  rpc GetPost(Req) returns (Res)
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(schema.Services))
	}
}

func TestParseCustomTypes(t *testing.T) {
	input := `
type Address {
  street: string
}

type User {
  address: Address @required
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Types) != 2 {
		t.Fatalf("Expected 2 types, got %d", len(schema.Types))
	}

	// Find User type
	var userType *ast.Type
	for _, typ := range schema.Types {
		if typ.Name == "User" {
			userType = typ
			break
		}
	}

	if userType == nil {
		t.Fatal("User type not found")
	}

	if userType.Fields[0].Type.Name != "Address" {
		t.Errorf("Expected field type 'Address', got %q", userType.Fields[0].Type.Name)
	}

	if userType.Fields[0].Type.IsBuiltin {
		t.Error("Expected Address to not be a builtin type")
	}
}

func TestParseDocumentation(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		checkType       string // "enum", "enumvalue", "type", "field", "service", "method"
		expectedGeneral string
		expectedProto   string
		expectedGraphQL string
		expectedOpenAPI string
	}{
		{
			name: "enum with general documentation",
			input: `
/// User role enumeration
enum UserRole {
  ADMIN
}`,
			checkType:       "enum",
			expectedGeneral: "User role enumeration",
		},
		{
			name: "enum with language-specific documentation",
			input: `
/// General description
/// @proto Proto-specific description
/// @graphql GraphQL-specific description
enum UserRole {
  ADMIN
}`,
			checkType:       "enum",
			expectedGeneral: "General description",
			expectedProto:   "Proto-specific description",
			expectedGraphQL: "GraphQL-specific description",
		},
		{
			name: "enum value with documentation",
			input: `
enum UserRole {
  /// Administrator with full access
  ADMIN
  USER
}`,
			checkType:       "enumvalue",
			expectedGeneral: "Administrator with full access",
		},
		{
			name: "type with documentation",
			input: `
/// User entity
/// @openapi User schema for REST API
type User {
  id: string
}`,
			checkType:       "type",
			expectedGeneral: "User entity",
			expectedOpenAPI: "User schema for REST API",
		},
		{
			name: "field with documentation",
			input: `
type User {
  /// Unique identifier
  /// @proto User ID field
  id: string
}`,
			checkType:       "field",
			expectedGeneral: "Unique identifier",
			expectedProto:   "User ID field",
		},
		{
			name: "service with documentation",
			input: `
/// User management service
service UserService {
  rpc GetUser(Req) returns (Res)
}`,
			checkType:       "service",
			expectedGeneral: "User management service",
		},
		{
			name: "method with documentation",
			input: `
service UserService {
  /// Get a user by ID
  /// @graphql Query to fetch user
  rpc GetUser(Req) returns (Res)
}`,
			checkType:       "method",
			expectedGeneral: "Get a user by ID",
			expectedGraphQL: "Query to fetch user",
		},
		{
			name: "multiline documentation",
			input: `
/// First line
/// Second line
/// Third line
enum Status {
  ACTIVE
}`,
			checkType:       "enum",
			expectedGeneral: "First line\nSecond line\nThird line",
		},
		{
			name: "mixed language-specific documentation",
			input: `
/// General doc
/// @proto Proto line 1
/// @proto Proto line 2
/// @graphql GraphQL doc
type User {
  id: string
}`,
			checkType:       "type",
			expectedGeneral: "General doc",
			expectedProto:   "Proto line 1\nProto line 2",
			expectedGraphQL: "GraphQL doc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			var doc *ast.Documentation

			switch tt.checkType {
			case "enum":
				if len(schema.Enums) != 1 {
					t.Fatalf("Expected 1 enum, got %d", len(schema.Enums))
				}
				doc = schema.Enums[0].Doc
			case "enumvalue":
				if len(schema.Enums) != 1 || len(schema.Enums[0].Values) == 0 {
					t.Fatalf("Expected 1 enum with values")
				}
				doc = schema.Enums[0].Values[0].Doc
			case "type":
				if len(schema.Types) != 1 {
					t.Fatalf("Expected 1 type, got %d", len(schema.Types))
				}
				doc = schema.Types[0].Doc
			case "field":
				if len(schema.Types) != 1 || len(schema.Types[0].Fields) == 0 {
					t.Fatalf("Expected 1 type with fields")
				}
				doc = schema.Types[0].Fields[0].Doc
			case "service":
				if len(schema.Services) != 1 {
					t.Fatalf("Expected 1 service, got %d", len(schema.Services))
				}
				doc = schema.Services[0].Doc
			case "method":
				if len(schema.Services) != 1 || len(schema.Services[0].Methods) == 0 {
					t.Fatalf("Expected 1 service with methods")
				}
				doc = schema.Services[0].Methods[0].Doc
			}

			if doc == nil {
				t.Fatal("Expected documentation but got nil")
			}

			if doc.General != tt.expectedGeneral {
				t.Errorf("Expected general doc %q, got %q", tt.expectedGeneral, doc.General)
			}

			if tt.expectedProto != "" {
				if protoDoc := doc.GetDoc("proto"); protoDoc != tt.expectedProto {
					t.Errorf("Expected proto doc %q, got %q", tt.expectedProto, protoDoc)
				}
			}

			if tt.expectedGraphQL != "" {
				if graphqlDoc := doc.GetDoc("graphql"); graphqlDoc != tt.expectedGraphQL {
					t.Errorf("Expected graphql doc %q, got %q", tt.expectedGraphQL, graphqlDoc)
				}
			}

			if tt.expectedOpenAPI != "" {
				if openapiDoc := doc.GetDoc("openapi"); openapiDoc != tt.expectedOpenAPI {
					t.Errorf("Expected openapi doc %q, got %q", tt.expectedOpenAPI, openapiDoc)
				}
			}
		})
	}
}

func TestParseFieldExclusion(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		fieldName        string
		expectedExclude  []string
		expectedOnly     []string
		shouldIncludeMap map[string]bool
	}{
		{
			name: "exclude from single generator",
			input: `type User {
  internal: string @exclude(proto)
}`,
			fieldName:       "internal",
			expectedExclude: []string{"proto"},
			shouldIncludeMap: map[string]bool{
				"proto":   false,
				"graphql": true,
				"openapi": true,
			},
		},
		{
			name: "exclude from multiple generators",
			input: `type User {
  dbVersion: int32 @exclude(graphql,openapi)
}`,
			fieldName:       "dbVersion",
			expectedExclude: []string{"graphql", "openapi"},
			shouldIncludeMap: map[string]bool{
				"proto":   true,
				"graphql": false,
				"openapi": false,
			},
		},
		{
			name: "only for single generator",
			input: `type User {
  passwordHash: string @only(proto)
}`,
			fieldName:    "passwordHash",
			expectedOnly: []string{"proto"},
			shouldIncludeMap: map[string]bool{
				"proto":   true,
				"graphql": false,
				"openapi": false,
			},
		},
		{
			name: "only for multiple generators",
			input: `type User {
  publicField: string @only(graphql,openapi)
}`,
			fieldName:    "publicField",
			expectedOnly: []string{"graphql", "openapi"},
			shouldIncludeMap: map[string]bool{
				"proto":   false,
				"graphql": true,
				"openapi": true,
			},
		},
		{
			name: "exclude with other attributes",
			input: `type User {
  id: string @required @exclude(proto)
}`,
			fieldName:       "id",
			expectedExclude: []string{"proto"},
			shouldIncludeMap: map[string]bool{
				"proto":   false,
				"graphql": true,
			},
		},
		{
			name: "only with other attributes",
			input: `type User {
  score: int32 @default(0) @only(openapi)
}`,
			fieldName:    "score",
			expectedOnly: []string{"openapi"},
			shouldIncludeMap: map[string]bool{
				"proto":   false,
				"openapi": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			if len(schema.Types) != 1 || len(schema.Types[0].Fields) != 1 {
				t.Fatalf("Expected 1 type with 1 field")
			}

			field := schema.Types[0].Fields[0]

			if field.Name != tt.fieldName {
				t.Errorf("Expected field name %q, got %q", tt.fieldName, field.Name)
			}

			// Check ExcludeFrom
			if len(tt.expectedExclude) > 0 {
				if len(field.ExcludeFrom) != len(tt.expectedExclude) {
					t.Errorf("Expected ExcludeFrom length %d, got %d", len(tt.expectedExclude), len(field.ExcludeFrom))
				}
				for i, expected := range tt.expectedExclude {
					if i >= len(field.ExcludeFrom) || field.ExcludeFrom[i] != expected {
						t.Errorf("Expected ExcludeFrom[%d]=%q, got %q", i, expected, field.ExcludeFrom[i])
					}
				}
			}

			// Check OnlyFor
			if len(tt.expectedOnly) > 0 {
				if len(field.OnlyFor) != len(tt.expectedOnly) {
					t.Errorf("Expected OnlyFor length %d, got %d", len(tt.expectedOnly), len(field.OnlyFor))
				}
				for i, expected := range tt.expectedOnly {
					if i >= len(field.OnlyFor) || field.OnlyFor[i] != expected {
						t.Errorf("Expected OnlyFor[%d]=%q, got %q", i, expected, field.OnlyFor[i])
					}
				}
			}

			// Check ShouldIncludeInGenerator
			for generator, expected := range tt.shouldIncludeMap {
				result := field.ShouldIncludeInGenerator(generator)
				if result != expected {
					t.Errorf("ShouldIncludeInGenerator(%q) = %v, want %v", generator, result, expected)
				}
			}
		})
	}
}

func TestParseMethodAnnotations(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		methodName          string
		expectedHTTP        string
		expectedGraphQL     string
		expectedHTTPLower   string
		expectedGQLResolved string
	}{
		{
			name: "explicit HTTP POST",
			input: `service UserService {
  rpc CreateUser(Req) returns (Res) @http(POST)
}`,
			methodName:        "CreateUser",
			expectedHTTP:      "POST",
			expectedHTTPLower: "post",
		},
		{
			name: "explicit HTTP GET",
			input: `service UserService {
  rpc GetUser(Req) returns (Res) @http(GET)
}`,
			methodName:        "GetUser",
			expectedHTTP:      "GET",
			expectedHTTPLower: "get",
		},
		{
			name: "explicit HTTP DELETE",
			input: `service UserService {
  rpc DeleteUser(Req) returns (Res) @http(DELETE)
}`,
			methodName:        "DeleteUser",
			expectedHTTP:      "DELETE",
			expectedHTTPLower: "delete",
		},
		{
			name: "explicit HTTP PUT",
			input: `service UserService {
  rpc UpdateUser(Req) returns (Res) @http(PUT)
}`,
			methodName:        "UpdateUser",
			expectedHTTP:      "PUT",
			expectedHTTPLower: "put",
		},
		{
			name: "explicit HTTP PATCH",
			input: `service UserService {
  rpc PatchUser(Req) returns (Res) @http(PATCH)
}`,
			methodName:        "PatchUser",
			expectedHTTP:      "PATCH",
			expectedHTTPLower: "patch",
		},
		{
			name: "explicit GraphQL query",
			input: `service UserService {
  rpc CreateUser(Req) returns (Res) @graphql(query)
}`,
			methodName:          "CreateUser",
			expectedGraphQL:     "query",
			expectedGQLResolved: "query",
		},
		{
			name: "explicit GraphQL mutation",
			input: `service UserService {
  rpc GetUser(Req) returns (Res) @graphql(mutation)
}`,
			methodName:          "GetUser",
			expectedGraphQL:     "mutation",
			expectedGQLResolved: "mutation",
		},
		{
			name: "both HTTP and GraphQL annotations",
			input: `service UserService {
  rpc CreateUser(Req) returns (Res) @http(POST) @graphql(mutation)
}`,
			methodName:          "CreateUser",
			expectedHTTP:        "POST",
			expectedHTTPLower:   "post",
			expectedGraphQL:     "mutation",
			expectedGQLResolved: "mutation",
		},
		{
			name: "heuristic fallback for Get prefix",
			input: `service UserService {
  rpc GetUser(Req) returns (Res)
}`,
			methodName:          "GetUser",
			expectedHTTPLower:   "get",
			expectedGQLResolved: "query",
		},
		{
			name: "heuristic fallback for List prefix",
			input: `service UserService {
  rpc ListUsers(Req) returns (Res)
}`,
			methodName:          "ListUsers",
			expectedHTTPLower:   "get",
			expectedGQLResolved: "query",
		},
		{
			name: "heuristic fallback for other methods",
			input: `service UserService {
  rpc CreateUser(Req) returns (Res)
}`,
			methodName:          "CreateUser",
			expectedHTTPLower:   "post",
			expectedGQLResolved: "mutation",
		},
		{
			name: "override heuristic with explicit annotation",
			input: `service UserService {
  rpc GetUser(Req) returns (Res) @http(POST) @graphql(mutation)
}`,
			methodName:          "GetUser",
			expectedHTTP:        "POST",
			expectedHTTPLower:   "post",
			expectedGraphQL:     "mutation",
			expectedGQLResolved: "mutation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			if len(schema.Services) != 1 || len(schema.Services[0].Methods) != 1 {
				t.Fatalf("Expected 1 service with 1 method")
			}

			method := schema.Services[0].Methods[0]

			if method.Name != tt.methodName {
				t.Errorf("Expected method name %q, got %q", tt.methodName, method.Name)
			}

			// Check stored HTTPMethod value
			if tt.expectedHTTP != "" && method.HTTPMethod != tt.expectedHTTP {
				t.Errorf("Expected HTTPMethod %q, got %q", tt.expectedHTTP, method.HTTPMethod)
			}

			// Check stored GraphQLType value
			if tt.expectedGraphQL != "" && method.GraphQLType != tt.expectedGraphQL {
				t.Errorf("Expected GraphQLType %q, got %q", tt.expectedGraphQL, method.GraphQLType)
			}

			// Check GetHTTPMethod() with heuristics
			if tt.expectedHTTPLower != "" {
				result := method.GetHTTPMethod()
				if result != tt.expectedHTTPLower {
					t.Errorf("GetHTTPMethod() = %q, want %q", result, tt.expectedHTTPLower)
				}
			}

			// Check GetGraphQLType() with heuristics
			if tt.expectedGQLResolved != "" {
				result := method.GetGraphQLType()
				if result != tt.expectedGQLResolved {
					t.Errorf("GetGraphQLType() = %q, want %q", result, tt.expectedGQLResolved)
				}
			}
		})
	}
}

func TestParseMethodWithDocumentationAndAnnotations(t *testing.T) {
	input := `
service UserService {
  /// Create a new user in the system
  /// @proto CreateUser RPC method
  /// @graphql Mutation to create user
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) @http(POST) @graphql(mutation)
}`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Services) != 1 || len(schema.Services[0].Methods) != 1 {
		t.Fatalf("Expected 1 service with 1 method")
	}

	method := schema.Services[0].Methods[0]

	// Check documentation
	if method.Doc == nil {
		t.Fatal("Expected documentation but got nil")
	}

	expectedGeneral := "Create a new user in the system"
	if method.Doc.General != expectedGeneral {
		t.Errorf("Expected general doc %q, got %q", expectedGeneral, method.Doc.General)
	}

	expectedProto := "CreateUser RPC method"
	if protoDoc := method.Doc.GetDoc("proto"); protoDoc != expectedProto {
		t.Errorf("Expected proto doc %q, got %q", expectedProto, protoDoc)
	}

	expectedGraphQL := "Mutation to create user"
	if graphqlDoc := method.Doc.GetDoc("graphql"); graphqlDoc != expectedGraphQL {
		t.Errorf("Expected graphql doc %q, got %q", expectedGraphQL, graphqlDoc)
	}

	// Check annotations
	if method.HTTPMethod != "POST" {
		t.Errorf("Expected HTTPMethod POST, got %q", method.HTTPMethod)
	}

	if method.GraphQLType != "mutation" {
		t.Errorf("Expected GraphQLType mutation, got %q", method.GraphQLType)
	}
}

func TestParsePathTemplate(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedPath string
	}{
		{
			name: "simple path template",
			input: `service UserService {
  rpc GetUser(Req) returns (Res) @path("/users/{id}")
}`,
			expectedPath: "/users/{id}",
		},
		{
			name: "path template with multiple parameters",
			input: `service API {
  rpc GetPost(Req) returns (Res) @path("/users/{userId}/posts/{postId}")
}`,
			expectedPath: "/users/{userId}/posts/{postId}",
		},
		{
			name: "path template with HTTP method",
			input: `service UserService {
  rpc GetUser(Req) returns (Res) @http(GET) @path("/api/v1/users/{id}")
}`,
			expectedPath: "/api/v1/users/{id}",
		},
		{
			name: "path template with all annotations",
			input: `service UserService {
  rpc UpdateUser(Req) returns (Res) @http(PUT) @path("/users/{id}") @graphql(mutation)
}`,
			expectedPath: "/users/{id}",
		},
		{
			name: "path template without parameters",
			input: `service UserService {
  rpc ListUsers(Req) returns (Res) @path("/api/users")
}`,
			expectedPath: "/api/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			if len(schema.Services) != 1 || len(schema.Services[0].Methods) != 1 {
				t.Fatalf("Expected 1 service with 1 method")
			}

			method := schema.Services[0].Methods[0]

			if method.PathTemplate != tt.expectedPath {
				t.Errorf("Expected PathTemplate %q, got %q", tt.expectedPath, method.PathTemplate)
			}
		})
	}
}

func TestParseErrorCodes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCodes []string
	}{
		{
			name: "single error code",
			input: `
service UserService {
  rpc GetUser(Req) returns (Res) @errors(404)
}`,
			expectedCodes: []string{"404"},
		},
		{
			name: "multiple error codes",
			input: `
service UserService {
  rpc CreateUser(Req) returns (Res) @errors(400,404,409)
}`,
			expectedCodes: []string{"400", "404", "409"},
		},
		{
			name: "with other annotations",
			input: `
service UserService {
  rpc GetUser(Req) returns (Res) @http(GET) @path("/users/{id}") @errors(404,500)
}`,
			expectedCodes: []string{"404", "500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			if len(schema.Services) != 1 || len(schema.Services[0].Methods) != 1 {
				t.Fatalf("Expected 1 service with 1 method")
			}

			method := schema.Services[0].Methods[0]

			if len(method.ErrorCodes) != len(tt.expectedCodes) {
				t.Errorf("Expected %d error codes, got %d", len(tt.expectedCodes), len(method.ErrorCodes))
			}

			for i, code := range tt.expectedCodes {
				if i >= len(method.ErrorCodes) {
					t.Errorf("Missing error code at index %d", i)
					continue
				}
				if method.ErrorCodes[i] != code {
					t.Errorf("Expected error code %q at index %d, got %q", code, i, method.ErrorCodes[i])
				}
			}
		})
	}
}

func TestParseSuccessCodes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCodes []string
	}{
		{
			name: "single success code",
			input: `
service UserService {
  rpc CreateUser(Req) returns (Res) @success(201)
}`,
			expectedCodes: []string{"201"},
		},
		{
			name: "multiple success codes",
			input: `
service UserService {
  rpc CreateUser(Req) returns (Res) @success(201,202,204)
}`,
			expectedCodes: []string{"201", "202", "204"},
		},
		{
			name: "with other annotations",
			input: `
service UserService {
  rpc CreateUser(Req) returns (Res) @http(POST) @success(201) @errors(400,409)
}`,
			expectedCodes: []string{"201"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			if len(schema.Services) != 1 || len(schema.Services[0].Methods) != 1 {
				t.Fatalf("Expected 1 service with 1 method")
			}

			method := schema.Services[0].Methods[0]

			if len(method.SuccessCodes) != len(tt.expectedCodes) {
				t.Errorf("Expected %d success codes, got %d", len(tt.expectedCodes), len(method.SuccessCodes))
			}

			for i, code := range tt.expectedCodes {
				if i >= len(method.SuccessCodes) {
					t.Errorf("Missing success code at index %d", i)
					continue
				}
				if method.SuccessCodes[i] != code {
					t.Errorf("Expected success code %q at index %d, got %q", code, i, method.SuccessCodes[i])
				}
			}
		})
	}
}

func TestParseImport(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedImports []string
	}{
		{
			name: "single import",
			input: `import "common.typemux"

			type User {
				id: string
			}`,
			expectedImports: []string{"common.typemux"},
		},
		{
			name: "multiple imports",
			input: `import "common.typemux"
			import "types/user.typemux"
			import "types/order.typemux"

			type User {
				id: string
			}`,
			expectedImports: []string{"common.typemux", "types/user.typemux", "types/order.typemux"},
		},
		{
			name: "relative path import",
			input: `import "../shared/common.typemux"

			type User {
				id: string
			}`,
			expectedImports: []string{"../shared/common.typemux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser returned errors: %v", p.Errors())
			}

			if len(schema.Imports) != len(tt.expectedImports) {
				t.Errorf("Expected %d imports, got %d", len(tt.expectedImports), len(schema.Imports))
			}

			for i, expectedImport := range tt.expectedImports {
				if i >= len(schema.Imports) {
					t.Errorf("Missing import at index %d", i)
					continue
				}
				if schema.Imports[i] != expectedImport {
					t.Errorf("Expected import %q at index %d, got %q", expectedImport, i, schema.Imports[i])
				}
			}
		})
	}
}

func TestParser_ParseNamespace(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedNamespace string
	}{
		{
			name:              "simple namespace",
			input:             "namespace api",
			expectedNamespace: "api",
		},
		{
			name:              "dotted namespace",
			input:             "namespace com.example.api",
			expectedNamespace: "com.example.api",
		},
		{
			name:              "deeply dotted namespace",
			input:             "namespace com.company.product.api.v1",
			expectedNamespace: "com.company.product.api.v1",
		},
		{
			name:              "namespace with schema elements",
			input:             "namespace myapi\n\nenum Status { ACTIVE INACTIVE }",
			expectedNamespace: "myapi",
		},
		{
			name:              "no namespace uses default",
			input:             "enum Status { ACTIVE }",
			expectedNamespace: "api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %s", p.PrintErrors())
			}

			if schema.Namespace != tt.expectedNamespace {
				t.Errorf("Expected namespace %q, got %q", tt.expectedNamespace, schema.Namespace)
			}
		})
	}
}

func TestParseLeadingAnnotations(t *testing.T) {
	input := `
namespace com.example.api

@proto.name("UserV2")
@graphql.name("UserAccount")
@openapi.name("UserProfile")
type User {
    id: string = 1 @required
    @required
    username: string = 2
    email: string = 3 @required
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %s", p.PrintErrors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]

	// Check that leading annotations were parsed for the type
	if typ.Annotations == nil {
		t.Fatal("Expected annotations to be set")
	}

	if typ.Annotations.ProtoName != "UserV2" {
		t.Errorf("Expected ProtoName %q, got %q", "UserV2", typ.Annotations.ProtoName)
	}

	if typ.Annotations.GraphQLName != "UserAccount" {
		t.Errorf("Expected GraphQLName %q, got %q", "UserAccount", typ.Annotations.GraphQLName)
	}

	if typ.Annotations.OpenAPIName != "UserProfile" {
		t.Errorf("Expected OpenAPIName %q, got %q", "UserProfile", typ.Annotations.OpenAPIName)
	}

	// Check fields
	if len(typ.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(typ.Fields))
	}

	// Check first field - trailing annotation
	if !typ.Fields[0].Required {
		t.Error("Expected id field to be required")
	}

	// Check second field - leading @required annotation
	if typ.Fields[1].Name != "username" {
		t.Errorf("Expected field name 'username', got %q", typ.Fields[1].Name)
	}

	// Leading @required on fields works
	if _, ok := typ.Fields[1].Attributes["required"]; !ok {
		t.Error("Expected username field to have required attribute (from leading @required)")
	}

	if !typ.Fields[1].Required {
		t.Error("Expected username field Required flag to be true")
	}
}

func TestParseMixedLeadingAndTrailingAnnotations(t *testing.T) {
	input := `
namespace test

@proto.name("TypeProto")
type Example @graphql.name("TypeGraphQL") {
    id: string = 1
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %s", p.PrintErrors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]

	// Both leading and trailing annotations should be merged
	if typ.Annotations == nil {
		t.Fatal("Expected annotations to be set")
	}

	if typ.Annotations.ProtoName != "TypeProto" {
		t.Errorf("Expected ProtoName from leading annotation: %q, got %q", "TypeProto", typ.Annotations.ProtoName)
	}

	if typ.Annotations.GraphQLName != "TypeGraphQL" {
		t.Errorf("Expected GraphQLName from trailing annotation: %q, got %q", "TypeGraphQL", typ.Annotations.GraphQLName)
	}
}

func TestParseMultilineLeadingAnnotations(t *testing.T) {
	input := `
@proto.name("V2")
@graphql.name("GQL")
@openapi.name("OA")
type MultiAnnotated {
    field: string = 1
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %s", p.PrintErrors())
	}

	typ := schema.Types[0]

	if typ.Annotations.ProtoName != "V2" {
		t.Errorf("Expected ProtoName %q, got %q", "V2", typ.Annotations.ProtoName)
	}

	if typ.Annotations.GraphQLName != "GQL" {
		t.Errorf("Expected GraphQLName %q, got %q", "GQL", typ.Annotations.GraphQLName)
	}

	if typ.Annotations.OpenAPIName != "OA" {
		t.Errorf("Expected OpenAPIName %q, got %q", "OA", typ.Annotations.OpenAPIName)
	}
}

func TestParseOptionalField(t *testing.T) {
	input := `@typemux("1.0.0")
namespace test

type User {
	name: string? = 1
	age: int32 = 2
}`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser had errors: %v", p.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]
	if len(typ.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(typ.Fields))
	}

	// Test optional field
	nameField := typ.Fields[0]
	if nameField.Name != "name" {
		t.Errorf("Expected field name 'name', got '%s'", nameField.Name)
	}
	if !nameField.Type.Optional {
		t.Error("Expected name field to be optional")
	}
	if nameField.Type.Name != "string" {
		t.Errorf("Expected type 'string', got '%s'", nameField.Type.Name)
	}

	// Test non-optional field
	ageField := typ.Fields[1]
	if ageField.Name != "age" {
		t.Errorf("Expected field name 'age', got '%s'", ageField.Name)
	}
	if ageField.Type.Optional {
		t.Error("Expected age field to not be optional")
	}
}

func TestParseOptionalArrayField(t *testing.T) {
	input := `@typemux("1.0.0")
namespace test

type User {
	tags: []string? = 1
}`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser had errors: %v", p.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]
	if len(typ.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(typ.Fields))
	}

	field := typ.Fields[0]
	if !field.Type.Optional {
		t.Error("Expected field to be optional")
	}
	if !field.Type.IsArray {
		t.Error("Expected field to be an array")
	}
	if field.Type.Name != "string" {
		t.Errorf("Expected element type 'string', got '%s'", field.Type.Name)
	}
}

func TestParseUnion(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedName    string
		expectedOptions []string
		expectErrors    bool
	}{
		{
			name: "simple union",
			input: `union Result {
				Success
				Error
			}`,
			expectedName:    "Result",
			expectedOptions: []string{"Success", "Error"},
			expectErrors:    false,
		},
		{
			name: "union with multiple options",
			input: `union Message {
				TextMessage
				ImageMessage
				VideoMessage
				AudioMessage
			}`,
			expectedName:    "Message",
			expectedOptions: []string{"TextMessage", "ImageMessage", "VideoMessage", "AudioMessage"},
			expectErrors:    false,
		},
		{
			name: "union with documentation",
			input: `/// Result type for API calls
			union Result {
				Success
				Error
			}`,
			expectedName:    "Result",
			expectedOptions: []string{"Success", "Error"},
			expectErrors:    false,
		},
		{
			name: "union with leading annotations",
			input: `@proto.name("ResultProto")
			union Result {
				Success
				Error
			}`,
			expectedName:    "Result",
			expectedOptions: []string{"Success", "Error"},
			expectErrors:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.expectErrors {
				if len(p.Errors()) == 0 {
					t.Error("Expected errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
				return
			}

			if len(schema.Unions) != 1 {
				t.Fatalf("Expected 1 union, got %d", len(schema.Unions))
			}

			union := schema.Unions[0]
			if union.Name != tt.expectedName {
				t.Errorf("Expected union name %q, got %q", tt.expectedName, union.Name)
			}

			if len(union.Options) != len(tt.expectedOptions) {
				t.Fatalf("Expected %d options, got %d", len(tt.expectedOptions), len(union.Options))
			}

			for i, expectedOption := range tt.expectedOptions {
				if union.Options[i] != expectedOption {
					t.Errorf("Option %d: expected %q, got %q", i, expectedOption, union.Options[i])
				}
			}
		})
	}
}

func TestParseStreamingMethods(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		methodName   string
		expectInput  bool
		expectOutput bool
	}{
		{
			name: "server streaming",
			input: `service StreamService {
				rpc WatchEvents(Request) returns (stream Event)
			}`,
			methodName:   "WatchEvents",
			expectOutput: true,
		},
		{
			name: "client streaming",
			input: `service StreamService {
				rpc UploadData(stream Data) returns (Response)
			}`,
			methodName:  "UploadData",
			expectInput: true,
		},
		{
			name: "bidirectional streaming",
			input: `service StreamService {
				rpc Chat(stream Message) returns (stream Message)
			}`,
			methodName:   "Chat",
			expectInput:  true,
			expectOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			if len(schema.Services) != 1 || len(schema.Services[0].Methods) != 1 {
				t.Fatalf("Expected 1 service with 1 method")
			}

			method := schema.Services[0].Methods[0]

			if method.Name != tt.methodName {
				t.Errorf("Expected method name %q, got %q", tt.methodName, method.Name)
			}

			if method.InputStream != tt.expectInput {
				t.Errorf("Expected InputStream=%v, got %v", tt.expectInput, method.InputStream)
			}

			if method.OutputStream != tt.expectOutput {
				t.Errorf("Expected OutputStream=%v, got %v", tt.expectOutput, method.OutputStream)
			}
		})
	}
}

// TestParseDeprecation tests deprecation field annotations
// Note: @deprecated annotation is currently only supported on fields, not enum values or methods
func TestParseDeprecation(t *testing.T) {
	input := `
namespace test

type Product {
  id: string @required
  oldField: string @deprecated("Use newField instead", since="2.0.0", removed="3.0.0")
  newField: string
  legacyFlag: bool @deprecated
}
`
	l := lexer.New(input)
	parser := New(l)
	schema := parser.Parse()

	if len(parser.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", parser.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	product := schema.Types[0]
	if len(product.Fields) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(product.Fields))
	}

	// Check id field is not deprecated
	if product.Fields[0].Deprecated != nil {
		t.Error("Expected id field to not be deprecated")
	}

	// Check oldField has full deprecation info
	oldField := product.Fields[1]
	if oldField.Deprecated == nil {
		t.Fatal("Expected oldField to be deprecated")
	}
	if oldField.Deprecated.Reason != "Use newField instead" {
		t.Errorf("Expected reason 'Use newField instead', got %q", oldField.Deprecated.Reason)
	}
	if oldField.Deprecated.Since != "2.0.0" {
		t.Errorf("Expected since '2.0.0', got %q", oldField.Deprecated.Since)
	}
	if oldField.Deprecated.Removed != "3.0.0" {
		t.Errorf("Expected removed '3.0.0', got %q", oldField.Deprecated.Removed)
	}

	// Check newField is not deprecated
	if product.Fields[2].Deprecated != nil {
		t.Error("Expected newField to not be deprecated")
	}

	// Check legacyFlag has deprecation marker without details
	legacyFlag := product.Fields[3]
	if legacyFlag.Deprecated == nil {
		t.Fatal("Expected legacyFlag to be deprecated")
	}
	if legacyFlag.Deprecated.Reason != "" {
		t.Errorf("Expected empty reason, got %q", legacyFlag.Deprecated.Reason)
	}
}

func TestParseEdgeCaseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed enum block",
			input: "enum Status { ACTIVE",
		},
		{
			name:  "unclosed type block",
			input: "type User { id: string",
		},
		{
			name:  "unclosed service block",
			input: "service API { rpc Get(Req) returns (Res)",
		},
		{
			name:  "unclosed union block",
			input: "union Result { Success",
		},
		{
			name:  "missing return type",
			input: "service API { rpc Get(Req) }",
		},
		{
			name:  "missing input type",
			input: "service API { rpc Get() returns (Res) }",
		},
		{
			name:  "invalid field separator",
			input: "type User { id string }",
		},
		{
			name:  "missing map value type",
			input: "type Config { data: map<string> }",
		},
		{
			name:  "invalid annotation syntax",
			input: "type User { id: string @invalid",
		},
		{
			name:  "missing annotation parenthesis",
			input: "type User { id: string @default(true }",
		},
		{
			name:  "invalid namespace",
			input: "namespace",
		},
		{
			name:  "invalid import",
			input: "import",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.Parse()

			if len(p.Errors()) == 0 {
				t.Error("Expected parser errors but got none")
			}
		})
	}
}

func TestParseComplexMapTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		fieldName   string
		expectKey   string
		expectValue string
	}{
		{
			name: "string to int map",
			input: `type Config {
				counts: map<string, int32>
			}`,
			fieldName:   "counts",
			expectKey:   "string",
			expectValue: "int32",
		},
		{
			name: "string to custom type map",
			input: `type UserMap {
				users: map<string, User>
			}`,
			fieldName:   "users",
			expectKey:   "string",
			expectValue: "User",
		},
		{
			name: "int to string map",
			input: `type Lookup {
				ids: map<int64, string>
			}`,
			fieldName:   "ids",
			expectKey:   "int64",
			expectValue: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if len(p.Errors()) > 0 {
				t.Fatalf("Unexpected errors: %s", p.PrintErrors())
			}

			if len(schema.Types) != 1 || len(schema.Types[0].Fields) != 1 {
				t.Fatalf("Expected 1 type with 1 field")
			}

			field := schema.Types[0].Fields[0]

			if field.Name != tt.fieldName {
				t.Errorf("Expected field name %q, got %q", tt.fieldName, field.Name)
			}

			if !field.Type.IsMap {
				t.Error("Expected field to be a map")
			}

			if field.Type.MapKey != tt.expectKey {
				t.Errorf("Expected map key %q, got %q", tt.expectKey, field.Type.MapKey)
			}

			if field.Type.MapValue != tt.expectValue {
				t.Errorf("Expected map value %q, got %q", tt.expectValue, field.Type.MapValue)
			}
		})
	}
}

func TestParseNestedComplexTypes(t *testing.T) {
	input := `
type Address {
	street: string @required
	city: string @required
}

type User {
	id: string @required
	addresses: []Address
	metadata: map<string, string>
}

type Company {
	owner: User @required
	employees: []User
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Types) != 3 {
		t.Fatalf("Expected 3 types, got %d", len(schema.Types))
	}

	// Find User type
	var userType *ast.Type
	for _, typ := range schema.Types {
		if typ.Name == "User" {
			userType = typ
			break
		}
	}

	if userType == nil {
		t.Fatal("User type not found")
	}

	// Check addresses field (array of custom type)
	var addressesField *ast.Field
	for _, field := range userType.Fields {
		if field.Name == "addresses" {
			addressesField = field
			break
		}
	}

	if addressesField == nil {
		t.Fatal("addresses field not found")
	}

	if !addressesField.Type.IsArray {
		t.Error("Expected addresses to be an array")
	}

	if addressesField.Type.Name != "Address" {
		t.Errorf("Expected array element type Address, got %q", addressesField.Type.Name)
	}

	if addressesField.Type.IsBuiltin {
		t.Error("Expected Address to not be a builtin type")
	}

	// Find Company type
	var companyType *ast.Type
	for _, typ := range schema.Types {
		if typ.Name == "Company" {
			companyType = typ
			break
		}
	}

	if companyType == nil {
		t.Fatal("Company type not found")
	}

	// Check that Company has User fields
	if companyType.Fields[0].Type.Name != "User" {
		t.Errorf("Expected owner to be User type, got %q", companyType.Fields[0].Type.Name)
	}
}

func TestParseFieldNumberEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectErrors bool
		fieldNumbers []int
	}{
		{
			name: "sequential numbering",
			input: `type User {
				id: string = 1
				name: string = 2
				age: int32 = 3
			}`,
			expectErrors: false,
			fieldNumbers: []int{1, 2, 3},
		},
		{
			name: "sparse numbering",
			input: `type User {
				id: string = 1
				name: string = 10
				age: int32 = 100
			}`,
			expectErrors: false,
			fieldNumbers: []int{1, 10, 100},
		},
		{
			name: "large field numbers",
			input: `type User {
				id: string = 536870911
			}`,
			expectErrors: false,
			fieldNumbers: []int{536870911},
		},
		{
			name: "mixed with and without numbers",
			input: `type User {
				id: string = 1
				name: string
				age: int32 = 3
			}`,
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.expectErrors {
				if len(p.Errors()) == 0 {
					t.Error("Expected errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
				return
			}

			if len(schema.Types) != 1 {
				t.Fatalf("Expected 1 type, got %d", len(schema.Types))
			}

			if len(tt.fieldNumbers) > 0 {
				typ := schema.Types[0]
				for i, expectedNum := range tt.fieldNumbers {
					if i >= len(typ.Fields) {
						t.Fatalf("Not enough fields parsed")
					}
					if typ.Fields[i].Number != expectedNum {
						t.Errorf("Field %d: expected number %d, got %d", i, expectedNum, typ.Fields[i].Number)
					}
				}
			}
		})
	}
}

func TestParseCombinedAnnotations(t *testing.T) {
	input := `
type User {
	id: string = 1 @required @exclude(graphql)
	email: string = 2 @required @only(proto,openapi)
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Types) != 1 || len(schema.Types[0].Fields) != 2 {
		t.Fatalf("Expected 1 type with 2 fields")
	}

	// Check first field has all annotations
	idField := schema.Types[0].Fields[0]
	if !idField.Required {
		t.Error("Expected id to be required")
	}
	if len(idField.ExcludeFrom) == 0 || idField.ExcludeFrom[0] != "graphql" {
		t.Error("Expected ExcludeFrom to contain graphql")
	}

	// Check second field
	emailField := schema.Types[0].Fields[1]
	if !emailField.Required {
		t.Error("Expected email to be required")
	}
	if len(emailField.OnlyFor) != 2 {
		t.Errorf("Expected 2 OnlyFor entries, got %d", len(emailField.OnlyFor))
	}
}

func TestParseEmptySchema(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "completely empty",
			input: "",
		},
		{
			name:  "only whitespace",
			input: "   \n\n\t\t  \n  ",
		},
		{
			name:  "only comments",
			input: "// comment\n// another comment",
		},
		{
			name:  "only doc comments",
			input: "/// documentation\n/// more docs",
		},
		{
			name:  "only namespace",
			input: "namespace api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			// Should not error on empty input
			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors on empty input: %s", p.PrintErrors())
			}

			// But should have empty structures
			if len(schema.Types) != 0 {
				t.Errorf("Expected 0 types in empty schema, got %d", len(schema.Types))
			}
			if len(schema.Enums) != 0 {
				t.Errorf("Expected 0 enums in empty schema, got %d", len(schema.Enums))
			}
			if len(schema.Services) != 0 {
				t.Errorf("Expected 0 services in empty schema, got %d", len(schema.Services))
			}
		})
	}
}

func TestParseComplexAnnotations(t *testing.T) {
	input := `
@proto.package("com.example.v1")
@graphql.schema_directive("@cacheControl(maxAge: 3600)")
namespace com.example.api

@proto.name("UserV2")
@graphql.name("UserAccount")
@openapi.name("UserProfile")
type User {
	id: string = 1 @required @proto.name("user_id") @graphql.name("userId")
	name: string? = 2 @proto.name("full_name")
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	// Check namespace annotations
	if schema.Namespace != "com.example.api" {
		t.Errorf("Expected namespace 'com.example.api', got %q", schema.Namespace)
	}

	// Check type annotations
	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]
	if typ.Annotations.ProtoName != "UserV2" {
		t.Errorf("Expected ProtoName 'UserV2', got %q", typ.Annotations.ProtoName)
	}

	// Check field annotations
	if len(typ.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(typ.Fields))
	}

	idField := typ.Fields[0]
	if idField.Annotations.ProtoName != "user_id" {
		t.Errorf("Expected field ProtoName 'user_id', got %q", idField.Annotations.ProtoName)
	}
	if idField.Annotations.GraphQLName != "userId" {
		t.Errorf("Expected field GraphQLName 'userId', got %q", idField.Annotations.GraphQLName)
	}
}

func TestParseTypeMUXVersion(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectVersion string
		expectError   bool
	}{
		{
			name: "version at start",
			input: `@typemux("1.0.0")
namespace api

type User {
	id: string
}`,
			expectVersion: "1.0.0",
			expectError:   false,
		},
		{
			name: "different version",
			input: `@typemux("2.5.1")
namespace api`,
			expectVersion: "2.5.1",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			schema := p.Parse()

			if tt.expectError {
				if len(p.Errors()) == 0 {
					t.Error("Expected errors but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("Unexpected errors: %s", p.PrintErrors())
			}

			if schema.TypeMUXVersion != tt.expectVersion {
				t.Errorf("Expected version %q, got %q", tt.expectVersion, schema.TypeMUXVersion)
			}
		})
	}
}

func TestParseQualifiedTypeReferences(t *testing.T) {
	input := `
namespace api

type User {
	id: string = 1
	address: common.Address = 2
	config: system.config.Settings = 3
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	fields := schema.Types[0].Fields
	if len(fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(fields))
	}

	// Check qualified type names are parsed correctly
	if fields[1].Type.Name != "common.Address" {
		t.Errorf("Expected type 'common.Address', got %q", fields[1].Type.Name)
	}

	if fields[2].Type.Name != "system.config.Settings" {
		t.Errorf("Expected type 'system.config.Settings', got %q", fields[2].Type.Name)
	}
}

func TestParseArrayOfArrays(t *testing.T) {
	input := `
namespace test

type Matrix {
  simple_array: []string
  nested_array: [][]int32
  triple_nested: [][][]bool
  optional_nested: [][]string?
}
`
	l := lexer.New(input)
	parser := New(l)
	schema := parser.Parse()

	if len(parser.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", parser.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	matrix := schema.Types[0]
	if len(matrix.Fields) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(matrix.Fields))
	}

	// Check simple_array
	if !matrix.Fields[0].Type.IsArray {
		t.Error("Expected simple_array to be an array")
	}
	if matrix.Fields[0].Type.Name != "string" {
		t.Errorf("Expected simple_array element type 'string', got %q", matrix.Fields[0].Type.Name)
	}

	// Check nested_array ([][]int32)
	if !matrix.Fields[1].Type.IsArray {
		t.Error("Expected nested_array to be an array")
	}
	if matrix.Fields[1].Type.Name != "[]int32" {
		t.Errorf("Expected nested_array Name '[]int32', got %q", matrix.Fields[1].Type.Name)
	}

	// Check triple_nested ([][][]bool)
	if !matrix.Fields[2].Type.IsArray {
		t.Error("Expected triple_nested to be an array")
	}
	if matrix.Fields[2].Type.Name != "[][]bool" {
		t.Errorf("Expected triple_nested Name '[][]bool', got %q", matrix.Fields[2].Type.Name)
	}

	// Check optional_nested is optional
	if !matrix.Fields[3].Type.Optional {
		t.Error("Expected optional_nested to be optional")
	}
	if !matrix.Fields[3].Type.IsArray {
		t.Error("Expected optional_nested to be an array")
	}
}

func TestParseOptionalMapField(t *testing.T) {
	input := `
namespace test

type Config {
  required_map: map<string, string>
  optional_map: map<string, int32>?
  optional_string: string?
  optional_array: []string?
}
`
	l := lexer.New(input)
	parser := New(l)
	schema := parser.Parse()

	if len(parser.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", parser.Errors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	config := schema.Types[0]
	if len(config.Fields) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(config.Fields))
	}

	// Check required_map is not optional
	if config.Fields[0].Type.Optional {
		t.Error("Expected required_map to not be optional")
	}
	if !config.Fields[0].Type.IsMap {
		t.Error("Expected required_map to be a map")
	}

	// Check optional_map is optional
	if !config.Fields[1].Type.Optional {
		t.Error("Expected optional_map to be optional")
	}
	if !config.Fields[1].Type.IsMap {
		t.Error("Expected optional_map to be a map")
	}

	// Check optional_string is optional
	if !config.Fields[2].Type.Optional {
		t.Error("Expected optional_string to be optional")
	}

	// Check optional_array is optional
	if !config.Fields[3].Type.Optional {
		t.Error("Expected optional_array to be optional")
	}
	if !config.Fields[3].Type.IsArray {
		t.Error("Expected optional_array to be an array")
	}
}

func TestParseMultipleImportsAndNamespaces(t *testing.T) {
	input := `
import "common.typemux"
import "types.typemux"

namespace com.example.api

type User {
	id: string
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(schema.Imports))
	}

	if schema.Namespace != "com.example.api" {
		t.Errorf("Expected namespace 'com.example.api', got %q", schema.Namespace)
	}
}

func TestParseMethodWithStatusCodes(t *testing.T) {
	input := `
service API {
	rpc CreateUser(Req) returns (Res) @success(201) @errors(400,409,500)
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Services) != 1 || len(schema.Services[0].Methods) != 1 {
		t.Fatalf("Expected 1 service with 1 method")
	}

	method := schema.Services[0].Methods[0]

	if len(method.SuccessCodes) != 1 || method.SuccessCodes[0] != "201" {
		t.Errorf("Expected success code '201', got %v", method.SuccessCodes)
	}

	if len(method.ErrorCodes) != 3 {
		t.Errorf("Expected 3 error codes, got %d", len(method.ErrorCodes))
	}
}

func TestParseUnionUsedInType(t *testing.T) {
	input := `
union Result {
	Success
	Error
}

type Response {
	data: Result = 1
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Unions) != 1 {
		t.Fatalf("Expected 1 union, got %d", len(schema.Unions))
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	field := schema.Types[0].Fields[0]
	if field.Type.Name != "Result" {
		t.Errorf("Expected field type 'Result', got %q", field.Type.Name)
	}
}

func TestParseEnumWithLargeNumbers(t *testing.T) {
	input := `
enum Status {
	UNKNOWN = 0
	ACTIVE = 1000
	INACTIVE = 2000
	DEPRECATED = 9999
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Enums) != 1 {
		t.Fatalf("Expected 1 enum, got %d", len(schema.Enums))
	}

	enum := schema.Enums[0]
	expectedNumbers := []int{0, 1000, 2000, 9999}

	for i, expected := range expectedNumbers {
		if enum.Values[i].Number != expected {
			t.Errorf("Value %d: expected number %d, got %d", i, expected, enum.Values[i].Number)
		}
	}
}

func TestParseServiceWithMultipleHTTPMethods(t *testing.T) {
	input := `
service UserService {
	rpc GetUser(GetReq) returns (GetRes) @http(GET) @path("/users/{id}")
	rpc CreateUser(CreateReq) returns (CreateRes) @http(POST) @path("/users")
	rpc UpdateUser(UpdateReq) returns (UpdateRes) @http(PUT) @path("/users/{id}")
	rpc DeleteUser(DeleteReq) returns (DeleteRes) @http(DELETE) @path("/users/{id}")
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(schema.Services))
	}

	methods := schema.Services[0].Methods
	if len(methods) != 4 {
		t.Fatalf("Expected 4 methods, got %d", len(methods))
	}

	expectedMethods := []struct {
		name string
		http string
		path string
	}{
		{"GetUser", "GET", "/users/{id}"},
		{"CreateUser", "POST", "/users"},
		{"UpdateUser", "PUT", "/users/{id}"},
		{"DeleteUser", "DELETE", "/users/{id}"},
	}

	for i, expected := range expectedMethods {
		if methods[i].Name != expected.name {
			t.Errorf("Method %d: expected name %q, got %q", i, expected.name, methods[i].Name)
		}
		if methods[i].HTTPMethod != expected.http {
			t.Errorf("Method %d: expected HTTP %q, got %q", i, expected.http, methods[i].HTTPMethod)
		}
		if methods[i].PathTemplate != expected.path {
			t.Errorf("Method %d: expected path %q, got %q", i, expected.path, methods[i].PathTemplate)
		}
	}
}

func TestParseTypeWithAllFieldTypes(t *testing.T) {
	input := `
type ComplexType {
	simpleString: string = 1
	optionalString: string? = 2
	stringArray: []string = 3
	stringMap: map<string, string> = 4
	customType: Address = 5
	customArray: []Address = 6
	qualifiedType: common.Config = 7
}
`

	l := lexer.New(input)
	p := New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("Unexpected errors: %s", p.PrintErrors())
	}

	if len(schema.Types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(schema.Types))
	}

	fields := schema.Types[0].Fields
	if len(fields) != 7 {
		t.Fatalf("Expected 7 fields, got %d", len(fields))
	}

	// Test each field type
	tests := []struct {
		idx      int
		name     string
		optional bool
		isArray  bool
		isMap    bool
	}{
		{0, "simpleString", false, false, false},
		{1, "optionalString", true, false, false},
		{2, "stringArray", false, true, false},
		{3, "stringMap", false, false, true},
		{4, "customType", false, false, false},
		{5, "customArray", false, true, false},
		{6, "qualifiedType", false, false, false},
	}

	for _, tt := range tests {
		field := fields[tt.idx]
		if field.Name != tt.name {
			t.Errorf("Field %d: expected name %q, got %q", tt.idx, tt.name, field.Name)
		}
		if field.Type.Optional != tt.optional {
			t.Errorf("Field %q: expected optional=%v, got %v", tt.name, tt.optional, field.Type.Optional)
		}
		if field.Type.IsArray != tt.isArray {
			t.Errorf("Field %q: expected isArray=%v, got %v", tt.name, tt.isArray, field.Type.IsArray)
		}
		if field.Type.IsMap != tt.isMap {
			t.Errorf("Field %q: expected isMap=%v, got %v", tt.name, tt.isMap, field.Type.IsMap)
		}
	}
}
