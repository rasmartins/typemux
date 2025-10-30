# com.example.optionals API Documentation

## Table of Contents

- [Types](#types)
- [Services](#services)

## Types

### UserProfile

Example demonstrating optional field syntax
Fields marked with ? are explicitly optional
Fields marked with @required are required
Fields without either are treated as optional by default
User profile with a mix of required and optional fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier (always required) |
| `username` | `string` | Yes | Username (always required) |
| `email` | `string?` | No | Email address (explicitly optional) |
| `displayName` | `string` | No | Display name (optional, has default behavior) |
| `bio` | `string?` | No | Bio text (explicitly optional) |
| `age` | `int32?` | No | Age in years (explicitly optional) |
| `avatarUrl` | `string?` | No | Profile picture URL (optional) |
| `preferences` | `map<string, string>` | No | User preferences (optional map) |
| `tags` | `[]string?` | No | List of tags (explicitly optional array) |
| `createdAt` | `timestamp` | Yes | Created timestamp (required) |
| `lastLoginAt` | `timestamp?` | No | Last login (explicitly optional) |


### Product

Product with pricing information

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `name` | `string` | Yes |  |
| `description` | `string?` | No | Optional description |
| `price` | `float64` | Yes | Price is required |
| `discountPercent` | `float32?` | No | Optional discount percentage |
| `stockQuantity` | `int32?` | No | Optional stock quantity |


### GetProfileRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |


### UpdateProfileRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | `string` | Yes |  |
| `email` | `string?` | No |  |
| `displayName` | `string?` | No |  |
| `bio` | `string?` | No |  |
| `avatarUrl` | `string?` | No |  |


## Services

### UserService

#### Methods

##### GetProfile

Get user profile by ID

**Request:** `GetProfileRequest`

**Response:** `UserProfile`

##### UpdateProfile

Update user profile

**Request:** `UpdateProfileRequest`

**Response:** `UserProfile`


