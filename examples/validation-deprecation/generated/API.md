# com.example.userservice API Documentation

## Table of Contents

- [Types](#types)
- [Services](#services)

## Types

### User

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `username` | `string` | No |  |
| `email` | `string` | No |  |
| `age` | `int32` | No |  |
| `fullName` | `string` | No |  |
| `displayName` | `string` | No | ⚠️ **DEPRECATED**: Use fullName instead |
| `website` | `string` | No |  |
| `balance` | `int64` | No |  |
| `createdAt` | `timestamp` | No |  |
| `lastLogin` | `timestamp` | No |  |
| `isActive` | `bool` | No |  |
| `tags` | `[]string` | No |  |
| `legacyEmail` | `string` | No | ⚠️ **DEPRECATED**: Use email field instead |


### UserPreferences

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | No |  |
| `theme` | `string` | No |  |
| `language` | `string` | No |  |
| `notificationsEnabled` | `bool` | No |  |


### UserIdRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |


### ListUsersRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `pageSize` | `int32` | No |  |
| `pageToken` | `string` | No |  |


### ListUsersResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `users` | `[]User` | No |  |
| `nextPageToken` | `string` | No |  |


### EmptyResponse


## Services

### UserService

#### Methods

##### CreateUser

**Request:** `User`

**Response:** `User`

##### GetUser

**Request:** `UserIdRequest`

**Response:** `User`

##### UpdateUser

**Request:** `User`

**Response:** `User`

##### DeleteUser

**Request:** `UserIdRequest`

**Response:** `EmptyResponse`

##### ListUsers

**Request:** `ListUsersRequest`

**Response:** `ListUsersResponse`


