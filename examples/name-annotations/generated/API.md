# com.example.api API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### User

User type with different names in each format:
- Protobuf: UserV2
- GraphQL: UserAccount
- OpenAPI: UserProfile

This example uses LEADING annotations (before the type keyword)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `username` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `status` | `Status` | Yes |  |
| `createdAt` | `timestamp` | Yes |  |


### Product

Product type with custom Protobuf name for versioning
This example uses TRAILING annotation (after the type name)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `price` | `float64` | Yes |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | Yes |  |
| `success` | `bool` | Yes |  |


### CreateProductRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes |  |
| `price` | `float64` | Yes |  |


### CreateProductResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `product` | `Product` | Yes |  |


## Enums

### Status

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 1 |  |
| `INACTIVE` | 2 |  |
| `DELETED` | 3 |  |


## Services

### UserService

User service demonstrating name annotations

#### Methods

##### GetUser

Get a user by ID

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

**HTTP:** `GET /api/v1/users/{userId}`

##### CreateProduct

Create a new product

**Request:** `CreateProductRequest`

**Response:** `CreateProductResponse`

**HTTP:** `POST /api/v1/products`


