#!/bin/bash

# Run golangci-lint using the local binary
# This ensures we use the version built with Go 1.25.0

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LINT_BIN="$SCRIPT_DIR/golangci-lint"

if [ ! -f "$LINT_BIN" ]; then
    echo "Error: golangci-lint binary not found at $LINT_BIN"
    exit 1
fi

if [ ! -x "$LINT_BIN" ]; then
    chmod +x "$LINT_BIN"
fi

echo "Using golangci-lint: $($LINT_BIN version)"
echo "Running linting..."

exec "$LINT_BIN" "$@"
