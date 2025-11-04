# Post

Post entity

## Format Representations

### GraphQL

```graphql
type Post {
  id: string!
  title: string!
  content: string!
  authorId: string!
  published: bool!
  createdAt: timestamp!
}```

### OpenAPI

```yaml
Post:
  type: object
  properties:
    id:
      type: string
    title:
      type: string
    content:
      type: string
    authorId:
      type: string
    published:
      type: boolean
    createdAt:
      $ref: '#/components/schemas/timestamp'
```

### Protobuf

```protobuf
message Post {
  string id = 1;
  string title = 2;
  string content = 3;
  string authorId = 4;
  bool published = 5;
  timestamp createdAt = 6;
}```

## Fields

### id

**Type:** `string` (required)

### title

**Type:** `string` (required)

### content

**Type:** `string` (required)

### authorId

**Type:** `string` (required)

### published

**Type:** `bool` (required)

### createdAt

**Type:** `timestamp` (required)

