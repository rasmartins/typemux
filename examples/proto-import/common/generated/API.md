# common API Documentation

## Table of Contents

- [Types](#types)
- [Enums](#enums)

## Types

### Address

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `street` | `string` | No |  |
| `city` | `string` | No |  |
| `state` | `string` | No |  |
| `zip_code` | `string` | No |  |
| `country` | `string` | No |  |


### Metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `labels` | `map<string, string>` | No |  |
| `tags` | `[]string` | No |  |


## Enums

### Status

| Value | Number | Description |
|-------|--------|-------------|
| `STATUS_UNSPECIFIED` | 0 |  |
| `STATUS_ACTIVE` | 1 |  |
| `STATUS_INACTIVE` | 2 |  |


