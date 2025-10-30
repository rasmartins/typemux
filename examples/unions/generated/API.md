# api API Documentation

## Table of Contents

- [Types](#types)
- [Unions](#unions)
- [Services](#services)

## Types

### TextMessage

Example demonstrating union/oneOf types

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | `string` | Yes |  |
| `timestamp` | `timestamp` | Yes |  |


### ImageMessage

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `imageUrl` | `string` | Yes |  |
| `thumbnail` | `string` | No |  |
| `timestamp` | `timestamp` | Yes |  |


### VideoMessage

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `videoUrl` | `string` | Yes |  |
| `duration` | `int32` | Yes |  |
| `thumbnail` | `string` | No |  |
| `timestamp` | `timestamp` | Yes |  |


### SendMessageRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `chatId` | `string` | Yes |  |
| `message` | `Message` | Yes |  |


### SendMessageResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `messageId` | `string` | Yes |  |
| `success` | `bool` | Yes |  |


### GetMessageRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `messageId` | `string` | Yes |  |


### GetMessageResponse

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message` | `Message` | Yes |  |


## Unions

### Message

A message can be text, image, or video

**Possible types:**

- `TextMessage`
- `ImageMessage`
- `VideoMessage`


## Services

### MessageService

#### Methods

##### SendMessage

Send a message (text, image, or video)

**Request:** `SendMessageRequest`

**Response:** `SendMessageResponse`

**HTTP:** `POST /api/v1/messages`

##### GetMessage

Get a message by ID

**Request:** `GetMessageRequest`

**Response:** `GetMessageResponse`

**HTTP:** `GET /api/v1/messages/{id}`


