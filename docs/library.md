# Using TypeMUX as a Go Library

TypeMUX can be used as a Go library in your applications. This guide shows how to use the public API to parse schemas, generate code, import external formats, and detect breaking changes.

## Installation

```bash
go get github.com/rasmartins/typemux
```

## Quick Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/rasmartins/typemux"
)

func main() {
    // Parse a TypeMUX schema
    idl := `
    namespace myapi

    type User {
      id: string @required
      email: string @required
      name: string
    }
    `

    schema, err := typemux.ParseSchema(idl)
    if err != nil {
        log.Fatal(err)
    }

    // Generate GraphQL
    factory := typemux.NewGeneratorFactory()
    graphql, err := factory.Generate("graphql", schema)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(graphql)
}
```

## API Reference

### Parsing Schemas

#### ParseSchema

Parse a TypeMUX IDL schema from a string:

```go
schema, err := typemux.ParseSchema(idlContent)
if err != nil {
    log.Fatal(err)
}
```

#### ParseWithAnnotations

Parse a schema and merge YAML annotations:

```go
idl := `
namespace myapi

type User {
  id: string
  email: string
}
`

yamlAnnotations := `
types:
  User:
    fields:
      id:
        annotations:
          - name: "@required"
      email:
        annotations:
          - name: "@required"
`

schema, err := typemux.ParseWithAnnotations(idl, yamlAnnotations)
if err != nil {
    log.Fatal(err)
}
```

You can provide multiple YAML annotation strings; they are merged in order:

```go
schema, err := typemux.ParseWithAnnotations(idl, annotations1, annotations2)
```

### Generating Output

#### Using the Generator Factory

Create a generator factory and generate output:

```go
factory := typemux.NewGeneratorFactory()

// Generate GraphQL
graphql, err := factory.Generate("graphql", schema)

// Generate Protobuf
proto, err := factory.Generate("protobuf", schema)
// Or use alias:
proto, err := factory.Generate("proto", schema)

// Generate OpenAPI
openapi, err := factory.Generate("openapi", schema)

// Generate Go code
goCode, err := factory.Generate("go", schema)
// Or use alias:
goCode, err := factory.Generate("golang", schema)
```

#### Generate All Formats

Generate all registered formats at once:

```go
factory := typemux.NewGeneratorFactory()
outputs, err := factory.GenerateAll(schema)
if err != nil {
    log.Fatal(err)
}

for format, content := range outputs {
    fmt.Printf("=== %s ===\n%s\n\n", format, content)
}
```

#### Custom Generators

Register your own custom generator:

```go
type CustomGenerator struct{}

func (g *CustomGenerator) Generate(schema *typemux.Schema) (string, error) {
    // Your custom generation logic
    return "custom output", nil
}

func (g *CustomGenerator) Format() string {
    return "custom"
}

func (g *CustomGenerator) FileExtension() string {
    return ".custom"
}

// Register the custom generator
factory := typemux.NewGeneratorFactory()
factory.Register(&CustomGenerator{})

// Use it
output, err := factory.Generate("custom", schema)
```

#### Check Available Formats

```go
factory := typemux.NewGeneratorFactory()

// Check if a format exists
if factory.HasFormat("graphql") {
    fmt.Println("GraphQL generator is available")
}

// Get all available formats
formats := factory.GetFormats()
fmt.Println("Available formats:", formats)
// Output: Available formats: [go graphql openapi protobuf]
```

### Configuration API

Use the builder pattern for fluent configuration:

```go
config, err := typemux.NewConfigBuilder().
    WithSchema(idlContent).
    WithFormats("graphql", "protobuf").
    WithOutputDir("./generated").
    WithNamespace("myapi").
    Build()

if err != nil {
    log.Fatal(err)
}

// Generate using config
factory := typemux.NewGeneratorFactory()
outputs, err := factory.GenerateWithConfig(config)
if err != nil {
    log.Fatal(err)
}

for format, content := range outputs {
    fmt.Printf("Generated %s\n", format)
}
```

#### Configuration Options

```go
config, err := typemux.NewConfigBuilder().
    // Input
    WithSchema(idlContent).
    WithSchemaFile("schema.typemux").  // Or read from file
    WithAnnotations(yamlAnnotation1, yamlAnnotation2).

    // Output
    WithFormats("graphql", "protobuf", "openapi", "go").
    WithOutputDir("./generated").

    // Generator options
    WithProtoPackage("myapi.v1").
    WithGraphQLNullable(true).

    Build()
```

### Importing External Formats

Convert external schema formats to TypeMUX IDL:

#### Import GraphQL

```go
factory := typemux.NewImporterFactory()

graphqlSchema := `
type User {
  id: ID!
  email: String!
  name: String
}

type Query {
  user(id: ID!): User
}
`

typemuxIDL, err := factory.ImportGraphQL(graphqlSchema)
if err != nil {
    log.Fatal(err)
}

fmt.Println(typemuxIDL)
```

#### Import Protobuf

```go
factory := typemux.NewImporterFactory()

protoContent := `
syntax = "proto3";

message User {
  string id = 1;
  string email = 2;
  string name = 3;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
}
`

typemuxIDL, err := factory.ImportProtobuf(protoContent)
if err != nil {
    log.Fatal(err)
}

fmt.Println(typemuxIDL)
```

#### Import OpenAPI

```go
factory := typemux.NewImporterFactory()

openapiYAML := `
openapi: 3.0.0
info:
  title: User API
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      required:
        - id
        - email
      properties:
        id:
          type: string
        email:
          type: string
        name:
          type: string
`

typemuxIDL, err := factory.ImportOpenAPI(openapiYAML)
if err != nil {
    log.Fatal(err)
}

fmt.Println(typemuxIDL)
```

#### Custom Importers

Register your own custom importer:

```go
type JSONSchemaImporter struct{}

func (i *JSONSchemaImporter) Import(content string) (string, error) {
    // Your import logic
    return "namespace myapi\n\ntype MyType { }", nil
}

func (i *JSONSchemaImporter) Format() string {
    return "jsonschema"
}

factory := typemux.NewImporterFactory()
factory.Register(&JSONSchemaImporter{})

typemuxIDL, err := factory.Import("jsonschema", jsonSchemaContent)
```

### Breaking Change Detection

Detect breaking changes between schema versions:

#### Basic Diff

```go
baseIDL := `
namespace myapi

type User {
  id: string @required
  email: string @required
}
`

headIDL := `
namespace myapi

type User {
  id: string @required
  email: string @required
  name: string
}
`

baseSchema, _ := typemux.ParseSchema(baseIDL)
headSchema, _ := typemux.ParseSchema(headIDL)

result, err := typemux.Diff(baseSchema, headSchema)
if err != nil {
    log.Fatal(err)
}

if result.HasBreakingChanges() {
    fmt.Println("⚠️ Breaking changes detected!")
}

if result.HasChanges() {
    fmt.Printf("Changes: %d breaking, %d dangerous, %d non-breaking\n",
        result.BreakingCount,
        result.DangerousCount,
        result.NonBreakingCount)
}
```

#### Generate Change Report

```go
result, err := typemux.Diff(baseSchema, headSchema)
if err != nil {
    log.Fatal(err)
}

// Detailed report
report := result.Report()
fmt.Println(report)

// Compact one-line summary
compact := result.CompactReport()
fmt.Println(compact)
// Output: 2 breaking, 1 dangerous, 3 non-breaking changes detected
```

#### Inspect Individual Changes

```go
result, err := typemux.Diff(baseSchema, headSchema)
if err != nil {
    log.Fatal(err)
}

for _, change := range result.Changes {
    fmt.Printf("Type: %s\n", change.Type)
    fmt.Printf("Severity: %s\n", change.Severity)
    fmt.Printf("Protocol: %s\n", change.Protocol)
    fmt.Printf("Path: %s\n", change.Path)
    fmt.Printf("Description: %s\n", change.Description)
    fmt.Println()
}
```

#### Diff with Options

Filter changes by protocol or ignore specific change types:

```go
result, err := typemux.DiffWithOptions(baseSchema, headSchema, typemux.DiffOptions{
    // Only show changes affecting GraphQL
    Protocol: typemux.ProtocolGraphQL,

    // Ignore field additions (non-breaking)
    IgnoreChanges: []typemux.ChangeType{
        typemux.ChangeTypeFieldAdded,
    },
})
```

#### Change Types

Available change type constants:

```go
typemux.ChangeTypeFieldAdded          // Field added to type
typemux.ChangeTypeFieldRemoved        // Field removed (breaking)
typemux.ChangeTypeFieldTypeChanged    // Field type changed (breaking)
typemux.ChangeTypeFieldMadeRequired   // Field made required (breaking)
typemux.ChangeTypeFieldMadeOptional   // Field made optional
typemux.ChangeTypeTypeAdded           // New type added
typemux.ChangeTypeTypeRemoved         // Type removed (breaking)
typemux.ChangeTypeEnumValueAdded      // Enum value added
typemux.ChangeTypeEnumValueRemoved    // Enum value removed (breaking)
typemux.ChangeTypeMethodAdded         // Service method added
typemux.ChangeTypeMethodRemoved       // Service method removed (breaking)
typemux.ChangeTypeMethodParamChanged  // Method parameters changed (breaking)
typemux.ChangeTypeMethodReturnChanged // Method return type changed (breaking)
```

#### Severity Levels

```go
typemux.SeverityBreaking    // Breaking change - requires major version bump
typemux.SeverityDangerous   // Potentially dangerous - review carefully
typemux.SeverityNonBreaking // Safe change - minor/patch version bump
```

#### Protocol Constants

```go
typemux.ProtocolGraphQL  // Change affects GraphQL
typemux.ProtocolProto    // Change affects Protobuf
typemux.ProtocolOpenAPI  // Change affects OpenAPI
typemux.ProtocolGo       // Change affects Go code
```

### Annotation Metadata

Query built-in annotation metadata:

#### Get All Annotations

```go
annotations := typemux.GetBuiltinAnnotations()

for _, ann := range annotations {
    fmt.Printf("%s: %s\n", ann.Name, ann.Description)
    fmt.Printf("  Scope: %v\n", ann.Scope)
    fmt.Printf("  Formats: %v\n", ann.Formats)
}
```

#### Filter by Scope

```go
// Get all field-level annotations
fieldAnnotations := typemux.FilterAnnotationsByScope("field")

for _, ann := range fieldAnnotations {
    fmt.Println(ann.Name)
}
```

Scopes: `"schema"`, `"type"`, `"field"`, `"method"`, `"enum"`, `"union"`

#### Filter by Format

```go
// Get all GraphQL-specific annotations
graphqlAnnotations := typemux.FilterAnnotationsByFormat("graphql")

for _, ann := range graphqlAnnotations {
    fmt.Printf("%s - %s\n", ann.Name, ann.Description)
}
```

Formats: `"graphql"`, `"proto"`, `"openapi"`, `"go"`, `"all"`

#### Get Specific Annotation

```go
ann, found := typemux.GetAnnotation("@required")
if found {
    fmt.Printf("Name: %s\n", ann.Name)
    fmt.Printf("Description: %s\n", ann.Description)
    fmt.Printf("Scope: %v\n", ann.Scope)
    fmt.Printf("Formats: %v\n", ann.Formats)

    if len(ann.Parameters) > 0 {
        fmt.Println("Parameters:")
        for _, param := range ann.Parameters {
            fmt.Printf("  - %s (%s): %s\n",
                param.Name, param.Type, param.Description)
        }
    }

    if len(ann.Examples) > 0 {
        fmt.Println("Examples:")
        for _, example := range ann.Examples {
            fmt.Printf("  %s\n", example)
        }
    }
}
```

## Complete Example

Here's a complete example showing multiple API features:

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/rasmartins/typemux"
)

func main() {
    // 1. Parse a schema
    idl := `
    namespace myapi

    type User {
      id: string @required
      email: string @required
      name: string
      age: int32
    }

    service UserService {
      rpc GetUser(GetUserRequest) returns (User)
        @http.method(GET)
        @http.path("/users/{id}")
    }

    type GetUserRequest {
      id: string @required
    }
    `

    schema, err := typemux.ParseSchema(idl)
    if err != nil {
        log.Fatal("Parse error:", err)
    }

    // 2. Generate multiple formats
    factory := typemux.NewGeneratorFactory()

    outputs, err := factory.GenerateAll(schema)
    if err != nil {
        log.Fatal("Generation error:", err)
    }

    // 3. Write outputs to files
    os.Mkdir("generated", 0755)

    for format, content := range outputs {
        var filename string
        switch format {
        case "graphql":
            filename = "generated/schema.graphql"
        case "protobuf":
            filename = "generated/schema.proto"
        case "openapi":
            filename = "generated/openapi.yaml"
        case "go":
            filename = "generated/types.go"
        }

        err = os.WriteFile(filename, []byte(content), 0644)
        if err != nil {
            log.Printf("Failed to write %s: %v", filename, err)
            continue
        }
        fmt.Printf("✅ Generated %s\n", filename)
    }

    // 4. Check for breaking changes against a modified schema
    modifiedIDL := `
    namespace myapi

    type User {
      id: string @required
      email: string @required
      // 'name' field removed - this is breaking!
      age: int32
      createdAt: timestamp  // New field added
    }

    service UserService {
      rpc GetUser(GetUserRequest) returns (User)
        @http.method(GET)
        @http.path("/users/{id}")
    }

    type GetUserRequest {
      id: string @required
    }
    `

    modifiedSchema, err := typemux.ParseSchema(modifiedIDL)
    if err != nil {
        log.Fatal("Parse error:", err)
    }

    diffResult, err := typemux.Diff(schema, modifiedSchema)
    if err != nil {
        log.Fatal("Diff error:", err)
    }

    fmt.Println("\n=== Schema Changes ===")
    if diffResult.HasBreakingChanges() {
        fmt.Println("⚠️  Breaking changes detected!")
    }

    fmt.Println(diffResult.Report())

    // 5. List available annotations
    fmt.Println("\n=== Available Field Annotations ===")
    fieldAnnotations := typemux.FilterAnnotationsByScope("field")
    for _, ann := range fieldAnnotations {
        fmt.Printf("  %s - %s\n", ann.Name, ann.Description)
    }
}
```

## Integration Patterns

### Pipeline Pattern

Build a schema transformation pipeline:

```go
func processSchema(inputFile string) error {
    // Read input
    content, err := os.ReadFile(inputFile)
    if err != nil {
        return err
    }

    // Parse
    schema, err := typemux.ParseSchema(string(content))
    if err != nil {
        return fmt.Errorf("parse error: %w", err)
    }

    // Generate
    factory := typemux.NewGeneratorFactory()
    outputs, err := factory.GenerateAll(schema)
    if err != nil {
        return fmt.Errorf("generation error: %w", err)
    }

    // Write outputs
    for format, content := range outputs {
        filename := fmt.Sprintf("output/%s.%s", format, getExtension(format))
        if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
            return err
        }
    }

    return nil
}
```

### CI/CD Integration

Use in CI to detect breaking changes:

```go
func checkBreakingChanges(baseRef, headRef string) error {
    // Get base schema
    baseContent, _ := exec.Command("git", "show", baseRef+":schema.typemux").Output()
    baseSchema, err := typemux.ParseSchema(string(baseContent))
    if err != nil {
        return err
    }

    // Get head schema
    headContent, _ := os.ReadFile("schema.typemux")
    headSchema, err := typemux.ParseSchema(string(headContent))
    if err != nil {
        return err
    }

    // Diff
    result, err := typemux.Diff(baseSchema, headSchema)
    if err != nil {
        return err
    }

    if result.HasBreakingChanges() {
        fmt.Println("❌ Breaking changes detected:")
        fmt.Println(result.Report())
        os.Exit(1)
    }

    fmt.Println("✅ No breaking changes")
    return nil
}
```

### Format Conversion Service

Create a web service that converts between formats:

```go
func convertHandler(w http.ResponseWriter, r *http.Request) {
    sourceFormat := r.URL.Query().Get("from")
    targetFormat := r.URL.Query().Get("to")

    body, _ := io.ReadAll(r.Body)

    // Import from source format
    importFactory := typemux.NewImporterFactory()
    typemuxIDL, err := importFactory.Import(sourceFormat, string(body))
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Parse
    schema, err := typemux.ParseSchema(typemuxIDL)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Generate target format
    genFactory := typemux.NewGeneratorFactory()
    output, err := genFactory.Generate(targetFormat, schema)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}
```

## Error Handling

All API functions return errors that should be checked:

```go
schema, err := typemux.ParseSchema(idl)
if err != nil {
    // Parse errors include line numbers and details
    log.Fatalf("Failed to parse schema: %v", err)
}

output, err := factory.Generate("graphql", schema)
if err != nil {
    // Generation errors indicate format issues
    log.Fatalf("Failed to generate output: %v", err)
}
```

## Best Practices

1. **Reuse factories**: Create `GeneratorFactory` and `ImporterFactory` once and reuse them
2. **Handle errors**: Always check errors from parsing and generation
3. **Validate early**: Parse schemas as early as possible to catch errors
4. **Use configuration**: For complex setups, use `ConfigBuilder` instead of multiple function calls
5. **Check breaking changes**: Run diff checks in CI/CD before merging schema changes
6. **Cache parsed schemas**: If processing the same schema multiple times, parse once and cache the result

## See Also

- [CLI Documentation](quickstart.md) - Using TypeMUX from the command line
- [Configuration Guide](configuration.md) - Advanced configuration options
- [Reference](reference.md) - TypeMUX language reference
- [Examples](examples.md) - More example schemas
