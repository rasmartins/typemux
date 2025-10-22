# YAML Annotations Specification

## Overview

TypeMUX supports defining annotations in a companion YAML file instead of inline in `.typemux` files. This allows for:
- Separation of concerns (schema vs metadata)
- Easier bulk annotation management
- External configuration without modifying schema files

## File Structure

```yaml
# Top-level structure
types:
  # You can use either simple names or qualified names (namespace.TypeName)
  TypeName:
    # Type-level annotations - organized by format
    proto:
      name: "CustomProtoName"
      option: "[packed = false]"
    graphql:
      name: "CustomGraphQLName"
      directive: "@key(fields: \"id\")"
    openapi:
      name: "CustomOpenAPIName"
      extension: '{"x-internal": true}'

    # Field-level annotations
    fields:
      fieldName:
        required: true
        default: "value"
        exclude: ["proto", "graphql"]
        only: ["openapi"]
        proto:
          name: "custom_field_name"
          option: "[deprecated = true]"
        graphql:
          directive: "@external"
        openapi:
          extension: '{"x-format": "email"}'

  # Using qualified names to disambiguate types in different namespaces
  com.example.users.User:
    proto:
      name: "UserAccount"

  com.example.products.User:
    proto:
      name: "ProductUser"

enums:
  # Simple or qualified names
  EnumName:
    proto:
      name: "CustomEnumName"
    graphql:
      name: "CustomEnumName"

  com.example.api.Status:
    proto:
      name: "ApiStatus"

unions:
  UnionName:
    proto:
      name: "CustomUnionName"
    graphql:
      name: "CustomUnionName"

services:
  # Simple or qualified names
  ServiceName:
    proto:
      name: "CustomServiceName"

    methods:
      MethodName:
        http: "GET"
        path: "/api/v1/resource/{id}"
        graphql: "query"
        success: [201, 202]
        errors: [400, 404, 500]
        proto:
          option: "[idempotency_level = IDEMPOTENT]"

  com.example.api.UserService:
    methods:
      GetUser:
        http: "GET"
```

## Usage

Specify the annotations file via CLI:

```bash
# Single annotation file
typemux generate schema.typemux --annotations annotations.yaml

# Multiple annotation files (merged in order)
typemux generate schema.typemux --annotations base.yaml --annotations overrides.yaml
```

## Annotation Types

### Type-Level Annotations

| Annotation | Value Type | Example |
|------------|-----------|---------|
| `proto.name` | string | `"UserV2"` |
| `graphql.name` | string | `"UserAccount"` |
| `openapi.name` | string | `"UserProfile"` |
| `proto.option` | string | `"[packed = false]"` |
| `graphql.directive` | string | `"@key(fields: \"id\")"` |
| `openapi.extension` | JSON string | `'{"x-internal": true}'` |

### Field-Level Annotations

| Annotation | Value Type | Example |
|------------|-----------|---------|
| `required` | boolean | `true` |
| `default` | string/number/boolean | `"default_value"` or `42` or `true` |
| `exclude` | array of strings | `["proto", "graphql"]` |
| `only` | array of strings | `["openapi"]` |
| `proto.name` | string | `"custom_name"` |
| `proto.option` | string | `"[packed = false]"` |
| `graphql.directive` | string | `"@external"` |
| `openapi.extension` | JSON string | `'{"x-format": "email"}'` |

### Method-Level Annotations

| Annotation | Value Type | Example |
|------------|-----------|---------|
| `http` | string | `"GET"` or `"POST"` |
| `path` | string | `"/api/v1/users/{id}"` |
| `graphql` | string | `"query"` or `"mutation"` |
| `success` | array of integers | `[201, 202]` |
| `errors` | array of integers | `[400, 404, 500]` |
| `proto.option` | string | `"[idempotency_level = IDEMPOTENT]"` |

## Namespace Support

TypeMUX supports using qualified names in YAML annotations to disambiguate types, enums, unions, and services that have the same name but exist in different namespaces.

### Qualified Names

Use the format `namespace.TypeName` to reference entities:

```yaml
types:
  # Simple name - works when there's no ambiguity
  User:
    proto.name: "UserV2"

  # Qualified name - required when multiple types have the same simple name
  com.example.users.User:
    proto.name: "UserAccount"

  com.example.products.User:
    proto.name: "ProductUser"
```

### Matching Rules

1. **Qualified names take precedence**: If you specify `com.example.api.User`, it will only match the `User` type in the `com.example.api` namespace
2. **Simple names match any namespace**: If you specify just `User`, it will match the first `User` type found (use with caution when you have duplicate names)
3. **Best practice**: Use qualified names when your schema has multiple types with the same name

### Example with Namespaces

**schema.typemux:**
```typemux
namespace com.example.users

type User {
    id: string
    username: string
}

namespace com.example.products

type User {
    id: string
    productName: string
}

service ProductService {
    rpc GetProduct(GetProductRequest) returns (GetProductResponse)
}
```

**annotations.yaml:**
```yaml
types:
  # Target the User type in users namespace
  com.example.users.User:
    proto:
      name: "UserAccount"
    graphql:
      name: "UserAccount"
    fields:
      username:
        required: true

  # Target the User type in products namespace
  com.example.products.User:
    proto:
      name: "ProductUser"
    graphql:
      name: "ProductOwner"
    fields:
      productName:
        required: true

services:
  # Target service with qualified name
  com.example.products.ProductService:
    methods:
      GetProduct:
        http: "GET"
        path: "/api/v1/products/{id}"
```

## Precedence Rules

When the same annotation exists in both YAML and inline:

1. **YAML annotations override inline annotations** (configurable override behavior)
2. Annotations merge at the field/method level
3. For list values (like `exclude`, `errors`), values are merged

## Validation

The parser validates:

1. **Reference validation**: All types, fields, enums, unions, services, and methods referenced in YAML must exist in the schema
2. **Annotation validity**: Annotations must be valid for their context (e.g., can't use `@http` on a type)
3. **Value validation**: Annotation values must be the correct type (e.g., `required` must be boolean)
4. **Conflict detection**: Warns or errors on conflicting annotations

## Errors

```
Error: YAML annotation references non-existent type: 'UserProfile'
Error: YAML annotation references non-existent field 'User.emailAddress'
Error: YAML annotation references non-existent method 'UserService.DeleteUser'
Error: Invalid annotation 'http' for type 'User' (only valid for service methods)
Error: Invalid value for 'required': expected boolean, got string
```

## Example

**schema.typemux:**
```typemux
namespace com.example.api

type User {
    id: string = 1 @required
    username: string = 2
    email: string = 3
}

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse)
}
```

**annotations.yaml:**
```yaml
types:
  User:
    proto.name: "UserV2"
    graphql.name: "UserAccount"
    openapi.name: "UserProfile"
    fields:
      username:
        required: true
      email:
        required: true
        proto.option: "[deprecated = true]"

services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{id}"
        graphql: "query"
        errors: [404, 500]
```

**Result**: The `User` type will be named differently in each format, `username` and `email` will be required (merged with inline `@required` on `id`), and the `GetUser` method will have HTTP and GraphQL mappings.
