# Examples

Real-world examples demonstrating TypeMUX features and use cases.

## Table of Contents

- [Basic Example](#basic-example)
- [Union Types Example](#union-types-example)
- [YAML Annotations Example](#yaml-annotations-example)
- [Namespace Example](#namespace-example)
- [Custom Field Numbers](#custom-field-numbers)
- [Status Codes Example](#status-codes-example)
- [Multi-File Imports](#multi-file-imports)

## Basic Example

A complete user management API with posts, demonstrating core TypeMUX features.

### Schema

**basic.typemux:**
```typemux
/// User role enumeration
/// Defines the different roles a user can have in the system
enum UserRole {
  /// Administrator with full access
  ADMIN = 10
  /// Regular user with limited access
  USER = 20
  /// Guest user with read-only access
  GUEST = 30
}

/// Status enumeration for various entities
enum Status {
  ACTIVE = 1
  INACTIVE = 2
  PENDING = 3
}

/// User entity representing a system user
type User {
  /// Unique identifier for the user
  id: string @required
  /// Full name of the user
  name: string @required
  /// Email address for contact
  email: string @required
  /// User's age in years
  age: int32
  /// Role assigned to the user
  role: UserRole @required
  /// Whether the user account is active
  isActive: bool @default(true)
  /// Timestamp when the user was created
  createdAt: timestamp @required
  /// Custom tags for categorization
  tags: []string
  /// Additional metadata key-value pairs
  metadata: map<string, string>
  /// Internal database version (excluded from GraphQL and OpenAPI)
  dbVersion: int32 @exclude(graphql,openapi)
  /// Password hash (only in Protobuf for internal services)
  passwordHash: string @only(proto)
}

type Post {
  id: string @required
  title: string @required
  content: string
  authorId: string @required
  status: Status @required
  publishedAt: timestamp
  viewCount: int64 @default(0)
  tags: []string
}

type CreateUserRequest {
  name: string @required
  email: string @required
  role: UserRole @required
}

type CreateUserResponse {
  user: User @required
  success: bool @required
}

type GetUserRequest {
  id: string @required
}

type GetUserResponse {
  user: User
}

type ListUsersRequest {
  limit: int32 @default(10)
  offset: int32 @default(0)
  role: UserRole
}

type ListUsersResponse {
  users: []User @required
  total: int32 @required
}

/// User service for managing users
service UserService {
  /// Create a new user
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
    @http.method(POST)
    @http.path("/api/v1/users")
    @graphql(mutation)

  /// Get a user by ID
  rpc GetUser(GetUserRequest) returns (GetUserResponse)
    @http.method(GET)
    @http.path("/api/v1/users/{id}")
    @graphql(query)

  /// List all users with pagination
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse)
    @http.method(GET)
    @http.path("/api/v1/users")
    @graphql(query)

  /// Delete a user
  rpc DeleteUser(GetUserRequest) returns (GetUserResponse)
    @http.method(DELETE)
    @http.path("/api/v1/users/{id}")
    @graphql(mutation)
}

/// Post service for managing blog posts
service PostService {
  /// Create a new post
  rpc CreatePost(Post) returns (Post)
    @http.method(POST)
    @http.path("/api/v1/posts")
    @graphql(mutation)

  /// Get a post by ID
  rpc GetPost(GetUserRequest) returns (Post)
    @http.method(GET)
    @http.path("/api/v1/posts/{id}")
    @graphql(query)
}
```

### Generate

```bash
cd examples/basic
typemux -input basic.typemux -output ./generated
```

### Key Features Demonstrated

- ✅ Enums with custom values
- ✅ Complex types with multiple field types
- ✅ Arrays and maps
- ✅ Field attributes (`@required`, `@default`)
- ✅ Field visibility control (`@exclude`, `@only`)
- ✅ Service definitions with multiple methods
- ✅ HTTP method and path annotations
- ✅ GraphQL operation types
- ✅ Documentation comments

### Generated Files

- `generated/schema.graphql` - GraphQL schema with User, Post types and Query/Mutation types
- `generated/schema.proto` - Protocol Buffers with messages and services
- `generated/openapi.yaml` - OpenAPI 3.0 specification with REST endpoints

## Union Types Example

Demonstrates polymorphic message types using unions.

### Schema

**union_example.typemux:**
```typemux
/// Example demonstrating union/oneOf types

type TextMessage {
    content: string @required
    timestamp: timestamp @required
}

type ImageMessage {
    imageUrl: string @required
    thumbnail: string
    timestamp: timestamp @required
}

type VideoMessage {
    videoUrl: string @required
    duration: int32 @required
    thumbnail: string
    timestamp: timestamp @required
}

/// A message can be text, image, or video
union Message {
    TextMessage
    ImageMessage
    VideoMessage
}

type SendMessageRequest {
    chatId: string @required
    message: Message @required
}

type SendMessageResponse {
    messageId: string @required
    success: bool @required
}

type GetMessageRequest {
    messageId: string @required
}

type GetMessageResponse {
    message: Message @required
}

service MessageService {
    /// Send a message (text, image, or video)
    rpc SendMessage(SendMessageRequest) returns (SendMessageResponse)
        @http.method(POST)
        @http.path("/api/v1/messages")
        @http.success(201)
        @http.errors(400,500)

    /// Get a message by ID
    rpc GetMessage(GetMessageRequest) returns (GetMessageResponse)
        @http.method(GET)
        @http.path("/api/v1/messages/{id}")
        @http.errors(404,500)
}
```

### Generate

```bash
cd examples/unions
typemux -input union_example.typemux -output ./generated
```

### Key Features Demonstrated

- ✅ Union type definitions
- ✅ Polymorphic service method parameters and returns
- ✅ HTTP status code annotations (`@http.success`, `@http.errors`)

### Generated Output

**GraphQL:**
```graphql
union Message @oneOf = TextMessage | ImageMessage | VideoMessage

type Mutation {
  SendMessage(input: SendMessageRequestInput!): SendMessageResponse
}

type Query {
  GetMessage(input: GetMessageRequestInput!): GetMessageResponse
}
```

**Protobuf:**
```protobuf
message Message {
  oneof value {
    TextMessage text_message = 1;
    ImageMessage image_message = 2;
    VideoMessage video_message = 3;
  }
}
```

**OpenAPI:**
```yaml
Message:
  oneOf:
    - $ref: '#/components/schemas/TextMessage'
    - $ref: '#/components/schemas/ImageMessage'
    - $ref: '#/components/schemas/VideoMessage'
```

## YAML Annotations Example

Demonstrates separating schema definition from metadata using external YAML files.

### Schema

**schema.typemux:**
```typemux
/// Example demonstrating YAML annotations
/// Annotations are defined in a separate annotations.yaml file
namespace com.example.api

enum Status {
    ACTIVE = 1
    INACTIVE = 2
    DELETED = 3
}

type User {
    id: string = 1
    username: string = 2
    email: string = 3
    status: Status = 4
    createdAt: timestamp = 5
}

type Product {
    id: string = 1
    name: string = 2
    price: float64 = 3
}

type GetUserRequest {
    userId: string
}

type GetUserResponse {
    user: User
    success: bool
}

type CreateProductRequest {
    name: string
    price: float64
}

type CreateProductResponse {
    product: Product
}

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse)
    rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse)
}
```

### Annotations

**annotations.yaml:**
```yaml
types:
  User:
    # Type-level annotations - custom names per generator
    proto:
      name: "UserV2"
    graphql:
      name: "UserAccount"
    openapi:
      name: "UserProfile"

    # Field-level annotations
    fields:
      username:
        required: true
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'
      status:
        required: true
      createdAt:
        required: true

  Product:
    proto:
      name: "ProductV3"
    fields:
      name:
        required: true
      price:
        required: true
        openapi:
          extension: '{"x-format": "currency"}'

  GetUserRequest:
    fields:
      userId:
        required: true

  GetUserResponse:
    fields:
      user:
        required: true
      success:
        required: true

  CreateProductRequest:
    fields:
      name:
        required: true
      price:
        required: true

  CreateProductResponse:
    fields:
      product:
        required: true

services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{userId}"
        graphql: "query"
        errors: [404, 500]

      CreateProduct:
        http: "POST"
        path: "/api/v1/products"
        graphql: "mutation"
        success: [201]
        errors: [400, 500]
```

### Generate

```bash
cd examples/yaml-annotations
typemux -input schema.typemux -annotations annotations.yaml -output ./generated
```

### Key Features Demonstrated

- ✅ External YAML annotation files
- ✅ Format-specific type name overrides
- ✅ Field requirements defined in YAML
- ✅ HTTP status codes in YAML
- ✅ OpenAPI extensions
- ✅ Separation of schema from metadata

### Benefits

- **Clean schemas**: Keep IDL files focused on structure
- **Configuration management**: Manage metadata separately
- **Multi-environment**: Different annotations for dev/staging/prod
- **Non-technical edits**: Product managers can update docs/metadata

## Namespace Example

Demonstrates organizing types in namespaces to handle naming conflicts.

### Schema

**schema.typemux:**
```typemux
// Multiple namespaces in one file
namespace com.example.users

type User {
    id: string
    username: string
    email: string
}

enum UserStatus {
    ACTIVE
    INACTIVE
    SUSPENDED
}

namespace com.example.products

type User {
    id: string
    productName: string
    ownerId: string
}

enum Status {
    AVAILABLE
    OUT_OF_STOCK
}

service ProductService {
    rpc GetProduct(GetProductRequest) returns (GetProductResponse)
}

type GetProductRequest {
    productId: string
}

type GetProductResponse {
    user: User
    success: bool
}
```

### Annotations with Qualified Names

**annotations.yaml:**
```yaml
types:
  com.example.users.User:
    fields:
      id:
        required: true
      username:
        required: true
      email:
        required: true

  com.example.products.User:
    fields:
      id:
        required: true
      productName:
        required: true

services:
  com.example.products.ProductService:
    methods:
      GetProduct:
        http: "GET"
        path: "/api/v1/products/{productId}"
```

### Generate

```bash
cd examples/namespaces
typemux -input schema.typemux -annotations annotations.yaml -output ./generated
```

### Generated Files

**Protobuf** (separate files per namespace):
- `generated/com.example.users.proto`
- `generated/com.example.products.proto`

**GraphQL** (single file, all types must have unique names):
- `generated/schema.graphql`

### Key Features Demonstrated

- ✅ Multiple namespaces in one file
- ✅ Types with same name in different namespaces
- ✅ Qualified names in YAML annotations
- ✅ Namespace-based proto file generation

## Custom Field Numbers

Control Protobuf field numbering for backward compatibility.

### Schema

```typemux
type Product {
  id: string = 1
  name: string = 2
  // Reserved field numbers 3-9 for future use
  price: float64 = 10
  category: string = 11
  // High numbers for rarely-used fields
  internalMetadata: string = 1000
}

enum Priority {
  LOW = 1
  MEDIUM = 2
  HIGH = 3
  CRITICAL = 10
}
```

### Generate

```bash
cd examples/custom-field-numbers
typemux -input schema.typemux -output ./generated
```

### Generated Protobuf

```protobuf
message Product {
  string id = 1;
  string name = 2;
  double price = 10;
  string category = 11;
  string internalMetadata = 1000;
}

enum Priority {
  LOW = 1;
  MEDIUM = 2;
  HIGH = 3;
  CRITICAL = 10;
}
```

### Key Features Demonstrated

- ✅ Explicit field numbers
- ✅ Non-sequential numbering
- ✅ Reserved number ranges
- ✅ Enum value assignments

### Use Cases

- Maintaining backward compatibility when evolving schemas
- Reserving field number ranges for future fields
- Optimizing wire format (1-15 are more efficient in Protobuf)
- Aligning with existing Protobuf schemas

## Status Codes Example

Comprehensive HTTP status code handling.

### Schema

```typemux
service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (Order)
    @http.method(POST)
    @http.path("/api/v1/orders")
    @http.success(201)
    @http.errors(400,409,500)

  rpc GetOrder(GetOrderRequest) returns (Order)
    @http.method(GET)
    @http.path("/api/v1/orders/{id}")
    @http.success(200)
    @http.errors(404,500)

  rpc UpdateOrder(UpdateOrderRequest) returns (Order)
    @http.method(PUT)
    @http.path("/api/v1/orders/{id}")
    @http.success(200,204)
    @http.errors(400,404,409,500)

  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse)
    @http.method(DELETE)
    @http.path("/api/v1/orders/{id}")
    @http.success(204)
    @http.errors(404,500)

  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse)
    @http.method(GET)
    @http.path("/api/v1/orders")
    @http.success(200)
    @http.errors(500)
}
```

### Generated OpenAPI

```yaml
paths:
  /api/v1/orders:
    post:
      operationId: CreateOrder
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Order'
        '400':
          description: Bad Request
        '409':
          description: Conflict
        '500':
          description: Internal Server Error
```

### Key Features Demonstrated

- ✅ Multiple success codes
- ✅ Multiple error codes
- ✅ RESTful status code conventions
- ✅ OpenAPI response generation

## Multi-File Imports

Organize large schemas across multiple files.

### File Structure

```
examples/imports/
├── types/
│   ├── user.typemux
│   ├── product.typemux
│   └── order.typemux
├── services/
│   ├── user_service.typemux
│   ├── product_service.typemux
│   └── order_service.typemux
└── main.typemux
```

### Schema Files

**types/user.typemux:**
```typemux
type User {
  id: string @required
  email: string @required
  name: string @required
}

type CreateUserRequest {
  email: string @required
  name: string @required
}
```

**types/product.typemux:**
```typemux
type Product {
  id: string @required
  name: string @required
  price: float64 @required
}

type CreateProductRequest {
  name: string @required
  price: float64 @required
}
```

**types/order.typemux:**
```typemux
import "types/user.typemux"
import "types/product.typemux"

type Order {
  id: string @required
  user: User @required
  products: []Product @required
  total: float64 @required
}
```

**services/user_service.typemux:**
```typemux
import "types/user.typemux"

service UserService {
  rpc CreateUser(CreateUserRequest) returns (User)
    @http.method(POST)
    @http.path("/api/v1/users")
}
```

**main.typemux:**
```typemux
import "types/user.typemux"
import "types/product.typemux"
import "types/order.typemux"
import "services/user_service.typemux"
```

### Generate

```bash
cd examples/imports
typemux -input main.typemux -output ./generated
```

### Key Features Demonstrated

- ✅ Multiple file organization
- ✅ Import statements
- ✅ Cross-file type references
- ✅ Modular schema design

### Best Practices

- Organize by domain (users, products, orders)
- Separate types from services
- Use a main file to aggregate imports
- Keep related types together
- Avoid circular dependencies

## Running All Examples

Generate all example schemas:

```bash
# From the repository root
make examples

# Or individually
cd examples/basic && typemux -input basic.typemux -output ./generated
cd examples/unions && typemux -input union_example.typemux -output ./generated
cd examples/yaml-annotations && typemux -input schema.typemux -annotations annotations.yaml -output ./generated
cd examples/namespaces && typemux -input schema.typemux -annotations annotations.yaml -output ./generated
```

## Next Steps

- Read the [Language Reference](reference.md) for complete syntax details
- Learn about [Configuration Options](configuration.md)
- Check the [Tutorial](tutorial.md) for step-by-step learning
