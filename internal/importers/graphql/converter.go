package graphql

import (
	"fmt"
	"strings"
)

type Converter struct {
	schema *GraphQLSchema
}

func NewConverter() *Converter {
	return &Converter{}
}

// isReservedKeyword checks if a field name is a TypeMUX reserved keyword
func isReservedKeyword(name string) bool {
	reserved := map[string]bool{
		"namespace": true,
		"import":    true,
		"enum":      true,
		"type":      true,
		"union":     true,
		"service":   true,
		"rpc":       true,
		"returns":   true,
		"stream":    true,
	}
	return reserved[name]
}

// escapeFieldName adds an underscore suffix if the name is a reserved keyword
func escapeFieldName(name string) string {
	if isReservedKeyword(name) {
		return name + "_"
	}
	return name
}

func (c *Converter) Convert(schema *GraphQLSchema) string {
	c.schema = schema

	var sb strings.Builder

	// Header
	sb.WriteString("@typemux(\"1.0.0\")\n")
	sb.WriteString("namespace graphql\n\n")

	// First, convert enums
	for _, enum := range schema.Enums {
		c.writeEnum(&sb, enum)
		sb.WriteString("\n\n")
	}

	// Convert scalars to type aliases or skip if they're common types
	for _, scalar := range schema.Scalars {
		c.writeScalar(&sb, scalar)
		sb.WriteString("\n\n")
	}

	// Convert input types
	for _, input := range schema.Inputs {
		c.writeInput(&sb, input)
		sb.WriteString("\n\n")
	}

	// Convert object types
	for _, typ := range schema.Types {
		c.writeType(&sb, typ)
		sb.WriteString("\n\n")
	}

	// Convert queries and mutations to a service
	if len(schema.Queries) > 0 || len(schema.Mutations) > 0 {
		c.writeService(&sb, schema)
	}

	return sb.String()
}

func (c *Converter) writeEnum(sb *strings.Builder, enum *GraphQLEnum) {
	// Write documentation
	if enum.Description != "" {
		c.writeDocumentation(sb, enum.Description)
	}

	// Write metadata if present
	if typemuxType, ok := enum.Metadata["typemux_type"]; ok {
		sb.WriteString(fmt.Sprintf("@graphql.type(\"%s\")\n", typemuxType))
	}

	sb.WriteString(fmt.Sprintf("enum %s {\n", enum.Name))

	for i, value := range enum.Values {
		if value.Description != "" {
			// Indent description
			lines := strings.Split(value.Description, "\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("  // %s\n", line))
			}
		}
		sb.WriteString(fmt.Sprintf("  %s = %d", value.Name, i))
		if i < len(enum.Values)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n}")
}

func (c *Converter) writeScalar(sb *strings.Builder, scalar *GraphQLScalar) {
	// Map common GraphQL scalars to TypeMUX types
	typemuxType := c.mapScalarType(scalar.Name)
	if typemuxType == scalar.Name {
		// Custom scalar, create a type alias
		if scalar.Description != "" {
			c.writeDocumentation(sb, scalar.Description)
		}
		// For now, map custom scalars to string
		sb.WriteString(fmt.Sprintf("// Custom scalar %s (mapped to string)\n", scalar.Name))
	}
}

func (c *Converter) writeInput(sb *strings.Builder, input *GraphQLInput) {
	// Write documentation
	if input.Description != "" {
		c.writeDocumentation(sb, input.Description)
	}

	// Write metadata if present
	if typemuxType, ok := input.Metadata["typemux_type"]; ok {
		sb.WriteString(fmt.Sprintf("@graphql.type(\"%s\")\n", typemuxType))
	}

	sb.WriteString(fmt.Sprintf("type %s {\n", input.Name))

	for i, field := range input.Fields {
		if field.Description != "" {
			// Indent description
			lines := strings.Split(field.Description, "\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("  // %s\n", line))
			}
		}

		typemuxType := c.convertGraphQLType(field.Type)
		fieldName := escapeFieldName(field.Name)
		sb.WriteString(fmt.Sprintf("  %s: %s = %d", fieldName, typemuxType, i+1))

		if field.DefaultValue != "" {
			sb.WriteString(fmt.Sprintf(" @graphql.default(%s)", field.DefaultValue))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}")
}

func (c *Converter) writeType(sb *strings.Builder, typ *GraphQLType) {
	// Write documentation
	if typ.Description != "" {
		c.writeDocumentation(sb, typ.Description)
	}

	// Write metadata if present
	if typemuxType, ok := typ.Metadata["typemux_type"]; ok {
		sb.WriteString(fmt.Sprintf("@graphql.type(\"%s\")\n", typemuxType))
	}

	sb.WriteString(fmt.Sprintf("type %s {\n", typ.Name))

	for i, field := range typ.Fields {
		if field.Description != "" {
			// Indent description
			lines := strings.Split(field.Description, "\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("  // %s\n", line))
			}
		}

		typemuxType := c.convertGraphQLType(field.Type)
		fieldName := escapeFieldName(field.Name)
		sb.WriteString(fmt.Sprintf("  %s: %s = %d", fieldName, typemuxType, i+1))

		if field.DefaultValue != "" {
			sb.WriteString(fmt.Sprintf(" @graphql.default(%s)", field.DefaultValue))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}")
}

func (c *Converter) writeService(sb *strings.Builder, schema *GraphQLSchema) {
	if len(schema.Queries) == 0 && len(schema.Mutations) == 0 && len(schema.Subscriptions) == 0 {
		return
	}

	sb.WriteString("service GraphQLService {\n")

	// Write queries
	for _, query := range schema.Queries {
		c.writeMethod(sb, query, "query", false)
	}

	// Write mutations
	for _, mutation := range schema.Mutations {
		c.writeMethod(sb, mutation, "mutation", false)
	}

	// Write subscriptions (as streaming RPCs)
	for _, subscription := range schema.Subscriptions {
		c.writeMethod(sb, subscription, "subscription", true)
	}

	sb.WriteString("}")
}

func (c *Converter) writeMethod(sb *strings.Builder, field *GraphQLField, methodType string, isStreaming bool) {
	if field.Description != "" {
		// Indent description
		lines := strings.Split(field.Description, "\n")
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("  // %s\n", line))
		}
	}

	// For GraphQL queries/mutations/subscriptions, we need to create request/response types
	// For now, we'll use the field arguments as input and the return type as output

	inputType := "Empty"
	if len(field.Arguments) > 0 {
		// We'd ideally generate an input type, but for simplicity, we'll inline
		inputType = fmt.Sprintf("%sRequest", strings.Title(field.Name))
	}

	// Wrap non-null types in a response type
	outputTypeName := UnwrapType(field.Type)

	// Note: Annotations on RPC methods are not yet supported in TypeMUX parser
	// We document the method type in comments instead
	sb.WriteString(fmt.Sprintf("  // GraphQL %s\n", methodType))

	// For subscriptions, add stream prefix to output
	if isStreaming {
		sb.WriteString(fmt.Sprintf("  rpc %s(%s) returns (stream %s)\n",
			strings.Title(field.Name),
			inputType,
			outputTypeName))
	} else {
		sb.WriteString(fmt.Sprintf("  rpc %s(%s) returns (%s)\n",
			strings.Title(field.Name),
			inputType,
			outputTypeName))
	}
}

func (c *Converter) writeDocumentation(sb *strings.Builder, doc string) {
	lines := strings.Split(doc, "\n")
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("// %s\n", line))
	}
}

func (c *Converter) convertGraphQLType(graphqlType string) string {
	// Handle non-null types (Type!)
	isRequired := false
	if strings.HasSuffix(graphqlType, "!") {
		isRequired = true
		graphqlType = strings.TrimSuffix(graphqlType, "!")
	}

	// Handle list types ([Type] or [Type!])
	if strings.HasPrefix(graphqlType, "[") && strings.HasSuffix(graphqlType, "]") {
		// Extract inner type
		innerType := graphqlType[1 : len(graphqlType)-1]
		innerType = strings.TrimSuffix(innerType, "!")
		typemuxInnerType := c.mapType(innerType)

		// In TypeMUX, arrays are represented as []Type
		result := "[]" + typemuxInnerType

		// Note: TypeMUX doesn't have a direct way to express non-null lists vs nullable lists
		// We'll use the required annotation if needed
		return result
	}

	// Map scalar and object types
	typemuxType := c.mapType(graphqlType)

	// Note: TypeMUX doesn't have optional/required at the type level for proto compatibility
	// Required/optional is handled by proto3 semantics
	_ = isRequired // We acknowledge the requirement but proto3 handles it differently

	return typemuxType
}

func (c *Converter) mapType(graphqlType string) string {
	// Map GraphQL built-in scalars to TypeMUX types
	switch graphqlType {
	case "String":
		return "string"
	case "Int":
		return "int32"
	case "Float":
		return "float"
	case "Boolean":
		return "bool"
	case "ID":
		return "string"
	default:
		// Check if it's a custom scalar
		return c.mapScalarType(graphqlType)
	}
}

func (c *Converter) mapScalarType(scalarName string) string {
	// Map common custom scalars
	switch scalarName {
	case "Time", "DateTime", "Timestamp":
		return "timestamp"
	case "Date":
		return "string" // Could be a custom type
	case "JSON":
		return "string" // JSON as string
	case "UUID":
		return "string"
	case "URL":
		return "string"
	default:
		// Unknown scalar, map to the name itself (assume it's a type name)
		return scalarName
	}
}
