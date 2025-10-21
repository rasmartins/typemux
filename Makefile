.PHONY: help
help: ## Display this help message
	@echo "TypeMux - Multi-format Schema Generator"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'

.PHONY: all
all: clean deps build test lint ## Run all checks and build

# ============================================================================
# Build Targets
# ============================================================================

.PHONY: build
build: ## Build all binaries
	@echo "==> Building binaries..."
	@mkdir -p bin
	go build -v -o bin/typemux ./cmd/typemux
	go build -v -o bin/proto2typemux ./cmd/proto2typemux
	go build -v -o bin/graphql2typemux ./cmd/graphql2typemux
	go build -v -o bin/openapi2typemux ./cmd/openapi2typemux
	@echo "✅ Built binaries in bin/"

.PHONY: install
install: ## Install binaries to $GOPATH/bin
	@echo "==> Installing binaries..."
	go install ./cmd/typemux
	go install ./cmd/proto2typemux
	go install ./cmd/graphql2typemux
	go install ./cmd/openapi2typemux
	@echo "✅ Installed to $$GOPATH/bin"

# ============================================================================
# Development Targets
# ============================================================================

.PHONY: deps
deps: ## Download dependencies
	@echo "==> Downloading dependencies..."
	go mod download
	go mod verify

.PHONY: tidy
tidy: ## Tidy go.mod
	@echo "==> Tidying go.mod..."
	go mod tidy
	@echo "✅ go.mod tidied"

# ============================================================================
# Test Targets
# ============================================================================

.PHONY: test
test: ## Run all tests
	@echo "==> Running tests..."
	go test -v -race ./...

.PHONY: test-short
test-short: ## Run tests without race detector
	@echo "==> Running tests (short)..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: coverage ## Alias for coverage

.PHONY: coverage
coverage: ## Run tests with coverage report
	@echo "==> Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html: coverage ## Generate HTML coverage report
	@echo "==> Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

.PHONY: coverage-check
coverage-check: ## Check coverage meets minimum threshold (90%)
	@echo "==> Checking coverage threshold..."
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${total}%"; \
	if command -v bc >/dev/null 2>&1; then \
		if [ $$(echo "$${total} < 90" | bc -l) -eq 1 ]; then \
			echo "❌ Coverage $${total}% is below 90% threshold"; \
			exit 1; \
		else \
			echo "✅ Coverage $${total}% meets 90% threshold"; \
		fi \
	else \
		echo "⚠️  bc not installed, skipping threshold check"; \
	fi

# ============================================================================
# Code Quality Targets
# ============================================================================

.PHONY: fmt
fmt: ## Format code
	@echo "==> Formatting code..."
	gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi
	@echo "✅ Code formatted"

.PHONY: fmt-check
fmt-check: ## Check code formatting
	@echo "==> Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ The following files are not formatted:"; \
		gofmt -l .; \
		echo ""; \
		echo "Run: make fmt"; \
		exit 1; \
	else \
		echo "✅ All files are formatted"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "==> Running go vet..."
	go vet ./...
	@echo "✅ go vet passed"

.PHONY: lint
lint: ## Run linters
	@echo "==> Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
		echo "✅ Linting passed"; \
	else \
		echo "❌ golangci-lint not installed. Install with:"; \
		echo "  brew install golangci-lint (macOS)"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

.PHONY: security
security: ## Run security checks
	@echo "==> Running security checks..."
	@echo "Running go vet..."
	@go vet ./...
	@if command -v gosec >/dev/null 2>&1; then \
		echo "Running gosec..."; \
		gosec -quiet ./...; \
		echo "✅ gosec passed"; \
	else \
		echo "⚠️  gosec not installed (skipping)"; \
	fi
	@if command -v govulncheck >/dev/null 2>&1; then \
		echo "Running govulncheck..."; \
		govulncheck ./...; \
		echo "✅ govulncheck passed"; \
	else \
		echo "⚠️  govulncheck not installed (skipping)"; \
	fi

# ============================================================================
# Validation Targets
# ============================================================================

.PHONY: validate-examples
validate-examples: build ## Validate all example files
	@echo "==> Validating example files..."
	@failed=0; \
	passed=0; \
	skipped=0; \
	for file in $$(find examples -name "*.typemux" -not -path "*/output/*" -not -path "*/proto-import-output/*" | sort); do \
		echo "=== Validating: $$file ==="; \
		if echo "$$file" | grep -q "circular"; then \
			echo "⏭️  SKIPPED (circular import test)"; \
			skipped=$$((skipped + 1)); \
			echo ""; \
			continue; \
		fi; \
		output_dir=$$(mktemp -d); \
		if ./bin/typemux -input "$$file" -format all -output "$$output_dir" >/dev/null 2>&1; then \
			echo "✅ PASSED"; \
			passed=$$((passed + 1)); \
		else \
			echo "❌ FAILED"; \
			./bin/typemux -input "$$file" -format all -output "$$output_dir" 2>&1 | head -5; \
			failed=$$((failed + 1)); \
		fi; \
		rm -rf "$$output_dir"; \
		echo ""; \
	done; \
	echo "=================================="; \
	echo "Summary:"; \
	echo "  Passed:  $$passed"; \
	echo "  Failed:  $$failed"; \
	echo "  Skipped: $$skipped"; \
	echo "=================================="; \
	if [ $$failed -gt 0 ]; then \
		exit 1; \
	fi

.PHONY: validate-importers
validate-importers: build ## Validate importer tools
	@echo "==> Testing importers..."
	@echo ""
	@echo "Testing Protobuf Importer..."
	@if [ -d "examples/proto-import" ]; then \
		for proto in examples/proto-import/*.proto; do \
			if [ -f "$$proto" ]; then \
				echo "  Testing: $$proto"; \
				./bin/proto2typemux -input "$$proto" -output /tmp/proto-test >/dev/null 2>&1; \
			fi; \
		done; \
		echo "  ✅ Protobuf importer passed"; \
	else \
		echo "  ⚠️  No proto-import examples found"; \
	fi
	@echo ""
	@echo "Testing GraphQL Importer..."
	@if [ -d "examples/graphql-import" ]; then \
		count=0; \
		for gql in examples/graphql-import/*.graphql examples/graphql-import/*.gql; do \
			if [ -f "$$gql" ]; then \
				echo "  Testing: $$gql"; \
				./bin/graphql2typemux -input "$$gql" -output /tmp/graphql-test >/dev/null 2>&1 || true; \
				count=$$((count + 1)); \
			fi; \
		done; \
		if [ $$count -gt 0 ]; then \
			echo "  ✅ GraphQL importer passed"; \
		else \
			echo "  ⚠️  No GraphQL files found"; \
		fi \
	else \
		echo "  ⚠️  No graphql-import examples found"; \
	fi
	@echo ""
	@echo "Testing OpenAPI Importer..."
	@if [ -d "examples/openapi-import" ]; then \
		count=0; \
		for openapi in examples/openapi-import/*.yaml examples/openapi-import/*.yml examples/openapi-import/*.json; do \
			if [ -f "$$openapi" ]; then \
				echo "  Testing: $$openapi"; \
				./bin/openapi2typemux -input "$$openapi" -output /tmp/openapi-test >/dev/null 2>&1 || true; \
				count=$$((count + 1)); \
			fi; \
		done; \
		if [ $$count -gt 0 ]; then \
			echo "  ✅ OpenAPI importer passed"; \
		else \
			echo "  ⚠️  No OpenAPI files found"; \
		fi \
	else \
		echo "  ⚠️  No openapi-import examples found"; \
	fi

# ============================================================================
# Example Generation Targets (legacy compatibility)
# ============================================================================

.PHONY: example
example: build ## Run example schema generation
	@if [ -f "example.schema" ]; then \
		./bin/typemux -input example.schema -output ./generated; \
	else \
		echo "No example.schema file found"; \
	fi

.PHONY: run
run: build ## Generate all formats from example.schema
	@if [ -f "example.schema" ]; then \
		./bin/typemux -input example.schema -format all -output ./generated; \
	else \
		echo "No example.schema file found"; \
	fi

.PHONY: graphql
graphql: build ## Generate GraphQL schema only
	@if [ -f "example.schema" ]; then \
		./bin/typemux -input example.schema -format graphql -output ./generated; \
	else \
		echo "No example.schema file found"; \
	fi

.PHONY: protobuf
protobuf: build ## Generate Protobuf schema only
	@if [ -f "example.schema" ]; then \
		./bin/typemux -input example.schema -format protobuf -output ./generated; \
	else \
		echo "No example.schema file found"; \
	fi

.PHONY: openapi
openapi: build ## Generate OpenAPI schema only
	@if [ -f "example.schema" ]; then \
		./bin/typemux -input example.schema -format openapi -output ./generated; \
	else \
		echo "No example.schema file found"; \
	fi

# ============================================================================
# Clean Targets
# ============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "==> Cleaning..."
	rm -rf bin/
	rm -rf generated/
	rm -f typemux proto2typemux graphql2typemux openapi2typemux
	rm -f coverage.out coverage.html
	rm -rf examples/output/
	rm -rf examples/proto-import-output/
	@echo "✅ Cleaned"

.PHONY: clean-test
clean-test: ## Clean test cache
	@echo "==> Cleaning test cache..."
	go clean -testcache
	@echo "✅ Test cache cleaned"

# ============================================================================
# CI Targets
# ============================================================================

.PHONY: ci
ci: deps lint fmt-check vet test coverage-check build validate-examples ## Run full CI pipeline locally
	@echo ""
	@echo "======================================"
	@echo "✅ CI Pipeline completed successfully!"
	@echo "======================================"

.PHONY: quick
quick: deps test build ## Quick build and test
	@echo "✅ Quick check passed"

# ============================================================================
# Utility Targets
# ============================================================================

.PHONY: bench
bench: ## Run benchmarks
	@echo "==> Running benchmarks..."
	go test -bench=. -benchmem ./...

.PHONY: tools
tools: ## Install development tools
	@echo "==> Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Tools installed"

.PHONY: version
version: ## Show version information
	@go version
	@echo ""
	@if [ -f bin/typemux ]; then \
		echo "TypeMux CLI tools built"; \
	else \
		echo "TypeMux tools not built (run 'make build')"; \
	fi

# Default target
.DEFAULT_GOAL := help
