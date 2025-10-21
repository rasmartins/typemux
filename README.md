# TypeMux

A powerful Interface Definition Language (IDL) that can generate GraphQL schemas, Protobuf definitions, and OpenAPI specifications from a single schema definition.

## Features

- **Single Source of Truth**: Define your API schema once in a simple, intuitive syntax
- **Multiple Output Formats**:
  - GraphQL schemas
  - Protocol Buffers (proto3)
  - OpenAPI 3.0 specifications
- **Type Safe**: Strongly typed with support for primitives, enums, arrays, maps, and unions
- **Service Definitions**: Define RPC-style service methods with REST and GraphQL annotations
- **Modular Design**:
  - Import/include support with circular dependency detection
  - Namespace support for organizing types across packages
  - Cross-namespace type references
- **Flexible Annotations**:
  - Inline annotations in schema files
  - External YAML annotation files
  - Leading or trailing annotation syntax
  - Format-specific name overrides (`@proto.name`, `@graphql.name`, `@openapi.name`)
- **Field Attributes**: Add metadata like `@required`, `@default`, `@exclude`, `@only`
- **Documentation Comments**: Triple-slash comments (`///`) for documenting types, fields, and methods
- **REST Annotations**: `@http`, `@path`, `@success`, `@errors` for OpenAPI generation
- **Custom Field Numbers**: Specify protobuf field numbers explicitly

## IDL Syntax

### Basic Types

Supported primitive types:
- `string` - String values
- `int32` - 32-bit integers
- `int64` - 64-bit integers
- `float32` - 32-bit floating point
- `float64` - 64-bit floating point
- `bool` - Boolean values
- `timestamp` - Timestamp/datetime values
- `bytes` - Binary data

### Complex Types

- **Arrays**: `[]typeName` - e.g., `[]string`, `[]User`
- **Maps**: `map<keyType, valueType>` - e.g., `map<string, string>`
- **Unions**: `union Name { Type1 Type2 Type3 }` - oneOf/sum types

### Imports and Namespaces

```
// Import other schema files
import "common.typemux"

// Define a namespace for this schema
namespace com.example.userservice
```

### Enums

```
enum UserRole {
  ADMIN = 1
  USER = 2
  GUEST = 3
}
```

Custom enum values are optional. If not specified, they will be auto-numbered starting from 1.

### Types

```
/// User entity with profile information
type User {
  id: string = 1 @required
  name: string = 2 @required
  email: string = 3 @required
  age: int32 = 4
  role: UserRole = 5 @required
  isActive: bool = 6 @default(true)
  tags: []string = 7
  metadata: map<string, string> = 8
  internalField: string = 9 @exclude(graphql)
}
```

Field numbers (e.g., `= 1`) are optional but recommended for Protobuf compatibility. Documentation comments use `///`.

### Unions

```
/// A payment method can be one of several types
union PaymentMethod {
  CreditCard
  BankTransfer
  PayPal
}
```

### Services

```
service UserService {
  /// Create a new user
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
    @http(POST)
    @path("/api/v1/users")
    @graphql(mutation)
    @success(201)
    @errors(400,409,500)

  /// Get a user by ID
  rpc GetUser(GetUserRequest) returns (GetUserResponse)
    @http(GET)
    @path("/api/v1/users/{id}")
    @graphql(query)
    @errors(404,500)
}
```

### Attributes

**Field Attributes:**
- `@required` - Mark field as required (non-nullable)
- `@default(value)` - Set default value for field
- `@exclude(format)` - Exclude field from specific format (e.g., `@exclude(graphql)`)
- `@only(format)` - Only include field in specific format (e.g., `@only(proto)`)

**Method Annotations:**
- `@http(METHOD)` - HTTP method for REST endpoints (GET, POST, PUT, DELETE, PATCH)
- `@path(template)` - URL path template (e.g., `/api/v1/users/{id}`)
- `@graphql(type)` - GraphQL operation type (query, mutation, subscription)
- `@success(code)` - Additional success HTTP status codes (e.g., `@success(201)`)
- `@errors(codes)` - Expected error HTTP status codes (e.g., `@errors(400,404,500)`)

**Format-Specific Name Overrides:**
- `@proto.name(name)` - Override type/field name in Protobuf output
- `@graphql.name(name)` - Override type/field name in GraphQL output
- `@openapi.name(name)` - Override type/field name in OpenAPI output

## Installation

```bash
# Clone the repository
git clone https://github.com/rasmartins/typemux.git
cd typemux

# Build the CLI tool
go build -o typemux ./cmd/typemux

# Or run directly
go run ./cmd/typemux/main.go
```

## Usage

### Generate all formats

```bash
./typemux -input example.typemux -output ./generated
```

### Generate specific format

```bash
# GraphQL only
./typemux -input example.typemux -format graphql -output ./generated

# Protobuf only
./typemux -input example.typemux -format protobuf -output ./generated

# OpenAPI only
./typemux -input example.typemux -format openapi -output ./generated
```

### CLI Options

- `-input` (required): Path to the input IDL schema file
- `-format`: Output format - `graphql`, `protobuf`, `openapi`, or `all` (default: `all`)
- `-output`: Output directory for generated files (default: `./generated`)
- `-annotations`: YAML annotations file (can be specified multiple times)

### YAML Annotations

TypeMux supports defining annotations in external YAML files instead of inline in `.typemux` files. This provides better separation of concerns and easier annotation management.

**Usage:**
```bash
# Single annotations file
./typemux -input schema.typemux -annotations annotations.yaml

# Multiple files (merged in order, later files override)
./typemux -input schema.typemux \
          -annotations base.yaml \
          -annotations overrides.yaml
```

**Example annotations.yaml:**
```yaml
types:
  User:
    proto:
      name: "UserV2"
    graphql:
      name: "UserAccount"
    openapi:
      name: "UserProfile"
    fields:
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'

services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{userId}"
        graphql: "query"
        errors: [404, 500]
```

**Features:**
- YAML annotations override inline annotations
- Full validation with helpful error messages
- Support for all annotation types (type names, field attributes, method configs)
- See `examples/yaml-annotations/` for a complete example
- Full documentation: `docs/YAML_ANNOTATIONS.md`

## Examples

The `examples/` directory contains comprehensive examples demonstrating all features:

- `examples/basic/` - Simple types, enums, and services
- `examples/unions/` - Union/oneOf types for polymorphism
- `examples/namespaces/` - Namespace organization
- `examples/imports/` - Multi-file schemas with imports
- `examples/yaml-annotations/` - External YAML annotations
- `examples/annotations/` - Inline format-specific annotations
- `examples/custom-field-numbers/` - Explicit protobuf field numbering
- `examples/status-codes/` - HTTP status code annotations

Run any example:

```bash
go run ./cmd/typemux/main.go -input examples/basic/basic.typemux -output examples/basic/generated
```

This will generate:
- `schema.graphql` - GraphQL schema
- `schema.proto` - Protocol Buffers definition
- `openapi.yaml` - OpenAPI specification

## Generated Output

### GraphQL Schema

The generator creates a complete GraphQL schema with:
- Type definitions with proper nullable/non-nullable fields
- Enum types
- Query and Mutation types based on service methods

### Protobuf Schema

The generator creates a proto3 schema with:
- Message types with proper field numbering
- Enum definitions with required UNSPECIFIED value
- Service definitions with RPC methods
- Proper type mappings including timestamp support

### OpenAPI Specification

The generator creates an OpenAPI 3.0 specification with:
- Schema definitions for all types
- Path operations based on service methods
- Request/response schemas
- Proper type and format specifications

## Project Structure

```
.
├── cmd/
│   └── typemux/
│       └── main.go         # CLI entry point
├── internal/
│   ├── ast/                # Abstract Syntax Tree definitions
│   │   ├── ast.go
│   │   └── ast_test.go
│   ├── lexer/              # Lexical analyzer
│   │   ├── lexer.go
│   │   └── lexer_test.go
│   ├── parser/             # Parser implementation
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── generator/          # Code generators
│   │   ├── graphql.go
│   │   ├── protobuf.go
│   │   ├── openapi.go
│   │   └── *_test.go
│   └── annotations/        # YAML annotations support
│       ├── yaml.go
│       ├── validator.go
│       └── merger.go
├── examples/               # Example schemas
├── docs/                   # Documentation
└── README.md
```

## Architecture

1. **Lexer**: Tokenizes the input schema file
2. **Parser**: Builds an Abstract Syntax Tree (AST) from tokens with support for:
   - Imports and namespace resolution
   - Type registry for cross-file references
   - Circular dependency detection
3. **Annotations**: YAML annotation loader, validator, and merger
4. **Generators**: Transform the AST into target formats (GraphQL, Protobuf, OpenAPI)
   - Format-specific type mapping
   - Documentation generation
   - Multi-file output for namespaces

## Future Enhancements

- [ ] Custom scalar types
- [ ] Advanced field validation annotations (min, max, pattern, etc.)
- [ ] Code generation for client/server stubs
- [ ] JSON Schema output format
- [ ] TypeScript type definitions
- [ ] gRPC gateway annotations
