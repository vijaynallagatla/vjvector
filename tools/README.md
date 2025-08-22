# Tools Directory

This directory contains pre-built binaries and scripts for CI/CD and local development.

## Contents

### golangci-lint
- **Version**: 2.4.0
- **Built with**: Go 1.25.0
- **Source**: Built from [golangci-lint v2.4.0](https://github.com/golangci/golangci-lint/releases/tag/v2.4.0)
- **Purpose**: Static code analysis and linting

### run-lint.sh
- **Purpose**: Shell script wrapper for golangci-lint
- **Usage**: `./tools/run-lint.sh [args]`
- **Features**: 
  - Automatically checks if binary exists and is executable
  - Shows version information before running
  - Passes through all arguments to golangci-lint

## Why Pre-built Binaries?

1. **Consistency**: Ensures the same version is used in CI and local development
2. **Go Version Compatibility**: Built with Go 1.25.0 to match our project requirements
3. **CI Reliability**: Eliminates dependency on external actions that may have version conflicts
4. **Performance**: No need to download and install tools during CI runs

## Updating Tools

To update golangci-lint to a newer version:

1. Clone the golangci-lint repository
2. Checkout the desired version tag
3. Build with Go 1.25.0: `go build -o tools/golangci-lint ./cmd/golangci-lint`
4. Update this README with the new version information

## Usage in CI

The CI workflow uses these tools directly:

```yaml
- name: golangci-lint
  run: |
    chmod +x ./tools/run-lint.sh
    ./tools/run-lint.sh run --timeout=5m
```

## Local Development

Use the Makefile targets that reference these tools:

```bash
make lint        # Runs golangci-lint using local binary
```
