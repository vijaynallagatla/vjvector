# VJVector Development Tools

This directory contains development tools and utilities for the VJVector project.

## Tools Included

### Air - Hot Reloading
**Binary**: `tools/bin/air`  
**Version**: v1.62.0  
**Go Version**: 1.24.6+  
**Purpose**: Hot reloading for Go development

#### Installation
```bash
# Install air globally
make install-air

# Or install all tools
make install-all
```

#### Usage
```bash
# Run with hot reloading
make dev

# Run with custom air script
make dev-air

# Run directly
./tools/bin/air
```

#### Configuration
The project includes a customized `.air.toml` configuration file that:
- Watches `cmd/`, `internal/`, and `pkg/` directories
- Excludes `deploy/`, `deployments/`, `docs/`, `examples/`, `scripts/`, `tools/`
- Builds to `./tmp/vjvector`
- Runs with `serve --config ./config.cluster.yaml`
- Includes Go, YAML, and YML files
- Enables screen clearing and timestamps

#### Customization
You can modify `.air.toml` to:
- Change watched directories
- Modify build commands
- Adjust file exclusions
- Customize build delays

### GolangCI-Lint
**Binary**: `tools/bin/golangci-lint`  
**Purpose**: Code linting and static analysis

#### Usage
```bash
# Run linter
make lint

# Run linter for CI
make lint-ci
```

### GoImports
**Purpose**: Go import formatting

#### Usage
```bash
# Format imports
make format
```

## Scripts

### run-air.sh
Custom script for running Air with project-specific configuration and checks.

**Features**:
- Automatic air installation if missing
- Configuration file validation
- Colored output
- Prerequisites checking

**Usage**:
```bash
./tools/run-air.sh
```

## Development Workflow

### 1. Setup Development Environment
```bash
make setup-dev
```

### 2. Start Development Cluster
```bash
make cluster-dev
```

### 3. Run with Hot Reloading
```bash
make dev          # Standard air
make dev-air      # Custom script
```

### 4. Code Quality Checks
```bash
make lint         # Run linter
make format       # Format code
make test         # Run tests
```

## File Structure

```
tools/
├── README.md           # This file
├── bin/                # Binary tools
│   ├── air            # Hot reloading tool
│   └── golangci-lint  # Linting tool
├── run-air.sh         # Custom air runner script
├── run-lint.sh        # Linting script
└── install-lint-linux.sh  # Linux lint installation
```

## Configuration Files

### .air.toml
Air configuration file with VJVector-specific settings:
- **Build Command**: `go build -o ./tmp/vjvector ./cmd/api`
- **Binary**: `./tmp/vjvector serve --config ./config.cluster.yaml`
- **Watch Directories**: `cmd/`, `internal/`, `pkg/`
- **Exclude Directories**: `deploy/`, `deployments/`, `docs/`, `examples/`, `scripts/`, `tools/`
- **File Extensions**: `go`, `yaml`, `yml`
- **Build Delay**: 1000ms
- **Kill Delay**: 0.5s
- **Rerun**: Enabled
- **Screen Clearing**: Enabled

## Troubleshooting

### Air Not Working
```bash
# Reinstall air
make install-air

# Check configuration
cat .air.toml

# Run with verbose output
./tools/bin/air -d
```

### Build Errors
```bash
# Clean build artifacts
make clean

# Check Go modules
go mod tidy
go mod download

# Verify configuration
make cluster-config
```

### Performance Issues
- Reduce watched directories in `.air.toml`
- Increase build delay if builds are too frequent
- Exclude large directories that don't need watching

## Best Practices

1. **Use make commands** instead of running tools directly
2. **Keep .air.toml** in version control for team consistency
3. **Exclude large directories** from watching to improve performance
4. **Use dev-cluster** for development with full infrastructure
5. **Run tests** before committing changes

## Integration with Makefile

The Makefile includes comprehensive targets for all tools:
- `make install-tools` - Install basic tools
- `make install-air` - Install air specifically
- `make install-all` - Install all development tools
- `make dev` - Run with hot reloading
- `make dev-air` - Run with custom air script
- `make setup-dev` - Complete development setup
