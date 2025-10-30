# api API Documentation

## Table of Contents

- [Types](#types)
- [Services](#services)

## Types

### User

Example demonstrating @success and @errors annotations

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |


### CreateUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |


### CreateUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | Yes |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |


### UpdateUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |
| `name` | `string` | No |  |
| `email` | `string` | No |  |


### UpdateUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | Yes |  |


## Services

### UserService

#### Methods

##### CreateUser

Create a new user - returns 201 Created

**Request:** `CreateUserRequest`

**Response:** `CreateUserResponse`

**HTTP:** `POST /api/v1/users`

##### GetUser

Get a user by ID - standard 200 response

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

**HTTP:** `GET /api/v1/users/{id}`

##### UpdateUser

Update a user - can return 200 or 204

**Request:** `UpdateUserRequest`

**Response:** `UpdateUserResponse`

**HTTP:** `PUT /api/v1/users/{id}`


