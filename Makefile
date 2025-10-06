.PHONY: build lint test run tui clean install deps

# Build the application
build:
	@echo "Building media2goodreads..."
	@mkdir -p bin
	@go build -o bin/media2goodreads ./cmd/media2goodreads

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the CLI
run: build
	@./bin/media2goodreads

# Run the TUI
tui: build
	@./bin/media2goodreads tui

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Install golangci-lint (if not already installed)
install-lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Install the binary to GOPATH/bin
install: build
	@echo "Installing to $(go env GOPATH)/bin..."
	@cp bin/media2goodreads $(shell go env GOPATH)/bin/

# Run all checks (format, lint, test)
check: fmt lint test
	@echo "All checks passed!"
