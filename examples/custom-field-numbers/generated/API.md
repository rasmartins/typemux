# api API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### User

Example demonstrating custom protobuf field numbers

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `age` | `int32` | No |  |
| `createdAt` | `timestamp` | No | This field will auto-assign number 11 (next after 10) |


### Product

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Custom sparse numbering (e.g., for backward compatibility) |
| `name` | `string` | No |  |
| `price` | `float64` | No | Reserved space for future fields (3-9) |
| `description` | `string` | No |  |
| `category` | `string` | No | Auto-assigned fields continue from 12 |
| `inStock` | `bool` | No |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |


## Enums

### Priority

| Value | Number | Description |
|-------|--------|-------------|
| `LOW` | 1 |  |
| `MEDIUM` | 2 |  |
| `HIGH` | 3 |  |
| `URGENT` | 10 |  |


## Services

### UserService

#### Methods

##### GetUser

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`


