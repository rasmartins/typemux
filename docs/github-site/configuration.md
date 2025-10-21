# Configuration Guide

Complete guide to configuring TypeMux code generation.

## Table of Contents

- [CLI Flags](#cli-flags)
- [Inline Annotations](#inline-annotations)
- [YAML Annotations](#yaml-annotations)
- [Annotation Merging](#annotation-merging)
- [Output Formats](#output-formats)
- [Best Practices](#best-practices)

## CLI Flags

TypeMux accepts the following command-line flags:

### -input

**Required.** Path to the input TypeMux schema file.

```bash
typemux -input schema.typemux
```

**Multiple files:** Use imports in your schema file instead of multiple `-input` flags.

### -format

Output format to generate. Default: `all`

**Options:**
- `all` - Generate all formats (default)
- `graphql` - Generate only GraphQL schema
- `protobuf` - Generate only Protocol Buffers
- `openapi` - Generate only OpenAPI specification

**Examples:**

```bash
# Generate all formats
typemux -input schema.typemux -format all

# Generate only GraphQL
typemux -input schema.typemux -format graphql

# Generate only Protobuf
typemux -input schema.typemux -format protobuf

# Generate only OpenAPI
typemux -input schema.typemux -format openapi
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

### -annotations

Path to YAML annotations file. Can be specified multiple times.

```bash
# Single annotations file
typemux -input schema.typemux -annotations annotations.yaml

# Multiple files (merged, later files override earlier)
typemux -input schema.typemux \
        -annotations base-annotations.yaml \
        -annotations environment-specific.yaml \
        -annotations overrides.yaml
```

### Complete Example

```bash
typemux \
  -input api/schema.typemux \
  -format all \
  -output ./generated/api \
  -annotations api/annotations.yaml \
  -annotations api/prod-overrides.yaml
```

## Inline Annotations

Annotations can be embedded directly in TypeMux schema files.

### Field Attributes

#### @required

Mark field as non-nullable.

```typemux
type User {
  id: string @required
  email: string @required
  nickname: string  // Optional
}
```

**Effect:**
- GraphQL: Field type becomes non-nullable (`String!`)
- OpenAPI: Field added to `required` array
- Protobuf: No effect (proto3 fields are optional by default)

#### @default

Set default value for a field.

```typemux
type Config {
  timeout: int32 @default("30")
  enabled: bool @default("true")
  mode: Mode @default("NORMAL")
  message: string @default("Hello, World!")
}
```

**Value format:**
- Strings: `@default("value")`
- Numbers: `@default("42")` or `@default("3.14")`
- Booleans: `@default("true")` or `@default("false")`
- Enums: `@default("ENUM_VALUE_NAME")`

**Effect:**
- GraphQL: Default value in schema
- OpenAPI: `default` property in schema
- Protobuf: Comment only (proto3 doesn't support defaults)

#### @exclude

Exclude field from specific output formats.

```typemux
type User {
  id: string @required
  email: string @required
  passwordHash: string @exclude(graphql,openapi)
  apiKey: string @exclude(graphql)
}
```

**Formats:**
- `graphql`
- `protobuf` (or `proto`)
- `openapi`

**Use cases:**
- Hide sensitive fields from public APIs
- Internal-only fields for backend services
- Format-specific fields

#### @only

Include field only in specific formats.

```typemux
type User {
  id: string @required
  email: string @required
  __typename: string @only(graphql)
  _links: map<string, string> @only(openapi)
}
```

**Note:** `@only` and `@exclude` are mutually exclusive.

### Method Annotations

#### @http

HTTP method for REST endpoints.

```typemux
service UserService {
  rpc GetUser(...) returns (...) @http(GET)
  rpc CreateUser(...) returns (...) @http(POST)
  rpc UpdateUser(...) returns (...) @http(PUT)
  rpc PatchUser(...) returns (...) @http(PATCH)
  rpc DeleteUser(...) returns (...) @http(DELETE)
}
```

**Methods:** `GET`, `POST`, `PUT`, `PATCH`, `DELETE`

#### @path

URL path template with variables.

```typemux
service UserService {
  rpc GetUser(...) returns (...)
    @http(GET)
    @path("/api/v1/users/{id}")

  rpc GetUserPosts(...) returns (...)
    @http(GET)
    @path("/api/v1/users/{userId}/posts/{postId}")

  rpc ListUsers(...) returns (...)
    @http(GET)
    @path("/api/v1/users")
}
```

**Path variables:**
- Enclosed in `{}`
- Match field names in request type
- Automatically extracted as path parameters in OpenAPI

#### @graphql

GraphQL operation type.

```typemux
service UserService {
  rpc GetUser(...) returns (...) @graphql(query)
  rpc ListUsers(...) returns (...) @graphql(query)
  rpc CreateUser(...) returns (...) @graphql(mutation)
  rpc WatchUsers(...) returns (...) @graphql(subscription)
}
```

**Types:** `query`, `mutation`, `subscription`

**Auto-detection:**
If not specified, inferred from method name:
- `Get*`, `List*`, `Find*`, `Search*` → `query`
- `Create*`, `Update*`, `Delete*`, `Set*` → `mutation`
- `Subscribe*`, `Watch*` → `subscription`

#### @success

HTTP success status codes.

```typemux
service UserService {
  rpc CreateUser(...) returns (...)
    @http(POST)
    @success(201)

  rpc UpdateUser(...) returns (...)
    @http(PUT)
    @success(200,204)
}
```

**Common codes:**
- `200` - OK
- `201` - Created
- `204` - No Content

#### @errors

HTTP error status codes.

```typemux
service UserService {
  rpc GetUser(...) returns (...)
    @http(GET)
    @errors(404,500)

  rpc CreateUser(...) returns (...)
    @http(POST)
    @errors(400,409,500)
}
```

**Common codes:**
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Unprocessable Entity
- `500` - Internal Server Error

### Combining Annotations

Multiple annotations can be applied:

```typemux
type User {
  id: string @required = 1
  email: string @required @default("user@example.com") = 2
  age: int32 @default("0") = 3
  internalId: int64 @exclude(graphql,openapi) = 100
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User)
    @http(GET)
    @path("/api/v1/users/{id}")
    @graphql(query)
    @success(200)
    @errors(404,500)
}
```

## YAML Annotations

External YAML files for annotations, useful for large projects or separating metadata from schema.

### File Structure

```yaml
types:
  TypeName:
    proto:
      name: "CustomProtoName"
      option: "option string"
    graphql:
      name: "CustomGraphQLName"
      directive: "directive string"
    openapi:
      name: "CustomOpenAPIName"
      extension: '{"x-custom": "value"}'
    fields:
      fieldName:
        required: true|false
        default: "value"
        proto:
          name: "custom_field_name"
        graphql:
          name: "customFieldName"
        openapi:
          name: "customFieldName"

enums:
  EnumName:
    proto:
      name: "CustomEnumName"

unions:
  UnionName:
    proto:
      name: "CustomUnionName"

services:
  ServiceName:
    methods:
      MethodName:
        http: "GET|POST|PUT|PATCH|DELETE"
        path: "/api/path"
        graphql: "query|mutation|subscription"
        success: [200, 201]
        errors: [400, 404, 500]
        proto:
          option: "option string"
```

### Example

**schema.typemux:**
```typemux
namespace com.example.api

type User {
  id: string
  email: string
  username: string
  role: UserRole
  createdAt: timestamp
}

enum UserRole {
  ADMIN
  USER
  GUEST
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User)
  rpc CreateUser(CreateUserRequest) returns (User)
}
```

**annotations.yaml:**
```yaml
types:
  User:
    proto:
      name: "UserMessage"
    graphql:
      name: "UserAccount"
    openapi:
      name: "UserProfile"
    fields:
      id:
        required: true
      email:
        required: true
        openapi:
          extension: '{"x-format": "email", "x-validate": "email"}'
      username:
        required: true
      role:
        required: true
        default: "USER"
      createdAt:
        required: true

  GetUserRequest:
    fields:
      id:
        required: true

enums:
  UserRole:
    proto:
      name: "UserRoleEnum"

services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{id}"
        graphql: "query"
        success: [200]
        errors: [404, 500]
      CreateUser:
        http: "POST"
        path: "/api/v1/users"
        graphql: "mutation"
        success: [201]
        errors: [400, 409, 500]
```

### Qualified Names (with Namespaces)

When using namespaces, use fully qualified names:

```yaml
types:
  com.example.users.User:
    fields:
      id:
        required: true

  com.example.products.Product:
    fields:
      id:
        required: true

services:
  com.example.users.UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{id}"
```

### Format-Specific Names

Override type and field names per format:

```yaml
types:
  Product:
    proto:
      name: "ProductMessage"
    graphql:
      name: "ProductType"
    openapi:
      name: "ProductSchema"
    fields:
      id:
        proto:
          name: "product_id"
        graphql:
          name: "productId"
        openapi:
          name: "id"
      categoryId:
        proto:
          name: "category_id"
        graphql:
          name: "categoryId"
```

### OpenAPI Extensions

Add custom OpenAPI extensions:

```yaml
types:
  User:
    openapi:
      extension: |
        {
          "x-displayName": "User Account",
          "x-tags": ["users", "accounts"],
          "x-permissions": ["user:read", "user:write"]
        }
    fields:
      email:
        openapi:
          extension: '{"x-format": "email", "x-validate": "email"}'
      age:
        openapi:
          extension: '{"x-minimum": 0, "x-maximum": 150}'
```

### Validation

TypeMux validates YAML annotations:

**Validated:**
- ✅ Referenced types exist
- ✅ Referenced fields exist
- ✅ Field requirements are boolean
- ✅ HTTP methods are valid
- ✅ Status codes are valid integers
- ✅ No conflicting annotations

**Errors reported:**
```
Error: Type 'InvalidType' referenced in annotations does not exist
Error: Field 'invalidField' in type 'User' does not exist
Error: Invalid HTTP method 'INVALID' for method 'GetUser'
```

## Annotation Merging

When multiple YAML annotation files are specified, they are merged with a last-wins strategy.

### Merge Order

```bash
typemux -input schema.typemux \
        -annotations base.yaml \
        -annotations env.yaml \
        -annotations overrides.yaml
```

**Order:** `base.yaml` → `env.yaml` → `overrides.yaml`

Later files override earlier ones at the field level.

### Example

**base.yaml:**
```yaml
types:
  User:
    fields:
      id:
        required: true
      email:
        required: true
      username:
        required: false
```

**overrides.yaml:**
```yaml
types:
  User:
    fields:
      username:
        required: true
        default: "anonymous"
```

**Result:**
```yaml
types:
  User:
    fields:
      id:
        required: true
      email:
        required: true
      username:
        required: true      # Overridden
        default: "anonymous" # Added
```

### Use Cases

**Environment-specific configurations:**
```bash
typemux -input schema.typemux \
        -annotations base-annotations.yaml \
        -annotations prod-annotations.yaml
```

**Layered configurations:**
```bash
typemux -input schema.typemux \
        -annotations team-standards.yaml \
        -annotations project-specific.yaml \
        -annotations local-overrides.yaml
```

**A/B testing:**
```bash
# Variant A
typemux -input schema.typemux \
        -annotations base.yaml \
        -annotations variant-a.yaml

# Variant B
typemux -input schema.typemux \
        -annotations base.yaml \
        -annotations variant-b.yaml
```

## Output Formats

### GraphQL

**File:** `schema.graphql`

**Features:**
- Type definitions
- Enum definitions
- Union types with `@oneOf` directive
- Input types for mutations (auto-generated with `Input` suffix)
- Query type (from `query` methods)
- Mutation type (from `mutation` methods)
- Subscription type (from `subscription` methods)
- Documentation comments as GraphQL descriptions

**Limitations:**
- All type names must be unique (no namespace support)
- Maps become `JSON` scalar
- Timestamps become `String`

### Protocol Buffers

**File:** `schema.proto` (or namespace-specific files)

**Features:**
- Proto3 syntax
- Message definitions
- Enum definitions (with `UNSPECIFIED` value at 0)
- Service definitions with RPC methods
- Union types as `oneof`
- Custom field numbers
- Google Protobuf Timestamp import for timestamp fields
- Map support
- Documentation comments

**Namespace handling:**
- Single namespace: `schema.proto`
- Multiple namespaces: `namespace.name.proto` per namespace

**Generated:**
```protobuf
syntax = "proto3";

import "google/protobuf/timestamp.proto";

package com.example.api;

message User {
  string id = 1;
  string email = 2;
  google.protobuf.Timestamp createdAt = 3;
  map<string, string> metadata = 4;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
}
```

### OpenAPI

**File:** `openapi.yaml`

**Features:**
- OpenAPI 3.0.0 specification
- Schema definitions for all types
- Path operations from service methods
- Path parameters from URL templates
- Request body schemas
- Response schemas
- HTTP status codes
- Enum support
- Array and map support
- Documentation as descriptions
- Custom extensions

**Generated structure:**
```yaml
openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths:
  /api/v1/users/{id}:
    get:
      operationId: GetUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: Not found
        '500':
          description: Internal server error
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
          format: email
```

## Best Practices

### When to Use Inline vs YAML Annotations

**Use inline annotations when:**
- ✅ Schema is small (<500 lines)
- ✅ Annotations are simple and stable
- ✅ Schema and metadata are tightly coupled
- ✅ Single-person or small team

**Use YAML annotations when:**
- ✅ Schema is large (>500 lines)
- ✅ Complex metadata (many custom names, extensions)
- ✅ Multi-environment deployments
- ✅ Non-technical team members need to edit metadata
- ✅ Annotations change frequently
- ✅ Sharing base schema across projects

### Annotation Organization

**Recommended structure for large projects:**

```
api/
├── schemas/
│   ├── types/
│   │   ├── user.typemux
│   │   ├── product.typemux
│   │   └── order.typemux
│   └── services/
│       ├── user_service.typemux
│       ├── product_service.typemux
│       └── order_service.typemux
├── annotations/
│   ├── base.yaml
│   ├── development.yaml
│   ├── staging.yaml
│   └── production.yaml
└── main.typemux
```

### Field Number Ranges

**Protobuf field numbering best practices:**

- **1-15:** Frequently used fields (more efficient encoding)
- **16-99:** Common fields
- **100-999:** Occasional fields
- **1000+:** Rare or internal fields
- **19000-19999:** Reserved by Protobuf (do not use)

```typemux
type User {
  id: string = 1              // Frequently used
  email: string = 2           // Frequently used
  name: string = 3            // Frequently used
  role: UserRole = 10         // Common
  createdAt: timestamp = 11   // Common
  lastLogin: timestamp = 100  // Occasional
  metadata: map<string, string> = 1000  // Rare
}
```

### Documentation

Always document:
- Public types and fields
- Service methods
- Complex enums
- Default values and their meaning
- Field constraints and validation rules

```typemux
/// User account with authentication and profile information
///
/// Users are created through the registration flow and can be
/// assigned different roles for access control.
type User {
  /// Unique immutable user identifier (UUID v4)
  id: string @required

  /// Email address used for login (must be unique)
  email: string @required

  /// User's role determining permissions (default: USER)
  role: UserRole @default("USER")
}
```

### Version Control

**Commit separately:**
1. Schema changes (`.typemux` files)
2. Annotation changes (`.yaml` files)
3. Generated code

**Add to `.gitignore`:**
```
generated/
*.pb.go
*.graphql.ts
```

**Track in version control:**
```
schemas/
annotations/
README.md
```

### Testing Generated Schemas

Validate generated schemas:

```bash
# Validate GraphQL schema
npx graphql-schema-linter generated/schema.graphql

# Compile Protobuf
protoc --proto_path=. --go_out=. generated/schema.proto

# Validate OpenAPI
npx @stoplight/spectral-cli lint generated/openapi.yaml
```

### Continuous Integration

Example GitHub Actions workflow:

```yaml
name: Generate Schemas

on: [push, pull_request]

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build TypeMux
        run: go build -o typemux

      - name: Generate schemas
        run: |
          ./typemux -input schema.typemux -output ./generated

      - name: Validate GraphQL
        run: npx graphql-schema-linter generated/schema.graphql

      - name: Validate Protobuf
        run: protoc --proto_path=. --descriptor_set_out=/dev/null generated/schema.proto

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: generated-schemas
          path: generated/
```

## Troubleshooting

### Common Issues

**Issue: Type not found in annotations**
```
Error: Type 'Usr' referenced in annotations does not exist
```
**Solution:** Check spelling and ensure type exists in schema. Use qualified names with namespaces.

**Issue: Circular import detected**
```
Error: Circular import detected: a.typemux -> b.typemux -> a.typemux
```
**Solution:** Extract common types to a third file, or restructure imports.

**Issue: GraphQL name collision**
```
Error: Duplicate type name 'User' in GraphQL generation
```
**Solution:** Use YAML annotations to give unique GraphQL names, or use `@only` to exclude from GraphQL.

**Issue: Invalid field number**
```
Error: Field number 19500 is in reserved range
```
**Solution:** Use field numbers outside 19000-19999 range.

### Debug Mode

Use verbose output to debug generation:

```bash
# Enable Go logging (if supported)
TYPEMUX_DEBUG=1 ./typemux -input schema.typemux -output ./generated
```

## See Also

- [Quick Start](quickstart.md) - Get started quickly
- [Tutorial](tutorial.md) - Learn step by step
- [Language Reference](reference.md) - Complete syntax reference
- [Examples](examples.md) - Real-world examples
