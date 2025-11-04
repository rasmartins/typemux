# Query

Query type demonstrating various field argument patterns

## Format Representations

### GraphQL

```graphql
type Query {
  allPosts: [Post]
  featuredPosts: [Post]
}```

### OpenAPI

```yaml
Query:
  type: object
  properties:
    allPosts:
      type: array
      items:
        $ref: '#/components/schemas/Post'
    featuredPosts:
      type: array
      items:
        $ref: '#/components/schemas/Post'
```

### Protobuf

```protobuf
message Query {
  repeated Post allPosts = 1;
  repeated Post featuredPosts = 2;
}```

## Fields

### user

Get a single user by ID
Simple required argument

**Type:** `User`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| id | `string` | Yes | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  user(id: "example") {
    id
    name
    email
  }
}
```

**REST:**
```bash
GET /user?id=example
```

**gRPC:**
```protobuf
// Call the User method
client.User(QueryUserRequest {
  id: example
})
```

### users

Get multiple users with optional pagination
Multiple arguments with defaults

**Type:** `[]User`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| limit | `int32` | No | 10 |  |
| offset | `int32` | No | 0 |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  users(limit: 10, offset: 0) {
    id
    name
    email
  }
}
```

**REST:**
```bash
GET /users?limit=10&offset=0
```

**gRPC:**
```protobuf
// Call the Users method
client.Users(QueryUsersRequest {
  limit: 10
  offset: 0
})
```

### searchUsers

Search users with validation
Argument with validation constraints

**Type:** `[]User`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| query | `string` | Yes | - |  |
| limit | `int32` | No | 20 |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  searchUsers(query: "example", limit: 20) {
    id
    name
    email
  }
}
```

**REST:**
```bash
GET /search-users?query=example&limit=20
```

**gRPC:**
```protobuf
// Call the SearchUsers method
client.SearchUsers(QuerySearchUsersRequest {
  query: example
  limit: 20
})
```

### findUser

Get user by username or email
Optional arguments - at least one should be provided

**Type:** `User`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| username | `string` | No | - |  |
| email | `string` | No | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  findUser(username: "example", email: "example") {
    id
    name
    email
  }
}
```

**REST:**
```bash
GET /find-user?username=example&email=example
```

**gRPC:**
```protobuf
// Call the FindUser method
client.FindUser(QueryFindUserRequest {
  username: example
  email: example
})
```

### posts

Get posts with complex filtering
Mix of required, optional, and filter objects

**Type:** `[]Post`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| authorId | `string` | No | - |  |
| published | `bool` | No | true |  |
| limit | `int32` | No | 10 |  |
| offset | `int32` | No | 0 |  |
| sortBy | `string` | No | createdAt |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  posts(authorId: "example", published: true, limit: 10, offset: 0, sortBy: createdAt) {
    id
    title
    content
  }
}
```

**REST:**
```bash
GET /posts?authorId=example&published=true&limit=10&offset=0&sortBy=createdAt
```

**gRPC:**
```protobuf
// Call the Posts method
client.Posts(QueryPostsRequest {
  authorId: example
  published: true
  limit: 10
  offset: 0
  sortBy: createdAt
})
```

### searchPosts

Advanced search with complex filter

**Type:** `PostSearchResults`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| query | `string` | Yes | - |  |
| filter | `PostFilter` | No | - |  |
| page | `int32` | No | 1 |  |
| pageSize | `int32` | No | 10 |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  searchPosts(query: "example", filter: "value", page: 1, pageSize: 10) {
    posts
    metadata
  }
}
```

**REST:**
```bash
GET /search-posts?query=example&filter=value&page=1&pageSize=10
```

**gRPC:**
```protobuf
// Call the SearchPosts method
client.SearchPosts(QuerySearchPostsRequest {
  query: example
  filter: value
  page: 1
  pageSize: 10
})
```

### post

Get a specific post

**Type:** `Post`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| id | `string` | Yes | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  post(id: "example") {
    id
    title
    content
  }
}
```

**REST:**
```bash
GET /post?id=example
```

**gRPC:**
```protobuf
// Call the Post method
client.Post(QueryPostRequest {
  id: example
})
```

### comments

Get comments for a post with pagination

**Type:** `[]Comment`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| postId | `string` | Yes | - |  |
| limit | `int32` | No | 10 |  |
| offset | `int32` | No | 0 |  |
| sortOrder | `string` | No | desc |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  comments(postId: "example", limit: 10, offset: 0, sortOrder: desc) {
    id
    postId
    authorId
  }
}
```

**REST:**
```bash
GET /comments?postId=example&limit=10&offset=0&sortOrder=desc
```

**gRPC:**
```protobuf
// Call the Comments method
client.Comments(QueryCommentsRequest {
  postId: example
  limit: 10
  offset: 0
  sortOrder: desc
})
```

### allPosts

Field without arguments (traditional style)

**Type:** `[]Post`

### featuredPosts

Get featured posts (no arguments)

**Type:** `[]Post`

