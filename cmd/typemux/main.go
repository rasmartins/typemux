package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rasmartins/typemux/internal/annotations"
	"github.com/rasmartins/typemux/internal/ast"
	"github.com/rasmartins/typemux/internal/generator"
	"github.com/rasmartins/typemux/internal/lexer"
	"github.com/rasmartins/typemux/internal/parser"
)

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
	inputFile := flag.String("input", "", "Input IDL schema file (required)")
	outputFormat := flag.String("format", "all", "Output format: graphql, protobuf, openapi, or all")
	outputDir := flag.String("output", "./generated", "Output directory for generated files")

	var annotationFiles arrayFlags
	flag.Var(&annotationFiles, "annotations", "YAML annotations file (can be specified multiple times)")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse the schema with imports
	schema, err := parseSchemaWithImports(*inputFile, make(map[string]bool))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Load and merge YAML annotations if provided
	if len(annotationFiles) > 0 {
		yamlAnnotations, err := annotations.MergeYAMLAnnotations(annotationFiles)
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

		fmt.Printf("Loaded annotations from %d file(s)\n", len(annotationFiles))
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate output based on format
	switch *outputFormat {
	case "graphql":
		generateGraphQL(schema, *outputDir)
	case "protobuf", "proto":
		generateProtobuf(schema, *outputDir)
	case "openapi":
		generateOpenAPI(schema, *outputDir)
	case "all":
		generateGraphQL(schema, *outputDir)
		generateProtobuf(schema, *outputDir)
		generateOpenAPI(schema, *outputDir)
	default:
		fmt.Printf("Unknown format: %s\n", *outputFormat)
		os.Exit(1)
	}

	fmt.Println("Code generation completed successfully!")
}

func generateGraphQL(schema *ast.Schema, outputDir string) {
	gen := generator.NewGraphQLGenerator()
	output := gen.Generate(schema)

	outputPath := filepath.Join(outputDir, "schema.graphql")
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
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
			if err := os.MkdirAll(nsDir, 0755); err != nil {
				fmt.Printf("Error creating namespace directory: %v\n", err)
				continue
			}

			// Write proto file (e.g., com/example/users.proto)
			outputPath := filepath.Join(outputDir, nsPath+".proto")
			if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
				fmt.Printf("Error writing Protobuf schema for %s: %v\n", ns, err)
				continue
			}
			fmt.Printf("Generated Protobuf schema: %s\n", outputPath)
		}
	} else {
		// Single namespace - generate single file
		output := gen.Generate(schema)
		outputPath := filepath.Join(outputDir, "schema.proto")
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
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
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		fmt.Printf("Error writing OpenAPI schema: %v\n", err)
		return
	}
	fmt.Printf("Generated OpenAPI schema: %s\n", outputPath)
}
