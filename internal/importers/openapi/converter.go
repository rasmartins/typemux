package openapi

import (
	"fmt"
	"strings"
)

type Converter struct {
	spec            *OpenAPISpec
	resolvedSchemas map[string]*Schema
	fieldCounter    map[string]int
}

func NewConverter() *Converter {
	return &Converter{
		resolvedSchemas: make(map[string]*Schema),
		fieldCounter:    make(map[string]int),
	}
}

func (c *Converter) Convert(spec *OpenAPISpec) string {
	c.spec = spec

	var sb strings.Builder

	// Header
	sb.WriteString("@typemux(\"1.0.0\")\n")

	// Namespace from info.title
	namespace := "api"
	if spec.Info != nil && spec.Info.Title != "" {
		namespace = sanitizeName(spec.Info.Title)
	}
	sb.WriteString(fmt.Sprintf("namespace %s\n\n", namespace))

	// Write description if present
	if spec.Info != nil && spec.Info.Description != "" {
		c.writeDocumentation(&sb, spec.Info.Description)
		sb.WriteString("\n")
	}

	// First, process all schemas from components
	if spec.Components != nil && spec.Components.Schemas != nil {
		// Collect all enums, unions, and types
		var enums []*Schema
		var unions []*Schema
		var types []*Schema

		for name, schema := range spec.Components.Schemas {
			resolvedSchema := c.resolveSchema(schema)
			resolvedSchema.Title = name

			if len(resolvedSchema.Enum) > 0 {
				enums = append(enums, resolvedSchema)
			} else if c.isUnionSchema(resolvedSchema) {
				unions = append(unions, resolvedSchema)
			} else {
				types = append(types, resolvedSchema)
			}
		}

		// Write enums
		for _, enum := range enums {
			c.writeEnum(&sb, enum)
			sb.WriteString("\n\n")
		}

		// Write unions
		for _, union := range unions {
			c.writeUnion(&sb, union)
			sb.WriteString("\n\n")
		}

		// Write types
		for _, typ := range types {
			c.writeType(&sb, typ)
			sb.WriteString("\n\n")
		}
	}

	// Convert paths to service methods
	if len(spec.Paths) > 0 {
		c.writeService(&sb, spec)
	}

	return sb.String()
}

func (c *Converter) resolveSchema(schema *Schema) *Schema {
	if schema.Ref != "" {
		// Check cache first
		if resolved, ok := c.resolvedSchemas[schema.Ref]; ok {
			return resolved
		}

		// Resolve reference
		componentName := ResolveRef(schema.Ref)
		if c.spec.Components != nil && c.spec.Components.Schemas != nil {
			if refSchema, ok := c.spec.Components.Schemas[componentName]; ok {
				c.resolvedSchemas[schema.Ref] = refSchema
				return refSchema
			}
		}
	}

	return schema
}

func (c *Converter) isUnionSchema(schema *Schema) bool {
	return len(schema.OneOf) > 0 || len(schema.AnyOf) > 0
}

func (c *Converter) writeEnum(sb *strings.Builder, schema *Schema) {
	if schema.Description != "" {
		c.writeDocumentation(sb, schema.Description)
	}

	enumName := schema.Title
	if enumName == "" {
		enumName = "UnknownEnum"
	}

	sb.WriteString(fmt.Sprintf("enum %s {\n", enumName))

	for i, value := range schema.Enum {
		// Convert enum value to string
		valueStr := fmt.Sprintf("%v", value)
		// Sanitize enum value name
		valueName := strings.ToUpper(sanitizeName(valueStr))
		sb.WriteString(fmt.Sprintf("  %s = %d\n", valueName, i))
	}

	sb.WriteString("}")
}

func (c *Converter) writeUnion(sb *strings.Builder, schema *Schema) {
	if schema.Description != "" {
		c.writeDocumentation(sb, schema.Description)
	}

	unionName := schema.Title
	if unionName == "" {
		unionName = "UnknownUnion"
	}

	sb.WriteString(fmt.Sprintf("union %s {\n", unionName))

	// Process oneOf schemas (these are the union variants)
	variants := schema.OneOf
	if len(variants) == 0 {
		// Fall back to anyOf if oneOf is not present
		variants = schema.AnyOf
	}

	for _, variant := range variants {
		resolvedVariant := c.resolveSchema(variant)

		// Determine variant name and type
		variantName := ""
		variantType := ""

		if resolvedVariant.Ref != "" {
			// Reference to another type
			typeName := ResolveRef(resolvedVariant.Ref)
			variantName = typeName
			variantType = typeName
		} else if resolvedVariant.Title != "" {
			// Named inline type
			variantName = resolvedVariant.Title
			variantType = resolvedVariant.Title
		} else if len(resolvedVariant.Properties) > 0 {
			// Anonymous object - create inline type definition
			// Generate a name based on the first property or use generic name
			variantName = "Variant"
			for propName := range resolvedVariant.Properties {
				variantName = strings.Title(propName) + "Variant"
				break
			}
			variantType = variantName
		} else {
			// Primitive type
			variantType = c.convertSchemaType(resolvedVariant)
			variantName = strings.Title(variantType) + "Value"
		}

		sb.WriteString(fmt.Sprintf("  %s: %s\n", variantName, variantType))
	}

	sb.WriteString("}")
}

func (c *Converter) writeType(sb *strings.Builder, schema *Schema) {
	if schema.Description != "" {
		c.writeDocumentation(sb, schema.Description)
	}

	typeName := schema.Title
	if typeName == "" {
		typeName = "UnknownType"
	}

	sb.WriteString(fmt.Sprintf("type %s {\n", typeName))

	// Write properties
	fieldNum := 1
	for propName, propSchema := range schema.Properties {
		resolvedSchema := c.resolveSchema(propSchema)

		if resolvedSchema.Description != "" {
			lines := strings.Split(resolvedSchema.Description, "\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("  // %s\n", strings.TrimSpace(line)))
			}
		}

		typemuxType := c.convertSchemaType(resolvedSchema)
		fieldName := escapeFieldName(propName)

		sb.WriteString(fmt.Sprintf("  %s: %s = %d\n", fieldName, typemuxType, fieldNum))
		fieldNum++
	}

	sb.WriteString("}")
}

func (c *Converter) writeService(sb *strings.Builder, spec *OpenAPISpec) {
	serviceName := "APIService"
	if spec.Info != nil && spec.Info.Title != "" {
		serviceName = sanitizeName(spec.Info.Title) + "Service"
	}

	sb.WriteString(fmt.Sprintf("service %s {\n", serviceName))

	// Process each path
	for path, pathItem := range spec.Paths {
		// Process each operation
		operations := map[string]*Operation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"DELETE": pathItem.Delete,
			"PATCH":  pathItem.Patch,
		}

		for method, operation := range operations {
			if operation != nil {
				c.writeMethod(sb, path, method, operation)
			}
		}
	}

	sb.WriteString("}")
}

func (c *Converter) writeMethod(sb *strings.Builder, path string, method string, operation *Operation) {
	if operation.Description != "" {
		lines := strings.Split(operation.Description, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				sb.WriteString(fmt.Sprintf("  // %s\n", trimmed))
			}
		}
	} else if operation.Summary != "" {
		sb.WriteString(fmt.Sprintf("  // %s\n", operation.Summary))
	}

	// Add HTTP method and path as comment
	sb.WriteString(fmt.Sprintf("  // %s %s\n", method, path))

	// Determine method name
	methodName := operation.OperationID
	if methodName == "" {
		// Generate from path and method
		methodName = generateMethodName(path, method)
	}
	methodName = strings.Title(sanitizeName(methodName))

	// Determine input type
	inputType := "Empty"
	if len(operation.Parameters) > 0 || operation.RequestBody != nil {
		inputType = methodName + "Request"
	}

	// Determine output type
	outputType := "Empty"
	// Try to find 200/201 response
	for statusCode, response := range operation.Responses {
		if statusCode == "200" || statusCode == "201" {
			if response.Content != nil {
				for _, mediaType := range response.Content {
					if mediaType.Schema != nil {
						resolvedSchema := c.resolveSchema(mediaType.Schema)
						if resolvedSchema.Ref != "" {
							outputType = ResolveRef(resolvedSchema.Ref)
						} else if resolvedSchema.Title != "" {
							outputType = resolvedSchema.Title
						} else {
							outputType = methodName + "Response"
						}
						break
					}
				}
			}
			break
		}
	}

	sb.WriteString(fmt.Sprintf("  rpc %s(%s) returns (%s)\n", methodName, inputType, outputType))
}

func (c *Converter) convertSchemaType(schema *Schema) string {
	// Handle references
	if schema.Ref != "" {
		return ResolveRef(schema.Ref)
	}

	// Handle arrays
	if schema.Type == "array" {
		if schema.Items != nil {
			itemType := c.convertSchemaType(schema.Items)
			return "[]" + itemType
		}
		return "[]string" // Default array type
	}

	// Handle objects
	if schema.Type == "object" {
		// If it has properties, it should be a named type
		if len(schema.Properties) > 0 && schema.Title != "" {
			return schema.Title
		}
		// Generic object
		return "map<string, string>"
	}

	// Map primitive types
	switch schema.Type {
	case "string":
		if schema.Format == "date-time" || schema.Format == "date" {
			return "timestamp"
		}
		return "string"
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int32"
	case "number":
		if schema.Format == "double" {
			return "double"
		}
		return "float"
	case "boolean":
		return "bool"
	default:
		return "string"
	}
}

func (c *Converter) writeDocumentation(sb *strings.Builder, doc string) {
	lines := strings.Split(doc, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			sb.WriteString(fmt.Sprintf("// %s\n", trimmed))
		}
	}
}

// sanitizeName converts a string to a valid identifier name
func sanitizeName(name string) string {
	// Replace common separators with nothing or underscore
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, "/", "")

	// Remove any non-alphanumeric characters
	var result strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// generateMethodName generates a method name from path and HTTP method
func generateMethodName(path string, method string) string {
	// Remove leading slash and split by /
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	var nameParts []string
	nameParts = append(nameParts, strings.ToLower(method))

	for _, part := range parts {
		// Skip path parameters
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		nameParts = append(nameParts, part)
	}

	// Join and convert to camelCase
	result := ""
	for i, part := range nameParts {
		cleaned := sanitizeName(part)
		if i == 0 {
			result = cleaned
		} else {
			result += strings.Title(cleaned)
		}
	}

	return result
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
