# api API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### User

User entity representing a system user

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier for the user |
| `name` | `string` | Yes | Full name of the user |
| `email` | `string` | Yes | Email address for contact |
| `age` | `int32` | No | User's age in years |
| `role` | `UserRole` | Yes | Role assigned to the user |
| `isActive` | `bool` | No | Whether the user account is active |
| `createdAt` | `timestamp` | Yes | Timestamp when the user was created |
| `tags` | `[]string` | No | Custom tags for categorization |
| `metadata` | `map<string, string>` | No | Additional metadata key-value pairs |
| `dbVersion` | `int32` | No | Internal database version (excluded from GraphQL and OpenAPI) |
| `passwordHash` | `string` | No | Password hash (only in Protobuf for internal services) |


### Post

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `title` | `string` | Yes |  |
| `content` | `string` | No |  |
| `authorId` | `string` | Yes |  |
| `status` | `Status` | Yes |  |
| `publishedAt` | `timestamp` | No |  |
| `viewCount` | `int64` | No |  |
| `tags` | `[]string` | No |  |


### CreateUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `role` | `UserRole` | Yes |  |


### CreateUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | Yes |  |
| `success` | `bool` | Yes |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |


### ListUsersRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `limit` | `int32` | No |  |
| `offset` | `int32` | No |  |
| `role` | `UserRole` | No |  |


### ListUsersResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `users` | `[]User` | Yes |  |
| `total` | `int32` | Yes |  |


## Enums

### UserRole

User role enumeration
Defines the different roles a user can have in the system

| Value | Number | Description |
|-------|--------|-------------|
| `ADMIN` | 10 | Administrator with full access |
| `USER` | 20 | Regular user with limited access |
| `GUEST` | 30 | Guest user with read-only access |


### Status

Status enumeration for various entities

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 1 |  |
| `INACTIVE` | 2 |  |
| `PENDING` | 3 |  |


## Services

### UserService

User service for managing users

#### Methods

##### CreateUser

Create a new user

**Request:** `CreateUserRequest`

**Response:** `CreateUserResponse`

**HTTP:** `POST /api/v1/users`

##### GetUser

Get a user by ID

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

**HTTP:** `GET /api/v1/users/{id}`

##### ListUsers

List all users with pagination

**Request:** `ListUsersRequest`

**Response:** `ListUsersResponse`

**HTTP:** `GET /api/v1/users`

##### DeleteUser

Delete a user

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

**HTTP:** `DELETE /api/v1/users/{id}`


### PostService

Post service for managing blog posts

#### Methods

##### CreatePost

Create a new post

**Request:** `Post`

**Response:** `Post`

**HTTP:** `POST /api/v1/posts`

##### GetPost

Get a post by ID

**Request:** `GetUserRequest`

**Response:** `Post`

**HTTP:** `GET /api/v1/posts/{id}`


