package typemux

import (
	"fmt"
	"sort"

	"github.com/rasmartins/typemux/internal/generator"
)

// Generator is the interface that all code generators must implement.
// Custom generators can be registered with the GeneratorFactory.
type Generator interface {
	// Generate produces output from a schema
	Generate(schema *Schema) (string, error)

	// Format returns the generator's format identifier (e.g., "graphql", "protobuf")
	Format() string

	// FileExtension returns the file extension for generated output (e.g., ".graphql", ".proto")
	FileExtension() string
}

// ConfigurableGenerator is an optional interface for generators that support
// configuration options.
type ConfigurableGenerator interface {
	Generator

	// GenerateWithConfig generates output using format-specific configuration
	GenerateWithConfig(schema *Schema, config map[string]interface{}) (string, error)
}

// GeneratorFactory manages generator registration and lookup.
// It provides built-in generators for GraphQL, Protobuf, OpenAPI, and Go,
// and allows registration of custom generators.
type GeneratorFactory struct {
	generators map[string]Generator
}

// NewGeneratorFactory creates a factory with all built-in generators pre-registered.
// Built-in generators include: graphql, protobuf (proto), openapi, go (golang).
func NewGeneratorFactory() *GeneratorFactory {
	factory := &GeneratorFactory{
		generators: make(map[string]Generator),
	}

	// Register built-in generators
	factory.Register(&builtinGraphQLGenerator{})
	factory.Register(&builtinProtobufGenerator{})
	factory.Register(&builtinOpenAPIGenerator{})
	factory.Register(&builtinGoGenerator{})

	return factory
}

// Register adds or replaces a generator in the factory.
// If a generator with the same format already exists, it will be replaced.
//
// Example:
//
//	type CustomGenerator struct{}
//	func (g *CustomGenerator) Generate(schema *typemux.Schema) (string, error) { ... }
//	func (g *CustomGenerator) Format() string { return "custom" }
//	func (g *CustomGenerator) FileExtension() string { return ".custom" }
//
//	factory := typemux.NewGeneratorFactory()
//	factory.Register(&CustomGenerator{})
func (f *GeneratorFactory) Register(gen Generator) {
	f.generators[gen.Format()] = gen

	// Also register common aliases
	if gen.Format() == "protobuf" {
		f.generators["proto"] = gen
	} else if gen.Format() == "go" {
		f.generators["golang"] = gen
	}
}

// Unregister removes a generator from the factory.
func (f *GeneratorFactory) Unregister(format string) {
	delete(f.generators, format)

	// Also remove aliases
	if format == "protobuf" {
		delete(f.generators, "proto")
	} else if format == "proto" {
		delete(f.generators, "protobuf")
	} else if format == "go" {
		delete(f.generators, "golang")
	} else if format == "golang" {
		delete(f.generators, "go")
	}
}

// Get retrieves a generator by format name.
// Returns an error if no generator is registered for the format.
func (f *GeneratorFactory) Get(format string) (Generator, error) {
	gen, ok := f.generators[format]
	if !ok {
		return nil, fmt.Errorf("no generator registered for format: %s", format)
	}
	return gen, nil
}

// Generate generates output for the specified format.
//
// Example:
//
//	factory := typemux.NewGeneratorFactory()
//	graphql, err := factory.Generate("graphql", schema)
//	protobuf, err := factory.Generate("protobuf", schema)
func (f *GeneratorFactory) Generate(format string, schema *Schema) (string, error) {
	gen, err := f.Get(format)
	if err != nil {
		return "", err
	}
	return gen.Generate(schema)
}

// GenerateAll generates output for all registered formats.
// Returns a map of format name to generated content.
//
// Example:
//
//	outputs, err := factory.GenerateAll(schema)
//	for format, content := range outputs {
//	    fmt.Printf("=== %s ===\n%s\n", format, content)
//	}
func (f *GeneratorFactory) GenerateAll(schema *Schema) (map[string]string, error) {
	outputs := make(map[string]string)
	seen := make(map[string]bool)

	for format, gen := range f.generators {
		// Skip aliases (proto/protobuf, go/golang)
		if seen[gen.Format()] {
			continue
		}
		seen[gen.Format()] = true

		output, err := gen.Generate(schema)
		if err != nil {
			return nil, fmt.Errorf("error generating %s: %w", format, err)
		}
		outputs[gen.Format()] = output
	}

	return outputs, nil
}

// GetFormats returns a sorted list of all registered format names.
func (f *GeneratorFactory) GetFormats() []string {
	seen := make(map[string]bool)
	formats := make([]string, 0, len(f.generators))

	for _, gen := range f.generators {
		primaryFormat := gen.Format()
		if !seen[primaryFormat] {
			seen[primaryFormat] = true
			formats = append(formats, primaryFormat)
		}
	}

	sort.Strings(formats)
	return formats
}

// HasFormat checks if a generator is registered for the given format.
func (f *GeneratorFactory) HasFormat(format string) bool {
	_, ok := f.generators[format]
	return ok
}

// GenerateWithConfig generates output using configuration options.
// This is the main method used when processing from a Config object.
func (f *GeneratorFactory) GenerateWithConfig(config *Config) (map[string]string, error) {
	// Parse schema with annotations
	schema, err := ParseWithAnnotations(config.Input.Schema, config.Input.Annotations...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	outputs := make(map[string]string)

	// Generate requested formats
	for _, formatName := range config.Output.Formats {
		// Handle "all" format
		if formatName == "all" {
			allOutputs, err := f.GenerateAll(schema)
			if err != nil {
				return nil, err
			}
			for k, v := range allOutputs {
				outputs[k] = v
			}
			continue
		}

		// Get the generator
		gen, err := f.Get(formatName)
		if err != nil {
			return nil, err
		}

		// Check if generator supports configuration
		var output string
		if configurable, ok := gen.(ConfigurableGenerator); ok {
			genConfig := config.getGeneratorConfig(formatName)
			output, err = configurable.GenerateWithConfig(schema, genConfig)
		} else {
			output, err = gen.Generate(schema)
		}

		if err != nil {
			return nil, fmt.Errorf("error generating %s: %w", formatName, err)
		}

		outputs[gen.Format()] = output
	}

	return outputs, nil
}

// Built-in generator wrappers

type builtinGraphQLGenerator struct{}

func (g *builtinGraphQLGenerator) Generate(schema *Schema) (string, error) {
	gen := generator.NewGraphQLGenerator()
	return gen.Generate(schema), nil
}

func (g *builtinGraphQLGenerator) Format() string {
	return "graphql"
}

func (g *builtinGraphQLGenerator) FileExtension() string {
	return ".graphql"
}

type builtinProtobufGenerator struct{}

func (g *builtinProtobufGenerator) Generate(schema *Schema) (string, error) {
	gen := generator.NewProtobufGenerator()
	return gen.Generate(schema), nil
}

func (g *builtinProtobufGenerator) Format() string {
	return "protobuf"
}

func (g *builtinProtobufGenerator) FileExtension() string {
	return ".proto"
}

type builtinOpenAPIGenerator struct{}

func (g *builtinOpenAPIGenerator) Generate(schema *Schema) (string, error) {
	gen := generator.NewOpenAPIGenerator()
	return gen.Generate(schema), nil
}

func (g *builtinOpenAPIGenerator) Format() string {
	return "openapi"
}

func (g *builtinOpenAPIGenerator) FileExtension() string {
	return ".yaml"
}

type builtinGoGenerator struct{}

func (g *builtinGoGenerator) Generate(schema *Schema) (string, error) {
	gen := generator.NewGoGenerator()
	return gen.Generate(schema), nil
}

func (g *builtinGoGenerator) Format() string {
	return "go"
}

func (g *builtinGoGenerator) FileExtension() string {
	return ".go"
}
