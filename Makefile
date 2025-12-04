.PHONY: build run test clean help

# Binary name
BINARY_NAME=salesforce-splunk-migration

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME).exe .
	@echo "Build complete!"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run .

# Run migration
migrate:
	@echo "Running migration..."
	go run . migrate --config credentials.json

# Validate configuration
validate:
	@echo "Validating configuration..."
	go run . validate --config credentials.json

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@if exist $(BINARY_NAME).exe del $(BINARY_NAME).exe
	@if exist coverage.out del coverage.out
	@if exist coverage.html del coverage.html
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  run       - Run the application"
	@echo "  migrate   - Run full migration"
	@echo "  validate  - Validate configuration"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage"
	@echo "  clean     - Clean build artifacts"
	@echo "  deps      - Install dependencies"
	@echo "  fmt       - Format code"
	@echo "  lint      - Run linter"
	@echo "  help      - Show this help message"
