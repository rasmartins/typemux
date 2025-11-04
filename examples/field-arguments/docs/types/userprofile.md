# UserProfile

Nested type to show field arguments work at any level

## Format Representations

### GraphQL

```graphql
type UserProfile {
  user: User!
  followerCount: int32
}```

### OpenAPI

```yaml
UserProfile:
  type: object
  properties:
    user:
      $ref: '#/components/schemas/User'
    followerCount:
      type: integer
```

### Protobuf

```protobuf
message UserProfile {
  User user = 1;
  optional int32 followerCount = 2;
}```

## Fields

### user

**Type:** `User` (required)

### posts

Posts authored by this user with arguments

**Type:** `[]Post`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| limit | `int32` | No | 5 |  |
| published | `bool` | No | true |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  posts(limit: 5, published: true) {
    id
    title
    content
  }
}
```

**REST:**
```bash
GET /user-profile/{id}/posts?limit=5&published=true
```

**gRPC:**
```protobuf
// Call the Posts method
client.Posts(UserProfilePostsRequest {
  limit: 5
  published: true
})
```

### recentComments

Recent comments with pagination

**Type:** `[]Comment`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| limit | `int32` | No | 10 |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  recentComments(limit: 10) {
    id
    postId
    authorId
  }
}
```

**REST:**
```bash
GET /user-profile/{id}/recent-comments?limit=10
```

**gRPC:**
```protobuf
// Call the RecentComments method
client.RecentComments(UserProfileRecentCommentsRequest {
  limit: 10
})
```

### followerCount

Follower count (no arguments)

**Type:** `int32`

