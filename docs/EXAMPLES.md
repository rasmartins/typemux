# TypeMux - Examples

This document shows example outputs for the different schema formats.

## Input IDL Schema

```
enum UserRole {
  ADMIN
  USER
  GUEST
}

type User {
  id: string @required
  name: string @required
  email: string @required
  age: int32
  role: UserRole @required
  isActive: bool @default(true)
  createdAt: timestamp @required
  tags: []string
  metadata: map<string, string>
}

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
  rpc GetUser(GetUserRequest) returns (GetUserResponse)
}
```

## Generated GraphQL Schema

```graphql
enum UserRole {
  ADMIN
  USER
  GUEST
}

type User {
  id: String!
  name: String!
  email: String!
  age: Int
  role: UserRole!
  isActive: Boolean
  createdAt: String!
  tags: [String]
  metadata: JSON
}

type Query {
  getUser(input: GetUserRequest): GetUserResponse
}

type Mutation {
  createUser(input: CreateUserRequest): CreateUserResponse
}
```

**Key Features:**
- Enums are preserved
- Required fields marked with `!`
- Arrays use `[Type]` syntax
- Maps converted to JSON scalar
- Service methods split into Query/Mutation based on naming

## Generated Protobuf Schema

```protobuf
syntax = "proto3";

package api;

import "google/protobuf/timestamp.proto";

enum UserRole {
  USERROLE_UNSPECIFIED = 0;
  ADMIN = 1;
  USER = 2;
  GUEST = 3;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
  int32 age = 4;
  UserRole role = 5;
  bool isActive = 6;
  google.protobuf.Timestamp createdAt = 7;
  repeated string tags = 8;
  map<string, string> metadata = 9;
}

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

**Key Features:**
- Proto3 syntax
- Automatic field numbering
- Enums include UNSPECIFIED value
- Arrays use `repeated` keyword
- Native map support
- Timestamp uses google.protobuf.Timestamp
- Service definitions preserved

## Generated OpenAPI Schema (excerpt)

```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "Generated API",
    "version": "1.0.0"
  },
  "paths": {
    "/userservice/createuser": {
      "post": {
        "summary": "CreateUser operation",
        "operationId": "CreateUser",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/CreateUserRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/CreateUserResponse"
                }
              }
            }
          }
        }
      }
    },
    "/userservice/getuser": {
      "get": {
        "summary": "GetUser operation",
        "operationId": "GetUser",
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/GetUserResponse"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "UserRole": {
        "type": "string",
        "enum": ["ADMIN", "USER", "GUEST"]
      },
      "User": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string"
          },
          "name": {
            "type": "string"
          },
          "email": {
            "type": "string"
          },
          "age": {
            "type": "integer",
            "format": "int32"
          },
          "role": {
            "$ref": "#/components/schemas/UserRole"
          },
          "isActive": {
            "type": "boolean",
            "default": "true"
          },
          "createdAt": {
            "type": "string",
            "format": "date-time"
          },
          "tags": {
            "type": "array",
            "items": {
              "type": "string"
            }
          },
          "metadata": {
            "type": "object"
          }
        },
        "required": ["id", "name", "email", "role", "createdAt"]
      }
    }
  }
}
```

**Key Features:**
- OpenAPI 3.0 compliant
- REST-style paths generated from services
- GET for queries, POST for mutations
- Schema components with $ref references
- Proper type formats (int32, date-time, etc.)
- Required fields array
- Default values preserved

## Type Mappings

| IDL Type | GraphQL | Protobuf | OpenAPI |
|----------|---------|----------|---------|
| string | String | string | string |
| int32 | Int | int32 | integer (format: int32) |
| int64 | Int | int64 | integer (format: int64) |
| float32 | Float | float | number (format: float) |
| float64 | Float | double | number (format: double) |
| bool | Boolean | bool | boolean |
| timestamp | String | google.protobuf.Timestamp | string (format: date-time) |
| bytes | String | bytes | string (format: byte) |
| []T | [T] | repeated T | array with items |
| map<K,V> | JSON | map<K,V> | object |

## Usage Examples

### Generate All Formats

```bash
./typemux -input myschema.typemux -output ./api
```

### Generate Single Format

```bash
# GraphQL only
./typemux -input myschema.typemux -format graphql -output ./api

# Protobuf only
./typemux -input myschema.typemux -format protobuf -output ./api

# OpenAPI only
./typemux -input myschema.typemux -format openapi -output ./api
```

### Using Make

```bash
# Build and run example
make example

# Generate specific format
make graphql
make protobuf
make openapi

# Clean generated files
make clean
```
