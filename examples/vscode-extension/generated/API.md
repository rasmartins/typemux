# api API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### User

User entity with custom field numbers

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `age` | `int32` | No |  |
| `status` | `Status` | Yes |  |
| `isActive` | `bool` | No |  |
| `createdAt` | `timestamp` | Yes |  |
| `tags` | `[]string` | No |  |
| `metadata` | `map<string, string>` | No |  |
| `dbVersion` | `int32` | No | Internal field excluded from GraphQL and OpenAPI |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |
| `success` | `bool` | Yes |  |


## Enums

### Status

Test file for TypeMUX VS Code extension
This demonstrates syntax highlighting and snippets

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 1 |  |
| `INACTIVE` | 2 |  |
| `PENDING` | 3 |  |


## Services

### UserService

User management service

#### Methods

##### GetUser

Get a user by ID

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

**HTTP:** `GET /api/v1/users/{id}`


