# Namespace Support Example

This example demonstrates how to use qualified names in YAML annotations to disambiguate types, enums, and services that have the same name but exist in different namespaces.

## Problem

When your schema has multiple types with the same name in different namespaces:

```typemux
namespace com.example.users
type User { ... }

namespace com.example.products
type User { ... }
```

You need a way to specify which `User` type you want to annotate in your YAML file.

## Solution

Use qualified names in the YAML annotations file:

```yaml
types:
  # Target the User type in users namespace
  com.example.users.User:
    proto.name: "UserAccount"

  # Target the User type in products namespace
  com.example.products.User:
    proto.name: "ProductUser"
```

## Files

- `schema.typemux` - Schema with two `User` types in different namespaces
- `annotations.yaml` - YAML annotations using qualified names
- `generated/` - Generated output showing different names for each User type

## Running the Example

```bash
# From the repository root
go run main.go -input examples/namespaces/schema.typemux \
               -annotations examples/namespaces/annotations.yaml \
               -output examples/namespaces/generated
```

## Expected Output

The generator will create:

1. **com/example/users.proto** - Contains `UserAccount` message (renamed from User)
2. **com/example/products.proto** - Contains `ProductUser` message (renamed from User)
3. **schema.graphql** - Shows error about duplicate User types (expected, as GraphQL doesn't support namespaces)
4. **openapi.yaml** - OpenAPI specification with both User types

## Key Points

1. **Qualified names** use the format `namespace.TypeName`
2. **Simple names** work when there's no ambiguity (e.g., `GetProductRequest`)
3. **Best practice**: Always use qualified names when you have duplicate names across namespaces
4. Works for types, enums, unions, and services
