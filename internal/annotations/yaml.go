package annotations

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLAnnotations represents the root structure of a YAML annotations file
type YAMLAnnotations struct {
	Types    map[string]*TypeAnnotations    `yaml:"types"`
	Enums    map[string]*EnumAnnotations    `yaml:"enums"`
	Unions   map[string]*UnionAnnotations   `yaml:"unions"`
	Services map[string]*ServiceAnnotations `yaml:"services"`
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
	Required bool                       `yaml:"required"`
	Default  string                     `yaml:"default"`
	Exclude  []string                   `yaml:"exclude"`
	Only     []string                   `yaml:"only"`
	Proto    *FormatSpecificAnnotations `yaml:"proto"`
	GraphQL  *FormatSpecificAnnotations `yaml:"graphql"`
	OpenAPI  *FormatSpecificAnnotations `yaml:"openapi"`
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

// MergeYAMLAnnotations merges multiple YAML annotation files
// Later files override earlier ones
func MergeYAMLAnnotations(files []string) (*YAMLAnnotations, error) {
	result := &YAMLAnnotations{
		Types:    make(map[string]*TypeAnnotations),
		Enums:    make(map[string]*EnumAnnotations),
		Unions:   make(map[string]*UnionAnnotations),
		Services: make(map[string]*ServiceAnnotations),
	}

	for _, file := range files {
		annotations, err := LoadYAMLAnnotations(file)
		if err != nil {
			return nil, err
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
