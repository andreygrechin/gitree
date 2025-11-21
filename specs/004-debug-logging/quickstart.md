# Developer Quickstart: Debug Logging

**Feature**: `004-debug-logging`
**Date**: 2025-11-21

## Overview

Quick reference guide for developers implementing the debug logging feature. Follow this guide to add debug output capability to gitree CLI tool.

## Prerequisites

- Go 1.25.4 installed
- Gitree repository cloned
- Familiarity with cobra CLI framework
- Understanding of go-git library basics

## 5-Minute Implementation Overview

### What You're Building

Add a `--debug` CLI flag that outputs diagnostic information to stderr, helping users understand:

- Why repositories are marked as needing attention
- Which specific files cause non-clean status
- Directory scanning decisions and timing

### Key Changes

1. **CLI Layer** (`cmd/gitree/root.go`): Add --debug flag parsing
2. **Scanner** (`internal/scanner/scanner.go`): Add debug output for directory scanning
3. **Git Status** (`internal/gitstatus/status.go`): Add debug output for status extraction
4. **Spinner**: Disable when debug enabled

### Architecture Flow

```
User: gitree --debug
         ↓
CLI parses flag → debugFlag = true
         ↓
ScanOptions{Debug: true} → scanner.Scan()
         ↓
ExtractOptions{Debug: true} → gitstatus.ExtractBatch()
         ↓
Debug messages → stderr
Tree output → stdout
```

## Implementation Steps

### Step 1: Add Debug Flag to Options Structs (2 min)

**File**: `internal/scanner/scanner.go`

```go
// ScanOptions configures the directory scanning behavior.
type ScanOptions struct {
    RootPath string // Root directory to start scanning from
    Debug    bool   // Enable debug output for scanning operations
}
```

**File**: `internal/gitstatus/status.go`

```go
// ExtractOptions configures the Git status extraction behavior.
type ExtractOptions struct {
    Timeout        time.Duration
    MaxConcurrency int
    Debug          bool   // Enable debug output for status extraction
}
```

### Step 2: Add Debug Helper Function (2 min)

**File**: `internal/scanner/scanner.go` (add at package level)

```go
// debugPrintf formats the message using fmt.Sprintf, adds a "DEBUG: " prefix,
// and outputs it to stderr if debug is enabled.
func debugPrintf(debug bool, format string, args ...interface{}) {
    if !debug {
        return
    }
    message := fmt.Sprintf(format, args...)
    fmt.Fprintf(os.Stderr, "DEBUG: %s\n", message)
}
```

**File**: `internal/gitstatus/status.go` (add same function)

```go
// debugPrintf formats the message using fmt.Sprintf, adds a "DEBUG: " prefix,
// and outputs it to stderr if debug is enabled.
func debugPrintf(debug bool, format string, args ...interface{}) {
    if !debug {
        return
    }
    message := fmt.Sprintf(format, args...)
    fmt.Fprintf(os.Stderr, "DEBUG: %s\n", message)
}
```

### Step 3: Add CLI Flag Parsing (3 min)

**File**: `cmd/gitree/root.go`

Add flag variable alongside existing flags:

```go
var (
    versionFlag bool
    noColorFlag bool
    allFlag     bool
    debugFlag   bool  // Add this line
)
```

Add flag definition in `init()`:

```go
func init() {
    rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
    rootCmd.Flags().BoolVar(&noColorFlag, "no-color", false, "Disable color output")
    rootCmd.Flags().BoolVarP(&allFlag, "all", "a", false,
        "Show all repositories including clean ones (default shows only repos needing attention)")
    rootCmd.Flags().BoolVar(&debugFlag, "debug", false, "Enable debug output")  // Add this line

    rootCmd.PersistentPreRun = handleGlobalFlags
    rootCmd.PreRunE = handleVersionFlag
}
```

Pass debug flag to options in `runGitree()`:

```go
func runGitree(_ *cobra.Command, _ []string) error {
    // ... existing code ...

    // Add debug to scan options
    scanOpts := scanner.ScanOptions{
        RootPath: cwd,
        Debug:    debugFlag,  // Add this line
    }

    // ... after scanning ...

    // Add debug to extract options
    statusOpts := &gitstatus.ExtractOptions{
        Timeout:        defaultTimeout,
        MaxConcurrency: maxConcurrentRequests,
        Debug:          debugFlag,  // Add this line
    }

    // ... rest of function ...
}
```

### Step 4: Disable Spinner When Debug Enabled (2 min)

**File**: `cmd/gitree/root.go` in `runGitree()`

```go
// Initialize spinner
s := spinner.New(spinner.CharSets[spinnerChar], spinnerDelay)
s.Suffix = " Scanning repositories..."
s.Writer = os.Stderr

// Only start spinner if debug is disabled
if !debugFlag {
    s.Start()
}

// ... later in function, update stops ...

if !debugFlag {
    s.Stop()
}
```

**Update all 3 spinner.Stop() calls in runGitree()** to be conditional:

- Line ~126: After scan error
- Line ~155: After status extraction warning
- Line ~183: Before tree output

### Step 5: Add Scanner Debug Output (5 min)

**File**: `internal/scanner/scanner.go` in `walkFunc()` method

Add debug output at key decision points:

```go
func (s *scanner) walkFunc(ctx context.Context, path string, d fs.DirEntry, err error) error {
    // ... existing context check ...

    // Handle permission errors (non-fatal)
    if err != nil {
        if os.IsPermission(err) {
            debugPrintf(s.opts.Debug, "Skipping %s: permission denied", path)  // Add this
            s.errors = append(s.errors, fmt.Errorf("permission denied: %s: %w", path, errPermissionDenied))
            return fs.SkipDir
        }
        return err
    }

    // Only process directories
    if !d.IsDir() {
        return nil
    }

    debugPrintf(s.opts.Debug, "Entering directory: %s", path)  // Add this

    s.dirCount++

    // Check for symlink loops
    shouldVisit, isSymlink, err := s.shouldVisit(path)
    if err != nil {
        debugPrintf(s.opts.Debug, "Skipping %s: %v", path, err)  // Add this
        s.errors = append(s.errors, fmt.Errorf("error checking path %s: %w", path, err))
        return fs.SkipDir
    }
    if !shouldVisit {
        debugPrintf(s.opts.Debug, "Skipping %s: already visited (symlink loop)", path)  // Add this
        return fs.SkipDir
    }

    // Check if this directory is a Git repository
    isRepo, isBare := IsGitRepository(path)
    if isRepo {
        repoType := "regular"
        if isBare {
            repoType = "bare"
        }
        debugPrintf(s.opts.Debug, "Found git repository: %s (%s)", path, repoType)  // Add this

        repo := &models.Repository{
            Path:      path,
            Name:      filepath.Base(path),
            IsBare:    isBare,
            IsSymlink: isSymlink,
        }

        s.repositories = append(s.repositories, repo)

        debugPrintf(s.opts.Debug, "Skipping %s: inside git repository", path)  // Add this
        return fs.SkipDir
    }

    return nil
}
```

**Note**: You'll need to change scanner struct to store opts:

```go
type scanner struct {
    rootPath     string
    opts         ScanOptions  // Change: store full options instead of just rootPath
    repositories []*models.Repository
    errors       []error
    visited      map[uint64]bool
    dirCount     int
}
```

Update `Scan()` function:

```go
s := &scanner{
    rootPath:     absPath,
    opts:         opts,  // Store opts
    repositories: make([]*models.Repository, 0),
    errors:       make([]error, 0),
    visited:      make(map[uint64]bool),
}
```

### Step 6: Add Git Status Debug Output (10 min)

**File**: `internal/gitstatus/status.go` in `extractGitStatus()` function

Add timing and status debug output:

```go
func extractGitStatus(repoPath string, opts *ExtractOptions) (*models.GitStatus, error) {
    startTime := time.Now()  // Add this

    // Open repository
    repo, err := git.PlainOpen(repoPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open repository: %w", err)
    }

    status := &models.GitStatus{}

    // ... existing status extraction ...

    // Debug output at end
    if opts != nil && opts.Debug {
        // Timing (only if >100ms)
        duration := time.Since(startTime)
        if duration > 100*time.Millisecond {
            debugPrintf(true, "Repository %s status extraction: %dms", repoPath, duration.Milliseconds())
        }

        // Status summary
        statusParts := []string{fmt.Sprintf("branch=%s", status.Branch)}
        statusParts = append(statusParts, fmt.Sprintf("hasChanges=%t", status.HasChanges))
        if status.HasRemote {
            statusParts = append(statusParts, fmt.Sprintf("hasRemote=true"))
            if status.Ahead > 0 {
                statusParts = append(statusParts, fmt.Sprintf("ahead=%d", status.Ahead))
            }
            if status.Behind > 0 {
                statusParts = append(statusParts, fmt.Sprintf("behind=%d", status.Behind))
            }
        } else {
            statusParts = append(statusParts, "hasRemote=false")
        }
        if status.HasStashes {
            statusParts = append(statusParts, "hasStashes=true")
        }

        debugPrintf(true, "Repository %s: %s", repoPath, strings.Join(statusParts, ", "))
    }

    return status, nil
}
```

**Update function signature** to accept opts:

```go
func extractGitStatus(repoPath string, opts *ExtractOptions) (*models.GitStatus, error) {
```

**Update call sites** in `Extract()`:

```go
go func() {
    status, err := extractGitStatus(repoPath, opts)  // Pass opts
    // ...
}()
```

### Step 7: Add File Listing Debug Output (10 min)

**File**: `internal/gitstatus/status.go` in `extractUncommittedChanges()` function

Add file categorization and listing:

```go
func extractUncommittedChanges(repo *git.Repository, status *models.GitStatus, opts *ExtractOptions) error {
    worktree, err := repo.Worktree()
    if err != nil {
        status.HasChanges = false
        return err
    }

    // ... existing gitignore loading ...

    wtStatus, err := worktree.Status()
    if err != nil {
        return fmt.Errorf("failed to get worktree status: %w", err)
    }

    status.HasChanges = !wtStatus.IsClean()

    // Debug file listing
    if opts != nil && opts.Debug && status.HasChanges {
        // Categorize files
        modifiedFiles := []string{}
        untrackedFiles := []string{}
        stagedFiles := []string{}
        deletedFiles := []string{}

        for filename, fileStatus := range wtStatus {
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

        // Print with truncation
        printFileList := func(category string, files []string) {
            if len(files) == 0 {
                return
            }
            if len(files) <= 20 {
                debugPrintf(true, "%s files (%d): %s", category, len(files), strings.Join(files, ", "))
            } else {
                debugPrintf(true, "%s files (%d): %s", category, len(files), strings.Join(files[:20], ", "))
                debugPrintf(true, "...and %d more %s files", len(files)-20, strings.ToLower(category))
            }
        }

        printFileList("Modified", modifiedFiles)
        printFileList("Untracked", untrackedFiles)
        printFileList("Staged", stagedFiles)
        printFileList("Deleted", deletedFiles)
    }

    return nil
}
```

**Update function signature**:

```go
func extractUncommittedChanges(repo *git.Repository, status *models.GitStatus, opts *ExtractOptions) error {
```

**Update call site** in `extractGitStatus()`:

```go
if err := extractUncommittedChanges(repo, status, opts); err != nil {  // Pass opts
    // ...
}
```

### Step 8: Test Locally (5 min)

Build and test:

```bash
# Build
make build

# Test basic debug
./bin/gitree --debug

# Test with other flags
./bin/gitree --debug --all
./bin/gitree --debug --no-color
./bin/gitree --debug 2>debug.log

# Test without debug (verify no overhead)
./bin/gitree
```

## Quick Reference

### File Changes Summary

| File | Changes | Lines Added |
|------|---------|-------------|
| `cmd/gitree/root.go` | Add flag, pass to options, conditional spinner | ~15 |
| `internal/scanner/scanner.go` | Add Debug field, helper func, debug output | ~30 |
| `internal/gitstatus/status.go` | Add Debug field, helper func, timing, file listing | ~60 |

**Total**: ~105 lines of new code

### Key Functions to Modify

1. `cmd/gitree/root.go`:
   - `init()` - add flag definition
   - `runGitree()` - pass debug to options, conditional spinner

2. `internal/scanner/scanner.go`:
   - `ScanOptions` struct - add Debug field
   - `walkFunc()` - add debug output
   - Add `debugPrintf()` helper

3. `internal/gitstatus/status.go`:
   - `ExtractOptions` struct - add Debug field
   - `extractGitStatus()` - add timing and status debug
   - `extractUncommittedChanges()` - add file listing
   - Add `debugPrintf()` helper

### Testing Checklist

- [ ] `--help` shows debug flag
- [ ] Debug output appears with `--debug`
- [ ] No debug output without `--debug`
- [ ] Stdout tree unchanged by debug flag
- [ ] `--debug --no-color` works
- [ ] `--debug --all` works
- [ ] Spinner not shown when debug enabled
- [ ] Timing shown for slow repos (>100ms)
- [ ] File truncation at 20 per category

## Common Pitfalls

### ❌ Pitfall 1: Forgetting to pass opts

**Problem**: Debug output never appears because opts not passed through

**Solution**: Verify opts parameter passed to all functions that need it:

- `extractGitStatus(repoPath, opts)`
- `extractUncommittedChanges(repo, status, opts)`

### ❌ Pitfall 2: Debug output to stdout

**Problem**: Debug messages mixed with tree output, breaks piping

**Solution**: Always use `os.Stderr` for debug output:

```go
fmt.Fprintf(os.Stderr, "DEBUG: %s\n", message)
```

### ❌ Pitfall 3: Not respecting color flag

**Problem**: Debug output has color codes even with --no-color

**Solution**: Use `color.NoColor` global variable (already set by handleGlobalFlags):

```go
// No action needed - color.NoColor is automatically respected
// by any code using fatih/color package
```

### ❌ Pitfall 4: Spinner still showing with debug

**Problem**: Spinner interferes with debug output, makes it hard to read

**Solution**: Wrap all spinner operations in conditionals:

```go
if !debugFlag {
    s.Start()
}
// ...
if !debugFlag {
    s.Stop()
}
```

### ❌ Pitfall 5: Not nil-checking opts

**Problem**: Panic when opts is nil

**Solution**: Always check before accessing Debug field:

```go
if opts != nil && opts.Debug {
    debugPrintf(true, "message")
}
```

## Next Steps

After implementing:

1. **Run tests**: `make test`
2. **Run linters**: `make lint`
3. **Test manually**: Try debug flag on real repositories
4. **Write tests**: See `tasks.md` for test specifications (Phase 2)
5. **Update documentation**: CHANGELOG, README if needed

## Architecture Decisions Reference

### Why boolean flag instead of verbosity levels?

**Decision**: Single `--debug` boolean flag

**Rationale**:

- YAGNI principle (Constitution V: Simplicity)
- Spec explicitly requires "Debug level will be boolean (on/off)"
- Can add verbosity levels later if needed

### Why stderr for debug output?

**Decision**: All debug output to stderr, tree to stdout

**Rationale**:

- Standard Unix convention: data to stdout, diagnostics to stderr
- Allows separation: `gitree --debug >tree.txt 2>debug.log`
- Spec requirement FR-005: "Debug output MUST be written to stderr"

### Why disable spinner when debug enabled?

**Decision**: Skip spinner.Start() when debug flag is true

**Rationale**:

- Both write to stderr, causes jumbled output
- Debug provides its own progress indication
- Spec requirement FR-010: "Progress spinner MUST be disabled entirely when debug mode is active"

### Why fmt.Fprintln instead of logging framework?

**Decision**: Use fmt.Fprintln(os.Stderr, ...) directly

**Rationale**:

- Spec requirement FR-006: "MUST use simple print statements (fmt.Fprintln/fmt.Fprint)"
- Simplicity principle - no dependencies needed
- Human-readable output, not structured logs

## Resources

- **Feature Spec**: `specs/004-debug-logging/spec.md`
- **Research**: `specs/004-debug-logging/research.md`
- **Data Model**: `specs/004-debug-logging/data-model.md`
- **CLI Contract**: `specs/004-debug-logging/contracts/cli-interface.md`
- **Implementation Plan**: `specs/004-debug-logging/plan.md`
- **Cobra Docs**: <https://github.com/spf13/cobra>
- **go-git Docs**: <https://pkg.go.dev/github.com/go-git/go-git/v5>

## Estimated Time

- **Total implementation**: ~45 minutes
- **Testing**: ~30 minutes
- **Code review fixes**: ~15 minutes

**Total**: ~1.5 hours for complete feature
