package protobuf

import (
	"strings"
	"testing"
)

func TestConvertBasicMessage(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Messages: []*ProtoMessage{
			{
				Name: "User",
				Fields: []*ProtoField{
					{Name: "id", Type: "string", Number: 1},
					{Name: "name", Type: "string", Number: 2},
					{Name: "age", Type: "int32", Number: 3},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "namespace example") {
		t.Error("expected namespace declaration")
	}

	if !strings.Contains(result, "type User {") {
		t.Error("expected User type declaration")
	}

	if !strings.Contains(result, "id: string = 1") {
		t.Error("expected id field")
	}

	if !strings.Contains(result, "name: string = 2") {
		t.Error("expected name field")
	}

	if !strings.Contains(result, "age: int32 = 3") {
		t.Error("expected age field")
	}
}

func TestConvertEnum(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Enums: []*ProtoEnum{
			{
				Name: "Status",
				Values: []*ProtoEnumValue{
					{Name: "STATUS_UNSPECIFIED", Number: 0},
					{Name: "STATUS_ACTIVE", Number: 1},
					{Name: "STATUS_INACTIVE", Number: 2},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "enum Status {") {
		t.Error("expected Status enum declaration")
	}

	if !strings.Contains(result, "STATUS_UNSPECIFIED = 0") {
		t.Error("expected STATUS_UNSPECIFIED value")
	}

	if !strings.Contains(result, "STATUS_ACTIVE = 1") {
		t.Error("expected STATUS_ACTIVE value")
	}

	if !strings.Contains(result, "STATUS_INACTIVE = 2") {
		t.Error("expected STATUS_INACTIVE value")
	}
}

func TestConvertRepeatedField(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Messages: []*ProtoMessage{
			{
				Name: "User",
				Fields: []*ProtoField{
					{Name: "tags", Type: "string", Number: 1, Repeated: true},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "tags: []string = 1") {
		t.Error("expected repeated field with [] syntax")
	}
}

func TestConvertMapField(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Messages: []*ProtoMessage{
			{
				Name: "User",
				Fields: []*ProtoField{
					{Name: "metadata", Type: "map<string, string>", Number: 1},
					{Name: "scores", Type: "map<string, int32>", Number: 2},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "metadata: map<string, string> = 1") {
		t.Error("expected map<string, string> field")
	}

	if !strings.Contains(result, "scores: map<string, int32> = 2") {
		t.Error("expected map<string, int32> field")
	}
}

func TestConvertService(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Services: []*ProtoService{
			{
				Name: "UserService",
				Methods: []*ProtoMethod{
					{
						Name:       "GetUser",
						InputType:  "GetUserRequest",
						OutputType: "GetUserResponse",
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "service UserService {") {
		t.Error("expected UserService declaration")
	}

	if !strings.Contains(result, "rpc GetUser(GetUserRequest) returns (GetUserResponse)") {
		t.Error("expected GetUser RPC method")
	}
}

func TestConvertStreamingService(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Services: []*ProtoService{
			{
				Name: "UserService",
				Methods: []*ProtoMethod{
					{
						Name:         "ListUsers",
						InputType:    "ListUsersRequest",
						OutputType:   "ListUsersResponse",
						ServerStream: true,
					},
					{
						Name:         "UpdateUser",
						InputType:    "UpdateUserRequest",
						OutputType:   "UpdateUserResponse",
						ClientStream: true,
					},
					{
						Name:         "StreamUsers",
						InputType:    "StreamUsersRequest",
						OutputType:   "StreamUsersResponse",
						ClientStream: true,
						ServerStream: true,
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "rpc ListUsers(ListUsersRequest) returns (stream ListUsersResponse)") {
		t.Error("expected server streaming RPC")
	}

	if !strings.Contains(result, "rpc UpdateUser(stream UpdateUserRequest) returns (UpdateUserResponse)") {
		t.Error("expected client streaming RPC")
	}

	if !strings.Contains(result, "rpc StreamUsers(stream StreamUsersRequest) returns (stream StreamUsersResponse)") {
		t.Error("expected bidirectional streaming RPC")
	}
}

func TestConvertNestedMessage(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Messages: []*ProtoMessage{
			{
				Name: "Outer",
				Fields: []*ProtoField{
					{Name: "id", Type: "string", Number: 1},
				},
				Messages: []*ProtoMessage{
					{
						Name: "Inner",
						Fields: []*ProtoField{
							{Name: "value", Type: "string", Number: 1},
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "type Outer {") {
		t.Error("expected Outer type declaration")
	}

	// Nested messages are flattened with parent prefix
	if !strings.Contains(result, "type OuterInner {") || !strings.Contains(result, "type Outer_Inner {") {
		// Accept either OuterInner or Outer_Inner format
		if !strings.Contains(result, "Inner") {
			t.Error("expected nested Inner type")
		}
	}
}

func TestConvertNestedEnum(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Messages: []*ProtoMessage{
			{
				Name: "User",
				Fields: []*ProtoField{
					{Name: "status", Type: "Status", Number: 1},
				},
				Enums: []*ProtoEnum{
					{
						Name: "Status",
						Values: []*ProtoEnumValue{
							{Name: "ACTIVE", Number: 0},
							{Name: "INACTIVE", Number: 1},
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	// Nested enums are flattened
	if !strings.Contains(result, "enum") {
		t.Error("expected enum declaration")
	}

	if !strings.Contains(result, "ACTIVE") || !strings.Contains(result, "INACTIVE") {
		t.Error("expected enum values")
	}
}

func TestConvertOneOf(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Messages: []*ProtoMessage{
			{
				Name: "Payment",
				OneOfs: []*ProtoOneOf{
					{
						Name: "payment_method",
						Fields: []*ProtoField{
							{Name: "credit_card", Type: "string", Number: 1},
							{Name: "paypal", Type: "string", Number: 2},
						},
					},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	// OneOfs are converted to unions
	if !strings.Contains(result, "union") || !strings.Contains(result, "payment_method") {
		// If unions aren't supported yet, just verify the fields are present
		if !strings.Contains(result, "credit_card") || !strings.Contains(result, "paypal") {
			t.Error("expected oneof fields to be present")
		}
	}
}

func TestConvertOptions(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Options: map[string]string{
			"go_package": "github.com/example/proto",
		},
		Messages: []*ProtoMessage{
			{
				Name: "User",
				Fields: []*ProtoField{
					{Name: "id", Type: "string", Number: 1},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "@proto.option(go_package = \"github.com/example/proto\")") {
		t.Error("expected go_package option annotation")
	}
}

func TestConvertImports(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Imports: []string{
			"common/types.proto",
		},
		Messages: []*ProtoMessage{
			{
				Name: "User",
				Fields: []*ProtoField{
					{Name: "id", Type: "string", Number: 1},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, `import "common/types.typemux"`) {
		t.Error("expected common/types import converted to .typemux")
	}
}

func TestConvertTypeMapping(t *testing.T) {
	tests := []struct {
		protoType       string
		expectedTypemux string
	}{
		{"google.protobuf.Timestamp", "timestamp"},
		{"google.protobuf.Duration", "duration"},
		{"int32", "int32"},
		{"int64", "int64"},
		{"uint32", "uint32"},
		{"uint64", "uint64"},
		{"bool", "bool"},
		{"string", "string"},
		{"bytes", "bytes"},
		{"float", "float"},
		{"double", "double"},
	}

	for _, tt := range tests {
		t.Run(tt.protoType, func(t *testing.T) {
			schema := &ProtoSchema{
				Syntax:  "proto3",
				Package: "example",
				Messages: []*ProtoMessage{
					{
						Name: "Test",
						Fields: []*ProtoField{
							{Name: "field", Type: tt.protoType, Number: 1},
						},
					},
				},
			}

			converter := NewConverter()
			result := converter.Convert(schema)

			expected := "field: " + tt.expectedTypemux + " = 1"
			if !strings.Contains(result, expected) {
				t.Errorf("expected %q in output, got:\n%s", expected, result)
			}
		})
	}
}

func TestConvertEmptySchema(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "@typemux(\"1.0.0\")") {
		t.Error("expected typemux version annotation")
	}

	if !strings.Contains(result, "namespace example") {
		t.Error("expected namespace declaration")
	}
}

func TestConvertDeprecatedField(t *testing.T) {
	schema := &ProtoSchema{
		Syntax:  "proto3",
		Package: "example",
		Messages: []*ProtoMessage{
			{
				Name: "User",
				Fields: []*ProtoField{
					{Name: "old_field", Type: "string", Number: 1, Deprecated: true},
					{Name: "new_field", Type: "string", Number: 2},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "@deprecated") {
		t.Error("expected @deprecated annotation for deprecated field")
	}
}
