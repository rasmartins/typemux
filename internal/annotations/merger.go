package annotations

import (
	"fmt"

	"github.com/rasmartins/typemux/internal/ast"
)

// Merger merges YAML annotations into an AST schema
type Merger struct {
	annotations *YAMLAnnotations
}

// NewMerger creates a new merger with the given YAML annotations
func NewMerger(annotations *YAMLAnnotations) *Merger {
	return &Merger{
		annotations: annotations,
	}
}

// Merge applies YAML annotations to the schema
// YAML annotations override inline annotations when there's a conflict
func (m *Merger) Merge(schema *ast.Schema) {
	// Merge type annotations (support both simple and qualified names)
	for _, schemaType := range schema.Types {
		qualifiedName := schemaType.Namespace + "." + schemaType.Name

		// Try qualified name first, then simple name
		if typeAnnotations, ok := m.annotations.Types[qualifiedName]; ok {
			m.mergeTypeAnnotations(schemaType, typeAnnotations)
		} else if typeAnnotations, ok := m.annotations.Types[schemaType.Name]; ok {
			m.mergeTypeAnnotations(schemaType, typeAnnotations)
		}
	}

	// Merge enum annotations (support both simple and qualified names)
	for _, schemaEnum := range schema.Enums {
		qualifiedName := schemaEnum.Namespace + "." + schemaEnum.Name

		if enumAnnotations, ok := m.annotations.Enums[qualifiedName]; ok {
			m.mergeEnumAnnotations(schemaEnum, enumAnnotations)
		} else if enumAnnotations, ok := m.annotations.Enums[schemaEnum.Name]; ok {
			m.mergeEnumAnnotations(schemaEnum, enumAnnotations)
		}
	}

	// Merge union annotations (support both simple and qualified names)
	for _, schemaUnion := range schema.Unions {
		qualifiedName := schemaUnion.Namespace + "." + schemaUnion.Name

		if unionAnnotations, ok := m.annotations.Unions[qualifiedName]; ok {
			m.mergeUnionAnnotations(schemaUnion, unionAnnotations)
		} else if unionAnnotations, ok := m.annotations.Unions[schemaUnion.Name]; ok {
			m.mergeUnionAnnotations(schemaUnion, unionAnnotations)
		}
	}

	// Merge service annotations (support both simple and qualified names)
	for _, schemaService := range schema.Services {
		qualifiedName := schemaService.Namespace + "." + schemaService.Name

		if serviceAnnotations, ok := m.annotations.Services[qualifiedName]; ok {
			m.mergeServiceAnnotations(schemaService, serviceAnnotations)
		} else if serviceAnnotations, ok := m.annotations.Services[schemaService.Name]; ok {
			m.mergeServiceAnnotations(schemaService, serviceAnnotations)
		}
	}
}

func (m *Merger) mergeTypeAnnotations(schemaType *ast.Type, annotations *TypeAnnotations) {
	// Initialize if nil
	if schemaType.Annotations == nil {
		schemaType.Annotations = ast.NewFormatAnnotations()
	}

	// Merge format-specific annotations
	m.applyFormatAnnotations(schemaType.Annotations, annotations.Proto, annotations.GraphQL, annotations.OpenAPI)

	// Merge field annotations
	if annotations.Fields != nil {
		for _, field := range schemaType.Fields {
			if fieldAnnotations, ok := annotations.Fields[field.Name]; ok {
				m.mergeFieldAnnotations(field, fieldAnnotations)
			}
		}
	}
}

func (m *Merger) mergeFieldAnnotations(field *ast.Field, annotations *FieldAnnotations) {
	// YAML overrides inline for required
	if annotations.Required {
		field.Required = true
		if field.Attributes == nil {
			field.Attributes = make(map[string]string)
		}
		field.Attributes["required"] = ""
	}

	// YAML overrides inline for default
	if annotations.Default != "" {
		field.Default = annotations.Default
		if field.Attributes == nil {
			field.Attributes = make(map[string]string)
		}
		field.Attributes["default"] = ""
	}

	// Merge exclude/only lists
	if len(annotations.Exclude) > 0 {
		field.ExcludeFrom = mergeLists(field.ExcludeFrom, annotations.Exclude)
		if field.Attributes == nil {
			field.Attributes = make(map[string]string)
		}
		field.Attributes["exclude"] = ""
	}
	if len(annotations.Only) > 0 {
		field.OnlyFor = mergeLists(field.OnlyFor, annotations.Only)
		if field.Attributes == nil {
			field.Attributes = make(map[string]string)
		}
		field.Attributes["only"] = ""
	}

	// Initialize field annotations if nil
	if field.Annotations == nil {
		field.Annotations = ast.NewFormatAnnotations()
	}

	// Merge format-specific annotations
	m.applyFormatAnnotations(field.Annotations, annotations.Proto, annotations.GraphQL, annotations.OpenAPI)
}

func (m *Merger) mergeEnumAnnotations(schemaEnum *ast.Enum, annotations *EnumAnnotations) {
	// Initialize if nil
	if schemaEnum.Annotations == nil {
		schemaEnum.Annotations = ast.NewFormatAnnotations()
	}

	// Merge format-specific annotations
	m.applyFormatAnnotations(schemaEnum.Annotations, annotations.Proto, annotations.GraphQL, annotations.OpenAPI)
}

func (m *Merger) mergeUnionAnnotations(schemaUnion *ast.Union, annotations *UnionAnnotations) {
	// Initialize if nil
	if schemaUnion.Annotations == nil {
		schemaUnion.Annotations = ast.NewFormatAnnotations()
	}

	// Merge format-specific annotations
	m.applyFormatAnnotations(schemaUnion.Annotations, annotations.Proto, annotations.GraphQL, annotations.OpenAPI)
}

func (m *Merger) mergeServiceAnnotations(schemaService *ast.Service, annotations *ServiceAnnotations) {
	// Note: Service doesn't have Annotations field in the current AST
	// Service name annotations would need to be added to the AST if needed

	// Merge method annotations
	if annotations.Methods != nil {
		for _, method := range schemaService.Methods {
			if methodAnnotations, ok := annotations.Methods[method.Name]; ok {
				m.mergeMethodAnnotations(method, methodAnnotations)
			}
		}
	}
}

func (m *Merger) mergeMethodAnnotations(method *ast.Method, annotations *MethodAnnotations) {
	// YAML overrides inline
	if annotations.HTTP != "" {
		method.HTTPMethod = annotations.HTTP
	}
	if annotations.Path != "" {
		method.PathTemplate = annotations.Path
	}
	if annotations.GraphQL != "" {
		method.GraphQLType = annotations.GraphQL
	}

	// Merge status code lists (convert int to string)
	if len(annotations.Success) > 0 {
		successStrs := make([]string, len(annotations.Success))
		for i, code := range annotations.Success {
			successStrs[i] = fmt.Sprintf("%d", code)
		}
		method.SuccessCodes = mergeLists(method.SuccessCodes, successStrs)
	}
	if len(annotations.Errors) > 0 {
		errorStrs := make([]string, len(annotations.Errors))
		for i, code := range annotations.Errors {
			errorStrs[i] = fmt.Sprintf("%d", code)
		}
		method.ErrorCodes = mergeLists(method.ErrorCodes, errorStrs)
	}

	// Note: Method doesn't have Annotations field for ProtoOption in current AST
	// This would need to be added if proto options on methods are needed
}

// applyFormatAnnotations applies format-specific annotations to an AST FormatAnnotations struct
func (m *Merger) applyFormatAnnotations(target *ast.FormatAnnotations, proto, graphql, openapi *FormatSpecificAnnotations) {
	// Apply proto annotations
	if proto != nil {
		if proto.Name != "" {
			target.ProtoName = proto.Name
		}
		if proto.Option != "" {
			target.Proto = append(target.Proto, proto.Option)
		}
	}

	// Apply graphql annotations
	if graphql != nil {
		if graphql.Name != "" {
			target.GraphQLName = graphql.Name
		}
		if graphql.Directive != "" {
			target.GraphQL = append(target.GraphQL, graphql.Directive)
		}
	}

	// Apply openapi annotations
	if openapi != nil {
		if openapi.Name != "" {
			target.OpenAPIName = openapi.Name
		}
		if openapi.Extension != "" {
			target.OpenAPI = append(target.OpenAPI, openapi.Extension)
		}
	}
}

// mergeLists merges two string lists, removing duplicates
func mergeLists(a, b []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(a)+len(b))

	for _, item := range a {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	for _, item := range b {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
