# Tutorial

This tutorial will teach you TypeMUX from the ground up, building a complete e-commerce API schema.

## Table of Contents

1. [Basic Types](#basic-types)
2. [Enums](#enums)
3. [Complex Types](#complex-types)
4. [Arrays and Maps](#arrays-and-maps)
5. [Services and Methods](#services-and-methods)
6. [Annotations and Attributes](#annotations-and-attributes)
7. [Union Types](#union-types)
8. [Namespaces](#namespaces)
9. [Multiple Files and Imports](#multiple-files-and-imports)
10. [External YAML Annotations](#external-yaml-annotations)

## Basic Types

TypeMUX supports these primitive types:

- `string` - Text data
- `int32` - 32-bit integer
- `int64` - 64-bit integer
- `float32` - 32-bit floating point
- `float64` - 64-bit floating point
- `bool` - Boolean (true/false)
- `timestamp` - Date and time
- `bytes` - Binary data

Let's create a simple product type:

```typemux
type Product {
  id: string
  name: string
  price: float64
  inStock: bool
}
```

## Enums

Enums define a set of named constants:

```typemux
enum ProductCategory {
  ELECTRONICS
  CLOTHING
  BOOKS
  FOOD
  OTHER
}
```

Add the category to the Product:

```typemux
type Product {
  id: string
  name: string
  price: float64
  inStock: bool
  category: ProductCategory
}
```

## Complex Types

Types can reference other types. Let's add a supplier:

```typemux
type Supplier {
  id: string
  name: string
  email: string
  phone: string
}

type Product {
  id: string
  name: string
  price: float64
  inStock: bool
  category: ProductCategory
  supplier: Supplier
}
```

## Arrays and Maps

### Arrays

Use `[]` to define arrays:

```typemux
type Product {
  id: string
  name: string
  price: float64
  inStock: bool
  category: ProductCategory
  supplier: Supplier
  tags: []string
  images: []string
}
```

### Maps

Use `map<KeyType, ValueType>` for key-value pairs:

```typemux
type Product {
  id: string
  name: string
  price: float64
  inStock: bool
  category: ProductCategory
  supplier: Supplier
  tags: []string
  images: []string
  attributes: map<string, string>
}
```

Maps are converted differently by each output format:

- **GraphQL:** Creates strongly-typed KeyValue entry types (e.g., `[StringStringEntry!]`)
- **Protobuf:** Uses native map syntax (e.g., `map<string, string>`)
- **OpenAPI:** Uses `additionalProperties` with proper typing

**Nested maps are fully supported:**

```typemux
type Product {
  // Simple map
  attributes: map<string, string>

  // Nested map - directly supported
  nested_metadata: map<string, map<string, string>>

  // Triple nested map - also supported
  deep_config: map<string, map<string, map<string, int32>>>
}
```

Each generator handles nested maps appropriately:
- **GraphQL**: Auto-generates wrapper types (MapWrapper0, MapWrapper1, etc.)
- **Protobuf**: Uses native nested `map<string, map<string, int32>>` syntax
- **OpenAPI**: Uses nested `additionalProperties` structures

## Services and Methods

Services define RPC-style API methods. Each method takes an input type and returns an output type.

```typemux
type GetProductRequest {
  id: string
}

type CreateProductRequest {
  name: string
  price: float64
  category: ProductCategory
}

type ListProductsRequest {
  category: ProductCategory
  limit: int32
  offset: int32
}

type ListProductsResponse {
  products: []Product
  total: int32
}

service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product)
  rpc CreateProduct(CreateProductRequest) returns (Product)
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse)
}
```

## Field Arguments (GraphQL-Style Queries)

Field arguments allow you to add parameters directly to fields, similar to GraphQL field arguments. This is an alternative to defining separate service methods and request/response types.

### Basic Field Arguments

```typemux
type Query {
  // Simple field with required argument
  user(id: string @required): User

  // Field with optional arguments and defaults
  users(limit: int32 @default(10), offset: int32 @default(0)): []User

  // Field with multiple arguments
  posts(
    authorId: string,
    published: bool @default(true),
    limit: int32 @default(10)
  ): []Post
}
```

### How Field Arguments Map to Different Formats

Field arguments are automatically converted to the appropriate pattern for each output format:

#### GraphQL
Fields with arguments generate native GraphQL field arguments:

```graphql
type Query {
  user(id: String!): User
  users(limit: Int = 10, offset: Int = 0): [User]
  posts(authorId: String, published: Boolean = true, limit: Int = 10): [Post]
}
```

#### OpenAPI/REST
Fields with arguments become separate REST endpoints:

```yaml
paths:
  /user:
    get:
      parameters:
        - name: id
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

  /users:
    get:
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 10
        - name: offset
          in: query
          schema:
            type: integer
            default: 0
```

#### Protobuf/gRPC
Fields with arguments become gRPC service methods with auto-generated request/response messages:

```protobuf
message QueryUserRequest {
  string id = 1;
}

message QueryUsersRequest {
  optional int32 limit = 1;
  optional int32 offset = 2;
}

message QueryUsersResponse {
  repeated User items = 1;
}

service QueryService {
  rpc User(QueryUserRequest) returns (User);
  rpc Users(QueryUsersRequest) returns (QueryUsersResponse);
  rpc Posts(QueryPostsRequest) returns (QueryPostsResponse);
}
```

### Argument Annotations

Field arguments support the same annotations as regular fields:

```typemux
type Query {
  // Required argument
  user(id: string @required): User

  // Default values
  users(limit: int32 @default(10)): []User

  // Validation constraints
  searchUsers(
    query: string @required @validate(minLength=3, maxLength=100),
    limit: int32 @default(20) @validate(min=1, max=100)
  ): []User
}
```

### When to Use Field Arguments vs Services

**Use Field Arguments when:**
- Building GraphQL-style query APIs
- You want a more concise, field-oriented syntax
- Arguments are simple query parameters
- You're primarily targeting GraphQL or REST

**Use Services when:**
- Building traditional RPC-style APIs
- You need complex request/response structures
- You want explicit control over request/response types
- You need streaming (Protobuf server/client streaming)

Both patterns can coexist in the same schema:

```typemux
// GraphQL-style queries using field arguments
type Query {
  user(id: string @required): User
  users(limit: int32 @default(10)): []User
}

// Traditional RPC-style mutations using services
type CreateUserRequest {
  name: string @required
  email: string @required
}

service UserService {
  rpc CreateUser(CreateUserRequest) returns (User)
    @http.method(POST)
    @http.path("/api/v1/users")
    @graphql(mutation)
}
```

## Annotations and Attributes

Annotations add metadata to control code generation.

### Field Attributes

#### @required

Mark fields as required (non-nullable):

```typemux
type Product {
  id: string @required
  name: string @required
  price: float64 @required
  inStock: bool @required
  category: ProductCategory
  supplier: Supplier
  tags: []string
  images: []string
  attributes: map<string, string>
}
```

#### @default

Provide default values:

```typemux
type Product {
  id: string @required
  name: string @required
  price: float64 @required
  inStock: bool @default("true")
  category: ProductCategory @default("OTHER")
  supplier: Supplier
  tags: []string
  images: []string
  attributes: map<string, string>
}
```

#### @exclude

Exclude fields from specific formats:

```typemux
type Product {
  id: string @required
  name: string @required
  price: float64 @required
  inStock: bool @default("true")
  category: ProductCategory @default("OTHER")
  supplier: Supplier
  tags: []string
  images: []string
  attributes: map<string, string>
  internalNotes: string @exclude(graphql,openapi)
}
```

The `internalNotes` field will only appear in Protobuf.

#### @only

Include fields only in specific formats:

```typemux
type Product {
  id: string @required
  name: string @required
  price: float64 @required
  inStock: bool @default("true")
  category: ProductCategory @default("OTHER")
  supplier: Supplier
  tags: []string
  images: []string
  attributes: map<string, string>
  internalNotes: string @exclude(graphql,openapi)
  graphqlMetadata: string @only(graphql)
}
```

### Method Annotations

#### @http.method

Specify HTTP method:

```typemux
service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product)
    @http.method(GET)

  rpc CreateProduct(CreateProductRequest) returns (Product)
    @http.method(POST)

  rpc UpdateProduct(UpdateProductRequest) returns (Product)
    @http.method(PUT)

  rpc DeleteProduct(DeleteProductRequest) returns (DeleteProductResponse)
    @http.method(DELETE)
}
```

#### @http.path

Define URL path templates:

```typemux
service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product)
    @http.method(GET)
    @http.path("/api/v1/products/{id}")

  rpc CreateProduct(CreateProductRequest) returns (Product)
    @http.method(POST)
    @http.path("/api/v1/products")

  rpc UpdateProduct(UpdateProductRequest) returns (Product)
    @http.method(PUT)
    @http.path("/api/v1/products/{id}")

  rpc DeleteProduct(DeleteProductRequest) returns (DeleteProductResponse)
    @http.method(DELETE)
    @http.path("/api/v1/products/{id}")
}
```

#### @graphql

Specify GraphQL operation type:

```typemux
service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product)
    @http.method(GET)
    @http.path("/api/v1/products/{id}")
    @graphql(query)

  rpc CreateProduct(CreateProductRequest) returns (Product)
    @http.method(POST)
    @http.path("/api/v1/products")
    @graphql(mutation)

  rpc UpdateProduct(UpdateProductRequest) returns (Product)
    @http.method(PUT)
    @http.path("/api/v1/products/{id}")
    @graphql(mutation)

  rpc DeleteProduct(DeleteProductRequest) returns (DeleteProductResponse)
    @http.method(DELETE)
    @http.path("/api/v1/products/{id}")
    @graphql(mutation)
}
```

#### @http.success and @http.errors

Define HTTP status codes:

```typemux
service ProductService {
  rpc CreateProduct(CreateProductRequest) returns (Product)
    @http.method(POST)
    @http.path("/api/v1/products")
    @graphql(mutation)
    @http.success(201)
    @http.errors(400,409,500)

  rpc GetProduct(GetProductRequest) returns (Product)
    @http.method(GET)
    @http.path("/api/v1/products/{id}")
    @graphql(query)
    @http.success(200)
    @http.errors(404,500)
}
```

## Union Types

Unions represent a value that can be one of several types. This is useful for polymorphic responses.

```typemux
type TextMessage {
  text: string @required
  timestamp: timestamp @required
}

type ImageMessage {
  imageUrl: string @required
  caption: string
  timestamp: timestamp @required
}

type VideoMessage {
  videoUrl: string @required
  thumbnail: string
  duration: int32
  timestamp: timestamp @required
}

union Message {
  TextMessage
  ImageMessage
  VideoMessage
}

type GetMessageRequest {
  id: string @required
}

service MessageService {
  rpc GetMessage(GetMessageRequest) returns (Message)
    @http.method(GET)
    @http.path("/api/v1/messages/{id}")
    @graphql(query)
}
```

**Generated GraphQL:**

```graphql
union Message @oneOf = TextMessage | ImageMessage | VideoMessage
```

**Generated Protobuf:**

```protobuf
message Message {
  oneof value {
    TextMessage text_message = 1;
    ImageMessage image_message = 2;
    VideoMessage video_message = 3;
  }
}
```

**Generated OpenAPI:**

```yaml
Message:
  oneOf:
    - $ref: '#/components/schemas/TextMessage'
    - $ref: '#/components/schemas/ImageMessage'
    - $ref: '#/components/schemas/VideoMessage'
```

## Namespaces

Namespaces organize types and prevent naming conflicts. Use reverse domain notation:

```typemux
namespace com.example.ecommerce

type Product {
  id: string @required
  name: string @required
}
```

### Multiple Types with Same Name

Different namespaces can have types with the same name:

**users.typemux:**
```typemux
namespace com.example.users

type User {
  id: string @required
  email: string @required
}

type Address {
  street: string @required
  city: string @required
}
```

**orders.typemux:**
```typemux
namespace com.example.orders

type Order {
  id: string @required
  userId: string @required
}

type Address {
  latitude: float64 @required
  longitude: float64 @required
}
```

Both namespaces have an `Address` type, but they're distinct.

### Protobuf Generation with Namespaces

When using namespaces, Protobuf generates separate files per namespace:

```bash
./typemux -input schema.typemux -format protobuf -output ./generated
```

Creates:
- `generated/com.example.users.proto`
- `generated/com.example.orders.proto`

## Multiple Files and Imports

Split large schemas into multiple files using imports.

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

type GetProductRequest {
  id: string @required
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User)
    @http.method(GET)
    @http.path("/users/{id}")
}

service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product)
    @http.method(GET)
    @http.path("/products/{id}")
}
```

### Cross-Namespace Imports

You can import types from different namespaces:

**com/example/users/types.typemux:**
```typemux
namespace com.example.users

type User {
  id: string @required
  email: string @required
}
```

**com/example/orders/types.typemux:**
```typemux
namespace com.example.orders

import "com/example/users/types.typemux"

type Order {
  id: string @required
  userId: string @required
  user: User
}
```

## External YAML Annotations

For large projects, you can separate annotations from your schema definition using YAML files.

**schema.typemux:**
```typemux
type Product {
  id: string
  name: string
  price: float64
  inStock: bool
}

service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product)
  rpc CreateProduct(CreateProductRequest) returns (Product)
}
```

**annotations.yaml:**
```yaml
types:
  Product:
    fields:
      id:
        required: true
      name:
        required: true
      price:
        required: true
      inStock:
        required: true
        default: true

services:
  ProductService:
    methods:
      GetProduct:
        http: GET
        path: /api/v1/products/{id}
        graphql: query
        success: [200]
        errors: [404, 500]
      CreateProduct:
        http: POST
        path: /api/v1/products
        graphql: mutation
        success: [201]
        errors: [400, 500]
```

Generate with annotations:

```bash
./typemux -input schema.typemux -annotations annotations.yaml -output ./generated
```

### Multiple Annotation Files

You can merge multiple annotation files. Later files override earlier ones:

```bash
./typemux -input schema.typemux \
          -annotations base-annotations.yaml \
          -annotations overrides.yaml \
          -output ./generated
```

### Format-Specific Names

Override type and field names per format:

**annotations.yaml:**
```yaml
types:
  Product:
    proto:
      name: ProductMessage
    graphql:
      name: ProductType
    openapi:
      name: ProductSchema
    fields:
      id:
        proto:
          name: product_id
        graphql:
          name: productId
```

### Qualified Names in Namespaces

When using namespaces, use fully qualified names in YAML:

**annotations.yaml:**
```yaml
types:
  com.example.users.User:
    fields:
      id:
        required: true

  com.example.orders.Order:
    fields:
      id:
        required: true
      userId:
        required: true
```

## Documentation Comments

Use triple-slash comments (`///`) to add documentation:

```typemux
/// User account with authentication details
type User {
  /// Unique user identifier
  id: string @required

  /// User's email address for login
  email: string @required

  /// Display name
  username: string @required
}

/// Service for managing user accounts
service UserService {
  /// Retrieves a user by their unique ID
  rpc GetUser(GetUserRequest) returns (User)
    @http.method(GET)
    @http.path("/api/v1/users/{id}")
    @graphql(query)
}
```

Documentation is included in all generated formats:

**GraphQL:**
```graphql
"""
User account with authentication details
"""
type User {
  """
  Unique user identifier
  """
  id: String!
}
```

**Protobuf:**
```protobuf
// User account with authentication details
message User {
  // Unique user identifier
  string id = 1;
}
```

**OpenAPI:**
```yaml
User:
  type: object
  description: User account with authentication details
  properties:
    id:
      type: string
      description: Unique user identifier
```

## Custom Field Numbers

Control Protobuf field numbering explicitly:

```typemux
type Product {
  id: string = 1
  name: string = 2
  price: float64 = 10
  category: ProductCategory = 100
}

enum ProductCategory {
  ELECTRONICS = 1
  CLOTHING = 2
  BOOKS = 3
}
```

This is useful for:
- Maintaining backward compatibility
- Reserving field number ranges
- Optimizing wire format size

## Complete Example

Here's a complete e-commerce schema combining all concepts:

**ecommerce.typemux:**
```typemux
namespace com.example.ecommerce

/// Product category enumeration
enum ProductCategory {
  ELECTRONICS = 1
  CLOTHING = 2
  BOOKS = 3
  FOOD = 4
  OTHER = 5
}

/// Product availability status
enum ProductStatus {
  AVAILABLE = 1
  OUT_OF_STOCK = 2
  DISCONTINUED = 3
}

/// Product supplier information
type Supplier {
  /// Unique supplier identifier
  id: string @required

  /// Supplier company name
  name: string @required

  /// Contact email
  email: string @required

  /// Contact phone number
  phone: string
}

/// Product in the catalog
type Product {
  /// Unique product identifier
  id: string @required

  /// Product name
  name: string @required

  /// Product description
  description: string

  /// Price in USD
  price: float64 @required

  /// Product category
  category: ProductCategory @required

  /// Current availability status
  status: ProductStatus @default("AVAILABLE")

  /// Product supplier
  supplier: Supplier

  /// Search tags
  tags: []string

  /// Product image URLs
  images: []string

  /// Custom attributes
  attributes: map<string, string>

  /// Creation timestamp
  createdAt: timestamp @required

  /// Last update timestamp
  updatedAt: timestamp @required
}

/// Request to retrieve a product by ID
type GetProductRequest {
  /// Product identifier
  id: string @required
}

/// Request to create a new product
type CreateProductRequest {
  name: string @required
  description: string
  price: float64 @required
  category: ProductCategory @required
  supplierId: string @required
  tags: []string
  images: []string
  attributes: map<string, string>
}

/// Request to list products with filtering
type ListProductsRequest {
  /// Filter by category
  category: ProductCategory

  /// Filter by status
  status: ProductStatus

  /// Maximum number of results
  limit: int32 @default("50")

  /// Offset for pagination
  offset: int32 @default("0")
}

/// Response containing product list
type ListProductsResponse {
  /// List of products
  products: []Product @required

  /// Total count of products matching filter
  total: int32 @required
}

/// Product catalog service
service ProductService {
  /// Get a product by its ID
  rpc GetProduct(GetProductRequest) returns (Product)
    @http.method(GET)
    @http.path("/api/v1/products/{id}")
    @graphql(query)
    @http.success(200)
    @http.errors(404,500)

  /// Create a new product
  rpc CreateProduct(CreateProductRequest) returns (Product)
    @http.method(POST)
    @http.path("/api/v1/products")
    @graphql(mutation)
    @http.success(201)
    @http.errors(400,500)

  /// List products with optional filtering
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse)
    @http.method(GET)
    @http.path("/api/v1/products")
    @graphql(query)
    @http.success(200)
    @http.errors(500)
}
```

Generate all formats:

```bash
./typemux -input ecommerce.typemux -output ./generated
```

## Next Steps

- Explore more [Examples](examples.md)
- Read the [Language Reference](reference.md) for complete syntax
- Check the [Configuration Guide](configuration.md) for advanced options
