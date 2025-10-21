# TypeMux

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/rasmartins/typemux)](https://goreportcard.com/report/github.com/rasmartins/typemux)

**One Schema, Three Protocols**

TypeMux is an Interface Definition Language (IDL) and code generator that converts a single schema definition into multiple output formats: GraphQL schemas, Protocol Buffers (proto3), and OpenAPI 3.0 specifications.

## Why TypeMux?

Stop maintaining separate schema definitions. Write your API schema once in TypeMux and generate GraphQL, Protobuf, and OpenAPI automatically.

```typescript
type User {
  id: string @required
  email: string @required
  role: UserRole @default("USER")
}

enum UserRole {
  ADMIN
  USER
  GUEST
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User)
    @http(GET)
    @path("/api/v1/users/{id}")
    @graphql(query)
}
```

**Generates:**
- ✅ GraphQL schema with queries and mutations
- ✅ Protocol Buffers (proto3) with services
- ✅ OpenAPI 3.0 specification with paths

## Quick Start

```bash
# Install
go install github.com/rasmartins/typemux@latest

# Generate all formats
typemux -input schema.typemux -output ./generated
```

## Documentation

📚 **[Full Documentation](https://rasmartins.github.io/typemux)**

- [Quick Start Guide](https://rasmartins.github.io/typemux/quickstart) - Get started in 5 minutes
- [Tutorial](https://rasmartins.github.io/typemux/tutorial) - Learn TypeMux step by step
- [Language Reference](https://rasmartins.github.io/typemux/reference) - Complete syntax specification
- [Examples](https://rasmartins.github.io/typemux/examples) - Real-world use cases
- [Configuration](https://rasmartins.github.io/typemux/configuration) - CLI and annotations guide

## Features

- **Single Source of Truth** - Write once, generate multiple formats
- **Type Safety** - Strongly typed with primitives, enums, arrays, maps, and unions
- **Namespace Support** - Organize types across multiple namespaces
- **Union Types** - OneOf/sum types for polymorphic data
- **Flexible Annotations** - Inline or external YAML annotations
- **Service Definitions** - RPC-style methods with HTTP and GraphQL mappings
- **Multi-File Support** - Import and modular schemas
- **Custom Field Numbers** - Protobuf field numbering control
- **Documentation Comments** - Auto-generated API documentation

## Type System

### Primitive Types
`string` · `int32` · `int64` · `float32` · `float64` · `bool` · `timestamp` · `bytes`

### Complex Types
- Arrays: `[]TypeName`
- Maps: `map<KeyType, ValueType>`
- Enums: Named constants
- Unions: OneOf/tagged unions
- User-defined types

### Annotations
- Field: `@required` · `@default("value")` · `@exclude(format)` · `@only(format)`
- Method: `@http(METHOD)` · `@path("/api/path")` · `@graphql(type)` · `@success(code)` · `@errors(code)`

## Example Output

From a single TypeMux schema, generate:

**GraphQL**
```graphql
type User {
  id: String!
  email: String!
  role: UserRole!
}

enum UserRole {
  ADMIN
  USER
  GUEST
}
```

**Protocol Buffers**
```protobuf
message User {
  string id = 1;
  string email = 2;
  UserRole role = 3;
}

enum UserRole {
  ADMIN = 0;
  USER = 1;
  GUEST = 2;
}
```

**OpenAPI**
```yaml
components:
  schemas:
    User:
      type: object
      required: [id, email, role]
      properties:
        id:
          type: string
        email:
          type: string
        role:
          $ref: '#/components/schemas/UserRole'
```

## CLI Usage

```bash
# Generate all formats
typemux -input schema.typemux -output ./generated

# Generate specific format
typemux -input schema.typemux -format graphql -output ./gen
typemux -input schema.typemux -format protobuf -output ./gen
typemux -input schema.typemux -format openapi -output ./gen

# With external annotations
typemux -input schema.typemux -annotations annotations.yaml -output ./gen
```

## Building from Source

```bash
git clone https://github.com/rasmartins/typemux.git
cd typemux
go build -o typemux ./cmd/typemux
```

## Testing

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Project maintains 90%+ test coverage.

## Contributing

Contributions are welcome! Please see our [Contributing Guide](https://rasmartins.github.io/typemux/CONTRIBUTING) for details.

Quick steps:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests (maintain 90%+ coverage)
4. Commit (`git commit -m 'Add amazing feature'`)
5. Push (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## Project Structure

```
typemux/
├── cmd/typemux/          # CLI entry point
├── internal/
│   ├── ast/              # Abstract syntax tree
│   ├── lexer/            # Tokenization
│   ├── parser/           # Parsing
│   ├── generator/        # Code generators (GraphQL, Protobuf, OpenAPI)
│   └── annotations/      # YAML annotation handling
├── examples/             # Usage examples
├── docs/                 # Documentation (GitHub Pages)
└── vscode-extension/     # VS Code language support
```

## VS Code Extension

Syntax highlighting and snippets for `.typemux` files are available in the `vscode-extension/` directory. See [installation guide](vscode-extension/INSTALL.md).

## Use Cases

- **API-First Development** - Design APIs before implementation
- **Multi-Protocol Support** - Support REST, GraphQL, and gRPC from one definition
- **Microservices** - Share consistent type definitions across services
- **Contract Testing** - Ensure API contracts are consistent
- **Documentation** - Auto-generate up-to-date API docs

## Development

This project was mostly **vibe-coded** with AI assistance.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Links

- 📖 [Documentation](https://rasmartins.github.io/typemux)
- 🐛 [Issues](https://github.com/rasmartins/typemux/issues)
- 💬 [Discussions](https://github.com/rasmartins/typemux/discussions)

---

**TypeMux** - Write once, generate everywhere.
