#!/bin/bash
# Quick pre-commit checks (faster than full CI)
# Run this for quick validation before full CI run

set -e

echo "=========================================="
echo "Running Quick Pre-Commit Checks"
echo "=========================================="
echo ""

# 1. Format check
echo "1/7 Checking code formatting..."
if [ -n "$(gofmt -l .)" ]; then
    echo "❌ The following files are not formatted:"
    gofmt -l .
    echo ""
    echo "Run: gofmt -w ."
    exit 1
fi
echo "✅ Code formatting OK"
echo ""

# 2. go mod tidy check
echo "2/7 Checking go.mod and go.sum..."
cp go.mod go.mod.bak
cp go.sum go.sum.bak
go mod tidy
if ! diff -q go.mod go.mod.bak > /dev/null || ! diff -q go.sum go.sum.bak > /dev/null; then
    echo "❌ go.mod or go.sum is not tidy"
    echo "Run: go mod tidy"
    mv go.mod.bak go.mod
    mv go.sum.bak go.sum
    exit 1
fi
rm go.mod.bak go.sum.bak
echo "✅ go.mod and go.sum are tidy"
echo ""

# 3. Build check
echo "3/7 Building binaries..."
go build -v ./cmd/typemux > /dev/null
go build -v ./cmd/proto2typemux > /dev/null
go build -v ./cmd/graphql2typemux > /dev/null
go build -v ./cmd/openapi2typemux > /dev/null
echo "✅ All binaries build successfully"
echo ""

# 4. Run tests
echo "4/7 Running tests..."
if ! go test ./... > /dev/null 2>&1; then
    echo "❌ Tests failed"
    go test ./...
    exit 1
fi
echo "✅ All tests passed"
echo ""

# 5. Coverage check
echo "5/7 Checking test coverage..."
go test -coverprofile=coverage.out ./... > /dev/null 2>&1
total_coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$total_coverage < 65" | bc -l) )); then
    echo "❌ Coverage ${total_coverage}% is below 65% threshold"
    rm coverage.out
    exit 1
fi
echo "✅ Coverage ${total_coverage}% meets threshold"
rm coverage.out
echo ""

# 6. Linting
echo "6/7 Running linter..."
if ! golangci-lint run --timeout=5m > /dev/null 2>&1; then
    echo "❌ Linting failed"
    golangci-lint run --timeout=5m
    exit 1
fi
echo "✅ Linting passed"
echo ""

# 7. Annotations check
echo "7/7 Checking annotations.json is up to date..."
if ! ./scripts/check-annotations.sh > /dev/null 2>&1; then
    echo "❌ annotations.json is out of date"
    ./scripts/check-annotations.sh
    exit 1
fi
echo "✅ annotations.json is up to date"
echo ""

echo "=========================================="
echo "✅ All quick checks passed!"
echo ""
echo "For full CI validation, run:"
echo "  ./scripts/pre-commit-ci.sh"
echo "=========================================="
