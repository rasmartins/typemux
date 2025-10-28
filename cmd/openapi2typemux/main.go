package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rasmartins/typemux/internal/importers/openapi"
)

func main() {
	inputFile := flag.String("input", "", "Input OpenAPI file (.yaml, .yml, or .json) (required)")
	outputDir := flag.String("output", "./imported", "Output directory for generated TypeMUX files")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read the OpenAPI file
	content, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse the OpenAPI spec
	parser := openapi.NewParser(content)
	spec, err := parser.Parse()
	if err != nil {
		fmt.Printf("Error parsing OpenAPI file: %v\n", err)
		os.Exit(1)
	}

	// Convert to TypeMUX IDL
	converter := openapi.NewConverter()
	typemuxIDL := converter.Convert(spec)

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0o750); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate output filename from input filename
	baseName := filepath.Base(*inputFile)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputFile := filepath.Join(*outputDir, baseName+".typemux")

	// Write the output file
	if err := os.WriteFile(outputFile, []byte(typemuxIDL), 0o600); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted OpenAPI spec to TypeMUX IDL\n")
	fmt.Printf("Input:  %s\n", *inputFile)
	fmt.Printf("Output: %s\n", outputFile)
}
