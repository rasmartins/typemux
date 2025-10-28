# GraphQL Import Example

This example demonstrates how to import existing GraphQL schema files and convert them to TypeMUX IDL format.

## Files

- `example.graphqls` - Original GraphQL schema
- `example.typemux` - Converted TypeMUX IDL (generated)
- `generated/schema.graphql` - Round-trip generated GraphQL (for verification)

## Converting GraphQL to TypeMUX

Use the `graphql2typemux` tool to convert GraphQL schema files:

```bash
graphql2typemux --input example.graphqls --output ./
```

Or from the repository root:

```bash
go run cmd/graphql2typemux/main.go --input examples/graphql-import/example.graphqls --output examples/graphql-import
```

## Features Demonstrated

### 1. Enums
```graphql
enum UserStatus {
  """
  User account is active
  """
  ACTIVE
  INACTIVE
  SUSPENDED
}
```

Converts to:
```typemux
enum UserStatus {
  // User account is active
  ACTIVE = 0
  INACTIVE = 1
  SUSPENDED = 2
}
```

### 2. Object Types
```graphql
type User {
  id: ID!
  name: String!
  email: String!
  age: Int
  status: UserStatus!
  tags: [String!]!
}
```

Converts to:
```typemux
type User {
  id: string = 1
  name: string = 2
  email: string = 3
  age: int32 = 4
  status: UserStatus = 5
  tags: []string = 6
}
```

### 3. Input Types
```graphql
input CreateUserInput {
  name: String!
  email: String!
  age: Int
}
```

Converts to:
```typemux
type CreateUserInput {
  name: string = 1
  email: string = 2
  age: int32 = 3
}
```

### 4. Queries, Mutations, and Subscriptions
```graphql
extend type Query {
  getUser(id: ID!): User!

  listUsers(
    limit: Int = 10
    offset: Int = 0
    status: UserStatus
  ): [User!]!
}

extend type Mutation {
  createUser(input: CreateUserInput!): CreateUserResult!
  deleteUser(id: ID!): Boolean!
}

extend type Subscription {
  """
  Subscribe to user updates
  """
  userUpdates: UserUpdate!

  """
  Subscribe to new messages
  """
  newMessages(roomId: String!): ChatMessage!
}
```

Converts to:
```typemux
service GraphQLService {
  // GraphQL query
  rpc GetUser(GetUserRequest) returns (User)
  // GraphQL query
  rpc ListUsers(ListUsersRequest) returns (User)
  // GraphQL mutation
  rpc CreateUser(CreateUserRequest) returns (CreateUserResult)
  // GraphQL mutation
  rpc DeleteUser(DeleteUserRequest) returns (Boolean)
  // Subscribe to user updates
  // GraphQL subscription
  rpc UserUpdates(Empty) returns (stream UserUpdate)
  // Subscribe to new messages
  // GraphQL subscription
  rpc NewMessages(NewMessagesRequest) returns (stream ChatMessage)
}
```

**Note:** GraphQL subscriptions are converted to **server-side streaming RPCs** using the `stream` keyword in the return type.

### 5. Custom Scalars
```graphql
scalar Time
```

Custom scalars like `Time`, `DateTime`, and `Timestamp` are automatically mapped to TypeMUX's `timestamp` type. Other scalars are preserved as type names.

### 6. Multi-line Descriptions
```graphql
"""
User represents a user account in the system.
This includes all user profile information.
"""
type User {
  # ...
}
```

Multi-line descriptions in GraphQL are converted to single-line or multi-line comments in TypeMUX:
```typemux
// User represents a user account in the system.
// This includes all user profile information.
type User {
  # ...
}
```

### 7. Field Arguments with Descriptions
```graphql
mutation {
  updateUser(
    """
    ID of the user to update
    """
    id: ID!
    """
    Fields to update
    """
    input: UpdateUserInput!
  ): User!
}
```

The converter handles multi-line field declarations with argument descriptions correctly.

## Round-Trip Conversion

After converting to TypeMUX, you can generate back to GraphQL:

```bash
# Convert TypeMUX back to GraphQL
typemux --input example.typemux --output ./generated --format graphql

# The generated schema.graphql will be functionally equivalent to the original
```

## Type Mappings

### GraphQL → TypeMUX

| GraphQL Type | TypeMUX Type | Notes |
|--------------|--------------|-------|
| `String` | `string` | |
| `Int` | `int32` | |
| `Float` | `float` | |
| `Boolean` | `bool` | |
| `ID` | `string` | ID is treated as string |
| `Time` / `DateTime` / `Timestamp` | `timestamp` | Custom scalars |
| `[Type]` | `[]Type` | Lists/Arrays |
| `Type!` | `Type` | Non-null (handled by proto3 semantics) |
| `[Type!]!` | `[]Type` | Non-null list of non-null items |

## What's Preserved

✅ Type and field names
✅ Enum values
✅ Field descriptions/documentation
✅ Input types
✅ Queries, mutations, and subscriptions (as service RPCs)
✅ Subscriptions as streaming RPCs (`stream` keyword)
✅ List types (arrays)
✅ Custom scalar types
✅ Multi-line field declarations with arguments
✅ Default values (as GraphQL-specific annotations)

## What's Not Preserved

❌ Non-null modifiers (`!`) - proto3 fields are implicitly optional
❌ Field arguments become separate Request types
❌ Query/Mutation distinction becomes comment-only
❌ Interfaces and Unions (converted to regular types)
❌ Directives (except documented in comments)
❌ Field-level directives

## Reserved Keywords

TypeMUX has reserved keywords that cannot be used as field names. If your GraphQL schema uses these as field names, they will be automatically renamed with a trailing underscore:

- `namespace` → `namespace_`
- `import` → `import_`
- `enum` → `enum_`
- `type` → `type_`
- `union` → `union_`
- `service` → `service_`
- `rpc` → `rpc_`
- `returns` → `returns_`
- `stream` → `stream_`

For example, if your GraphQL schema has:
```graphql
type FilterConfig {
  type: FilterType!
}
```

It will be converted to:
```typemux
type FilterConfig {
  type_: FilterType = 1
}
```

## Use Cases

1. **Schema Migration**: Convert existing GraphQL schemas to TypeMUX IDL for multi-format code generation
2. **API Unification**: Combine GraphQL APIs with Protobuf/OpenAPI using a single schema source
3. **Multi-Format Generation**: Import GraphQL, then generate Protobuf and OpenAPI from the same source
4. **Documentation**: Generate comprehensive API documentation across multiple formats
5. **Polyglot Services**: Use TypeMUX as a lingua franca between GraphQL, gRPC, and REST services

## Known Limitations

1. **Service Methods**: GraphQL queries and mutations are converted to service RPC methods, but the original method/field structure may differ slightly in the round-trip
2. **Request Types**: Arguments to queries/mutations are referenced as `MethodNameRequest` types, but these types are not auto-generated from arguments yet
3. **Non-null Semantics**: GraphQL's explicit non-null (`!`) is lost since proto3 treats all fields as optional by default
4. **Interfaces and Unions**: Advanced GraphQL features like interfaces and unions are converted to regular types

## Testing with Real-World Schemas

This importer has been tested with production GraphQL schemas from Anchorage's statement service, including:
- Complex nested types with descriptions
- Multi-line field declarations
- Multiple enum values
- Default parameter values
- Mix of queries and mutations

Example:
```bash
graphql2typemux --input ~/path/to/production/schema.graphqls --output ./imported
```
