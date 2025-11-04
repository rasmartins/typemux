# Cross-Format Usage Guide

This guide shows how to use the same API across different formats.

## Overview

The API is available in three formats, each optimized for different use cases:

### GraphQL

- **Best for:** Web and mobile apps that need flexible data fetching
- **Advantages:** Request exactly the data you need, reduce over-fetching
- **Protocol:** HTTP/JSON
- **Endpoint:** `/graphql`

### REST/OpenAPI

- **Best for:** Traditional web apps, simple integrations
- **Advantages:** Standard HTTP methods, easy to cache, widely supported
- **Protocol:** HTTP/JSON
- **Base URL:** `/api/v1`

### gRPC/Protobuf

- **Best for:** Microservices, high-performance systems
- **Advantages:** Binary protocol, type-safe, efficient, streaming support
- **Protocol:** HTTP/2 with Protocol Buffers

## Field Arguments Across Formats

TypeMUX field arguments map naturally to each format:

| TypeMUX | GraphQL | REST | gRPC |
|---------|---------|------|------|
| `user(id: string)` | Field arguments | Query parameters | Request message fields |
| Required args | `id: String!` | Required query param | Non-optional field |
| Optional args | `limit: Int` | Optional query param | Optional field |
| Default values | `limit: Int = 10` | Default in docs | Default in proto3 |

## Example: Query.user

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

## Example: Mutation.createUser

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

## Example: AdminQuery.userById

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

## Example: UserProfile.posts

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

## Example: Dashboard.notifications

#### Usage Examples

**GraphQL:**
```graphql
{
  notifications(limit: 20, unreadOnly: false) {
  }
}
```

**REST:**
```bash
GET /dashboard/{id}/notifications?limit=20&unreadOnly=false
```

**gRPC:**
```protobuf
// Call the Notifications method
client.Notifications(DashboardNotificationsRequest {
  limit: 20
  unreadOnly: false
})
```

