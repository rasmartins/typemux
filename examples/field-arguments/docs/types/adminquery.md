# AdminQuery

Example with format-specific annotations on arguments

## Format Representations

### GraphQL

```graphql
type AdminQuery {
}```

### OpenAPI

```yaml
AdminQuery:
  type: object
  properties:
```

### Protobuf

```protobuf
message AdminQuery {
}```

## Fields

### userById

Get user with GraphQL-specific annotations on arguments

**Type:** `User`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| id | `string` | Yes | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  userById(id: "example") {
    id
    name
    email
  }
}
```

**REST:**
```bash
GET /admin-query/{id}/user-by-id?id=example
```

**gRPC:**
```protobuf
// Call the UserById method
client.UserById(AdminQueryUserByIdRequest {
  id: example
})
```

### advancedSearch

Search with multiple format-specific customizations

**Type:** `[]Post`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| query | `string` | Yes | - |  |
| filters | `string` | No | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  advancedSearch(query: "example", filters: "example") {
    id
    title
    content
  }
}
```

**REST:**
```bash
GET /admin-query/{id}/advanced-search?query=example&filters=example
```

**gRPC:**
```protobuf
// Call the AdvancedSearch method
client.AdvancedSearch(AdminQueryAdvancedSearchRequest {
  query: example
  filters: example
})
```

