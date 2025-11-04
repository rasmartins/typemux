# Mutation

Mutation type demonstrating field arguments for mutations

## Format Representations

### GraphQL

```graphql
type Mutation {
}```

### OpenAPI

```yaml
Mutation:
  type: object
  properties:
```

### Protobuf

```protobuf
message Mutation {
}```

## Fields

### createUser

Create a new user

**Type:** `User`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| name | `string` | Yes | - |  |
| email | `string` | Yes | - |  |
| username | `string` | Yes | - |  |
| age | `int32` | No | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  createUser(name: "example", email: "example", username: "example", age: 10) {
    id
    name
    email
  }
}
```

**REST:**
```bash
GET /create-user?name=example&email=example&username=example&age=10
```

**gRPC:**
```protobuf
// Call the CreateUser method
client.CreateUser(MutationCreateUserRequest {
  name: example
  email: example
  username: example
  age: 10
})
```

### updateUser

Update user information

**Type:** `User`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| id | `string` | Yes | - |  |
| name | `string` | No | - |  |
| email | `string` | No | - |  |
| age | `int32` | No | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  updateUser(id: "example", name: "example", email: "example", age: 10) {
    id
    name
    email
  }
}
```

**REST:**
```bash
GET /update-user?id=example&name=example&email=example&age=10
```

**gRPC:**
```protobuf
// Call the UpdateUser method
client.UpdateUser(MutationUpdateUserRequest {
  id: example
  name: example
  email: example
  age: 10
})
```

### deleteUser

Delete a user

**Type:** `bool`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| id | `string` | Yes | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  deleteUser(id: "example") {
  }
}
```

**REST:**
```bash
GET /delete-user?id=example
```

**gRPC:**
```protobuf
// Call the DeleteUser method
client.DeleteUser(MutationDeleteUserRequest {
  id: example
})
```

### createPost

Create a new post

**Type:** `Post`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| title | `string` | Yes | - |  |
| content | `string` | Yes | - |  |
| published | `bool` | No | false |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  createPost(title: "example", content: "example", published: false) {
    id
    title
    content
  }
}
```

**REST:**
```bash
GET /create-post?title=example&content=example&published=false
```

**gRPC:**
```protobuf
// Call the CreatePost method
client.CreatePost(MutationCreatePostRequest {
  title: example
  content: example
  published: false
})
```

### publishPost

Publish a post

**Type:** `Post`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| id | `string` | Yes | - |  |
| publishedAt | `timestamp` | No | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  publishPost(id: "example", publishedAt: "value") {
    id
    title
    content
  }
}
```

**REST:**
```bash
GET /publish-post?id=example&publishedAt=value
```

**gRPC:**
```protobuf
// Call the PublishPost method
client.PublishPost(MutationPublishPostRequest {
  id: example
  publishedAt: value
})
```

### addComment

Add a comment to a post

**Type:** `Comment`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| postId | `string` | Yes | - |  |
| content | `string` | Yes | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  addComment(postId: "example", content: "example") {
    id
    postId
    authorId
  }
}
```

**REST:**
```bash
GET /add-comment?postId=example&content=example
```

**gRPC:**
```protobuf
// Call the AddComment method
client.AddComment(MutationAddCommentRequest {
  postId: example
  content: example
})
```

