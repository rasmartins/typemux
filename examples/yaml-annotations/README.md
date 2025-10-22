# YAML Annotations Example

This example demonstrates how to use YAML files to annotate TypeMUX schemas instead of inline annotations.

## Files

- `schema.typemux` - The TypeMUX schema without any annotations
- `annotations.yaml` - YAML file containing all annotations
- `invalid.yaml` - Example of invalid annotations (for testing validation)

## Benefits of YAML Annotations

1. **Separation of concerns** - Keep schema definitions separate from metadata
2. **Easier maintenance** - Update annotations without touching schema
3. **Bulk operations** - Easier to manage annotations across many types
4. **External configuration** - Annotations can be managed separately from code

## Usage

```bash
# Generate with YAML annotations
typemux -input schema.typemux \
        -annotations annotations.yaml \
        -output generated

# Multiple annotation files (merged in order)
typemux -input schema.typemux \
        -annotations base.yaml \
        -annotations overrides.yaml \
        -output generated
```

## Annotation Structure

### Type Annotations

```yaml
types:
  User:
    # Type-level name annotations
    proto.name: "UserV2"
    graphql.name: "UserAccount"
    openapi.name: "UserProfile"

    # Field annotations
    fields:
      email:
        required: true
        openapi.extension: '{"x-format": "email"}'
```

### Service Annotations

```yaml
services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{userId}"
        graphql: "query"
        errors: [404, 500]
```

## Generated Output

The example generates three schemas:

1. **Protobuf** (`schema.proto`) - Uses `UserV2` and `ProductV3` names
2. **GraphQL** (`schema.graphql`) - Uses `UserAccount` name
3. **OpenAPI** (`openapi.yaml`) - Uses `UserProfile` name

## Validation

The tool validates YAML annotations and reports errors:

```bash
$ typemux -input schema.typemux -annotations invalid.yaml

Found 4 validation error(s) in YAML annotations:

  • YAML annotation error at types.User.fields.nonexistent: references non-existent field 'User.nonexistent'
  • YAML annotation error at types.NonExistentType: references non-existent type 'NonExistentType'
  • YAML annotation error at services.NonExistentService: references non-existent service 'NonExistentService'
  • YAML annotation error at services.UserService.methods.NonExistentMethod: references non-existent method 'UserService.NonExistentMethod'
```

## Precedence Rules

- YAML annotations override inline annotations
- Multiple YAML files are merged in order (later files override earlier ones)
- List values (exclude, errors, success) are merged, not replaced
