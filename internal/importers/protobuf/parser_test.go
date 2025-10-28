package protobuf

import (
	"testing"
)

func TestParseSyntax(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "proto3 syntax",
			input:    `syntax = "proto3";`,
			expected: "proto3",
		},
		{
			name:     "proto2 syntax",
			input:    `syntax = "proto2";`,
			expected: "proto2",
		},
		{
			name:     "default proto3",
			input:    `invalid syntax`,
			expected: "proto3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			result := p.parseSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParsePackage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple package",
			input:    "package example;",
			expected: "example",
		},
		{
			name:     "nested package",
			input:    "package com.example.api;",
			expected: "com.example.api",
		},
		{
			name:     "empty package",
			input:    "invalid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			result := p.parsePackage(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseImport(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple import",
			input:    `import "common/types.proto";`,
			expected: "common/types.proto",
		},
		{
			name:     "google import",
			input:    `import "google/protobuf/timestamp.proto";`,
			expected: "google/protobuf/timestamp.proto",
		},
		{
			name:     "empty import",
			input:    "invalid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			result := p.parseImport(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseOption(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedKey   string
		expectedValue string
	}{
		{
			name:          "go_package option",
			input:         `option go_package = "github.com/example/proto";`,
			expectedKey:   "go_package",
			expectedValue: "github.com/example/proto",
		},
		{
			name:          "java_package option",
			input:         `option java_package = "com.example";`,
			expectedKey:   "java_package",
			expectedValue: "com.example",
		},
		{
			name:          "invalid option",
			input:         "invalid",
			expectedKey:   "",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			key, value := p.parseOption(tt.input)
			if key != tt.expectedKey {
				t.Errorf("expected key %q, got %q", tt.expectedKey, key)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestParseField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ProtoField
		wantErr  bool
	}{
		{
			name:  "simple field",
			input: "string name = 1;",
			expected: &ProtoField{
				Type:   "string",
				Name:   "name",
				Number: 1,
			},
			wantErr: false,
		},
		{
			name:  "repeated field",
			input: "repeated string tags = 2;",
			expected: &ProtoField{
				Type:     "string",
				Name:     "tags",
				Number:   2,
				Repeated: true,
			},
			wantErr: false,
		},
		{
			name:  "optional field",
			input: "optional int32 age = 3;",
			expected: &ProtoField{
				Type:     "int32",
				Name:     "age",
				Number:   3,
				Optional: true,
			},
			wantErr: false,
		},
		{
			name:  "map field",
			input: "map<string, int32> scores = 4;",
			expected: &ProtoField{
				Type:   "map<string, int32>",
				Name:   "scores",
				Number: 4,
			},
			wantErr: false,
		},
		{
			name:  "deprecated field",
			input: "string old_field = 5 [deprecated = true];",
			expected: &ProtoField{
				Type:       "string",
				Name:       "old_field",
				Number:     5,
				Deprecated: true,
			},
			wantErr: false,
		},
		{
			name:    "invalid field",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			result, err := p.parseField(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("expected type %q, got %q", tt.expected.Type, result.Type)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("expected name %q, got %q", tt.expected.Name, result.Name)
			}
			if result.Number != tt.expected.Number {
				t.Errorf("expected number %d, got %d", tt.expected.Number, result.Number)
			}
			if result.Repeated != tt.expected.Repeated {
				t.Errorf("expected repeated %v, got %v", tt.expected.Repeated, result.Repeated)
			}
			if result.Optional != tt.expected.Optional {
				t.Errorf("expected optional %v, got %v", tt.expected.Optional, result.Optional)
			}
			if result.Deprecated != tt.expected.Deprecated {
				t.Errorf("expected deprecated %v, got %v", tt.expected.Deprecated, result.Deprecated)
			}
		})
	}
}

func TestParseEnum(t *testing.T) {
	input := `syntax = "proto3";

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}`

	p := NewParser(input)
	schema, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(schema.Enums))
	}

	enum := schema.Enums[0]
	if enum.Name != "Status" {
		t.Errorf("expected enum name %q, got %q", "Status", enum.Name)
	}

	if len(enum.Values) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(enum.Values))
	}

	expectedValues := []struct {
		name   string
		number int
	}{
		{"STATUS_UNSPECIFIED", 0},
		{"STATUS_ACTIVE", 1},
		{"STATUS_INACTIVE", 2},
	}

	for i, expected := range expectedValues {
		if enum.Values[i].Name != expected.name {
			t.Errorf("expected value name %q, got %q", expected.name, enum.Values[i].Name)
		}
		if enum.Values[i].Number != expected.number {
			t.Errorf("expected value number %d, got %d", expected.number, enum.Values[i].Number)
		}
	}
}

func TestParseMessage(t *testing.T) {
	input := `syntax = "proto3";

message User {
  string id = 1;
  string name = 2;
  int32 age = 3;
  repeated string tags = 4;
}`

	p := NewParser(input)
	schema, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(schema.Messages))
	}

	msg := schema.Messages[0]
	if msg.Name != "User" {
		t.Errorf("expected message name %q, got %q", "User", msg.Name)
	}

	if len(msg.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(msg.Fields))
	}

	expectedFields := []struct {
		name     string
		typ      string
		number   int
		repeated bool
	}{
		{"id", "string", 1, false},
		{"name", "string", 2, false},
		{"age", "int32", 3, false},
		{"tags", "string", 4, true},
	}

	for i, expected := range expectedFields {
		if msg.Fields[i].Name != expected.name {
			t.Errorf("expected field name %q, got %q", expected.name, msg.Fields[i].Name)
		}
		if msg.Fields[i].Type != expected.typ {
			t.Errorf("expected field type %q, got %q", expected.typ, msg.Fields[i].Type)
		}
		if msg.Fields[i].Number != expected.number {
			t.Errorf("expected field number %d, got %d", expected.number, msg.Fields[i].Number)
		}
		if msg.Fields[i].Repeated != expected.repeated {
			t.Errorf("expected field repeated %v, got %v", expected.repeated, msg.Fields[i].Repeated)
		}
	}
}

func TestParseNestedMessage(t *testing.T) {
	input := `syntax = "proto3";

message Outer {
  string id = 1;

  message Inner {
    string value = 1;
  }

  Inner inner = 2;
}`

	p := NewParser(input)
	schema, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(schema.Messages))
	}

	outer := schema.Messages[0]
	if outer.Name != "Outer" {
		t.Errorf("expected message name %q, got %q", "Outer", outer.Name)
	}

	if len(outer.Messages) != 1 {
		t.Fatalf("expected 1 nested message, got %d", len(outer.Messages))
	}

	inner := outer.Messages[0]
	if inner.Name != "Inner" {
		t.Errorf("expected nested message name %q, got %q", "Inner", inner.Name)
	}

	if len(inner.Fields) != 1 {
		t.Fatalf("expected 1 field in nested message, got %d", len(inner.Fields))
	}
}

func TestParseService(t *testing.T) {
	input := `syntax = "proto3";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (stream ListUsersResponse);
  rpc UpdateUser(stream UpdateUserRequest) returns (UpdateUserResponse);
  rpc StreamUsers(stream StreamUsersRequest) returns (stream StreamUsersResponse);
}`

	p := NewParser(input)
	schema, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(schema.Services))
	}

	service := schema.Services[0]
	if service.Name != "UserService" {
		t.Errorf("expected service name %q, got %q", "UserService", service.Name)
	}

	if len(service.Methods) != 4 {
		t.Fatalf("expected 4 methods, got %d", len(service.Methods))
	}

	expectedMethods := []struct {
		name         string
		inputType    string
		outputType   string
		clientStream bool
		serverStream bool
	}{
		{"GetUser", "GetUserRequest", "GetUserResponse", false, false},
		{"ListUsers", "ListUsersRequest", "ListUsersResponse", false, true},
		{"UpdateUser", "UpdateUserRequest", "UpdateUserResponse", true, false},
		{"StreamUsers", "StreamUsersRequest", "StreamUsersResponse", true, true},
	}

	for i, expected := range expectedMethods {
		method := service.Methods[i]
		if method.Name != expected.name {
			t.Errorf("expected method name %q, got %q", expected.name, method.Name)
		}
		if method.InputType != expected.inputType {
			t.Errorf("expected input type %q, got %q", expected.inputType, method.InputType)
		}
		if method.OutputType != expected.outputType {
			t.Errorf("expected output type %q, got %q", expected.outputType, method.OutputType)
		}
		if method.ClientStream != expected.clientStream {
			t.Errorf("expected client stream %v, got %v", expected.clientStream, method.ClientStream)
		}
		if method.ServerStream != expected.serverStream {
			t.Errorf("expected server stream %v, got %v", expected.serverStream, method.ServerStream)
		}
	}
}

func TestParseOneOf(t *testing.T) {
	input := `syntax = "proto3";

message Payment {
  oneof payment_method {
    string credit_card = 1;
    string paypal = 2;
    string bank_account = 3;
  }
}`

	p := NewParser(input)
	schema, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(schema.Messages))
	}

	msg := schema.Messages[0]
	if len(msg.OneOfs) != 1 {
		t.Fatalf("expected 1 oneof, got %d", len(msg.OneOfs))
	}

	oneof := msg.OneOfs[0]
	if oneof.Name != "payment_method" {
		t.Errorf("expected oneof name %q, got %q", "payment_method", oneof.Name)
	}

	if len(oneof.Fields) != 3 {
		t.Fatalf("expected 3 fields in oneof, got %d", len(oneof.Fields))
	}
}

func TestParseReserved(t *testing.T) {
	input := `syntax = "proto3";

message Reserved {
  reserved 2, 3, 4;
  reserved "old_field", "deprecated_field";
  string name = 1;
}`

	p := NewParser(input)
	schema, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(schema.Messages))
	}

	msg := schema.Messages[0]
	if len(msg.Reserved) == 0 {
		t.Error("expected reserved fields to be parsed")
	}
}

func TestParseCompleteSchema(t *testing.T) {
	input := `syntax = "proto3";

package example;

option go_package = "github.com/example/proto";

import "google/protobuf/timestamp.proto";

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}

message User {
  string id = 1;
  string name = 2;
  Status status = 3;
  repeated string tags = 4;
  map<string, string> metadata = 5;
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  User user = 1;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}`

	p := NewParser(input)
	schema, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Syntax != "proto3" {
		t.Errorf("expected syntax %q, got %q", "proto3", schema.Syntax)
	}

	if schema.Package != "example" {
		t.Errorf("expected package %q, got %q", "example", schema.Package)
	}

	if len(schema.Options) == 0 {
		t.Error("expected options to be parsed")
	}

	if schema.Options["go_package"] != "github.com/example/proto" {
		t.Errorf("expected go_package option %q, got %q", "github.com/example/proto", schema.Options["go_package"])
	}

	if len(schema.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(schema.Imports))
	}

	if len(schema.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(schema.Enums))
	}

	if len(schema.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(schema.Messages))
	}

	if len(schema.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(schema.Services))
	}
}

func TestIsWellKnownType(t *testing.T) {
	p := NewParser("")

	tests := []struct {
		input    string
		expected bool
	}{
		{"google/protobuf/timestamp.proto", true},
		{"google/protobuf/duration.proto", true},
		{"google/api/annotations.proto", true},
		{"common/types.proto", false},
		{"user.proto", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := p.isWellKnownType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, result)
			}
		})
	}
}
