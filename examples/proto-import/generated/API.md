# example API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### User

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `name` | `string` | No |  |
| `email` | `string` | No |  |
| `age` | `int32` | No |  |
| `status` | `UserStatus` | No |  |
| `created_at` | `timestamp` | No |  |
| `tags` | `[]string` | No |  |
| `metadata` | `map<string, string>` | No |  |


### CreateUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | No |  |
| `email` | `string` | No |  |
| `age` | `int32` | No |  |


### CreateUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |


### GetUserRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |


### GetUserResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |


### ListUsersRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `page_size` | `int32` | No |  |
| `page_token` | `string` | No |  |


### ListUsersResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `users` | `[]User` | No |  |
| `next_page_token` | `string` | No |  |


### StreamUsersRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `filter` | `string` | No |  |


### StreamUsersResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |


## Enums

### UserStatus

| Value | Number | Description |
|-------|--------|-------------|
| `USER_STATUS_UNSPECIFIED` | 0 |  |
| `USER_STATUS_ACTIVE` | 1 |  |
| `USER_STATUS_INACTIVE` | 2 |  |
| `USER_STATUS_SUSPENDED` | 3 |  |


## Services

### UserService

#### Methods

##### CreateUser

**Request:** `CreateUserRequest`

**Response:** `CreateUserResponse`

##### GetUser

**Request:** `GetUserRequest`

**Response:** `GetUserResponse`

##### ListUsers

**Request:** `ListUsersRequest`

**Response:** `ListUsersResponse`

##### StreamUsers (server streaming)

**Request:** `StreamUsersRequest`

**Response:** `StreamUsersResponse`


