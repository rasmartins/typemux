# com.example.api API Documentation

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
| `status` | `Status` | No |  |
| `createdAt` | `timestamp` | No |  |


### Product

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `name` | `string` | No |  |
| `price` | `float64` | No |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | No |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |
| `success` | `bool` | No |  |


### CreateProductRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | No |  |
| `price` | `float64` | No |  |


### CreateProductResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `product` | `Product` | No |  |


## Enums

### Status

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 1 |  |
| `INACTIVE` | 2 |  |
| `DELETED` | 3 |  |


## Services

### UserService

#### Methods

##### GetUser

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

##### CreateProduct

**Request:** `CreateProductRequest`

**Response:** `CreateProductResponse`


