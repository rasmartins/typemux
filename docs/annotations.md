---
layout: default
title: Annotation Reference
---

# Annotation Reference

This reference documents all built-in TypeMUX annotations. Annotations provide metadata to customize code generation for different output formats (Protobuf, GraphQL, OpenAPI, Go).

> **📝 Note:** This documentation is automatically generated from [`annotations.json`](https://github.com/rasmartins/typemux/blob/main/annotations.json). To see all annotations programmatically:
>
> ```bash
> typemux annotations
> ```

---

## Table of Contents

- [Schema-Level Annotations](#schema-level-annotations)
- [Namespace-Level Annotations](#namespace-level-annotations)
- [Type-Level Annotations](#type-level-annotations)
- [Field-Level Annotations](#field-level-annotations)
- [Method-Level Annotations](#method-level-annotations)

---

## Schema-Level Annotations

These annotations apply to the entire schema file.

### @typemux

Specifies the TypeMUX IDL format version

**Applies to:** `all`


**Parameters:**

- **version** (string) *required*: Version string (e.g., '1.0.0')


**Examples:**

```typemux
@typemux("1.0.0")
```

### @version

Specifies the schema/API version

**Applies to:** `all`


**Parameters:**

- **version** (string) *required*: Version string (e.g., '2.1.0')


**Examples:**

```typemux
@version("2.1.0")
```

---

## Namespace-Level Annotations

These annotations apply to namespace declarations.

### @proto.option

Adds Protobuf file-level or message-level options

**Applies to:** `Protobuf`


**Parameters:**

- **option** (string) *required*: Protobuf option declaration


**Examples:**

```typemux
@proto.option(go_package="github.com/example/api")
```

```typemux
@proto.option([packed = false])
```

### @graphql.directive

Adds GraphQL directives to schema elements

**Applies to:** `GraphQL`


**Parameters:**

- **directive** (string) *required*: GraphQL directive (e.g., @key, @external)


**Examples:**

```typemux
@graphql.directive(@key(fields: "id"))
```

```typemux
@graphql.directive(@external)
```

### @go.package

Overrides the Go package name for generated code

**Applies to:** `Go`


**Parameters:**

- **package** (string) *required*: Go package name


**Examples:**

```typemux
@go.package("mypackage")
```

---

## Type-Level Annotations

These annotations apply to type, enum, and union definitions.

### @proto.option

Adds Protobuf file-level or message-level options

**Applies to:** `Protobuf`


**Parameters:**

- **option** (string) *required*: Protobuf option declaration


**Examples:**

```typemux
@proto.option(go_package="github.com/example/api")
```

```typemux
@proto.option([packed = false])
```

### @graphql.directive

Adds GraphQL directives to schema elements

**Applies to:** `GraphQL`


**Parameters:**

- **directive** (string) *required*: GraphQL directive (e.g., @key, @external)


**Examples:**

```typemux
@graphql.directive(@key(fields: "id"))
```

```typemux
@graphql.directive(@external)
```

### @proto.name

Overrides the Protobuf name for the element

**Applies to:** `Protobuf`


**Parameters:**

- **name** (string) *required*: Protobuf name


**Examples:**

```typemux
@proto.name("UserV2")
```

```typemux
@proto.name("user_id")
```

### @graphql.name

Overrides the GraphQL name for the element

**Applies to:** `GraphQL`


**Parameters:**

- **name** (string) *required*: GraphQL name


**Examples:**

```typemux
@graphql.name("UserAccount")
```

```typemux
@graphql.name("userId")
```

### @openapi.name

Overrides the OpenAPI schema or property name

**Applies to:** `OpenAPI`


**Parameters:**

- **name** (string) *required*: OpenAPI name


**Examples:**

```typemux
@openapi.name("UserProfile")
```

```typemux
@openapi.name("user_id")
```

### @openapi.extension

Adds OpenAPI vendor extensions (x-* fields)

**Applies to:** `OpenAPI`


**Parameters:**

- **extension** (object) *required*: JSON object with vendor extensions


**Examples:**

```typemux
@openapi.extension({"x-internal": true, "x-format": "currency"})
```

### @deprecated

Marks element as deprecated with version information

**Applies to:** `all`


**Parameters:**

- **reason** (string) *required*: Reason for deprecation
- **since** (string) *optional*: Version when deprecated
- **removed** (string) *optional*: Version when it will be removed


**Examples:**

```typemux
@deprecated("Use fullName instead")
```

```typemux
@deprecated("Use email field", since="2.0.0", removed="3.0.0")
```

### @since

Marks when an element was added to the schema

**Applies to:** `all`


**Parameters:**

- **version** (string) *required*: Version when element was added


**Examples:**

```typemux
@since("2.0.0")
```

---

## Field-Level Annotations

These annotations apply to fields within types.

### @graphql.directive

Adds GraphQL directives to schema elements

**Applies to:** `GraphQL`


**Parameters:**

- **directive** (string) *required*: GraphQL directive (e.g., @key, @external)


**Examples:**

```typemux
@graphql.directive(@key(fields: "id"))
```

```typemux
@graphql.directive(@external)
```

### @proto.name

Overrides the Protobuf name for the element

**Applies to:** `Protobuf`


**Parameters:**

- **name** (string) *required*: Protobuf name


**Examples:**

```typemux
@proto.name("UserV2")
```

```typemux
@proto.name("user_id")
```

### @graphql.name

Overrides the GraphQL name for the element

**Applies to:** `GraphQL`


**Parameters:**

- **name** (string) *required*: GraphQL name


**Examples:**

```typemux
@graphql.name("UserAccount")
```

```typemux
@graphql.name("userId")
```

### @openapi.name

Overrides the OpenAPI schema or property name

**Applies to:** `OpenAPI`


**Parameters:**

- **name** (string) *required*: OpenAPI name


**Examples:**

```typemux
@openapi.name("UserProfile")
```

```typemux
@openapi.name("user_id")
```

### @openapi.extension

Adds OpenAPI vendor extensions (x-* fields)

**Applies to:** `OpenAPI`


**Parameters:**

- **extension** (object) *required*: JSON object with vendor extensions


**Examples:**

```typemux
@openapi.extension({"x-internal": true, "x-format": "currency"})
```

### @required

Marks a field as required/non-nullable

**Applies to:** `all`


**Examples:**

```typemux
id: string @required
```

### @default

Sets a default value for the field

**Applies to:** `all`


**Parameters:**

- **value** (any) *required*: Default value (string, number, or boolean)


**Examples:**

```typemux
age: int32 @default(0)
```

```typemux
active: bool @default(true)
```

```typemux
status: string @default("pending")
```

### @exclude

Excludes field from specific output formats

**Applies to:** `all`


**Parameters:**

- **formats** (list) *required*: Comma-separated list of formats to exclude from


**Examples:**

```typemux
internal: string @exclude(graphql,openapi)
```

```typemux
debug: string @exclude(proto)
```

### @only

Includes field only in specific output formats

**Applies to:** `all`


**Parameters:**

- **formats** (list) *required*: Comma-separated list of formats to include in


**Examples:**

```typemux
protoField: string @only(proto)
```

```typemux
graphqlField: string @only(graphql)
```

### @deprecated

Marks element as deprecated with version information

**Applies to:** `all`


**Parameters:**

- **reason** (string) *required*: Reason for deprecation
- **since** (string) *optional*: Version when deprecated
- **removed** (string) *optional*: Version when it will be removed


**Examples:**

```typemux
@deprecated("Use fullName instead")
```

```typemux
@deprecated("Use email field", since="2.0.0", removed="3.0.0")
```

### @since

Marks when an element was added to the schema

**Applies to:** `all`


**Parameters:**

- **version** (string) *required*: Version when element was added


**Examples:**

```typemux
@since("2.0.0")
```

### @validate

Defines validation rules for the field

**Applies to:** `all`


**Parameters:**

- **format** (string) *optional*: String format (email, uuid, uri, etc.)
  - Valid values: `email`, `uuid`, `uri`, `date`, `time`, `datetime`
- **pattern** (string) *optional*: Regular expression pattern
- **minLength** (number) *optional*: Minimum string length
- **maxLength** (number) *optional*: Maximum string length
- **min** (number) *optional*: Minimum numeric value
- **max** (number) *optional*: Maximum numeric value
- **exclusiveMin** (boolean) *optional*: Whether min is exclusive
- **exclusiveMax** (boolean) *optional*: Whether max is exclusive
- **multipleOf** (number) *optional*: Number must be multiple of this value
- **minItems** (number) *optional*: Minimum array length
- **maxItems** (number) *optional*: Maximum array length
- **uniqueItems** (boolean) *optional*: Whether array items must be unique
- **enum** (list) *optional*: List of allowed values


**Examples:**

```typemux
@validate(format="email", maxLength=100)
```

```typemux
@validate(min=0, max=150)
```

```typemux
@validate(pattern="^[A-Z]{3}$")
```

---

## Method-Level Annotations

These annotations apply to service methods (RPC definitions).

### @deprecated

Marks element as deprecated with version information

**Applies to:** `all`


**Parameters:**

- **reason** (string) *required*: Reason for deprecation
- **since** (string) *optional*: Version when deprecated
- **removed** (string) *optional*: Version when it will be removed


**Examples:**

```typemux
@deprecated("Use fullName instead")
```

```typemux
@deprecated("Use email field", since="2.0.0", removed="3.0.0")
```

### @since

Marks when an element was added to the schema

**Applies to:** `all`


**Parameters:**

- **version** (string) *required*: Version when element was added


**Examples:**

```typemux
@since("2.0.0")
```

### @http.method

Specifies the HTTP method for REST API mapping

**Applies to:** `OpenAPI`


**Parameters:**

- **method** (string) *required*: HTTP method
  - Valid values: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`


**Examples:**

```typemux
@http.method(GET)
```

```typemux
@http.method(POST)
```

### @http.path

Specifies the URL path template for REST API mapping

**Applies to:** `OpenAPI`


**Parameters:**

- **path** (string) *required*: URL path template with parameters in {braces}


**Examples:**

```typemux
@http.path("/api/v1/users")
```

```typemux
@http.path("/api/v1/users/{id}")
```

### @http.success

Specifies additional success HTTP status codes beyond 200

**Applies to:** `OpenAPI`


**Parameters:**

- **codes** (list) *required*: Comma-separated list of HTTP status codes


**Examples:**

```typemux
@http.success(201)
```

```typemux
@http.success(201,204)
```

### @http.errors

Specifies expected error HTTP status codes

**Applies to:** `OpenAPI`


**Parameters:**

- **codes** (list) *required*: Comma-separated list of HTTP status codes


**Examples:**

```typemux
@http.errors(404,500)
```

```typemux
@http.errors(400,404,409,500)
```

### @graphql

Specifies the GraphQL operation type

**Applies to:** `GraphQL`


**Parameters:**

- **operation** (string) *required*: GraphQL operation type
  - Valid values: `query`, `mutation`, `subscription`


**Examples:**

```typemux
@graphql(query)
```

```typemux
@graphql(mutation)
```

```typemux
@graphql(subscription)
```

---

---

## Need More Help?

- See the [Tutorial](tutorial) for practical examples
- Check the [Quick Start](quickstart) guide to get started
- View [Examples](examples) for complete use cases
- Browse the full [Reference](reference) documentation

**Generated from:** [`annotations.json`](https://github.com/rasmartins/typemux/blob/main/annotations.json)
**Last updated:** 2025-10-31
