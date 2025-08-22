# VJVector Makefile
# Common development tasks for the vector database project

.PHONY: help build test clean lint format coverage docker-build docker-run install-tools

# Default target
help:
	@echo "VJVector - AI-First Vector Database"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"
	@echo "  coverage      - Run tests with coverage"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  install-tools - Install development tools"
	@echo "  deps          - Download dependencies"
	@echo "  run           - Run the application locally"

# Build the application
build:
	@echo "Building vjvector..."
	@mkdir -p bin
	go build -o bin/vjvector ./cmd/vjvector
	@echo "Build complete: bin/vjvector"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf coverage.out
	@echo "Clean complete"

# Run linter
lint:
	@echo "Running linter..."
	./tools/run-lint.sh run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@if [ "$(shell go env CGO_ENABLED)" = "1" ]; then \
		go test -v -race -coverprofile=coverage.out -covermode=atomic ./...; \
	else \
		go test -v -coverprofile=coverage.out -covermode=atomic ./...; \
	fi
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t vjvector:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/data:/app/data vjvector:latest

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run the application locally
run: build
	@echo "Running vjvector..."
	./bin/vjvector serve

# Development mode with hot reload (requires air)
dev:
	@echo "Starting development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		make run; \
	fi

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	@if command -v mockgen > /dev/null; then \
		mockgen -source=pkg/core/vector.go -destination=pkg/core/mocks.go; \
	else \
		echo "Mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"; \
	fi

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "Gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Race condition detection
race:
	@echo "Running tests with race detection..."
	@if [ "$(shell go env CGO_ENABLED)" = "1" ]; then \
		go test -race ./...; \
	else \
		echo "CGO not enabled, skipping race detection"; \
		go test ./...; \
	fi

# Generate documentation
docs:
	@echo "Generating documentation..."
	@if command -v godoc > /dev/null; then \
		godoc -http=:6060; \
	else \
		echo "Godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Pre-commit checks
pre-commit: format lint test
	@echo "Pre-commit checks completed successfully"

# CI checks
ci: deps test lint security
	@echo "CI checks completed successfully"
