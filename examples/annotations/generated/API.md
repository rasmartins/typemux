# api API Documentation

## Table of Contents

- [Types](#types)
- [Services](#services)

## Types

### User

Example demonstrating format-specific annotations
This shows how to add Protobuf options, GraphQL directives, and OpenAPI extensions
User type with GraphQL Federation support

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `tags` | `[]string` | No |  |


### Product

Product with OpenAPI extensions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `price` | `float64` | Yes |  |
| `inStock` | `bool` | Yes |  |


### Config

Configuration with retention and nested OpenAPI metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiKey` | `bytes` | No |  |
| `timeout` | `int32` | No |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | Yes |  |


## Services

### UserService

#### Methods

##### GetUser

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

**HTTP:** `GET /api/v1/users/{userId}`


