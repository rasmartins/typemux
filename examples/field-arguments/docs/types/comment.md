# Comment

Comment entity

## Format Representations

### GraphQL

```graphql
type Comment {
  id: string!
  postId: string!
  authorId: string!
  content: string!
  createdAt: timestamp!
}```

### OpenAPI

```yaml
Comment:
  type: object
  properties:
    id:
      type: string
    postId:
      type: string
    authorId:
      type: string
    content:
      type: string
    createdAt:
      $ref: '#/components/schemas/timestamp'
```

### Protobuf

```protobuf
message Comment {
  string id = 1;
  string postId = 2;
  string authorId = 3;
  string content = 4;
  timestamp createdAt = 5;
}```

## Fields

### id

**Type:** `string` (required)

### postId

**Type:** `string` (required)

### authorId

**Type:** `string` (required)

### content

**Type:** `string` (required)

### createdAt

**Type:** `timestamp` (required)

