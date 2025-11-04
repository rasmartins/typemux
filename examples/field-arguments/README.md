# Field Arguments Example

This example demonstrates the new **field-level parameterized queries** feature in TypeMUX, similar to GraphQL field arguments.

## Feature Overview

Field arguments allow you to add parameters directly to fields in your type definitions, enabling more flexible and powerful query patterns without having to create separate request/response types for every query.

## Syntax

```typemux
type Query {
  // Simple required argument
  user(id: string @required): User

  // Multiple arguments with defaults
  users(
    limit: int32 @default(10),
    offset: int32 @default(0)
  ): []User

  // Arguments with validation
  searchUsers(
    query: string @required @validate(minLength=3, maxLength=100),
    limit: int32 @default(20) @validate(min=1, max=100)
  ): []User

  // Optional arguments
  findUser(
    username: string?,
    email: string?
  ): User?

  // Mix of required, optional, and default arguments
  posts(
    authorId: string?,
    published: bool @default(true),
    limit: int32 @default(10),
    offset: int32 @default(0),
    sortBy: string @default("createdAt")
  ): []Post
}
```

## Features Demonstrated

### 1. Simple Arguments
```typemux
user(id: string @required): User
```
**Generates (GraphQL):**
```graphql
user(id: String!): User
```

### 2. Multiple Arguments with Defaults
```typemux
users(limit: int32 @default(10), offset: int32 @default(0)): []User
```
**Generates (GraphQL):**
```graphql
users(limit: Int = 10, offset: Int = 0): [User]
```

### 3. Arguments with Validation
```typemux
searchUsers(
  query: string @required @validate(minLength=3),
  limit: int32 @default(20) @validate(min=1, max=100)
): []User
```
**Generates (GraphQL):**
```graphql
searchUsers(query: String!, limit: Int = 20): [User]
```

### 4. Fields with and without Arguments
```typemux
type UserProfile {
  user: User @required
  posts(limit: int32 @default(5)): []Post
  followerCount: int32
}
```
**Generates (GraphQL):**
```graphql
type UserProfile {
  user: User!
  posts(limit: Int = 5): [Post]
  followerCount: Int
}
```

## Supported Annotations on Arguments

- `@required` - Mark argument as required (adds `!` in GraphQL)
- `@default(value)` - Set default value for argument
- `@validate(...)` - Add validation constraints (format, min, max, etc.)
- `@graphql.name("customName")` - Override argument name in GraphQL output
- `@proto.name("customName")` - Override argument name in Protobuf output
- `@openapi.name("customName")` - Override argument name in OpenAPI output

## Argument Types

All TypeMUX types are supported for arguments:

- **Primitives**: `string`, `int32`, `int64`, `uint32`, `float32`, `bool`, `timestamp`, `bytes`
- **Arrays**: `[]string`, `[]User`
- **Optional**: `string?`, `User?`
- **Custom types**: Any defined type can be used as an argument type

## Compiling the Example

```bash
# Generate GraphQL schema
./bin/typemux -input examples/field-arguments/field-arguments.typemux \
  -format graphql \
  -output examples/field-arguments/generated

# Generate all formats
./bin/typemux -input examples/field-arguments/field-arguments.typemux \
  -format all \
  -output examples/field-arguments/generated
```

## Generated Output

The example generates:
- **GraphQL**: Field arguments are natively supported and generated correctly
- **Protobuf**: Fields with arguments are preserved (though gRPC methods typically use request messages)
- **OpenAPI**: Can be mapped to query/path parameters (future enhancement)
- **Go**: Generates appropriate function signatures (future enhancement)

## Benefits Over Request/Response Pattern

### Before (Request/Response Pattern)
```typemux
type GetUserRequest {
  id: string @required
}

type GetUserResponse {
  user: User
}

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse)
}
```

### After (Field Arguments)
```typemux
type Query {
  user(id: string @required): User
}
```

**Advantages:**
- ✅ More concise and readable
- ✅ Matches GraphQL conventions naturally
- ✅ Reduces type proliferation
- ✅ Better for simple queries
- ✅ Easier to maintain

**Note:** Request/Response pattern is still recommended for:
- Complex mutations with many fields
- Operations requiring extensive validation
- gRPC services with specific protobuf requirements

## Files

- `field-arguments.typemux` - Comprehensive example with various field argument patterns
- `simple.typemux` - Minimal example for quick testing
- `generated/` - Generated GraphQL, Protobuf, OpenAPI, and Go code

## Comparison with GraphQL

TypeMUX field arguments are designed to match GraphQL field arguments closely:

| TypeMUX | GraphQL |
|---------|---------|
| `user(id: string @required): User` | `user(id: String!): User` |
| `users(limit: int32 @default(10)): []User` | `users(limit: Int = 10): [User]` |
| `search(query: string @required, filter: Filter?): []Result` | `search(query: String!, filter: Filter): [Result]` |

This makes it easy to:
1. Import existing GraphQL schemas
2. Generate GraphQL APIs from TypeMUX
3. Transition between TypeMUX and GraphQL
