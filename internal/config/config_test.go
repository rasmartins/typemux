package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.config.yaml")

	configContent := `version: "1.0.0"
input:
  schema: schema.typemux
  annotations:
    - annotations.yaml
output:
  directory: ./generated
  formats:
    - graphql
    - protobuf
  clean: true
generators:
  graphql:
    filename: custom.graphql
  protobuf:
    filename: custom.proto
    import_buf_validate: true
  openapi:
    filename: custom.yaml
    version: "3.1.0"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load the config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify basic fields
	if cfg.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", cfg.Version)
	}

	// Verify input
	expectedSchema := filepath.Join(tmpDir, "schema.typemux")
	if cfg.Input.Schema != expectedSchema {
		t.Errorf("Expected schema %s, got %s", expectedSchema, cfg.Input.Schema)
	}

	if len(cfg.Input.Annotations) != 1 {
		t.Errorf("Expected 1 annotation file, got %d", len(cfg.Input.Annotations))
	}

	// Verify output
	expectedDir := filepath.Join(tmpDir, "generated")
	if cfg.Output.Directory != expectedDir {
		t.Errorf("Expected directory %s, got %s", expectedDir, cfg.Output.Directory)
	}

	if len(cfg.Output.Formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(cfg.Output.Formats))
	}

	if !cfg.Output.Clean {
		t.Error("Expected clean to be true")
	}

	// Verify generator settings
	if cfg.Generators.GraphQL == nil {
		t.Fatal("GraphQL generator config is nil")
	}
	if cfg.Generators.GraphQL.Filename != "custom.graphql" {
		t.Errorf("Expected GraphQL filename custom.graphql, got %s", cfg.Generators.GraphQL.Filename)
	}

	if cfg.Generators.Protobuf == nil {
		t.Fatal("Protobuf generator config is nil")
	}
	if cfg.Generators.Protobuf.Filename != "custom.proto" {
		t.Errorf("Expected Protobuf filename custom.proto, got %s", cfg.Generators.Protobuf.Filename)
	}
	if !cfg.Generators.Protobuf.ImportBufValidate {
		t.Error("Expected ImportBufValidate to be true")
	}

	if cfg.Generators.OpenAPI == nil {
		t.Fatal("OpenAPI generator config is nil")
	}
	if cfg.Generators.OpenAPI.Filename != "custom.yaml" {
		t.Errorf("Expected OpenAPI filename custom.yaml, got %s", cfg.Generators.OpenAPI.Filename)
	}
	if cfg.Generators.OpenAPI.Version != "3.1.0" {
		t.Errorf("Expected OpenAPI version 3.1.0, got %s", cfg.Generators.OpenAPI.Version)
	}
}

func TestValidate_MissingSchema(t *testing.T) {
	cfg := &Config{
		Output: OutputConfig{
			Formats: []string{"graphql"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing schema")
	}
}

func TestValidate_MissingFormats(t *testing.T) {
	cfg := &Config{
		Input: InputConfig{
			Schema: "schema.typemux",
		},
		Output: OutputConfig{
			Formats: []string{},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing formats")
	}
}

func TestValidate_InvalidFormat(t *testing.T) {
	cfg := &Config{
		Input: InputConfig{
			Schema: "schema.typemux",
		},
		Output: OutputConfig{
			Formats: []string{"invalid"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestShouldGenerateFormat(t *testing.T) {
	tests := []struct {
		name     string
		formats  []string
		check    string
		expected bool
	}{
		{
			name:     "all includes graphql",
			formats:  []string{"all"},
			check:    "graphql",
			expected: true,
		},
		{
			name:     "specific format match",
			formats:  []string{"graphql", "protobuf"},
			check:    "graphql",
			expected: true,
		},
		{
			name:     "no match",
			formats:  []string{"graphql"},
			check:    "openapi",
			expected: false,
		},
		{
			name:     "proto alias",
			formats:  []string{"proto"},
			check:    "protobuf",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Output: OutputConfig{
					Formats: tt.formats,
				},
			}

			result := cfg.ShouldGenerateFormat(tt.check)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{
		Input: InputConfig{
			Schema: "schema.typemux",
		},
		Output: OutputConfig{
			Formats: []string{"all"},
		},
		Generators: GeneratorConfig{
			GraphQL:  &GraphQLConfig{},
			Protobuf: &ProtobufConfig{},
			OpenAPI:  &OpenAPIConfig{},
		},
	}

	cfg.ApplyDefaults()

	// Check default output directory
	if cfg.Output.Directory != "./generated" {
		t.Errorf("Expected default directory ./generated, got %s", cfg.Output.Directory)
	}

	// Check GraphQL defaults
	if cfg.Generators.GraphQL.Filename != "schema.graphql" {
		t.Errorf("Expected default GraphQL filename schema.graphql, got %s", cfg.Generators.GraphQL.Filename)
	}

	// Check Protobuf defaults
	if cfg.Generators.Protobuf.Filename != "schema.proto" {
		t.Errorf("Expected default Protobuf filename schema.proto, got %s", cfg.Generators.Protobuf.Filename)
	}

	// Check OpenAPI defaults
	if cfg.Generators.OpenAPI.Filename != "openapi.yaml" {
		t.Errorf("Expected default OpenAPI filename openapi.yaml, got %s", cfg.Generators.OpenAPI.Filename)
	}
	if cfg.Generators.OpenAPI.Version != "3.0.0" {
		t.Errorf("Expected default OpenAPI version 3.0.0, got %s", cfg.Generators.OpenAPI.Version)
	}
}

func TestResolvePaths(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Input: InputConfig{
			Schema: "schema.typemux",
			Annotations: []string{
				"ann1.yaml",
				"ann2.yaml",
			},
		},
		Output: OutputConfig{
			Directory: "./generated",
		},
	}

	err := cfg.ResolvePaths(tmpDir)
	if err != nil {
		t.Fatalf("ResolvePaths failed: %v", err)
	}

	// Check schema path was resolved
	expectedSchema := filepath.Join(tmpDir, "schema.typemux")
	if cfg.Input.Schema != expectedSchema {
		t.Errorf("Expected schema %s, got %s", expectedSchema, cfg.Input.Schema)
	}

	// Check annotation paths were resolved
	for i, ann := range cfg.Input.Annotations {
		expected := filepath.Join(tmpDir, filepath.Base(ann))
		if ann != expected {
			t.Errorf("Expected annotation[%d] %s, got %s", i, expected, ann)
		}
	}

	// Check output directory was resolved
	expectedDir := filepath.Join(tmpDir, "generated")
	if cfg.Output.Directory != expectedDir {
		t.Errorf("Expected directory %s, got %s", expectedDir, cfg.Output.Directory)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `
input:
  schema: test
  invalid yaml here:::
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
