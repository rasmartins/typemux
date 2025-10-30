# PetStoreAPI API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)
- [Services](#services)

## Types

### Error

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | `string` | No |  |
| `message` | `string` | No |  |
| `details` | `string` | No |  |


### Pet

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `name` | `string` | No |  |
| `status` | `string` | No |  |
| `species` | `string` | No |  |
| `breed` | `string` | No |  |
| `age` | `int32` | No |  |
| `tags` | `[]string` | No |  |
| `createdAt` | `timestamp` | No |  |


### NewPet

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | No |  |
| `status` | `string` | No |  |
| `species` | `string` | No |  |
| `breed` | `string` | No |  |
| `age` | `int32` | No |  |
| `tags` | `[]string` | No |  |


### UpdatePet

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `age` | `int32` | No |  |
| `tags` | `[]string` | No |  |
| `name` | `string` | No |  |
| `status` | `string` | No |  |
| `breed` | `string` | No |  |


### PetList

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `total` | `int32` | No |  |
| `nextOffset` | `int32` | No |  |
| `pets` | `[]Pet` | No |  |


## Enums

### PetStatus

| Value | Number | Description |
|-------|--------|-------------|
| `AVAILABLE` | 0 |  |
| `PENDING` | 1 |  |
| `SOLD` | 2 |  |


## Services

### PetStoreAPIService

#### Methods

##### ListPets

**Request:** `ListPetsRequest`

**Response:** `PetList`

##### CreatePet

**Request:** `CreatePetRequest`

**Response:** `Pet`

##### GetPetById

**Request:** `GetPetByIdRequest`

**Response:** `Pet`

##### UpdatePet

**Request:** `UpdatePetRequest`

**Response:** `Pet`

##### DeletePet

**Request:** `DeletePetRequest`

**Response:** `Empty`


