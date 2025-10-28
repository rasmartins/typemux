package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rasmartins/typemux/internal/annotations"
	"github.com/rasmartins/typemux/internal/ast"
	"github.com/rasmartins/typemux/internal/config"
	"github.com/rasmartins/typemux/internal/docgen"
	"github.com/rasmartins/typemux/internal/generator"
	"github.com/rasmartins/typemux/internal/lexer"
	"github.com/rasmartins/typemux/internal/parser"
)

// CurrentTypeMUXVersion is the TypeMUX IDL version supported by this compiler.
const CurrentTypeMUXVersion = "1.0.0"

// arrayFlags is a custom flag type that accumulates multiple values
type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// parseSchemaWithImports recursively parses a schema file and all its imports
func parseSchemaWithImports(filePath string, visited map[string]bool) (*ast.Schema, error) {
	// Get absolute path to handle relative imports correctly
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %s: %v", filePath, err)
	}

	// Check for circular imports
	if visited[absPath] {
		return nil, fmt.Errorf("circular import detected: %s", absPath)
	}
	visited[absPath] = true

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", absPath, err)
	}

	// Parse the file
	l := lexer.New(string(content))
	p := parser.New(l)
	schema := p.Parse()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parser errors in %s:\n%s", absPath, p.PrintErrors())
	}

	// Validate TypeMUX version if specified
	if err := validateTypeMUXVersion(schema.TypeMUXVersion, absPath); err != nil {
		return nil, err
	}

	// Initialize type registry if not already present
	if schema.TypeRegistry == nil {
		schema.TypeRegistry = ast.NewTypeRegistry()
	}

	// Register all types, enums, and unions from this schema
	for _, enum := range schema.Enums {
		schema.TypeRegistry.RegisterEnum(enum)
	}
	for _, typ := range schema.Types {
		schema.TypeRegistry.RegisterType(typ)
	}
	for _, union := range schema.Unions {
		schema.TypeRegistry.RegisterUnion(union)
	}

	// Process imports
	baseDir := filepath.Dir(absPath)
	for _, importPath := range schema.Imports {
		// Resolve import path relative to the current file
		resolvedPath := filepath.Join(baseDir, importPath)

		// Parse the imported file
		importedSchema, err := parseSchemaWithImports(resolvedPath, visited)
		if err != nil {
			return nil, err
		}

		// Merge imported schema into current schema (preserving namespaces)
		schema.Enums = append(schema.Enums, importedSchema.Enums...)
		schema.Types = append(schema.Types, importedSchema.Types...)
		schema.Unions = append(schema.Unions, importedSchema.Unions...)
		schema.Services = append(schema.Services, importedSchema.Services...)

		// Merge type registries
		for qualName, enum := range importedSchema.TypeRegistry.Enums {
			schema.TypeRegistry.Enums[qualName] = enum
		}
		for qualName, typ := range importedSchema.TypeRegistry.Types {
			schema.TypeRegistry.Types[qualName] = typ
		}
		for qualName, union := range importedSchema.TypeRegistry.Unions {
			schema.TypeRegistry.Unions[qualName] = union
		}
	}

	return schema, nil
}

func main() {
	// Config file flag
	configFile := flag.String("config", "", "Configuration file (YAML)")

	// Direct flags (used when no config file is provided)
	inputFile := flag.String("input", "", "Input IDL schema file")
	outputFormat := flag.String("format", "all", "Output format: graphql, protobuf, openapi, go, or all")
	outputDir := flag.String("output", "./generated", "Output directory for generated files")

	var annotationFiles arrayFlags
	flag.Var(&annotationFiles, "annotations", "YAML annotations file (can be specified multiple times)")

	flag.Parse()

	var (
		schemaFile       string
		formats          []string
		outputDirectory  string
		annotationFiles2 []string
	)

	// Load configuration
	if *configFile != "" {
		// Load from config file
		cfg, err := config.Load(*configFile)
		if err != nil {
			fmt.Printf("Error loading config file: %v\n", err)
			os.Exit(1)
		}

		schemaFile = cfg.Input.Schema
		outputDirectory = cfg.Output.Directory
		annotationFiles2 = cfg.Input.Annotations

		// Convert formats
		if cfg.ShouldGenerateFormat("all") {
			formats = []string{"all"}
		} else {
			if cfg.ShouldGenerateFormat("graphql") {
				formats = append(formats, "graphql")
			}
			if cfg.ShouldGenerateFormat("protobuf") {
				formats = append(formats, "protobuf")
			}
			if cfg.ShouldGenerateFormat("openapi") {
				formats = append(formats, "openapi")
			}
			if cfg.ShouldGenerateFormat("go") {
				formats = append(formats, "go")
			}
		}

		// Clean output directory if requested
		if cfg.Output.Clean {
			if err := os.RemoveAll(outputDirectory); err != nil {
				fmt.Printf("Error cleaning output directory: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Loaded configuration from: %s\n", *configFile)
	} else {
		// Use command-line flags
		if *inputFile == "" {
			fmt.Println("Error: -input flag or -config flag is required")
			flag.Usage()
			os.Exit(1)
		}

		schemaFile = *inputFile
		outputDirectory = *outputDir
		annotationFiles2 = annotationFiles
		formats = []string{*outputFormat}
	}

	// Parse the schema with imports
	schema, err := parseSchemaWithImports(schemaFile, make(map[string]bool))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Load and merge YAML annotations if provided
	if len(annotationFiles2) > 0 {
		yamlAnnotations, err := annotations.MergeYAMLAnnotations(annotationFiles2)
		if err != nil {
			fmt.Printf("Error loading YAML annotations: %v\n", err)
			os.Exit(1)
		}

		// Validate annotations
		validator := annotations.NewValidator(schema)
		validationErrors := validator.Validate(yamlAnnotations)
		if len(validationErrors) > 0 {
			fmt.Print(validator.FormatErrors())
			os.Exit(1)
		}

		// Merge annotations into schema
		merger := annotations.NewMerger(yamlAnnotations)
		merger.Merge(schema)

		fmt.Printf("Loaded annotations from %d file(s)\n", len(annotationFiles2))
	}

	// Create output directory
	if err := os.MkdirAll(outputDirectory, 0o750); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate output based on formats
	for _, format := range formats {
		switch format {
		case "graphql":
			generateGraphQL(schema, outputDirectory)
		case "protobuf", "proto":
			generateProtobuf(schema, outputDirectory)
		case "openapi":
			generateOpenAPI(schema, outputDirectory)
		case "go", "golang":
			generateGo(schema, outputDirectory)
		case "docs", "markdown", "md":
			generateMarkdownDocs(schema, outputDirectory)
		case "all":
			generateGraphQL(schema, outputDirectory)
			generateProtobuf(schema, outputDirectory)
			generateOpenAPI(schema, outputDirectory)
			generateGo(schema, outputDirectory)
			generateMarkdownDocs(schema, outputDirectory)
		default:
			fmt.Printf("Unknown format: %s\n", format)
			os.Exit(1)
		}
	}

	fmt.Println("Code generation completed successfully!")
}

func generateGraphQL(schema *ast.Schema, outputDir string) {
	gen := generator.NewGraphQLGenerator()
	output := gen.Generate(schema)

	outputPath := filepath.Join(outputDir, "schema.graphql")
	if err := os.WriteFile(outputPath, []byte(output), 0o600); err != nil {
		fmt.Printf("Error writing GraphQL schema: %v\n", err)
		return
	}
	fmt.Printf("Generated GraphQL schema: %s\n", outputPath)
}

func generateProtobuf(schema *ast.Schema, outputDir string) {
	gen := generator.NewProtobufGenerator()

	// Check if we have multiple namespaces
	namespaces := collectNamespaces(schema)

	if len(namespaces) > 1 {
		// Generate separate proto files per namespace
		protoFiles := gen.GenerateByNamespace(schema)

		for ns, content := range protoFiles {
			// Create namespace directory structure (e.g., com/example/users/)
			nsPath := strings.ReplaceAll(ns, ".", "/")
			nsDir := filepath.Join(outputDir, filepath.Dir(nsPath))
			if err := os.MkdirAll(nsDir, 0o750); err != nil {
				fmt.Printf("Error creating namespace directory: %v\n", err)
				continue
			}

			// Write proto file (e.g., com/example/users.proto)
			outputPath := filepath.Join(outputDir, nsPath+".proto")
			if err := os.WriteFile(outputPath, []byte(content), 0o600); err != nil {
				fmt.Printf("Error writing Protobuf schema for %s: %v\n", ns, err)
				continue
			}
			fmt.Printf("Generated Protobuf schema: %s\n", outputPath)
		}
	} else {
		// Single namespace - generate single file
		output := gen.Generate(schema)
		outputPath := filepath.Join(outputDir, "schema.proto")
		if err := os.WriteFile(outputPath, []byte(output), 0o600); err != nil {
			fmt.Printf("Error writing Protobuf schema: %v\n", err)
			return
		}
		fmt.Printf("Generated Protobuf schema: %s\n", outputPath)
	}
}

// collectNamespaces returns all unique namespaces in the schema
func collectNamespaces(schema *ast.Schema) []string {
	nsMap := make(map[string]bool)

	for _, enum := range schema.Enums {
		ns := enum.Namespace
		if ns == "" {
			ns = "api"
		}
		nsMap[ns] = true
	}

	for _, typ := range schema.Types {
		ns := typ.Namespace
		if ns == "" {
			ns = "api"
		}
		nsMap[ns] = true
	}

	for _, union := range schema.Unions {
		ns := union.Namespace
		if ns == "" {
			ns = "api"
		}
		nsMap[ns] = true
	}

	for _, service := range schema.Services {
		ns := service.Namespace
		if ns == "" {
			ns = "api"
		}
		nsMap[ns] = true
	}

	result := make([]string, 0, len(nsMap))
	for ns := range nsMap {
		result = append(result, ns)
	}
	return result
}

func generateOpenAPI(schema *ast.Schema, outputDir string) {
	gen := generator.NewOpenAPIGenerator()
	output := gen.Generate(schema)

	outputPath := filepath.Join(outputDir, "openapi.yaml")
	if err := os.WriteFile(outputPath, []byte(output), 0o600); err != nil {
		fmt.Printf("Error writing OpenAPI schema: %v\n", err)
		return
	}
	fmt.Printf("Generated OpenAPI schema: %s\n", outputPath)
}

func generateGo(schema *ast.Schema, outputDir string) {
	gen := generator.NewGoGenerator()
	output := gen.Generate(schema)

	outputPath := filepath.Join(outputDir, "types.go")
	if err := os.WriteFile(outputPath, []byte(output), 0o600); err != nil {
		fmt.Printf("Error writing Go code: %v\n", err)
		return
	}
	fmt.Printf("Generated Go code: %s\n", outputPath)
}

func generateMarkdownDocs(schema *ast.Schema, outputDir string) {
	gen := docgen.NewMarkdownGenerator()
	output := gen.Generate(schema)

	outputPath := filepath.Join(outputDir, "API.md")
	if err := os.WriteFile(outputPath, []byte(output), 0o600); err != nil {
		fmt.Printf("Error writing Markdown documentation: %v\n", err)
		return
	}
	fmt.Printf("Generated Markdown documentation: %s\n", outputPath)
}

// validateTypeMUXVersion validates that the schema's TypeMUX version is compatible
func validateTypeMUXVersion(schemaVersion, filePath string) error {
	// If no version is specified, accept it (backward compatibility)
	if schemaVersion == "" {
		fmt.Printf("Warning: No @typemux version specified in %s\n", filePath)
		return nil
	}

	// Parse versions (simple major.minor.patch comparison)
	if schemaVersion != CurrentTypeMUXVersion {
		// For now, only accept exact version match
		// In the future, we could implement more sophisticated version compatibility
		return fmt.Errorf("incompatible TypeMUX version in %s: schema requires %s, but compiler supports %s",
			filePath, schemaVersion, CurrentTypeMUXVersion)
	}

	return nil
}
