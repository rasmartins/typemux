# com.example.users API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)

## Types

### User

User entity for the users service

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes |  |
| `username` | `string` | Yes |  |
| `email` | `string` | Yes |  |
| `status` | `UserStatus` | Yes |  |
| `role` | `UserRole` | Yes |  |
| `createdAt` | `timestamp` | Yes |  |


## Enums

### UserStatus

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 1 |  |
| `INACTIVE` | 2 |  |
| `SUSPENDED` | 3 |  |


### UserRole

| Value | Number | Description |
|-------|--------|-------------|
| `CUSTOMER` | 1 |  |
| `ADMIN` | 2 |  |
| `MODERATOR` | 3 |  |


