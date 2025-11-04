package docgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
	"github.com/rasmartins/typemux/internal/generator"
)

// Generator generates documentation from schemas
type Generator struct {
	schema     *ast.Schema
	outputDir  string
	graphqlGen *generator.GraphQLGenerator
	protoGen   *generator.ProtobufGenerator
	openapiGen *generator.OpenAPIGenerator
}

// NewGenerator creates a new documentation generator
func NewGenerator(schema *ast.Schema, outputDir string) *Generator {
	return &Generator{
		schema:     schema,
		outputDir:  outputDir,
		graphqlGen: generator.NewGraphQLGenerator(),
		protoGen:   generator.NewProtobufGenerator(),
		openapiGen: generator.NewOpenAPIGenerator(),
	}
}

// Generate creates all documentation files
func (g *Generator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate index page
	if err := g.generateIndex(); err != nil {
		return fmt.Errorf("failed to generate index: %w", err)
	}

	// Generate type documentation
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	// Generate service documentation
	if err := g.generateServices(); err != nil {
		return fmt.Errorf("failed to generate services: %w", err)
	}

	// Generate enum documentation
	if err := g.generateEnums(); err != nil {
		return fmt.Errorf("failed to generate enums: %w", err)
	}

	// Generate cross-format guide
	if err := g.generateCrossFormatGuide(); err != nil {
		return fmt.Errorf("failed to generate cross-format guide: %w", err)
	}

	return nil
}

func (g *Generator) generateIndex() error {
	var sb strings.Builder

	sb.WriteString("# API Documentation\n\n")

	if g.schema.Namespace != "" {
		sb.WriteString(fmt.Sprintf("**Namespace:** `%s`\n\n", g.schema.Namespace))
	}

	sb.WriteString("This API is available in multiple formats:\n\n")
	sb.WriteString("- 🔷 **GraphQL** - Query language with flexible data fetching\n")
	sb.WriteString("- 🔶 **REST/OpenAPI** - Standard HTTP/JSON API\n")
	sb.WriteString("- 🔸 **gRPC/Protobuf** - High-performance binary protocol\n\n")

	sb.WriteString("## Table of Contents\n\n")

	if len(g.schema.Types) > 0 {
		sb.WriteString("### Types\n\n")
		for _, typ := range g.schema.Types {
			sb.WriteString(fmt.Sprintf("- [%s](types/%s.md)\n", typ.Name, strings.ToLower(typ.Name)))
		}
		sb.WriteString("\n")
	}

	if len(g.schema.Enums) > 0 {
		sb.WriteString("### Enums\n\n")
		for _, enum := range g.schema.Enums {
			sb.WriteString(fmt.Sprintf("- [%s](enums/%s.md)\n", enum.Name, strings.ToLower(enum.Name)))
		}
		sb.WriteString("\n")
	}

	if len(g.schema.Services) > 0 {
		sb.WriteString("### Services\n\n")
		for _, svc := range g.schema.Services {
			sb.WriteString(fmt.Sprintf("- [%s](services/%s.md)\n", svc.Name, strings.ToLower(svc.Name)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Guides\n\n")
	sb.WriteString("- [Cross-Format Usage Guide](cross-format-guide.md) - How to use the same API across different formats\n")

	return g.writeFile("README.md", sb.String())
}

func (g *Generator) generateTypes() error {
	typesDir := filepath.Join(g.outputDir, "types")
	if err := os.MkdirAll(typesDir, 0o750); err != nil {
		return err
	}

	for _, typ := range g.schema.Types {
		if err := g.generateTypeDoc(typ, typesDir); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateTypeDoc(typ *ast.Type, outputDir string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", typ.Name))

	// Add documentation if present
	if doc := typ.Doc.GetDoc(""); doc != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", doc))
	}

	sb.WriteString("## Format Representations\n\n")

	// GraphQL representation
	sb.WriteString("### GraphQL\n\n")
	sb.WriteString("```graphql\n")
	sb.WriteString(g.generateGraphQLType(typ))
	sb.WriteString("```\n\n")

	// OpenAPI representation
	sb.WriteString("### OpenAPI\n\n")
	sb.WriteString("```yaml\n")
	sb.WriteString(g.generateOpenAPIType(typ))
	sb.WriteString("```\n\n")

	// Protobuf representation
	sb.WriteString("### Protobuf\n\n")
	sb.WriteString("```protobuf\n")
	sb.WriteString(g.generateProtobufType(typ))
	sb.WriteString("```\n\n")

	// Fields documentation
	if len(typ.Fields) > 0 {
		sb.WriteString("## Fields\n\n")
		for _, field := range typ.Fields {
			sb.WriteString(g.generateFieldDoc(typ, field))
		}
	}

	fileName := filepath.Join(outputDir, strings.ToLower(typ.Name)+".md")
	return g.writeFile(fileName, sb.String())
}

func (g *Generator) generateFieldDoc(typ *ast.Type, field *ast.Field) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### %s\n\n", field.Name))

	// Field documentation
	if doc := field.Doc.GetDoc(""); doc != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", doc))
	}

	// Type information
	typeStr := field.Type.Name
	if field.Type.IsArray {
		typeStr = "[]" + typeStr
	}
	sb.WriteString(fmt.Sprintf("**Type:** `%s`", typeStr))
	if field.Required {
		sb.WriteString(" (required)")
	}
	sb.WriteString("\n\n")

	// Field arguments (if any)
	if len(field.Arguments) > 0 {
		sb.WriteString("**Arguments:**\n\n")
		sb.WriteString("| Argument | Type | Required | Default | Description |\n")
		sb.WriteString("|----------|------|----------|---------|-------------|\n")

		for _, arg := range field.Arguments {
			required := "No"
			if arg.Required {
				required = "Yes"
			}
			defaultVal := arg.Default
			if defaultVal == "" {
				defaultVal = "-"
			}
			description := ""
			if doc := arg.Doc.GetDoc(""); doc != "" {
				description = strings.ReplaceAll(doc, "\n", " ")
			}

			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s | %s |\n",
				arg.Name, arg.Type.Name, required, defaultVal, description))
		}
		sb.WriteString("\n")

		// Usage examples for fields with arguments
		sb.WriteString(g.generateFieldArgumentExamples(typ, field))
	}

	return sb.String()
}

func (g *Generator) generateFieldArgumentExamples(typ *ast.Type, field *ast.Field) string {
	var sb strings.Builder

	sb.WriteString("#### Usage Examples\n\n")

	// GraphQL example
	sb.WriteString("**GraphQL:**\n```graphql\n")
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  %s", field.Name))

	if len(field.Arguments) > 0 {
		sb.WriteString("(")
		args := []string{}
		for _, arg := range field.Arguments {
			exampleVal := g.getExampleValue(arg.Type.Name, arg.Default)
			args = append(args, fmt.Sprintf("%s: %s", arg.Name, exampleVal))
		}
		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	}

	sb.WriteString(" {\n")
	// Show a few fields from the return type
	if returnType := g.findType(field.Type.Name); returnType != nil {
		count := 0
		for _, f := range returnType.Fields {
			if count < 3 && len(f.Arguments) == 0 {
				sb.WriteString(fmt.Sprintf("    %s\n", f.Name))
				count++
			}
		}
	}
	sb.WriteString("  }\n}\n")
	sb.WriteString("```\n\n")

	// REST example
	sb.WriteString("**REST:**\n```bash\n")

	// Build URL with query parameters
	path := "/" + g.toKebabCase(typ.Name)
	if typ.Name == "Query" || typ.Name == "Mutation" {
		path = "/" + g.toKebabCase(field.Name)
	} else {
		path = path + "/{id}/" + g.toKebabCase(field.Name)
	}

	sb.WriteString(fmt.Sprintf("GET %s", path))

	if len(field.Arguments) > 0 {
		queryParams := []string{}
		for _, arg := range field.Arguments {
			exampleVal := g.getExampleValue(arg.Type.Name, arg.Default)
			// Remove quotes for URL
			exampleVal = strings.Trim(exampleVal, "\"")
			queryParams = append(queryParams, fmt.Sprintf("%s=%s", arg.Name, exampleVal))
		}
		sb.WriteString("?" + strings.Join(queryParams, "&"))
	}

	sb.WriteString("\n```\n\n")

	// gRPC example
	sb.WriteString("**gRPC:**\n```protobuf\n")

	methodName := g.capitalize(field.Name)
	requestName := typ.Name + methodName + "Request"

	sb.WriteString(fmt.Sprintf("// Call the %s method\n", methodName))
	sb.WriteString(fmt.Sprintf("client.%s(%s {\n", methodName, requestName))

	for _, arg := range field.Arguments {
		exampleVal := g.getExampleValue(arg.Type.Name, arg.Default)
		// Remove quotes for protobuf
		exampleVal = strings.Trim(exampleVal, "\"")
		sb.WriteString(fmt.Sprintf("  %s: %s\n", arg.Name, exampleVal))
	}

	sb.WriteString("})\n```\n\n")

	return sb.String()
}

func (g *Generator) generateEnums() error {
	enumsDir := filepath.Join(g.outputDir, "enums")
	if err := os.MkdirAll(enumsDir, 0o750); err != nil {
		return err
	}

	for _, enum := range g.schema.Enums {
		if err := g.generateEnumDoc(enum, enumsDir); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateEnumDoc(enum *ast.Enum, outputDir string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", enum.Name))

	if doc := enum.Doc.GetDoc(""); doc != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", doc))
	}

	sb.WriteString("## Values\n\n")
	sb.WriteString("| Value | Description |\n")
	sb.WriteString("|-------|-------------|\n")

	for _, val := range enum.Values {
		description := ""
		if val.Doc != nil {
			if doc := val.Doc.GetDoc(""); doc != "" {
				description = strings.ReplaceAll(doc, "\n", " ")
			}
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", val.Name, description))
	}

	sb.WriteString("\n")

	fileName := filepath.Join(outputDir, strings.ToLower(enum.Name)+".md")
	return g.writeFile(fileName, sb.String())
}

func (g *Generator) generateServices() error {
	if len(g.schema.Services) == 0 {
		return nil
	}

	servicesDir := filepath.Join(g.outputDir, "services")
	if err := os.MkdirAll(servicesDir, 0o750); err != nil {
		return err
	}

	for _, svc := range g.schema.Services {
		if err := g.generateServiceDoc(svc, servicesDir); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateServiceDoc(svc *ast.Service, outputDir string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", svc.Name))

	if doc := svc.Doc.GetDoc(""); doc != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", doc))
	}

	sb.WriteString("## Methods\n\n")

	for _, method := range svc.Methods {
		sb.WriteString(fmt.Sprintf("### %s\n\n", method.Name))

		if doc := method.Doc.GetDoc(""); doc != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", doc))
		}

		sb.WriteString(fmt.Sprintf("**Input:** `%s`\n\n", method.InputType))
		sb.WriteString(fmt.Sprintf("**Output:** `%s`\n\n", method.OutputType))
	}

	fileName := filepath.Join(outputDir, strings.ToLower(svc.Name)+".md")
	return g.writeFile(fileName, sb.String())
}

func (g *Generator) generateCrossFormatGuide() error {
	var sb strings.Builder

	sb.WriteString("# Cross-Format Usage Guide\n\n")
	sb.WriteString("This guide shows how to use the same API across different formats.\n\n")

	sb.WriteString("## Overview\n\n")
	sb.WriteString("The API is available in three formats, each optimized for different use cases:\n\n")

	sb.WriteString("### GraphQL\n\n")
	sb.WriteString("- **Best for:** Web and mobile apps that need flexible data fetching\n")
	sb.WriteString("- **Advantages:** Request exactly the data you need, reduce over-fetching\n")
	sb.WriteString("- **Protocol:** HTTP/JSON\n")
	sb.WriteString("- **Endpoint:** `/graphql`\n\n")

	sb.WriteString("### REST/OpenAPI\n\n")
	sb.WriteString("- **Best for:** Traditional web apps, simple integrations\n")
	sb.WriteString("- **Advantages:** Standard HTTP methods, easy to cache, widely supported\n")
	sb.WriteString("- **Protocol:** HTTP/JSON\n")
	sb.WriteString("- **Base URL:** `/api/v1`\n\n")

	sb.WriteString("### gRPC/Protobuf\n\n")
	sb.WriteString("- **Best for:** Microservices, high-performance systems\n")
	sb.WriteString("- **Advantages:** Binary protocol, type-safe, efficient, streaming support\n")
	sb.WriteString("- **Protocol:** HTTP/2 with Protocol Buffers\n\n")

	sb.WriteString("## Field Arguments Across Formats\n\n")
	sb.WriteString("TypeMUX field arguments map naturally to each format:\n\n")

	sb.WriteString("| TypeMUX | GraphQL | REST | gRPC |\n")
	sb.WriteString("|---------|---------|------|------|\n")
	sb.WriteString("| `user(id: string)` | Field arguments | Query parameters | Request message fields |\n")
	sb.WriteString("| Required args | `id: String!` | Required query param | Non-optional field |\n")
	sb.WriteString("| Optional args | `limit: Int` | Optional query param | Optional field |\n")
	sb.WriteString("| Default values | `limit: Int = 10` | Default in docs | Default in proto3 |\n\n")

	// Find a good example type with field arguments
	for _, typ := range g.schema.Types {
		for _, field := range typ.Fields {
			if len(field.Arguments) > 0 {
				sb.WriteString(fmt.Sprintf("## Example: %s.%s\n\n", typ.Name, field.Name))
				sb.WriteString(g.generateFieldArgumentExamples(typ, field))
				break
			}
		}
	}

	return g.writeFile("cross-format-guide.md", sb.String())
}

// Helper methods

func (g *Generator) generateGraphQLType(typ *ast.Type) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s {\n", typ.Name))

	for _, field := range typ.Fields {
		// Skip fields with arguments - they're shown separately
		if len(field.Arguments) > 0 {
			continue
		}

		fieldType := field.Type.Name
		if field.Type.IsArray {
			fieldType = "[" + fieldType + "]"
		}
		if field.Required {
			fieldType += "!"
		}

		sb.WriteString(fmt.Sprintf("  %s: %s\n", field.Name, fieldType))
	}

	sb.WriteString("}")
	return sb.String()
}

func (g *Generator) generateOpenAPIType(typ *ast.Type) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s:\n", typ.Name))
	sb.WriteString("  type: object\n")

	if len(typ.Fields) > 0 {
		sb.WriteString("  properties:\n")
		for _, field := range typ.Fields {
			// Skip fields with arguments
			if len(field.Arguments) > 0 {
				continue
			}

			sb.WriteString(fmt.Sprintf("    %s:\n", field.Name))

			if field.Type.IsArray {
				sb.WriteString("      type: array\n")
				sb.WriteString("      items:\n")
				sb.WriteString(fmt.Sprintf("        $ref: '#/components/schemas/%s'\n", field.Type.Name))
			} else {
				openAPIType := g.mapToOpenAPIType(field.Type.Name)
				if openAPIType == "ref" {
					sb.WriteString(fmt.Sprintf("      $ref: '#/components/schemas/%s'\n", field.Type.Name))
				} else {
					sb.WriteString(fmt.Sprintf("      type: %s\n", openAPIType))
				}
			}
		}
	}

	return sb.String()
}

func (g *Generator) generateProtobufType(typ *ast.Type) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("message %s {\n", typ.Name))

	fieldNum := 1
	for _, field := range typ.Fields {
		// Skip fields with arguments
		if len(field.Arguments) > 0 {
			continue
		}

		optional := ""
		if !field.Required {
			optional = "optional "
		}

		protoType := g.mapToProtoType(field.Type.Name)

		if field.Type.IsArray {
			sb.WriteString(fmt.Sprintf("  repeated %s %s = %d;\n", protoType, field.Name, fieldNum))
		} else {
			sb.WriteString(fmt.Sprintf("  %s%s %s = %d;\n", optional, protoType, field.Name, fieldNum))
		}

		fieldNum++
	}

	sb.WriteString("}")
	return sb.String()
}

func (g *Generator) findType(name string) *ast.Type {
	for _, typ := range g.schema.Types {
		if typ.Name == name {
			return typ
		}
	}
	return nil
}

func (g *Generator) getExampleValue(typeName, defaultValue string) string {
	if defaultValue != "" {
		return defaultValue
	}

	switch typeName {
	case "string":
		return "\"example\""
	case "int32", "int64", "uint32", "uint64":
		return "10"
	case "float32", "float64":
		return "10.5"
	case "bool":
		return "true"
	default:
		return "\"value\""
	}
}

func (g *Generator) mapToOpenAPIType(typeName string) string {
	switch typeName {
	case "string":
		return "string"
	case "int32", "int64", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "ref"
	}
}

func (g *Generator) mapToProtoType(typeName string) string {
	switch typeName {
	case "string":
		return "string"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "bool":
		return "bool"
	default:
		return typeName
	}
}

func (g *Generator) toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func (g *Generator) capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (g *Generator) writeFile(filePath string, content string) error {
	// If filePath is already absolute or starts with outputDir, use it as-is
	// Otherwise, make it relative to outputDir
	if !filepath.IsAbs(filePath) && !strings.HasPrefix(filePath, g.outputDir) {
		filePath = filepath.Join(g.outputDir, filePath)
	}

	// Create parent directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(content), 0o600)
}
