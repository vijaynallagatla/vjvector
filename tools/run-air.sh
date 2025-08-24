#!/bin/bash

# VJVector Air Runner Script
# This script runs air for hot reloading during development

set -e

# Colors for output
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

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  VJVector Air Hot Reloader${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Check if air configuration exists
check_config() {
    if [ ! -f ".air.toml" ]; then
        print_warning "No .air.toml configuration found. Creating default configuration..."
        ./tools/bin/air init
        print_status "Default .air.toml created. You may want to customize it for your needs."
    fi
}

# Check if air binary exists
check_air() {
    if [ ! -f "./tools/bin/air" ]; then
        print_warning "Air binary not found in tools/bin/. Installing air..."
        go install github.com/air-verse/air@latest
        cp $(go env GOPATH)/bin/air tools/bin/
        print_status "Air installed and copied to tools/bin/"
    fi
}

# Main function
main() {
    print_header
    
    # Check prerequisites
    check_air
    check_config
    
    print_status "Starting VJVector with hot reloading..."
    print_status "Configuration: .air.toml"
    print_status "Watching directories: cmd/, internal/, pkg/"
    print_status "Excluding: deploy/, deployments/, docs/, examples/, scripts/, tools/"
    print_status ""
    print_status "Press Ctrl+C to stop"
    print_status ""
    
    # Run air with project configuration
    ./tools/bin/air
}

# Run main function
main "$@"
