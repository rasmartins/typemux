package generator

import (
	"fmt"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
)

// ProtobufGenerator generates Protocol Buffers (proto3) schemas from TypeMUX schemas.
type ProtobufGenerator struct{}

// NewProtobufGenerator creates a new Protobuf schema generator.
func NewProtobufGenerator() *ProtobufGenerator {
	return &ProtobufGenerator{}
}

// Note on nested maps:
// Protobuf supports maps natively using the map<K,V> syntax.
// For nested maps (e.g., map<string, map<string, string>>), users should
// create explicit wrapper message types. For example:
//
//   type StringMapValue {
//     data: map<string, string>
//   }
//
//   type MyType {
//     nested_data: map<string, StringMapValue>
//   }
//
// This will generate valid Protobuf with proper message types:
//
//   message StringMapValue {
//     map<string, string> data = 1;
//   }
//
//   message MyType {
//     map<string, StringMapValue> nested_data = 1;
//   }

// GenerateByNamespace generates separate Protobuf files per namespace
// Returns a map of namespace -> proto file content
func (g *ProtobufGenerator) GenerateByNamespace(schema *ast.Schema) map[string]string {
	result := make(map[string]string)

	// Helper function to create namespace schema with annotations
	createNamespaceSchema := func(ns string) *ast.Schema {
		nsSchema := &ast.Schema{
			Namespace: ns,
			Enums:     []*ast.Enum{},
			Types:     []*ast.Type{},
			Unions:    []*ast.Union{},
			Services:  []*ast.Service{},
		}
		// Copy namespace annotations if this is the main namespace
		if ns == schema.Namespace && schema.NamespaceAnnotations != nil {
			nsSchema.NamespaceAnnotations = schema.NamespaceAnnotations
		}
		return nsSchema
	}

	// Group types, enums, unions, and services by namespace
	namespaceData := make(map[string]*ast.Schema)

	for _, enum := range schema.Enums {
		ns := enum.Namespace
		if ns == "" {
			ns = "api"
		}
		if namespaceData[ns] == nil {
			namespaceData[ns] = createNamespaceSchema(ns)
		}
		namespaceData[ns].Enums = append(namespaceData[ns].Enums, enum)
	}

	for _, typ := range schema.Types {
		ns := typ.Namespace
		if ns == "" {
			ns = "api"
		}
		if namespaceData[ns] == nil {
			namespaceData[ns] = createNamespaceSchema(ns)
		}
		namespaceData[ns].Types = append(namespaceData[ns].Types, typ)
	}

	for _, union := range schema.Unions {
		ns := union.Namespace
		if ns == "" {
			ns = "api"
		}
		if namespaceData[ns] == nil {
			namespaceData[ns] = createNamespaceSchema(ns)
		}
		namespaceData[ns].Unions = append(namespaceData[ns].Unions, union)
	}

	for _, service := range schema.Services {
		ns := service.Namespace
		if ns == "" {
			ns = "api"
		}
		if namespaceData[ns] == nil {
			namespaceData[ns] = createNamespaceSchema(ns)
		}
		namespaceData[ns].Services = append(namespaceData[ns].Services, service)
	}

	// Generate a proto file for each namespace
	for ns, nsSchema := range namespaceData {
		result[ns] = g.generateForNamespace(nsSchema)
	}

	return result
}

// generateForNamespace generates a single proto file for a specific namespace
func (g *ProtobufGenerator) generateForNamespace(nsSchema *ast.Schema) string {
	var sb strings.Builder

	sb.WriteString("// Generated Protobuf Schema\n")
	sb.WriteString("syntax = \"proto3\";\n\n")
	sb.WriteString(fmt.Sprintf("package %s;\n\n", nsSchema.Namespace))

	// Add namespace-level protobuf options
	if nsSchema.NamespaceAnnotations != nil && len(nsSchema.NamespaceAnnotations.Proto) > 0 {
		for _, option := range nsSchema.NamespaceAnnotations.Proto {
			// Options should be in format: go_package="value" or option_name="value"
			sb.WriteString(fmt.Sprintf("option %s;\n", option))
		}
		sb.WriteString("\n")
	}

	// Collect required imports from other namespaces
	requiredNamespaces := g.findRequiredNamespaces(nsSchema)

	// Add imports for other namespace proto files
	for _, reqNs := range requiredNamespaces {
		if reqNs != nsSchema.Namespace {
			// Convert namespace to file path (e.g., com.example.users -> com/example/users.proto)
			protoPath := strings.ReplaceAll(reqNs, ".", "/") + ".proto"
			sb.WriteString(fmt.Sprintf("import \"%s\";\n", protoPath))
		}
	}

	sb.WriteString("import \"google/protobuf/timestamp.proto\";\n\n")

	// Generate enums
	for _, enum := range nsSchema.Enums {
		sb.WriteString(g.generateEnum(enum))
		sb.WriteString("\n\n")
	}

	// Generate message types
	for _, typ := range nsSchema.Types {
		sb.WriteString(g.generateMessageWithNamespace(typ, nsSchema.Namespace))
		sb.WriteString("\n\n")
	}

	// Generate unions
	for _, union := range nsSchema.Unions {
		sb.WriteString(g.generateUnion(union))
		sb.WriteString("\n\n")
	}

	// Generate services
	for _, service := range nsSchema.Services {
		sb.WriteString(g.generateService(service))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// findRequiredNamespaces finds all namespaces that are referenced by types in the given schema
func (g *ProtobufGenerator) findRequiredNamespaces(nsSchema *ast.Schema) []string {
	required := make(map[string]bool)

	// Check all field types in messages
	for _, typ := range nsSchema.Types {
		for _, field := range typ.Fields {
			if strings.Contains(field.Type.Name, ".") {
				// This is a qualified name, extract the namespace
				parts := strings.Split(field.Type.Name, ".")
				if len(parts) > 1 {
					// Namespace is everything except the last part
					ns := strings.Join(parts[:len(parts)-1], ".")
					required[ns] = true
				}
			}
		}
	}

	// Check service method types
	for _, service := range nsSchema.Services {
		for _, method := range service.Methods {
			if strings.Contains(method.InputType, ".") {
				parts := strings.Split(method.InputType, ".")
				if len(parts) > 1 {
					ns := strings.Join(parts[:len(parts)-1], ".")
					required[ns] = true
				}
			}
			if strings.Contains(method.OutputType, ".") {
				parts := strings.Split(method.OutputType, ".")
				if len(parts) > 1 {
					ns := strings.Join(parts[:len(parts)-1], ".")
					required[ns] = true
				}
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(required))
	for ns := range required {
		result = append(result, ns)
	}
	return result
}

// Generate creates a Protocol Buffers (proto3) schema string from the given schema.
func (g *ProtobufGenerator) Generate(schema *ast.Schema) string {
	var sb strings.Builder

	sb.WriteString("// Generated Protobuf Schema\n")
	sb.WriteString("syntax = \"proto3\";\n\n")

	// Use namespace from schema, default to "api" if empty
	namespace := schema.Namespace
	if namespace == "" {
		namespace = "api"
	}
	sb.WriteString(fmt.Sprintf("package %s;\n\n", namespace))

	// Add namespace-level protobuf options
	if schema.NamespaceAnnotations != nil && len(schema.NamespaceAnnotations.Proto) > 0 {
		for _, option := range schema.NamespaceAnnotations.Proto {
			// Options should be in format: go_package="value" or option_name="value"
			sb.WriteString(fmt.Sprintf("option %s;\n", option))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("import \"google/protobuf/timestamp.proto\";\n\n")

	// Build a map of original type names to their custom Protobuf names
	typeNameMap := make(map[string]string)
	for _, typ := range schema.Types {
		if typ.Annotations != nil && typ.Annotations.ProtoName != "" {
			typeNameMap[typ.Name] = typ.Annotations.ProtoName
		}
	}

	// Generate enums
	for _, enum := range schema.Enums {
		sb.WriteString(g.generateEnum(enum))
		sb.WriteString("\n\n")
	}

	// Generate message types
	for _, typ := range schema.Types {
		sb.WriteString(g.generateMessageWithMap(typ, typeNameMap))
		sb.WriteString("\n\n")
	}

	// Generate unions as messages with oneof
	for _, union := range schema.Unions {
		sb.WriteString(g.generateUnion(union))
		sb.WriteString("\n\n")
	}

	// Generate services
	for _, service := range schema.Services {
		sb.WriteString(g.generateService(service))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (g *ProtobufGenerator) generateEnum(enum *ast.Enum) string {
	var sb strings.Builder

	// Add enum documentation
	if doc := enum.Doc.GetDoc("proto"); doc != "" {
		for _, line := range strings.Split(doc, "\n") {
			sb.WriteString(fmt.Sprintf("// %s\n", line))
		}
	}

	sb.WriteString(fmt.Sprintf("enum %s {\n", enum.Name))

	// Check if there's already a value with number 0
	hasZeroValue := false
	for _, value := range enum.Values {
		if value.HasNumber && value.Number == 0 {
			hasZeroValue = true
			break
		}
	}

	// Only add UNSPECIFIED if there's no value with number 0
	if !hasZeroValue {
		sb.WriteString(fmt.Sprintf("  %s_UNSPECIFIED = 0;\n", strings.ToUpper(enum.Name)))
	}

	nextAutoNumber := 1
	for _, value := range enum.Values {
		// Add enum value documentation
		if doc := value.Doc.GetDoc("proto"); doc != "" {
			for _, line := range strings.Split(doc, "\n") {
				sb.WriteString(fmt.Sprintf("  // %s\n", line))
			}
		}

		var number int
		if value.HasNumber {
			number = value.Number
			// Update nextAutoNumber to be after this custom number
			if value.Number >= nextAutoNumber {
				nextAutoNumber = value.Number + 1
			}
		} else {
			number = nextAutoNumber
			nextAutoNumber++
		}
		sb.WriteString(fmt.Sprintf("  %s = %d;\n", value.Name, number))
	}
	sb.WriteString("}")
	return sb.String()
}

func (g *ProtobufGenerator) generateMessage(typ *ast.Type) string {
	return g.generateMessageWithNamespace(typ, typ.Namespace)
}

func (g *ProtobufGenerator) generateMessageWithMap(typ *ast.Type, typeNameMap map[string]string) string {
	return g.generateMessageWithNamespaceAndMap(typ, typ.Namespace, typeNameMap)
}

func (g *ProtobufGenerator) generateMessageWithNamespace(typ *ast.Type, currentNamespace string) string {
	return g.generateMessageWithNamespaceAndMap(typ, currentNamespace, make(map[string]string))
}

func (g *ProtobufGenerator) generateMessageWithNamespaceAndMap(typ *ast.Type, currentNamespace string, typeNameMap map[string]string) string {
	var sb strings.Builder

	// Add type documentation
	if doc := typ.Doc.GetDoc("proto"); doc != "" {
		for _, line := range strings.Split(doc, "\n") {
			sb.WriteString(fmt.Sprintf("// %s\n", line))
		}
	}

	// Use ProtoName override if specified, otherwise use typ.Name
	messageName := typ.Name
	if typ.Annotations != nil && typ.Annotations.ProtoName != "" {
		messageName = typ.Annotations.ProtoName
	}

	sb.WriteString(fmt.Sprintf("message %s {\n", messageName))
	nextAutoNumber := 1
	for _, field := range typ.Fields {
		// Skip excluded fields
		if !field.ShouldIncludeInGenerator("proto") {
			continue
		}

		// Add field documentation
		if doc := field.Doc.GetDoc("proto"); doc != "" {
			for _, line := range strings.Split(doc, "\n") {
				sb.WriteString(fmt.Sprintf("  // %s\n", line))
			}
		}

		// Add deprecation warning
		if field.Deprecated != nil {
			sb.WriteString("  // DEPRECATED")
			if field.Deprecated.Since != "" {
				sb.WriteString(fmt.Sprintf(" (since %s)", field.Deprecated.Since))
			}
			if field.Deprecated.Removed != "" {
				sb.WriteString(fmt.Sprintf(" - will be removed in %s", field.Deprecated.Removed))
			}
			sb.WriteString("\n")
			if field.Deprecated.Reason != "" {
				sb.WriteString(fmt.Sprintf("  // %s\n", field.Deprecated.Reason))
			}
		}

		// Add since version info
		if field.Since != "" {
			sb.WriteString(fmt.Sprintf("  // Added in version %s\n", field.Since))
		}

		// Determine field number (custom or auto)
		var fieldNum int
		if field.HasNumber {
			fieldNum = field.Number
			// Update nextAutoNumber to be after this custom number
			if field.Number >= nextAutoNumber {
				nextAutoNumber = field.Number + 1
			}
		} else {
			fieldNum = nextAutoNumber
			nextAutoNumber++
		}

		fieldStr := g.generateMessageFieldWithNamespaceAndMap(field, fieldNum, currentNamespace, typeNameMap)
		sb.WriteString(fmt.Sprintf("  %s\n", fieldStr))
	}
	sb.WriteString("}")
	return sb.String()
}

func (g *ProtobufGenerator) generateUnion(union *ast.Union) string {
	var sb strings.Builder

	// Add union documentation
	if doc := union.Doc.GetDoc("proto"); doc != "" {
		for _, line := range strings.Split(doc, "\n") {
			sb.WriteString(fmt.Sprintf("// %s\n", line))
		}
	}

	sb.WriteString(fmt.Sprintf("message %s {\n", union.Name))
	sb.WriteString("  oneof value {\n")

	// Generate oneof options
	fieldNum := 1
	for _, option := range union.Options {
		sb.WriteString(fmt.Sprintf("    %s %s = %d;\n",
			option,
			strings.ToLower(option[:1])+option[1:], // camelCase the field name
			fieldNum))
		fieldNum++
	}

	sb.WriteString("  }\n")
	sb.WriteString("}")
	return sb.String()
}

func (g *ProtobufGenerator) generateMessageField(field *ast.Field, fieldNum int) string {
	return g.generateMessageFieldWithNamespace(field, fieldNum, "")
}

func (g *ProtobufGenerator) generateMessageFieldWithNamespace(field *ast.Field, fieldNum int, currentNamespace string) string {
	return g.generateMessageFieldWithNamespaceAndMap(field, fieldNum, currentNamespace, make(map[string]string))
}

func (g *ProtobufGenerator) generateMessageFieldWithNamespaceAndMap(field *ast.Field, fieldNum int, currentNamespace string, typeNameMap map[string]string) string {
	var protoType string
	if currentNamespace != "" {
		protoType = g.mapTypeToProtobufWithNamespaceAndMap(field.Type, currentNamespace, typeNameMap)
	} else {
		protoType = g.mapTypeToProtobufWithMap(field.Type, typeNameMap)
	}

	// Build field options
	var optionParts []string

	// Add deprecation option if field is deprecated
	if field.Deprecated != nil {
		optionParts = append(optionParts, "deprecated = true")
	}

	// Add validation rules (using buf validate constraints)
	if field.Validation != nil {
		validationOpts := g.buildValidationOptions(field)
		optionParts = append(optionParts, validationOpts...)
	}

	// Add format-specific annotations
	if field.Annotations != nil && len(field.Annotations.Proto) > 0 {
		optionParts = append(optionParts, field.Annotations.Proto...)
	}

	var options string
	if len(optionParts) > 0 {
		options = " [" + strings.Join(optionParts, ", ") + "]"
	}

	if field.Type.IsMap {
		var keyType, valueType string
		keyType = g.mapScalarTypeWithPackageAndMap(field.Type.MapKey, currentNamespace, typeNameMap)

		// Handle the value type - it can be a nested map or simple type
		valueFieldType := field.Type.GetMapValueType()
		if valueFieldType.IsMap {
			// Recursively handle nested map
			valueType = g.generateMapTypeString(valueFieldType, currentNamespace, typeNameMap)
		} else if currentNamespace != "" {
			valueType = g.mapScalarTypeWithPackageAndMap(valueFieldType.Name, currentNamespace, typeNameMap)
		} else {
			valueType = g.mapScalarTypeWithMap(valueFieldType.Name, typeNameMap)
		}

		return fmt.Sprintf("map<%s, %s> %s = %d%s;",
			keyType,
			valueType,
			field.Name,
			fieldNum,
			options)
	}

	if field.Type.IsArray {
		return fmt.Sprintf("repeated %s %s = %d%s;", protoType, field.Name, fieldNum, options)
	}

	// Handle optional fields (proto3 optional keyword)
	if field.Type.Optional {
		return fmt.Sprintf("optional %s %s = %d%s;", protoType, field.Name, fieldNum, options)
	}

	// Proto3 doesn't have required keyword, all fields are optional by default
	return fmt.Sprintf("%s %s = %d%s;", protoType, field.Name, fieldNum, options)
}

// generateMapTypeString recursively generates the protobuf type string for a map (including nested maps)
func (g *ProtobufGenerator) generateMapTypeString(fieldType *ast.FieldType, currentNamespace string, typeNameMap map[string]string) string {
	if !fieldType.IsMap {
		// Base case: not a map, just return the type name
		if currentNamespace != "" {
			return g.mapScalarTypeWithPackageAndMap(fieldType.Name, currentNamespace, typeNameMap)
		}
		return g.mapScalarTypeWithMap(fieldType.Name, typeNameMap)
	}

	// Recursive case: this is a map
	keyType := g.mapScalarTypeWithPackageAndMap(fieldType.MapKey, currentNamespace, typeNameMap)

	valueFieldType := fieldType.GetMapValueType()
	var valueType string
	if valueFieldType.IsMap {
		// Recursively handle nested map
		valueType = g.generateMapTypeString(valueFieldType, currentNamespace, typeNameMap)
	} else if currentNamespace != "" {
		valueType = g.mapScalarTypeWithPackageAndMap(valueFieldType.Name, currentNamespace, typeNameMap)
	} else {
		valueType = g.mapScalarTypeWithMap(valueFieldType.Name, typeNameMap)
	}

	return fmt.Sprintf("map<%s, %s>", keyType, valueType)
}

func (g *ProtobufGenerator) mapScalarType(typeName string) string {
	typeMap := map[string]string{
		"string":    "string",
		"int32":     "int32",
		"int64":     "int64",
		"uint8":     "uint32", // Protobuf has no uint8, use uint32
		"uint16":    "uint32", // Protobuf has no uint16, use uint32
		"uint32":    "uint32",
		"uint64":    "uint64",
		"float32":   "float",
		"float64":   "double",
		"bool":      "bool",
		"timestamp": "google.protobuf.Timestamp",
		"bytes":     "bytes",
	}

	if protoType, ok := typeMap[typeName]; ok {
		return protoType
	}

	// Custom type - use unqualified name for output
	return ast.GetUnqualifiedName(typeName)
}

// mapScalarTypeWithPackage maps a type name to protobuf type, using package-qualified names for cross-package references
func (g *ProtobufGenerator) mapScalarTypeWithPackage(typeName string, currentNamespace string) string {
	return g.mapScalarTypeWithPackageAndMap(typeName, currentNamespace, make(map[string]string))
}

func (g *ProtobufGenerator) mapTypeToProtobufWithMap(fieldType *ast.FieldType, typeNameMap map[string]string) string {
	if fieldType.IsMap {
		return "map"
	}

	return g.mapScalarTypeWithMap(fieldType.Name, typeNameMap)
}

func (g *ProtobufGenerator) mapTypeToProtobufWithNamespaceAndMap(fieldType *ast.FieldType, currentNamespace string, typeNameMap map[string]string) string {
	if fieldType.IsMap {
		return "map"
	}

	return g.mapScalarTypeWithPackageAndMap(fieldType.Name, currentNamespace, typeNameMap)
}

func (g *ProtobufGenerator) mapScalarTypeWithMap(typeName string, typeNameMap map[string]string) string {
	typeMap := map[string]string{
		"string":    "string",
		"int32":     "int32",
		"int64":     "int64",
		"uint8":     "uint32", // Protobuf has no uint8, use uint32
		"uint16":    "uint32", // Protobuf has no uint16, use uint32
		"uint32":    "uint32",
		"uint64":    "uint64",
		"float32":   "float",
		"float64":   "double",
		"bool":      "bool",
		"timestamp": "google.protobuf.Timestamp",
		"bytes":     "bytes",
	}

	if protoType, ok := typeMap[typeName]; ok {
		return protoType
	}

	// Get unqualified name for lookup
	unqualifiedName := ast.GetUnqualifiedName(typeName)

	// Check if this type has a custom Protobuf name
	if customName, ok := typeNameMap[unqualifiedName]; ok {
		return customName
	}

	// Custom type - use unqualified name for output
	return unqualifiedName
}

func (g *ProtobufGenerator) mapScalarTypeWithPackageAndMap(typeName string, currentNamespace string, typeNameMap map[string]string) string {
	typeMap := map[string]string{
		"string":    "string",
		"int32":     "int32",
		"int64":     "int64",
		"uint8":     "uint32", // Protobuf has no uint8, use uint32
		"uint16":    "uint32", // Protobuf has no uint16, use uint32
		"uint32":    "uint32",
		"uint64":    "uint64",
		"float32":   "float",
		"float64":   "double",
		"bool":      "bool",
		"timestamp": "google.protobuf.Timestamp",
		"bytes":     "bytes",
	}

	if protoType, ok := typeMap[typeName]; ok {
		return protoType
	}

	// Check if this is a qualified type name (contains dots)
	if strings.Contains(typeName, ".") {
		// Extract namespace from qualified name
		parts := strings.Split(typeName, ".")
		if len(parts) > 1 {
			typeNs := strings.Join(parts[:len(parts)-1], ".")
			unqualifiedType := parts[len(parts)-1]

			// Check if this type has a custom Protobuf name
			if customName, ok := typeNameMap[unqualifiedType]; ok {
				unqualifiedType = customName
			}

			// If the type is from a different namespace, use fully qualified name
			if typeNs != currentNamespace {
				return typeNs + "." + unqualifiedType
			}
			// Same namespace - use unqualified
			return unqualifiedType
		}
	}

	// Unqualified custom type - check for custom name
	if customName, ok := typeNameMap[typeName]; ok {
		return customName
	}

	// Use as-is
	return typeName
}

func (g *ProtobufGenerator) generateService(service *ast.Service) string {
	var sb strings.Builder

	// Add service documentation
	if doc := service.Doc.GetDoc("proto"); doc != "" {
		for _, line := range strings.Split(doc, "\n") {
			sb.WriteString(fmt.Sprintf("// %s\n", line))
		}
	}

	sb.WriteString(fmt.Sprintf("service %s {\n", service.Name))
	for _, method := range service.Methods {
		// Add method documentation
		if doc := method.Doc.GetDoc("proto"); doc != "" {
			for _, line := range strings.Split(doc, "\n") {
				sb.WriteString(fmt.Sprintf("  // %s\n", line))
			}
		}

		// Build input type with optional stream prefix
		inputType := method.InputType
		if method.InputStream {
			inputType = "stream " + inputType
		}

		// Build output type with optional stream prefix
		outputType := method.OutputType
		if method.OutputStream {
			outputType = "stream " + outputType
		}

		sb.WriteString(fmt.Sprintf("  rpc %s(%s) returns (%s);\n",
			method.Name,
			inputType,
			outputType))
	}
	sb.WriteString("}")
	return sb.String()
}

// buildValidationOptions builds protobuf validation options from validation rules
// Uses buf validate constraint syntax
func (g *ProtobufGenerator) buildValidationOptions(field *ast.Field) []string {
	var opts []string
	v := field.Validation

	// String validation
	if v.MinLength != nil {
		opts = append(opts, fmt.Sprintf("(buf.validate.field).string.min_len = %d", *v.MinLength))
	}
	if v.MaxLength != nil {
		opts = append(opts, fmt.Sprintf("(buf.validate.field).string.max_len = %d", *v.MaxLength))
	}
	if v.Pattern != "" {
		// Escape quotes in pattern
		pattern := strings.ReplaceAll(v.Pattern, "\"", "\\\"")
		opts = append(opts, fmt.Sprintf("(buf.validate.field).string.pattern = \"%s\"", pattern))
	}
	if v.Format != "" {
		switch v.Format {
		case "email":
			opts = append(opts, "(buf.validate.field).string.email = true")
		case "uuid":
			opts = append(opts, "(buf.validate.field).string.uuid = true")
		case "uri", "url":
			opts = append(opts, "(buf.validate.field).string.uri = true")
		case "hostname":
			opts = append(opts, "(buf.validate.field).string.hostname = true")
		case "ipv4":
			opts = append(opts, "(buf.validate.field).string.ipv4 = true")
		case "ipv6":
			opts = append(opts, "(buf.validate.field).string.ipv6 = true")
		}
	}

	// Numeric validation (for int/float types)
	if v.Min != nil {
		if field.Type.Name == "int32" || field.Type.Name == "int64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).int.gte = %d", int64(*v.Min)))
		} else if field.Type.Name == "float32" || field.Type.Name == "float64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).float.gte = %f", *v.Min))
		}
	}
	if v.Max != nil {
		if field.Type.Name == "int32" || field.Type.Name == "int64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).int.lte = %d", int64(*v.Max)))
		} else if field.Type.Name == "float32" || field.Type.Name == "float64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).float.lte = %f", *v.Max))
		}
	}
	if v.ExclusiveMin != nil {
		if field.Type.Name == "int32" || field.Type.Name == "int64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).int.gt = %d", int64(*v.ExclusiveMin)))
		} else if field.Type.Name == "float32" || field.Type.Name == "float64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).float.gt = %f", *v.ExclusiveMin))
		}
	}
	if v.ExclusiveMax != nil {
		if field.Type.Name == "int32" || field.Type.Name == "int64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).int.lt = %d", int64(*v.ExclusiveMax)))
		} else if field.Type.Name == "float32" || field.Type.Name == "float64" {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).float.lt = %f", *v.ExclusiveMax))
		}
	}

	// Array validation (for repeated fields)
	if field.Type.IsArray {
		if v.MinItems != nil {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).repeated.min_items = %d", *v.MinItems))
		}
		if v.MaxItems != nil {
			opts = append(opts, fmt.Sprintf("(buf.validate.field).repeated.max_items = %d", *v.MaxItems))
		}
		if v.UniqueItems {
			opts = append(opts, "(buf.validate.field).repeated.unique = true")
		}
	}

	// Enum validation
	if len(v.Enum) > 0 {
		// For string enums, use the in constraint
		opts = append(opts, fmt.Sprintf("(buf.validate.field).string.in = [\"%s\"]", strings.Join(v.Enum, "\", \"")))
	}

	return opts
}
