# TypeMUX

**One Schema, Three Protocols**

TypeMUX is a powerful Interface Definition Language (IDL) that generates GraphQL schemas, Protocol Buffers, and OpenAPI specifications from a single source of truth.

## Why TypeMUX?

Maintaining separate schema definitions for GraphQL, Protocol Buffers, and OpenAPI is tedious and error-prone. TypeMUX solves this by letting you write one `.typemux` schema file and automatically generating all three formats, ensuring consistency across your API specifications.

```typemux
/// User entity with authentication roles
type User {
  id: string @required
  email: string @required
  username: string @required
  role: UserRole @default("USER")
  createdAt: timestamp @required
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

From this single schema, TypeMUX generates:
- GraphQL schema with types, queries, and mutations
- Protocol Buffers (proto3) with services and messages
- OpenAPI 3.0 specification with paths and schemas

## Key Features

- **Single Source of Truth**: Write once, generate multiple formats
- **Type Safety**: Strongly typed system with primitives, enums, arrays, and maps
- **Namespace Support**: Organize types across multiple namespaces
- **Union Types**: OneOf/union support for representing multiple possible types
- **Flexible Annotations**: Inline or external YAML annotations for metadata
- **Service Definitions**: RPC-style service methods with HTTP and GraphQL mappings
- **Multi-File Schemas**: Import and include support for modular schemas
- **Custom Field Numbers**: Protobuf field numbering control
- **Documentation Comments**: Triple-slash comments for API documentation

## Quick Links

- [Quick Start](quickstart.md) - Get started in 5 minutes
- [Tutorial](tutorial.md) - Learn TypeMUX step by step
- [Language Reference](reference.md) - Complete language specification
- [Configuration](configuration.md) - CLI flags and annotations
- [Examples](examples.md) - Real-world use cases

## Installation

### From Source

```bash
git clone https://github.com/rasmartins/typemux.git
cd typemux
go build -o typemux
```

### Using Go Install

```bash
go install github.com/rasmartins/typemux@latest
```

## Basic Usage

```bash
# Generate all formats
typemux -input schema.typemux -output ./generated

# Generate specific format
typemux -input schema.typemux -format graphql -output ./generated

# With external annotations
typemux -input schema.typemux -annotations annotations.yaml -output ./generated
```

## Example Output

Given a TypeMUX schema, you get:

**GraphQL** (`schema.graphql`):
```graphql
type User {
  id: String!
  email: String!
  username: String!
  role: UserRole
  createdAt: String!
}

enum UserRole {
  ADMIN
  USER
  GUEST
}

type Query {
  GetUser(input: GetUserRequestInput!): User
}
```

**Protocol Buffers** (`schema.proto`):
```protobuf
syntax = "proto3";

message User {
  string id = 1;
  string email = 2;
  string username = 3;
  UserRole role = 4;
  google.protobuf.Timestamp createdAt = 5;
}

enum UserRole {
  ADMIN = 0;
  USER = 1;
  GUEST = 2;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
}
```

**OpenAPI** (`openapi.yaml`):
```yaml
paths:
  /api/v1/users/{id}:
    get:
      operationId: GetUser
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

components:
  schemas:
    User:
      type: object
      required: [id, email, username, createdAt]
      properties:
        id:
          type: string
        email:
          type: string
        username:
          type: string
        role:
          $ref: '#/components/schemas/UserRole'
        createdAt:
          type: string
          format: date-time
```

## Use Cases

- **API-First Development**: Design your API schema before implementation
- **Multi-Protocol Support**: Support REST, GraphQL, and gRPC from one definition
- **Microservices**: Share consistent type definitions across services
- **Code Generation**: Generate client/server code for multiple languages
- **Documentation**: Maintain up-to-date API documentation automatically

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

[License information to be added]

## Support

- GitHub Issues: [Report bugs and request features](https://github.com/rasmartins/typemux/issues)
- Documentation: [Full documentation](https://rasmartins.github.io/typemux)
