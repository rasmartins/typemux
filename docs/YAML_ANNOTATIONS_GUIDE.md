# YAML Annotations Guide

A comprehensive guide to using YAML annotations in TypeMUX for managing API metadata separately from schema definitions.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Tutorial](#tutorial)
  - [Basic Annotations](#basic-annotations)
  - [Field Annotations](#field-annotations)
  - [Service Annotations](#service-annotations)
  - [Working with Namespaces](#working-with-namespaces)
  - [Multiple Annotation Files](#multiple-annotation-files)
- [Reference](#reference)
  - [YAML Structure](#yaml-structure)
  - [Type Annotations](#type-annotations)
  - [Field Annotations](#field-annotations-reference)
  - [Enum Annotations](#enum-annotations)
  - [Union Annotations](#union-annotations)
  - [Service Annotations](#service-annotations-reference)
  - [Method Annotations](#method-annotations)
- [Best Practices](#best-practices)
- [Validation](#validation)
- [Examples](#examples)

---

## Overview

YAML annotations allow you to define TypeMUX metadata in external YAML files instead of inline in your `.typemux` schema files. This provides:

- **Separation of Concerns** - Keep schema definitions clean and focused
- **Easier Maintenance** - Update annotations without modifying schema files
- **Bulk Management** - Apply annotations across multiple types easily
- **External Configuration** - Manage annotations separately from code
- **Better Organization** - Group related annotations by format (proto/graphql/openapi)

## Quick Start

### 1. Create Your Schema

**schema.typemux:**
```typemux
namespace com.example.api

type User {
    id: string
    username: string
    email: string
    createdAt: timestamp
}

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse)
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
}
```

### 2. Create YAML Annotations

**annotations.yaml:**
```yaml
types:
  User:
    proto:
      name: "UserV2"
    graphql:
      name: "UserAccount"
    fields:
      username:
        required: true
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'

services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{id}"
        graphql: "query"
      CreateUser:
        http: "POST"
        path: "/api/v1/users"
        graphql: "mutation"
```

### 3. Generate Code

```bash
typemux -input schema.typemux -annotations annotations.yaml -output ./generated
```

That's it! Your generated code will use:
- Protobuf: `UserV2` message name
- GraphQL: `UserAccount` type name
- OpenAPI: Email field with format validation

---

## Tutorial

### Basic Annotations

Let's start with a simple example of renaming types for different output formats.

**Step 1: Define your schema**

```typemux
type Product {
    id: string
    name: string
    price: float64
}
```

**Step 2: Add basic annotations**

```yaml
types:
  Product:
    proto:
      name: "ProductV2"
    graphql:
      name: "ProductItem"
    openapi:
      name: "ProductResource"
```

**Result:**
- Protobuf generates: `message ProductV2 { ... }`
- GraphQL generates: `type ProductItem { ... }`
- OpenAPI generates: `ProductResource` schema

### Field Annotations

Control how individual fields are generated across formats.

**Schema:**
```typemux
type User {
    id: string
    username: string
    email: string
    phoneNumber: string
    internalNotes: string
}
```

**Annotations:**
```yaml
types:
  User:
    fields:
      username:
        required: true

      email:
        required: true
        openapi:
          extension: '{"x-format": "email", "x-example": "user@example.com"}'

      phoneNumber:
        proto:
          name: "phone_number"  # Snake case for proto
        openapi:
          extension: '{"x-format": "phone"}'

      internalNotes:
        exclude: ["openapi"]  # Don't expose in OpenAPI
```

**What this does:**
- Makes `username` and `email` required (non-nullable) in all formats
- Adds email validation hint for OpenAPI
- Renames `phoneNumber` to `phone_number` in Protobuf only
- Excludes `internalNotes` from OpenAPI (keeps it in Proto and GraphQL)

### Service Annotations

Define HTTP endpoints and GraphQL operations for your services.

**Schema:**
```typemux
service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse)
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse)
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
    rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse)
    rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse)
}
```

**Annotations:**
```yaml
services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{userId}"
        graphql: "query"
        errors: [404, 500]

      ListUsers:
        http: "GET"
        path: "/api/v1/users"
        graphql: "query"
        success: [200]
        errors: [500]

      CreateUser:
        http: "POST"
        path: "/api/v1/users"
        graphql: "mutation"
        success: [201]
        errors: [400, 409, 500]

      UpdateUser:
        http: "PUT"
        path: "/api/v1/users/{userId}"
        graphql: "mutation"
        errors: [400, 404, 500]

      DeleteUser:
        http: "DELETE"
        path: "/api/v1/users/{userId}"
        graphql: "mutation"
        success: [204]
        errors: [404, 500]
```

**Benefits:**
- OpenAPI generates proper REST endpoints with path parameters
- GraphQL categorizes operations correctly (query vs mutation)
- HTTP status codes documented for each endpoint

### Working with Namespaces

When you have types with the same name in different namespaces, use qualified names to disambiguate.

**Schema:**
```typemux
namespace com.example.users

type User {
    id: string
    username: string
    email: string
}

enum UserStatus {
    ACTIVE
    INACTIVE
}

namespace com.example.products

type User {
    id: string
    productName: string
    ownerId: string
}

enum Status {
    AVAILABLE
    SOLD_OUT
}
```

**Annotations:**
```yaml
types:
  # Use qualified names to target specific User types
  com.example.users.User:
    proto:
      name: "UserAccount"
    graphql:
      name: "UserAccount"
    fields:
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'

  com.example.products.User:
    proto:
      name: "ProductUser"
    graphql:
      name: "ProductOwner"
    fields:
      ownerId:
        required: true

enums:
  # Qualified name for UserStatus
  com.example.users.UserStatus:
    proto:
      name: "UserAccountStatus"

  # Simple name works when no ambiguity
  Status:
    proto:
      name: "ProductAvailability"
```

**Key Points:**
- Use `namespace.TypeName` format for qualified names
- Simple names work when there's no ambiguity
- Both approaches can be mixed in the same file

### Multiple Annotation Files

Split your annotations across multiple files and merge them. Later files override earlier ones.

**base.yaml** (common annotations):
```yaml
types:
  User:
    fields:
      id:
        required: true
      createdAt:
        required: true

  Product:
    fields:
      id:
        required: true
      name:
        required: true
```

**production.yaml** (production overrides):
```yaml
types:
  User:
    proto:
      name: "UserV2"  # Override for production
    fields:
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'
```

**development.yaml** (development overrides):
```yaml
types:
  User:
    proto:
      name: "UserDev"  # Different name for dev
    fields:
      debugInfo:
        exclude: []  # Include debug fields in dev
```

**Usage:**
```bash
# Production
typemux -input schema.typemux \
        -annotations base.yaml \
        -annotations production.yaml \
        -output ./generated

# Development
typemux -input schema.typemux \
        -annotations base.yaml \
        -annotations development.yaml \
        -output ./generated
```

**Merging Rules:**
- Later files override earlier files
- List values (exclude, only, errors, success) are merged
- Format-specific annotations are merged per format

---

## Reference

### YAML Structure

```yaml
types:           # Type definitions
  TypeName:
    proto:       # Protobuf-specific annotations
    graphql:     # GraphQL-specific annotations
    openapi:     # OpenAPI-specific annotations
    fields:      # Field-level annotations

enums:           # Enum definitions
  EnumName:
    proto:
    graphql:
    openapi:

unions:          # Union definitions
  UnionName:
    proto:
    graphql:
    openapi:

services:        # Service definitions
  ServiceName:
    proto:
    graphql:
    openapi:
    methods:     # Method-level annotations
```

### Type Annotations

Customize how types are generated in each format.

```yaml
types:
  TypeName:
    # Protobuf annotations
    proto:
      name: "CustomProtoName"          # Custom message name
      option: "[packed = false]"       # Proto options

    # GraphQL annotations
    graphql:
      name: "CustomGraphQLName"        # Custom type name
      directive: "@key(fields: \"id\")" # GraphQL directives

    # OpenAPI annotations
    openapi:
      name: "CustomOpenAPIName"        # Custom schema name
      extension: '{"x-internal": true}' # OpenAPI extensions (JSON)
```

**Common Use Cases:**

| Use Case | Example |
|----------|---------|
| Versioned API | `proto: { name: "UserV2" }` |
| Domain-specific naming | `graphql: { name: "UserAccount" }` |
| Federation support | `graphql: { directive: "@key(fields: \"id\")" }` |
| Internal schemas | `openapi: { extension: '{"x-internal": true}' }` |

### Field Annotations Reference

Control individual field behavior.

```yaml
types:
  TypeName:
    fields:
      fieldName:
        # Required/optional
        required: true                 # Make field required (non-nullable)
        default: "default_value"       # Set default value

        # Generator control
        exclude: ["proto", "graphql"]  # Exclude from specific generators
        only: ["openapi"]              # Include only in specific generators

        # Format-specific names
        proto:
          name: "custom_field_name"    # Custom field name for proto
          option: "[deprecated = true]" # Proto field options

        graphql:
          name: "customFieldName"      # Custom field name for GraphQL
          directive: "@external"       # GraphQL field directives

        openapi:
          name: "customFieldName"      # Custom field name for OpenAPI
          extension: '{"x-format": "email"}' # OpenAPI extensions
```

**Common Patterns:**

**Email Field:**
```yaml
email:
  required: true
  openapi:
    extension: '{"x-format": "email", "x-example": "user@example.com"}'
```

**Internal Field:**
```yaml
internalNotes:
  exclude: ["openapi"]  # Don't expose in public API
```

**Deprecated Field:**
```yaml
oldField:
  proto:
    option: "[deprecated = true]"
  graphql:
    directive: "@deprecated(reason: \"Use newField instead\")"
```

**Phone Number:**
```yaml
phoneNumber:
  openapi:
    extension: '{"x-format": "phone", "pattern": "^\\+?[1-9]\\d{1,14}$"}'
```

### Enum Annotations

```yaml
enums:
  EnumName:
    proto:
      name: "CustomEnumName"
      option: "[allow_alias = true]"

    graphql:
      name: "CustomEnumName"
      directive: "@deprecated"

    openapi:
      name: "CustomEnumName"
      extension: '{"x-extensible-enum": true}'
```

### Union Annotations

```yaml
unions:
  UnionName:
    proto:
      name: "CustomUnionName"

    graphql:
      name: "CustomUnionName"

    openapi:
      name: "CustomUnionName"
      extension: '{"discriminator": {"propertyName": "type"}}'
```

### Service Annotations Reference

Customize service generation.

```yaml
services:
  ServiceName:
    proto:
      name: "CustomServiceName"

    graphql:
      name: "CustomServiceName"

    openapi:
      name: "CustomServiceName"
```

### Method Annotations

Define HTTP mappings, GraphQL operations, and status codes.

```yaml
services:
  ServiceName:
    methods:
      MethodName:
        http: "GET"                           # HTTP method
        path: "/api/v1/resource/{id}"         # URL path with parameters
        graphql: "query"                      # GraphQL operation type
        success: [200, 201, 204]              # Success status codes
        errors: [400, 404, 500]               # Error status codes
        proto:
          option: "[idempotency_level = IDEMPOTENT]"
```

**HTTP Methods:**
- `GET` - Retrieve resources
- `POST` - Create new resources
- `PUT` - Update/replace resources
- `PATCH` - Partial update resources
- `DELETE` - Delete resources

**GraphQL Operations:**
- `query` - Read operations (GET-like)
- `mutation` - Write operations (POST/PUT/DELETE-like)
- `subscription` - Real-time updates

**Path Parameters:**
Use `{paramName}` in paths:
```yaml
path: "/api/v1/users/{userId}/posts/{postId}"
```

**Status Codes:**

| Code | Meaning | When to Use |
|------|---------|-------------|
| 200 | OK | Successful GET/PUT/PATCH |
| 201 | Created | Successful POST |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid input |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource already exists |
| 500 | Internal Server Error | Server error |

---

## Best Practices

### 1. Use Qualified Names for Namespaces

When working with multiple namespaces, always use qualified names to avoid ambiguity:

✅ **Good:**
```yaml
types:
  com.example.users.User:
    proto:
      name: "UserAccount"
```

❌ **Avoid:**
```yaml
types:
  User:  # Ambiguous when multiple User types exist
    proto:
      name: "UserAccount"
```

### 2. Organize Annotations by Environment

Split annotations into base and environment-specific files:

```
annotations/
  base.yaml           # Common annotations
  production.yaml     # Production-specific
  development.yaml    # Development-specific
  staging.yaml        # Staging-specific
```

### 3. Group Related Annotations

Keep related annotations together for readability:

```yaml
types:
  User:
    # All format names together
    proto:
      name: "UserV2"
    graphql:
      name: "UserAccount"
    openapi:
      name: "UserProfile"

    # All fields together
    fields:
      # Related fields grouped
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'

      phoneNumber:
        required: false
        openapi:
          extension: '{"x-format": "phone"}'
```

### 4. Document Your Annotations

Add comments to explain non-obvious choices:

```yaml
types:
  User:
    proto:
      name: "UserV2"  # V2 for backward compatibility with existing clients

    fields:
      internalScore:
        exclude: ["openapi"]  # Internal scoring, not exposed to public API
```

### 5. Validate Early and Often

Run with validation after making changes:

```bash
# This will validate before generating
typemux -input schema.typemux -annotations annotations.yaml -output ./generated
```

### 6. Use Consistent Naming Conventions

Choose a naming convention and stick with it:

**Proto:** `snake_case` for fields, `PascalCase` for messages
```yaml
proto:
  name: "UserAccount"
  fields:
    phoneNumber:
      proto:
        name: "phone_number"
```

**GraphQL:** `camelCase` for fields, `PascalCase` for types
```yaml
graphql:
  name: "UserAccount"
  fields:
    phoneNumber:
      graphql:
        name: "phoneNumber"
```

### 7. Leverage Exclude/Only Strategically

Use `exclude` for sensitive data:
```yaml
fields:
  passwordHash:
    exclude: ["graphql", "openapi"]  # Only in proto for storage

  internalNotes:
    exclude: ["openapi"]  # Internal use only
```

Use `only` when a field is format-specific:
```yaml
fields:
  graphqlCursor:
    only: ["graphql"]  # GraphQL pagination cursor
```

---

## Validation

TypeMUX validates all YAML annotations before code generation. Validation catches:

### Reference Validation

Ensures all referenced entities exist:

```yaml
types:
  NonExistentType:  # ❌ Error: type doesn't exist
    proto:
      name: "Foo"

  User:
    fields:
      nonExistentField:  # ❌ Error: field doesn't exist
        required: true
```

**Error Output:**
```
Found 2 validation error(s) in YAML annotations:

  • YAML annotation error at types.NonExistentType: references non-existent type 'NonExistentType'
  • YAML annotation error at types.User.fields.nonExistentField: references non-existent field 'User.nonExistentField'
```

### Conflict Detection

Prevents conflicting annotations:

```yaml
fields:
  email:
    exclude: ["proto"]
    only: ["graphql"]  # ❌ Error: can't have both exclude and only
```

**Error Output:**
```
YAML annotation error at types.User.fields.email: cannot specify both 'exclude' and 'only' annotations
```

### Value Validation

Validates annotation values:

```yaml
services:
  UserService:
    methods:
      GetUser:
        http: "INVALID"  # ❌ Error: invalid HTTP method
        graphql: "invalid"  # ❌ Error: invalid GraphQL operation
        errors: [999]  # ❌ Error: invalid status code
```

**Error Output:**
```
Found 3 validation error(s) in YAML annotations:

  • YAML annotation error at services.UserService.methods.GetUser: invalid HTTP method: 'INVALID'
  • YAML annotation error at services.UserService.methods.GetUser: invalid GraphQL operation type: 'invalid'
  • YAML annotation error at services.UserService.methods.GetUser: invalid HTTP status code in errors: 999
```

### Valid Values

**HTTP Methods:** `GET`, `POST`, `PUT`, `PATCH`, `DELETE`

**GraphQL Operations:** `query`, `mutation`, `subscription`

**Generators:** `proto`, `graphql`, `openapi`

**Status Codes:** `100-599`

---

## Examples

### Complete E-Commerce API

**schema.typemux:**
```typemux
namespace com.example.shop

type Product {
    id: string
    name: string
    description: string
    price: float64
    stockCount: int32
    imageUrls: []string
    internalNotes: string
}

type Order {
    id: string
    userId: string
    productIds: []string
    totalAmount: float64
    status: OrderStatus
    createdAt: timestamp
}

enum OrderStatus {
    PENDING
    CONFIRMED
    SHIPPED
    DELIVERED
    CANCELLED
}

service ShopService {
    rpc GetProduct(GetProductRequest) returns (GetProductResponse)
    rpc ListProducts(ListProductsRequest) returns (ListProductsResponse)
    rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse)
    rpc GetOrder(GetOrderRequest) returns (GetOrderResponse)
}
```

**annotations.yaml:**
```yaml
types:
  Product:
    proto:
      name: "ProductV2"
    graphql:
      name: "ProductItem"
    fields:
      name:
        required: true
      description:
        required: true
      price:
        required: true
        openapi:
          extension: '{"x-format": "currency", "minimum": 0}'
      stockCount:
        required: true
        proto:
          name: "stock_count"
        openapi:
          extension: '{"minimum": 0}'
      imageUrls:
        proto:
          name: "image_urls"
      internalNotes:
        exclude: ["openapi", "graphql"]  # Internal only

  Order:
    proto:
      name: "OrderV2"
    fields:
      userId:
        required: true
        proto:
          name: "user_id"
      productIds:
        required: true
        proto:
          name: "product_ids"
      totalAmount:
        required: true
        proto:
          name: "total_amount"
        openapi:
          extension: '{"x-format": "currency"}'
      status:
        required: true
      createdAt:
        required: true
        proto:
          name: "created_at"

enums:
  OrderStatus:
    proto:
      name: "OrderStatusEnum"

services:
  ShopService:
    methods:
      GetProduct:
        http: "GET"
        path: "/api/v1/products/{productId}"
        graphql: "query"
        errors: [404, 500]

      ListProducts:
        http: "GET"
        path: "/api/v1/products"
        graphql: "query"
        errors: [500]

      CreateOrder:
        http: "POST"
        path: "/api/v1/orders"
        graphql: "mutation"
        success: [201]
        errors: [400, 409, 500]

      GetOrder:
        http: "GET"
        path: "/api/v1/orders/{orderId}"
        graphql: "query"
        errors: [404, 500]
```

### Multi-Tenant System

**schema.typemux:**
```typemux
namespace com.example.tenants

type Tenant {
    id: string
    name: string
    subdomain: string
    planType: PlanType
    apiKey: string
    secretKey: string
}

type User {
    id: string
    tenantId: string
    email: string
    role: UserRole
}

enum PlanType {
    FREE
    PROFESSIONAL
    ENTERPRISE
}

enum UserRole {
    ADMIN
    MEMBER
    GUEST
}
```

**annotations.yaml:**
```yaml
types:
  Tenant:
    fields:
      name:
        required: true
      subdomain:
        required: true
        openapi:
          extension: '{"pattern": "^[a-z0-9-]+$", "x-example": "acme-corp"}'
      planType:
        required: true
        proto:
          name: "plan_type"
      apiKey:
        required: true
        exclude: ["graphql"]  # Don't expose in GraphQL
        proto:
          name: "api_key"
        openapi:
          extension: '{"x-format": "uuid"}'
      secretKey:
        required: true
        exclude: ["graphql", "openapi"]  # Never expose secret
        proto:
          name: "secret_key"

  User:
    fields:
      tenantId:
        required: true
        proto:
          name: "tenant_id"
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'
      role:
        required: true

enums:
  PlanType:
    proto:
      name: "TenantPlanType"

  UserRole:
    proto:
      name: "TenantUserRole"
```

### Microservices with Namespaces

**schema.typemux:**
```typemux
namespace com.example.users

type User {
    id: string
    username: string
    email: string
}

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse)
}

namespace com.example.orders

type User {
    id: string
    orderCount: int32
    totalSpent: float64
}

service OrderService {
    rpc GetUserOrders(GetUserOrdersRequest) returns (GetUserOrdersResponse)
}
```

**annotations.yaml:**
```yaml
types:
  com.example.users.User:
    proto:
      name: "UserProfile"
    graphql:
      name: "UserAccount"
    fields:
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'

  com.example.orders.User:
    proto:
      name: "UserOrderStats"
    graphql:
      name: "UserOrderSummary"
    fields:
      orderCount:
        required: true
        proto:
          name: "order_count"
      totalSpent:
        required: true
        proto:
          name: "total_spent"

services:
  com.example.users.UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{userId}"
        graphql: "query"

  com.example.orders.OrderService:
    methods:
      GetUserOrders:
        http: "GET"
        path: "/api/v1/users/{userId}/orders"
        graphql: "query"
```

---

## Troubleshooting

### Common Issues

**Issue: "references non-existent type"**
```
Error: YAML annotation error at types.User: references non-existent type 'User'
```

**Solution:** Check your type name matches exactly (case-sensitive). If using namespaces, use the qualified name:
```yaml
types:
  com.example.api.User:  # Not just 'User'
```

---

**Issue: "cannot specify both 'exclude' and 'only'"**
```
Error: cannot specify both 'exclude' and 'only' annotations
```

**Solution:** Use either `exclude` OR `only`, never both:
```yaml
# ✅ Correct
fields:
  field1:
    exclude: ["proto"]

# ✅ Also correct
fields:
  field2:
    only: ["graphql"]

# ❌ Wrong
fields:
  field3:
    exclude: ["proto"]
    only: ["graphql"]
```

---

**Issue: Annotations not being applied**

**Solution:** Verify annotation file is specified:
```bash
typemux -input schema.typemux -annotations annotations.yaml -output ./generated
#                              ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ Don't forget this!
```

---

**Issue: Wrong type being annotated with namespaces**

**Solution:** Use qualified names to be explicit:
```yaml
types:
  # ❌ Ambiguous - could match any User
  User:
    proto:
      name: "UserV2"

  # ✅ Explicit - matches specific User
  com.example.users.User:
    proto:
      name: "UserAccountV2"
```

---

## Additional Resources

- **Specification:** See [YAML_ANNOTATIONS.md](YAML_ANNOTATIONS.md) for technical specification
- **Examples:** Check `examples/yaml-annotations/` and `examples/namespaces/` directories
- **Schema Syntax:** See main [README.md](../README.md) for TypeMUX IDL syntax

---

## Feedback

Found an issue or have a suggestion? Please [open an issue](https://github.com/rasmartins/typemux/issues) on GitHub.
