package generator

import (
	"encoding/json"
	"fmt"

	"github.com/rasmartins/typemux/internal/ast"
	"gopkg.in/yaml.v3"
)

// OpenAPIAnnotator adds x-go-type and x-go-name annotations to OpenAPI specs
// to enable oapi-codegen to use existing Go models instead of generating new ones
type OpenAPIAnnotator struct {
	baseGenerator *OpenAPIGenerator
}

// NewOpenAPIAnnotator creates a new OpenAPI annotator
func NewOpenAPIAnnotator() *OpenAPIAnnotator {
	return &OpenAPIAnnotator{
		baseGenerator: NewOpenAPIGenerator(),
	}
}

// OpenAPIAnnotatorOptions contains options for OpenAPI annotation
type OpenAPIAnnotatorOptions struct {
	ModelsPackage string // Go package path for models (e.g., "github.com/user/project/models")
	TypePrefix    string // Optional prefix for type names in package
}

// Generate creates an annotated OpenAPI specification with x-go-type extensions
func (a *OpenAPIAnnotator) Generate(schema *ast.Schema, opts *OpenAPIAnnotatorOptions) (string, error) {
	if opts == nil {
		opts = &OpenAPIAnnotatorOptions{
			ModelsPackage: "models",
		}
	}

	// First generate the base OpenAPI spec
	baseYAML := a.baseGenerator.Generate(schema)

	// Parse it back to add annotations
	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(baseYAML), &spec); err != nil {
		return "", fmt.Errorf("failed to parse base OpenAPI spec: %w", err)
	}

	// Add x-go-type annotations to schemas
	a.annotateSchemas(&spec, schema, opts)

	// Marshal back to YAML
	data, err := yaml.Marshal(&spec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal annotated OpenAPI spec: %w", err)
	}

	return string(data), nil
}

// annotateSchemas adds x-go-type and x-go-name annotations to all schemas
func (a *OpenAPIAnnotator) annotateSchemas(spec *OpenAPISpec, schema *ast.Schema, opts *OpenAPIAnnotatorOptions) {
	if spec.Components.Schemas == nil {
		return
	}

	// Build a map of TypeMUX types for quick lookup
	typeMap := make(map[string]*ast.Type)
	for _, typ := range schema.Types {
		typeMap[typ.Name] = typ
	}

	// Build a map of TypeMUX enums
	enumMap := make(map[string]*ast.Enum)
	for _, enum := range schema.Enums {
		enumMap[enum.Name] = enum
	}

	// Build a map of TypeMUX unions
	unionMap := make(map[string]*ast.Union)
	for _, union := range schema.Unions {
		unionMap[union.Name] = union
	}

	// Annotate each schema
	for schemaName, schemaObj := range spec.Components.Schemas {
		// Check if this is a TypeMUX type
		if _, isType := typeMap[schemaName]; isType {
			a.annotateTypeSchema(schemaName, &schemaObj, opts)
		}

		// Check if this is a TypeMUX enum
		if _, isEnum := enumMap[schemaName]; isEnum {
			a.annotateEnumSchema(schemaName, &schemaObj, opts)
		}

		// Check if this is a TypeMUX union
		if union, isUnion := unionMap[schemaName]; isUnion {
			a.annotateUnionSchema(schemaName, &schemaObj, union, opts)
		}

		// Update the schema in the spec
		spec.Components.Schemas[schemaName] = schemaObj
	}
}

// annotateTypeSchema adds x-go-type annotation for regular types
func (a *OpenAPIAnnotator) annotateTypeSchema(typeName string, schema *OpenAPISchema, opts *OpenAPIAnnotatorOptions) {
	if schema.Extensions == nil {
		schema.Extensions = make(map[string]interface{})
	}

	// Add x-go-type with import path
	goType := fmt.Sprintf("%s.%s", opts.ModelsPackage, typeName)
	schema.Extensions["x-go-type"] = goType
	schema.Extensions["x-go-type-import"] = map[string]string{
		"path": opts.ModelsPackage,
	}

	// Annotate properties with x-go-name if needed (for proper casing)
	if schema.Properties != nil {
		for propName, prop := range schema.Properties {
			a.annotateProperty(propName, &prop, opts)
			schema.Properties[propName] = prop
		}
	}
}

// annotateEnumSchema adds x-go-type annotation for enums
func (a *OpenAPIAnnotator) annotateEnumSchema(enumName string, schema *OpenAPISchema, opts *OpenAPIAnnotatorOptions) {
	if schema.Extensions == nil {
		schema.Extensions = make(map[string]interface{})
	}

	// Add x-go-type with import path
	goType := fmt.Sprintf("%s.%s", opts.ModelsPackage, enumName)
	schema.Extensions["x-go-type"] = goType
	schema.Extensions["x-go-type-import"] = map[string]string{
		"path": opts.ModelsPackage,
	}
}

// annotateUnionSchema adds x-go-type annotation for unions (interfaces)
func (a *OpenAPIAnnotator) annotateUnionSchema(unionName string, schema *OpenAPISchema, union *ast.Union, opts *OpenAPIAnnotatorOptions) {
	if schema.Extensions == nil {
		schema.Extensions = make(map[string]interface{})
	}

	// Unions map to Go interfaces
	goType := fmt.Sprintf("%s.%s", opts.ModelsPackage, unionName)
	schema.Extensions["x-go-type"] = goType
	schema.Extensions["x-go-type-import"] = map[string]string{
		"path": opts.ModelsPackage,
	}
}

// annotateProperty adds x-go-name annotation to properties if needed
func (a *OpenAPIAnnotator) annotateProperty(propName string, prop *OpenAPIProperty, opts *OpenAPIAnnotatorOptions) {
	// oapi-codegen uses the property name as-is, but we want proper Go casing
	// The Go generator already handles this, so we just need to mark it
	// Only add if the property references a custom type

	if prop.Extensions == nil {
		prop.Extensions = make(map[string]interface{})
	}

	// For custom type references, add x-go-type
	if prop.Ref != "" {
		// Extract type name from $ref (e.g., "#/components/schemas/User" -> "User")
		// This will be handled by the schema-level annotation
		return
	}

	// For array types with custom items
	if prop.Type == "array" && prop.Items != nil && prop.Items.Ref != "" {
		// The item type will be handled by its own schema annotation
		return
	}

	// For map types with custom values
	if prop.Type == "object" && prop.AdditionalProperties != nil && prop.AdditionalProperties.Ref != "" {
		// The value type will be handled by its own schema annotation
		return
	}
}

// GenerateJSON creates an annotated OpenAPI specification in JSON format
func (a *OpenAPIAnnotator) GenerateJSON(schema *ast.Schema, opts *OpenAPIAnnotatorOptions) (string, error) {
	yamlOutput, err := a.Generate(schema, opts)
	if err != nil {
		return "", err
	}

	// Parse YAML and convert to JSON
	var spec OpenAPISpec
	if err := yaml.Unmarshal([]byte(yamlOutput), &spec); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(jsonData), nil
}
