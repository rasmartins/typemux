# chat API Documentation

## Table of Contents

- [Types](#types)
- [Services](#services)

## Types

### Message

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | No |  |
| `content` | `string` | No |  |
| `sender` | `string` | No |  |
| `timestamp` | `timestamp` | No |  |


### MessageRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | `string` | No |  |
| `sender` | `string` | No |  |


### MessageQuery

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `messageId` | `string` | No |  |


### Empty


## Services

### ChatService

Chat service with queries, mutations, and subscriptions

#### Methods

##### GetMessage

**Request:** `MessageQuery`

**Response:** `Message`

##### ListMessages

**Request:** `Empty`

**Response:** `Message`

##### SendMessage

**Request:** `MessageRequest`

**Response:** `Message`

##### DeleteMessage

**Request:** `MessageQuery`

**Response:** `Empty`

##### WatchMessages (server streaming)

**Request:** `Empty`

**Response:** `Message`

##### WatchMessagesBySender (server streaming)

**Request:** `MessageQuery`

**Response:** `Message`


