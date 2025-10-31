package typemux

import (
	"fmt"

	"github.com/rasmartins/typemux/internal/importers/graphql"
	"github.com/rasmartins/typemux/internal/importers/openapi"
	"github.com/rasmartins/typemux/internal/importers/protobuf"
)

// Importer is the interface for converting external schema formats to TypeMUX IDL.
// Custom importers can be registered with the ImporterFactory.
type Importer interface {
	// Import converts an external format to TypeMUX IDL
	Import(content string) (string, error)

	// Format returns the source format name (e.g., "protobuf", "graphql", "openapi")
	Format() string
}

// MultiFileImporter is an optional interface for importers that can handle
// multiple files with imports (like Protobuf).
type MultiFileImporter interface {
	Importer

	// ImportWithPaths converts multiple related files to TypeMUX IDL
	// Returns a map of file paths to TypeMUX IDL content
	ImportWithPaths(content string, importPaths []string) (map[string]string, error)
}

// ImporterFactory manages importer registration and lookup.
// It provides built-in importers for Protobuf, GraphQL, and OpenAPI.
type ImporterFactory struct {
	importers map[string]Importer
}

// NewImporterFactory creates a factory with all built-in importers pre-registered.
// Built-in importers include: protobuf, graphql, openapi.
func NewImporterFactory() *ImporterFactory {
	factory := &ImporterFactory{
		importers: make(map[string]Importer),
	}

	// Register built-in importers
	factory.Register(&builtinProtobufImporter{})
	factory.Register(&builtinGraphQLImporter{})
	factory.Register(&builtinOpenAPIImporter{})

	return factory
}

// Register adds or replaces an importer in the factory.
//
// Example:
//
//	type JSONSchemaImporter struct{}
//	func (i *JSONSchemaImporter) Import(content string) (string, error) { ... }
//	func (i *JSONSchemaImporter) Format() string { return "jsonschema" }
//
//	factory := typemux.NewImporterFactory()
//	factory.Register(&JSONSchemaImporter{})
func (f *ImporterFactory) Register(importer Importer) {
	f.importers[importer.Format()] = importer

	// Register common aliases
	if importer.Format() == "protobuf" {
		f.importers["proto"] = importer
	}
}

// Unregister removes an importer from the factory.
func (f *ImporterFactory) Unregister(format string) {
	delete(f.importers, format)

	// Also remove aliases
	if format == "protobuf" {
		delete(f.importers, "proto")
	} else if format == "proto" {
		delete(f.importers, "protobuf")
	}
}

// Get retrieves an importer by format name.
func (f *ImporterFactory) Get(format string) (Importer, error) {
	imp, ok := f.importers[format]
	if !ok {
		return nil, fmt.Errorf("no importer registered for format: %s", format)
	}
	return imp, nil
}

// Import converts an external schema format to TypeMUX IDL.
//
// Example:
//
//	factory := typemux.NewImporterFactory()
//	typemuxIDL, err := factory.Import("graphql", graphqlSchema)
//	typemuxIDL, err := factory.Import("protobuf", protoFile)
func (f *ImporterFactory) Import(format string, content string) (string, error) {
	imp, err := f.Get(format)
	if err != nil {
		return "", err
	}
	return imp.Import(content)
}

// ImportProtobuf converts a Protobuf schema to TypeMUX IDL.
// For simple single-file conversion.
//
// Example:
//
//	factory := typemux.NewImporterFactory()
//	typemuxIDL, err := factory.ImportProtobuf(protoContent)
func (f *ImporterFactory) ImportProtobuf(content string) (string, error) {
	return f.Import("protobuf", content)
}

// ImportProtobufWithPaths converts Protobuf schemas with import resolution.
// Returns a map of proto file paths to TypeMUX IDL content.
//
// Example:
//
//	factory := typemux.NewImporterFactory()
//	schemas, err := factory.ImportProtobufWithPaths(protoContent, []string{"./protos", "/usr/include"})
//	for path, idl := range schemas {
//	    fmt.Printf("=== %s ===\n%s\n", path, idl)
//	}
func (f *ImporterFactory) ImportProtobufWithPaths(content string, importPaths []string) (map[string]string, error) {
	imp, err := f.Get("protobuf")
	if err != nil {
		return nil, err
	}

	multiFileImp, ok := imp.(MultiFileImporter)
	if !ok {
		return nil, fmt.Errorf("protobuf importer does not support multi-file imports")
	}

	return multiFileImp.ImportWithPaths(content, importPaths)
}

// ImportGraphQL converts a GraphQL schema to TypeMUX IDL.
//
// Example:
//
//	factory := typemux.NewImporterFactory()
//	typemuxIDL, err := factory.ImportGraphQL(graphqlSchema)
func (f *ImporterFactory) ImportGraphQL(content string) (string, error) {
	return f.Import("graphql", content)
}

// ImportOpenAPI converts an OpenAPI specification to TypeMUX IDL.
//
// Example:
//
//	factory := typemux.NewImporterFactory()
//	typemuxIDL, err := factory.ImportOpenAPI(openapiYAML)
func (f *ImporterFactory) ImportOpenAPI(content string) (string, error) {
	return f.Import("openapi", content)
}

// HasFormat checks if an importer is registered for the given format.
func (f *ImporterFactory) HasFormat(format string) bool {
	_, ok := f.importers[format]
	return ok
}

// GetFormats returns a list of all registered format names.
func (f *ImporterFactory) GetFormats() []string {
	seen := make(map[string]bool)
	formats := make([]string, 0, len(f.importers))

	for _, imp := range f.importers {
		primaryFormat := imp.Format()
		if !seen[primaryFormat] {
			seen[primaryFormat] = true
			formats = append(formats, primaryFormat)
		}
	}

	return formats
}

// Built-in importer implementations

type builtinProtobufImporter struct{}

func (i *builtinProtobufImporter) Import(content string) (string, error) {
	parser := protobuf.NewParser(content)
	schema, err := parser.Parse()
	if err != nil {
		return "", fmt.Errorf("failed to parse protobuf: %w", err)
	}

	converter := protobuf.NewConverter()
	return converter.Convert(schema), nil
}

func (i *builtinProtobufImporter) ImportWithPaths(content string, importPaths []string) (map[string]string, error) {
	// Use empty filename since we're working with content string
	parser := protobuf.NewParserWithImports(content, "", importPaths)
	schemas, err := parser.ParseWithImports()
	if err != nil {
		return nil, fmt.Errorf("failed to parse protobuf with imports: %w", err)
	}

	converter := protobuf.NewConverter()
	results := make(map[string]string)

	for path, schema := range schemas {
		results[path] = converter.Convert(schema)
	}

	return results, nil
}

func (i *builtinProtobufImporter) Format() string {
	return "protobuf"
}

type builtinGraphQLImporter struct{}

func (i *builtinGraphQLImporter) Import(content string) (string, error) {
	parser := graphql.NewParser(content)
	schema, err := parser.Parse()
	if err != nil {
		return "", fmt.Errorf("failed to parse GraphQL: %w", err)
	}

	converter := graphql.NewConverter()
	return converter.Convert(schema), nil
}

func (i *builtinGraphQLImporter) Format() string {
	return "graphql"
}

type builtinOpenAPIImporter struct{}

func (i *builtinOpenAPIImporter) Import(content string) (string, error) {
	parser := openapi.NewParser([]byte(content))
	spec, err := parser.Parse()
	if err != nil {
		return "", fmt.Errorf("failed to parse OpenAPI: %w", err)
	}

	converter := openapi.NewConverter()
	return converter.Convert(spec), nil
}

func (i *builtinOpenAPIImporter) Format() string {
	return "openapi"
}
