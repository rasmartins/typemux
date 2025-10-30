# Map Types Example

This example demonstrates how TypeMux handles `map<K, V>` types with custom values across different output formats.

## Overview

TypeMux supports map types with both primitive and custom type values. Each target format handles maps differently:

### Protobuf
Maps are represented using Protocol Buffers' native `map<K, V>` syntax:

```protobuf
message Inventory {
  map<string, Product> productsByWarehouse = 1;
  map<string, int32> quantities = 2;
  map<string, Settings> featureSettings = 3;
}
```

**Key points:**
- Direct support for `map<string, CustomType>`
- Keys must be string or integral types
- Values can be any message type or primitive

### GraphQL
Maps are converted to arrays of key-value entry types:

```graphql
type StringProductEntry {
  key: String!
  value: Product!
}

type Inventory {
  productsByWarehouse: [StringProductEntry!]!
  quantities: [StringIntEntry!]!
}
```

**Key points:**
- Each unique map signature gets its own entry type
- Entry types are automatically generated
- Both output and input variants are created
- Provides type safety in GraphQL schema

### OpenAPI
Maps are represented as objects with `additionalProperties`:

```yaml
Inventory:
  type: object
  properties:
    productsByWarehouse:
      type: object
      description: Map of string to Product
      additionalProperties:
        $ref: '#/components/schemas/Product'
    quantities:
      type: object
      description: Map of string to int32
      additionalProperties:
        type: integer
        format: int32
```

**Key points:**
- Uses JSON object representation
- `additionalProperties` defines value type
- Can reference complex schemas
- Standard REST API pattern

## Example Types

### Simple Map (Primitives)
```typemux
type Example {
    metadata: map<string, string>
}
```

- **Proto:** `map<string, string> metadata = 1;`
- **GraphQL:** `metadata: [StringStringEntry!]`
- **OpenAPI:** `additionalProperties: { type: string }`

### Map with Custom Type
```typemux
type Inventory {
    productsByWarehouse: map<string, Product>
}
```

- **Proto:** `map<string, Product> productsByWarehouse = 1;`
- **GraphQL:** `productsByWarehouse: [StringProductEntry!]!`
- **OpenAPI:** `additionalProperties: { $ref: '#/components/schemas/Product' }`

### Map with Multiple Custom Types
```typemux
type UserPreferences {
    featureSettings: map<string, Settings>
    friends: map<string, User>
}
```

Each map gets its own entry type in GraphQL:
- `StringSettingsEntry` with `Settings` value
- `StringUserEntry` with `User` value

## Use Cases Demonstrated

1. **Inventory Management** - Maps for warehouse-to-product relationships
2. **Shopping Cart** - Product maps with quantities
3. **User Preferences** - Configuration maps with custom settings
4. **Metadata Storage** - Simple string-to-string maps

## Running the Example

Generate all formats:
```bash
typemux -input map_example.typemux -output generated -format all
```

Generate specific format:
```bash
typemux -input map_example.typemux -output generated -format protobuf
typemux -input map_example.typemux -output generated -format graphql
typemux -input map_example.typemux -output generated -format openapi
```

## Generated Files

- `generated/schema.proto` - Protocol Buffers schema
- `generated/schema.graphql` - GraphQL schema with entry types
- `generated/openapi.yaml` - OpenAPI specification
- `generated/types.go` - Go type definitions
- `generated/API.md` - API documentation

## Best Practices

1. **Use descriptive map names** - `productsByWarehouse` is better than `items`
2. **Document map semantics** - Explain what keys and values represent
3. **Consider alternatives** - For small, fixed sets, use regular fields
4. **Test with generators** - Verify output in all target formats
5. **Mind performance** - Large maps may have implications in some formats
