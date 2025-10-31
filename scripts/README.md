# Pre-Commit Validation Scripts

This directory contains scripts to validate changes locally before committing.

## Scripts

### `quick-check.sh` (Recommended for frequent use)

Fast validation that runs the most important checks:
- Code formatting (`gofmt`)
- `go.mod` tidiness
- Build verification
- Unit tests
- Coverage threshold
- Linting

**Usage:**
```bash
./scripts/quick-check.sh
```

**Time:** ~30-60 seconds

### `pre-commit-ci.sh` (Run before pushing)

Runs the **complete CI pipeline locally** using [act](https://github.com/nektos/act).
This simulates the exact GitHub Actions workflow in Docker containers.

**Usage:**
```bash
./scripts/pre-commit-ci.sh
```

**Time:** ~5-10 minutes (depends on Docker cache)

**Requirements:**
- Docker must be running
- `act` must be installed

## Recommended Workflow

1. **During development:** Run `quick-check.sh` frequently
2. **Before committing:** Run `quick-check.sh`
3. **Before pushing:** Run `pre-commit-ci.sh` to ensure CI will pass

## Installing act

If `act` is not installed:

```bash
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

Or on macOS:
```bash
brew install act
```

## CI Pipeline Jobs

The CI pipeline includes:
- **Test** - All tests across OS/Go version matrix
- **Coverage** - Verify ≥65% coverage
- **Lint** - `golangci-lint` checks
- **Build** - Build all binaries
- **Format** - Code formatting checks
- **Validate Examples** - Test all `.typemux` examples
- **Validate Importers** - Test importer tools
- **Security** - `gosec`, `go vet`, `govulncheck`

## Tips

- Both scripts will exit with error code 1 if any check fails
- Use `quick-check.sh` in a git pre-commit hook for automatic validation
- The full CI run with `act` requires Docker and may download images on first run
