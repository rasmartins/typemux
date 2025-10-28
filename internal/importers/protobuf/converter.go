package protobuf

import (
	"fmt"
	"strings"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) Convert(schema *ProtoSchema) string {
	var sb strings.Builder

	// Write TypeMUX version
	sb.WriteString("@typemux(\"1.0.0\")\n")

	// Write namespace with options as annotations
	if schema.Package != "" {
		sb.WriteString(fmt.Sprintf("namespace %s", schema.Package))

		// Add proto options as annotations
		if goPackage, ok := schema.Options["go_package"]; ok {
			sb.WriteString(fmt.Sprintf(" @proto.option(go_package = \"%s\")", goPackage))
		}

		sb.WriteString("\n\n")
	}

	// Write imports (skip google and common imports for now)
	for _, imp := range schema.Imports {
		if !strings.Contains(imp, "google/") && !strings.Contains(imp, "commonwatcherproto") {
			// Convert .proto to .typemux
			impName := strings.TrimSuffix(imp, ".proto")
			// Extract just the filename without directory path for same-package imports
			if strings.Contains(impName, "/") {
				parts := strings.Split(impName, "/")
				// If it's in the same package directory (e.g., mypackage/file.proto)
				// just use the filename
				if len(parts) == 2 && parts[0] == schema.Package {
					impName = parts[1]
				}
			}
			sb.WriteString(fmt.Sprintf("import \"%s.typemux\"\n", impName))
		}
	}
	if len(schema.Imports) > 0 {
		sb.WriteString("\n")
	}

	// Write enums
	for _, enum := range schema.Enums {
		c.writeEnum(&sb, enum, 0)
		sb.WriteString("\n")
	}

	// Write messages as types
	for _, msg := range schema.Messages {
		c.writeMessage(&sb, msg, 0)
		sb.WriteString("\n")
	}

	// Write services
	for _, service := range schema.Services {
		c.writeService(&sb, service)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (c *Converter) writeEnum(sb *strings.Builder, enum *ProtoEnum, indent int) {
	indentStr := strings.Repeat("  ", indent)

	sb.WriteString(fmt.Sprintf("%senum %s {\n", indentStr, enum.Name))

	for _, value := range enum.Values {
		sb.WriteString(fmt.Sprintf("%s  %s = %d\n", indentStr, value.Name, value.Number))
	}

	sb.WriteString(fmt.Sprintf("%s}\n", indentStr))
}

func (c *Converter) writeMessage(sb *strings.Builder, msg *ProtoMessage, indent int) {
	indentStr := strings.Repeat("  ", indent)

	// Write nested enums first
	for _, enum := range msg.Enums {
		c.writeEnum(sb, enum, indent)
		sb.WriteString("\n")
	}

	// Write nested messages
	for _, nested := range msg.Messages {
		c.writeMessage(sb, nested, indent)
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("%stype %s {\n", indentStr, msg.Name))

	// Write fields
	for _, field := range msg.Fields {
		c.writeField(sb, field, indent+1)
	}

	// Write oneof fields as optional union members
	for _, oneof := range msg.OneOfs {
		sb.WriteString(fmt.Sprintf("%s  // oneof %s\n", indentStr, oneof.Name))
		for _, field := range oneof.Fields {
			field.Optional = true
			c.writeField(sb, field, indent+1)
		}
	}

	sb.WriteString(fmt.Sprintf("%s}\n", indentStr))
}

func (c *Converter) writeField(sb *strings.Builder, field *ProtoField, indent int) {
	indentStr := strings.Repeat("  ", indent)

	// Convert proto type to TypeMUX type
	typemuxType := c.convertType(field.Type)

	// Handle repeated (arrays)
	if field.Repeated {
		typemuxType = "[]" + typemuxType
	}

	// In proto3, all fields are optional by default, so we don't need the ? marker
	// TypeMUX will treat them as optional unless marked @required

	// Build field line
	fieldLine := fmt.Sprintf("%s%s: %s = %d",
		indentStr,
		field.Name,
		typemuxType,
		field.Number)

	// Add deprecated annotation
	if field.Deprecated {
		fieldLine += " @deprecated"
	}

	sb.WriteString(fieldLine + "\n")
}

func (c *Converter) convertType(protoType string) string {
	// Map protobuf types to TypeMUX types
	typeMap := map[string]string{
		"string":   "string",
		"int32":    "int32",
		"int64":    "int64",
		"uint32":   "uint32",
		"uint64":   "uint64",
		"sint32":   "int32",
		"sint64":   "int64",
		"fixed32":  "uint32",
		"fixed64":  "uint64",
		"sfixed32": "int32",
		"sfixed64": "int64",
		"bool":     "bool",
		"bytes":    "bytes",
		"float":    "float",
		"double":   "double",
	}

	if mapped, ok := typeMap[protoType]; ok {
		return mapped
	}

	// Handle map types
	if strings.HasPrefix(protoType, "map<") {
		// map<string, Type> -> map<string, Type>
		return protoType
	}

	// Handle google.protobuf types
	if strings.HasPrefix(protoType, "google.protobuf.") {
		switch protoType {
		case "google.protobuf.Timestamp":
			return "timestamp"
		case "google.protobuf.Duration":
			return "duration"
		case "google.type.Decimal":
			return "string" // Represent as string for now
		default:
			// Remove google.protobuf prefix
			return strings.TrimPrefix(protoType, "google.protobuf.")
		}
	}

	// Handle google.type types
	if strings.HasPrefix(protoType, "google.type.") {
		return "string" // Represent as string for now
	}

	// Handle google.rpc types
	if strings.HasPrefix(protoType, "google.rpc.") {
		// Remove prefix for custom types
		return strings.TrimPrefix(protoType, "google.rpc.")
	}

	// Handle package-prefixed types (e.g., protolib.AssetID)
	if strings.Contains(protoType, ".") {
		parts := strings.Split(protoType, ".")
		// Return just the type name for now
		return parts[len(parts)-1]
	}

	// Return as-is for custom types
	return protoType
}

func (c *Converter) writeService(sb *strings.Builder, service *ProtoService) {
	sb.WriteString(fmt.Sprintf("service %s {\n", service.Name))

	for _, method := range service.Methods {
		// Build method signature with streaming support
		inputType := method.InputType
		if method.ClientStream {
			inputType = "stream " + inputType
		}

		outputType := method.OutputType
		if method.ServerStream {
			outputType = "stream " + outputType
		}

		sb.WriteString(fmt.Sprintf("  rpc %s(%s) returns (%s)\n",
			method.Name,
			inputType,
			outputType))
	}

	sb.WriteString("}\n")
}
