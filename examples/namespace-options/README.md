# Namespace-Level Annotations Examples

This directory demonstrates how to use namespace-level annotations in TypeMUX to specify package-level configuration for different output formats.

## Overview

Namespace annotations allow you to configure package-level settings such as:
- **Protobuf**: `go_package`, `java_package`, `csharp_namespace`, etc.
- **GraphQL**: Schema-level directives (e.g., Apollo Federation)
- **OpenAPI**: Info section (title, version, description) and extensions

## Two Approaches

### Option 1: Inline Annotations

Annotations are specified directly before the `namespace` declaration in your `.typemux` file.

**Example (`schema.typemux`):**
```typemux
@proto.option(go_package="github.com/rasmartins/typemux/examples/proto")
@proto.option(java_package="com.example.proto")
@proto.option(java_multiple_files=true)
namespace com.example.users
```

**Generate:**
```bash
typemux -input schema.typemux -format protobuf -output ./generated
```

**Output (`schema.proto`):**
```protobuf
syntax = "proto3";

package com.example.users;

option go_package = "github.com/rasmartins/typemux/examples/proto";
option java_package = "com.example.proto";
option java_multiple_files = true;
```

### Option 2: YAML Annotations

Annotations are defined in a separate YAML file, keeping your schema clean and allowing environment-specific configurations.

**Schema (`schema-yaml.typemux`):**
```typemux
namespace com.example.users

type User {
  id: string = 1
  email: string = 2
}
```

**Annotations (`annotations.yaml`):**
```yaml
namespaces:
  com.example.users:
    proto:
      options:
        go_package: "github.com/rasmartins/typemux/examples/proto"
        java_package: "com.example.proto"
        java_multiple_files: "true"
        csharp_namespace: "Example.Proto"
        objc_class_prefix: "EXP"

    graphql:
      directive: "@link(url: \"https://specs.apollo.dev/federation/v2.0\")"

    openapi:
      info:
        title: "User Service API"
        version: "1.0.0"
        description: "User management service"
      extensions:
        x-api-id: "user-service"
        x-internal-id: "prod-users-001"
```

**Generate:**
```bash
typemux -input schema-yaml.typemux -annotations annotations.yaml -output ./generated
```

## Format-Specific Examples

### Protobuf Options

All standard Protobuf file-level options are supported:

```typemux
@proto.option(go_package="github.com/example/proto")
@proto.option(java_package="com.example.proto")
@proto.option(java_multiple_files=true)
@proto.option(java_outer_classname="MyProtos")
@proto.option(csharp_namespace="Example.Proto")
@proto.option(objc_class_prefix="EXP")
@proto.option(php_namespace="Example\\Proto")
@proto.option(php_metadata_namespace="Example\\Proto\\Metadata")
@proto.option(ruby_package="Example::Proto")
namespace com.example.myservice
```

### GraphQL Directives

Useful for Apollo Federation and other schema-level directives:

**Inline:**
```typemux
@graphql.directive("@link(url: \"https://specs.apollo.dev/federation/v2.0\", import: [\"@key\", \"@shareable\"])")
namespace com.example.products
```

**YAML:**
```yaml
namespaces:
  com.example.products:
    graphql:
      directive: "@link(url: \"https://specs.apollo.dev/federation/v2.0\")"
```

**Output (`schema.graphql`):**
```graphql
# Generated GraphQL Schema
# Namespace: com.example.products

extend schema @link(url: "https://specs.apollo.dev/federation/v2.0")

scalar JSON
...
```

### OpenAPI Info

Configure the OpenAPI info section:

```yaml
namespaces:
  com.example.api:
    openapi:
      info:
        title: "My Awesome API"
        version: "2.0.0"
        description: "A comprehensive API for managing resources"
      extensions:
        x-api-id: "awesome-api"
        x-audience: "external"
        x-support-contact: "api-support@example.com"
```

**Output (`openapi.yaml`):**
```yaml
openapi: 3.0.0
info:
    title: My Awesome API
    version: 2.0.0
    description: A comprehensive API for managing resources
paths:
  ...
```

## Files in This Directory

- `schema.typemux` - Example with inline Protobuf annotations
- `schema-yaml.typemux` - Clean schema for YAML annotations
- `schema-graphql.typemux` - Example with inline GraphQL directives
- `annotations.yaml` - Comprehensive YAML annotations for all formats

## Use Cases

### 1. Multi-Environment Configurations

Use YAML annotations to maintain different configurations for dev, staging, and production:

```bash
# Development
typemux -input schema.typemux -annotations annotations-dev.yaml -output ./dev

# Production
typemux -input schema.typemux -annotations annotations-prod.yaml -output ./prod
```

### 2. Apollo Federation

Configure GraphQL schemas for Apollo Federation:

```typemux
@graphql.directive("@link(url: \"https://specs.apollo.dev/federation/v2.0\", import: [\"@key\", \"@shareable\", \"@external\"])")
namespace com.example.users
```

### 3. Language-Specific Protobuf Packages

Configure different package names for different languages:

```yaml
namespaces:
  com.example.api:
    proto:
      options:
        go_package: "github.com/myorg/myapp/proto/api"
        java_package: "com.myorg.myapp.proto.api"
        csharp_namespace: "MyOrg.MyApp.Proto.Api"
        php_namespace: "MyOrg\\MyApp\\Proto\\Api"
```

### 4. API Versioning

Maintain version information in OpenAPI:

```yaml
namespaces:
  com.example.api.v2:
    openapi:
      info:
        title: "My API v2"
        version: "2.1.0"
        description: "Version 2 of the API with breaking changes"
```

## Best Practices

1. **Choose the Right Approach**:
   - Use **inline** for simple, static configurations
   - Use **YAML** for complex, environment-specific configurations

2. **Protobuf Package Naming**:
   - Always set `go_package` for Go projects
   - Follow language conventions for each target language

3. **GraphQL Federation**:
   - Include necessary directives at the namespace level
   - Use type-level annotations for entity-specific directives

4. **OpenAPI Metadata**:
   - Provide clear title and description
   - Use semantic versioning for the version field
   - Add extensions (x-*) for custom metadata

5. **Version Control**:
   - Commit YAML annotation files to version control
   - Use separate files for different environments
   - Document the purpose of each configuration

## Generating Output

```bash
# All formats with YAML annotations
typemux -input schema-yaml.typemux -annotations annotations.yaml -output ./generated

# Protobuf only
typemux -input schema.typemux -format protobuf -output ./proto

# GraphQL with directives
typemux -input schema-graphql.typemux -format graphql -output ./graphql

# Multiple annotation files (later files override earlier ones)
typemux -input schema.typemux -annotations base.yaml -annotations overrides.yaml -output ./gen
```

## Learn More

- [TypeMUX Documentation](https://rasmartins.github.io/typemux)
- [Protocol Buffers Options](https://protobuf.dev/reference/protobuf/proto3-spec/#options)
- [Apollo Federation](https://www.apollographql.com/docs/federation/)
- [OpenAPI Specification](https://swagger.io/specification/)
