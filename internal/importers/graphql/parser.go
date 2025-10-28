package graphql

import (
	"fmt"
	"regexp"
	"strings"
)

type Parser struct {
	content string
	lines   []string
	pos     int
}

func NewParser(content string) *Parser {
	lines := strings.Split(content, "\n")
	return &Parser{
		content: content,
		lines:   lines,
		pos:     0,
	}
}

func (p *Parser) Parse() (*GraphQLSchema, error) {
	schema := &GraphQLSchema{
		Types:         []*GraphQLType{},
		Inputs:        []*GraphQLInput{},
		Enums:         []*GraphQLEnum{},
		Scalars:       []*GraphQLScalar{},
		Queries:       []*GraphQLField{},
		Mutations:     []*GraphQLField{},
		Subscriptions: []*GraphQLField{},
		Interfaces:    []*GraphQLInterface{},
		Unions:        []*GraphQLUnion{},
	}

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// Check for metadata comments (typemux:begin, typemux:end)
		if strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		// Collect description
		description := p.parseDescription()

		// Now check what definition follows
		line = strings.TrimSpace(p.lines[p.pos])

		if strings.HasPrefix(line, "type ") {
			typ, err := p.parseType(description)
			if err != nil {
				return nil, err
			}
			schema.Types = append(schema.Types, typ)
			continue
		} else if strings.HasPrefix(line, "input ") {
			input, err := p.parseInput(description)
			if err != nil {
				return nil, err
			}
			schema.Inputs = append(schema.Inputs, input)
			continue
		} else if strings.HasPrefix(line, "enum ") {
			enum, err := p.parseEnum(description)
			if err != nil {
				return nil, err
			}
			schema.Enums = append(schema.Enums, enum)
			continue
		} else if strings.HasPrefix(line, "scalar ") {
			scalar := p.parseScalar(description)
			schema.Scalars = append(schema.Scalars, scalar)
			continue
		} else if strings.HasPrefix(line, "interface ") {
			iface, err := p.parseInterface(description)
			if err != nil {
				return nil, err
			}
			schema.Interfaces = append(schema.Interfaces, iface)
			continue
		} else if strings.HasPrefix(line, "union ") {
			union := p.parseUnion(description)
			schema.Unions = append(schema.Unions, union)
			continue
		} else if strings.HasPrefix(line, "extend type Query") {
			queries, err := p.parseExtendQuery()
			if err != nil {
				return nil, err
			}
			schema.Queries = append(schema.Queries, queries...)
			continue
		} else if strings.HasPrefix(line, "extend type Mutation") {
			mutations, err := p.parseExtendMutation()
			if err != nil {
				return nil, err
			}
			schema.Mutations = append(schema.Mutations, mutations...)
			continue
		} else if strings.HasPrefix(line, "extend type Subscription") {
			subscriptions, err := p.parseExtendSubscription()
			if err != nil {
				return nil, err
			}
			schema.Subscriptions = append(schema.Subscriptions, subscriptions...)
			continue
		}

		p.pos++
	}

	return schema, nil
}

func (p *Parser) parseDescription() string {
	var description strings.Builder
	startPos := p.pos

	// Check for multi-line description with """
	if p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])
		if strings.HasPrefix(line, `"""`) {
			p.pos++
			for p.pos < len(p.lines) {
				line := p.lines[p.pos]
				if strings.Contains(line, `"""`) {
					// End of description
					// Extract any text before the closing """
					endIdx := strings.Index(line, `"""`)
					if endIdx > 0 {
						description.WriteString(strings.TrimSpace(line[:endIdx]))
					}
					p.pos++
					break
				}
				description.WriteString(strings.TrimSpace(line))
				description.WriteString("\n")
				p.pos++
			}
			return strings.TrimSpace(description.String())
		}
	}

	// No description found, reset position
	p.pos = startPos
	return ""
}

func (p *Parser) parseType(description string) (*GraphQLType, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// Parse metadata from previous comment if present
	metadata := p.parseMetadataFromPreviousLine()

	// type TypeName implements Interface1, Interface2 {
	// or: type TypeName {
	re := regexp.MustCompile(`type\s+(\w+)(?:\s+implements\s+([^{]+))?`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid type declaration: %s", line)
	}

	typ := &GraphQLType{
		Name:        matches[1],
		Description: description,
		Fields:      []*GraphQLField{},
		Implements:  []string{},
		Directives:  []*GraphQLDirective{},
		Metadata:    metadata,
	}

	// Parse implements clause
	if len(matches) > 2 && matches[2] != "" {
		implements := strings.Split(matches[2], ",")
		for _, iface := range implements {
			typ.Implements = append(typ.Implements, strings.TrimSpace(iface))
		}
	}

	p.pos++ // Move past type declaration

	// Parse type body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// End of type
		if line == "}" {
			p.pos++
			break
		}

		// Parse field with description
		fieldDescription := p.parseDescription()
		if p.pos >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.pos])
		if line == "}" {
			p.pos++
			break
		}

		field := p.parseField(line, fieldDescription)
		if field != nil {
			typ.Fields = append(typ.Fields, field)
		}
		p.pos++
	}

	return typ, nil
}

func (p *Parser) parseInput(description string) (*GraphQLInput, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	metadata := p.parseMetadataFromPreviousLine()

	// input InputName {
	re := regexp.MustCompile(`input\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid input declaration: %s", line)
	}

	input := &GraphQLInput{
		Name:        matches[1],
		Description: description,
		Fields:      []*GraphQLField{},
		Directives:  []*GraphQLDirective{},
		Metadata:    metadata,
	}

	p.pos++ // Move past input declaration

	// Parse input body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// End of input
		if line == "}" {
			p.pos++
			break
		}

		// Parse field with description
		fieldDescription := p.parseDescription()
		if p.pos >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.pos])
		if line == "}" {
			p.pos++
			break
		}

		field := p.parseField(line, fieldDescription)
		if field != nil {
			input.Fields = append(input.Fields, field)
		}
		p.pos++
	}

	return input, nil
}

func (p *Parser) parseField(line string, description string) *GraphQLField {
	// field: Type
	// field: Type!
	// field: [Type!]!
	// field(arg: Type): Type
	// field(arg1: Type1, arg2: Type2 = "default"): Type

	// Remove trailing comments
	if idx := strings.Index(line, "#"); idx != -1 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)

	field := &GraphQLField{
		Description: description,
		Arguments:   []*GraphQLArgument{},
		Directives:  []*GraphQLDirective{},
	}

	// Check for arguments
	if strings.Contains(line, "(") {
		// field(args): Type
		nameEnd := strings.Index(line, "(")
		field.Name = strings.TrimSpace(line[:nameEnd])

		argsEnd := strings.Index(line, ")")
		if argsEnd == -1 {
			return nil
		}

		argsStr := line[nameEnd+1 : argsEnd]
		field.Arguments = p.parseArguments(argsStr)

		// Parse return type after colon
		rest := strings.TrimSpace(line[argsEnd+1:])
		if strings.HasPrefix(rest, ":") {
			rest = strings.TrimSpace(rest[1:])
			// Remove default value if present
			if idx := strings.Index(rest, "="); idx != -1 {
				field.DefaultValue = strings.TrimSpace(rest[idx+1:])
				rest = strings.TrimSpace(rest[:idx])
			}
			field.Type = rest
		}
	} else {
		// Simple field: name: Type
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			return nil
		}

		field.Name = strings.TrimSpace(parts[0])
		typeAndDefault := strings.TrimSpace(parts[1])

		// Check for default value
		if idx := strings.Index(typeAndDefault, "="); idx != -1 {
			field.Type = strings.TrimSpace(typeAndDefault[:idx])
			field.DefaultValue = strings.TrimSpace(typeAndDefault[idx+1:])
		} else {
			field.Type = typeAndDefault
		}
	}

	return field
}

func (p *Parser) parseArguments(argsStr string) []*GraphQLArgument {
	var args []*GraphQLArgument

	// Split by comma, but be careful with nested types like [Type!]
	var currentArg strings.Builder
	depth := 0

	for _, ch := range argsStr {
		if ch == '[' {
			depth++
		} else if ch == ']' {
			depth--
		} else if ch == ',' && depth == 0 {
			// Process current argument
			if arg := p.parseArgument(currentArg.String()); arg != nil {
				args = append(args, arg)
			}
			currentArg.Reset()
			continue
		}
		currentArg.WriteRune(ch)
	}

	// Process last argument
	if currentArg.Len() > 0 {
		if arg := p.parseArgument(currentArg.String()); arg != nil {
			args = append(args, arg)
		}
	}

	return args
}

func (p *Parser) parseArgument(argStr string) *GraphQLArgument {
	argStr = strings.TrimSpace(argStr)
	if argStr == "" {
		return nil
	}

	arg := &GraphQLArgument{
		Directives: []*GraphQLDirective{},
	}

	// arg: Type = "default"
	parts := strings.Split(argStr, ":")
	if len(parts) < 2 {
		return nil
	}

	arg.Name = strings.TrimSpace(parts[0])
	typeAndDefault := strings.TrimSpace(parts[1])

	// Check for default value
	if idx := strings.Index(typeAndDefault, "="); idx != -1 {
		arg.Type = strings.TrimSpace(typeAndDefault[:idx])
		arg.DefaultValue = strings.TrimSpace(typeAndDefault[idx+1:])
	} else {
		arg.Type = typeAndDefault
	}

	return arg
}

func (p *Parser) parseEnum(description string) (*GraphQLEnum, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	metadata := p.parseMetadataFromPreviousLine()

	// enum EnumName {
	re := regexp.MustCompile(`enum\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid enum declaration: %s", line)
	}

	enum := &GraphQLEnum{
		Name:        matches[1],
		Description: description,
		Values:      []*GraphQLEnumValue{},
		Directives:  []*GraphQLDirective{},
		Metadata:    metadata,
	}

	p.pos++ // Move past enum declaration

	// Parse enum body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// End of enum
		if line == "}" {
			p.pos++
			break
		}

		// Parse enum value with description
		valueDescription := p.parseDescription()
		if p.pos >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.pos])
		if line == "}" {
			p.pos++
			break
		}

		// Remove trailing comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		if line != "" {
			enum.Values = append(enum.Values, &GraphQLEnumValue{
				Name:        line,
				Description: valueDescription,
				Directives:  []*GraphQLDirective{},
			})
		}
		p.pos++
	}

	return enum, nil
}

func (p *Parser) parseScalar(description string) *GraphQLScalar {
	line := strings.TrimSpace(p.lines[p.pos])

	// scalar ScalarName
	re := regexp.MustCompile(`scalar\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}

	p.pos++

	return &GraphQLScalar{
		Name:        matches[1],
		Description: description,
		Directives:  []*GraphQLDirective{},
	}
}

func (p *Parser) parseInterface(description string) (*GraphQLInterface, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// interface InterfaceName {
	re := regexp.MustCompile(`interface\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid interface declaration: %s", line)
	}

	iface := &GraphQLInterface{
		Name:        matches[1],
		Description: description,
		Fields:      []*GraphQLField{},
		Directives:  []*GraphQLDirective{},
	}

	p.pos++ // Move past interface declaration

	// Parse interface body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// End of interface
		if line == "}" {
			p.pos++
			break
		}

		// Parse field with description
		fieldDescription := p.parseDescription()
		if p.pos >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.pos])
		if line == "}" {
			p.pos++
			break
		}

		field := p.parseField(line, fieldDescription)
		if field != nil {
			iface.Fields = append(iface.Fields, field)
		}
		p.pos++
	}

	return iface, nil
}

func (p *Parser) parseUnion(description string) *GraphQLUnion {
	line := strings.TrimSpace(p.lines[p.pos])

	// union UnionName = Type1 | Type2 | Type3
	re := regexp.MustCompile(`union\s+(\w+)\s*=\s*(.+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	union := &GraphQLUnion{
		Name:        matches[1],
		Description: description,
		Types:       []string{},
		Directives:  []*GraphQLDirective{},
	}

	// Parse union types
	typesStr := matches[2]
	types := strings.Split(typesStr, "|")
	for _, t := range types {
		union.Types = append(union.Types, strings.TrimSpace(t))
	}

	p.pos++
	return union
}

func (p *Parser) parseExtendQuery() ([]*GraphQLField, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	if !strings.HasPrefix(line, "extend type Query") {
		return nil, fmt.Errorf("invalid extend Query declaration: %s", line)
	}

	p.pos++ // Move past extend declaration

	var fields []*GraphQLField

	// Parse query fields
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// End of extend block
		if line == "}" {
			p.pos++
			break
		}

		// Parse field with description
		fieldDescription := p.parseDescription()
		if p.pos >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.pos])
		if line == "}" {
			p.pos++
			break
		}

		// Check if this is a multi-line field (has opening paren but no closing paren or colon)
		fieldLine := p.collectMultilineField()
		field := p.parseField(fieldLine, fieldDescription)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func (p *Parser) parseExtendMutation() ([]*GraphQLField, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	if !strings.HasPrefix(line, "extend type Mutation") {
		return nil, fmt.Errorf("invalid extend Mutation declaration: %s", line)
	}

	p.pos++ // Move past extend declaration

	var fields []*GraphQLField

	// Parse mutation fields
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// End of extend block
		if line == "}" {
			p.pos++
			break
		}

		// Parse field with description
		fieldDescription := p.parseDescription()
		if p.pos >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.pos])
		if line == "}" {
			p.pos++
			break
		}

		// Check if this is a multi-line field (has opening paren but no closing paren or colon)
		fieldLine := p.collectMultilineField()
		field := p.parseField(fieldLine, fieldDescription)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func (p *Parser) parseExtendSubscription() ([]*GraphQLField, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	if !strings.HasPrefix(line, "extend type Subscription") {
		return nil, fmt.Errorf("invalid extend Subscription declaration: %s", line)
	}

	p.pos++ // Move past extend declaration

	var fields []*GraphQLField

	// Parse subscription fields
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// End of extend block
		if line == "}" {
			p.pos++
			break
		}

		// Parse field with description
		fieldDescription := p.parseDescription()
		if p.pos >= len(p.lines) {
			break
		}

		line = strings.TrimSpace(p.lines[p.pos])
		if line == "}" {
			p.pos++
			break
		}

		// Check if this is a multi-line field (has opening paren but no closing paren or colon)
		fieldLine := p.collectMultilineField()
		field := p.parseField(fieldLine, fieldDescription)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func (p *Parser) collectMultilineField() string {
	// Collect lines until we find a complete field definition
	// A field is complete when we have a colon with the return type
	var lines []string
	startPos := p.pos
	inDescription := false

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		trimmed := strings.TrimSpace(line)

		// Handle description blocks - skip them
		if strings.HasPrefix(trimmed, `"""`) {
			if inDescription {
				// End of description
				inDescription = false
			} else {
				// Start of description
				inDescription = true
			}
			p.pos++
			continue
		}

		// Skip lines inside descriptions
		if inDescription {
			p.pos++
			continue
		}

		// Skip empty lines within the field
		if trimmed == "" {
			p.pos++
			continue
		}

		lines = append(lines, trimmed)

		// Check if this line has a colon (potential return type)
		if strings.Contains(trimmed, ":") {
			// Check if we have complete field with matching parentheses
			fullLine := strings.Join(lines, " ")
			openParens := strings.Count(fullLine, "(")
			closeParens := strings.Count(fullLine, ")")

			// Field is complete if:
			// 1. No parentheses (simple field)
			// 2. Or matching parentheses (field with args)
			if openParens == closeParens {
				p.pos++
				return fullLine
			}
		}

		p.pos++

		// Safety check: if we've gone too far, reset and return what we have
		if p.pos-startPos > 50 {
			return strings.Join(lines, " ")
		}
	}

	return strings.Join(lines, " ")
}

func (p *Parser) parseMetadataFromPreviousLine() map[string]string {
	metadata := make(map[string]string)

	// Look back at previous line for typemux:begin comment
	if p.pos > 0 {
		prevLine := strings.TrimSpace(p.lines[p.pos-1])
		if strings.Contains(prevLine, "typemux:begin") {
			// Extract the value after typemux:begin
			re := regexp.MustCompile(`#\s*typemux:begin\s+(.+)`)
			matches := re.FindStringSubmatch(prevLine)
			if len(matches) > 1 {
				metadata["typemux_type"] = strings.TrimSpace(matches[1])
			}
		}
	}

	return metadata
}
