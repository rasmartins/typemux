# Dashboard

Type showing mix of fields with and without arguments

## Format Representations

### GraphQL

```graphql
type Dashboard {
  currentUser: User!
  stats: map
}```

### OpenAPI

```yaml
Dashboard:
  type: object
  properties:
    currentUser:
      $ref: '#/components/schemas/User'
    stats:
      $ref: '#/components/schemas/map'
```

### Protobuf

```protobuf
message Dashboard {
  User currentUser = 1;
  optional map stats = 2;
}```

## Fields

### currentUser

Current user (no arguments)

**Type:** `User` (required)

### notifications

Notifications with pagination

**Type:** `[]string`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| limit | `int32` | No | 20 |  |
| unreadOnly | `bool` | No | false |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  notifications(limit: 20, unreadOnly: false) {
  }
}
```

**REST:**
```bash
GET /dashboard/{id}/notifications?limit=20&unreadOnly=false
```

**gRPC:**
```protobuf
// Call the Notifications method
client.Notifications(DashboardNotificationsRequest {
  limit: 20
  unreadOnly: false
})
```

### activityFeed

Recent activity feed

**Type:** `[]string`

**Arguments:**

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| limit | `int32` | No | 50 |  |
| types | `string` | No | - |  |

#### Usage Examples

**GraphQL:**
```graphql
{
  activityFeed(limit: 50, types: "example") {
  }
}
```

**REST:**
```bash
GET /dashboard/{id}/activity-feed?limit=50&types=example
```

**gRPC:**
```protobuf
// Call the ActivityFeed method
client.ActivityFeed(DashboardActivityFeedRequest {
  limit: 50
  types: example
})
```

### stats

Summary stats (no arguments)

**Type:** `map`

