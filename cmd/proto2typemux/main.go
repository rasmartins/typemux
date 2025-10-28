package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rasmartins/typemux/internal/importers/protobuf"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	inputFile := flag.String("input", "", "Input .proto file (required)")
	outputDir := flag.String("output", "./imported", "Output directory for generated TypeMUX files")
	var importPaths arrayFlags
	flag.Var(&importPaths, "I", "Import search path (can be specified multiple times)")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Add the directory of the input file to import paths
	inputDir := filepath.Dir(*inputFile)
	importPaths = append([]string{inputDir}, importPaths...)

	// Read the proto file
	content, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse the proto file with imports
	parser := protobuf.NewParserWithImports(string(content), *inputFile, importPaths)
	schemas, err := parser.ParseWithImports()
	if err != nil {
		fmt.Printf("Error parsing proto file: %v\n", err)
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0o750); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Convert each schema to TypeMUX and write to separate files
	converter := protobuf.NewConverter()
	filesGenerated := 0

	for protoPath, schema := range schemas {
		// Convert to TypeMUX
		typemuxIDL := converter.Convert(schema)

		// Generate output path preserving directory structure
		// Remove the base import path to get relative path
		relPath := protoPath
		for _, importPath := range importPaths {
			if strings.HasPrefix(protoPath, importPath) {
				relPath = strings.TrimPrefix(protoPath, importPath)
				relPath = strings.TrimPrefix(relPath, "/")
				break
			}
		}

		// Replace .proto with .typemux
		relPath = strings.TrimSuffix(relPath, ".proto") + ".typemux"

		// Create full output path
		outputPath := filepath.Join(*outputDir, relPath)

		// Create subdirectories if needed
		outputSubDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputSubDir, 0o750); err != nil {
			fmt.Printf("Error creating output subdirectory %s: %v\n", outputSubDir, err)
			continue
		}

		// Write the TypeMUX file
		if err := os.WriteFile(outputPath, []byte(typemuxIDL), 0o600); err != nil {
			fmt.Printf("Error writing output file %s: %v\n", outputPath, err)
			continue
		}

		fmt.Printf("Generated %s\n", outputPath)
		filesGenerated++
	}

	fmt.Printf("\nSuccessfully converted %d file(s)\n", filesGenerated)
}
