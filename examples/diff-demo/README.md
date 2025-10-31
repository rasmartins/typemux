# Schema Diff Demo

This directory contains example schemas demonstrating TypeMUX's breaking change detection feature.

## Overview

The `typemux diff` command compares two schema versions and detects breaking changes across all three protocols (Protobuf, GraphQL, OpenAPI).

## Files

- `base.typemux` - Original schema (v1.0.0)
- `modified.typemux` - Modified schema with various changes

## Running the Demo

```bash
# Full report
typemux diff -base examples/diff-demo/base.typemux -head examples/diff-demo/modified.typemux

# Compact summary
typemux diff -base examples/diff-demo/base.typemux -head examples/diff-demo/modified.typemux -compact

# Exit with error code if breaking changes detected (useful for CI)
typemux diff -base examples/diff-demo/base.typemux -head examples/diff-demo/modified.typemux -exit-on-breaking
```

## Changes Demonstrated

### Breaking Changes (8)

1. **Field Removed** - `User.name` removed (replaced with `fullName`)
2. **Field Type Changed** - `User.age` changed from `int32` to `int64`
3. **Enum Value Removed** - `UserRole.GUEST` removed
4. **Method Removed** - `UserService.DeleteUser` removed
5. **Required Field Added** (3x) - New required fields in request types

### Non-Breaking Changes (6)

1. **Optional Field Added** - `User.createdAt`, `Product.description`
2. **New Type Added** - `UserProfile`, `UpdateUserRequest`
3. **Enum Value Added** - `UserRole.MODERATOR`
4. **Method Added** - `UserService.UpdateUser`

## Use Cases

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Check for breaking changes
  run: |
    typemux diff \
      -base main-branch-schema.typemux \
      -head feature-branch-schema.typemux \
      -exit-on-breaking
```

### Semver Automation

The tool recommends version bump strategy:
- **MAJOR** - Breaking changes detected
- **MINOR** - New features or dangerous changes
- **PATCH** - Safe changes only

### Multi-Protocol Analysis

Changes are analyzed per-protocol:
- **Protobuf** - Field number changes, reserved fields
- **GraphQL** - Type changes, field additions/removals
- **OpenAPI** - Required field changes, endpoint modifications

## Best Practices

1. **Run diff before releases** to understand impact
2. **Document migration paths** for breaking changes
3. **Use deprecation** instead of immediate removal
4. **Add integration tests** for critical changes
5. **Version your APIs** appropriately based on recommendations
