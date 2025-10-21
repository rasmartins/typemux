package graphql

import (
	"strings"
	"testing"
)

func TestConvertBasicType(t *testing.T) {
	schema := &GraphQLSchema{
		Types: []*GraphQLType{
			{
				Name: "User",
				Fields: []*GraphQLField{
					{Name: "id", Type: "ID!"},
					{Name: "name", Type: "String!"},
					{Name: "age", Type: "Int"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

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

func TestConvertEnum(t *testing.T) {
	schema := &GraphQLSchema{
		Enums: []*GraphQLEnum{
			{
				Name: "UserStatus",
				Values: []*GraphQLEnumValue{
					{Name: "ACTIVE"},
					{Name: "INACTIVE"},
					{Name: "PENDING"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "enum UserStatus {") {
		t.Error("expected UserStatus enum declaration")
	}

	if !strings.Contains(result, "ACTIVE") {
		t.Error("expected ACTIVE value")
	}

	if !strings.Contains(result, "INACTIVE") {
		t.Error("expected INACTIVE value")
	}

	if !strings.Contains(result, "PENDING") {
		t.Error("expected PENDING value")
	}
}

func TestConvertInput(t *testing.T) {
	schema := &GraphQLSchema{
		Inputs: []*GraphQLInput{
			{
				Name: "CreateUserInput",
				Fields: []*GraphQLField{
					{Name: "name", Type: "String!"},
					{Name: "email", Type: "String!"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "type CreateUserInput {") {
		t.Error("expected CreateUserInput type declaration")
	}

	if !strings.Contains(result, "name: string") {
		t.Error("expected name field")
	}

	if !strings.Contains(result, "email: string") {
		t.Error("expected email field")
	}
}

func TestConvertListType(t *testing.T) {
	schema := &GraphQLSchema{
		Types: []*GraphQLType{
			{
				Name: "User",
				Fields: []*GraphQLField{
					{Name: "friends", Type: "[User!]!"},
					{Name: "tags", Type: "[String]"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "friends: []User") {
		t.Error("expected friends as list of User")
	}

	if !strings.Contains(result, "tags: []string") {
		t.Error("expected tags as list of string")
	}
}

func TestConvertTypeMapping(t *testing.T) {
	tests := []struct {
		graphqlType     string
		expectedTypemux string
	}{
		{"ID!", "string"},
		{"String!", "string"},
		{"Int!", "int32"},
		{"Float!", "float"},
		{"Boolean!", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.graphqlType, func(t *testing.T) {
			schema := &GraphQLSchema{
				Types: []*GraphQLType{
					{
						Name: "Test",
						Fields: []*GraphQLField{
							{Name: "field", Type: tt.graphqlType},
						},
					},
				},
			}

			converter := NewConverter()
			result := converter.Convert(schema)

			expected := "field: " + tt.expectedTypemux
			if !strings.Contains(result, expected) {
				t.Errorf("expected %q in output for GraphQL type %q", expected, tt.graphqlType)
			}
		})
	}
}

func TestConvertFieldDescription(t *testing.T) {
	schema := &GraphQLSchema{
		Types: []*GraphQLType{
			{
				Name:        "User",
				Description: "A user in the system",
				Fields: []*GraphQLField{
					{Name: "id", Type: "ID!", Description: "The unique identifier"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	// Descriptions should be converted to comments
	if !strings.Contains(result, "//") {
		t.Error("expected descriptions to be converted to comments")
	}
}

func TestConvertEmptySchema(t *testing.T) {
	schema := &GraphQLSchema{}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "@typemux(\"1.0.0\")") {
		t.Error("expected typemux version annotation")
	}

	if !strings.Contains(result, "namespace graphql") {
		t.Error("expected namespace declaration")
	}
}

func TestConvertCompleteSchema(t *testing.T) {
	schema := &GraphQLSchema{
		Scalars: []*GraphQLScalar{
			{Name: "DateTime"},
		},
		Enums: []*GraphQLEnum{
			{Name: "Status", Values: []*GraphQLEnumValue{{Name: "ACTIVE"}, {Name: "INACTIVE"}}},
		},
		Types: []*GraphQLType{
			{
				Name: "User",
				Fields: []*GraphQLField{
					{Name: "id", Type: "ID!"},
					{Name: "name", Type: "String!"},
				},
			},
		},
		Inputs: []*GraphQLInput{
			{
				Name: "CreateUserInput",
				Fields: []*GraphQLField{
					{Name: "name", Type: "String!"},
				},
			},
		},
		Queries: []*GraphQLField{
			{Name: "getUser", Type: "User"},
		},
		Mutations: []*GraphQLField{
			{Name: "createUser", Type: "User"},
		},
		Subscriptions: []*GraphQLField{
			{Name: "userUpdated", Type: "User!"},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	if !strings.Contains(result, "enum Status") {
		t.Error("expected enum declaration")
	}

	if !strings.Contains(result, "type User") {
		t.Error("expected type declaration")
	}

	if !strings.Contains(result, "type CreateUserInput") {
		t.Error("expected input type declaration")
	}

	if !strings.Contains(result, "service") {
		t.Error("expected service declaration")
	}

	if !strings.Contains(result, "GetUser") {
		t.Error("expected GetUser method")
	}

	if !strings.Contains(result, "CreateUser") {
		t.Error("expected CreateUser method")
	}

	if !strings.Contains(result, "UserUpdated") {
		t.Error("expected UserUpdated subscription")
	}

	if !strings.Contains(result, "stream") {
		t.Error("expected stream keyword for subscription")
	}
}

func TestConvertSubscriptionsInService(t *testing.T) {
	// Subscriptions are included in the service along with queries/mutations
	schema := &GraphQLSchema{
		Queries: []*GraphQLField{
			{Name: "getMessage", Type: "Message"},
		},
		Subscriptions: []*GraphQLField{
			{Name: "messageReceived", Type: "Message!"},
		},
		Types: []*GraphQLType{
			{
				Name: "Message",
				Fields: []*GraphQLField{
					{Name: "text", Type: "String!"},
				},
			},
		},
	}

	converter := NewConverter()
	result := converter.Convert(schema)

	// Subscription methods should be in the service
	if !strings.Contains(result, "service") {
		t.Error("expected service declaration")
	}

	if !strings.Contains(result, "MessageReceived") {
		t.Error("expected MessageReceived method")
	}

	if !strings.Contains(result, "stream") {
		t.Error("expected stream keyword for subscription")
	}
}
