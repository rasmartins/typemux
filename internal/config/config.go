package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the TypeMUX configuration
type Config struct {
	// TypeMUX version this config is compatible with
	Version string `yaml:"version"`

	// Input configuration
	Input InputConfig `yaml:"input"`

	// Output configuration
	Output OutputConfig `yaml:"output"`

	// Generator-specific settings
	Generators GeneratorConfig `yaml:"generators,omitempty"`
}

// InputConfig defines input sources
type InputConfig struct {
	// Main schema file (required)
	Schema string `yaml:"schema"`

	// Additional annotation files
	Annotations []string `yaml:"annotations,omitempty"`
}

// OutputConfig defines output settings
type OutputConfig struct {
	// Output directory (default: ./generated)
	Directory string `yaml:"directory"`

	// Formats to generate (graphql, protobuf, openapi, or all)
	Formats []string `yaml:"formats"`

	// Clean output directory before generation
	Clean bool `yaml:"clean,omitempty"`
}

// GeneratorConfig holds generator-specific configurations
type GeneratorConfig struct {
	// GraphQL-specific settings
	GraphQL *GraphQLConfig `yaml:"graphql,omitempty"`

	// Protobuf-specific settings
	Protobuf *ProtobufConfig `yaml:"protobuf,omitempty"`

	// OpenAPI-specific settings
	OpenAPI *OpenAPIConfig `yaml:"openapi,omitempty"`
}

// GraphQLConfig holds GraphQL generator settings
type GraphQLConfig struct {
	// Output filename (default: schema.graphql)
	Filename string `yaml:"filename,omitempty"`
}

// ProtobufConfig holds Protobuf generator settings
type ProtobufConfig struct {
	// Output filename for single namespace (default: schema.proto)
	Filename string `yaml:"filename,omitempty"`

	// Import buf validate for validation rules
	ImportBufValidate bool `yaml:"import_buf_validate,omitempty"`
}

// OpenAPIConfig holds OpenAPI generator settings
type OpenAPIConfig struct {
	// Output filename (default: openapi.yaml)
	Filename string `yaml:"filename,omitempty"`

	// OpenAPI version (default: 3.0.0)
	Version string `yaml:"version,omitempty"`
}

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Resolve relative paths
	configDir := filepath.Dir(path)
	if err := config.ResolvePaths(configDir); err != nil {
		return nil, err
	}

	// Apply defaults
	config.ApplyDefaults()

	return &config, nil
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	// Check required fields
	if c.Input.Schema == "" {
		return fmt.Errorf("input.schema is required")
	}

	if len(c.Output.Formats) == 0 {
		return fmt.Errorf("output.formats must specify at least one format")
	}

	// Validate format names
	validFormats := map[string]bool{
		"graphql":  true,
		"protobuf": true,
		"proto":    true,
		"openapi":  true,
		"all":      true,
	}

	for _, format := range c.Output.Formats {
		if !validFormats[format] {
			return fmt.Errorf("invalid format: %s (must be graphql, protobuf, openapi, or all)", format)
		}
	}

	return nil
}

// ResolvePaths converts relative paths to absolute paths based on config file location
func (c *Config) ResolvePaths(configDir string) error {
	// Resolve schema path
	if !filepath.IsAbs(c.Input.Schema) {
		c.Input.Schema = filepath.Join(configDir, c.Input.Schema)
	}

	// Resolve annotation paths
	for i, ann := range c.Input.Annotations {
		if !filepath.IsAbs(ann) {
			c.Input.Annotations[i] = filepath.Join(configDir, ann)
		}
	}

	// Resolve output directory
	if c.Output.Directory != "" && !filepath.IsAbs(c.Output.Directory) {
		c.Output.Directory = filepath.Join(configDir, c.Output.Directory)
	}

	return nil
}

// ApplyDefaults sets default values for optional fields
func (c *Config) ApplyDefaults() {
	// Default output directory
	if c.Output.Directory == "" {
		c.Output.Directory = "./generated"
	}

	// Generator defaults
	if c.Generators.GraphQL != nil && c.Generators.GraphQL.Filename == "" {
		c.Generators.GraphQL.Filename = "schema.graphql"
	}

	if c.Generators.Protobuf != nil && c.Generators.Protobuf.Filename == "" {
		c.Generators.Protobuf.Filename = "schema.proto"
	}

	if c.Generators.OpenAPI != nil {
		if c.Generators.OpenAPI.Filename == "" {
			c.Generators.OpenAPI.Filename = "openapi.yaml"
		}
		if c.Generators.OpenAPI.Version == "" {
			c.Generators.OpenAPI.Version = "3.0.0"
		}
	}
}

// ShouldGenerateFormat checks if a specific format should be generated
func (c *Config) ShouldGenerateFormat(format string) bool {
	for _, f := range c.Output.Formats {
		if f == "all" || f == format || (f == "proto" && format == "protobuf") || (f == "golang" && format == "go") {
			return true
		}
	}
	return false
}
