# Schema Validation Tests

This directory contains validation tests for the IDL code generator output.

## Setup

Install dependencies using uv:

```bash
uv sync
```

## Running Tests

Run all validation tests:

```bash
cd validate
uv run pytest
```

Run with verbose output:

```bash
uv run pytest -v
```

Run specific test class:

```bash
uv run pytest tests/test_schemas.py::TestProtobufValidation
uv run pytest tests/test_schemas.py::TestGraphQLValidation
uv run pytest tests/test_schemas.py::TestOpenAPIValidation
```

## CI Integration

Add this to your CI pipeline:

```bash
# Generate schemas
go run main.go -input example.schema -output ./final_output

# Run validation tests
cd validate && uv run pytest
```

## What's Tested

- **Protobuf**: Validates using `protoc` compiler
- **GraphQL**: Validates using `graphql-core` library
- **OpenAPI**: Validates using `openapi-spec-validator` library

## Requirements

- Python >= 3.11
- protoc (Protocol Buffers compiler)
- uv (Python package manager)
