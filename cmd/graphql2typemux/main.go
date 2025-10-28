package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rasmartins/typemux/internal/importers/graphql"
)

func main() {
	inputFile := flag.String("input", "", "Input .graphql or .graphqls file (required)")
	outputDir := flag.String("output", "./imported", "Output directory for generated TypeMUX files")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read the GraphQL file
	content, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse the GraphQL schema
	parser := graphql.NewParser(string(content))
	schema, err := parser.Parse()
	if err != nil {
		fmt.Printf("Error parsing GraphQL file: %v\n", err)
		os.Exit(1)
	}

	// Convert to TypeMUX IDL
	converter := graphql.NewConverter()
	typemuxIDL := converter.Convert(schema)

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

	fmt.Printf("Successfully converted GraphQL schema to TypeMUX IDL\n")
	fmt.Printf("Input:  %s\n", *inputFile)
	fmt.Printf("Output: %s\n", outputFile)
}
