package typemux

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents TypeMUX configuration for code generation.
type Config struct {
	// Version of the TypeMUX config format
	Version string

	// Input configuration
	Input InputConfig

	// Output configuration
	Output OutputConfig

	// Generator-specific settings
	Generators GeneratorConfig

	// Custom generator options (for user-registered generators)
	CustomGenerators map[string]map[string]interface{}
}

// InputConfig defines input sources for schema and annotations.
type InputConfig struct {
	// Schema is the TypeMUX IDL content or file path
	Schema string

	// Annotations are optional YAML annotation strings or file paths
	Annotations []string

	// BaseDir is the base directory for resolving relative imports
	BaseDir string
}

// OutputConfig defines output settings for generated code.
type OutputConfig struct {
	// Directory for generated output files
	Directory string

	// Formats to generate (e.g., "graphql", "protobuf", "openapi", "go", "all")
	Formats []string

	// Clean the output directory before generation
	Clean bool

	// Custom filenames for each format (format -> filename)
	Filenames map[string]string
}

// GeneratorConfig holds settings for built-in generators.
type GeneratorConfig struct {
	GraphQL  *GraphQLConfig
	Protobuf *ProtobufConfig
	OpenAPI  *OpenAPIConfig
	Go       *GoConfig
}

// GraphQLConfig configures the GraphQL generator.
type GraphQLConfig struct {
	Filename          string
	IncludeDeprecated bool
}

// ProtobufConfig configures the Protobuf generator.
type ProtobufConfig struct {
	Filename          string
	ImportBufValidate bool
	PackagePrefix     string
}

// OpenAPIConfig configures the OpenAPI generator.
type OpenAPIConfig struct {
	Filename string
	Version  string // e.g., "3.0.0", "3.1.0"
}

// GoConfig configures the Go generator.
type GoConfig struct {
	Filename     string
	PackageName  string
	JSONTags     bool
	ValidateTags bool
}

// NewConfig creates a new configuration with default values.
func NewConfig() *Config {
	config := &Config{
		Version: Version,
		Input:   InputConfig{},
		Output: OutputConfig{
			Directory: "./generated",
			Formats:   []string{"all"},
			Filenames: make(map[string]string),
		},
		Generators:       GeneratorConfig{},
		CustomGenerators: make(map[string]map[string]interface{}),
	}
	config.ApplyDefaults()
	return config
}

// LoadConfig reads and parses a configuration from a YAML file.
//
// Example:
//
//	config, err := typemux.LoadConfig("typemux.config.yaml")
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return LoadConfigFromBytes(data)
}

// LoadConfigFromBytes parses configuration from YAML bytes.
func LoadConfigFromBytes(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	config.ApplyDefaults()
	return &config, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
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
		"go":       true,
		"golang":   true,
		"all":      true,
	}

	for _, format := range c.Output.Formats {
		if !validFormats[format] {
			return fmt.Errorf("invalid format: %s", format)
		}
	}

	return nil
}

// ApplyDefaults sets default values for optional configuration fields.
func (c *Config) ApplyDefaults() {
	if c.Output.Directory == "" {
		c.Output.Directory = "./generated"
	}

	if c.Output.Filenames == nil {
		c.Output.Filenames = make(map[string]string)
	}

	// Apply generator-specific defaults
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

	if c.Generators.Go != nil && c.Generators.Go.Filename == "" {
		c.Generators.Go.Filename = "types.go"
	}
}

// ResolvePaths converts relative paths to absolute paths based on base directory.
func (c *Config) ResolvePaths(baseDir string) error {
	if !filepath.IsAbs(c.Input.Schema) {
		c.Input.Schema = filepath.Join(baseDir, c.Input.Schema)
	}

	for i, ann := range c.Input.Annotations {
		if !filepath.IsAbs(ann) {
			c.Input.Annotations[i] = filepath.Join(baseDir, ann)
		}
	}

	if c.Output.Directory != "" && !filepath.IsAbs(c.Output.Directory) {
		c.Output.Directory = filepath.Join(baseDir, c.Output.Directory)
	}

	return nil
}

// getGeneratorConfig returns configuration for a specific generator format.
func (c *Config) getGeneratorConfig(format string) map[string]interface{} {
	config := make(map[string]interface{})

	switch format {
	case "graphql":
		if c.Generators.GraphQL != nil {
			config["filename"] = c.Generators.GraphQL.Filename
			config["include_deprecated"] = c.Generators.GraphQL.IncludeDeprecated
		}
	case "protobuf", "proto":
		if c.Generators.Protobuf != nil {
			config["filename"] = c.Generators.Protobuf.Filename
			config["import_buf_validate"] = c.Generators.Protobuf.ImportBufValidate
			config["package_prefix"] = c.Generators.Protobuf.PackagePrefix
		}
	case "openapi":
		if c.Generators.OpenAPI != nil {
			config["filename"] = c.Generators.OpenAPI.Filename
			config["version"] = c.Generators.OpenAPI.Version
		}
	case "go", "golang":
		if c.Generators.Go != nil {
			config["filename"] = c.Generators.Go.Filename
			config["package_name"] = c.Generators.Go.PackageName
			config["json_tags"] = c.Generators.Go.JSONTags
			config["validate_tags"] = c.Generators.Go.ValidateTags
		}
	}

	// Merge custom generator config if present
	if customConfig, ok := c.CustomGenerators[format]; ok {
		for k, v := range customConfig {
			config[k] = v
		}
	}

	return config
}

// ConfigBuilder provides a fluent API for building configurations.
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new configuration builder with defaults.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: NewConfig(),
	}
}

// WithSchema sets the TypeMUX IDL schema content.
func (b *ConfigBuilder) WithSchema(schema string) *ConfigBuilder {
	b.config.Input.Schema = schema
	return b
}

// WithAnnotations adds YAML annotation content.
func (b *ConfigBuilder) WithAnnotations(annotations ...string) *ConfigBuilder {
	b.config.Input.Annotations = append(b.config.Input.Annotations, annotations...)
	return b
}

// WithBaseDir sets the base directory for resolving relative paths.
func (b *ConfigBuilder) WithBaseDir(dir string) *ConfigBuilder {
	b.config.Input.BaseDir = dir
	return b
}

// WithOutputDir sets the output directory for generated files.
func (b *ConfigBuilder) WithOutputDir(dir string) *ConfigBuilder {
	b.config.Output.Directory = dir
	return b
}

// WithFormats sets the output formats to generate.
func (b *ConfigBuilder) WithFormats(formats ...string) *ConfigBuilder {
	b.config.Output.Formats = formats
	return b
}

// WithCleanOutput sets whether to clean the output directory before generation.
func (b *ConfigBuilder) WithCleanOutput(clean bool) *ConfigBuilder {
	b.config.Output.Clean = clean
	return b
}

// WithGraphQLConfig sets GraphQL generator configuration.
func (b *ConfigBuilder) WithGraphQLConfig(cfg *GraphQLConfig) *ConfigBuilder {
	b.config.Generators.GraphQL = cfg
	return b
}

// WithProtobufConfig sets Protobuf generator configuration.
func (b *ConfigBuilder) WithProtobufConfig(cfg *ProtobufConfig) *ConfigBuilder {
	b.config.Generators.Protobuf = cfg
	return b
}

// WithOpenAPIConfig sets OpenAPI generator configuration.
func (b *ConfigBuilder) WithOpenAPIConfig(cfg *OpenAPIConfig) *ConfigBuilder {
	b.config.Generators.OpenAPI = cfg
	return b
}

// WithGoConfig sets Go generator configuration.
func (b *ConfigBuilder) WithGoConfig(cfg *GoConfig) *ConfigBuilder {
	b.config.Generators.Go = cfg
	return b
}

// WithCustomGenerator adds configuration for a custom generator.
func (b *ConfigBuilder) WithCustomGenerator(format string, opts map[string]interface{}) *ConfigBuilder {
	if b.config.CustomGenerators == nil {
		b.config.CustomGenerators = make(map[string]map[string]interface{})
	}
	b.config.CustomGenerators[format] = opts
	return b
}

// Build validates and returns the constructed configuration.
func (b *ConfigBuilder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}
