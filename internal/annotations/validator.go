package annotations

import (
	"fmt"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
)

// ValidationError represents an error found during YAML annotation validation
type ValidationError struct {
	Message string
	Path    string // e.g., "types.User.fields.email"
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("YAML annotation error at %s: %s", e.Path, e.Message)
}

// Validator validates YAML annotations against a schema
type Validator struct {
	schema *ast.Schema
	errors []*ValidationError
}

// NewValidator creates a new validator for the given schema
func NewValidator(schema *ast.Schema) *Validator {
	return &Validator{
		schema: schema,
		errors: make([]*ValidationError, 0),
	}
}

// Validate validates YAML annotations and returns any errors found
func (v *Validator) Validate(annotations *YAMLAnnotations) []*ValidationError {
	v.errors = make([]*ValidationError, 0)

	// Validate types
	for typeName, typeAnnotations := range annotations.Types {
		path := fmt.Sprintf("types.%s", typeName)

		// Check if type exists (support both simple and qualified names)
		typeExists := false

		for _, schemaType := range v.schema.Types {
			// Match by qualified name (e.g., "com.example.api.User")
			qualifiedName := schemaType.Namespace + "." + schemaType.Name
			if qualifiedName == typeName || schemaType.Name == typeName {
				typeExists = true

				// Validate field annotations
				if typeAnnotations.Fields != nil {
					v.validateFieldAnnotations(schemaType, typeAnnotations.Fields, path)
				}
				break
			}
		}

		if !typeExists {
			v.addError(path, fmt.Sprintf("references non-existent type '%s'", typeName))
		}
	}

	// Validate enums
	for enumName := range annotations.Enums {
		path := fmt.Sprintf("enums.%s", enumName)

		enumExists := false
		for _, schemaEnum := range v.schema.Enums {
			qualifiedName := schemaEnum.Namespace + "." + schemaEnum.Name
			if qualifiedName == enumName || schemaEnum.Name == enumName {
				enumExists = true
				break
			}
		}

		if !enumExists {
			v.addError(path, fmt.Sprintf("references non-existent enum '%s'", enumName))
		}
	}

	// Validate unions
	for unionName := range annotations.Unions {
		path := fmt.Sprintf("unions.%s", unionName)

		unionExists := false
		for _, schemaUnion := range v.schema.Unions {
			qualifiedName := schemaUnion.Namespace + "." + schemaUnion.Name
			if qualifiedName == unionName || schemaUnion.Name == unionName {
				unionExists = true
				break
			}
		}

		if !unionExists {
			v.addError(path, fmt.Sprintf("references non-existent union '%s'", unionName))
		}
	}

	// Validate services
	for serviceName, serviceAnnotations := range annotations.Services {
		path := fmt.Sprintf("services.%s", serviceName)

		serviceExists := false
		for _, schemaService := range v.schema.Services {
			qualifiedName := schemaService.Namespace + "." + schemaService.Name
			if qualifiedName == serviceName || schemaService.Name == serviceName {
				serviceExists = true

				// Validate method annotations
				if serviceAnnotations.Methods != nil {
					v.validateMethodAnnotations(schemaService, serviceAnnotations.Methods, path)
				}
				break
			}
		}

		if !serviceExists {
			v.addError(path, fmt.Sprintf("references non-existent service '%s'", serviceName))
		}
	}

	return v.errors
}

func (v *Validator) validateFieldAnnotations(schemaType *ast.Type, fieldAnnotations map[string]*FieldAnnotations, basePath string) {
	for fieldName, annotations := range fieldAnnotations {
		path := fmt.Sprintf("%s.fields.%s", basePath, fieldName)

		// Check if field exists
		fieldExists := false
		for _, field := range schemaType.Fields {
			if field.Name == fieldName {
				fieldExists = true
				break
			}
		}

		if !fieldExists {
			v.addError(path, fmt.Sprintf("references non-existent field '%s.%s'", schemaType.Name, fieldName))
		}

		// Validate annotation values
		if annotations.Exclude != nil && annotations.Only != nil {
			v.addError(path, "cannot specify both 'exclude' and 'only' annotations")
		}

		// Validate generator names
		validGenerators := map[string]bool{"proto": true, "graphql": true, "openapi": true}
		for _, gen := range annotations.Exclude {
			if !validGenerators[gen] {
				v.addError(path, fmt.Sprintf("invalid generator name in exclude: '%s'", gen))
			}
		}
		for _, gen := range annotations.Only {
			if !validGenerators[gen] {
				v.addError(path, fmt.Sprintf("invalid generator name in only: '%s'", gen))
			}
		}
	}
}

func (v *Validator) validateMethodAnnotations(schemaService *ast.Service, methodAnnotations map[string]*MethodAnnotations, basePath string) {
	for methodName, annotations := range methodAnnotations {
		path := fmt.Sprintf("%s.methods.%s", basePath, methodName)

		// Check if method exists
		methodExists := false
		for _, method := range schemaService.Methods {
			if method.Name == methodName {
				methodExists = true
				break
			}
		}

		if !methodExists {
			v.addError(path, fmt.Sprintf("references non-existent method '%s.%s'", schemaService.Name, methodName))
		}

		// Validate HTTP method
		if annotations.HTTP != "" {
			validHTTPMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
			}
			httpMethod := strings.ToUpper(annotations.HTTP)
			if !validHTTPMethods[httpMethod] {
				v.addError(path, fmt.Sprintf("invalid HTTP method: '%s'", annotations.HTTP))
			}
		}

		// Validate GraphQL operation type
		if annotations.GraphQL != "" {
			validGraphQLTypes := map[string]bool{
				"query": true, "mutation": true, "subscription": true,
			}
			if !validGraphQLTypes[annotations.GraphQL] {
				v.addError(path, fmt.Sprintf("invalid GraphQL operation type: '%s'", annotations.GraphQL))
			}
		}

		// Validate status codes
		for _, code := range annotations.Success {
			if code < 100 || code > 599 {
				v.addError(path, fmt.Sprintf("invalid HTTP status code in success: %d", code))
			}
		}
		for _, code := range annotations.Errors {
			if code < 100 || code > 599 {
				v.addError(path, fmt.Sprintf("invalid HTTP status code in errors: %d", code))
			}
		}
	}
}

func (v *Validator) addError(path, message string) {
	v.errors = append(v.errors, &ValidationError{
		Path:    path,
		Message: message,
	})
}

// HasErrors returns true if validation found any errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// FormatErrors returns a formatted string of all validation errors
func (v *Validator) FormatErrors() string {
	if len(v.errors) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d validation error(s) in YAML annotations:\n\n", len(v.errors)))
	for _, err := range v.errors {
		sb.WriteString(fmt.Sprintf("  • %s\n", err.Error()))
	}
	return sb.String()
}
