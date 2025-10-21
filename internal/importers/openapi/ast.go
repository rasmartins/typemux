package openapi

// OpenAPISpec represents a complete OpenAPI specification
type OpenAPISpec struct {
	OpenAPI    string
	Info       *Info
	Servers    []*Server
	Paths      map[string]*PathItem
	Components *Components
	Security   []map[string][]string
	Tags       []*Tag
}

// Info represents the API metadata
type Info struct {
	Title       string
	Version     string
	Description string
	Contact     *Contact
}

// Contact represents contact information
type Contact struct {
	Name  string
	Email string
	URL   string
}

// Server represents a server definition
type Server struct {
	URL         string
	Description string
}

// PathItem represents operations available on a single path
type PathItem struct {
	Ref         string
	Summary     string
	Description string
	Get         *Operation
	Post        *Operation
	Put         *Operation
	Delete      *Operation
	Patch       *Operation
	Options     *Operation
	Head        *Operation
	Trace       *Operation
}

// Operation represents a single API operation
type Operation struct {
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Parameters  []*Parameter
	RequestBody *RequestBody
	Responses   map[string]*Response
	Security    []map[string][]string
}

// Parameter represents a parameter in an operation
type Parameter struct {
	Name        string
	In          string // query, path, header, cookie
	Description string
	Required    bool
	Schema      *Schema
	Example     interface{}
}

// RequestBody represents a request body
type RequestBody struct {
	Description string
	Required    bool
	Content     map[string]*MediaType
}

// Response represents a response
type Response struct {
	Description string
	Content     map[string]*MediaType
	Headers     map[string]*Header
}

// MediaType represents a media type (e.g., application/json)
type MediaType struct {
	Schema   *Schema
	Example  interface{}
	Examples map[string]*Example
}

// Example represents an example value
type Example struct {
	Summary     string
	Description string
	Value       interface{}
}

// Header represents a header parameter
type Header struct {
	Description string
	Required    bool
	Schema      *Schema
}

// Components represents reusable components
type Components struct {
	Schemas         map[string]*Schema
	Responses       map[string]*Response
	Parameters      map[string]*Parameter
	Examples        map[string]*Example
	RequestBodies   map[string]*RequestBody
	Headers         map[string]*Header
	SecuritySchemes map[string]*SecurityScheme
}

// Schema represents a schema object
type Schema struct {
	Ref                  string
	Type                 string
	Format               string
	Title                string
	Description          string
	Properties           map[string]*Schema
	Required             []string
	Items                *Schema // For arrays
	Enum                 []interface{}
	Default              interface{}
	Example              interface{}
	Nullable             bool
	ReadOnly             bool
	WriteOnly            bool
	Minimum              *float64
	Maximum              *float64
	Pattern              string
	MinLength            *int
	MaxLength            *int
	AdditionalProperties interface{} // Can be bool or Schema
	AllOf                []*Schema
	OneOf                []*Schema
	AnyOf                []*Schema
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type             string
	Description      string
	Name             string
	In               string
	Scheme           string
	BearerFormat     string
	Flows            *OAuthFlows
	OpenIDConnectURL string
}

// OAuthFlows represents OAuth flows
type OAuthFlows struct {
	Implicit          *OAuthFlow
	Password          *OAuthFlow
	ClientCredentials *OAuthFlow
	AuthorizationCode *OAuthFlow
}

// OAuthFlow represents an OAuth flow
type OAuthFlow struct {
	AuthorizationURL string
	TokenURL         string
	RefreshURL       string
	Scopes           map[string]string
}

// Tag represents a tag with metadata
type Tag struct {
	Name        string
	Description string
}

// IsRequired checks if a property name is in the required list
func (s *Schema) IsRequired(propertyName string) bool {
	for _, req := range s.Required {
		if req == propertyName {
			return true
		}
	}
	return false
}

// ResolveRef extracts the component name from a $ref string
// E.g., "#/components/schemas/User" returns "User"
func ResolveRef(ref string) string {
	if ref == "" {
		return ""
	}

	// Split by / and get the last part
	parts := []rune(ref)
	lastSlash := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return ref
	}

	return string(parts[lastSlash+1:])
}
