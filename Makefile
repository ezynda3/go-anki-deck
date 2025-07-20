.PHONY: all test bench coverage lint clean build install help

# Default target
all: test

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f coverage.txt coverage.html
	@rm -f example/*.apkg
	@rm -f example/main
	@go clean -cache

# Build the package
build:
	@echo "Building..."
	@go build -v ./...

# Install the package
install:
	@echo "Installing..."
	@go install ./...

# Run example
example:
	@echo "Running example..."
	@cd example && go run main.go

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  make test      - Run tests"
	@echo "  make bench     - Run benchmarks"
	@echo "  make coverage  - Generate test coverage report"
	@echo "  make lint      - Run linter"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make build     - Build the package"
	@echo "  make install   - Install the package"
	@echo "  make example   - Run example"
	@echo "  make fmt       - Format code"
	@echo "  make tidy      - Tidy dependencies"
	@echo "  make help      - Show this help message"