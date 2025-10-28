package ast

import "strings"

// Schema represents the entire IDL schema
type Schema struct {
	Namespace            string             // Optional namespace (e.g., "com.example.api")
	TypeMUXVersion       string             // TypeMUX IDL format version (e.g., "1.0.0")
	Version              string             // Schema version (e.g., "1.0.0", "2.1.3")
	NamespaceAnnotations *FormatAnnotations // Namespace-level annotations
	Imports              []string           // Imported file paths
	Enums                []*Enum
	Types                []*Type
	Unions               []*Union
	Services             []*Service
	TypeRegistry         *TypeRegistry // Registry for resolving qualified type names
}

// Enum represents an enumeration type
type Enum struct {
	Name        string
	Namespace   string // Namespace this enum belongs to
	Values      []*EnumValue
	Doc         *Documentation
	Annotations *FormatAnnotations // Format-specific annotations
}

// EnumValue represents a single enum value with optional number
type EnumValue struct {
	Name      string
	Number    int  // Protobuf field number
	HasNumber bool // Whether a custom number was specified
	Doc       *Documentation
}

// Type represents a data type definition
type Type struct {
	Name        string
	Namespace   string // Namespace this type belongs to
	Fields      []*Field
	Doc         *Documentation
	Annotations *FormatAnnotations // Format-specific annotations
}

// Union represents a union/oneOf type (can be one of several types)
type Union struct {
	Name        string
	Namespace   string   // Namespace this union belongs to
	Options     []string // Names of the types that can be in this union
	Doc         *Documentation
	Annotations *FormatAnnotations // Format-specific annotations
}

// Field represents a field in a type
type Field struct {
	Name        string
	Type        *FieldType
	Required    bool
	Default     string
	Attributes  map[string]string
	Doc         *Documentation
	ExcludeFrom []string           // List of generators to exclude this field from
	OnlyFor     []string           // If set, only include in these generators
	Number      int                // Protobuf field number
	HasNumber   bool               // Whether a custom number was specified
	Annotations *FormatAnnotations // Format-specific annotations
	Deprecated  *DeprecationInfo   // Deprecation information
	Validation  *ValidationRules   // Validation rules
	Since       string             // Version when this field was added (e.g., "2.0.0")
}

// ShouldIncludeInGenerator checks if a field should be included in a specific generator
func (f *Field) ShouldIncludeInGenerator(generator string) bool {
	// If OnlyFor is specified, only include if generator is in the list
	if len(f.OnlyFor) > 0 {
		for _, g := range f.OnlyFor {
			if g == generator {
				return true
			}
		}
		return false
	}

	// If ExcludeFrom is specified, exclude if generator is in the list
	for _, g := range f.ExcludeFrom {
		if g == generator {
			return false
		}
	}

	return true
}

// FieldType represents the type of a field
type FieldType struct {
	Name         string // base type name (set to "map" for map types)
	IsArray      bool
	IsMap        bool
	MapKey       string     // for map types - the key type (must be string or int)
	MapValue     string     // for simple map value types (deprecated - use MapValueType for new code)
	MapValueType *FieldType // for complex map value types (supports nested maps, arrays, etc.)
	IsBuiltin    bool
	Optional     bool // true if the type has a ? suffix (e.g., string?)
}

// GetMapValueType returns the map value type, supporting both simple string values and complex FieldType values
func (ft *FieldType) GetMapValueType() *FieldType {
	if ft.MapValueType != nil {
		return ft.MapValueType
	}
	// Fallback to simple string-based MapValue for backward compatibility
	if ft.MapValue != "" {
		return &FieldType{
			Name:      ft.MapValue,
			IsBuiltin: IsBuiltinType(ft.MapValue),
		}
	}
	return nil
}

// GetMapValueTypeName returns the type name for simple cases (backward compatibility)
func (ft *FieldType) GetMapValueTypeName() string {
	if ft.MapValueType != nil {
		return ft.MapValueType.Name
	}
	return ft.MapValue
}

// Service represents a service definition
type Service struct {
	Name        string
	Namespace   string // Namespace this service belongs to
	Methods     []*Method
	Doc         *Documentation
	Annotations *FormatAnnotations // Format-specific annotations
}

// Method represents an RPC method
type Method struct {
	Name         string
	InputType    string
	OutputType   string
	InputStream  bool // Client-side streaming
	OutputStream bool // Server-side streaming
	Doc          *Documentation
	HTTPMethod   string   // HTTP method for OpenAPI (GET, POST, PUT, DELETE, PATCH)
	GraphQLType  string   // GraphQL operation type (query, mutation, subscription)
	PathTemplate string   // URL path template for OpenAPI (e.g., "/users/{id}")
	SuccessCodes []string // Additional success HTTP codes beyond 200 (e.g., "201", "204")
	ErrorCodes   []string // Expected HTTP error codes (e.g., "400", "404", "500")
}

// GetHTTPMethod returns the HTTP method, using heuristics if not explicitly set
func (m *Method) GetHTTPMethod() string {
	if m.HTTPMethod != "" {
		return strings.ToLower(m.HTTPMethod)
	}
	// Default heuristic: Get/List are GET, everything else is POST
	if strings.HasPrefix(m.Name, "Get") || strings.HasPrefix(m.Name, "List") {
		return "get"
	}
	return "post"
}

// GetGraphQLType returns the GraphQL operation type, using heuristics if not explicitly set
func (m *Method) GetGraphQLType() string {
	if m.GraphQLType != "" {
		return m.GraphQLType
	}
	// Heuristic: methods with OutputStream (stream returns) are subscriptions
	if m.OutputStream {
		return "subscription"
	}
	// Default heuristic: Get/List are queries, everything else is mutations
	if strings.HasPrefix(m.Name, "Get") || strings.HasPrefix(m.Name, "List") {
		return "query"
	}
	return "mutation"
}

// Documentation represents documentation comments
type Documentation struct {
	General  string            // General documentation for all languages
	Specific map[string]string // Language-specific documentation (proto, graphql, openapi)
}

// GetDoc returns the documentation for a specific language, falling back to general doc
func (d *Documentation) GetDoc(lang string) string {
	if d == nil {
		return ""
	}
	if specific, ok := d.Specific[lang]; ok && specific != "" {
		return specific
	}
	return d.General
}

// BuiltinTypes maps primitive type names to their existence in the type system.
var BuiltinTypes = map[string]bool{
	"string":    true,
	"int32":     true,
	"int64":     true,
	"float32":   true,
	"float64":   true,
	"bool":      true,
	"timestamp": true,
	"bytes":     true,
}

// IsBuiltinType checks if a type name is a builtin type
func IsBuiltinType(typeName string) bool {
	return BuiltinTypes[typeName]
}

// FormatAnnotations holds format-specific annotations for types and fields
type FormatAnnotations struct {
	Proto       []string // Protobuf options: ["packed = false", "retention = RETENTION_SOURCE"]
	GraphQL     []string // GraphQL directives: ["@key(fields: \"id\")", "@external"]
	OpenAPI     []string // OpenAPI extensions: ["x-internal-id: prod", "x-format: currency"]
	Go          []string // Go options: ["package = \"mypackage\""]
	ProtoName   string   // Override name for Protobuf generation (from @proto.name annotation)
	GraphQLName string   // Override name for GraphQL generation (from @graphql.name annotation)
	OpenAPIName string   // Override name for OpenAPI generation (from @openapi.name annotation)
	GoName      string   // Override name for Go generation (from @go.name annotation)
}

// NewFormatAnnotations creates a new FormatAnnotations instance
func NewFormatAnnotations() *FormatAnnotations {
	return &FormatAnnotations{
		Proto:   []string{},
		GraphQL: []string{},
		OpenAPI: []string{},
	}
}

// TypeRegistry maintains a registry of all types, enums, and unions for namespace resolution
type TypeRegistry struct {
	// Map from qualified name (namespace.TypeName) to the definition
	Types  map[string]*Type
	Enums  map[string]*Enum
	Unions map[string]*Union
}

// NewTypeRegistry creates a new empty type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		Types:  make(map[string]*Type),
		Enums:  make(map[string]*Enum),
		Unions: make(map[string]*Union),
	}
}

// RegisterType registers a type in the registry
func (tr *TypeRegistry) RegisterType(typ *Type) {
	qualifiedName := typ.Namespace + "." + typ.Name
	tr.Types[qualifiedName] = typ
}

// RegisterEnum registers an enum in the registry
func (tr *TypeRegistry) RegisterEnum(enum *Enum) {
	qualifiedName := enum.Namespace + "." + enum.Name
	tr.Enums[qualifiedName] = enum
}

// RegisterUnion registers a union in the registry
func (tr *TypeRegistry) RegisterUnion(union *Union) {
	qualifiedName := union.Namespace + "." + union.Name
	tr.Unions[qualifiedName] = union
}

// ResolveType resolves a type name (qualified or unqualified) to its qualified name
// If the name is already qualified (contains a dot), it returns it as-is
// Otherwise, it tries to find it in the given namespace
func (tr *TypeRegistry) ResolveType(name string, currentNamespace string) (string, bool) {
	// If already qualified (contains dot), return as-is
	if strings.Contains(name, ".") {
		// Check if it exists
		if _, ok := tr.Types[name]; ok {
			return name, true
		}
		if _, ok := tr.Enums[name]; ok {
			return name, true
		}
		if _, ok := tr.Unions[name]; ok {
			return name, true
		}
		return name, false
	}

	// Try current namespace first
	qualifiedName := currentNamespace + "." + name
	if _, ok := tr.Types[qualifiedName]; ok {
		return qualifiedName, true
	}
	if _, ok := tr.Enums[qualifiedName]; ok {
		return qualifiedName, true
	}
	if _, ok := tr.Unions[qualifiedName]; ok {
		return qualifiedName, true
	}

	// Try all namespaces (for unqualified lookups)
	// Collect all matches
	var matches []string
	for qualName := range tr.Types {
		if strings.HasSuffix(qualName, "."+name) {
			matches = append(matches, qualName)
		}
	}
	for qualName := range tr.Enums {
		if strings.HasSuffix(qualName, "."+name) {
			matches = append(matches, qualName)
		}
	}
	for qualName := range tr.Unions {
		if strings.HasSuffix(qualName, "."+name) {
			matches = append(matches, qualName)
		}
	}

	// If exactly one match, return it
	if len(matches) == 1 {
		return matches[0], true
	}

	// If multiple matches, it's ambiguous - return false
	// If no matches, return false
	return name, false
}

// GetUnqualifiedName extracts the unqualified name from a qualified name
func GetUnqualifiedName(qualifiedName string) string {
	parts := strings.Split(qualifiedName, ".")
	return parts[len(parts)-1]
}

// DeprecationInfo holds information about deprecated fields/types
type DeprecationInfo struct {
	Reason  string // Why it's deprecated and what to use instead
	Since   string // Version when it was deprecated (e.g., "2.0.0")
	Removed string // Version when it will be removed (optional, e.g., "3.0.0")
}

// ValidationRules holds validation constraints for a field
type ValidationRules struct {
	// String validation
	MinLength *int   `json:"minLength,omitempty"`
	MaxLength *int   `json:"maxLength,omitempty"`
	Pattern   string `json:"pattern,omitempty"` // Regex pattern
	Format    string `json:"format,omitempty"`  // email, url, uuid, etc.

	// Numeric validation
	Min          *float64 `json:"min,omitempty"`          // Minimum value (inclusive)
	Max          *float64 `json:"max,omitempty"`          // Maximum value (inclusive)
	ExclusiveMin *float64 `json:"exclusiveMin,omitempty"` // Minimum value (exclusive)
	ExclusiveMax *float64 `json:"exclusiveMax,omitempty"` // Maximum value (exclusive)
	MultipleOf   *float64 `json:"multipleOf,omitempty"`   // Must be multiple of this value

	// Array validation
	MinItems    *int `json:"minItems,omitempty"`
	MaxItems    *int `json:"maxItems,omitempty"`
	UniqueItems bool `json:"uniqueItems,omitempty"`

	// General
	Enum []string `json:"enum,omitempty"` // Allowed values
}
