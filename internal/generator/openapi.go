package generator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
	"gopkg.in/yaml.v3"
)

// OpenAPIGenerator generates OpenAPI 3.0 specifications from TypeMUX schemas.
type OpenAPIGenerator struct{}

// NewOpenAPIGenerator creates a new OpenAPI specification generator.
func NewOpenAPIGenerator() *OpenAPIGenerator {
	return &OpenAPIGenerator{}
}

// OpenAPISpec represents the root OpenAPI 3.0 specification structure.
type OpenAPISpec struct {
	OpenAPI    string                                 `json:"openapi" yaml:"openapi"`
	Info       OpenAPIInfo                            `json:"info" yaml:"info"`
	Paths      map[string]map[string]OpenAPIOperation `json:"paths" yaml:"paths"`
	Components OpenAPIComponents                      `json:"components" yaml:"components"`
}

// OpenAPIInfo contains metadata about the API.
type OpenAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// OpenAPIOperation describes a single API operation on a path.
type OpenAPIOperation struct {
	Summary     string                     `json:"summary" yaml:"summary"`
	OperationID string                     `json:"operationId" yaml:"operationId"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses" yaml:"responses"`
}

// OpenAPIParameter describes a single operation parameter.
type OpenAPIParameter struct {
	Name        string                 `json:"name" yaml:"name"`
	In          string                 `json:"in" yaml:"in"` // "path", "query", "header", "cookie"
	Required    bool                   `json:"required,omitempty" yaml:"required,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      OpenAPIParameterSchema `json:"schema" yaml:"schema"`
}

// OpenAPIParameterSchema describes the schema of a parameter.
type OpenAPIParameterSchema struct {
	Type string `json:"type" yaml:"type"`
}

// OpenAPIRequestBody describes a request body.
type OpenAPIRequestBody struct {
	Required bool                        `json:"required" yaml:"required"`
	Content  map[string]OpenAPIMediaType `json:"content" yaml:"content"`
}

// OpenAPIMediaType describes the media type of a request or response body.
type OpenAPIMediaType struct {
	Schema OpenAPISchemaRef `json:"schema" yaml:"schema"`
}

// OpenAPIResponse describes a single response from an API operation.
type OpenAPIResponse struct {
	Description string                      `json:"description" yaml:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// OpenAPISchemaRef is a reference to a schema or an inline schema definition.
type OpenAPISchemaRef struct {
	Ref        string                     `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type       string                     `json:"type,omitempty" yaml:"type,omitempty"`
	Properties map[string]OpenAPIProperty `json:"properties,omitempty" yaml:"properties,omitempty"`
}

// OpenAPIComponents holds reusable schema definitions.
type OpenAPIComponents struct {
	Schemas map[string]OpenAPISchema `json:"schemas" yaml:"schemas"`
}

// OpenAPIDiscriminator specifies the discriminator for polymorphic types.
type OpenAPIDiscriminator struct {
	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

// OpenAPISchema describes the structure of request/response bodies or schema components.
type OpenAPISchema struct {
	Type          string                     `json:"type,omitempty" yaml:"type,omitempty"`
	Description   string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Properties    map[string]OpenAPIProperty `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required      []string                   `json:"required,omitempty" yaml:"required,omitempty"`
	Enum          []string                   `json:"enum,omitempty" yaml:"enum,omitempty"`
	OneOf         []OpenAPISchemaRef         `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	Discriminator *OpenAPIDiscriminator      `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	Extensions    map[string]interface{}     `json:",inline" yaml:",inline"` // x- prefixed extensions
}

// OpenAPIProperty describes a property within a schema including validation constraints.
type OpenAPIProperty struct {
	Type                 string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Description          string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Ref                  string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Items                *OpenAPIPropertyItems  `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalProperties *OpenAPIPropertyItems  `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Default              interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	Enum                 []string               `json:"enum,omitempty" yaml:"enum,omitempty"`
	Deprecated           bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	MinLength            *int                   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength            *int                   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMinimum     *float64               `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *float64               `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	MultipleOf           *float64               `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	MinItems             *int                   `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems             *int                   `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	Extensions           map[string]interface{} `json:",inline" yaml:",inline"` // x- prefixed extensions
}

// OpenAPIPropertyItems describes the items of an array-type property or additionalProperties for maps.
type OpenAPIPropertyItems struct {
	Type                 string                `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string                `json:"format,omitempty" yaml:"format,omitempty"`
	Ref                  string                `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	AdditionalProperties *OpenAPIPropertyItems `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
}

// Generate creates an OpenAPI 3.0 YAML specification from the given schema.
func (g *OpenAPIGenerator) Generate(schema *ast.Schema) string {
	// Use namespace for title if available
	title := "Generated API"
	version := "1.0.0"
	description := ""

	if schema.Namespace != "" {
		title = schema.Namespace + " API"
	}

	// Apply namespace-level OpenAPI info from annotations
	if schema.NamespaceAnnotations != nil && len(schema.NamespaceAnnotations.OpenAPI) > 0 {
		for _, info := range schema.NamespaceAnnotations.OpenAPI {
			// Parse info string format: "key:value"
			parts := strings.SplitN(info, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Handle info section keys (non x- prefixed)
				if !strings.HasPrefix(key, "x-") {
					switch key {
					case "title":
						title = value
					case "version":
						version = value
					case "description":
						description = value
					}
				}
			}
		}
	}

	spec := OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       title,
			Version:     version,
			Description: description,
		},
		Paths: make(map[string]map[string]OpenAPIOperation),
		Components: OpenAPIComponents{
			Schemas: make(map[string]OpenAPISchema),
		},
	}

	// Build a map of original type names to their custom OpenAPI names
	typeNameMap := make(map[string]string)
	for _, typ := range schema.Types {
		if typ.Annotations != nil && typ.Annotations.OpenAPIName != "" {
			typeNameMap[typ.Name] = typ.Annotations.OpenAPIName
		}
	}

	// Generate schemas for enums
	for _, enum := range schema.Enums {
		enumValues := make([]string, len(enum.Values))
		for i, val := range enum.Values {
			enumValues[i] = val.Name
		}
		enumSchema := OpenAPISchema{
			Type: "string",
			Enum: enumValues,
		}
		if doc := enum.Doc.GetDoc("openapi"); doc != "" {
			enumSchema.Description = doc
		}
		spec.Components.Schemas[enum.Name] = enumSchema
	}

	// Generate schemas for types
	for _, typ := range schema.Types {
		// Use OpenAPIName override if specified
		schemaName := typ.Name
		if typ.Annotations != nil && typ.Annotations.OpenAPIName != "" {
			schemaName = typ.Annotations.OpenAPIName
		}
		spec.Components.Schemas[schemaName] = g.generateSchema(typ, typeNameMap)
	}

	// Generate schemas for unions
	for _, union := range schema.Unions {
		spec.Components.Schemas[union.Name] = g.generateUnionSchema(union)
	}

	// Generate paths from services
	for _, service := range schema.Services {
		for _, method := range service.Methods {
			g.addServiceMethod(&spec, service, method, typeNameMap)
		}
	}

	yamlBytes, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Sprintf("Error generating OpenAPI spec: %v", err)
	}

	return string(yamlBytes)
}

func (g *OpenAPIGenerator) generateSchema(typ *ast.Type, typeNameMap map[string]string) OpenAPISchema {
	schema := OpenAPISchema{
		Type:       "object",
		Properties: make(map[string]OpenAPIProperty),
		Required:   []string{},
		Extensions: make(map[string]interface{}),
	}

	if doc := typ.Doc.GetDoc("openapi"); doc != "" {
		schema.Description = doc
	}

	// Add OpenAPI extensions from type annotations
	if typ.Annotations != nil && len(typ.Annotations.OpenAPI) > 0 {
		for _, ext := range typ.Annotations.OpenAPI {
			extensions := g.parseExtensions(ext)
			for k, v := range extensions {
				schema.Extensions[k] = v
			}
		}
	}

	for _, field := range typ.Fields {
		// Skip excluded fields
		if !field.ShouldIncludeInGenerator("openapi") {
			continue
		}

		property := g.convertFieldToProperty(field, typeNameMap)
		schema.Properties[field.Name] = property

		// Fields are required if explicitly marked with @required annotation
		// Fields marked with ? are explicitly optional
		if field.Required && !field.Type.Optional {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	return schema
}

func (g *OpenAPIGenerator) generateUnionSchema(union *ast.Union) OpenAPISchema {
	schema := OpenAPISchema{
		OneOf: []OpenAPISchemaRef{},
	}

	if doc := union.Doc.GetDoc("openapi"); doc != "" {
		schema.Description = doc
	}

	// Add discriminator for better client generation
	// Uses "type" as the discriminator property
	discriminator := &OpenAPIDiscriminator{
		PropertyName: "type",
		Mapping:      make(map[string]string),
	}

	// Add each union option as a oneOf reference
	for _, option := range union.Options {
		schema.OneOf = append(schema.OneOf, OpenAPISchemaRef{
			Ref: fmt.Sprintf("#/components/schemas/%s", option),
		})
		// Map the type name to the schema reference
		discriminator.Mapping[option] = fmt.Sprintf("#/components/schemas/%s", option)
	}

	schema.Discriminator = discriminator

	return schema
}

// generateMapDescription creates a human-readable description for map types
func (g *OpenAPIGenerator) generateMapDescription(fieldType *ast.FieldType) string {
	if !fieldType.IsMap {
		return ""
	}

	valueFieldType := fieldType.GetMapValueType()
	if valueFieldType == nil {
		return fmt.Sprintf("Map of %s to unknown", fieldType.MapKey)
	}

	var valueDesc string
	if valueFieldType.IsMap {
		// Recursively describe nested maps
		valueDesc = g.generateMapDescription(valueFieldType)
	} else {
		valueDesc = valueFieldType.Name
	}

	return fmt.Sprintf("Map of %s to %s", fieldType.MapKey, valueDesc)
}

// generateAdditionalProperties recursively generates OpenAPI additionalProperties for map value types
func (g *OpenAPIGenerator) generateAdditionalProperties(valueFieldType *ast.FieldType, typeNameMap map[string]string) *OpenAPIPropertyItems {
	if valueFieldType.IsMap {
		// Nested map case: recursively generate additionalProperties structure
		// Example: map<string, map<string, int32>> becomes:
		// additionalProperties:
		//   type: object
		//   additionalProperties:
		//     type: integer
		//     format: int32
		nestedValueType := valueFieldType.GetMapValueType()
		if nestedValueType != nil {
			return &OpenAPIPropertyItems{
				Type:                 "object",
				AdditionalProperties: g.generateAdditionalProperties(nestedValueType, typeNameMap),
			}
		}
		// Fallback if nested value type is unknown
		return &OpenAPIPropertyItems{
			Type: "object",
		}
	}

	// Non-map case: simple type or reference
	valueType := g.mapTypeToOpenAPI(valueFieldType.Name)
	valueFormat := g.getFormatForType(valueFieldType.Name)

	additionalProps := &OpenAPIPropertyItems{
		Type:   valueType,
		Format: valueFormat,
	}

	// If the value is a custom type, use a reference
	if !ast.IsBuiltinType(valueFieldType.Name) {
		unqualifiedName := ast.GetUnqualifiedName(valueFieldType.Name)
		schemaName := unqualifiedName
		if customName, ok := typeNameMap[unqualifiedName]; ok {
			schemaName = customName
		}
		additionalProps.Ref = fmt.Sprintf("#/components/schemas/%s", schemaName)
		additionalProps.Type = ""   // Clear type when using ref
		additionalProps.Format = "" // Clear format when using ref
	}

	return additionalProps
}

func (g *OpenAPIGenerator) convertFieldToProperty(field *ast.Field, typeNameMap map[string]string) OpenAPIProperty {
	property := OpenAPIProperty{
		Extensions: make(map[string]interface{}),
	}

	// Add field documentation
	if doc := field.Doc.GetDoc("openapi"); doc != "" {
		property.Description = doc
	}

	// Add deprecation
	if field.Deprecated != nil {
		property.Deprecated = true
		// Add deprecation info to description
		if property.Description != "" {
			property.Description += "\n\n"
		}
		property.Description += "**DEPRECATED**"
		if field.Deprecated.Since != "" {
			property.Description += fmt.Sprintf(" (since %s)", field.Deprecated.Since)
		}
		if field.Deprecated.Removed != "" {
			property.Description += fmt.Sprintf(" - will be removed in %s", field.Deprecated.Removed)
		}
		if field.Deprecated.Reason != "" {
			property.Description += fmt.Sprintf(": %s", field.Deprecated.Reason)
		}
	}

	// Add validation rules
	if field.Validation != nil {
		if field.Validation.MinLength != nil {
			property.MinLength = field.Validation.MinLength
		}
		if field.Validation.MaxLength != nil {
			property.MaxLength = field.Validation.MaxLength
		}
		if field.Validation.Pattern != "" {
			property.Pattern = field.Validation.Pattern
		}
		if field.Validation.Format != "" {
			property.Format = field.Validation.Format
		}
		if field.Validation.Min != nil {
			property.Minimum = field.Validation.Min
		}
		if field.Validation.Max != nil {
			property.Maximum = field.Validation.Max
		}
		if field.Validation.ExclusiveMin != nil {
			property.ExclusiveMinimum = field.Validation.ExclusiveMin
		}
		if field.Validation.ExclusiveMax != nil {
			property.ExclusiveMaximum = field.Validation.ExclusiveMax
		}
		if field.Validation.MultipleOf != nil {
			property.MultipleOf = field.Validation.MultipleOf
		}
		if field.Validation.MinItems != nil {
			property.MinItems = field.Validation.MinItems
		}
		if field.Validation.MaxItems != nil {
			property.MaxItems = field.Validation.MaxItems
		}
		if field.Validation.UniqueItems {
			property.UniqueItems = true
		}
		if len(field.Validation.Enum) > 0 {
			property.Enum = field.Validation.Enum
		}
	}

	// Add OpenAPI extensions from field annotations
	if field.Annotations != nil && len(field.Annotations.OpenAPI) > 0 {
		for _, ext := range field.Annotations.OpenAPI {
			extensions := g.parseExtensions(ext)
			for k, v := range extensions {
				property.Extensions[k] = v
			}
		}
	}

	if field.Type.IsMap {
		property.Type = "object"

		// Get the map value type using the new API
		valueFieldType := field.Type.GetMapValueType()
		if valueFieldType == nil {
			property.Description = fmt.Sprintf("Map of %s to unknown", field.Type.MapKey)
			return property
		}

		property.Description = g.generateMapDescription(field.Type)

		// Use additionalProperties to specify the value type
		additionalProps := g.generateAdditionalProperties(valueFieldType, typeNameMap)
		property.AdditionalProperties = additionalProps
		return property
	}

	if field.Type.IsArray {
		property.Type = "array"
		property.Items = &OpenAPIPropertyItems{}

		baseType := g.mapTypeToOpenAPI(field.Type.Name)
		if baseType == "object" || !ast.IsBuiltinType(field.Type.Name) {
			// Use unqualified name for schema reference lookup
			unqualifiedName := ast.GetUnqualifiedName(field.Type.Name)
			// Check if this type has a custom OpenAPI name
			schemaName := unqualifiedName
			if customName, ok := typeNameMap[unqualifiedName]; ok {
				schemaName = customName
			}
			property.Items.Ref = fmt.Sprintf("#/components/schemas/%s", schemaName)
		} else {
			property.Items.Type = baseType
		}
		return property
	}

	// Scalar or custom type
	if ast.IsBuiltinType(field.Type.Name) {
		oaType := g.mapTypeToOpenAPI(field.Type.Name)
		property.Type = oaType
		if format := g.getFormatForType(field.Type.Name); format != "" {
			property.Format = format
		}

		// Set properly typed default values
		if field.Default != "" {
			property.Default = g.convertDefaultValue(field.Default, field.Type.Name)
		}
	} else {
		// Reference to custom type - only set Ref, no other fields
		// Use unqualified name for schema reference lookup
		unqualifiedName := ast.GetUnqualifiedName(field.Type.Name)
		// Check if this type has a custom OpenAPI name
		schemaName := unqualifiedName
		if customName, ok := typeNameMap[unqualifiedName]; ok {
			schemaName = customName
		}
		property.Ref = fmt.Sprintf("#/components/schemas/%s", schemaName)
		return property
	}

	return property
}

func (g *OpenAPIGenerator) mapTypeToOpenAPI(typeName string) string {
	typeMap := map[string]string{
		"string":    "string",
		"int32":     "integer",
		"int64":     "integer",
		"float32":   "number",
		"float64":   "number",
		"bool":      "boolean",
		"timestamp": "string",
		"bytes":     "string",
	}

	if oaType, ok := typeMap[typeName]; ok {
		return oaType
	}

	return "object"
}

func (g *OpenAPIGenerator) getFormatForType(typeName string) string {
	formatMap := map[string]string{
		"int32":     "int32",
		"int64":     "int64",
		"float32":   "float",
		"float64":   "double",
		"timestamp": "date-time",
		"bytes":     "byte",
	}

	return formatMap[typeName]
}

func (g *OpenAPIGenerator) convertDefaultValue(defaultStr string, typeName string) interface{} {
	// Convert string default values to proper types for YAML/JSON
	switch typeName {
	case "int32", "int64":
		// Parse as integer
		var val int64
		if _, err := fmt.Sscanf(defaultStr, "%d", &val); err == nil {
			return val
		}
		return defaultStr
	case "float32", "float64":
		// Parse as float
		var val float64
		if _, err := fmt.Sscanf(defaultStr, "%f", &val); err == nil {
			return val
		}
		return defaultStr
	case "bool":
		// Parse as boolean
		return defaultStr == "true"
	default:
		// Keep as string for other types
		return defaultStr
	}
}

func (g *OpenAPIGenerator) addServiceMethod(spec *OpenAPISpec, service *ast.Service, method *ast.Method, typeNameMap map[string]string) {
	// Use custom path template if provided, otherwise generate from service/method name
	var path string
	if method.PathTemplate != "" {
		path = method.PathTemplate
	} else {
		path = fmt.Sprintf("/%s/%s", strings.ToLower(service.Name), strings.ToLower(method.Name))
	}

	// Use GetHTTPMethod which checks annotation or uses heuristics
	httpMethod := method.GetHTTPMethod()

	operation := OpenAPIOperation{
		Summary:     fmt.Sprintf("%s operation", method.Name),
		OperationID: method.Name,
		Responses:   make(map[string]OpenAPIResponse),
	}

	// Extract and add path parameters
	pathParams := g.extractPathParameters(path)
	if len(pathParams) > 0 {
		operation.Parameters = pathParams
	}

	// Resolve input type name (check for custom name)
	inputTypeName := method.InputType
	if customName, ok := typeNameMap[method.InputType]; ok {
		inputTypeName = customName
	}

	// Resolve output type name (check for custom name)
	outputTypeName := method.OutputType
	if customName, ok := typeNameMap[method.OutputType]; ok {
		outputTypeName = customName
	}

	// Add request body for POST/PUT/PATCH methods
	if httpMethod == "post" || httpMethod == "put" || httpMethod == "patch" {
		operation.RequestBody = &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: OpenAPISchemaRef{
						Ref: fmt.Sprintf("#/components/schemas/%s", inputTypeName),
					},
				},
			},
		}
	}

	// Add default 200 response
	operation.Responses["200"] = OpenAPIResponse{
		Description: "Successful response",
		Content: map[string]OpenAPIMediaType{
			"application/json": {
				Schema: OpenAPISchemaRef{
					Ref: fmt.Sprintf("#/components/schemas/%s", outputTypeName),
				},
			},
		},
	}

	// Add additional success responses
	for _, code := range method.SuccessCodes {
		operation.Responses[code] = OpenAPIResponse{
			Description: g.getSuccessDescription(code),
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: OpenAPISchemaRef{
						Ref: fmt.Sprintf("#/components/schemas/%s", outputTypeName),
					},
				},
			},
		}
	}

	// Add error responses
	for _, code := range method.ErrorCodes {
		operation.Responses[code] = OpenAPIResponse{
			Description: g.getErrorDescription(code),
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: OpenAPISchemaRef{
						Type: "object",
						Properties: map[string]OpenAPIProperty{
							"error": {
								Type:        "string",
								Description: "Error message",
							},
							"code": {
								Type:        "string",
								Description: "Error code",
							},
						},
					},
				},
			},
		}
	}

	if spec.Paths[path] == nil {
		spec.Paths[path] = make(map[string]OpenAPIOperation)
	}
	spec.Paths[path][httpMethod] = operation
}

// getSuccessDescription returns a description for common HTTP success codes
func (g *OpenAPIGenerator) getSuccessDescription(code string) string {
	descriptions := map[string]string{
		"200": "OK - Successful response",
		"201": "Created - Resource created successfully",
		"202": "Accepted - Request accepted for processing",
		"204": "No Content - Successful request with no response body",
		"206": "Partial Content - Partial resource returned",
	}

	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return fmt.Sprintf("Success response (%s)", code)
}

// getErrorDescription returns a description for common HTTP error codes
func (g *OpenAPIGenerator) getErrorDescription(code string) string {
	descriptions := map[string]string{
		"400": "Bad Request - Invalid input parameters",
		"401": "Unauthorized - Authentication required",
		"403": "Forbidden - Insufficient permissions",
		"404": "Not Found - Resource not found",
		"409": "Conflict - Resource already exists or conflict",
		"422": "Unprocessable Entity - Validation error",
		"429": "Too Many Requests - Rate limit exceeded",
		"500": "Internal Server Error",
		"502": "Bad Gateway",
		"503": "Service Unavailable",
		"504": "Gateway Timeout",
	}

	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return fmt.Sprintf("Error response (%s)", code)
}

func (g *OpenAPIGenerator) extractPathParameters(path string) []OpenAPIParameter {
	var params []OpenAPIParameter

	// Find all {paramName} patterns in the path
	start := -1
	for i := 0; i < len(path); i++ {
		if path[i] == '{' {
			start = i + 1
		} else if path[i] == '}' && start != -1 {
			paramName := path[start:i]
			params = append(params, OpenAPIParameter{
				Name:     paramName,
				In:       "path",
				Required: true,
				Schema: OpenAPIParameterSchema{
					Type: "string",
				},
			})
			start = -1
		}
	}

	return params
}

// parseExtensions parses a JSON string into a map of extensions
// Supports both JSON objects: {"x-custom": "value", "x-another": 123}
// The function expects valid JSON format
func (g *OpenAPIGenerator) parseExtensions(extJSON string) map[string]interface{} {
	extensions := make(map[string]interface{})

	// Try to parse as JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(extJSON), &jsonData)
	if err != nil {
		// If JSON parsing fails, return empty map (could log error in production)
		return extensions
	}

	// Copy all fields to extensions map
	for k, v := range jsonData {
		extensions[k] = v
	}

	return extensions
}
