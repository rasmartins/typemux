# PostFilter

Filter options for posts

## Format Representations

### GraphQL

```graphql
type PostFilter {
  published: bool
  authorId: string
  minDate: timestamp
  maxDate: timestamp
}```

### OpenAPI

```yaml
PostFilter:
  type: object
  properties:
    published:
      type: boolean
    authorId:
      type: string
    minDate:
      $ref: '#/components/schemas/timestamp'
    maxDate:
      $ref: '#/components/schemas/timestamp'
```

### Protobuf

```protobuf
message PostFilter {
  optional bool published = 1;
  optional string authorId = 2;
  optional timestamp minDate = 3;
  optional timestamp maxDate = 4;
}```

## Fields

### published

**Type:** `bool`

### authorId

**Type:** `string`

### minDate

**Type:** `timestamp`

### maxDate

**Type:** `timestamp`

