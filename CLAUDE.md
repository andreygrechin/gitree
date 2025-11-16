# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`gitree` is a CLI tool that recursively scans directories for Git repositories and displays them in a tree structure with status information. It uses go-git for Git operations and implements concurrent status extraction with configurable timeouts.

## Build & Development Commands

### Building

```bash
make build              # Build binary to bin/gitree
```

### Testing

```bash
make test               # Run all tests
go test ./...           # Alternative: run all tests
go test -run TestName ./internal/scanner  # Run specific test
```

### Code Quality

```bash
make lint               # Format code and run all linters (includes fmt, go vet, staticcheck, golangci-lint)
make format             # Format code with gofumpt and auto-fix some issues
make fmt                # Alias for format
make vuln               # Run security checks (gosec, govulncheck)
```

### Coverage

```bash
make cov-unit           # Generate unit test coverage report (outputs to covdatafiles/)
make cov-integration    # Generate integration coverage with instrumented binary
```

### Release

```bash
make release-test       # Test release process with snapshot
make release            # Create actual release (requires clean working tree)
make clean              # Remove build artifacts and caches
```

## Architecture

### Core Flow

1. **Scanner** (`internal/scanner/`) - Recursively walks directory tree, detects Git repositories (both regular and bare)
2. **Git Status** (`internal/gitstatus/`) - Concurrently extracts status for multiple repos with timeout controls
3. **Tree Formatter** (`internal/tree/`) - Builds hierarchical tree structure and formats ASCII output
4. **Main CLI** (`cmd/gitree/`) - Orchestrates scanner → status extraction → tree formatting pipeline

### Key Packages

**internal/scanner**

- `Scan()` - Main entry point for recursive directory scanning
- `IsGitRepository()` - Detects regular (.git dir) and bare repos (HEAD, refs/, objects/)
- Implements symlink loop detection via inode tracking
- Skips into repository contents once detected (FR-018)

**internal/gitstatus**

- `Extract()` - Single repo status extraction with timeout
- `ExtractBatch()` - Concurrent batch processing with semaphore-based concurrency control
- Extracts: branch, detached HEAD, remote status, ahead/behind counts, stashes, uncommitted changes
- Uses go-git library, not shell commands

**internal/tree**

- `Build()` - Constructs hierarchical tree from flat repository list
- `Format()` - Generates ASCII tree with box-drawing characters
- Handles intermediate directory nodes for nested repos

**internal/models**

- Core data structures: `Repository`, `GitStatus`, `TreeNode`, `ScanResult`
- Validation methods on all models
- `GitStatus.Format()` returns display string like `[main ↑2 ↓1 $ *]`

### Concurrency Model

- Main program uses context with 5-minute timeout
- Git status extraction uses worker pool pattern with semaphore (default: 10 concurrent)
- Each repository status extraction has individual timeout (default: 10s)
- All operations respect context cancellation

### Error Handling

- Scanner: permission errors are non-fatal, collected in `ScanResult.Errors`
- Git status: partial status returned on errors (e.g., branch="unknown", Error field set)
- Validation errors use sentinel errors (`errScanResultValidation`, `errRepositoryValidation`, etc.)

## Testing Strategy

- All packages have `_test.go` files using testify/assert
- Scanner tests: use temp directories with actual Git repos
- Git status tests: create test repos with various states
- Formatter tests: verify tree structure and ASCII output
- Use `t.TempDir()` for test isolation
- File permissions in tests use octal notation (0o755, 0o644)

## Code Conventions

- Go 1.25+
- Uses golangci-lint with extensive linters enabled
- Line length limit: 140 characters
- Error wrapping: use `fmt.Errorf("message: %w", err)` with sentinel errors
- Constants: defined at package level with descriptive names (e.g., `defaultTimeout`, `maxConcurrentRequests`)
- Structured logging: currently prints to stderr/stdout directly (spinner to stderr, output to stdout)

## Dependencies

- `github.com/go-git/go-git/v5` - Pure Go Git implementation (no git binary required)
- `github.com/briandowns/spinner` - Terminal spinner for progress indication
- `github.com/stretchr/testify` - Testing assertions and mocking

## Key Implementation Details

- **Bare repository detection**: Checks for HEAD file + refs/ and objects/ directories in repo root
- **Symlink handling**: Uses inode tracking to prevent loops, marks repos reached via symlinks
- **Status display**: Symbols in git status: ○ (no remote), ↑ (ahead), ↓ (behind), $ (stashes), * (uncommitted)
- **Tree building**: Converts flat repository list into hierarchy, creates intermediate directory nodes as needed
- **Performance**: Spinner updates during long operations, concurrent git status extraction for speed

## Active Technologies

- Go 1.25.4 + fatih/color (already present via go-git dependencies) or golang.org/x/term for color suppor (002-colorized-status-output)

## Recent Changes

- 002-colorized-status-output: Added Go 1.25.4 + fatih/color (already present via go-git dependencies) or golang.org/x/term for color suppor
