# VJVector Makefile
# Common development tasks for the vector database project

.PHONY: help build test clean lint format coverage docker-build docker-run install-tools cluster-prod cluster-dev cluster-stop cluster-status cluster-logs cluster-scale cluster-cleanup

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
	@echo "  install-air   - Install Air for hot reloading"
	@echo "  deps          - Download dependencies"
	@echo "  run           - Run the application locally"
	@echo "  dev           - Run with hot reload (Air)"
	@echo "  dev-air       - Run with Air using custom script"
	@echo ""
	@echo "Cluster Management:"
	@echo "  cluster-prod     - Start production cluster"
	@echo "  cluster-dev      - Start development cluster"
	@echo "  cluster-stop     - Stop all clusters"
	@echo "  cluster-status   - Show cluster status"
	@echo "  cluster-logs     - Show cluster logs"
	@echo "  cluster-scale    - Scale cluster nodes"
	@echo "  cluster-cleanup  - Clean up all cluster resources"
	@echo "  cluster-info     - Show comprehensive cluster information"
	@echo "  cluster-test     - Test cluster connectivity"
	@echo "  cluster-resources - Show cluster resource usage"
	@echo "  cluster-backup   - Backup cluster data"
	@echo "  cluster-restart  - Restart production cluster"
	@echo ""
	@echo "Setup & Utilities:"
	@echo "  setup-dev        - Full development environment setup"
	@echo "  setup-prod       - Full production environment setup"
	@echo "  install-all      - Install all development tools"
	@echo "  commands         - Show all available commands"

# Build the application
build:
	@echo "Building vjvector..."
	@mkdir -p bin
	go build -o bin/vjvector ./cmd/api
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

lint-ci:
	@echo "Installing and running linter for CI..."
	./tools/install-lint-linux.sh
	./tools/bin/golangci-lint run

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

# Build development Docker image
docker-build-dev:
	@echo "Building development Docker image..."
	docker build -f Dockerfile.dev -t vjvector:dev .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/data:/app/data vjvector:latest

# Run development Docker container
docker-run-dev:
	@echo "Running development Docker container..."
	docker run -p 8080:8080 -p 8081:8081 -p 2345:2345 \
		-v $(PWD)/data:/app/data \
		-v $(PWD):/app/src \
		-v $(PWD)/config.cluster.yaml:/app/config.cluster.yaml \
		vjvector:dev



# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v goimports > /dev/null; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@if ! command -v air > /dev/null; then \
		echo "Installing Air for hot reloading..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@echo "Development tools installed!"

# Install air specifically
install-air:
	@echo "Installing Air for hot reloading..."
	@go install github.com/air-verse/air@latest
	@if [ -f "$(shell go env GOPATH)/bin/air" ]; then \
		echo "Copying Air to tools/bin/..."; \
		cp $(shell go env GOPATH)/bin/air tools/bin/; \
		echo "Air installed and copied to tools/bin/"; \
	else \
		echo "Failed to install Air"; \
		exit 1; \
	fi

# Run the application locally
run: build
	@echo "Running vjvector..."
	./bin/vjvector serve

# Development mode with hot reload (requires air)
dev:
	@echo "Starting development mode with hot reload..."
	@if [ -f "./tools/bin/air" ]; then \
		./tools/bin/air; \
	elif command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing air..."; \
		go install github.com/air-verse/air@latest; \
		cp $(shell go env GOPATH)/bin/air tools/bin/; \
		echo "Air installed. Starting hot reload..."; \
		./tools/bin/air; \
	fi

# Run air with custom script
dev-air:
	@echo "Starting development mode with air script..."
	@./tools/run-air.sh

# Start development with cluster
dev-cluster: cluster-dev
	@echo "Development cluster started. You can now run:"
	@echo "  make run          # Run locally"
	@echo "  make dev          # Run with hot reload"
	@echo "  make docker-run-dev # Run in development container"

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	@if command -v mockgen > /dev/null; then \
		mockgen -source=pkg/core/vector.go -destination=pkg/core/mocks.go; \
	else \
		echo "Mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"; \
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
ci: deps test lint
	@echo "CI checks completed successfully"

# =============================================================================
# Cluster Management Commands
# =============================================================================

# Start production cluster
cluster-prod:
	@echo "Starting production cluster..."
	@if [ -f "docker-compose.yml" ]; then \
		docker compose up -d; \
		echo "Production cluster started successfully!"; \
		echo "Services available at:"; \
		echo "  - VJVector Master: http://localhost:8080"; \
		echo "  - VJVector Slave 1: http://localhost:8082"; \
		echo "  - VJVector Slave 2: http://localhost:8084"; \
		echo "  - Load Balancer: http://localhost:80"; \
		echo "  - Prometheus: http://localhost:9090"; \
		echo "  - Grafana: http://localhost:3000 (admin:admin)"; \
		echo "  - HAProxy Stats: http://localhost:8404 (admin:admin123)"; \
	else \
		echo "Error: docker-compose.yml not found"; \
		exit 1; \
	fi

# Start development cluster
cluster-dev:
	@echo "Starting development cluster..."
	@if [ -f "docker-compose.dev.yml" ]; then \
		docker compose -f docker-compose.dev.yml up -d; \
		echo "Development cluster started successfully!"; \
		echo "Services available at:"; \
		echo "  - VJVector Master: http://localhost:8080"; \
		echo "  - VJVector Slave: http://localhost:8082"; \
		echo "  - Load Balancer: http://localhost:80"; \
		echo "  - Prometheus: http://localhost:9090"; \
		echo "  - Grafana: http://localhost:3000 (admin:admin)"; \
		echo "  - Jaeger: http://localhost:16686"; \
	else \
		echo "Error: docker-compose.dev.yml not found"; \
		exit 1; \
	fi

# Stop all clusters
cluster-stop:
	@echo "Stopping all clusters..."
	@if [ -f "docker-compose.dev.yml" ]; then \
		docker compose -f docker-compose.dev.yml down; \
	fi
	@if [ -f "docker-compose.yml" ]; then \
		docker compose down; \
	fi
	@echo "All clusters stopped successfully!"

# Show cluster status
cluster-status:
	@echo "Cluster Status:"
	@echo ""
	@if [ -f "docker-compose.dev.yml" ]; then \
		echo "Development cluster:"; \
		docker compose -f docker-compose.dev.yml ps; \
		echo ""; \
	fi
	@if [ -f "docker-compose.yml" ]; then \
		echo "Production cluster:"; \
		docker compose ps; \
	fi

# Show cluster logs
cluster-logs:
	@echo "Showing cluster logs..."
	@if [ -f "docker-compose.yml" ]; then \
		docker compose logs -f; \
	else \
		echo "No production cluster found"; \
	fi

# Scale cluster nodes (default: 3)
cluster-scale:
	@echo "Scaling cluster to $(or $(SCALE),3) slave nodes..."
	@if [ -f "docker-compose.yml" ]; then \
		docker compose up -d --scale vjvector-slave=$(or $(SCALE),3); \
		echo "Cluster scaled successfully!"; \
	else \
		echo "Error: docker-compose.yml not found"; \
		exit 1; \
	fi

# Clean up all cluster resources
cluster-cleanup:
	@echo "This will remove all containers, volumes, and networks. Are you sure? (y/N)"
	@read -p "Enter 'y' to confirm: " response; \
	if [ "$$response" = "y" ] || [ "$$response" = "Y" ]; then \
		echo "Cleaning up cluster resources..."; \
		if [ -f "docker-compose.dev.yml" ]; then \
			docker compose -f docker-compose.dev.yml down -v --remove-orphans; \
		fi; \
		if [ -f "docker-compose.yml" ]; then \
			docker compose down -v --remove-orphans; \
		fi; \
		docker system prune -f; \
		echo "Cleanup completed!"; \
	else \
		echo "Cleanup cancelled."; \
	fi

# Build and start production cluster
cluster-build-prod: docker-build cluster-prod
	@echo "Production cluster built and started successfully!"

# Build and start development cluster
cluster-build-dev: docker-build cluster-dev
	@echo "Development cluster built and started successfully!"

# Quick cluster restart
cluster-restart: cluster-stop
	@echo "Waiting 5 seconds before restarting..."
	@sleep 5
	@make cluster-prod

# Show cluster health
cluster-health:
	@echo "Checking cluster health..."
	@if [ -f "docker-compose.yml" ]; then \
		echo "Production cluster health:"; \
		curl -s http://localhost:8080/health | jq . 2>/dev/null || echo "Master node not responding"; \
		curl -s http://localhost:8082/health | jq . 2>/dev/null || echo "Slave 1 not responding"; \
		curl -s http://localhost:8084/health | jq . 2>/dev/null || echo "Slave 2 not responding"; \
		curl -s http://localhost/health | jq . 2>/dev/null || echo "Load balancer not responding"; \
	else \
		echo "No production cluster found"; \
	fi

# Show cluster metrics
cluster-metrics:
	@echo "Cluster metrics:"
	@if [ -f "docker-compose.yml" ]; then \
		echo "Master node metrics:"; \
		curl -s http://localhost:8081/metrics | head -20; \
		echo ""; \
		echo "Load balancer stats:"; \
		curl -s -u admin:admin123 http://localhost:8404/stats | head -20; \
	else \
		echo "No production cluster found"; \
	fi

# Backup cluster data
cluster-backup:
	@echo "Creating cluster backup..."
	@mkdir -p backups/$(shell date +%Y%m%d_%H%M%S)
	@if [ -f "docker-compose.yml" ]; then \
		docker compose exec -T etcd-0 etcdctl snapshot save /tmp/backup.db; \
		docker cp etcd-0:/tmp/backup.db backups/$(shell date +%Y%m%d_%H%M%S)/etcd-backup.db; \
		echo "etcd backup created"; \
	fi
	@echo "Backup completed in backups/$(shell date +%Y%m%d_%H%M%S)/"

# Show cluster configuration
cluster-config:
	@echo "Cluster configuration:"
	@if [ -f "config.cluster.yaml" ]; then \
		echo "Cluster config:"; \
		cat config.cluster.yaml; \
	else \
		echo "No cluster configuration found"; \
	fi

# Show cluster endpoints
cluster-endpoints:
	@echo "Cluster endpoints:"
	@echo "Production cluster:"
	@echo "  - VJVector Master: http://localhost:8080"
	@echo "  - VJVector Slave 1: http://localhost:8082"
	@echo "  - VJVector Slave 2: http://localhost:8084"
	@echo "  - Load Balancer: http://localhost:80"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000 (admin:admin)"
	@echo "  - HAProxy Stats: http://localhost:8404 (admin:admin123)"
	@echo ""
	@echo "Development cluster:"
	@echo "  - VJVector Master: http://localhost:8080"
	@echo "  - VJVector Slave: http://localhost:8082"
	@echo "  - Load Balancer: http://localhost:80"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000 (admin:admin)"
	@echo "  - Jaeger: http://localhost:16686"

# Test cluster connectivity
cluster-test:
	@echo "Testing cluster connectivity..."
	@if [ -f "docker-compose.yml" ]; then \
		echo "Testing production cluster:"; \
		curl -s -o /dev/null -w "Master: %{http_code}\n" http://localhost:8080/health || echo "Master: FAILED"; \
		curl -s -o /dev/null -w "Slave 1: %{http_code}\n" http://localhost:8082/health || echo "Slave 1: FAILED"; \
		curl -s -o /dev/null -w "Slave 2: %{http_code}\n" http://localhost:8084/health || echo "Slave 2: FAILED"; \
		curl -s -o /dev/null -w "Load Balancer: %{http_code}\n" http://localhost/health || echo "Load Balancer: FAILED"; \
	else \
		echo "No production cluster found"; \
	fi

# Show cluster resource usage
cluster-resources:
	@echo "Cluster resource usage:"
	@if [ -f "docker-compose.yml" ]; then \
		docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"; \
	else \
		echo "No production cluster found"; \
	fi

# Quick cluster info
cluster-info: cluster-status cluster-endpoints
	@echo ""
	@echo "Use 'make cluster-health' to check health status"
	@echo "Use 'make cluster-metrics' to view metrics"
	@echo "Use 'make cluster-logs' to view logs"

# =============================================================================
# Utility Commands
# =============================================================================

# Install all development tools
install-all: install-tools
	@echo "Installing additional development tools..."
	@if ! command -v air > /dev/null; then \
		echo "Installing Air for hot reloading..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@if [ -f "$(shell go env GOPATH)/bin/air" ] && [ ! -f "./tools/bin/air" ]; then \
		echo "Copying Air to tools/bin/..."; \
		cp $(shell go env GOPATH)/bin/air tools/bin/; \
	fi
	@if ! command -v mockgen > /dev/null; then \
		echo "Installing Mockgen for mock generation..."; \
		go install github.com/golang/mock/mockgen@latest; \
	fi
	@if ! command -v godoc > /dev/null; then \
		echo "Installing Godoc for documentation..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "All development tools installed!"

# Full development setup
setup-dev: install-all deps cluster-dev
	@echo "Development environment setup complete!"
	@echo "You can now run:"
	@echo "  make run          # Run locally"
	@echo "  make dev          # Run with hot reload"
	@echo "  make cluster-info # Show cluster information"

# Full production setup
setup-prod: deps cluster-build-prod
	@echo "Production environment setup complete!"
	@echo "You can now run:"
	@echo "  make cluster-info # Show cluster information"
	@echo "  make cluster-health # Check cluster health"

# Show all available commands
commands:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make setup-dev    # Full development setup"
	@echo "  make dev-cluster  # Start development cluster"
	@echo "  make run          # Run locally"
	@echo "  make dev          # Run with hot reload (Air)"
	@echo "  make dev-air      # Run with Air using custom script"
	@echo ""
	@echo "Production:"
	@echo "  make setup-prod   # Full production setup"
	@echo "  make cluster-prod # Start production cluster"
	@echo "  make cluster-info # Show cluster information"
	@echo ""
	@echo "Management:"
	@echo "  make cluster-stop    # Stop all clusters"
	@echo "  make cluster-restart # Restart production cluster"
	@echo "  make cluster-scale   # Scale cluster nodes"
	@echo "  make cluster-cleanup # Clean up resources"
	@echo ""
	@echo "Monitoring:"
	@echo "  make cluster-health   # Check cluster health"
	@echo "  make cluster-metrics  # View metrics"
	@echo "  make cluster-logs     # View logs"
	@echo "  make cluster-test     # Test connectivity"
	@echo "  make cluster-resources # Show resource usage"
