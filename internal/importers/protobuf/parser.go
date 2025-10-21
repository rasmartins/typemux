package protobuf

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Parser struct {
	content     string
	lines       []string
	pos         int
	fileName    string
	importPaths []string
	processed   map[string]bool
}

func NewParser(content string) *Parser {
	lines := strings.Split(content, "\n")
	return &Parser{
		content:   content,
		lines:     lines,
		pos:       0,
		processed: make(map[string]bool),
	}
}

func NewParserWithImports(content string, fileName string, importPaths []string) *Parser {
	lines := strings.Split(content, "\n")
	return &Parser{
		content:     content,
		lines:       lines,
		pos:         0,
		fileName:    fileName,
		importPaths: importPaths,
		processed:   make(map[string]bool),
	}
}

func (p *Parser) Parse() (*ProtoSchema, error) {
	schema := &ProtoSchema{
		Options:  make(map[string]string),
		Messages: []*ProtoMessage{},
		Enums:    []*ProtoEnum{},
		Services: []*ProtoService{},
		Imports:  []string{},
	}

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			p.pos++
			continue
		}

		if strings.HasPrefix(line, "syntax") {
			schema.Syntax = p.parseSyntax(line)
		} else if strings.HasPrefix(line, "package") {
			schema.Package = p.parsePackage(line)
		} else if strings.HasPrefix(line, "import") {
			schema.Imports = append(schema.Imports, p.parseImport(line))
		} else if strings.HasPrefix(line, "option") {
			key, value := p.parseOption(line)
			schema.Options[key] = value
		} else if strings.HasPrefix(line, "message") {
			msg, err := p.parseMessage(0)
			if err != nil {
				return nil, err
			}
			schema.Messages = append(schema.Messages, msg)
			continue // parseMessage advances pos
		} else if strings.HasPrefix(line, "enum") {
			enum, err := p.parseEnum(0)
			if err != nil {
				return nil, err
			}
			schema.Enums = append(schema.Enums, enum)
			continue // parseEnum advances pos
		} else if strings.HasPrefix(line, "service") {
			service, err := p.parseService()
			if err != nil {
				return nil, err
			}
			schema.Services = append(schema.Services, service)
			continue // parseService advances pos
		}

		p.pos++
	}

	return schema, nil
}

func (p *Parser) ParseWithImports() (map[string]*ProtoSchema, error) {
	schemas := make(map[string]*ProtoSchema)

	// Mark this file as processed
	if p.fileName != "" {
		p.processed[p.fileName] = true
	}

	// Parse the main file
	schema, err := p.Parse()
	if err != nil {
		return nil, err
	}

	// Store the main schema
	schemas[p.fileName] = schema

	// Process imports recursively
	if len(p.importPaths) > 0 {
		for _, importPath := range schema.Imports {
			// Skip well-known types
			if p.isWellKnownType(importPath) {
				continue
			}

			// Try to resolve and parse the import
			importedSchemas, err := p.resolveAndParseImport(importPath)
			if err != nil {
				// If we can't find the import, just log and continue
				// (it might be a system proto that we don't need to include)
				fmt.Printf("Warning: could not resolve import %s: %v\n", importPath, err)
				continue
			}

			// Add all imported schemas to our map
			for path, importedSchema := range importedSchemas {
				if _, exists := schemas[path]; !exists {
					schemas[path] = importedSchema
				}
			}
		}
	}

	return schemas, nil
}

func (p *Parser) isWellKnownType(importPath string) bool {
	wellKnown := []string{
		"google/protobuf/",
		"google/api/",
	}

	for _, prefix := range wellKnown {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

func (p *Parser) resolveAndParseImport(importPath string) (map[string]*ProtoSchema, error) {
	// Try each import path
	for _, searchPath := range p.importPaths {
		fullPath := filepath.Join(searchPath, importPath)

		// Check if already processed
		if p.processed[fullPath] {
			return make(map[string]*ProtoSchema), nil // Already processed, return empty map
		}

		// Try to read the file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue // Try next search path
		}

		// Mark as processed
		p.processed[fullPath] = true

		// Parse the imported file
		importParser := &Parser{
			content:     string(content),
			lines:       strings.Split(string(content), "\n"),
			pos:         0,
			fileName:    fullPath,
			importPaths: p.importPaths,
			processed:   p.processed, // Share the processed map
		}

		return importParser.ParseWithImports()
	}

	return nil, fmt.Errorf("import not found in any search path: %s", importPath)
}

func (p *Parser) parseSyntax(line string) string {
	// syntax = "proto3";
	re := regexp.MustCompile(`syntax\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return "proto3"
}

func (p *Parser) parsePackage(line string) string {
	// package mypackage;
	re := regexp.MustCompile(`package\s+([^;]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (p *Parser) parseImport(line string) string {
	// import "path/to/file.proto";
	re := regexp.MustCompile(`import\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *Parser) parseOption(line string) (string, string) {
	// option go_package = "...";
	re := regexp.MustCompile(`option\s+([^\s=]+)\s*=\s*"?([^";]+)"?`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 2 {
		return matches[1], strings.Trim(matches[2], `"`)
	}
	return "", ""
}

func (p *Parser) parseMessage(indent int) (*ProtoMessage, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// message MessageName {
	re := regexp.MustCompile(`message\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid message declaration: %s", line)
	}

	msg := &ProtoMessage{
		Name:     matches[1],
		Fields:   []*ProtoField{},
		Enums:    []*ProtoEnum{},
		Messages: []*ProtoMessage{},
		Options:  make(map[string]string),
		Reserved: []string{},
		OneOfs:   []*ProtoOneOf{},
	}

	p.pos++ // Move past message declaration

	// Parse message body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			p.pos++
			continue
		}

		// End of message
		if line == "}" {
			p.pos++
			break
		}

		// Nested message
		if strings.HasPrefix(line, "message") {
			nested, err := p.parseMessage(indent + 1)
			if err != nil {
				return nil, err
			}
			msg.Messages = append(msg.Messages, nested)
			continue
		}

		// Nested enum
		if strings.HasPrefix(line, "enum") {
			enum, err := p.parseEnum(indent + 1)
			if err != nil {
				return nil, err
			}
			msg.Enums = append(msg.Enums, enum)
			continue
		}

		// Reserved fields
		if strings.HasPrefix(line, "reserved") {
			reserved := p.parseReserved(line)
			msg.Reserved = append(msg.Reserved, reserved...)
			p.pos++
			continue
		}

		// OneOf
		if strings.HasPrefix(line, "oneof") {
			oneof, err := p.parseOneOf()
			if err != nil {
				return nil, err
			}
			msg.OneOfs = append(msg.OneOfs, oneof)
			continue
		}

		// Regular field
		field, err := p.parseField(line)
		if err != nil {
			// Skip lines we can't parse
			p.pos++
			continue
		}
		msg.Fields = append(msg.Fields, field)
		p.pos++
	}

	return msg, nil
}

func (p *Parser) parseField(line string) (*ProtoField, error) {
	field := &ProtoField{}

	// Check for deprecated
	if strings.Contains(line, "[deprecated = true]") {
		field.Deprecated = true
	}

	// Check for optional/repeated
	if strings.HasPrefix(line, "optional ") {
		field.Optional = true
		line = strings.TrimPrefix(line, "optional ")
	} else if strings.HasPrefix(line, "repeated ") {
		field.Repeated = true
		line = strings.TrimPrefix(line, "repeated ")
	}

	// Parse field: type name = number;
	// More complex: map<string, Type> name = number;

	// Try to match map<K,V> type first
	mapRe := regexp.MustCompile(`^\s*(map<[^>]+>)\s+(\w+)\s*=\s*(\d+)`)
	mapMatches := mapRe.FindStringSubmatch(line)
	if len(mapMatches) >= 4 {
		field.Type = mapMatches[1]
		field.Name = mapMatches[2]
		number, _ := strconv.Atoi(mapMatches[3])
		field.Number = number
		return field, nil
	}

	// Regular field
	re := regexp.MustCompile(`^\s*(\S+)\s+(\w+)\s*=\s*(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid field: %s", line)
	}

	field.Type = matches[1]
	field.Name = matches[2]
	number, _ := strconv.Atoi(matches[3])
	field.Number = number

	return field, nil
}

func (p *Parser) parseEnum(indent int) (*ProtoEnum, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// enum EnumName {
	re := regexp.MustCompile(`enum\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid enum declaration: %s", line)
	}

	enum := &ProtoEnum{
		Name:   matches[1],
		Values: []*ProtoEnumValue{},
	}

	p.pos++ // Move past enum declaration

	// Parse enum body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			p.pos++
			continue
		}

		// End of enum
		if line == "}" {
			p.pos++
			break
		}

		// Parse enum value: NAME = number;
		re := regexp.MustCompile(`^(\w+)\s*=\s*(\d+)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			number, _ := strconv.Atoi(matches[2])
			enum.Values = append(enum.Values, &ProtoEnumValue{
				Name:   matches[1],
				Number: number,
			})
		}

		p.pos++
	}

	return enum, nil
}

func (p *Parser) parseService() (*ProtoService, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// service ServiceName {
	re := regexp.MustCompile(`service\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid service declaration: %s", line)
	}

	service := &ProtoService{
		Name:    matches[1],
		Methods: []*ProtoMethod{},
	}

	p.pos++ // Move past service declaration

	// Parse service body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			p.pos++
			continue
		}

		// End of service
		if line == "}" {
			p.pos++
			break
		}

		// Parse RPC method (may span multiple lines)
		if strings.HasPrefix(line, "rpc") {
			method := p.parseMultilineMethod()
			if method != nil {
				service.Methods = append(service.Methods, method)
			}
			continue
		}

		p.pos++
	}

	return service, nil
}

func (p *Parser) parseMethod(line string) *ProtoMethod {
	// rpc MethodName(RequestType) returns (ResponseType);
	// rpc MethodName(stream RequestType) returns (stream ResponseType);
	re := regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(stream\s+)?(\w+)\s*\)\s*returns\s*\(\s*(stream\s+)?(\w+)\s*\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 6 {
		return nil
	}

	return &ProtoMethod{
		Name:         matches[1],
		InputType:    matches[3],
		OutputType:   matches[5],
		ClientStream: matches[2] != "",
		ServerStream: matches[4] != "",
	}
}

func (p *Parser) parseMultilineMethod() *ProtoMethod {
	// Collect lines until we find the semicolon or closing brace
	var methodLines []string
	startPos := p.pos

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		methodLines = append(methodLines, line)
		p.pos++

		// Check if this line completes the method (has ; or })
		if strings.Contains(line, ";") || strings.Contains(line, "}") {
			break
		}
	}

	// Join all lines and parse
	fullMethod := strings.Join(methodLines, " ")
	method := p.parseMethod(fullMethod)

	// If parsing failed, reset position
	if method == nil {
		p.pos = startPos + 1
	}

	return method
}

func (p *Parser) parseReserved(line string) []string {
	// reserved 2, 3, 4, 5;
	// reserved "field_name";
	var reserved []string

	line = strings.TrimPrefix(line, "reserved")
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)

	parts := strings.Split(line, ",")
	for _, part := range parts {
		reserved = append(reserved, strings.TrimSpace(part))
	}

	return reserved
}

func (p *Parser) parseOneOf() (*ProtoOneOf, error) {
	line := strings.TrimSpace(p.lines[p.pos])

	// oneof name {
	re := regexp.MustCompile(`oneof\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid oneof declaration: %s", line)
	}

	oneof := &ProtoOneOf{
		Name:   matches[1],
		Fields: []*ProtoField{},
	}

	p.pos++ // Move past oneof declaration

	// Parse oneof body
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			p.pos++
			continue
		}

		// End of oneof
		if line == "}" {
			p.pos++
			break
		}

		// Parse field
		field, err := p.parseField(line)
		if err == nil {
			oneof.Fields = append(oneof.Fields, field)
		}

		p.pos++
	}

	return oneof, nil
}
