# user API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### User

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `username` | `string` | No |  |
| `email` | `string` | No |  |
| `role` | `UserRole` | No |  |
| `createdAt` | `timestamp` | No |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |


### CreateUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | `string` | No |  |
| `email` | `string` | No |  |
| `role` | `UserRole` | No |  |


### UpdateUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `username` | `string` | No |  |
| `email` | `string` | No |  |


### ListUsersRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `limit` | `int32` | No |  |
| `offset` | `int32` | No |  |


### UserListResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `users` | `[]User` | No |  |
| `total` | `int32` | No |  |


### Empty


## Enums

### UserRole

| Value | Number | Description |
|-------|--------|-------------|
| `ADMIN` | 0 |  |
| `USER` | 0 |  |
| `GUEST` | 0 |  |


## Services

### UserService

#### Methods

##### GetUser

**Request:** `GetUserRequest`

**Response:** `User`

**HTTP:** `GET /api/v1/users/{id}`

##### ListUsers

**Request:** `ListUsersRequest`

**Response:** `UserListResponse`

**HTTP:** `GET /api/v1/users`

##### CreateUser

**Request:** `CreateUserRequest`

**Response:** `User`

**HTTP:** `POST /api/v1/users`

##### UpdateUser

**Request:** `UpdateUserRequest`

**Response:** `User`

**HTTP:** `PUT /api/v1/users/{id}`

##### DeleteUser

**Request:** `GetUserRequest`

**Response:** `Empty`

**HTTP:** `DELETE /api/v1/users/{id}`


