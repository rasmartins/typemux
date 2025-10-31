package annotations

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLAnnotations represents the root structure of a YAML annotations file
type YAMLAnnotations struct {
	Version    string                           `yaml:"version"`
	Namespaces map[string]*NamespaceAnnotations `yaml:"namespaces"`
	Types      map[string]*TypeAnnotations      `yaml:"types"`
	Enums      map[string]*EnumAnnotations      `yaml:"enums"`
	Unions     map[string]*UnionAnnotations     `yaml:"unions"`
	Services   map[string]*ServiceAnnotations   `yaml:"services"`
}

// NamespaceAnnotations represents annotations for a namespace
type NamespaceAnnotations struct {
	Proto   *NamespaceProtoAnnotations   `yaml:"proto"`
	GraphQL *FormatSpecificAnnotations   `yaml:"graphql"`
	OpenAPI *NamespaceOpenAPIAnnotations `yaml:"openapi"`
}

// NamespaceProtoAnnotations represents protobuf-specific namespace annotations
type NamespaceProtoAnnotations struct {
	Options map[string]string `yaml:"options"` // e.g., go_package, java_package, etc.
}

// NamespaceOpenAPIAnnotations represents OpenAPI-specific namespace annotations
type NamespaceOpenAPIAnnotations struct {
	Info       map[string]string `yaml:"info"`       // OpenAPI info section
	Extensions map[string]string `yaml:"extensions"` // OpenAPI extensions (x-*)
}

// DeprecationAnnotations represents deprecation information for fields/types
type DeprecationAnnotations struct {
	Reason  string `yaml:"reason"`  // Why it's deprecated and what to use instead
	Since   string `yaml:"since"`   // Version when it was deprecated
	Removed string `yaml:"removed"` // Version when it will be removed (optional)
}

// ValidationAnnotations represents validation constraints for fields
type ValidationAnnotations struct {
	// String validation
	MinLength *int   `yaml:"minLength,omitempty"`
	MaxLength *int   `yaml:"maxLength,omitempty"`
	Pattern   string `yaml:"pattern,omitempty"` // Regex pattern
	Format    string `yaml:"format,omitempty"`  // email, url, uuid, etc.

	// Numeric validation
	Min          *float64 `yaml:"min,omitempty"`          // Minimum value (inclusive)
	Max          *float64 `yaml:"max,omitempty"`          // Maximum value (inclusive)
	ExclusiveMin *float64 `yaml:"exclusiveMin,omitempty"` // Minimum value (exclusive)
	ExclusiveMax *float64 `yaml:"exclusiveMax,omitempty"` // Maximum value (exclusive)
	MultipleOf   *float64 `yaml:"multipleOf,omitempty"`   // Must be multiple of this value

	// Array validation
	MinItems    *int `yaml:"minItems,omitempty"`
	MaxItems    *int `yaml:"maxItems,omitempty"`
	UniqueItems bool `yaml:"uniqueItems,omitempty"`

	// General
	Enum []string `yaml:"enum,omitempty"` // Allowed values
}

// FormatSpecificAnnotations represents annotations for a specific format (proto/graphql/openapi)
type FormatSpecificAnnotations struct {
	Name      string `yaml:"name"`
	Option    string `yaml:"option"`
	Directive string `yaml:"directive"`
	Extension string `yaml:"extension"`
}

// TypeAnnotations represents annotations for a type
type TypeAnnotations struct {
	Proto   *FormatSpecificAnnotations   `yaml:"proto"`
	GraphQL *FormatSpecificAnnotations   `yaml:"graphql"`
	OpenAPI *FormatSpecificAnnotations   `yaml:"openapi"`
	Fields  map[string]*FieldAnnotations `yaml:"fields"`
}

// FieldAnnotations represents annotations for a field
type FieldAnnotations struct {
	Required   bool                       `yaml:"required"`
	Default    string                     `yaml:"default"`
	Exclude    []string                   `yaml:"exclude"`
	Only       []string                   `yaml:"only"`
	Proto      *FormatSpecificAnnotations `yaml:"proto"`
	GraphQL    *FormatSpecificAnnotations `yaml:"graphql"`
	OpenAPI    *FormatSpecificAnnotations `yaml:"openapi"`
	Deprecated *DeprecationAnnotations    `yaml:"deprecated"`
	Validation *ValidationAnnotations     `yaml:"validation"`
	Since      string                     `yaml:"since"`
}

// EnumAnnotations represents annotations for an enum
type EnumAnnotations struct {
	Proto   *FormatSpecificAnnotations `yaml:"proto"`
	GraphQL *FormatSpecificAnnotations `yaml:"graphql"`
	OpenAPI *FormatSpecificAnnotations `yaml:"openapi"`
}

// UnionAnnotations represents annotations for a union
type UnionAnnotations struct {
	Proto   *FormatSpecificAnnotations `yaml:"proto"`
	GraphQL *FormatSpecificAnnotations `yaml:"graphql"`
	OpenAPI *FormatSpecificAnnotations `yaml:"openapi"`
}

// ServiceAnnotations represents annotations for a service
type ServiceAnnotations struct {
	Proto   *FormatSpecificAnnotations    `yaml:"proto"`
	GraphQL *FormatSpecificAnnotations    `yaml:"graphql"`
	OpenAPI *FormatSpecificAnnotations    `yaml:"openapi"`
	Methods map[string]*MethodAnnotations `yaml:"methods"`
}

// MethodAnnotations represents annotations for an RPC method
type MethodAnnotations struct {
	HTTP    string                     `yaml:"http"`
	Path    string                     `yaml:"path"`
	GraphQL string                     `yaml:"graphql"`
	Success []int                      `yaml:"success"`
	Errors  []int                      `yaml:"errors"`
	Proto   *FormatSpecificAnnotations `yaml:"proto"`
}

// LoadYAMLAnnotations loads annotations from a YAML file
func LoadYAMLAnnotations(filepath string) (*YAMLAnnotations, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read annotations file %s: %w", filepath, err)
	}

	var annotations YAMLAnnotations
	if err := yaml.Unmarshal(data, &annotations); err != nil {
		return nil, fmt.Errorf("failed to parse YAML annotations file %s: %w", filepath, err)
	}

	return &annotations, nil
}

// ParseYAMLAnnotations parses YAML annotations from a string content.
func ParseYAMLAnnotations(content string) (*YAMLAnnotations, error) {
	var annotations YAMLAnnotations
	if err := yaml.Unmarshal([]byte(content), &annotations); err != nil {
		return nil, fmt.Errorf("failed to parse YAML annotations: %w", err)
	}

	return &annotations, nil
}

// MergeYAMLAnnotations merges multiple YAML annotation files
// Later files override earlier ones
func MergeYAMLAnnotations(files []string) (*YAMLAnnotations, error) {
	result := &YAMLAnnotations{
		Namespaces: make(map[string]*NamespaceAnnotations),
		Types:      make(map[string]*TypeAnnotations),
		Enums:      make(map[string]*EnumAnnotations),
		Unions:     make(map[string]*UnionAnnotations),
		Services:   make(map[string]*ServiceAnnotations),
	}

	for _, file := range files {
		annotations, err := LoadYAMLAnnotations(file)
		if err != nil {
			return nil, err
		}

		// Merge namespaces
		for namespaceName, namespaceAnnotations := range annotations.Namespaces {
			result.Namespaces[namespaceName] = namespaceAnnotations
		}

		// Merge types
		for typeName, typeAnnotations := range annotations.Types {
			if result.Types[typeName] == nil {
				result.Types[typeName] = typeAnnotations
			} else {
				// Merge type-level annotations (later overrides)
				result.Types[typeName].Proto = mergeFormatAnnotations(result.Types[typeName].Proto, typeAnnotations.Proto)
				result.Types[typeName].GraphQL = mergeFormatAnnotations(result.Types[typeName].GraphQL, typeAnnotations.GraphQL)
				result.Types[typeName].OpenAPI = mergeFormatAnnotations(result.Types[typeName].OpenAPI, typeAnnotations.OpenAPI)

				// Merge fields
				if result.Types[typeName].Fields == nil {
					result.Types[typeName].Fields = make(map[string]*FieldAnnotations)
				}
				for fieldName, fieldAnnotations := range typeAnnotations.Fields {
					if result.Types[typeName].Fields[fieldName] == nil {
						result.Types[typeName].Fields[fieldName] = fieldAnnotations
					} else {
						// Merge field-level annotations
						existingField := result.Types[typeName].Fields[fieldName]
						if fieldAnnotations.Required {
							existingField.Required = true
						}
						if fieldAnnotations.Default != "" {
							existingField.Default = fieldAnnotations.Default
						}
						existingField.Proto = mergeFormatAnnotations(existingField.Proto, fieldAnnotations.Proto)
						existingField.GraphQL = mergeFormatAnnotations(existingField.GraphQL, fieldAnnotations.GraphQL)
						existingField.OpenAPI = mergeFormatAnnotations(existingField.OpenAPI, fieldAnnotations.OpenAPI)
						// Merge lists
						existingField.Exclude = mergeStringLists(existingField.Exclude, fieldAnnotations.Exclude)
						existingField.Only = mergeStringLists(existingField.Only, fieldAnnotations.Only)
					}
				}
			}
		}

		// Merge enums
		for enumName, enumAnnotations := range annotations.Enums {
			result.Enums[enumName] = enumAnnotations
		}

		// Merge unions
		for unionName, unionAnnotations := range annotations.Unions {
			result.Unions[unionName] = unionAnnotations
		}

		// Merge services
		for serviceName, serviceAnnotations := range annotations.Services {
			if result.Services[serviceName] == nil {
				result.Services[serviceName] = serviceAnnotations
			} else {
				// Merge service-level annotations
				result.Services[serviceName].Proto = mergeFormatAnnotations(result.Services[serviceName].Proto, serviceAnnotations.Proto)
				result.Services[serviceName].GraphQL = mergeFormatAnnotations(result.Services[serviceName].GraphQL, serviceAnnotations.GraphQL)
				result.Services[serviceName].OpenAPI = mergeFormatAnnotations(result.Services[serviceName].OpenAPI, serviceAnnotations.OpenAPI)

				// Merge methods
				if result.Services[serviceName].Methods == nil {
					result.Services[serviceName].Methods = make(map[string]*MethodAnnotations)
				}
				for methodName, methodAnnotations := range serviceAnnotations.Methods {
					result.Services[serviceName].Methods[methodName] = methodAnnotations
				}
			}
		}
	}

	return result, nil
}

// MergeYAMLAnnotationsFromContent merges multiple YAML annotation strings.
// Later annotations override earlier ones. This is similar to MergeYAMLAnnotations
// but takes content strings instead of file paths.
func MergeYAMLAnnotationsFromContent(contents []string) (*YAMLAnnotations, error) {
	result := &YAMLAnnotations{
		Namespaces: make(map[string]*NamespaceAnnotations),
		Types:      make(map[string]*TypeAnnotations),
		Enums:      make(map[string]*EnumAnnotations),
		Unions:     make(map[string]*UnionAnnotations),
		Services:   make(map[string]*ServiceAnnotations),
	}

	for _, content := range contents {
		annotations, err := ParseYAMLAnnotations(content)
		if err != nil {
			return nil, err
		}

		// Merge namespaces
		for namespaceName, namespaceAnnotations := range annotations.Namespaces {
			result.Namespaces[namespaceName] = namespaceAnnotations
		}

		// Merge types
		for typeName, typeAnnotations := range annotations.Types {
			if result.Types[typeName] == nil {
				result.Types[typeName] = typeAnnotations
			} else {
				// Merge type-level annotations (later overrides)
				result.Types[typeName].Proto = mergeFormatAnnotations(result.Types[typeName].Proto, typeAnnotations.Proto)
				result.Types[typeName].GraphQL = mergeFormatAnnotations(result.Types[typeName].GraphQL, typeAnnotations.GraphQL)
				result.Types[typeName].OpenAPI = mergeFormatAnnotations(result.Types[typeName].OpenAPI, typeAnnotations.OpenAPI)

				// Merge fields
				if result.Types[typeName].Fields == nil {
					result.Types[typeName].Fields = make(map[string]*FieldAnnotations)
				}
				for fieldName, fieldAnnotations := range typeAnnotations.Fields {
					if result.Types[typeName].Fields[fieldName] == nil {
						result.Types[typeName].Fields[fieldName] = fieldAnnotations
					} else {
						// Merge field-level annotations
						existingField := result.Types[typeName].Fields[fieldName]
						if fieldAnnotations.Required {
							existingField.Required = true
						}
						if fieldAnnotations.Default != "" {
							existingField.Default = fieldAnnotations.Default
						}
						existingField.Proto = mergeFormatAnnotations(existingField.Proto, fieldAnnotations.Proto)
						existingField.GraphQL = mergeFormatAnnotations(existingField.GraphQL, fieldAnnotations.GraphQL)
						existingField.OpenAPI = mergeFormatAnnotations(existingField.OpenAPI, fieldAnnotations.OpenAPI)
					}
				}
			}
		}

		// Merge enums
		for enumName, enumAnnotations := range annotations.Enums {
			result.Enums[enumName] = enumAnnotations
		}

		// Merge unions
		for unionName, unionAnnotations := range annotations.Unions {
			result.Unions[unionName] = unionAnnotations
		}

		// Merge services
		for serviceName, serviceAnnotations := range annotations.Services {
			if result.Services[serviceName] == nil {
				result.Services[serviceName] = serviceAnnotations
			} else {
				// Merge service-level annotations
				result.Services[serviceName].Proto = mergeFormatAnnotations(result.Services[serviceName].Proto, serviceAnnotations.Proto)
				result.Services[serviceName].GraphQL = mergeFormatAnnotations(result.Services[serviceName].GraphQL, serviceAnnotations.GraphQL)
				result.Services[serviceName].OpenAPI = mergeFormatAnnotations(result.Services[serviceName].OpenAPI, serviceAnnotations.OpenAPI)

				// Merge methods
				if result.Services[serviceName].Methods == nil {
					result.Services[serviceName].Methods = make(map[string]*MethodAnnotations)
				}
				for methodName, methodAnnotations := range serviceAnnotations.Methods {
					result.Services[serviceName].Methods[methodName] = methodAnnotations
				}
			}
		}
	}

	return result, nil
}

// mergeFormatAnnotations merges two FormatSpecificAnnotations, with b taking precedence
func mergeFormatAnnotations(a, b *FormatSpecificAnnotations) *FormatSpecificAnnotations {
	if b == nil {
		return a
	}
	if a == nil {
		return b
	}

	result := &FormatSpecificAnnotations{
		Name:      a.Name,
		Option:    a.Option,
		Directive: a.Directive,
		Extension: a.Extension,
	}

	if b.Name != "" {
		result.Name = b.Name
	}
	if b.Option != "" {
		result.Option = b.Option
	}
	if b.Directive != "" {
		result.Directive = b.Directive
	}
	if b.Extension != "" {
		result.Extension = b.Extension
	}

	return result
}

// mergeStringLists merges two string lists, removing duplicates
func mergeStringLists(a, b []string) []string {
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
