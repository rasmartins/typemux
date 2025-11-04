# User

Field Arguments Example
This example demonstrates the new field-level parameterized query feature
similar to GraphQL field arguments
User entity

## Format Representations

### GraphQL

```graphql
type User {
  id: string!
  name: string!
  email: string!
  username: string!
  age: int32
  isActive: bool
}```

### OpenAPI

```yaml
User:
  type: object
  properties:
    id:
      type: string
    name:
      type: string
    email:
      type: string
    username:
      type: string
    age:
      type: integer
    isActive:
      type: boolean
```

### Protobuf

```protobuf
message User {
  string id = 1;
  string name = 2;
  string email = 3;
  string username = 4;
  optional int32 age = 5;
  optional bool isActive = 6;
}```

## Fields

### id

**Type:** `string` (required)

### name

**Type:** `string` (required)

### email

**Type:** `string` (required)

### username

**Type:** `string` (required)

### age

**Type:** `int32`

### isActive

**Type:** `bool`

