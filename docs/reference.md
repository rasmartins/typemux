# Language Reference

Complete specification of the TypeMUX IDL syntax and semantics.

## Table of Contents

- [File Structure](#file-structure)
- [Primitive Types](#primitive-types)
- [Type Definitions](#type-definitions)
- [Enum Definitions](#enum-definitions)
- [Union Definitions](#union-definitions)
- [Service Definitions](#service-definitions)
- [Field Attributes](#field-attributes)
- [Method Annotations](#method-annotations)
- [Documentation Comments](#documentation-comments)
- [Namespaces](#namespaces)
- [Imports](#imports)
- [Type Mappings](#type-mappings)

## File Structure

A TypeMUX schema file (`.typemux`) has the following structure:

```typemux
[namespace IDENTIFIER]

[import "path/to/file.typemux"]*

[enum DEFINITION]*
[type DEFINITION]*
[union DEFINITION]*
[service DEFINITION]*
```

Elements can appear in any order, but best practice is:
1. Namespace declaration (if used)
2. Import statements
3. Enums
4. Types
5. Unions
6. Services

## Primitive Types

TypeMUX supports these built-in types:

| Type | Description | Size/Range |
|------|-------------|------------|
| `string` | UTF-8 text | Variable length |
| `int32` | Signed 32-bit integer | -2,147,483,648 to 2,147,483,647 |
| `int64` | Signed 64-bit integer | -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807 |
| `uint8` | Unsigned 8-bit integer | 0 to 255 |
| `uint16` | Unsigned 16-bit integer | 0 to 65,535 |
| `uint32` | Unsigned 32-bit integer | 0 to 4,294,967,295 |
| `uint64` | Unsigned 64-bit integer | 0 to 18,446,744,073,709,551,615 |
| `float32` | 32-bit floating point | IEEE 754 single precision |
| `float64` | 64-bit floating point | IEEE 754 double precision |
| `bool` | Boolean value | `true` or `false` |
| `timestamp` | Date and time | ISO 8601 / Unix timestamp |
| `bytes` | Binary data | Variable length |

## Type Definitions

### Basic Syntax

```typemux
type TYPENAME {
  fieldName: fieldType
  [fieldName: fieldType]*
}
```

### Example

```typemux
type User {
  id: string
  email: string
  age: int32
  active: bool
  createdAt: timestamp
}
```

### Field Types

Fields can be:
- **Primitive types**: `string`, `int32`, etc.
- **User-defined types**: Other type or enum names
- **Arrays**: `[]TypeName`
- **Maps**: `map<KeyType, ValueType>`

### Arrays

Array syntax uses brackets:

```typemux
type Post {
  tags: []string
  comments: []Comment
}
```

### Maps

Map syntax specifies key and value types:

```typemux
type Configuration {
  settings: map<string, string>
  scores: map<string, int32>
}
```

**Map constraints:**
- Key type must be `string` or an integer type
- Value type can be any type (primitive, user-defined, array, or nested map)
- Nested maps are fully supported: `map<string, map<string, int32>>`

**Nested map support:**

TypeMUX fully supports nested map syntax:

```typemux
type NestedMapExample {
  // Simple map
  settings: map<string, string>

  // Nested map (two levels)
  nested: map<string, map<string, int32>>

  // Triple nested map (three levels)
  deep: map<string, map<string, map<string, bool>>>
}
```

Each generator handles nested maps according to its schema capabilities:
- **GraphQL**: Auto-generates wrapper types (MapWrapper0, MapWrapper1, etc.)
- **Protobuf**: Uses native nested map syntax
- **OpenAPI**: Uses nested additionalProperties structures

### Nested Types

Types can reference other types:

```typemux
type Address {
  street: string
  city: string
  zipCode: string
}

type User {
  id: string
  name: string
  address: Address
}
```

## Enum Definitions

### Basic Syntax

```typemux
enum ENUMNAME {
  VALUE1
  VALUE2
  [VALUE_N]*
}
```

### Example

```typemux
enum UserRole {
  ADMIN
  MODERATOR
  USER
  GUEST
}
```

### Custom Values

Assign explicit numeric values:

```typemux
enum Status {
  UNKNOWN = 0
  ACTIVE = 1
  INACTIVE = 2
  DELETED = 99
}
```

**Rules:**
- Values must be non-negative integers
- Values need not be sequential
- If no value is specified, auto-incrementing starts from 0

### Protobuf Enum Generation

Protobuf enums always include an `UNSPECIFIED` value at 0:

```protobuf
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  ADMIN = 1;
  MODERATOR = 2;
  USER = 3;
  GUEST = 4;
}
```

## Union Definitions

Unions represent a value that can be one of several types (sum types, tagged unions, oneOf).

### Syntax

```typemux
union UNIONNAME {
  TypeName1
  TypeName2
  [TypeName_N]*
}
```

### Example

```typemux
type TextContent {
  text: string
}

type ImageContent {
  url: string
  width: int32
  height: int32
}

type VideoContent {
  url: string
  duration: int32
}

union Content {
  TextContent
  ImageContent
  VideoContent
}
```

### Generated Code

**GraphQL:**
```graphql
union Content @oneOf = TextContent | ImageContent | VideoContent
```

**Protobuf:**
```protobuf
message Content {
  oneof value {
    TextContent text_content = 1;
    ImageContent image_content = 2;
    VideoContent video_content = 3;
  }
}
```

**OpenAPI:**
```yaml
Content:
  oneOf:
    - $ref: '#/components/schemas/TextContent'
    - $ref: '#/components/schemas/ImageContent'
    - $ref: '#/components/schemas/VideoContent'
```

## Service Definitions

Services define RPC-style methods for APIs.

### Syntax

```typemux
service SERVICENAME {
  rpc MethodName(InputType) returns (OutputType)
    [@ANNOTATION]*
  [rpc ...]*
}
```

### Example

```typemux
service UserService {
  rpc GetUser(GetUserRequest) returns (User)
  rpc CreateUser(CreateUserRequest) returns (User)
  rpc UpdateUser(UpdateUserRequest) returns (User)
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse)
}
```

### Method Naming Conventions

GraphQL operation types are inferred from method names:

- Methods starting with `Get`, `List`, `Find`, `Search` → `query`
- Methods starting with `Create`, `Update`, `Delete`, `Set` → `mutation`
- Methods starting with `Subscribe`, `Watch` → `subscription`

Override with `@graphql()` annotation.

## Field Attributes

Attributes modify field behavior and generation.

### @required

Marks a field as non-nullable.

**Syntax:** `@required`

**Example:**
```typemux
type User {
  id: string @required
  email: string @required
  nickname: string
}
```

**Generated GraphQL:**
```graphql
type User {
  id: String!
  email: String!
  nickname: String
}
```

**Generated OpenAPI:**
```yaml
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
    nickname:
      type: string
```

### @default

Sets a default value for a field.

**Syntax:** `@default("value")`

**Example:**
```typemux
type User {
  id: string @required
  active: bool @default("true")
  role: UserRole @default("USER")
  loginCount: int32 @default("0")
}
```

**Value format:**
- Strings: `@default("hello")`
- Numbers: `@default("42")` or `@default("3.14")`
- Booleans: `@default("true")` or `@default("false")`
- Enums: `@default("ENUM_VALUE")`

### @exclude

Excludes a field from specific output formats.

**Syntax:** `@exclude(format1,format2,...)`

**Formats:** `graphql`, `protobuf`, `openapi`

**Example:**
```typemux
type User {
  id: string @required
  email: string @required
  passwordHash: string @exclude(graphql,openapi)
  internalId: int64 @exclude(graphql)
}
```

The `passwordHash` appears only in Protobuf.
The `internalId` appears in Protobuf and OpenAPI, but not GraphQL.

### @only

Includes a field only in specific output formats.

**Syntax:** `@only(format)`

**Example:**
```typemux
type User {
  id: string @required
  email: string @required
  __typename: string @only(graphql)
}
```

The `__typename` field appears only in GraphQL schema.

### Custom Field Numbers

Assign explicit Protobuf field numbers using `= N`.

**Syntax:** `fieldName: type = NUMBER`

**Example:**
```typemux
type User {
  id: string = 1
  email: string = 2
  name: string = 10
  age: int32 = 20
}
```

**Generated Protobuf:**
```protobuf
message User {
  string id = 1;
  string email = 2;
  string name = 10;
  int32 age = 20;
}
```

**Rules:**
- Field numbers must be positive integers (1-536,870,911)
- Field numbers 19000-19999 are reserved by Protobuf
- Field numbers should be unique within a message

### Combining Attributes

Multiple attributes can be applied to a single field:

```typemux
type Product {
  id: string @required = 1
  name: string @required @default("Unnamed") = 2
  price: float64 @required = 3
  internalNotes: string @exclude(graphql,openapi) = 100
}
```

## Method Annotations

Annotations provide metadata for service methods.

### @http.method

Specifies the HTTP method for REST endpoints.

**Syntax:** `@http.method(METHOD)`

**Methods:** `GET`, `POST`, `PUT`, `PATCH`, `DELETE`

**Example:**
```typemux
service UserService {
  rpc GetUser(GetUserRequest) returns (User)
    @http.method(GET)

  rpc CreateUser(CreateUserRequest) returns (User)
    @http.method(POST)

  rpc UpdateUser(UpdateUserRequest) returns (User)
    @http.method(PUT)

  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse)
    @http.method(DELETE)
}
```

### @http.path

Defines the URL path template.

**Syntax:** `@http.path("URL_PATH")`

**Path parameters:** Use `{paramName}` for variables

**Example:**
```typemux
service UserService {
  rpc GetUser(GetUserRequest) returns (User)
    @http.method(GET)
    @http.path("/api/v1/users/{id}")

  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse)
    @http.method(GET)
    @http.path("/api/v1/users")

  rpc GetUserPosts(GetUserPostsRequest) returns (GetUserPostsResponse)
    @http.method(GET)
    @http.path("/api/v1/users/{userId}/posts")
}
```

Path parameters are extracted from request type fields.

### @graphql

Specifies the GraphQL operation type.

**Syntax:** `@graphql(OPERATION_TYPE)`

**Operation types:** `query`, `mutation`, `subscription`

**Example:**
```typemux
service UserService {
  rpc GetUser(GetUserRequest) returns (User)
    @graphql(query)

  rpc CreateUser(CreateUserRequest) returns (User)
    @graphql(mutation)

  rpc WatchUser(WatchUserRequest) returns (User)
    @graphql(subscription)
}
```

**Default behavior:**
- If not specified, operation type is inferred from method name
- `Get*`, `List*`, `Find*`, `Search*` → `query`
- `Create*`, `Update*`, `Delete*`, `Set*` → `mutation`
- `Subscribe*`, `Watch*` → `subscription`

### @http.success

Lists HTTP success status codes.

**Syntax:** `@http.success(CODE1,CODE2,...)`

**Example:**
```typemux
service UserService {
  rpc CreateUser(CreateUserRequest) returns (User)
    @http.method(POST)
    @http.path("/api/v1/users")
    @http.success(201)

  rpc UpdateUser(UpdateUserRequest) returns (User)
    @http.method(PUT)
    @http.path("/api/v1/users/{id}")
    @http.success(200,204)
}
```

**Common codes:**
- `200` - OK
- `201` - Created
- `204` - No Content

### @http.errors

Lists HTTP error status codes.

**Syntax:** `@http.errors(CODE1,CODE2,...)`

**Example:**
```typemux
service UserService {
  rpc GetUser(GetUserRequest) returns (User)
    @http.method(GET)
    @http.path("/api/v1/users/{id}")
    @http.success(200)
    @http.errors(404,500)

  rpc CreateUser(CreateUserRequest) returns (User)
    @http.method(POST)
    @http.path("/api/v1/users")
    @http.success(201)
    @http.errors(400,409,500)
}
```

**Common codes:**
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

### Complete Method Example

```typemux
service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product)
    @http.method(GET)
    @http.path("/api/v1/products/{id}")
    @graphql(query)
    @http.success(200)
    @http.errors(404,500)

  rpc CreateProduct(CreateProductRequest) returns (Product)
    @http.method(POST)
    @http.path("/api/v1/products")
    @graphql(mutation)
    @http.success(201)
    @http.errors(400,409,500)
}
```

### Annotation Order

Annotations can be in any order:

```typemux
// Both are valid
rpc GetUser(GetUserRequest) returns (User)
  @http.method(GET)
  @http.path("/users/{id}")
  @graphql(query)

rpc GetUser(GetUserRequest) returns (User)
  @graphql(query)
  @http.path("/users/{id}")
  @http.method(GET)
```

Best practice: Use consistent ordering (http, path, graphql, success, errors).

## Documentation Comments

Add documentation using triple-slash comments (`///`).

### Syntax

```typemux
/// Documentation comment
/// Can span multiple lines
type TypeName {
  /// Field documentation
  fieldName: fieldType
}
```

### Example

```typemux
/// User account with authentication details
///
/// Users can have different roles and permissions
/// based on their account type.
type User {
  /// Unique user identifier
  ///
  /// This ID is immutable once created.
  id: string @required

  /// User's email address for login
  email: string @required

  /// Display name shown to other users
  username: string @required

  /// Account role determining permissions
  role: UserRole @required
}

/// User role enumeration
///
/// Roles are hierarchical: ADMIN > MODERATOR > USER > GUEST
enum UserRole {
  /// Full system access
  ADMIN

  /// Can moderate content
  MODERATOR

  /// Regular user
  USER

  /// Limited read-only access
  GUEST
}

/// Service for managing user accounts
///
/// Provides CRUD operations for user management.
service UserService {
  /// Retrieves a user by their unique ID
  ///
  /// Returns 404 if user is not found.
  rpc GetUser(GetUserRequest) returns (User)
    @http.method(GET)
    @http.path("/api/v1/users/{id}")
    @graphql(query)
}
```

### Documentation in Generated Code

**GraphQL:**
```graphql
"""
User account with authentication details

Users can have different roles and permissions
based on their account type.
"""
type User {
  """
  Unique user identifier

  This ID is immutable once created.
  """
  id: String!
}
```

**Protobuf:**
```protobuf
// User account with authentication details
//
// Users can have different roles and permissions
// based on their account type.
message User {
  // Unique user identifier
  //
  // This ID is immutable once created.
  string id = 1;
}
```

**OpenAPI:**
```yaml
User:
  type: object
  description: |
    User account with authentication details

    Users can have different roles and permissions
    based on their account type.
  properties:
    id:
      type: string
      description: |
        Unique user identifier

        This ID is immutable once created.
```

## Namespaces

Namespaces organize types and prevent naming conflicts.

### Syntax

```typemux
namespace NAMESPACE_IDENTIFIER

type TypeName {
  // ...
}
```

### Namespace Format

Use reverse domain notation:

```typemux
namespace com.example.api
namespace org.mycompany.users
namespace io.github.username.project
```

### Example

```typemux
namespace com.example.ecommerce

type Product {
  id: string @required
  name: string @required
}

type Order {
  id: string @required
  product: Product @required
}
```

### Multiple Namespaces

Different files can have different namespaces:

**com/example/users.typemux:**
```typemux
namespace com.example.users

type User {
  id: string @required
  email: string @required
}
```

**com/example/orders.typemux:**
```typemux
namespace com.example.orders

import "com/example/users.typemux"

type Order {
  id: string @required
  userId: string @required
}
```

### Namespace Effects

**GraphQL:**
- Namespace is ignored (GraphQL doesn't support namespaces)
- All types must have unique names across all namespaces

**Protobuf:**
- Generates separate `.proto` files per namespace
- File named: `namespace.proto` (e.g., `com.example.users.proto`)
- Package declaration: `package namespace;`

**OpenAPI:**
- Namespace can be added as schema prefix (configurable)
- Typically ignored for flat schema structure

## Imports

Import types from other files.

### Syntax

```typemux
import "relative/path/to/file.typemux"
```

### Example

**types.typemux:**
```typemux
type User {
  id: string @required
  email: string @required
}

type Product {
  id: string @required
  name: string @required
}
```

**services.typemux:**
```typemux
import "types.typemux"

type GetUserRequest {
  id: string @required
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User)
}
```

### Import Resolution

- Paths are relative to the importing file
- Use forward slashes (`/`) for path separators
- File extension `.typemux` is required

### Circular Imports

TypeMUX detects circular imports and reports an error:

**a.typemux:**
```typemux
import "b.typemux"

type A {
  b: B
}
```

**b.typemux:**
```typemux
import "a.typemux"  // Error: circular import

type B {
  a: A
}
```

**Solution:** Extract common types to a third file.

## Type Mappings

How TypeMUX types map to output formats.

### Complete Mapping Table

| TypeMUX | GraphQL | Protobuf | OpenAPI |
|---------|---------|----------|---------|
| `string` | `String` | `string` | `type: string` |
| `int32` | `Int` | `int32` | `type: integer, format: int32` |
| `int64` | `Int` | `int64` | `type: integer, format: int64` |
| `uint8` | `Int` | `uint32` | `type: integer, format: int32, minimum: 0` |
| `uint16` | `Int` | `uint32` | `type: integer, format: int32, minimum: 0` |
| `uint32` | `Int` | `uint32` | `type: integer, format: int32, minimum: 0` |
| `uint64` | `Int` | `uint64` | `type: integer, format: int64, minimum: 0` |
| `float32` | `Float` | `float` | `type: number, format: float` |
| `float64` | `Float` | `double` | `type: number, format: double` |
| `bool` | `Boolean` | `bool` | `type: boolean` |
| `timestamp` | `String` | `google.protobuf.Timestamp` | `type: string, format: date-time` |
| `bytes` | `String` | `bytes` | `type: string, format: byte` |
| `[]T` | `[T]` | `repeated T` | `type: array, items: {T}` |
| `map<K,V>` | `[KeyValueEntry!]` (typed) | `map<K, V>` | `type: object, additionalProperties: {V}` |

### Nullability

**TypeMUX:**
```typemux
type User {
  id: string @required
  nickname: string
}
```

**GraphQL:**
```graphql
type User {
  id: String!      # Non-null
  nickname: String # Nullable
}
```

**OpenAPI:**
```yaml
User:
  required:
    - id
  properties:
    id:
      type: string
    nickname:
      type: string
```

**Protobuf:**
```protobuf
// All fields are optional in proto3
message User {
  string id = 1;
  string nickname = 2;
}
```

### Arrays

**TypeMUX:**
```typemux
type Post {
  tags: []string
}
```

**GraphQL:**
```graphql
type Post {
  tags: [String]
}
```

**Protobuf:**
```protobuf
message Post {
  repeated string tags = 1;
}
```

**OpenAPI:**
```yaml
Post:
  properties:
    tags:
      type: array
      items:
        type: string
```

### Maps

**TypeMUX:**
```typemux
type Config {
  settings: map<string, string>
  scores: map<string, int64>
}
```

**GraphQL:**
Maps are converted to strongly-typed KeyValue entry lists:
```graphql
"StringStringEntry represents a key-value pair for map<string, string>"
type StringStringEntry {
  key: String!
  value: String!
}

"StringStringEntryInput represents a key-value pair for map<string, string>"
input StringStringEntryInput {
  key: String!
  value: String!
}

"StringIntEntry represents a key-value pair for map<string, int64>"
type StringIntEntry {
  key: String!
  value: Int!
}

"StringIntEntryInput represents a key-value pair for map<string, int64>"
input StringIntEntryInput {
  key: String!
  value: Int!
}

type Config {
  settings: [StringStringEntry!]
  scores: [StringIntEntry!]
}
```

**Protobuf:**
Uses native map syntax:
```protobuf
message Config {
  map<string, string> settings = 1;
  map<string, int64> scores = 2;
}
```

**OpenAPI:**
Uses `additionalProperties` to specify value types:
```yaml
Config:
  type: object
  properties:
    settings:
      type: object
      description: Map of string to string
      additionalProperties:
        type: string
    scores:
      type: object
      description: Map of string to int64
      additionalProperties:
        type: integer
        format: int64
```

**Maps with custom types:**

When using custom types as map values, all formats properly reference the type:

```typemux
type User {
  name: string
  age: int32
}

type Department {
  users: map<string, User>
}
```

- **GraphQL:** Generates `StringUserEntry` type with `value: User!`
- **Protobuf:** `map<string, User> users = 1;`
- **OpenAPI:** `additionalProperties: { $ref: '#/components/schemas/User' }`

**Nested maps:**

TypeMUX supports nested maps, with each generator handling them according to its schema capabilities:

```typemux
type NestedMapExample {
  // Nested map (two levels)
  nested: map<string, map<string, int32>>

  // Triple nested map (three levels)
  deep: map<string, map<string, map<string, bool>>>
}
```

**GraphQL:**
Auto-generates wrapper types for nested map levels:
```graphql
"MapWrapper0 is an auto-generated wrapper for nested map"
type MapWrapper0 {
  value: [StringIntEntry!]!
}

"StringMapWrapper0Entry represents a key-value pair for map<string, MapWrapper0>"
type StringMapWrapper0Entry {
  key: String!
  value: MapWrapper0!
}

type NestedMapExample {
  nested: [StringMapWrapper0Entry!]
  deep: [StringMapWrapper1Entry!]
}
```

**Protobuf:**
Uses native nested map syntax:
```protobuf
message NestedMapExample {
  map<string, map<string, int32>> nested = 1;
  map<string, map<string, map<string, bool>>> deep = 2;
}
```

**OpenAPI:**
Uses nested additionalProperties structures:
```yaml
NestedMapExample:
  type: object
  properties:
    nested:
      type: object
      description: Map of string to Map of string to int32
      additionalProperties:
        type: object
        additionalProperties:
          type: integer
          format: int32
    deep:
      type: object
      description: Map of string to Map of string to Map of string to bool
      additionalProperties:
        type: object
        additionalProperties:
          type: object
          additionalProperties:
            type: boolean
```

## Lexical Elements

### Identifiers

- Start with a letter (`a-z`, `A-Z`)
- Followed by letters, digits (`0-9`), or underscores (`_`)
- Case-sensitive

**Valid:**
- `User`
- `user_id`
- `Product123`
- `_internal`

**Invalid:**
- `123User` (starts with digit)
- `user-name` (hyphen not allowed)
- `type` (reserved keyword)

### Keywords

Reserved keywords:
- `namespace`
- `import`
- `type`
- `enum`
- `union`
- `service`
- `rpc`
- `returns`
- `map`

### Comments

**Documentation comments:**
```typemux
/// This is documentation
/// that appears in generated code
```

**Regular comments:**
```typemux
// This is a regular comment
// Not included in generated code
```

### String Literals

String literals use double quotes:

```typemux
@default("hello world")
@http.path("/api/v1/users/{id}")
```

**Escape sequences:**
- `\"` - Double quote
- `\\` - Backslash
- `\n` - Newline
- `\t` - Tab

## Best Practices

### Naming Conventions

**Types:** PascalCase
```typemux
type UserAccount { }
type OrderHistory { }
```

**Fields:** camelCase
```typemux
type User {
  firstName: string
  lastName: string
  emailAddress: string
}
```

**Enums:** UPPER_SNAKE_CASE
```typemux
enum UserRole {
  SUPER_ADMIN
  REGULAR_USER
  GUEST_USER
}
```

**Services:** PascalCase with "Service" suffix
```typemux
service UserService { }
service OrderService { }
```

**Methods:** PascalCase with verb prefix
```typemux
rpc GetUser(...)
rpc CreateOrder(...)
rpc UpdateProfile(...)
rpc DeleteAccount(...)
```

### File Organization

**Small projects:**
```
schema.typemux
```

**Medium projects:**
```
types.typemux
enums.typemux
services.typemux
```

**Large projects:**
```
types/
  user.typemux
  order.typemux
  product.typemux
services/
  user_service.typemux
  order_service.typemux
  product_service.typemux
```

### Documentation

- Document all public types and fields
- Explain business logic and constraints
- Include examples in documentation when helpful
- Document error conditions for service methods

### Field Numbering

- Reserve 1-15 for frequently used fields (more efficient in Protobuf)
- Reserve ranges for future use
- Document field number assignments in complex schemas

### Annotations

- Use inline annotations for simple cases
- Use YAML annotations for complex configurations
- Keep annotations consistent across similar types
