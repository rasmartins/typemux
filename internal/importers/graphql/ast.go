package graphql

// GraphQLSchema represents a complete GraphQL schema
type GraphQLSchema struct {
	Types         []*GraphQLType
	Inputs        []*GraphQLInput
	Enums         []*GraphQLEnum
	Scalars       []*GraphQLScalar
	Queries       []*GraphQLField
	Mutations     []*GraphQLField
	Subscriptions []*GraphQLField
	Interfaces    []*GraphQLInterface
	Unions        []*GraphQLUnion
}

// GraphQLType represents a GraphQL object type
type GraphQLType struct {
	Name        string
	Description string
	Fields      []*GraphQLField
	Implements  []string // Interface names this type implements
	Directives  []*GraphQLDirective
	Metadata    map[string]string // For storing typemux:begin/end comments
}

// GraphQLInput represents a GraphQL input type
type GraphQLInput struct {
	Name        string
	Description string
	Fields      []*GraphQLField
	Directives  []*GraphQLDirective
	Metadata    map[string]string
}

// GraphQLField represents a field in a type or input
type GraphQLField struct {
	Name         string
	Type         string
	Description  string
	Arguments    []*GraphQLArgument
	DefaultValue string
	Directives   []*GraphQLDirective
}

// GraphQLArgument represents an argument to a field
type GraphQLArgument struct {
	Name         string
	Type         string
	Description  string
	DefaultValue string
	Directives   []*GraphQLDirective
}

// GraphQLEnum represents a GraphQL enum type
type GraphQLEnum struct {
	Name        string
	Description string
	Values      []*GraphQLEnumValue
	Directives  []*GraphQLDirective
	Metadata    map[string]string
}

// GraphQLEnumValue represents a value in an enum
type GraphQLEnumValue struct {
	Name        string
	Description string
	Directives  []*GraphQLDirective
}

// GraphQLScalar represents a custom scalar type
type GraphQLScalar struct {
	Name        string
	Description string
	Directives  []*GraphQLDirective
}

// GraphQLInterface represents a GraphQL interface
type GraphQLInterface struct {
	Name        string
	Description string
	Fields      []*GraphQLField
	Directives  []*GraphQLDirective
}

// GraphQLUnion represents a GraphQL union type
type GraphQLUnion struct {
	Name        string
	Description string
	Types       []string // Names of types in the union
	Directives  []*GraphQLDirective
}

// GraphQLDirective represents a directive applied to a schema element
type GraphQLDirective struct {
	Name      string
	Arguments map[string]string
}

// IsNonNull returns true if the type is non-null (ends with !)
func IsNonNull(typeStr string) bool {
	return len(typeStr) > 0 && typeStr[len(typeStr)-1] == '!'
}

// IsList returns true if the type is a list (contains [ ])
func IsList(typeStr string) bool {
	return len(typeStr) > 0 && typeStr[0] == '['
}

// UnwrapType removes type modifiers (!, []) and returns the base type
func UnwrapType(typeStr string) string {
	// Remove trailing !
	for len(typeStr) > 0 && typeStr[len(typeStr)-1] == '!' {
		typeStr = typeStr[:len(typeStr)-1]
	}

	// Remove [ and ]
	if len(typeStr) > 0 && typeStr[0] == '[' {
		typeStr = typeStr[1:]
		if len(typeStr) > 0 && typeStr[len(typeStr)-1] == ']' {
			typeStr = typeStr[:len(typeStr)-1]
		}
	}

	// Remove any remaining !
	for len(typeStr) > 0 && typeStr[len(typeStr)-1] == '!' {
		typeStr = typeStr[:len(typeStr)-1]
	}

	return typeStr
}
