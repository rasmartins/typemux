# graphql API Documentation

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
| `status` | `UserStatus` | No |  |


### UserUpdate

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |
| `updateType` | `string` | No |  |
| `timestamp` | `timestamp` | No |  |


### ChatMessage

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `userId` | `string` | No |  |
| `text` | `string` | No |  |
| `timestamp` | `timestamp` | No |  |


## Enums

### UserStatus

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 0 |  |
| `INACTIVE` | 1 |  |
| `SUSPENDED` | 2 |  |


## Services

### GraphQLService

#### Methods

##### GetUser

**Request:** `GetUserRequest`

**Response:** `User`

##### UserUpdates (server streaming)

**Request:** `UserUpdatesRequest`

**Response:** `UserUpdate`

##### NewMessages (server streaming)

**Request:** `NewMessagesRequest`

**Response:** `ChatMessage`

##### UserStatusChanged (server streaming)

**Request:** `Empty`

**Response:** `User`


