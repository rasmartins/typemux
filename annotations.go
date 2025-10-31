package typemux

import (
	"github.com/rasmartins/typemux/internal/annotations"
)

// AnnotationMetadata describes a built-in TypeMUX annotation.
type AnnotationMetadata struct {
	// Name of the annotation (e.g., "@required", "@deprecated")
	Name string

	// Scope indicates where the annotation can be used
	// (e.g., "schema", "type", "field", "method")
	Scope []string

	// Formats indicates which output formats support this annotation
	// (e.g., "proto", "graphql", "openapi", "go", "all")
	Formats []string

	// Parameters describes the annotation's parameters
	Parameters []AnnotationParameter

	// Description is a human-readable description of what the annotation does
	Description string

	// Examples shows usage examples
	Examples []string
}

// AnnotationParameter describes a parameter for an annotation.
type AnnotationParameter struct {
	// Name of the parameter
	Name string

	// Type of the parameter (e.g., "string", "number", "boolean", "list", "object")
	Type string

	// Required indicates if the parameter is required
	Required bool

	// Description explains what the parameter does
	Description string

	// ValidValues lists valid values for enum-like parameters
	ValidValues []string
}

// GetBuiltinAnnotations returns metadata for all built-in TypeMUX annotations.
//
// Example:
//
//	annotations := typemux.GetBuiltinAnnotations()
//	for _, ann := range annotations {
//	    fmt.Printf("%s: %s\n", ann.Name, ann.Description)
//	}
func GetBuiltinAnnotations() []*AnnotationMetadata {
	registry := annotations.GetBuiltinAnnotations()
	allAnnotations := registry.GetAll()

	result := make([]*AnnotationMetadata, len(allAnnotations))
	for i, ann := range allAnnotations {
		result[i] = convertAnnotationMetadata(ann)
	}

	return result
}

// FilterAnnotationsByScope returns annotations that can be used in a specific scope.
//
// Example:
//
//	fieldAnnotations := typemux.FilterAnnotationsByScope("field")
//	for _, ann := range fieldAnnotations {
//	    fmt.Println(ann.Name)
//	}
func FilterAnnotationsByScope(scope string) []*AnnotationMetadata {
	registry := annotations.GetBuiltinAnnotations()
	filteredAnnotations := registry.GetByScope(scope)

	result := make([]*AnnotationMetadata, len(filteredAnnotations))
	for i, ann := range filteredAnnotations {
		result[i] = convertAnnotationMetadata(ann)
	}

	return result
}

// FilterAnnotationsByFormat returns annotations that apply to a specific output format.
//
// Example:
//
//	graphqlAnnotations := typemux.FilterAnnotationsByFormat("graphql")
//	for _, ann := range graphqlAnnotations {
//	    fmt.Println(ann.Name)
//	}
func FilterAnnotationsByFormat(format string) []*AnnotationMetadata {
	registry := annotations.GetBuiltinAnnotations()
	filteredAnnotations := registry.GetByFormat(format)

	result := make([]*AnnotationMetadata, len(filteredAnnotations))
	for i, ann := range filteredAnnotations {
		result[i] = convertAnnotationMetadata(ann)
	}

	return result
}

// GetAnnotation returns metadata for a specific annotation by name.
//
// Example:
//
//	ann, found := typemux.GetAnnotation("@required")
//	if found {
//	    fmt.Println(ann.Description)
//	}
func GetAnnotation(name string) (*AnnotationMetadata, bool) {
	registry := annotations.GetBuiltinAnnotations()
	allAnnotations := registry.GetAll()

	for _, ann := range allAnnotations {
		if ann.Name == name {
			return convertAnnotationMetadata(ann), true
		}
	}

	return nil, false
}

// convertAnnotationMetadata converts internal annotation metadata to public API type.
func convertAnnotationMetadata(ann *annotations.AnnotationMetadata) *AnnotationMetadata {
	result := &AnnotationMetadata{
		Name:        ann.Name,
		Scope:       ann.Scope,
		Formats:     ann.Formats,
		Description: ann.Description,
		Examples:    ann.Examples,
		Parameters:  make([]AnnotationParameter, len(ann.Parameters)),
	}

	for i, param := range ann.Parameters {
		result.Parameters[i] = AnnotationParameter{
			Name:        param.Name,
			Type:        param.Type,
			Required:    param.Required,
			Description: param.Description,
			ValidValues: param.ValidValues,
		}
	}

	return result
}
