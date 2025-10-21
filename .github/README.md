# CI/CD Pipeline

This directory contains the GitHub Actions CI configuration for TypeMux.

## What Gets Tested

The CI pipeline runs on every push to `main` or `develop` and on all pull requests:

- **Tests**: Cross-platform (Linux, macOS, Windows) on Go 1.21, 1.22, 1.23
- **Coverage**: Enforces 90% minimum coverage threshold
- **Linting**: golangci-lint with 15+ linters (see `.golangci.yml`)
- **Security**: gosec, go vet, govulncheck
- **Build**: All 4 binaries (`typemux`, `proto2typemux`, `graphql2typemux`, `openapi2typemux`)
- **Examples**: Validates all `.typemux` examples generate correctly
- **Importers**: Tests all 3 import tools work

## Run Locally

```bash
# Full CI pipeline
make ci

# Quick check (tests + build)
make quick

# Individual checks
make test
make lint
make coverage-check
make validate-examples
```

## Install Dev Tools

```bash
make tools
```

This installs: golangci-lint, gosec, govulncheck, goimports

## Files

- **`workflows/ci.yml`** - GitHub Actions workflow
- **`.golangci.yml`** - Linter configuration (in project root)
- **`Makefile`** - Build targets (in project root)
