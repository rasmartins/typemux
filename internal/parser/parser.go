package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
	"github.com/rasmartins/typemux/internal/lexer"
)

// Parser transforms a stream of tokens from the lexer into an abstract syntax tree (AST).
type Parser struct {
	lexer   *lexer.Lexer
	curTok  lexer.Token
	peekTok lexer.Token
	errors  []string
}

// New creates a new parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{lexer: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.lexer.NextToken()
}

// Errors returns all parsing errors encountered during parsing.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("Line %d:%d - %s", p.curTok.Line, p.curTok.Column, msg))
}

func (p *Parser) expectToken(t lexer.TokenType) bool {
	if p.curTok.Type == t {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("expected %s, got %s", t, p.curTok.Type))
	return false
}

// Parse parses the input tokens into an abstract syntax tree (AST) representing a TypeMUX schema.
func (p *Parser) Parse() *ast.Schema {
	schema := &ast.Schema{
		Namespace: "api", // default namespace
		Imports:   []string{},
		Enums:     []*ast.Enum{},
		Types:     []*ast.Type{},
		Unions:    []*ast.Union{},
		Services:  []*ast.Service{},
	}

	// Parse schema-level annotations at the beginning (@typemux, @version)
	// These must appear before any other declarations
	for p.curTok.Type == lexer.TOKEN_AT {
		// Peek at what comes after @
		if p.peekTok.Type == lexer.TOKEN_IDENT {
			attrName := p.peekTok.Literal
			if attrName == "typemux" || attrName == "version" {
				// Parse this schema-level annotation
				p.nextToken() // consume @
				p.nextToken() // consume identifier
				if p.curTok.Type == lexer.TOKEN_LPAREN {
					p.nextToken()
					if p.curTok.Type == lexer.TOKEN_STRING || p.curTok.Type == lexer.TOKEN_IDENT {
						value := strings.Trim(p.curTok.Literal, "\"'")
						if attrName == "typemux" {
							schema.TypeMUXVersion = value
						} else {
							schema.Version = value
						}
						p.nextToken()
						p.expectToken(lexer.TOKEN_RPAREN)
					}
				}
			} else {
				// Not a schema-level annotation, stop looking
				break
			}
		} else {
			// Not an identifier after @, stop looking
			break
		}
	}

	for p.curTok.Type != lexer.TOKEN_EOF {
		// Collect documentation that might precede the next declaration
		doc := p.parseDocumentation()

		// Collect leading annotations that might precede the next declaration
		leadingAnnotations := p.parseLeadingAnnotations()

		switch p.curTok.Type {
		case lexer.TOKEN_NAMESPACE:
			namespace := p.parseNamespace()
			if namespace != "" {
				schema.Namespace = namespace

				// Only store leading annotations if they exist (these would be annotations before the namespace keyword)
				if leadingAnnotations != nil && (len(leadingAnnotations.Proto) > 0 || len(leadingAnnotations.GraphQL) > 0 || len(leadingAnnotations.OpenAPI) > 0 || len(leadingAnnotations.Go) > 0) {
					schema.NamespaceAnnotations = leadingAnnotations
				}
				// Note: We do NOT parse trailing annotations here because annotations that appear
				// after the namespace declaration should be treated as leading annotations for
				// the next declaration (type, enum, etc.), not as namespace annotations.
			}
		case lexer.TOKEN_AT:
			// Handle special schema-level annotations before anything else
			p.nextToken()
			if p.curTok.Type == lexer.TOKEN_IDENT {
				attrName := p.curTok.Literal
				p.nextToken()

				if attrName == "typemux" {
					if p.curTok.Type == lexer.TOKEN_LPAREN {
						p.nextToken()
						if p.curTok.Type == lexer.TOKEN_STRING || p.curTok.Type == lexer.TOKEN_IDENT {
							schema.TypeMUXVersion = strings.Trim(p.curTok.Literal, "\"'")
							p.nextToken()
							p.expectToken(lexer.TOKEN_RPAREN)
						}
					}
				} else if attrName == "version" {
					if p.curTok.Type == lexer.TOKEN_LPAREN {
						p.nextToken()
						if p.curTok.Type == lexer.TOKEN_STRING || p.curTok.Type == lexer.TOKEN_IDENT {
							schema.Version = strings.Trim(p.curTok.Literal, "\"'")
							p.nextToken()
							p.expectToken(lexer.TOKEN_RPAREN)
						}
					}
				}
			}
		case lexer.TOKEN_IMPORT:
			importPath := p.parseImport()
			if importPath != "" {
				schema.Imports = append(schema.Imports, importPath)
			}
		case lexer.TOKEN_ENUM:
			enum := p.parseEnumWithDocAndAnnotations(doc, leadingAnnotations, schema.Namespace)
			if enum != nil {
				schema.Enums = append(schema.Enums, enum)
			}
		case lexer.TOKEN_TYPE:
			typ := p.parseTypeWithDocAndAnnotations(doc, leadingAnnotations, schema.Namespace)
			if typ != nil {
				schema.Types = append(schema.Types, typ)
			}
		case lexer.TOKEN_UNION:
			union := p.parseUnionWithDocAndAnnotations(doc, leadingAnnotations, schema.Namespace)
			if union != nil {
				schema.Unions = append(schema.Unions, union)
			}
		case lexer.TOKEN_SERVICE:
			service := p.parseServiceWithDocAndAnnotations(doc, leadingAnnotations, schema.Namespace)
			if service != nil {
				schema.Services = append(schema.Services, service)
			}
		default:
			p.nextToken()
		}
	}

	return schema
}

func (p *Parser) parseEnumWithDocAndAnnotations(doc *ast.Documentation, leadingAnnotations *ast.FormatAnnotations, namespace string) *ast.Enum {
	p.nextToken() // consume 'enum'

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected enum name")
		return nil
	}

	enum := &ast.Enum{
		Name:      p.curTok.Literal,
		Namespace: namespace,
		Values:    []*ast.EnumValue{},
		Doc:       doc,
	}

	p.nextToken()

	// Parse trailing enum-level annotations
	trailingAnnotations := p.parseLeadingAnnotations()

	// Merge leading and trailing annotations
	enum.Annotations = p.mergeAnnotations(leadingAnnotations, trailingAnnotations)

	if !p.expectToken(lexer.TOKEN_LBRACE) {
		return nil
	}

	for p.curTok.Type == lexer.TOKEN_IDENT || p.curTok.Type == lexer.TOKEN_DOC_COMMENT {
		// Parse documentation for enum value
		valueDoc := p.parseDocumentation()

		if p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError("expected enum value name")
			return nil
		}

		enumValue := &ast.EnumValue{
			Name: p.curTok.Literal,
			Doc:  valueDoc,
		}
		p.nextToken()

		// Check for optional = number syntax
		if p.curTok.Type == lexer.TOKEN_EQUALS {
			p.nextToken()
			if p.curTok.Type == lexer.TOKEN_NUMBER {
				// Parse the number
				var num int
				if _, err := fmt.Sscanf(p.curTok.Literal, "%d", &num); err == nil {
					enumValue.Number = num
					enumValue.HasNumber = true
				}
				p.nextToken()
			} else {
				p.addError("expected number after =")
				return nil
			}
		}

		enum.Values = append(enum.Values, enumValue)
	}

	if !p.expectToken(lexer.TOKEN_RBRACE) {
		return nil
	}

	return enum
}

func (p *Parser) parseTypeWithDocAndAnnotations(doc *ast.Documentation, leadingAnnotations *ast.FormatAnnotations, namespace string) *ast.Type {
	p.nextToken() // consume 'type'

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected type name")
		return nil
	}

	typ := &ast.Type{
		Name:      p.curTok.Literal,
		Namespace: namespace,
		Fields:    []*ast.Field{},
		Doc:       doc,
	}

	p.nextToken()

	// Parse trailing type-level annotations like @graphql.directive(...) @openapi.extension(...)
	trailingAnnotations := p.parseLeadingAnnotations() // reuse the same method

	// Merge leading and trailing annotations
	typ.Annotations = p.mergeAnnotations(leadingAnnotations, trailingAnnotations)

	if !p.expectToken(lexer.TOKEN_LBRACE) {
		return nil
	}

	for p.curTok.Type == lexer.TOKEN_IDENT || p.curTok.Type == lexer.TOKEN_DOC_COMMENT || p.curTok.Type == lexer.TOKEN_AT {
		// Collect field documentation
		fieldDoc := p.parseDocumentation()

		// Collect field leading annotations and attributes
		fieldLeadingAnnotations := ast.NewFormatAnnotations()
		leadingAttributes := make(map[string]string)

		for p.curTok.Type == lexer.TOKEN_AT {
			// Peek ahead to determine if this is a simple attribute or format annotation
			p.nextToken() // consume @

			if p.curTok.Type != lexer.TOKEN_IDENT {
				p.addError("expected annotation name")
				break
			}

			attrName := p.curTok.Literal
			p.nextToken()

			// Check if this is a format-specific annotation (has a dot)
			if p.curTok.Type == lexer.TOKEN_DOT && (attrName == "proto" || attrName == "graphql" || attrName == "openapi") {
				// This is a format annotation like @proto.name("foo")
				// Back up and let parseSingleAnnotation handle it
				p.curTok = lexer.Token{Type: lexer.TOKEN_AT, Literal: "@"}
				p.parseSingleAnnotation(fieldLeadingAnnotations)
			} else if p.curTok.Type == lexer.TOKEN_LPAREN {
				// This is an attribute with parameters - store it for later parsing
				// Mark that it has parameters by storing a special value
				leadingAttributes[attrName] = "NEEDS_PARSING"
			} else {
				// This is a simple attribute like @required
				leadingAttributes[attrName] = ""
			}
		}

		field := p.parseFieldWithLeadingAnnotations(fieldDoc, fieldLeadingAnnotations, leadingAttributes)
		if field != nil {
			typ.Fields = append(typ.Fields, field)
		}
	}

	if !p.expectToken(lexer.TOKEN_RBRACE) {
		return nil
	}

	return typ
}

func (p *Parser) parseUnionWithDocAndAnnotations(doc *ast.Documentation, leadingAnnotations *ast.FormatAnnotations, namespace string) *ast.Union {
	p.nextToken() // consume 'union'

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected union name")
		return nil
	}

	union := &ast.Union{
		Name:      p.curTok.Literal,
		Namespace: namespace,
		Options:   []string{},
		Doc:       doc,
	}

	p.nextToken()

	// Parse trailing union-level annotations
	trailingAnnotations := p.parseLeadingAnnotations()

	// Merge leading and trailing annotations
	union.Annotations = p.mergeAnnotations(leadingAnnotations, trailingAnnotations)

	if !p.expectToken(lexer.TOKEN_LBRACE) {
		return nil
	}

	// Parse union options (list of type names)
	for p.curTok.Type != lexer.TOKEN_RBRACE && p.curTok.Type != lexer.TOKEN_EOF {
		if p.curTok.Type == lexer.TOKEN_IDENT {
			union.Options = append(union.Options, p.curTok.Literal)
			p.nextToken()
		} else {
			p.addError("expected type name in union")
			p.nextToken()
		}
	}

	if !p.expectToken(lexer.TOKEN_RBRACE) {
		return nil
	}

	return union
}

func (p *Parser) parseFieldWithLeadingAnnotations(doc *ast.Documentation, leadingAnnotations *ast.FormatAnnotations, leadingAttributes map[string]string) *ast.Field {
	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected field name")
		return nil
	}

	field := &ast.Field{
		Name:       p.curTok.Literal,
		Attributes: make(map[string]string),
		Doc:        doc,
	}

	// Apply leading attributes (like @required)
	for k, v := range leadingAttributes {
		field.Attributes[k] = v
		if k == "required" {
			field.Required = true
		}
	}

	p.nextToken()

	if !p.expectToken(lexer.TOKEN_COLON) {
		return nil
	}

	// Parse field type
	field.Type = p.parseFieldType()
	if field.Type == nil {
		return nil
	}

	// Check for optional = number syntax (for protobuf field numbers)
	fieldLine := p.curTok.Line // Track the line where the field type/number is
	if p.curTok.Type == lexer.TOKEN_EQUALS {
		p.nextToken()
		if p.curTok.Type == lexer.TOKEN_NUMBER {
			// Parse the number
			var num int
			if _, err := fmt.Sscanf(p.curTok.Literal, "%d", &num); err == nil {
				field.Number = num
				field.HasNumber = true
			}
			fieldLine = p.curTok.Line // Update to the line of the number
			p.nextToken()
		} else {
			p.addError("expected number after =")
			return nil
		}
	}

	// Parse attributes (@required, @default, @exclude, @only, etc.) and trailing annotations
	// Only consume @ tokens on the same line as the field (to avoid consuming leading annotations for the next field)
	trailingFieldAnnotations := ast.NewFormatAnnotations()
	for p.curTok.Type == lexer.TOKEN_AT && p.curTok.Line == fieldLine {
		p.nextToken()
		if p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError("expected attribute name")
			return nil
		}

		attrName := p.curTok.Literal
		p.nextToken()

		if attrName == "required" {
			field.Required = true
			field.Attributes[attrName] = ""
		} else if attrName == "default" {
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				if p.curTok.Type == lexer.TOKEN_IDENT || p.curTok.Type == lexer.TOKEN_NUMBER {
					field.Default = p.curTok.Literal
					p.nextToken()
					p.expectToken(lexer.TOKEN_RPAREN)
				}
			}
			field.Attributes[attrName] = ""
		} else if attrName == "exclude" {
			// Parse @exclude(proto,graphql)
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				generators := p.parseGeneratorList()
				field.ExcludeFrom = generators
				p.expectToken(lexer.TOKEN_RPAREN)
			}
			field.Attributes[attrName] = ""
		} else if attrName == "only" {
			// Parse @only(openapi)
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				generators := p.parseGeneratorList()
				field.OnlyFor = generators
				p.expectToken(lexer.TOKEN_RPAREN)
			}
			field.Attributes[attrName] = ""
		} else if attrName == "deprecated" {
			// Parse @deprecated("reason", since="2.0.0", removed="3.0.0")
			if field.Deprecated == nil {
				field.Deprecated = &ast.DeprecationInfo{}
			}
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				p.parseDeprecationInfo(field.Deprecated)
				p.expectToken(lexer.TOKEN_RPAREN)
			}
		} else if attrName == "since" {
			// Parse @since("2.0.0")
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				if p.curTok.Type == lexer.TOKEN_STRING || p.curTok.Type == lexer.TOKEN_IDENT {
					field.Since = strings.Trim(p.curTok.Literal, "\"'")
					p.nextToken()
					p.expectToken(lexer.TOKEN_RPAREN)
				}
			}
		} else if attrName == "validate" {
			// Parse @validate(format="email", min=0, max=100, etc.)
			if field.Validation == nil {
				field.Validation = &ast.ValidationRules{}
			}
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				p.parseValidationRules(field.Validation)
				p.expectToken(lexer.TOKEN_RPAREN)
			}
		} else if attrName == "proto" || attrName == "graphql" || attrName == "openapi" || attrName == "json" {
			// Parse format-specific annotations like @proto.option([packed = false]), @proto.name("TypeName"), or @json.name("field_name")
			// Expect a dot
			if p.curTok.Type != lexer.TOKEN_DOT {
				p.addError(fmt.Sprintf("expected . after @%s", attrName))
				return nil
			}
			p.nextToken()

			// Expect subtype identifier
			if p.curTok.Type != lexer.TOKEN_IDENT {
				p.addError(fmt.Sprintf("expected subtype after @%s.", attrName))
				return nil
			}
			subtype := p.curTok.Literal
			p.nextToken()

			// Handle JSON annotations specially (some don't require parentheses)
			if attrName == "json" {
				if subtype == "nullable" {
					field.JSONNullable = true
				} else if subtype == "omitempty" {
					field.JSONOmitEmpty = true
				} else if subtype == "name" {
					// @json.name requires a parameter
					if p.curTok.Type == lexer.TOKEN_LPAREN {
						p.nextToken()
						content := p.parseAnnotationContent()
						p.expectToken(lexer.TOKEN_RPAREN)
						field.JSONName = strings.Trim(content, "\"'")
					} else {
						p.addError("@json.name requires a parameter: @json.name(\"field_name\")")
					}
				}
				continue
			}

			// Parse the content in parentheses
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				content := p.parseAnnotationContent()
				p.expectToken(lexer.TOKEN_RPAREN)

				// Handle name annotation specially
				if subtype == "name" {
					name := strings.Trim(content, "\"'")
					if attrName == "proto" {
						trailingFieldAnnotations.ProtoName = name
					} else if attrName == "graphql" {
						trailingFieldAnnotations.GraphQLName = name
					} else if attrName == "openapi" {
						trailingFieldAnnotations.OpenAPIName = name
					}
				} else {
					// Store in appropriate list for other subtypes
					if attrName == "proto" {
						trailingFieldAnnotations.Proto = append(trailingFieldAnnotations.Proto, content)
					} else if attrName == "graphql" {
						trailingFieldAnnotations.GraphQL = append(trailingFieldAnnotations.GraphQL, content)
					} else if attrName == "openapi" {
						trailingFieldAnnotations.OpenAPI = append(trailingFieldAnnotations.OpenAPI, content)
					}
				}
			}
		} else {
			field.Attributes[attrName] = ""
		}
	}

	// Merge leading and trailing field annotations
	field.Annotations = p.mergeAnnotations(leadingAnnotations, trailingFieldAnnotations)

	return field
}

// parseAnnotationContent reads everything inside annotation parentheses as a string
func (p *Parser) parseAnnotationContent() string {
	var content string
	depth := 1 // We're already inside the first (

	for depth > 0 && p.curTok.Type != lexer.TOKEN_EOF {
		if p.curTok.Type == lexer.TOKEN_LPAREN {
			depth++
			content += "("
		} else if p.curTok.Type == lexer.TOKEN_RPAREN {
			depth--
			if depth > 0 {
				content += ")"
			}
		} else if p.curTok.Type == lexer.TOKEN_LBRACKET {
			content += "["
		} else if p.curTok.Type == lexer.TOKEN_RBRACKET {
			content += "]"
		} else if p.curTok.Type == lexer.TOKEN_COLON {
			content += ":"
		} else if p.curTok.Type == lexer.TOKEN_COMMA {
			content += ", "
		} else if p.curTok.Type == lexer.TOKEN_EQUALS {
			content += " = "
		} else if p.curTok.Type == lexer.TOKEN_AT {
			content += "@"
		} else if p.curTok.Type == lexer.TOKEN_STRING {
			content += "\"" + p.curTok.Literal + "\""
		} else {
			content += p.curTok.Literal
		}

		if depth > 0 {
			p.nextToken()
		}
	}

	return content
}

// parseGeneratorList parses a comma-separated list of generator names
func (p *Parser) parseGeneratorList() []string {
	var generators []string

	if p.curTok.Type == lexer.TOKEN_IDENT {
		generators = append(generators, p.curTok.Literal)
		p.nextToken()
	}

	for p.curTok.Type == lexer.TOKEN_COMMA {
		p.nextToken()
		if p.curTok.Type == lexer.TOKEN_IDENT {
			generators = append(generators, p.curTok.Literal)
			p.nextToken()
		}
	}

	return generators
}

func (p *Parser) parseFieldType() *ast.FieldType {
	return p.parseFieldTypeInternal(true) // Allow optional marker at top level
}

func (p *Parser) parseFieldTypeInternal(allowOptional bool) *ast.FieldType {
	fieldType := &ast.FieldType{}

	// Check for array type []
	if p.curTok.Type == lexer.TOKEN_LBRACKET {
		p.nextToken()
		if p.curTok.Type == lexer.TOKEN_RBRACKET {
			p.nextToken()
			// Recursively parse the element type (supports nested arrays like [][])
			// Do not allow ? on element types - it should only appear at the end
			elementType := p.parseFieldTypeInternal(false)
			if elementType == nil {
				p.addError("expected element type after []")
				return nil
			}
			// For arrays, store the element type in Name and set IsArray
			fieldType.IsArray = true
			fieldType.Name = elementType.Name
			fieldType.IsBuiltin = elementType.IsBuiltin

			// If the element is also an array or map, we need to preserve that structure
			// For now, nested arrays will have the inner array type in the Name
			// Example: [][]string becomes Name="[]string", IsArray=true
			if elementType.IsArray {
				fieldType.Name = "[]" + elementType.Name
			} else if elementType.IsMap {
				// For arrays of maps, we need special handling
				fieldType.Name = "map"
				fieldType.MapKey = elementType.MapKey
				fieldType.MapValueType = elementType.MapValueType
				fieldType.MapValue = elementType.MapValue
			}

			// Check for optional marker (?) only if allowed
			if allowOptional && p.curTok.Type == lexer.TOKEN_QUESTION {
				fieldType.Optional = true
				p.nextToken()
			}

			return fieldType
		}
	}

	// Check for map type
	if p.curTok.Type == lexer.TOKEN_IDENT && p.curTok.Literal == "map" {
		p.nextToken()
		if p.curTok.Type == lexer.TOKEN_LT {
			p.nextToken()
			if p.curTok.Type == lexer.TOKEN_IDENT {
				fieldType.MapKey = p.curTok.Literal
				p.nextToken()
				if p.curTok.Type == lexer.TOKEN_COMMA {
					p.nextToken()
					// Recursively parse the value type (supports nested maps, arrays, etc.)
					valueType := p.parseFieldType()
					if valueType == nil {
						p.addError("expected value type in map")
						return nil
					}

					// Store both old format (MapValue string) for backward compatibility
					// and new format (MapValueType) for complex types
					fieldType.MapValueType = valueType
					if !valueType.IsMap && !valueType.IsArray {
						// Simple type - also store in MapValue for backward compatibility
						fieldType.MapValue = valueType.Name
					}

					if p.curTok.Type == lexer.TOKEN_GT {
						p.nextToken()
						fieldType.IsMap = true
						fieldType.Name = "map"
						fieldType.IsBuiltin = false

						// Check for optional marker (?) after map type only if allowed
						if allowOptional && p.curTok.Type == lexer.TOKEN_QUESTION {
							fieldType.Optional = true
							p.nextToken()
						}

						return fieldType
					} else {
						p.addError("expected '>' to close map type")
						return nil
					}
				}
			}
		}
	}

	// Parse base type name (may be qualified like com.example.User)
	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected type name")
		return nil
	}

	// Build the type name, supporting dotted notation for qualified names
	var nameParts []string
	nameParts = append(nameParts, p.curTok.Literal)
	p.nextToken()

	// Continue reading dots and identifiers for qualified type names
	for p.curTok.Type == lexer.TOKEN_DOT {
		p.nextToken() // consume '.'
		if p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError("expected identifier after '.' in type name")
			break
		}
		nameParts = append(nameParts, p.curTok.Literal)
		p.nextToken()
	}

	fieldType.Name = strings.Join(nameParts, ".")
	fieldType.IsBuiltin = ast.IsBuiltinType(fieldType.Name)

	// Check for optional marker (?) only if allowed at this level
	if allowOptional && p.curTok.Type == lexer.TOKEN_QUESTION {
		fieldType.Optional = true
		p.nextToken()
	}

	return fieldType
}

func (p *Parser) parseServiceWithDocAndAnnotations(doc *ast.Documentation, leadingAnnotations *ast.FormatAnnotations, namespace string) *ast.Service {
	p.nextToken() // consume 'service'

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected service name")
		return nil
	}

	service := &ast.Service{
		Name:      p.curTok.Literal,
		Namespace: namespace,
		Methods:   []*ast.Method{},
		Doc:       doc,
	}

	p.nextToken()

	// Parse trailing service-level annotations
	trailingAnnotations := p.parseLeadingAnnotations()

	// Merge leading and trailing annotations
	service.Annotations = p.mergeAnnotations(leadingAnnotations, trailingAnnotations)

	if !p.expectToken(lexer.TOKEN_LBRACE) {
		return nil
	}

	for p.curTok.Type == lexer.TOKEN_RPC || p.curTok.Type == lexer.TOKEN_DOC_COMMENT {
		method := p.parseMethod()
		if method != nil {
			service.Methods = append(service.Methods, method)
		}
	}

	if !p.expectToken(lexer.TOKEN_RBRACE) {
		return nil
	}

	return service
}

func (p *Parser) parseMethod() *ast.Method {
	// Collect documentation before 'rpc' keyword
	doc := p.parseDocumentation()

	p.nextToken() // consume 'rpc'

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected method name")
		return nil
	}

	method := &ast.Method{
		Name: p.curTok.Literal,
		Doc:  doc,
	}

	p.nextToken()

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	// Check for stream keyword before input type
	if p.curTok.Type == lexer.TOKEN_STREAM {
		method.InputStream = true
		p.nextToken()
	}

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected input type")
		return nil
	}

	method.InputType = p.curTok.Literal
	p.nextToken()

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectToken(lexer.TOKEN_RETURNS) {
		return nil
	}

	if !p.expectToken(lexer.TOKEN_LPAREN) {
		return nil
	}

	// Check for stream keyword before output type
	if p.curTok.Type == lexer.TOKEN_STREAM {
		method.OutputStream = true
		p.nextToken()
	}

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected output type")
		return nil
	}

	method.OutputType = p.curTok.Literal
	p.nextToken()

	if !p.expectToken(lexer.TOKEN_RPAREN) {
		return nil
	}

	// Parse method attributes (@http, @graphql)
	for p.curTok.Type == lexer.TOKEN_AT {
		p.nextToken()
		if p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError("expected attribute name")
			return nil
		}

		attrName := p.curTok.Literal
		p.nextToken()

		// Handle @http with dotted notation: @http.method, @http.path, @http.success, @http.errors
		if attrName == "http" {
			// Check if this is dotted notation (@http.method, @http.path, etc.)
			if p.curTok.Type == lexer.TOKEN_DOT {
				p.nextToken()
				if p.curTok.Type == lexer.TOKEN_IDENT {
					subtype := p.curTok.Literal
					p.nextToken()

					if p.curTok.Type == lexer.TOKEN_LPAREN {
						p.nextToken()

						switch subtype {
						case "method":
							// Parse @http.method(GET)
							if p.curTok.Type == lexer.TOKEN_IDENT {
								method.HTTPMethod = strings.ToUpper(p.curTok.Literal)
								p.nextToken()
							}
						case "path":
							// Parse @http.path("/users/{id}")
							if p.curTok.Type == lexer.TOKEN_STRING {
								method.PathTemplate = p.curTok.Literal
								p.nextToken()
							}
						case "success":
							// Parse @http.success(201,204)
							successCodes := p.parseStatusCodeList()
							method.SuccessCodes = successCodes
						case "errors":
							// Parse @http.errors(400,404,500)
							errorCodes := p.parseStatusCodeList()
							method.ErrorCodes = errorCodes
						}

						p.expectToken(lexer.TOKEN_RPAREN)
					}
				}
			}
		} else if attrName == "graphql" {
			// Parse @graphql(query) or @graphql(mutation)
			if p.curTok.Type == lexer.TOKEN_LPAREN {
				p.nextToken()
				if p.curTok.Type == lexer.TOKEN_IDENT {
					method.GraphQLType = strings.ToLower(p.curTok.Literal)
					p.nextToken()
					p.expectToken(lexer.TOKEN_RPAREN)
				}
			}
		}
	}

	return method
}

// parseStatusCodeList parses a comma-separated list of HTTP status codes
func (p *Parser) parseStatusCodeList() []string {
	var codes []string

	if p.curTok.Type == lexer.TOKEN_NUMBER {
		codes = append(codes, p.curTok.Literal)
		p.nextToken()

		for p.curTok.Type == lexer.TOKEN_COMMA {
			p.nextToken()
			if p.curTok.Type == lexer.TOKEN_NUMBER {
				codes = append(codes, p.curTok.Literal)
				p.nextToken()
			}
		}
	}

	return codes
}

// PrintErrors returns all parsing errors as a single formatted string.
func (p *Parser) PrintErrors() string {
	return strings.Join(p.errors, "\n")
}

// parseImport parses an import statement: import "path/to/file.typemux"
func (p *Parser) parseImport() string {
	p.nextToken() // consume 'import'

	if p.curTok.Type != lexer.TOKEN_STRING {
		p.addError("expected string after import")
		return ""
	}

	importPath := p.curTok.Literal
	p.nextToken()

	return importPath
}

func (p *Parser) parseNamespace() string {
	p.nextToken() // consume 'namespace'

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected namespace identifier")
		return ""
	}

	// Build the namespace, supporting dotted notation (e.g., com.example.api)
	var parts []string
	parts = append(parts, p.curTok.Literal)
	p.nextToken()

	// Continue reading dots and identifiers for dotted namespaces
	for p.curTok.Type == lexer.TOKEN_DOT {
		p.nextToken() // consume '.'
		if p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError("expected identifier after '.' in namespace")
			return strings.Join(parts, ".")
		}
		parts = append(parts, p.curTok.Literal)
		p.nextToken()
	}

	return strings.Join(parts, ".")
}

// parseDocumentation collects doc comments and parses them into Documentation
func (p *Parser) parseDocumentation() *ast.Documentation {
	var docLines []string

	// Collect all consecutive doc comment tokens
	for p.curTok.Type == lexer.TOKEN_DOC_COMMENT {
		docLines = append(docLines, p.curTok.Literal)
		p.nextToken()
	}

	if len(docLines) == 0 {
		return nil
	}

	doc := &ast.Documentation{
		Specific: make(map[string]string),
	}

	// Regex to match language-specific comments: @proto, @graphql, @openapi
	langRegex := regexp.MustCompile(`^@(proto|graphql|openapi)\s+(.*)$`)

	var generalLines []string

	for _, line := range docLines {
		if matches := langRegex.FindStringSubmatch(line); matches != nil {
			lang := matches[1]
			text := matches[2]
			// Append to existing doc for this language
			if existing := doc.Specific[lang]; existing != "" {
				doc.Specific[lang] = existing + "\n" + text
			} else {
				doc.Specific[lang] = text
			}
		} else {
			// General documentation
			generalLines = append(generalLines, line)
		}
	}

	if len(generalLines) > 0 {
		doc.General = strings.Join(generalLines, "\n")
	}

	return doc
}

// parseLeadingAnnotations parses annotations that appear before a declaration
// It collects all @format.subtype(...) annotations until it hits a non-@ token
func (p *Parser) parseLeadingAnnotations() *ast.FormatAnnotations {
	annotations := ast.NewFormatAnnotations()

	for p.curTok.Type == lexer.TOKEN_AT {
		p.parseSingleAnnotation(annotations)
	}

	return annotations
}

// parseSingleAnnotation parses a single @format.subtype(...) annotation
// and adds it to the provided FormatAnnotations object
func (p *Parser) parseSingleAnnotation(annotations *ast.FormatAnnotations) {
	p.nextToken() // consume @

	if p.curTok.Type != lexer.TOKEN_IDENT {
		p.addError("expected annotation name")
		return
	}

	formatName := p.curTok.Literal
	p.nextToken()

	// Check for dot notation: @format.subtype(...)
	if formatName == "proto" || formatName == "graphql" || formatName == "openapi" || formatName == "go" {
		// Expect a dot
		if p.curTok.Type != lexer.TOKEN_DOT {
			p.addError(fmt.Sprintf("expected . after @%s", formatName))
			return
		}
		p.nextToken()

		// Expect subtype identifier (option, directive, extension, name)
		if p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError(fmt.Sprintf("expected subtype after @%s.", formatName))
			return
		}
		subtype := p.curTok.Literal
		p.nextToken()

		// Parse the content in parentheses
		if p.curTok.Type == lexer.TOKEN_LPAREN {
			p.nextToken()
			content := p.parseAnnotationContent()
			p.expectToken(lexer.TOKEN_RPAREN)

			// Handle name annotation specially
			if subtype == "name" {
				// Extract the name from quotes
				name := strings.Trim(content, "\"'")
				if formatName == "proto" {
					annotations.ProtoName = name
				} else if formatName == "graphql" {
					annotations.GraphQLName = name
				} else if formatName == "openapi" {
					annotations.OpenAPIName = name
				} else if formatName == "go" {
					annotations.GoName = name
				}
			} else if subtype == "package" && formatName == "go" {
				// Handle @go.package("packagename") for namespace-level annotations
				packageName := strings.Trim(content, "\"'")
				annotations.Go = append(annotations.Go, fmt.Sprintf("package = \"%s\"", packageName))
			} else {
				// Store in appropriate list for other subtypes
				if formatName == "proto" {
					annotations.Proto = append(annotations.Proto, content)
				} else if formatName == "graphql" {
					annotations.GraphQL = append(annotations.GraphQL, content)
				} else if formatName == "openapi" {
					annotations.OpenAPI = append(annotations.OpenAPI, content)
				} else if formatName == "go" {
					annotations.Go = append(annotations.Go, content)
				}
			}
		}
	}
}

// mergeAnnotations merges leading and trailing annotations
// If both have the same annotation, trailing takes precedence
func (p *Parser) mergeAnnotations(leading, trailing *ast.FormatAnnotations) *ast.FormatAnnotations {
	if leading == nil && trailing == nil {
		return nil
	}
	if leading == nil {
		return trailing
	}
	if trailing == nil {
		return leading
	}

	merged := ast.NewFormatAnnotations()

	// Merge Proto annotations
	merged.Proto = append(merged.Proto, leading.Proto...)
	merged.Proto = append(merged.Proto, trailing.Proto...)

	// Merge GraphQL annotations
	merged.GraphQL = append(merged.GraphQL, leading.GraphQL...)
	merged.GraphQL = append(merged.GraphQL, trailing.GraphQL...)

	// Merge OpenAPI annotations
	merged.OpenAPI = append(merged.OpenAPI, leading.OpenAPI...)
	merged.OpenAPI = append(merged.OpenAPI, trailing.OpenAPI...)

	// Merge Go annotations
	merged.Go = append(merged.Go, leading.Go...)
	merged.Go = append(merged.Go, trailing.Go...)

	// For name annotations, trailing takes precedence
	if trailing.ProtoName != "" {
		merged.ProtoName = trailing.ProtoName
	} else {
		merged.ProtoName = leading.ProtoName
	}

	if trailing.GraphQLName != "" {
		merged.GraphQLName = trailing.GraphQLName
	} else {
		merged.GraphQLName = leading.GraphQLName
	}

	if trailing.OpenAPIName != "" {
		merged.OpenAPIName = trailing.OpenAPIName
	} else {
		merged.OpenAPIName = leading.OpenAPIName
	}

	if trailing.GoName != "" {
		merged.GoName = trailing.GoName
	} else {
		merged.GoName = leading.GoName
	}

	return merged
}

// parseDeprecationInfo parses deprecation annotation parameters
// Format: @deprecated("reason", since="version", removed="version")
func (p *Parser) parseDeprecationInfo(info *ast.DeprecationInfo) {
	// First parameter can be a reason string (optional)
	if p.curTok.Type == lexer.TOKEN_STRING {
		info.Reason = strings.Trim(p.curTok.Literal, "\"'")
		p.nextToken()

		// If there's a comma, parse named parameters
		if p.curTok.Type == lexer.TOKEN_COMMA {
			p.nextToken()
		}
	}

	// Parse named parameters: since="version", removed="version"
	for p.curTok.Type != lexer.TOKEN_RPAREN && p.curTok.Type != lexer.TOKEN_EOF {
		if p.curTok.Type != lexer.TOKEN_IDENT {
			break
		}

		paramName := p.curTok.Literal
		p.nextToken()

		if p.curTok.Type != lexer.TOKEN_EQUALS {
			p.addError("expected = after parameter name in @deprecated")
			return
		}
		p.nextToken()

		if p.curTok.Type != lexer.TOKEN_STRING && p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError("expected string value after = in @deprecated")
			return
		}

		value := strings.Trim(p.curTok.Literal, "\"'")

		switch paramName {
		case "since":
			info.Since = value
		case "removed":
			info.Removed = value
		}

		p.nextToken()

		// Skip comma if present
		if p.curTok.Type == lexer.TOKEN_COMMA {
			p.nextToken()
		}
	}
}

// parseValidationRules parses validation parameters from @validate(...)
// Format: format="email", min=0, max=100, pattern="regex", etc.
func (p *Parser) parseValidationRules(rules *ast.ValidationRules) {
	for p.curTok.Type != lexer.TOKEN_RPAREN && p.curTok.Type != lexer.TOKEN_EOF {
		// Get the parameter name
		if p.curTok.Type != lexer.TOKEN_IDENT {
			p.addError("expected validation parameter name")
			return
		}

		paramName := p.curTok.Literal
		p.nextToken()

		// Expect =
		if p.curTok.Type != lexer.TOKEN_EQUALS {
			p.addError("expected = after validation parameter")
			return
		}
		p.nextToken()

		// Get the value
		paramValue := ""
		if p.curTok.Type == lexer.TOKEN_STRING {
			paramValue = strings.Trim(p.curTok.Literal, "\"'")
		} else if p.curTok.Type == lexer.TOKEN_NUMBER {
			paramValue = p.curTok.Literal
		} else if p.curTok.Type == lexer.TOKEN_IDENT {
			paramValue = p.curTok.Literal
		} else {
			p.addError("expected value after =")
			return
		}
		p.nextToken()

		// Apply the parameter
		p.applyValidationParameter(rules, paramName, paramValue)

		// Skip comma if present
		if p.curTok.Type == lexer.TOKEN_COMMA {
			p.nextToken()
		}
	}
}

// applyValidationParameter sets the validation rule parameter
func (p *Parser) applyValidationParameter(rules *ast.ValidationRules, name, value string) {
	switch name {
	case "format":
		rules.Format = value
	case "pattern":
		rules.Pattern = value
	case "minLength":
		if val, err := parseInt(value); err == nil {
			rules.MinLength = &val
		}
	case "maxLength":
		if val, err := parseInt(value); err == nil {
			rules.MaxLength = &val
		}
	case "min":
		if val, err := parseFloat(value); err == nil {
			rules.Min = &val
		}
	case "max":
		if val, err := parseFloat(value); err == nil {
			rules.Max = &val
		}
	case "exclusiveMin":
		if val, err := parseFloat(value); err == nil {
			rules.ExclusiveMin = &val
		}
	case "exclusiveMax":
		if val, err := parseFloat(value); err == nil {
			rules.ExclusiveMax = &val
		}
	case "multipleOf":
		if val, err := parseFloat(value); err == nil {
			rules.MultipleOf = &val
		}
	case "minItems":
		if val, err := parseInt(value); err == nil {
			rules.MinItems = &val
		}
	case "maxItems":
		if val, err := parseInt(value); err == nil {
			rules.MaxItems = &val
		}
	case "uniqueItems":
		rules.UniqueItems = (value == "true")
	}
}

// parseInt helper
func parseInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

// parseFloat helper
func parseFloat(s string) (float64, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	return val, err
}
