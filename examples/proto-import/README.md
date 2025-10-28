# Protobuf Import Example

This example demonstrates how to import existing Protobuf (.proto) files and convert them to TypeMUX IDL format.

## Files

- `example.proto` - Original Protobuf schema
- `example.typemux` - Converted TypeMUX IDL (generated)

## Converting Protobuf to TypeMUX

Use the `proto2typemux` tool to convert Protobuf files:

```bash
proto2typemux --input example.proto --output ./
```

Or from the repository root:

```bash
go run cmd/proto2typemux/main.go --input examples/proto-import/example.proto --output examples/proto-import
```

## Features Demonstrated

### 1. Enums
```protobuf
enum UserStatus {
  USER_STATUS_UNSPECIFIED = 0;
  USER_STATUS_ACTIVE = 1;
  // ...
}
```

Converts to:
```typemux
enum UserStatus {
  USER_STATUS_UNSPECIFIED = 0
  USER_STATUS_ACTIVE = 1
  // ...
}
```

### 2. Messages with Various Field Types
```protobuf
message User {
  string id = 1;
  optional int32 age = 4;
  repeated string tags = 7;
  map<string, string> metadata = 8;
}
```

Converts to:
```typemux
type User {
  id: string = 1
  age: int32 = 4
  tags: []string = 7
  metadata: map<string, string> = 8
}
```

### 3. Services with Streaming
```protobuf
service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc StreamUsers(StreamUsersRequest) returns (stream StreamUsersResponse);
}
```

Converts to:
```typemux
service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
  rpc StreamUsers(StreamUsersRequest) returns (stream StreamUsersResponse)
}
```

### 4. Proto Options
```protobuf
package example;
option go_package = "github.com/example/proto/example";
```

Converts to:
```typemux
namespace example @proto.option(go_package = "github.com/example/proto/example")
```

## Round-Trip Conversion

After converting to TypeMUX, you can generate back to Protobuf:

```bash
# Convert TypeMUX back to Protobuf
typemux --input example.typemux --output ./generated --format proto

# The generated schema.proto will be functionally equivalent to the original
```

## What's Preserved

✅ Package/namespace names
✅ Enum names and values
✅ Message names and field definitions
✅ Field numbers (critical for compatibility)
✅ Service and RPC method definitions
✅ Streaming RPC indicators
✅ Proto options (go_package, etc.)
✅ Map types
✅ Repeated fields (arrays)
✅ Optional fields
✅ Deprecated field markers

## What's Not Preserved

❌ Comments (not parsed from proto files)
❌ Reserved field numbers (not critical for TypeMUX)
❌ Some advanced proto3 features (oneof converted to optional fields)

## Use Cases

1. **Migration**: Convert existing Protobuf schemas to TypeMUX IDL
2. **Multi-format Generation**: Import proto, generate GraphQL and OpenAPI from the same schema
3. **Schema Unification**: Combine proto-based services with new TypeMUX-defined services
4. **Documentation**: Generate multiple formats from a single source of truth
