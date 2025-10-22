package generator

import (
	"fmt"
	"strings"

	"github.com/rasmartins/typemux/internal/ast"
	"gopkg.in/yaml.v3"
)

// GqlgenConfigGenerator generates gqlgen configuration files
type GqlgenConfigGenerator struct{}

// NewGqlgenConfigGenerator creates a new gqlgen config generator
func NewGqlgenConfigGenerator() *GqlgenConfigGenerator {
	return &GqlgenConfigGenerator{}
}

// GqlgenConfig represents the gqlgen.yml configuration structure
type GqlgenConfig struct {
	Schema   []string            `yaml:"schema"`
	Exec     ExecConfig          `yaml:"exec"`
	Model    ModelConfig         `yaml:"model"`
	Resolver ResolverConfig      `yaml:"resolver,omitempty"`
	Autobind []string            `yaml:"autobind,omitempty"`
	Models   map[string]ModelDef `yaml:"models,omitempty"`
	Structs  map[string]ModelDef `yaml:"structs,omitempty"`
}

// ExecConfig defines the exec package configuration
type ExecConfig struct {
	Filename string `yaml:"filename"`
	Package  string `yaml:"package,omitempty"`
}

// ModelConfig defines the model package configuration
type ModelConfig struct {
	Filename string `yaml:"filename"`
	Package  string `yaml:"package,omitempty"`
}

// ResolverConfig defines the resolver package configuration
type ResolverConfig struct {
	Filename string `yaml:"filename,omitempty"`
	Package  string `yaml:"package,omitempty"`
	Type     string `yaml:"type,omitempty"`
}

// ModelDef defines a model mapping
type ModelDef struct {
	Model  []string          `yaml:"model,omitempty"`
	Fields map[string]string `yaml:"fields,omitempty"`
}

// GqlgenOptions contains options for gqlgen config generation
type GqlgenOptions struct {
	SchemaFiles   []string // GraphQL schema file paths
	ModelsPackage string   // Go package path for models (e.g., "github.com/user/project/models")
	ExecPackage   string   // Package for generated exec code
	ResolverType  string   // Resolver struct type name
	GenerateStubs bool     // Whether to generate resolver stubs
}

// Generate creates a gqlgen configuration from a TypeMUX schema
func (g *GqlgenConfigGenerator) Generate(schema *ast.Schema, opts *GqlgenOptions) (string, error) {
	if opts == nil {
		opts = &GqlgenOptions{
			SchemaFiles:   []string{"schema.graphql"},
			ModelsPackage: "models",
			ExecPackage:   "generated",
		}
	}

	config := GqlgenConfig{
		Schema: opts.SchemaFiles,
		Exec: ExecConfig{
			Filename: "generated/exec.go",
			Package:  opts.ExecPackage,
		},
		Model: ModelConfig{
			Filename: "models_gen.go",
		},
		Models:  g.generateModelMappings(schema, opts),
		Structs: g.generateStructMappings(schema, opts),
	}

	// Add autobind if models package is specified
	if opts.ModelsPackage != "" {
		config.Autobind = []string{opts.ModelsPackage}
	}

	// Add resolver config if requested
	if opts.GenerateStubs && opts.ResolverType != "" {
		config.Resolver = ResolverConfig{
			Filename: "resolver.go",
			Type:     opts.ResolverType,
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gqlgen config: %w", err)
	}

	return string(data), nil
}

// generateModelMappings creates model mappings for TypeMUX types
func (g *GqlgenConfigGenerator) generateModelMappings(schema *ast.Schema, opts *GqlgenOptions) map[string]ModelDef {
	models := make(map[string]ModelDef)

	// Map TypeMUX types to Go models
	for _, typ := range schema.Types {
		// Map to Go model in the models package
		modelPath := fmt.Sprintf("%s.%s", opts.ModelsPackage, typ.Name)
		models[typ.Name] = ModelDef{
			Model: []string{modelPath},
		}
	}

	// Unions are let gqlgen generate them - they need Go interfaces
	// We don't map them here

	// Map enums to Go enums
	for _, enum := range schema.Enums {
		modelPath := fmt.Sprintf("%s.%s", opts.ModelsPackage, enum.Name)
		models[enum.Name] = ModelDef{
			Model: []string{modelPath},
		}
	}

	// Add scalar mappings for TypeMUX built-in types
	models["Timestamp"] = ModelDef{
		Model: []string{"time.Time"},
	}

	models["Bytes"] = ModelDef{
		Model: []string{"[]byte"},
	}

	return models
}

// generateStructMappings creates struct mappings for input types
func (g *GqlgenConfigGenerator) generateStructMappings(schema *ast.Schema, opts *GqlgenOptions) map[string]ModelDef {
	structs := make(map[string]ModelDef)

	// Look for input types (types used as method parameters)
	inputTypes := g.findInputTypes(schema)

	for typeName := range inputTypes {
		modelPath := fmt.Sprintf("%s.%s", opts.ModelsPackage, typeName)
		structs[typeName] = ModelDef{
			Model: []string{modelPath},
		}
	}

	return structs
}

// findInputTypes identifies types that are used as service method inputs
func (g *GqlgenConfigGenerator) findInputTypes(schema *ast.Schema) map[string]bool {
	inputTypes := make(map[string]bool)

	for _, service := range schema.Services {
		for _, method := range service.Methods {
			// Input parameter type
			if method.InputType != "" && !g.isPrimitiveType(method.InputType) {
				inputTypes[method.InputType] = true
			}
		}
	}

	return inputTypes
}

// isPrimitiveType checks if a type is a primitive type
func (g *GqlgenConfigGenerator) isPrimitiveType(typeName string) bool {
	primitives := map[string]bool{
		"string":    true,
		"int32":     true,
		"int64":     true,
		"float32":   true,
		"float64":   true,
		"bool":      true,
		"bytes":     true,
		"timestamp": true,
	}

	// Handle arrays and maps
	if strings.HasPrefix(typeName, "[]") || strings.HasPrefix(typeName, "map<") {
		return false
	}

	return primitives[strings.ToLower(typeName)]
}
