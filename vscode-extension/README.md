# TypeMUX Schema Language Support

VS Code extension providing language support for TypeMUX IDL schema files (`.typemux`).

## Features

- **Syntax Highlighting**: Full syntax highlighting for TypeMUX schema files
  - Keywords: `enum`, `type`, `service`, `rpc`, `returns`
  - Types: `string`, `int32`, `int64`, `float32`, `float64`, `bool`, `timestamp`, `bytes`
  - Attributes: `@required`, `@default`, `@exclude`, `@only`, `@http.method`, `@graphql`, `@http.path`
  - Format-specific annotations: `@proto.name()`, `@graphql.name()`, `@openapi.name()`
  - Comments: Single-line (`//`) and documentation (`///`)

- **Code Snippets**: Quick insertion of common patterns
  - `enum` - Create enum definition
  - `enumnum` - Create enum with custom numbers
  - `type` - Create type definition
  - `typenum` - Create type with custom field numbers
  - `typenames` - Create type with leading name annotations
  - `service` - Create service definition
  - `rpc` - Create RPC method
  - `field`, `fieldreq`, `fielddef`, `fieldnum`, `fieldleading` - Create fields
  - `array`, `map` - Create array/map fields
  - `doc`, `doct` - Add documentation
  - `http`, `graphql`, `exclude`, `only` - Add annotations
  - `protoname`, `graphqlname`, `openapiname` - Add format-specific name annotations
  - `schema` - Complete schema template

- **Auto-Completion**: Bracket and quote auto-closing
- **Code Folding**: Support for folding type and service definitions
- **Indentation**: Automatic indentation for nested structures

## Examples

### Basic Type Definition

```typemux
/// User entity
type User {
    id: string = 1 @required
    name: string = 2 @required
    email: string = 5 @required
    age: int32 = 10
    createdAt: timestamp
}

enum UserRole {
    ADMIN = 1
    USER = 2
    GUEST = 3
}

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse)
}
```

### Leading Annotations (NEW)

```typemux
/// Type with different names in each format
@proto.name("UserV2")
@graphql.name("UserAccount")
@openapi.name("UserProfile")
type User {
    id: string = 1 @required
    @required
    username: string = 2
    email: string = 3 @required
}
```

## Installation

### From Source

1. Navigate to the extension directory:
   ```bash
   cd vscode-extension
   ```

2. Install using VS Code CLI:
   ```bash
   code --install-extension typemux-schema-0.1.0.vsix
   ```

3. Reload VS Code

## Usage

1. Create a new file with `.typemux` extension
2. Start typing and use snippets:
   - Type `type` and press Tab for a type definition
   - Type `service` and press Tab for a service definition
   - Type `fieldnum` and press Tab for a field with custom number

## Requirements

- Visual Studio Code 1.70.0 or higher

## Release Notes

### 0.4.0

- Added support for leading annotations (`@proto.name()`, `@graphql.name()`, `@openapi.name()`)
- New snippets: `typenames`, `protoname`, `graphqlname`, `openapiname`, `fieldleading`
- Enhanced syntax highlighting for format-specific name annotations
- Updated documentation with leading annotation examples

### 0.1.0

Initial release:
- Syntax highlighting for TypeMUX schema files
- 20+ code snippets
- Auto-completion and bracket matching
- Code folding support

## License

MIT
