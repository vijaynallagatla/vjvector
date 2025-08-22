#!/bin/bash

# Install golangci-lint for Linux (CI environment)
# This script downloads and installs the correct binary for GitHub Actions runners

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
LINT_VERSION="${GO_LINT_VERSION:-v2.4.0}"

echo "Installing golangci-lint $LINT_VERSION for Linux..."

# Create bin directory if it doesn't exist
mkdir -p "$BIN_DIR"

# Download and install golangci-lint for Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$BIN_DIR" "$LINT_VERSION"

# Make it executable
chmod +x "$BIN_DIR/golangci-lint"

# Verify installation
echo "Installed golangci-lint:"
"$BIN_DIR/golangci-lint" version

echo "Installation complete. Binary location: $BIN_DIR/golangci-lint"
