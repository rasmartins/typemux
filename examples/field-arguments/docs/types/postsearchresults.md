# PostSearchResults

Search results for posts

## Format Representations

### GraphQL

```graphql
type PostSearchResults {
  posts: [Post]!
  metadata: SearchMetadata!
}```

### OpenAPI

```yaml
PostSearchResults:
  type: object
  properties:
    posts:
      type: array
      items:
        $ref: '#/components/schemas/Post'
    metadata:
      $ref: '#/components/schemas/SearchMetadata'
```

### Protobuf

```protobuf
message PostSearchResults {
  repeated Post posts = 1;
  SearchMetadata metadata = 2;
}```

## Fields

### posts

**Type:** `[]Post` (required)

### metadata

**Type:** `SearchMetadata` (required)

