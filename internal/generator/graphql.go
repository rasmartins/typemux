package generator

import (
	"fmt"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
)

type GraphQLGenerator struct{}

func NewGraphQLGenerator() *GraphQLGenerator {
	return &GraphQLGenerator{}
}

func (g *GraphQLGenerator) Generate(schema *ast.Schema) string {
	var sb strings.Builder

	// Check for duplicate type names across namespaces
	if err := g.checkForDuplicates(schema); err != nil {
		sb.WriteString(fmt.Sprintf("# ERROR: %s\n", err.Error()))
		sb.WriteString("# GraphQL does not support multiple types with the same name.\n")
		sb.WriteString("# Please rename one of the conflicting types or use separate GraphQL schemas.\n")
		return sb.String()
	}

	sb.WriteString("# Generated GraphQL Schema\n")
	if schema.Namespace != "" {
		sb.WriteString(fmt.Sprintf("# Namespace: %s\n", schema.Namespace))
	}
	sb.WriteString("\n")

	// Add namespace-level GraphQL directives (e.g., federation directives)
	if schema.NamespaceAnnotations != nil && len(schema.NamespaceAnnotations.GraphQL) > 0 {
		for _, directive := range schema.NamespaceAnnotations.GraphQL {
			sb.WriteString(fmt.Sprintf("extend schema %s\n", directive))
		}
		sb.WriteString("\n")
	}

	// Add JSON scalar definition for map types
	sb.WriteString("scalar JSON\n\n")

	// Add @oneOf directive for union input types
	sb.WriteString("directive @oneOf on INPUT_OBJECT\n\n")

	// Build a map of union names for quick lookup
	unionNames := make(map[string]bool)
	for _, union := range schema.Unions {
		unionNames[union.Name] = true
	}

	// Generate enums
	for _, enum := range schema.Enums {
		sb.WriteString(g.generateEnum(enum))
		sb.WriteString("\n\n")
	}

	// Determine which types are used as inputs, outputs, or both
	typeUsage := g.analyzeTypeUsage(schema)

	// Build a map of original type names to their custom GraphQL names
	typeNameMap := make(map[string]string)
	for _, typ := range schema.Types {
		if typ.Annotations != nil && typ.Annotations.GraphQLName != "" {
			typeNameMap[typ.Name] = typ.Annotations.GraphQLName
		}
	}

	// Find all types that are union options - they need input versions
	unionOptionTypes := make(map[string]bool)
	for _, union := range schema.Unions {
		for _, option := range union.Options {
			unionOptionTypes[option] = true
		}
	}

	// Generate types
	for _, typ := range schema.Types {
		usage := typeUsage[typ.Name]
		isUnionOption := unionOptionTypes[typ.Name]

		// If used as both input and output, generate both versions
		if usage == "both" || isUnionOption {
			// Generate input version with "Input" suffix
			sb.WriteString(g.generateType(typ, true, true, unionNames, typeUsage, typeNameMap))
			sb.WriteString("\n\n")
			// Generate output version (regular type)
			sb.WriteString(g.generateType(typ, false, false, unionNames, typeUsage, typeNameMap))
			sb.WriteString("\n\n")
		} else if usage == "input" {
			// Only used as input
			sb.WriteString(g.generateType(typ, true, false, unionNames, typeUsage, typeNameMap))
			sb.WriteString("\n\n")
		} else {
			// Only used as output or not used in methods
			sb.WriteString(g.generateType(typ, false, false, unionNames, typeUsage, typeNameMap))
			sb.WriteString("\n\n")
		}
	}

	// Generate unions
	for _, union := range schema.Unions {
		sb.WriteString(g.generateUnion(union))
		sb.WriteString("\n\n")

		// Also generate a @oneOf input type for this union
		sb.WriteString(g.generateUnionInput(union))
		sb.WriteString("\n\n")
	}

	// Generate Query and Mutation types from services
	queryMethods := []string{}
	mutationMethods := []string{}

	for _, service := range schema.Services {
		for _, method := range service.Methods {
			methodStr := g.generateServiceMethod(method, typeUsage)
			// Use GetGraphQLType which checks annotation or uses heuristics
			if method.GetGraphQLType() == "query" {
				queryMethods = append(queryMethods, methodStr)
			} else if method.GetGraphQLType() == "mutation" {
				mutationMethods = append(mutationMethods, methodStr)
			}
			// Note: subscriptions would go here in the future
		}
	}

	if len(queryMethods) > 0 {
		sb.WriteString("type Query {\n")
		for _, method := range queryMethods {
			sb.WriteString("  " + method + "\n")
		}
		sb.WriteString("}\n\n")
	}

	if len(mutationMethods) > 0 {
		sb.WriteString("type Mutation {\n")
		for _, method := range mutationMethods {
			sb.WriteString("  " + method + "\n")
		}
		sb.WriteString("}\n")
	}

	return sb.String()
}

func (g *GraphQLGenerator) analyzeTypeUsage(schema *ast.Schema) map[string]string {
	inputTypes := make(map[string]bool)
	outputTypes := make(map[string]bool)

	// Build a map of types for quick lookup
	typeMap := make(map[string]*ast.Type)
	for _, typ := range schema.Types {
		typeMap[typ.Name] = typ
	}

	// Find all types used as input/output parameters in service methods
	for _, service := range schema.Services {
		for _, method := range service.Methods {
			inputTypes[method.InputType] = true
			outputTypes[method.OutputType] = true
		}
	}

	// Recursively find all types referenced by input types
	visited := make(map[string]bool)
	var findReferencedTypes func(typeName string, asInput bool)
	findReferencedTypes = func(typeName string, asInput bool) {
		if visited[typeName] {
			return
		}
		visited[typeName] = true

		typ := typeMap[typeName]
		if typ == nil {
			return
		}

		for _, field := range typ.Fields {
			// Skip excluded fields
			if !field.ShouldIncludeInGenerator("graphql") {
				continue
			}

			// If this is a custom type (not a primitive), mark it and recurse
			fieldTypeName := field.Type.Name
			if _, exists := typeMap[fieldTypeName]; exists {
				if asInput {
					inputTypes[fieldTypeName] = true
				} else {
					outputTypes[fieldTypeName] = true
				}
				findReferencedTypes(fieldTypeName, asInput)
			}
		}
	}

	// Process all directly used input types
	directInputs := make([]string, 0, len(inputTypes))
	for typeName := range inputTypes {
		directInputs = append(directInputs, typeName)
	}
	for _, typeName := range directInputs {
		findReferencedTypes(typeName, true)
	}

	// Process all directly used output types
	directOutputs := make([]string, 0, len(outputTypes))
	for typeName := range outputTypes {
		directOutputs = append(directOutputs, typeName)
	}
	for _, typeName := range directOutputs {
		findReferencedTypes(typeName, false)
	}

	// Categorize each type
	usage := make(map[string]string)
	allTypes := make(map[string]bool)

	// Collect all type names
	for name := range inputTypes {
		allTypes[name] = true
	}
	for name := range outputTypes {
		allTypes[name] = true
	}

	// Determine usage for each type
	for typeName := range allTypes {
		isInput := inputTypes[typeName]
		isOutput := outputTypes[typeName]

		if isInput && isOutput {
			usage[typeName] = "both"
		} else if isInput {
			usage[typeName] = "input"
		} else if isOutput {
			usage[typeName] = "output"
		}
	}

	return usage
}

func (g *GraphQLGenerator) generateEnum(enum *ast.Enum) string {
	var sb strings.Builder

	// Add documentation - combine multi-line docs into a single string
	if doc := enum.Doc.GetDoc("graphql"); doc != "" {
		// Replace newlines with spaces to create a single-line description
		singleLineDoc := strings.ReplaceAll(doc, "\n", " ")
		sb.WriteString(fmt.Sprintf("\"%s\"\n", singleLineDoc))
	}

	sb.WriteString(fmt.Sprintf("enum %s {\n", enum.Name))
	for _, value := range enum.Values {
		sb.WriteString(fmt.Sprintf("  %s\n", value.Name))
	}
	sb.WriteString("}")
	return sb.String()
}

func (g *GraphQLGenerator) generateUnion(union *ast.Union) string {
	var sb strings.Builder

	// Add documentation
	if doc := union.Doc.GetDoc("graphql"); doc != "" {
		singleLineDoc := strings.ReplaceAll(doc, "\n", " ")
		sb.WriteString(fmt.Sprintf("\"%s\"\n", singleLineDoc))
	}

	sb.WriteString(fmt.Sprintf("union %s = ", union.Name))
	sb.WriteString(strings.Join(union.Options, " | "))
	return sb.String()
}

func (g *GraphQLGenerator) generateUnionInput(union *ast.Union) string {
	var sb strings.Builder

	// Add documentation
	if doc := union.Doc.GetDoc("graphql"); doc != "" {
		singleLineDoc := strings.ReplaceAll(doc, "\n", " ")
		sb.WriteString(fmt.Sprintf("\"%s (Input variant with @oneOf)\"\n", singleLineDoc))
	}

	sb.WriteString(fmt.Sprintf("input %sInput @oneOf {\n", union.Name))
	for _, option := range union.Options {
		// Create optional field for each option (oneOf requires exactly one field to be set)
		fieldName := strings.ToLower(option[:1]) + option[1:] // camelCase
		sb.WriteString(fmt.Sprintf("  %s: %sInput\n", fieldName, option))
	}
	sb.WriteString("}")
	return sb.String()
}

func (g *GraphQLGenerator) generateType(typ *ast.Type, isInput bool, addInputSuffix bool, unionNames map[string]bool, typeUsage map[string]string, typeNameMap map[string]string) string {
	var sb strings.Builder

	// Add documentation - combine multi-line docs into a single string
	if doc := typ.Doc.GetDoc("graphql"); doc != "" {
		// Replace newlines with spaces to create a single-line description
		singleLineDoc := strings.ReplaceAll(doc, "\n", " ")
		sb.WriteString(fmt.Sprintf("\"%s\"\n", singleLineDoc))
	}

	// Use 'input' keyword for types used as input parameters
	keyword := "type"
	typeName := typ.Name

	// Use GraphQLName override if specified
	if typ.Annotations != nil && typ.Annotations.GraphQLName != "" {
		typeName = typ.Annotations.GraphQLName
	}

	if isInput {
		keyword = "input"
		if addInputSuffix {
			// Only add Input suffix if we're not using a custom name
			if typ.Annotations == nil || typ.Annotations.GraphQLName == "" {
				typeName = typ.Name + "Input"
			}
		}
	}

	// Add GraphQL directives to type
	directives := ""
	if !isInput && typ.Annotations != nil && len(typ.Annotations.GraphQL) > 0 {
		directives = " " + strings.Join(typ.Annotations.GraphQL, " ")
	}

	sb.WriteString(fmt.Sprintf("%s %s%s {\n", keyword, typeName, directives))
	for _, field := range typ.Fields {
		// Skip excluded fields
		if !field.ShouldIncludeInGenerator("graphql") {
			continue
		}

		// Build field directives
		var fieldDirectiveParts []string

		// Add @deprecated directive if field is deprecated
		if !isInput && field.Deprecated != nil {
			var deprecationReason string
			if field.Deprecated.Reason != "" {
				// Escape quotes in reason
				deprecationReason = strings.ReplaceAll(field.Deprecated.Reason, "\"", "\\\"")
				fieldDirectiveParts = append(fieldDirectiveParts, fmt.Sprintf("@deprecated(reason: \"%s\")", deprecationReason))
			} else {
				fieldDirectiveParts = append(fieldDirectiveParts, "@deprecated")
			}
		}

		// Add custom GraphQL directives
		if !isInput && field.Annotations != nil && len(field.Annotations.GraphQL) > 0 {
			fieldDirectiveParts = append(fieldDirectiveParts, field.Annotations.GraphQL...)
		}

		fieldDirectives := ""
		if len(fieldDirectiveParts) > 0 {
			fieldDirectives = " " + strings.Join(fieldDirectiveParts, " ")
		}

		// Use UnionInput type for union fields in input types
		if isInput && unionNames[field.Type.Name] {
			gqlType := field.Type.Name + "Input"
			if field.Required {
				gqlType += "!"
			}
			sb.WriteString(fmt.Sprintf("  %s: %s%s\n", field.Name, gqlType, fieldDirectives))
		} else {
			sb.WriteString(fmt.Sprintf("  %s: %s%s\n", field.Name, g.convertFieldType(field, isInput, typeUsage, typeNameMap), fieldDirectives))
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func (g *GraphQLGenerator) convertFieldType(field *ast.Field, isInput bool, typeUsage map[string]string, typeNameMap map[string]string) string {
	gqlType := g.mapTypeToGraphQL(field.Type)

	// Use unqualified name for lookups
	fieldTypeName := ast.GetUnqualifiedName(field.Type.Name)

	// Check if this type has a custom GraphQL name
	if customName, ok := typeNameMap[fieldTypeName]; ok {
		gqlType = customName
	}

	// If this is an input context and the field type is a custom type that has both/input usage,
	// use the Input suffix
	if isInput {
		usage := typeUsage[fieldTypeName]
		if usage == "both" || usage == "input" {
			// If there's a custom name, don't add Input suffix (it's already the custom name)
			// Otherwise add Input suffix to the original name
			if _, hasCustomName := typeNameMap[fieldTypeName]; !hasCustomName {
				gqlType = fieldTypeName + "Input"
			}
		}
	}

	if field.Type.IsArray {
		gqlType = fmt.Sprintf("[%s]", gqlType)
	}

	// In GraphQL, non-null (!) is the default for required fields
	// If the field is explicitly optional (has ?), don't add !
	// If the field is required (@required), add !
	if field.Required && !field.Type.Optional {
		gqlType = gqlType + "!"
	} else if !field.Type.Optional && !field.Required {
		// By default, if not marked as optional and not explicitly required,
		// GraphQL leaves it nullable (no ! suffix)
	}

	return gqlType
}

func (g *GraphQLGenerator) mapTypeToGraphQL(fieldType *ast.FieldType) string {
	if fieldType.IsMap {
		// GraphQL doesn't have native map support, use JSON scalar
		return "JSON"
	}

	typeMap := map[string]string{
		"string":    "String",
		"int32":     "Int",
		"int64":     "Int",
		"float32":   "Float",
		"float64":   "Float",
		"bool":      "Boolean",
		"timestamp": "String", // or use a custom DateTime scalar
		"bytes":     "String", // base64 encoded
	}

	if gqlType, ok := typeMap[fieldType.Name]; ok {
		return gqlType
	}

	// Custom type - use unqualified name for output
	return ast.GetUnqualifiedName(fieldType.Name)
}

func (g *GraphQLGenerator) generateServiceMethod(method *ast.Method, typeUsage map[string]string) string {
	// Convert method name to camelCase
	methodName := strings.ToLower(method.Name[:1]) + method.Name[1:]

	// If the input type is used as both input and output, add "Input" suffix
	inputTypeName := method.InputType
	if typeUsage[method.InputType] == "both" {
		inputTypeName = method.InputType + "Input"
	}

	return fmt.Sprintf("%s(input: %s): %s", methodName, inputTypeName, method.OutputType)
}

// checkForDuplicates checks if there are multiple types/enums with the same unqualified name
// across different namespaces, which would cause conflicts in GraphQL
func (g *GraphQLGenerator) checkForDuplicates(schema *ast.Schema) error {
	typeNames := make(map[string][]string) // unqualified name -> list of namespaces
	enumNames := make(map[string][]string)
	unionNames := make(map[string][]string)

	// Collect all type names with their namespaces
	for _, typ := range schema.Types {
		unqualified := ast.GetUnqualifiedName(typ.Name)
		ns := typ.Namespace
		if ns == "" {
			ns = "default"
		}
		typeNames[unqualified] = append(typeNames[unqualified], ns)
	}

	// Collect all enum names
	for _, enum := range schema.Enums {
		unqualified := ast.GetUnqualifiedName(enum.Name)
		ns := enum.Namespace
		if ns == "" {
			ns = "default"
		}
		enumNames[unqualified] = append(enumNames[unqualified], ns)
	}

	// Collect all union names
	for _, union := range schema.Unions {
		unqualified := ast.GetUnqualifiedName(union.Name)
		ns := union.Namespace
		if ns == "" {
			ns = "default"
		}
		unionNames[unqualified] = append(unionNames[unqualified], ns)
	}

	// Check for duplicates
	for name, namespaces := range typeNames {
		if len(namespaces) > 1 {
			// Remove duplicates from namespace list
			nsSet := make(map[string]bool)
			for _, ns := range namespaces {
				nsSet[ns] = true
			}
			if len(nsSet) > 1 {
				nsList := make([]string, 0, len(nsSet))
				for ns := range nsSet {
					nsList = append(nsList, ns)
				}
				return fmt.Errorf("duplicate type name '%s' found in namespaces: %s", name, strings.Join(nsList, ", "))
			}
		}
	}

	for name, namespaces := range enumNames {
		if len(namespaces) > 1 {
			nsSet := make(map[string]bool)
			for _, ns := range namespaces {
				nsSet[ns] = true
			}
			if len(nsSet) > 1 {
				nsList := make([]string, 0, len(nsSet))
				for ns := range nsSet {
					nsList = append(nsList, ns)
				}
				return fmt.Errorf("duplicate enum name '%s' found in namespaces: %s", name, strings.Join(nsList, ", "))
			}
		}
	}

	for name, namespaces := range unionNames {
		if len(namespaces) > 1 {
			nsSet := make(map[string]bool)
			for _, ns := range namespaces {
				nsSet[ns] = true
			}
			if len(nsSet) > 1 {
				nsList := make([]string, 0, len(nsSet))
				for ns := range nsSet {
					nsList = append(nsList, ns)
				}
				return fmt.Errorf("duplicate union name '%s' found in namespaces: %s", name, strings.Join(nsList, ", "))
			}
		}
	}

	return nil
}
