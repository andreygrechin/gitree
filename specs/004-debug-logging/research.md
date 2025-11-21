# Research: Debug Logging

**Feature**: `004-debug-logging`
**Date**: 2025-11-21

## Overview

This document consolidates research findings for implementing debug logging in gitree. The feature adds a `--debug` CLI flag that outputs diagnostic information to stderr explaining git status determination and directory scanning behavior.

## Research Questions & Decisions

### 1. How to pass debug flag through the application

**Decision**: Add debug flag to context or create a Config/Options struct passed through the call chain

**Rationale**:

- Gitree already uses context for cancellation (`context.Context` passed to `Scan()` and `ExtractBatch()`)
- Scanner and gitstatus packages accept options structs (`ScanOptions`, `ExtractOptions`)
- Extending these options structs is the most Go-idiomatic approach
- Avoids global state and keeps debug mode explicit in function signatures

**Implementation approach**:

1. Add `Debug bool` field to `ScanOptions` in `internal/scanner/scanner.go`
2. Add `Debug bool` field to `ExtractOptions` in `internal/gitstatus/status.go`
3. Parse `--debug` flag in `cmd/gitree/root.go` and pass through options structs
4. Pass debug flag down from scanner to git status extraction

**Alternatives considered**:

- Context values: Less explicit, harder to test, discouraged in Go best practices
- Global variable: Violates constitution simplicity principle, makes testing difficult
- Logging framework: Explicitly rejected by spec requirement FR-006

### 2. Debug output format and conventions

**Decision**: Use `"DEBUG: "` prefix for all debug output to stderr

**Rationale**:

- Distinguishes debug output from regular status messages
- Simple and searchable pattern
- Aligns with spec assumption: "Debug output will use a simple prefix like 'DEBUG:'"
- Stderr separation already established in codebase (`s.Writer = os.Stderr` for spinner)

**Format examples**:

```
DEBUG: Scanning directory: /path/to/dir
DEBUG: Found git repository: /path/to/repo (regular)
DEBUG: Found git repository: /path/to/bare.git (bare)
DEBUG: Skipping directory: /path/to/repo/.git (inside git repo)
DEBUG: Repository /path/to/repo status extraction: 45ms
DEBUG: Repository /path/to/repo: branch=main, hasChanges=true
DEBUG: Modified files: file1.go, file2.go, file3.go
DEBUG: Untracked files: newfile.txt
DEBUG: ...and 15 more modified files
```

**Color handling**:

- Debug output will respect existing `color.NoColor` flag from `fatih/color`
- Already handled by `handleGlobalFlags()` in root.go
- No additional code needed - `color.NoColor` is a package-level variable

### 3. How to detect "slow" operations for timing output

**Decision**: Show timing for all git status operations that exceed 100ms

**Rationale**:

- Spec SC-003 requires: "timing information for all git status operations taking longer than 100 milliseconds"
- 100ms threshold balances signal-to-noise ratio
- Fast operations (<100ms) don't need timing shown to reduce clutter
- Already have `time.Since()` usage in scanner (lines 73, 118 in scanner.go)

**Implementation approach**:

1. Record start time before `extractGitStatus()` call
2. Calculate duration after completion
3. If duration > 100ms, output: `DEBUG: Repository {path} status extraction: {duration}ms`
4. Format duration in milliseconds with 0 decimal places for readability

**Alternatives considered**:

- Show all timing: Too verbose, clutters output for fast repos
- Different threshold (50ms, 200ms): 100ms specified in spec, good balance

### 4. File listing limits and truncation

**Decision**: Show first 20 files per category, then add "...and N more files" summary line

**Rationale**:

- Spec FR-003 specifies: "When file count exceeds 20 per category, show first 20 files followed by '...and N more files' summary line"
- Categories are: Modified, Untracked, Staged, Deleted
- Per-category limit prevents overwhelming output when one category has many files
- Example: 25 modified + 5 untracked shows all 5 untracked, 20 of 25 modified, then "...and 5 more modified files"

**Implementation approach**:

1. Get worktree status from go-git (already done in `extractUncommittedChanges()`)
2. Iterate through status file entries, categorize by staging state
3. For each category with >20 files:
   - Print first 20 file paths
   - Print summary: `DEBUG: ...and {N} more {category} files`
4. Categories from `go-git/v5.Status`:
   - Modified: `Staging == git.Unmodified && Worktree == git.Modified`
   - Untracked: `Staging == git.Untracked && Worktree == git.Untracked`
   - Staged: `Staging != git.Unmodified && Staging != git.Untracked`
   - Deleted: `Worktree == git.Deleted`

### 5. Spinner interaction with debug output

**Decision**: Disable spinner completely when debug mode is active

**Rationale**:

- Spec FR-010: "Progress spinner MUST be disabled entirely when debug mode is active"
- Spec edge case: "Spinner is disabled entirely when debug mode is active"
- Spinner writes to stderr, debug output writes to stderr - conflict causes jumbled output
- Debug output provides its own progress indication

**Implementation approach**:

1. Check debug flag before creating spinner in `runGitree()`
2. Skip `s.Start()` and `s.Stop()` calls when debug enabled
3. Still create spinner object to avoid nil pointer checks throughout code
4. Alternative: Use null/no-op spinner pattern

**Code location**: `cmd/gitree/root.go` lines 110-113, 139, 155, 183

### 6. Best practices for debug output in CLI tools

**Research findings**:

- **Separation**: Diagnostic output to stderr, data output to stdout (already done in gitree)
- **Verbosity levels**: Single boolean flag sufficient for gitree's scope (per spec assumption)
- **Performance**: Debug checks should be fast `if debug { ... }` branches are near-zero cost when false
- **Testability**: Debug output should be captured in tests to verify correctness
- **Color handling**: Should respect NO_COLOR environment variable (already implemented)

**Go standard library patterns**:

- `flag` package: Uses stderr for help output, stdout for program data
- `log` package: Defaults to stderr for log messages
- `fmt.Fprintln(os.Stderr, ...)`: Standard Go idiom for diagnostic output

**Examples from popular Go CLI tools**:

- `kubectl --v=5`: Numeric verbosity levels (more complex than needed)
- `docker --debug`: Boolean flag, outputs to stderr (matches our approach)
- `git --verbose`: Boolean flag, inline with regular output (not suitable - we need separation)

### 7. Directory scanning debug details

**Decision**: Output debug messages for key scanning decisions: directory entry, repo detection, skipping

**Rationale**:

- Spec FR-004: "Debug output MUST show directory scanning decisions (which directories are entered, which are skipped, and why)"
- Helps diagnose performance issues (why is scanning slow?)
- Helps understand why repos might be missed (symlink loops, permissions)
- Current scanner already tracks: visited inodes, permission errors, repo detection

**Debug points to add in scanner.walkFunc()**:

1. When entering a directory: `DEBUG: Entering directory: {path}`
2. When skipping due to permission: `DEBUG: Skipping {path}: permission denied`
3. When detecting repo: `DEBUG: Found git repository: {path} (regular|bare)`
4. When skipping due to symlink loop: `DEBUG: Skipping {path}: already visited (symlink loop)`
5. When skipping repo contents: `DEBUG: Skipping {path}: inside git repository`

**Verbosity concern**: For large directory trees, this could be very verbose

- Mitigation: This is the purpose of debug mode - detailed diagnostic output
- Users can pipe to file or grep if needed: `gitree --debug 2>debug.log`
- Only enabled when explicitly requested via --debug flag

## Technology Stack

**Existing dependencies (no new dependencies needed)**:

- `github.com/spf13/cobra v1.10.1` - CLI flag parsing (add --debug flag)
- `github.com/go-git/go-git/v5 v5.16.3` - Git operations (already exposes Status() for file listing)
- `github.com/fatih/color v1.18.0` - Color output (already handles --no-color)
- `github.com/briandowns/spinner v1.23.2` - Progress spinner (conditionally disable)

**Go standard library**:

- `fmt` - Fprintln/Fprint for debug output (per spec FR-006)
- `os` - stderr access via `os.Stderr`
- `time` - Duration timing for operations

## Integration Points

### CLI Layer (`cmd/gitree/root.go`)

- Add `debugFlag bool` variable alongside existing `versionFlag`, `noColorFlag`, `allFlag`
- Add flag definition: `rootCmd.Flags().BoolVar(&debugFlag, "debug", false, "Enable debug output")`
- Pass debug flag to `ScanOptions` and `ExtractOptions`
- Conditionally skip spinner start/stop when debug enabled

### Scanner Layer (`internal/scanner/scanner.go`)

- Add `Debug bool` field to `ScanOptions` struct
- Add debug output in `walkFunc()` for scanning decisions
- Pass debug flag to any nested calls that need it

### Git Status Layer (`internal/gitstatus/status.go`)

- Add `Debug bool` field to `ExtractOptions` struct
- Add debug output in `extractGitStatus()` for status determination
- Add timing measurement around status extraction
- Add file listing in `extractUncommittedChanges()` when changes detected
- Handle file listing truncation at 20 files per category

### Models Layer (`internal/models/repository.go`)

- No changes needed - models are data structures only

### Tree Formatter Layer (`internal/tree/formatter.go`)

- No changes needed per spec: "Debug output for tree formatting logic (focus on scanning and status)" is out of scope

## Open Questions Resolved

All questions from the implementation plan have been resolved:

1. ✅ **How to pass debug flag**: Via options structs (`ScanOptions`, `ExtractOptions`)
2. ✅ **Debug output format**: `"DEBUG: "` prefix to stderr
3. ✅ **Timing threshold**: 100ms (per spec)
4. ✅ **File listing limits**: 20 per category (per spec)
5. ✅ **Spinner interaction**: Disable completely when debug enabled (per spec)
6. ✅ **Go best practices**: Stderr for diagnostics, fmt.Fprintln, respect NO_COLOR
7. ✅ **Scanning debug details**: Log directory entry, repo detection, skip reasons

## Performance Considerations

**Debug disabled (default)**:

- Near-zero overhead: `if opts.Debug { ... }` check is negligible
- No string formatting occurs when debug is false
- No additional allocations
- Target: <1% overhead when debug disabled

**Debug enabled**:

- Acceptable overhead: stderr writes are buffered
- File listing in git status is already computed by go-git
- Timing measurement adds two `time.Now()` calls per repo (~100ns each)
- String formatting only when debug enabled
- Target: <50ms additional per repository (per spec)

## Test Strategy

**Unit tests** (in `*_test.go` files):

1. Test debug flag parsing in CLI
2. Test debug output appears when flag enabled
3. Test debug output absent when flag disabled
4. Test debug output respects --no-color flag
5. Test file listing truncation at 20 files
6. Test timing output for slow operations (>100ms)
7. Test spinner is not started when debug enabled

**Integration tests** (`tests/integration/`):

1. Run gitree with --debug on test directory tree
2. Capture stderr output
3. Verify expected debug messages present
4. Verify stdout tree output unaffected

**Test fixtures**:

- Create test repos with various states: clean, modified, untracked, staged
- Create test repo with >20 modified files for truncation test
- Create directory tree with symlinks for scanning debug test
- Use `t.TempDir()` for isolation

## Implementation Order

**Phase 1: Infrastructure** (supports TDD principle)

1. Add `Debug bool` to `ScanOptions` and `ExtractOptions` structs
2. Add `--debug` flag to CLI flag parsing
3. Wire debug flag through options from CLI to scanner to gitstatus
4. Add helper function for debug output: `debugPrintf(debug bool, format string, args ...interface{})`

**Phase 2: Scanner Debug Output**

1. Add debug output for directory entry in `walkFunc()`
2. Add debug output for repo detection
3. Add debug output for skip decisions (permission, symlink, repo contents)

**Phase 3: Git Status Debug Output**

1. Add timing measurement in `Extract()` function
2. Add debug output for status determination (branch, remote, ahead/behind)
3. Implement file listing in `extractUncommittedChanges()` with categorization
4. Implement 20-file truncation per category

**Phase 4: Spinner Integration**

1. Conditionally disable spinner start/stop in `runGitree()`

**Phase 5: Testing**

1. Write unit tests for debug output
2. Write integration tests with test fixtures
3. Verify performance targets met

## References

- Feature spec: `/Users/andrey/repos/gitree/specs/004-debug-logging/spec.md`
- Gitree constitution: `/Users/andrey/repos/gitree/.specify/memory/constitution.md`
- go-git Status API: <https://pkg.go.dev/github.com/go-git/go-git/v5#Status>
- Go fmt package: <https://pkg.go.dev/fmt>
- Cobra CLI framework: <https://github.com/spf13/cobra>
