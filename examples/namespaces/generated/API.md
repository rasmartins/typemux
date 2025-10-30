# com.example.products API Documentation

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


### User

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `productName` | `string` | No |  |
| `ownerId` | `string` | No |  |


### GetProductRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `productId` | `string` | No |  |


### GetProductResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user` | `User` | No |  |
| `success` | `bool` | No |  |


## Enums

### UserStatus

| Value | Number | Description |
|-------|--------|-------------|
| `ACTIVE` | 0 |  |
| `INACTIVE` | 0 |  |
| `SUSPENDED` | 0 |  |


### Status

| Value | Number | Description |
|-------|--------|-------------|
| `AVAILABLE` | 0 |  |
| `OUT_OF_STOCK` | 0 |  |


## Services

### ProductService

#### Methods

##### GetProduct

**Request:** `GetProductRequest`

**Response:** `GetProductResponse`


