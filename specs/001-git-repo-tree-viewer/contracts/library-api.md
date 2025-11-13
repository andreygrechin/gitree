# Library API Contract

**Version**: 1.0.0
**Date**: 2025-11-13
**Language**: Go 1.25+

## Overview

The gitree library provides modular components for scanning directories, extracting Git status information, and formatting tree output. Each library is independently usable and testable.

## Package Structure

```text
internal/
├── scanner/     # Directory traversal and Git detection
├── gitstatus/   # Git status information extraction
├── tree/        # Tree formatting and output generation
└── models/      # Shared data structures
```

## 1. Scanner Library (`internal/scanner`)

### Purpose

Recursively scans directories to detect Git repositories.

### Public API

#### Function: Scan

```go
func Scan(ctx context.Context, rootPath string, opts ScanOptions) (*models.ScanResult, error)
```

**Parameters**:

- `ctx context.Context`: Context for cancellation and timeout control
- `rootPath string`: Absolute path to start scanning from
- `opts ScanOptions`: Configuration options for scanning behavior

**Returns**:

- `*models.ScanResult`: Results containing found repositories and scan metadata
- `error`: Error if fatal failure occurs (e.g., root path inaccessible)

**Behavior**:

- Recursively traverses directories starting from `rootPath`
- Detects Git repositories by checking for `.git` directory or bare repo structure
- Follows symbolic links once (tracks visited inodes to prevent loops)
- Skips nested repository contents after detecting Git root
- Returns partial results even if some directories are inaccessible
- Non-fatal errors are collected in `ScanResult.Errors`, not returned as function error

**Example**:

```go
ctx := context.Background()
opts := scanner.ScanOptions{
    FollowSymlinks: true,
    MaxDepth:       0, // unlimited
}

result, err := scanner.Scan(ctx, "/path/to/projects", opts)
if err != nil {
    log.Fatalf("Fatal scan error: %v", err)
}

fmt.Printf("Found %d repositories\n", result.TotalRepos)
for _, repo := range result.Repositories {
    fmt.Printf("  %s\n", repo.Path)
}
```

#### Type: ScanOptions

```go
type ScanOptions struct {
    FollowSymlinks bool // Follow symbolic links once
    MaxDepth       int  // Maximum traversal depth (0 = unlimited)
    SkipHidden     bool // Skip hidden directories (starting with .)
}
```

**Defaults**:

- `FollowSymlinks`: `true`
- `MaxDepth`: `0` (unlimited)
- `SkipHidden`: `false`

#### Function: IsGitRepository

```go
func IsGitRepository(path string) (isRepo bool, isBare bool, err error)
```

**Parameters**:

- `path string`: Path to check

**Returns**:

- `isRepo bool`: True if path contains a Git repository
- `isBare bool`: True if repository is bare (only valid if isRepo is true)
- `error`: Error if path is inaccessible

**Behavior**:

- Checks for `.git` directory (regular repo)
- Checks for `HEAD`, `refs/`, `objects/` (bare repo)
- Returns error only for access issues, not for "not a repo"

**Example**:

```go
isRepo, isBare, err := scanner.IsGitRepository("/path/to/repo")
if err != nil {
    log.Printf("Access error: %v", err)
}
if isRepo {
    if isBare {
        fmt.Println("Bare repository")
    } else {
        fmt.Println("Regular repository")
    }
}
```

### Error Handling

**Fatal Errors** (returned as function error):

- Root path does not exist
- Root path is not accessible (permission denied)
- Context cancelled

**Non-Fatal Errors** (collected in `ScanResult.Errors`):

- Individual directory access denied
- Symbolic link resolution failure
- Individual repository access issues

### Testing Contract

**Unit Tests Must Cover**:

- Detecting regular Git repositories
- Detecting bare repositories
- Following symlinks once
- Preventing symlink loops
- Skipping nested repository contents
- Handling permission denied errors
- Respecting context cancellation
- MaxDepth limiting

## 2. Git Status Library (`internal/gitstatus`)

### Purpose

Extracts Git status information from repositories using go-git library.

### Public API

#### Function: Extract

```go
func Extract(ctx context.Context, repoPath string, opts ExtractOptions) (*models.GitStatus, error)
```

**Parameters**:

- `ctx context.Context`: Context for timeout control (recommended: 5-10 second timeout)
- `repoPath string`: Absolute path to Git repository
- `opts ExtractOptions`: Configuration for status extraction

**Returns**:

- `*models.GitStatus`: Git status information
- `error`: Error if status cannot be extracted (repository not found, corrupted, timeout)

**Behavior**:

- Opens repository using go-git
- Extracts branch name or detects detached HEAD
- Queries remote tracking information (ahead/behind counts)
- Checks for stashed changes
- Checks for uncommitted changes (working directory + staging area)
- Handles bare repositories (HasChanges will be false)
- Respects context timeout
- Returns partial information with error field set if some operations fail

**Example**:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

opts := gitstatus.ExtractOptions{
    CheckRemote: true,
    CheckStash:  true,
}

status, err := gitstatus.Extract(ctx, "/path/to/repo", opts)
if err != nil {
    log.Printf("Status extraction failed: %v", err)
    return
}

fmt.Printf("Branch: %s\n", status.Branch)
if status.HasChanges {
    fmt.Println("Has uncommitted changes")
}
```

#### Type: ExtractOptions

```go
type ExtractOptions struct {
    CheckRemote bool // Query remote tracking info (ahead/behind)
    CheckStash  bool // Check for stashes
}
```

**Defaults**:

- `CheckRemote`: `true`
- `CheckStash`: `true`

**Note**: Disabling checks can improve performance but reduces information completeness.

#### Function: ExtractBatch

```go
func ExtractBatch(ctx context.Context, repoPaths []string, opts ExtractOptions, workers int) map[string]*models.GitStatus
```

**Parameters**:

- `ctx context.Context`: Context for cancellation and timeout
- `repoPaths []string`: List of repository paths to process
- `opts ExtractOptions`: Configuration for status extraction
- `workers int`: Number of concurrent workers (recommended: 10-20)

**Returns**:

- `map[string]*models.GitStatus`: Map of repo path to status (nil value if extraction failed)

**Behavior**:

- Processes repositories concurrently using worker pool
- Each repository has independent timeout (set via ctx or per-repo timeout)
- Returns map with all paths (nil value indicates failure)
- Does not return error (failures are reflected in nil map values)

**Example**:

```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

paths := []string{"/repo1", "/repo2", "/repo3"}
opts := gitstatus.ExtractOptions{CheckRemote: true, CheckStash: true}

statusMap := gitstatus.ExtractBatch(ctx, paths, opts, 10)

for path, status := range statusMap {
    if status == nil {
        fmt.Printf("%s: FAILED\n", path)
    } else {
        fmt.Printf("%s: [%s]\n", path, status.Format())
    }
}
```

### Error Handling

**Errors Returned**:

- Repository not found or not a Git repository
- Repository corrupted (cannot open)
- Context timeout exceeded
- Permission denied

**Partial Success**:

- If branch info succeeds but remote query fails, returns GitStatus with `Error` field set and `HasRemote = false`
- Allows graceful degradation

### Testing Contract

**Unit Tests Must Cover**:

- Extracting branch name
- Detecting detached HEAD
- Counting ahead/behind commits
- Detecting stashes
- Detecting uncommitted changes (modified files)
- Detecting staged changes
- Handling no remote configured
- Handling bare repositories
- Handling corrupted repositories
- Respecting context timeout
- Concurrent extraction (ExtractBatch)

## 3. Tree Formatter Library (`internal/tree`)

### Purpose

Formats repository information as ASCII tree structure.

### Public API

#### Function: Format

```go
func Format(root *models.TreeNode, opts FormatOptions) (string, error)
```

**Parameters**:

- `root *models.TreeNode`: Root node of the tree structure
- `opts FormatOptions`: Formatting configuration

**Returns**:

- `string`: Formatted tree as string (suitable for stdout)
- `error`: Error if tree is invalid or formatting fails

**Behavior**:

- Recursively traverses tree structure
- Generates ASCII box-drawing characters (├──, └──, │)
- Includes Git status inline with repository names
- Sorts children alphabetically for deterministic output
- Handles indicator formatting (ahead/behind, stashes, changes)
- Applies indentation based on depth

**Example**:

```go
opts := tree.FormatOptions{
    ShowRelativePaths: true,
    IncludeRoot:       true,
}

output, err := tree.Format(treeRoot, opts)
if err != nil {
    log.Fatalf("Format error: %v", err)
}

fmt.Print(output)
```

#### Type: FormatOptions

```go
type FormatOptions struct {
    ShowRelativePaths bool // Show relative paths instead of names
    IncludeRoot       bool // Include root directory (.) in output
    CompactEmpty      bool // Omit empty parent directories
}
```

**Defaults**:

- `ShowRelativePaths`: `false`
- `IncludeRoot`: `true`
- `CompactEmpty`: `false`

#### Function: Build

```go
func Build(repos []*models.Repository, rootPath string) (*models.TreeNode, error)
```

**Parameters**:

- `repos []*models.Repository`: List of repositories to include in tree
- `rootPath string`: Root path for calculating relative paths

**Returns**:

- `*models.TreeNode`: Root node of constructed tree
- `error`: Error if tree construction fails

**Behavior**:

- Constructs hierarchical tree structure from flat repository list
- Calculates relative paths from rootPath
- Sorts nodes alphabetically
- Sets depth levels correctly
- Marks last children for connector formatting

**Example**:

```go
treeRoot, err := tree.Build(scanResult.Repositories, scanResult.RootPath)
if err != nil {
    log.Fatalf("Tree build error: %v", err)
}

// Now format the tree
output, _ := tree.Format(treeRoot, tree.FormatOptions{})
fmt.Print(output)
```

### Error Handling

**Errors Returned**:

- Tree structure invalid (nil root, inconsistent depth)
- Repository paths inconsistent with root path
- Format generation failure

### Testing Contract

**Unit Tests Must Cover**:

- Single repository formatting
- Multiple repositories at same level
- Nested repositories
- Last child connector formatting (└── vs ├──)
- Indentation at various depths
- Git status inline formatting
- Empty repository list
- Tree structure building from repository list
- Deterministic sorting

## 4. Models Package (`internal/models`)

### Purpose

Shared data structures used across all libraries.

### Public Types

See `data-model.md` for complete type definitions and validation rules.

**Core Types**:

- `Repository`: Represents a Git repository with path and status
- `GitStatus`: Git status information (branch, ahead/behind, changes)
- `TreeNode`: Node in tree structure for formatting
- `ScanResult`: Complete scan result with metadata

**All types provide**:

- `Validate() error`: Validates struct invariants
- Constructor functions (e.g., `NewRepository()`)
- Useful methods (e.g., `GitStatus.Format()`)

### Example Usage

```go
import "github.com/user/gitree/internal/models"

// Create repository
repo := models.NewRepository("/path/to/repo")

// Create Git status
status := &models.GitStatus{
    Branch:     "main",
    IsDetached: false,
    HasRemote:  true,
    Ahead:      2,
    Behind:     1,
    HasStashes: false,
    HasChanges: true,
}

// Validate
if err := status.Validate(); err != nil {
    log.Fatalf("Invalid status: %v", err)
}

// Format
fmt.Println(status.Format()) // Output: [main ↑2 ↓1 *]
```

## Integration Contract

### Typical Usage Flow

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/user/gitree/internal/scanner"
    "github.com/user/gitree/internal/gitstatus"
    "github.com/user/gitree/internal/tree"
)

func main() {
    // 1. Scan for repositories
    ctx := context.Background()
    scanOpts := scanner.ScanOptions{FollowSymlinks: true}

    scanResult, err := scanner.Scan(ctx, ".", scanOpts)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if scanResult.TotalRepos == 0 {
        fmt.Println("No Git repositories found")
        return
    }

    // 2. Extract Git status (batch)
    ctx2, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    paths := make([]string, len(scanResult.Repositories))
    for i, repo := range scanResult.Repositories {
        paths[i] = repo.Path
    }

    statusOpts := gitstatus.ExtractOptions{CheckRemote: true, CheckStash: true}
    statusMap := gitstatus.ExtractBatch(ctx2, paths, statusOpts, 10)

    // 3. Populate repositories with status
    for _, repo := range scanResult.Repositories {
        repo.GitStatus = statusMap[repo.Path]
    }

    // 4. Build tree structure
    treeRoot, err := tree.Build(scanResult.Repositories, scanResult.RootPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // 5. Format and output
    formatOpts := tree.FormatOptions{IncludeRoot: true}
    output, err := tree.Format(treeRoot, formatOpts)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Print(output)
}
```

## Thread Safety

- **scanner**: Safe for concurrent use (stateless functions, or internal synchronization)
- **gitstatus**: Safe for concurrent use (stateless functions)
- **tree**: Safe for concurrent use for read operations; Build/Format are not thread-safe on same TreeNode
- **models**: Data structures are NOT thread-safe; do not mutate concurrently

**Recommendation**: Use `gitstatus.ExtractBatch` for concurrent processing rather than manually managing goroutines.

## Performance Guarantees

- **Scanner**: O(n) where n = number of directories, early exit on Git detection
- **GitStatus Extract**: O(m) where m = number of commits in ahead/behind calculation (bounded by timeout)
- **Tree Format**: O(k log k) where k = number of repositories (due to sorting)

## Version Compatibility

**Semantic Versioning**:

- MAJOR: Breaking API changes (function signature changes, removed functions)
- MINOR: New functions or backward-compatible enhancements
- PATCH: Bug fixes, performance improvements

**Current Version**: 1.0.0 (initial release)
