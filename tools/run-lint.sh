#!/bin/bash

# Run golangci-lint using the appropriate binary
# Prefers Linux binary for CI, falls back to local macOS binary for development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LINT_BIN="$SCRIPT_DIR/golangci-lint"
LINUX_BIN="$SCRIPT_DIR/bin/golangci-lint"

# Check if we're in CI (GitHub Actions sets CI=true)
if [ "$CI" = "true" ] && [ -f "$LINUX_BIN" ]; then
    echo "CI environment detected, using Linux binary"
    LINT_BIN="$LINUX_BIN"
elif [ -f "$LINUX_BIN" ]; then
    echo "Linux binary found, using it"
    LINT_BIN="$LINUX_BIN"
else
    echo "Using local macOS binary"
fi

if [ ! -f "$LINT_BIN" ]; then
    echo "Error: golangci-lint binary not found at $LINT_BIN"
    echo "For CI, run: ./tools/install-lint-linux.sh"
    echo "For local development, ensure golangci-lint is installed"
    exit 1
fi

if [ ! -x "$LINT_BIN" ]; then
    chmod +x "$LINT_BIN"
fi

echo "Using golangci-lint: $($LINT_BIN version)"
echo "Running linting..."

exec "$LINT_BIN" "$@"
