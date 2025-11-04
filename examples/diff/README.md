# Breaking Change Detection Examples

This directory contains examples demonstrating TypeMUX's breaking change detection feature.

## Overview

The `typemux diff` command compares two schema versions and detects:
- **Breaking changes** that will break existing clients
- **Dangerous changes** that might cause subtle bugs
- **Safe changes** that are backwards compatible

## Example Schemas

### schema-v1.typemux
Base schema with:
- Types with various fields
- Fields with arguments (GraphQL-style parameterized queries)
- Mutations with required and optional arguments

### schema-v2-breaking.typemux
Version with **breaking changes**:
- Removed field (`User.age`)
- Removed field argument (`Query.user(id)`)
- Changed argument type (`Query.users(limit)` from int32 to string)
- Added required argument (`Query.posts(category)`)
- Made optional argument required (`Mutation.createUser(age)`)
- Changed method signature (`Mutation.updateUser` - removed `name` and `email` args)

### schema-v2-safe.typemux
Version with **only safe changes**:
- Added new optional fields
- Added new optional field arguments
- Added new mutations
- All backwards compatible

## Usage

### Compare schemas and detect breaking changes:
```bash
typemux diff -base schema-v1.typemux -head schema-v2-breaking.typemux
```

**Output:**
```
Summary:
  Total changes: 9
  ❌ Breaking:     8
  ✅ Dangerous:    0
  ✨ Non-breaking: 1

Recommendation: ⚠️  MAJOR version bump required
```

### Compare schemas with safe changes:
```bash
typemux diff -base schema-v1.typemux -head schema-v2-safe.typemux
```

**Output:**
```
Summary:
  Total changes: 5
  ✅ Breaking:     0
  ✅ Dangerous:    0
  ✨ Non-breaking: 5

Recommendation: ✨ MINOR version bump recommended
```

### Compact output:
```bash
typemux diff -base schema-v1.typemux -head schema-v2-breaking.typemux -compact
```

### Exit with error code on breaking changes (for CI/CD):
```bash
typemux diff -base schema-v1.typemux -head schema-v2-breaking.typemux -exit-on-breaking
```

## Breaking Change Detection for Field Arguments

TypeMUX detects the following breaking changes for field arguments:

| Change | Breaking? | Example |
|--------|-----------|---------|
| Argument removed | ✅ Breaking | `user(id: string)` → `user()` |
| Argument type changed | ✅ Breaking | `users(limit: int32)` → `users(limit: string)` |
| Optional → Required | ✅ Breaking | `posts(published: bool)` → `posts(published: bool @required)` |
| Required argument added | ✅ Breaking | `posts()` → `posts(category: string @required)` |
| Optional argument added | ✅ Safe | `users()` → `users(offset: int32)` |
| Required → Optional | ✅ Safe | `user(id: string @required)` → `user(id: string)` |

## CI/CD Integration

Add schema validation to your CI pipeline:

```yaml
# .github/workflows/schema-check.yml
name: Schema Validation
on: [pull_request]

jobs:
  check-breaking-changes:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
        with:
          fetch-depth: 0  # Fetch all history

      - name: Get base schema
        run: git show main:schema.typemux > schema-base.typemux

      - name: Check for breaking changes
        run: |
          typemux diff \
            -base schema-base.typemux \
            -head schema.typemux \
            -exit-on-breaking
```

## Across All Formats

TypeMUX detects breaking changes for:
- **GraphQL**: Field arguments, types, directives
- **OpenAPI/REST**: Query parameters, request/response schemas, endpoints
- **Protobuf**: Field numbers, messages, service methods, enum values

This ensures your API stays compatible across all three formats simultaneously.
