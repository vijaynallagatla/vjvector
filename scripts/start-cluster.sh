#!/bin/bash

# VJVector Cluster Startup Script
# This script provides easy commands to start, stop, and manage the VJVector cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  VJVector Cluster Manager${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running. Please start Docker first."
        exit 1
    fi
    
    print_status "Prerequisites check passed!"
}

# Function to start production cluster
start_production() {
    print_status "Starting production cluster..."
    docker-compose up -d
    
    print_status "Waiting for services to be healthy..."
    sleep 30
    
    print_status "Production cluster started successfully!"
    print_status "Services available at:"
    echo "  - VJVector Master: http://localhost:8080"
    echo "  - VJVector Slave 1: http://localhost:8082"
    echo "  - VJVector Slave 2: http://localhost:8084"
    echo "  - Load Balancer: http://localhost:80"
    echo "  - Prometheus: http://localhost:9090"
    echo "  - Grafana: http://localhost:3000 (admin:admin)"
    echo "  - HAProxy Stats: http://localhost:8404 (admin:admin123)"
}

# Function to start development cluster
start_development() {
    print_status "Starting development cluster..."
    docker-compose -f docker-compose.dev.yml up -d
    
    print_status "Waiting for services to be healthy..."
    sleep 30
    
    print_status "Development cluster started successfully!"
    print_status "Services available at:"
    echo "  - VJVector Master: http://localhost:8080"
    echo "  - VJVector Slave: http://localhost:8082"
    echo "  - Load Balancer: http://localhost:80"
    echo "  - Prometheus: http://localhost:9090"
    echo "  - Grafana: http://localhost:3000 (admin:admin)"
    echo "  - Jaeger: http://localhost:16686"
}

# Function to stop cluster
stop_cluster() {
    print_status "Stopping cluster..."
    
    if [ -f "docker-compose.dev.yml" ]; then
        docker-compose -f docker-compose.dev.yml down
    fi
    
    docker-compose down
    
    print_status "Cluster stopped successfully!"
}

# Function to restart cluster
restart_cluster() {
    print_status "Restarting cluster..."
    stop_cluster
    sleep 5
    start_production
}

# Function to show cluster status
show_status() {
    print_status "Cluster status:"
    echo ""
    
    if [ -f "docker-compose.dev.yml" ]; then
        echo "Development cluster:"
        docker-compose -f docker-compose.dev.yml ps
        echo ""
    fi
    
    echo "Production cluster:"
    docker-compose ps
}

# Function to show logs
show_logs() {
    local service=${1:-"vjvector-master"}
    print_status "Showing logs for $service..."
    docker-compose logs -f "$service"
}

# Function to scale cluster
scale_cluster() {
    local count=${1:-3}
    print_status "Scaling cluster to $count slave nodes..."
    docker-compose up -d --scale vjvector-slave=$count
    print_status "Cluster scaled successfully!"
}

# Function to clean up
cleanup() {
    print_warning "This will remove all containers, volumes, and networks. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Cleaning up cluster..."
        docker-compose down -v --remove-orphans
        docker-compose -f docker-compose.dev.yml down -v --remove-orphans
        docker system prune -f
        print_status "Cleanup completed!"
    else
        print_status "Cleanup cancelled."
    fi
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  start-prod     Start production cluster"
    echo "  start-dev      Start development cluster"
    echo "  stop           Stop all clusters"
    echo "  restart        Restart production cluster"
    echo "  status         Show cluster status"
    echo "  logs [SERVICE] Show logs for service (default: vjvector-master)"
    echo "  scale [COUNT]  Scale cluster to N slave nodes (default: 3)"
    echo "  cleanup        Remove all containers, volumes, and networks"
    echo "  help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start-prod          # Start production cluster"
    echo "  $0 start-dev           # Start development cluster"
    echo "  $0 logs vjvector-slave # Show slave node logs"
    echo "  $0 scale 5             # Scale to 5 slave nodes"
}

# Main script logic
main() {
    print_header
    
    case "${1:-help}" in
        "start-prod"|"start-production")
            check_prerequisites
            start_production
            ;;
        "start-dev"|"start-development")
            check_prerequisites
            start_development
            ;;
        "stop")
            stop_cluster
            ;;
        "restart")
            restart_cluster
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs "$2"
            ;;
        "scale")
            scale_cluster "$2"
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
