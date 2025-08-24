# VJVector Makefile Quick Reference

## 🚀 Quick Start Commands

### Development Setup
```bash
make setup-dev          # Complete development environment setup
make dev-cluster        # Start development cluster
make run                # Run VJVector locally
make dev                # Run with hot reload (Air)
make dev-air            # Run with Air using custom script
```

### Production Setup
```bash
make setup-prod         # Complete production environment setup
make cluster-prod        # Start production cluster
make cluster-info        # Show cluster information
```

## 🏗️ Building & Running

### Local Development
```bash
make build              # Build the application
make run                # Build and run locally
make test               # Run all tests
make lint               # Run linter
make format             # Format code
```

### Docker Operations
```bash
make docker-build       # Build production Docker image
make docker-build-dev   # Build development Docker image
make docker-run         # Run production container
make docker-run-dev     # Run development container
```

## 🎯 Cluster Management

### Start/Stop Clusters
```bash
make cluster-prod       # Start production cluster
make cluster-dev        # Start development cluster
make cluster-stop       # Stop all clusters
make cluster-restart    # Restart production cluster
```

### Cluster Information
```bash
make cluster-status     # Show cluster status
make cluster-info       # Comprehensive cluster information
make cluster-endpoints  # Show all service endpoints
make cluster-config     # Show cluster configuration
```

### Cluster Monitoring
```bash
make cluster-health     # Check cluster health
make cluster-metrics    # View metrics
make cluster-logs       # View logs
make cluster-test       # Test connectivity
make cluster-resources  # Show resource usage
```

### Cluster Operations
```bash
make cluster-scale      # Scale cluster (default: 3 nodes)
make cluster-scale SCALE=5  # Scale to 5 nodes
make cluster-backup     # Backup cluster data
make cluster-cleanup    # Clean up all resources
```

## 🛠️ Development Tools

### Tool Installation
```bash
make install-tools      # Install basic development tools
make install-air        # Install Air for hot reloading
make install-all        # Install all development tools
make deps               # Download Go dependencies
```

### Code Quality
```bash
make lint               # Run linter
make format             # Format code
make coverage           # Run tests with coverage
make race               # Run tests with race detection
make bench              # Run benchmarks
```

## 📊 Monitoring & Debugging

### Health Checks
```bash
make cluster-health     # Check all nodes health
make cluster-test       # Test connectivity to all services
```

### Metrics & Logs
```bash
make cluster-metrics    # View Prometheus metrics
make cluster-logs       # View real-time logs
make cluster-resources  # Monitor resource usage
```

## 🔧 Utility Commands

### Help & Information
```bash
make help               # Show all available targets
make commands           # Show categorized commands
```

### Cleanup & Maintenance
```bash
make clean              # Clean build artifacts
make cluster-cleanup    # Remove all cluster resources
```

## 🌐 Service Endpoints

### Production Cluster
- **VJVector API**: http://localhost:80 (load balanced)
- **Master Node**: http://localhost:8080
- **Slave Nodes**: http://localhost:8082, 8084
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin:admin)
- **HAProxy Stats**: http://localhost:8404 (admin:admin123)

### Development Cluster
- **VJVector API**: http://localhost:80 (load balanced)
- **Master Node**: http://localhost:8080
- **Slave Node**: http://localhost:8082
- **Jaeger**: http://localhost:16686
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin:admin)

## 📝 Common Workflows

### Daily Development
```bash
make setup-dev          # First time setup
make cluster-dev        # Start development cluster
make dev                # Run with hot reloading
# ... make changes ... (automatic rebuild)
make test               # Run tests
make format             # Format code
make lint               # Check code quality
```

### Production Deployment
```bash
make setup-prod         # Setup production environment
make cluster-prod        # Start production cluster
make cluster-info        # Verify cluster status
make cluster-health      # Check health
```

### Troubleshooting
```bash
make cluster-status      # Check what's running
make cluster-logs        # View logs
make cluster-health      # Check health status
make cluster-test        # Test connectivity
make cluster-resources   # Check resource usage
```

### Scaling Operations
```bash
make cluster-scale SCALE=5  # Scale to 5 nodes
make cluster-status          # Verify scaling
make cluster-health          # Check all nodes health
```

## ⚠️ Important Notes

- **Prerequisites**: Docker and Docker Compose must be installed
- **Ports**: Ensure required ports (80, 8080-8085, 9090, 3000, 6379) are available
- **Resources**: Production cluster requires at least 8GB RAM and 20GB disk space
- **Cleanup**: Use `make cluster-cleanup` carefully as it removes all data
- **Scaling**: Scaling operations may require load balancer configuration updates

## 🆘 Getting Help

```bash
make help               # Show all available targets
make commands           # Show categorized commands
```

For more detailed information, see:
- [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md)
- [deploy/local/README.md](./deploy/local/README.md)
