# Quickstart: CLI Framework with Flags

**Feature**: 003-cobra-cli-flags

## Overview

This guide provides a quick reference for developers implementing the CLI flags feature.

## Key Changes

1. **Default Behavior Change**: By default, gitree now shows only repositories needing attention
2. **New --all Flag**: Shows all repositories including clean ones (previous default behavior)
3. **New --version Flag**: Displays version, build time, and commit hash
4. **New --no-color Flag**: Suppresses all color output

## Quick Reference

### Flag Usage

```bash
# Default: show only repos needing attention
gitree

# Show all repos including clean ones
gitree --all

# Show version
gitree --version

# Suppress colors
gitree --no-color

# Combine flags
gitree --all --no-color
```

### Implementation Packages

- **cmd/gitree/**: Cobra command setup, version variables
- **internal/cli/**: Filtering logic (IsClean, FilterRepositories)
- **Makefile**: Build-time variable injection

### Testing

```bash
# Run tests
make test

# Build with version info
make build
```

## Development Workflow (TDD)

Per Constitution III (Test-First):

1. Write tests first
2. Get user approval of tests
3. Implement (Red-Green-Refactor)

## Next Steps

See  command to generate implementation tasks.
