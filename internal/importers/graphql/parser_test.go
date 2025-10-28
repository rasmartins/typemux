package graphql

import (
	"testing"
)

func TestParseBasicType(t *testing.T) {
	input := `type User {
  id: ID!
  name: String!
  age: Int
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]
	if typ.Name != "User" {
		t.Errorf("expected type name %q, got %q", "User", typ.Name)
	}

	if len(typ.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(typ.Fields))
	}

	// Check field properties
	if typ.Fields[0].Name != "id" {
		t.Errorf("expected field name %q, got %q", "id", typ.Fields[0].Name)
	}
	if typ.Fields[0].Type != "ID!" {
		t.Errorf("expected field type %q, got %q", "ID!", typ.Fields[0].Type)
	}
	if !IsNonNull(typ.Fields[0].Type) {
		t.Error("expected id field to be required")
	}
}

func TestParseEnum(t *testing.T) {
	input := `enum Status {
  ACTIVE
  INACTIVE
  PENDING
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
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
		t.Fatalf("expected 3 values, got %d", len(enum.Values))
	}

	expectedValues := []string{"ACTIVE", "INACTIVE", "PENDING"}
	for i, expected := range expectedValues {
		if enum.Values[i].Name != expected {
			t.Errorf("expected value %q, got %q", expected, enum.Values[i].Name)
		}
	}
}

func TestParseInput(t *testing.T) {
	input := `input CreateUserInput {
  name: String!
  email: String!
  age: Int
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(schema.Inputs))
	}

	inputType := schema.Inputs[0]
	if inputType.Name != "CreateUserInput" {
		t.Errorf("expected input name %q, got %q", "CreateUserInput", inputType.Name)
	}

	if len(inputType.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(inputType.Fields))
	}
}

func TestParseInterface(t *testing.T) {
	input := `interface Node {
  id: ID!
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(schema.Interfaces))
	}

	iface := schema.Interfaces[0]
	if iface.Name != "Node" {
		t.Errorf("expected interface name %q, got %q", "Node", iface.Name)
	}

	if len(iface.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(iface.Fields))
	}
}

func TestParseUnion(t *testing.T) {
	input := `union SearchResult = User | Post | Comment`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Unions) != 1 {
		t.Fatalf("expected 1 union, got %d", len(schema.Unions))
	}

	union := schema.Unions[0]
	if union.Name != "SearchResult" {
		t.Errorf("expected union name %q, got %q", "SearchResult", union.Name)
	}

	if len(union.Types) != 3 {
		t.Fatalf("expected 3 types, got %d", len(union.Types))
	}

	expectedTypes := []string{"User", "Post", "Comment"}
	for i, expected := range expectedTypes {
		if union.Types[i] != expected {
			t.Errorf("expected type %q, got %q", expected, union.Types[i])
		}
	}
}

func TestParseScalar(t *testing.T) {
	input := `scalar DateTime`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Scalars) != 1 {
		t.Fatalf("expected 1 scalar, got %d", len(schema.Scalars))
	}

	scalar := schema.Scalars[0]
	if scalar.Name != "DateTime" {
		t.Errorf("expected scalar name %q, got %q", "DateTime", scalar.Name)
	}
}

func TestParseExtendTypeQuery(t *testing.T) {
	input := `extend type Query {
  getUser(id: ID!): User
  listUsers: [User!]!
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Queries) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(schema.Queries))
	}

	query := schema.Queries[0]
	if query.Name != "getUser" {
		t.Errorf("expected query name %q, got %q", "getUser", query.Name)
	}

	if len(query.Arguments) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(query.Arguments))
	}

	if query.Arguments[0].Name != "id" {
		t.Errorf("expected arg name %q, got %q", "id", query.Arguments[0].Name)
	}
}

func TestParseExtendTypeMutation(t *testing.T) {
	input := `extend type Mutation {
  createUser(input: CreateUserInput!): User
  deleteUser(id: ID!): Boolean
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Mutations) != 2 {
		t.Fatalf("expected 2 mutations, got %d", len(schema.Mutations))
	}

	mutation := schema.Mutations[0]
	if mutation.Name != "createUser" {
		t.Errorf("expected mutation name %q, got %q", "createUser", mutation.Name)
	}
}

func TestParseExtendTypeSubscription(t *testing.T) {
	input := `extend type Subscription {
  userUpdated(id: ID!): User!
  messageReceived: Message!
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Subscriptions) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(schema.Subscriptions))
	}

	subscription := schema.Subscriptions[0]
	if subscription.Name != "userUpdated" {
		t.Errorf("expected subscription name %q, got %q", "userUpdated", subscription.Name)
	}
}

func TestParseFieldWithDescription(t *testing.T) {
	input := `type User {
  """
  The unique identifier for the user
  """
  id: ID!

  # The user's name
  name: String!
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]
	if typ.Fields[0].Description == "" {
		t.Error("expected description for id field")
	}
}

func TestParseListType(t *testing.T) {
	input := `type User {
  friends: [User!]!
  tags: [String]
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(schema.Types))
	}

	typ := schema.Types[0]

	if !IsList(typ.Fields[0].Type) {
		t.Error("expected friends field to be a list")
	}

	if !IsNonNull(typ.Fields[0].Type) {
		t.Error("expected friends field to be required")
	}

	if !IsList(typ.Fields[1].Type) {
		t.Error("expected tags field to be a list")
	}
}

func TestParseCompleteSchema(t *testing.T) {
	input := `scalar DateTime

enum UserStatus {
  ACTIVE
  INACTIVE
}

type User {
  id: ID!
  name: String!
  status: UserStatus!
  createdAt: DateTime!
}

input CreateUserInput {
  name: String!
  email: String!
}

extend type Query {
  getUser(id: ID!): User
}

extend type Mutation {
  createUser(input: CreateUserInput!): User
}`

	parser := NewParser(input)
	schema, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Scalars) != 1 {
		t.Errorf("expected 1 scalar, got %d", len(schema.Scalars))
	}

	if len(schema.Enums) != 1 {
		t.Errorf("expected 1 enum, got %d", len(schema.Enums))
	}

	if len(schema.Types) != 1 {
		t.Errorf("expected 1 type, got %d", len(schema.Types))
	}

	if len(schema.Inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(schema.Inputs))
	}

	if len(schema.Queries) != 1 {
		t.Errorf("expected 1 query, got %d", len(schema.Queries))
	}

	if len(schema.Mutations) != 1 {
		t.Errorf("expected 1 mutation, got %d", len(schema.Mutations))
	}
}
