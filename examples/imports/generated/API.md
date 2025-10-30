# api API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### User

User service types

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `address` | `Address` | No |  |
| `status` | `Status` | Yes |  |


### CreateUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `address` | `Address` | Yes |  |


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


### Address

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `street` | `string` | Yes |  |
| `city` | `string` | Yes |  |
| `country` | `string` | Yes |  |
| `zipCode` | `string` | No |  |


## Enums

### Status

Common types shared across services

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 1 |  |
| `INACTIVE` | 2 |  |
| `PENDING` | 3 |  |


## Services

### UserService

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


