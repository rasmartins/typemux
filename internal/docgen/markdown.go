package docgen

import (
	"fmt"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
)

// MarkdownGenerator generates Markdown API documentation from TypeMUX schemas.
type MarkdownGenerator struct{}

// NewMarkdownGenerator creates a new Markdown documentation generator.
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{}
}

// Generate creates a Markdown documentation string from the given schema.
func (g *MarkdownGenerator) Generate(schema *ast.Schema) string {
	var sb strings.Builder

	// Title
	if schema.Namespace != "" {
		sb.WriteString(fmt.Sprintf("# %s API Documentation\n\n", schema.Namespace))
	} else {
		sb.WriteString("# API Documentation\n\n")
	}

	// Table of Contents
	sb.WriteString("## Table of Contents\n\n")
	if len(schema.Types) > 0 {
		sb.WriteString("- [Types](#types)\n")
	}
	if len(schema.Enums) > 0 {
		sb.WriteString("- [Enums](#enums)\n")
	}
	if len(schema.Unions) > 0 {
		sb.WriteString("- [Unions](#unions)\n")
	}
	if len(schema.Services) > 0 {
		sb.WriteString("- [Services](#services)\n")
	}
	sb.WriteString("\n")

	// Types Section
	if len(schema.Types) > 0 {
		sb.WriteString("## Types\n\n")
		for _, typ := range schema.Types {
			sb.WriteString(g.generateTypeDoc(typ))
			sb.WriteString("\n")
		}
	}

	// Enums Section
	if len(schema.Enums) > 0 {
		sb.WriteString("## Enums\n\n")
		for _, enum := range schema.Enums {
			sb.WriteString(g.generateEnumDoc(enum))
			sb.WriteString("\n")
		}
	}

	// Unions Section
	if len(schema.Unions) > 0 {
		sb.WriteString("## Unions\n\n")
		for _, union := range schema.Unions {
			sb.WriteString(g.generateUnionDoc(union))
			sb.WriteString("\n")
		}
	}

	// Services Section
	if len(schema.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, service := range schema.Services {
			sb.WriteString(g.generateServiceDoc(service))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (g *MarkdownGenerator) generateTypeDoc(typ *ast.Type) string {
	var sb strings.Builder

	// Type name as heading
	sb.WriteString(fmt.Sprintf("### %s\n\n", typ.Name))

	// Documentation
	if typ.Doc != nil {
		if doc := typ.Doc.GetDoc(""); doc != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", doc))
		}
	}

	// Fields table
	if len(typ.Fields) > 0 {
		sb.WriteString("| Field | Type | Required | Description |\n")
		sb.WriteString("|-------|------|----------|-------------|\n")

		for _, field := range typ.Fields {
			typeName := g.formatFieldType(field.Type)
			required := "No"
			if field.Required && !field.Type.Optional {
				required = "Yes"
			} else if field.Type.Optional {
				required = "No"
			}

			description := ""
			if field.Doc != nil {
				description = strings.ReplaceAll(field.Doc.GetDoc(""), "\n", " ")
			}

			// Add deprecation notice
			if field.Deprecated != nil {
				deprecationNote := "⚠️ **DEPRECATED**"
				if field.Deprecated.Reason != "" {
					deprecationNote += ": " + field.Deprecated.Reason
				}
				if description != "" {
					description = deprecationNote + " " + description
				} else {
					description = deprecationNote
				}
			}

			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n",
				field.Name,
				typeName,
				required,
				description))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *MarkdownGenerator) generateEnumDoc(enum *ast.Enum) string {
	var sb strings.Builder

	// Enum name as heading
	sb.WriteString(fmt.Sprintf("### %s\n\n", enum.Name))

	// Documentation
	if enum.Doc != nil {
		if doc := enum.Doc.GetDoc(""); doc != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", doc))
		}
	}

	// Values table
	if len(enum.Values) > 0 {
		sb.WriteString("| Value | Number | Description |\n")
		sb.WriteString("|-------|--------|-------------|\n")

		for _, value := range enum.Values {
			description := ""
			if value.Doc != nil {
				description = strings.ReplaceAll(value.Doc.GetDoc(""), "\n", " ")
			}

			sb.WriteString(fmt.Sprintf("| `%s` | %d | %s |\n",
				value.Name,
				value.Number,
				description))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *MarkdownGenerator) generateUnionDoc(union *ast.Union) string {
	var sb strings.Builder

	// Union name as heading
	sb.WriteString(fmt.Sprintf("### %s\n\n", union.Name))

	// Documentation
	if union.Doc != nil {
		if doc := union.Doc.GetDoc(""); doc != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", doc))
		}
	}

	// Options
	sb.WriteString("**Possible types:**\n\n")
	for _, option := range union.Options {
		sb.WriteString(fmt.Sprintf("- `%s`\n", option))
	}
	sb.WriteString("\n")

	return sb.String()
}

func (g *MarkdownGenerator) generateServiceDoc(service *ast.Service) string {
	var sb strings.Builder

	// Service name as heading
	sb.WriteString(fmt.Sprintf("### %s\n\n", service.Name))

	// Documentation
	if service.Doc != nil {
		if doc := service.Doc.GetDoc(""); doc != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", doc))
		}
	}

	// Methods
	if len(service.Methods) > 0 {
		sb.WriteString("#### Methods\n\n")

		for _, method := range service.Methods {
			sb.WriteString(g.generateMethodDoc(method))
		}
	}

	return sb.String()
}

func (g *MarkdownGenerator) generateMethodDoc(method *ast.Method) string {
	var sb strings.Builder

	// Method signature
	streaming := ""
	if method.OutputStream && method.InputStream {
		streaming = " (bidirectional streaming)"
	} else if method.OutputStream {
		streaming = " (server streaming)"
	} else if method.InputStream {
		streaming = " (client streaming)"
	}

	sb.WriteString(fmt.Sprintf("##### %s%s\n\n", method.Name, streaming))

	// Documentation
	if method.Doc != nil {
		if doc := method.Doc.GetDoc(""); doc != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", doc))
		}
	}

	// Request/Response
	sb.WriteString(fmt.Sprintf("**Request:** `%s`\n\n", method.InputType))
	sb.WriteString(fmt.Sprintf("**Response:** `%s`\n\n", method.OutputType))

	// HTTP mapping (if available)
	if method.HTTPMethod != "" && method.PathTemplate != "" {
		sb.WriteString(fmt.Sprintf("**HTTP:** `%s %s`\n\n", method.HTTPMethod, method.PathTemplate))
	}

	return sb.String()
}

func (g *MarkdownGenerator) formatFieldType(fieldType *ast.FieldType) string {
	var typeName string

	if fieldType.IsMap {
		typeName = fmt.Sprintf("map<%s, %s>", fieldType.MapKey, fieldType.MapValue)
	} else {
		typeName = fieldType.Name
	}

	if fieldType.IsArray {
		typeName = "[]" + typeName
	}

	if fieldType.Optional {
		typeName += "?"
	}

	return typeName
}
