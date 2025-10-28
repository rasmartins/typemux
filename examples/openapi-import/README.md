## OpenAPI Import Example

This example demonstrates how to import existing OpenAPI 3.0 specifications and convert them to TypeMUX IDL format.

## Files

- `example.yaml` - Original OpenAPI 3.0 specification
- `example.typemux` - Converted TypeMUX IDL (generated)
- `generated/openapi.yaml` - Round-trip generated OpenAPI (for verification)

## Converting OpenAPI to TypeMUX

Use the `openapi2typemux` tool to convert OpenAPI specifications:

```bash
openapi2typemux --input example.yaml --output ./
```

Or from the repository root:

```bash
go run cmd/openapi2typemux/main.go --input examples/openapi-import/example.yaml --output examples/openapi-import
```

## Features Demonstrated

### 1. Schemas to Types

```yaml
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: string
          description: Unique identifier
        name:
          type: string
        age:
          type: integer
          format: int32
        tags:
          type: array
          items:
            type: string
```

Converts to:
```typemux
type Pet {
  // Unique identifier
  id: string = 1
  name: string = 2
  age: int32 = 3
  tags: []string = 4
}
```

### 2. Enums

```yaml
PetStatus:
  type: string
  enum:
    - available
    - pending
    - sold
```

Converts to:
```typemux
enum PetStatus {
  AVAILABLE = 0
  PENDING = 1
  SOLD = 2
}
```

### 3. REST Endpoints to RPC Methods

```yaml
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PetList'
```

Converts to:
```typemux
service PetStoreAPIService {
  // List all pets
  // GET /pets
  rpc ListPets(ListPetsRequest) returns (PetList)
}
```

### 4. Request/Response Mappings

- **GET** requests with parameters → Request type with parameters as fields
- **POST/PUT** requests with body → Request type from requestBody schema
- **200/201** responses → Return type from response schema
- **204** No Content → Returns `Empty`

### 5. Type Mappings

| OpenAPI Type | Format | TypeMUX Type | Notes |
|--------------|--------|--------------|-------|
| `string` | - | `string` | |
| `string` | `date-time` | `timestamp` | |
| `string` | `date` | `timestamp` | |
| `integer` | `int32` | `int32` | |
| `integer` | `int64` | `int64` | |
| `number` | `float` | `float` | |
| `number` | `double` | `double` | |
| `boolean` | - | `bool` | |
| `array` | - | `[]Type` | |
| `object` | - | Named type or `map<string, string>` | |

### 6. References

OpenAPI `$ref` references are automatically resolved:

```yaml
schema:
  $ref: '#/components/schemas/Pet'
```

Converts to direct type usage:
```typemux
returns (Pet)
```

### 7. Descriptions and Documentation

All descriptions from OpenAPI are preserved as comments in TypeMUX:

```yaml
description: |
  Returns a list of all pets in the store.
  Supports pagination via limit and offset parameters.
```

Converts to:
```typemux
// Returns a list of all pets in the store.
// Supports pagination via limit and offset parameters.
```

## Round-Trip Conversion

After converting to TypeMUX, you can generate back to OpenAPI:

```bash
# Convert TypeMUX back to OpenAPI
typemux --input example.typemux --output ./generated --format openapi

# The generated openapi.yaml will be functionally equivalent
```

## What's Preserved

✅ Schema definitions (types)
✅ Enum values (auto-numbered)
✅ Field descriptions
✅ Array types
✅ Required fields (mapped to proto semantics)
✅ Operation summaries and descriptions
✅ HTTP methods (as comments)
✅ Path templates (as comments)
✅ Response types
✅ Nested objects
✅ Type formats (int32, int64, timestamp, etc.)

## What's Not Preserved

❌ HTTP-specific details (status codes, content types)
❌ Request/response headers
❌ Security schemes
❌ Servers and base URLs
❌ Parameter locations (query vs path - all become fields)
❌ Example values
❌ Validation constraints (min/max, patterns)
❌ oneOf/anyOf/allOf (merged or simplified)
❌ Callbacks and webhooks

## Reserved Keywords

TypeMUX reserved keywords are automatically escaped with a trailing underscore:

- `type` → `type_`
- `namespace` → `namespace_`
- `service` → `service_`
- etc.

For example:
```yaml
properties:
  type:
    type: string
```

Converts to:
```typemux
type_: string = 1
```

## Use Cases

1. **API Unification**: Convert REST APIs to TypeMUX for multi-format generation
2. **gRPC from REST**: Import OpenAPI specs and generate gRPC/Protobuf services
3. **GraphQL from REST**: Generate GraphQL schemas from existing REST APIs
4. **Schema Migration**: Modernize legacy OpenAPI specs with TypeMUX
5. **Multi-Protocol Services**: Support REST, gRPC, and GraphQL from one schema

## Testing with Real-World Specs

This importer has been tested with production OpenAPI specifications from Anchorage Digital's API, including:
- Complex nested schemas
- Multiple endpoints with various HTTP methods
- Query and path parameters
- Request bodies and response schemas
- Enumerations
- Array types and references

Example:
```bash
openapi2typemux --input ~/path/to/production/openapi.yaml --output ./imported
```

## Known Limitations

1. **REST to RPC Mapping**: OpenAPI's REST semantics are mapped to RPC-style methods, losing some REST-specific patterns
2. **Parameter Flattening**: Query, path, and header parameters all become fields in the request type
3. **Status Codes**: Only 200/201 responses are used for return types; error responses are not modeled
4. **Content Negotiation**: Only `application/json` is considered; other content types are ignored
5. **HTTP Semantics**: Idempotency, caching, and other HTTP-specific behaviors are not preserved

## Best Practices

1. **OperationIDs**: Use explicit `operationId` fields in your OpenAPI spec for cleaner method names
2. **Schema References**: Use `$ref` for reusable schemas rather than inline definitions
3. **Descriptions**: Add detailed descriptions to schemas and operations - they're preserved as comments
4. **Component Schemas**: Define all types in `components/schemas` for better organization
5. **Consistent Naming**: Use consistent naming conventions for schemas and operations

## Example Workflow

1. **Import existing OpenAPI**:
   ```bash
   openapi2typemux --input api.yaml --output ./schemas
   ```

2. **Generate multiple formats**:
   ```bash
   typemux --input schemas/api.typemux --output ./generated --format all
   ```

3. **Result**: You now have GraphQL, Protobuf, and OpenAPI from your original REST API specification!
