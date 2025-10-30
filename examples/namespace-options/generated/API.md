# com.example.users API Documentation

## Table of Contents

- [Types](#types)
- [Services](#services)

## Types

### User

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `email` | `string` | No |  |
| `username` | `string` | No |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |


## Services

### UserService

#### Methods

##### GetUser

**Request:** `GetUserRequest`

**Response:** `User`


