# Configuration Guide

Guide to configuring and using TypeMUX.

## Table of Contents

- [CLI Flags](#cli-flags)
- [Config File](#config-file)
- [YAML Annotations](#yaml-annotations)
- [Annotation Merging](#annotation-merging)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## CLI Flags

TypeMUX accepts the following command-line flags:

### -input

**Required.** Path to the input TypeMUX schema file.

```bash
typemux -input schema.typemux
```

**Multiple files:** Use imports in your schema file instead of multiple `-input` flags.

### -format

Output format to generate. Default: `all`

**Options:**
- `all` - Generate all formats (default)
- `graphql` - Generate only GraphQL schema
- `protobuf` (or `proto`) - Generate only Protocol Buffers
- `openapi` - Generate only OpenAPI specification
- `go` (or `golang`) - Generate only Go code
- `markdown` (or `docs`) - Generate only documentation

**Examples:**

```bash
# Generate all formats
typemux -input schema.typemux -format all

# Generate only GraphQL
typemux -input schema.typemux -format graphql

# Generate only Go code
typemux -input schema.typemux -format go

# Generate only documentation
typemux -input schema.typemux -format markdown
```

### -output

Output directory for generated files. Default: `./generated`

```bash
typemux -input schema.typemux -output ./gen
typemux -input schema.typemux -output /tmp/api-schemas
```

**Created files:**
- GraphQL: `<output>/schema.graphql`
- Protobuf: `<output>/schema.proto` (or namespace-specific files)
- OpenAPI: `<output>/openapi.yaml`
- Go: `<output>/types.go`
- Markdown: `<output>/schema.md`

### -annotations

Path to YAML annotations file. Can be specified multiple times.

```bash
# Single annotations file
typemux -input schema.typemux -annotations annotations.yaml

# Multiple files (merged, later files override earlier)
typemux -input schema.typemux \
        -annotations base-annotations.yaml \
        -annotations environment-specific.yaml
```

### -config

Path to configuration file. See [Config File](#config-file) section.

```bash
typemux -config typemux.config.yaml
```

### Complete Example

```bash
typemux \
  -input api/schema.typemux \
  -format all \
  -output ./generated/api \
  -annotations api/annotations.yaml
```

## Config File

Instead of CLI flags, you can use a YAML configuration file:

### File Format

```yaml
# typemux.config.yaml
input: schema.typemux
output:
  directory: ./generated
  formats:
    - graphql
    - protobuf
    - openapi
    - go
annotations:
  - base-annotations.yaml
  - overrides.yaml
```

### Configuration Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `input` | string | Path to schema file | Required |
| `output.directory` | string | Output directory | `./generated` |
| `output.formats` | array | Formats to generate | `["all"]` |
| `annotations` | array | YAML annotation files | `[]` |

### Usage

```bash
# Use config file
typemux -config typemux.config.yaml

# Override with CLI flags
typemux -config typemux.config.yaml -format graphql -output ./custom
```

**Priority:** CLI flags override config file settings.

## YAML Annotations

YAML annotation files provide format-specific metadata without cluttering schema files.

> **Note:** For annotation syntax and available annotations, see the [Language Reference](reference.md#field-attributes).

### File Structure

```yaml
types:
  TypeName:
    proto:
      - option_name = "value"
    graphql:
      - '@directive'
    openapi:
      - x-extension: value
    go:
      - package = "pkgname"

    fields:
      fieldName:
        proto:
          - option = "value"
        graphql:
          - '@directive(arg: "value")'
        openapi:
          - x-field-extension: value

enums:
  EnumName:
    proto:
      - allow_alias = true

    values:
      VALUE_NAME:
        proto:
          - option = "value"

services:
  ServiceName:
    proto:
      - option = "value"

    methods:
      MethodName:
        proto:
          - option = "value"
        openapi:
          - x-method-extension: value
```

### Complete Example

**annotations.yaml:**
```yaml
types:
  User:
    proto:
      - deprecated = true
    graphql:
      - '@key(fields: "id")'
    openapi:
      - x-internal: true

    fields:
      email:
        proto:
          - json_name = "emailAddress"
        graphql:
          - '@deprecated(reason: "Use emailAddress")'

      password_hash:
        proto:
          - json_name = "password"

enums:
  UserRole:
    proto:
      - allow_alias = true

    values:
      ADMIN:
        proto:
          - option = "(custom.role_level) = 100"

services:
  UserService:
    proto:
      - deprecated = false

    methods:
      GetUser:
        openapi:
          - x-rate-limit: 1000
```

### Qualified Names (with Namespaces)

When using namespaces, reference types by their qualified names:

```yaml
types:
  # Fully qualified name
  com.example.users.User:
    proto:
      - option = "value"

  # Another namespace
  com.example.orders.Order:
    fields:
      userId:
        proto:
          - json_name = "user_id"
```

### Format-Specific Options

#### Protobuf Options

Standard Protobuf options:
- `deprecated = true`
- `json_name = "customName"`
- `packed = false` (for repeated fields)

Custom options:
```yaml
proto:
  - (my.custom.option) = "value"
  - (validate.rules).string.email = true
```

#### GraphQL Directives

```yaml
graphql:
  - '@key(fields: "id")'
  - '@external'
  - '@requires(fields: "userId")'
  - '@deprecated(reason: "Use newField")'
```

#### OpenAPI Extensions

Any property starting with `x-`:
```yaml
openapi:
  - x-internal: true
  - x-rate-limit: 1000
  - x-nullable: true
  - x-format: email
```

#### Go Options

```yaml
go:
  - package = "mypackage"
  - json = "omitempty"
```

### Validation and Deprecation

See [Language Reference - Validation](reference.md#validation) for details on validation annotations.

## Annotation Merging

When annotations are specified in multiple places, TypeMUX merges them with a specific priority.

### Merge Order (highest to lowest priority)

1. **YAML annotations** (last file wins if multiple files)
2. **Inline annotations** in schema file

**Example:**

**schema.typemux:**
```typemux
type User {
  email: string @required
}
```

**annotations.yaml:**
```yaml
types:
  User:
    fields:
      email:
        proto:
          - json_name = "emailAddress"
```

**Result:** Field `email` has both `@required` (from inline) and `json_name` (from YAML).

### Multiple YAML Files

When using multiple YAML annotation files:

```bash
typemux -input schema.typemux \
        -annotations base.yaml \
        -annotations override.yaml
```

**Merge behavior:**
- Files are processed left to right
- Later files override earlier files
- Within each file, annotations are additive (not replaced)

**Example:**

**base.yaml:**
```yaml
types:
  User:
    proto:
      - deprecated = false
    fields:
      email:
        proto:
          - json_name = "email"
```

**override.yaml:**
```yaml
types:
  User:
    proto:
      - deprecated = true
    fields:
      email:
        graphql:
          - '@deprecated(reason: "Old field")'
```

**Result:**
- `User` has: `proto: [deprecated = true]` (overridden)
- `email` has: both proto `json_name` and graphql `@deprecated` (additive)

### Use Cases

**Environment-specific configuration:**
```bash
typemux -input schema.typemux \
        -annotations base-annotations.yaml \
        -annotations prod-overrides.yaml
```

**Team-specific overrides:**
```bash
typemux -input schema.typemux \
        -annotations shared-annotations.yaml \
        -annotations team-customizations.yaml
```

## Best Practices

### When to Use Inline vs YAML Annotations

**Use inline annotations when:**
- Annotation is core to the type's semantics (`@required`, `@default`)
- Schema is small and simple
- Team prefers single-file schemas

**Use YAML annotations when:**
- Format-specific customization (Protobuf options, GraphQL directives)
- Environment-specific configuration
- Separating concerns (schema vs. metadata)
- Multiple teams with different requirements
- Large schemas that would become cluttered

**Example split:**

**schema.typemux** (business logic):
```typemux
type User {
  id: string @required
  email: string @required
  age: int32
}
```

**annotations.yaml** (format-specific):
```yaml
types:
  User:
    proto:
      - deprecated = false
    graphql:
      - '@key(fields: "id")'
    fields:
      email:
        proto:
          - json_name = "emailAddress"
```

### Annotation Organization

**Small projects:**
```
schema.typemux
annotations.yaml
```

**Medium projects:**
```
schema.typemux
annotations/
  protobuf.yaml
  graphql.yaml
  openapi.yaml
```

**Large projects:**
```
schemas/
  users.typemux
  orders.typemux
annotations/
  base/
    users.yaml
    orders.yaml
  environments/
    prod.yaml
    staging.yaml
  teams/
    backend.yaml
    mobile.yaml
```

### Configuration File Usage

**Simple project:**
```yaml
# typemux.config.yaml
input: schema.typemux
output:
  formats: [graphql, protobuf]
```

**Complex project:**
```yaml
# typemux.config.yaml
input: schemas/main.typemux
output:
  directory: generated
  formats:
    - graphql
    - protobuf
    - openapi
    - go
annotations:
  - annotations/base.yaml
  - annotations/prod.yaml
```

### Version Control

**Always commit:**
- Schema files (`.typemux`)
- Configuration files
- YAML annotation files

**Add to .gitignore:**
```
generated/
*.proto
*.graphql
openapi.yaml
types.go
schema.md
```

## Troubleshooting

### Common Issues

**Error: "file not found"**
- Ensure paths in `-input` are correct
- Use relative paths from current directory
- Check import paths in schema files

**Error: "unknown annotation"**
- See [Language Reference](reference.md) for valid annotations
- Check YAML syntax (indentation, quotes)
- Ensure annotation is supported for the target format

**Generated files are empty**
- Check that types/enums/services are defined
- Verify namespace declarations if using multiple files
- Check for parser errors in output

**Annotations not applied**
- Check YAML indentation (use spaces, not tabs)
- Verify type/field names match exactly (case-sensitive)
- Check annotation merge order if using multiple files
- Use qualified names when using namespaces

**Import errors**
- Verify import paths are relative to importing file
- Check for circular imports
- Ensure imported files have `.typemux` extension

### Getting Help

If you encounter issues:

1. Check the [Language Reference](reference.md) for syntax
2. Review [examples](examples.md) for similar use cases
3. Enable verbose output to see what TypeMUX is doing
4. File an issue on [GitHub](https://github.com/rasmartins/typemux/issues)

## See Also

- [Language Reference](reference.md) - Complete TypeMUX syntax and annotations
- [Tutorial](tutorial.md) - Step-by-step guide
- [Quick Start](quickstart.md) - Get started in 5 minutes
- [Examples](examples.md) - Real-world examples
