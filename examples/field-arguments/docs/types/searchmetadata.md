# SearchMetadata

Search result metadata

## Format Representations

### GraphQL

```graphql
type SearchMetadata {
  totalResults: int32!
  page: int32!
  pageSize: int32!
  hasNextPage: bool!
}```

### OpenAPI

```yaml
SearchMetadata:
  type: object
  properties:
    totalResults:
      type: integer
    page:
      type: integer
    pageSize:
      type: integer
    hasNextPage:
      type: boolean
```

### Protobuf

```protobuf
message SearchMetadata {
  int32 totalResults = 1;
  int32 page = 2;
  int32 pageSize = 3;
  bool hasNextPage = 4;
}```

## Fields

### totalResults

**Type:** `int32` (required)

### page

**Type:** `int32` (required)

### pageSize

**Type:** `int32` (required)

### hasNextPage

**Type:** `bool` (required)

