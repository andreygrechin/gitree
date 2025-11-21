# Data Model: Debug Logging

**Feature**: `004-debug-logging`
**Date**: 2025-11-21

## Overview

This document defines the data structures and types needed for the debug logging feature. Since this feature primarily adds diagnostic output rather than new business entities, the data model is minimal and focuses on configuration structures.

## Data Structures

### 1. ScanOptions (Modified)

**Location**: `internal/scanner/scanner.go`

**Purpose**: Configure directory scanning behavior, extended to include debug flag

```go
// ScanOptions configures the directory scanning behavior.
type ScanOptions struct {
    RootPath string // Root directory to start scanning from
    Debug    bool   // Enable debug output for scanning operations
}
```

**Fields**:

- `RootPath string` - Existing field, root directory to scan
- `Debug bool` - **NEW** - Enables debug output during scanning

**Validation**:

- `RootPath` must be non-empty (existing validation)
- `RootPath` must be an absolute path (existing validation)
- `Debug` is boolean, no validation needed

**Usage**:

- Created in `cmd/gitree/root.go` runGitree()
- Passed to `scanner.Scan(ctx, opts)`
- Accessed in `scanner.walkFunc()` to conditionally output debug messages

### 2. ExtractOptions (Modified)

**Location**: `internal/gitstatus/status.go`

**Purpose**: Configure git status extraction behavior, extended to include debug flag

```go
// ExtractOptions configures the Git status extraction behavior.
type ExtractOptions struct {
    // Timeout is the maximum time to spend extracting status for a single repository
    Timeout time.Duration

    // MaxConcurrency limits the number of repositories processed concurrently in ExtractBatch
    MaxConcurrency int

    // Debug enables debug output for status extraction operations
    Debug bool
}
```

**Fields**:

- `Timeout time.Duration` - Existing field, timeout per repository
- `MaxConcurrency int` - Existing field, concurrent workers limit
- `Debug bool` - **NEW** - Enables debug output during status extraction

**Validation**:

- `Timeout` >= 0 (existing validation)
- `MaxConcurrency` > 0 (existing validation)
- `Debug` is boolean, no validation needed

**Usage**:

- Created in `cmd/gitree/root.go` runGitree()
- Passed to `gitstatus.ExtractBatch(ctx, repoMap, opts)`
- Accessed in `extractGitStatus()` to conditionally output debug messages

### 3. Debug Output Helper (New)

**Location**: `internal/scanner/scanner.go` and `internal/gitstatus/status.go`

**Purpose**: Helper function to conditionally output debug messages

```go
// debugPrintln outputs a debug message to stderr if debug is enabled.
// Format follows fmt.Fprintf conventions.
func debugPrintln(debug bool, format string, args ...interface{}) {
    if !debug {
        return
    }

    message := fmt.Sprintf(format, args...)
    fmt.Fprintf(os.Stderr, "DEBUG: %s\n", message)
}
```

**Parameters**:

- `debug bool` - Whether debug mode is enabled
- `format string` - Printf-style format string
- `args ...interface{}` - Printf-style arguments

**Behavior**:

- Returns immediately if `debug` is false (near-zero overhead)
- Formats message using `fmt.Sprintf`
- Writes to `os.Stderr` with "DEBUG: " prefix
- Adds newline at end
- No color handling needed - relies on `color.NoColor` package variable

**Alternative**: Could be method on options structs, but free function is simpler and matches spec requirement for `fmt.Fprintln`

## File Status Categories (Go-Git Mapping)

**Purpose**: Categorize uncommitted changes for debug output

**Source**: `github.com/go-git/go-git/v5` Status API

```go
// From go-git/v5: Status represents the current status of a Worktree
type Status map[string]*FileStatus

// FileStatus contains the status of a file in the worktree
type FileStatus struct {
    Staging  StatusCode
    Worktree StatusCode
}

// StatusCode status code of a file in the worktree
type StatusCode byte

const (
    Unmodified         StatusCode = ' '
    Untracked          StatusCode = '?'
    Modified           StatusCode = 'M'
    Added              StatusCode = 'A'
    Deleted            StatusCode = 'D'
    Renamed            StatusCode = 'R'
    Copied             StatusCode = 'C'
    UpdatedButUnmerged StatusCode = 'U'
)
```

**Category Mapping for Debug Output**:

| Category | Condition | Example |
|----------|-----------|---------|
| **Modified** | `Worktree == Modified && Staging == Unmodified` | File edited, not staged |
| **Untracked** | `Staging == Untracked && Worktree == Untracked` | New file, not tracked |
| **Staged** | `Staging != Unmodified && Staging != Untracked` | Changes added to index |
| **Deleted** | `Worktree == Deleted` | File deleted from working dir |

**Implementation Logic**:

```go
// Pseudocode for categorizing files
for filename, fileStatus := range worktreeStatus {
    if fileStatus.Worktree == git.Modified && fileStatus.Staging == git.Unmodified {
        modifiedFiles = append(modifiedFiles, filename)
    } else if fileStatus.Staging == git.Untracked && fileStatus.Worktree == git.Untracked {
        untrackedFiles = append(untrackedFiles, filename)
    } else if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked {
        stagedFiles = append(stagedFiles, filename)
    } else if fileStatus.Worktree == git.Deleted {
        deletedFiles = append(deletedFiles, filename)
    }
}
```

## State Transitions

### Debug Mode Enablement

```
User runs CLI
    ↓
Cobra parses flags → debugFlag bool
    ↓
debugFlag passed to ScanOptions.Debug
    ↓
debugFlag passed to ExtractOptions.Debug
    ↓
Debug output conditionally emitted based on Debug bool
```

**Key Property**: Debug state is immutable after initialization (set once at start, never changed)

## Data Flow

```
Command Line
    ↓
    [--debug flag]
    ↓
cmd/gitree/root.go (runGitree)
    ├─→ ScanOptions{Debug: debugFlag}
    │       ↓
    │   scanner.Scan(ctx, opts)
    │       ↓
    │   scanner.walkFunc() checks opts.Debug
    │       ↓
    │   debugPrintln(opts.Debug, "Entering directory: %s", path)
    │
    └─→ ExtractOptions{Debug: debugFlag}
            ↓
        gitstatus.ExtractBatch(ctx, repos, opts)
            ↓
        extractGitStatus() checks opts.Debug
            ↓
        debugPrintln(opts.Debug, "Repository %s: branch=%s", path, branch)
```

## Relationships

### Existing Models (Unchanged)

The following existing models are **not modified** by this feature:

- `models.Repository` - Repository metadata
- `models.GitStatus` - Git status information
- `models.ScanResult` - Scanning results
- `models.TreeNode` - Tree structure for output

**Rationale**: Debug feature only adds diagnostic output, doesn't change data capture or storage

### New Relationships

- `ScanOptions` → scanner behavior (debug output enabled/disabled)
- `ExtractOptions` → status extraction behavior (debug output enabled/disabled)
- Both options structs are independent; debug flag set on both from CLI layer

## Performance Implications

### Memory

**Debug disabled**: Zero additional memory overhead

- Boolean flag adds 1 byte per options struct (negligible)
- No debug strings allocated when debug is false

**Debug enabled**: Minimal memory overhead

- Debug strings allocated only when outputted
- File lists capped at 20 per category (maximum ~2KB per repository for file paths)
- Stderr buffering handled by OS

### CPU

**Debug disabled**: Near-zero CPU overhead

- `if debug { ... }` check is single boolean comparison (~1 CPU cycle)
- No string formatting occurs
- Target: <1% overhead

**Debug enabled**: Acceptable CPU overhead

- String formatting: ~1-10μs per debug message
- Stderr write: ~10-100μs per message (buffered)
- File categorization: already computed by go-git, just iterate and print
- Target: <50ms per repository (per spec)

### I/O

**Debug disabled**: Zero I/O overhead

**Debug enabled**: Stderr writes

- Stderr is buffered by OS (typically 4KB-8KB buffer)
- Messages flushed on newline or buffer full
- Does not block scanner or status extraction (async I/O)

## Validation Rules

### ScanOptions.Debug

- Type: `bool`
- Valid values: `true` or `false`
- Default: `false`
- No validation needed (type-safe)

### ExtractOptions.Debug

- Type: `bool`
- Valid values: `true` or `false`
- Default: `false`
- No validation needed (type-safe)

### Debug Output Format

- Prefix: Must be exactly `"DEBUG: "` (uppercase, colon, space)
- Output: Must go to `os.Stderr`
- Newline: Must end with `\n`
- Color: Must respect `color.NoColor` global variable

## Error Handling

**Debug output failures**:

- Stderr write failures are **non-fatal**
- Errors ignored - debug is best-effort diagnostic
- Main functionality (scanning, status extraction) continues even if debug output fails

**Rationale**:

- Debug is supplementary, not critical
- Stderr write failures are rare (disk full, broken pipe)
- User experience priority: show tree output even if debug fails

## Testing Considerations

### Test Data Needed

**For scanner debug output**:

1. Directory with nested repositories
2. Directory with symlinks (to test loop detection debug output)
3. Directory with permission-denied subdirectory
4. Bare repository
5. Regular repository

**For git status debug output**:

1. Repository with 5 modified files (below threshold)
2. Repository with 25 modified files (above threshold - test truncation)
3. Repository with mixed status: modified, untracked, staged, deleted
4. Repository with slow status extraction (mock 150ms delay to test timing output)
5. Clean repository (no debug output for files)

### Test Assertions

1. When `Debug: false` → no "DEBUG:" strings in stderr
2. When `Debug: true` → "DEBUG:" strings present in stderr for expected events
3. Stdout output unchanged regardless of debug flag
4. File truncation at 20 per category
5. Timing output only for operations >100ms
6. Spinner not started when debug enabled

## Future Considerations

**Not in current scope (potential future enhancements)**:

- Multiple verbosity levels (--debug, --verbose, --trace)
- Debug output filtering by package or operation type
- Debug output to file instead of stderr
- Structured JSON debug output for machine parsing
- Performance profiling integration

**Why deferred**: YAGNI principle (Constitution Principle V: Simplicity)

- Current boolean flag sufficient for stated requirements
- Can add complexity later if demonstrated need arises
