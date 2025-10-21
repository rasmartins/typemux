# CI/CD Pipeline

This directory contains the Continuous Integration (CI) pipeline configuration for TypeMux.

## Pipeline Overview

The CI pipeline runs automatically on:
- Push to `main` or `develop` branches
- Pull requests targeting `main` or `develop` branches

## CI Jobs

### 1. **Test** (`test`)
Runs tests across multiple platforms and Go versions to ensure compatibility.

**Matrix:**
- **OS**: Ubuntu, macOS, Windows
- **Go versions**: 1.21, 1.22, 1.23

**Steps:**
- Download dependencies
- Verify dependencies
- Run tests with race detector
- Generate coverage report
- Upload coverage to Codecov (Ubuntu + Go 1.23 only)

### 2. **Coverage Check** (`coverage`)
Ensures test coverage meets minimum thresholds.

**Checks:**
- **Total coverage**: Must be ≥ 90%
- **Package-level coverage**: Reports individual package coverage
- **Critical packages**: Lexer, Parser, Generator

### 3. **Lint** (`lint`)
Runs static analysis to ensure code quality.

**Linters:**
- golangci-lint (see `.golangci.yml` for configuration)
- Checks for:
  - Code formatting
  - Unused code
  - Security issues
  - Best practices violations
  - Common mistakes

### 4. **Build** (`build`)
Builds all binaries across multiple platforms.

**Builds:**
- `typemux` - Main CLI tool
- `proto2typemux` - Protobuf importer
- `graphql2typemux` - GraphQL importer
- `openapi2typemux` - OpenAPI importer

**Platforms:**
- Ubuntu, macOS, Windows

### 5. **Validate Examples** (`validate-examples`)
Validates all example `.typemux` files can generate valid output.

**Tests:**
- Finds all `.typemux` files in `examples/`
- Generates all formats (Protobuf, GraphQL, OpenAPI, Markdown)
- Skips circular import examples (expected to fail)
- Reports pass/fail/skip statistics

### 6. **Validate Importers** (`validate-importers`)
Tests that import tools work correctly.

**Tests:**
- Protobuf importer with `.proto` files
- GraphQL importer with `.graphql`/`.gql` files
- OpenAPI importer with `.yaml`/`.yml`/`.json` files

### 7. **Security** (`security`)
Runs security scans to detect vulnerabilities.

**Scans:**
- `gosec` - Security checker for Go code
- `go vet` - Official Go static analyzer
- `govulncheck` - Vulnerability scanner for Go dependencies

### 8. **Format Check** (`format`)
Ensures all code is properly formatted.

**Checks:**
- `gofmt` - Standard Go formatting
- `go mod tidy` - Dependency management

### 9. **Summary** (`summary`)
Aggregates results from all jobs and provides final status.

## Running CI Locally

### Full CI Pipeline
```bash
make ci
```

This runs:
1. Download dependencies
2. Lint code
3. Check formatting
4. Run `go vet`
5. Run all tests
6. Check coverage threshold
7. Build all binaries
8. Validate examples

### Individual Steps

```bash
# Run tests
make test

# Check coverage
make coverage-check

# Run linters
make lint

# Build binaries
make build

# Validate examples
make validate-examples

# Format code
make fmt
```

### Quick Check
```bash
make quick
```
Runs just dependencies, tests, and build (faster than full CI).

## CI Configuration Files

- **`.github/workflows/ci.yml`** - GitHub Actions workflow
- **`.golangci.yml`** - Linter configuration
- **`Makefile`** - Build automation (works locally and in CI)

## Development Tools

Install all development tools:
```bash
make tools
```

This installs:
- `golangci-lint` - Linter
- `gosec` - Security scanner
- `govulncheck` - Vulnerability checker
- `goimports` - Import formatter

## Coverage Requirements

The CI enforces the following coverage requirements:

### Minimum Thresholds
- **Total project coverage**: 90%

### Current Coverage (as of last update)
- **Lexer**: 88.5%
- **Parser**: 70.8%
- **Generator**: 77.4%
- **Docgen**: 84.8%
- **AST**: 81.0%
- **Config**: 95.6%
- **Annotations**: 77.0%

## Troubleshooting

### CI Failures

**Test Failures:**
```bash
# Run tests locally
make test

# Run specific package
go test -v ./internal/parser
```

**Linting Failures:**
```bash
# Run linters locally
make lint

# Fix formatting
make fmt
```

**Coverage Failures:**
```bash
# Check coverage
make coverage-check

# View detailed coverage
make coverage-html
```

**Build Failures:**
```bash
# Build locally
make build

# Clean and rebuild
make clean build
```

### Example Validation Failures

If example validation fails in CI:

1. **Run locally:**
   ```bash
   make validate-examples
   ```

2. **Test specific file:**
   ```bash
   ./bin/typemux -input examples/path/to/file.typemux -format all -output /tmp/test
   ```

3. **Check for syntax errors** in the `.typemux` file

### Importer Failures

If importer validation fails:

1. **Run locally:**
   ```bash
   make validate-importers
   ```

2. **Test specific importer:**
   ```bash
   # Protobuf
   ./bin/proto2typemux -input examples/proto-import/file.proto -output /tmp/test

   # GraphQL
   ./bin/graphql2typemux -input examples/graphql-import/file.graphql -output /tmp/test

   # OpenAPI
   ./bin/openapi2typemux -input examples/openapi-import/file.yaml -output /tmp/test
   ```

## Badges

Add these badges to your README.md:

```markdown
![CI](https://github.com/YOUR_ORG/typemux/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/YOUR_ORG/typemux/branch/main/graph/badge.svg)](https://codecov.io/gh/YOUR_ORG/typemux)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_ORG/typemux)](https://goreportcard.com/report/github.com/YOUR_ORG/typemux)
```

## Future Enhancements

Potential improvements to the CI pipeline:

1. **Performance Testing**: Add benchmark comparisons
2. **Release Automation**: Automatic binary releases on tags
3. **Documentation**: Auto-generate and deploy docs
4. **Docker**: Build and publish Docker images
5. **Integration Tests**: End-to-end testing with real projects
6. **Dependency Updates**: Automated dependency update PRs
