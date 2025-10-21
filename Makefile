.PHONY: build run clean test example

# Build the CLI binary
build:
	go build -o typemux ./cmd/typemux

# Run the example
example: build
	./typemux -input example.schema -output ./generated

# Run with all formats
run: build
	./typemux -input example.schema -format all -output ./generated

# Generate GraphQL only
graphql: build
	./typemux -input example.schema -format graphql -output ./generated

# Generate Protobuf only
protobuf: build
	./typemux -input example.schema -format protobuf -output ./generated

# Generate OpenAPI only
openapi: build
	./typemux -input example.schema -format openapi -output ./generated

# Clean generated files and binary
clean:
	rm -rf generated typemux

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run tests
test:
	go test ./... -v

# Run tests with coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the CLI binary"
	@echo "  example   - Run the example schema generation"
	@echo "  run       - Generate all formats from example.schema"
	@echo "  graphql   - Generate GraphQL schema only"
	@echo "  protobuf  - Generate Protobuf schema only"
	@echo "  openapi   - Generate OpenAPI schema only"
	@echo "  clean     - Remove generated files and binary"
	@echo "  deps      - Install Go dependencies"
	@echo "  fmt       - Format Go code"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  coverage-html - Generate HTML coverage report"
	@echo "  help      - Show this help message"
