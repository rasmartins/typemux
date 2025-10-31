package annotations

// AnnotationMetadata describes a single annotation supported by TypeMUX
type AnnotationMetadata struct {
	// Name of the annotation (e.g., "@http.method", "@proto.name")
	Name string `json:"name"`

	// Scope indicates where this annotation can be used
	Scope []string `json:"scope"` // ["method", "field", "type", "enum", "union", "namespace", "schema"]

	// Formats indicates which output formats this annotation affects
	Formats []string `json:"formats"` // ["proto", "graphql", "openapi", "go", "all"]

	// Parameters describes the parameters this annotation accepts
	Parameters []ParameterMetadata `json:"parameters,omitempty"`

	// Description explains what this annotation does
	Description string `json:"description"`

	// Examples shows usage examples
	Examples []string `json:"examples,omitempty"`

	// Deprecated: indicates if this annotation is deprecated
	Deprecated bool `json:"deprecated,omitempty"`

	// DeprecatedMessage provides migration guidance if deprecated
	DeprecatedMessage string `json:"deprecatedMessage,omitempty"`
}

// ParameterMetadata describes a parameter for an annotation
type ParameterMetadata struct {
	// Name of the parameter
	Name string `json:"name"`

	// Type of the parameter (string, number, boolean, list, etc.)
	Type string `json:"type"`

	// Required indicates if this parameter is mandatory
	Required bool `json:"required"`

	// Description explains what this parameter does
	Description string `json:"description"`

	// ValidValues lists allowed values (for enums)
	ValidValues []string `json:"validValues,omitempty"`

	// Default value if not specified
	Default interface{} `json:"default,omitempty"`
}

// AnnotationRegistry collects all annotation metadata from all components
type AnnotationRegistry struct {
	annotations []*AnnotationMetadata
}

// NewAnnotationRegistry creates a new registry
func NewAnnotationRegistry() *AnnotationRegistry {
	return &AnnotationRegistry{
		annotations: make([]*AnnotationMetadata, 0),
	}
}

// Register adds an annotation to the registry
func (r *AnnotationRegistry) Register(meta *AnnotationMetadata) {
	r.annotations = append(r.annotations, meta)
}

// GetAll returns all registered annotations
func (r *AnnotationRegistry) GetAll() []*AnnotationMetadata {
	return r.annotations
}

// GetByScope returns annotations that can be used in a given scope
func (r *AnnotationRegistry) GetByScope(scope string) []*AnnotationMetadata {
	result := make([]*AnnotationMetadata, 0)
	for _, meta := range r.annotations {
		for _, s := range meta.Scope {
			if s == scope {
				result = append(result, meta)
				break
			}
		}
	}
	return result
}

// GetByFormat returns annotations that affect a given format
func (r *AnnotationRegistry) GetByFormat(format string) []*AnnotationMetadata {
	result := make([]*AnnotationMetadata, 0)
	for _, meta := range r.annotations {
		for _, f := range meta.Formats {
			if f == format || f == "all" {
				result = append(result, meta)
				break
			}
		}
	}
	return result
}
