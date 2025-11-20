# Data Model: CLI Framework with Command-Line Flags

**Feature**: 003-cobra-cli-flags
**Date**: 2025-11-17

## Overview

This feature primarily extends existing data structures rather than creating new ones. The data model changes are minimal, focusing on configuration structures for filtering options and version information.

**Note**: The default behavior is to show only repositories needing attention (filtering enabled). The `--all` flag disables filtering to show all repositories.

## Entities

### 1. FilterOptions (New)

**Package**: `internal/cli`

**Purpose**: Configuration for repository filtering behavior. By default, filtering is enabled to show only repos needing attention; the --all flag disables filtering.

**Fields**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ShowAll` | `bool` | When true, disables filtering (shows all repos including clean ones). Default: false | N/A (boolean) |

**Relationships**:

- Used by `cli.FilterRepositories()` function
- Operates on `[]*models.Repository` collections

**State Transitions**: N/A (simple configuration struct)

**Example**:

```go
type FilterOptions struct {
    ShowAll bool  // Default: false (filtering enabled)
}

// Default usage (show only repos needing attention)
opts := FilterOptions{ShowAll: false}
filtered := cli.FilterRepositories(repos, opts)

// With --all flag (show all repos)
opts := FilterOptions{ShowAll: true}
all := cli.FilterRepositories(repos, opts)  // Returns unfiltered list
```

---

### 2. VersionInfo (New)

**Package**: `cmd/gitree` (main package)

**Purpose**: Holds build-time version metadata for display with --version flag.

**Fields**:

| Field | Type | Description | Validation | Default Value |
|-------|------|-------------|------------|---------------|
| `Version` | `string` | Semantic version (from git tags) | Non-empty | `"dev"` |
| `Commit` | `string` | Git commit hash (short) | Non-empty | `"none"` |
| `BuildTime` | `string` | ISO 8601 timestamp of build | RFC3339 format | `"unknown"` |

**Relationships**:

- Populated by Go linker at build time via `-ldflags`
- Displayed by version command

**State Transitions**: Immutable (set at build time)

**Example**:

```go
var (
    version   = "dev"      // -ldflags override
    commit    = "none"     // -ldflags override
    buildTime = "unknown"  // -ldflags override
)

// Usage in version command
func displayVersion() {
    fmt.Printf("gitree version %s\n", version)
    fmt.Printf("  commit: %s\n", commit)
    fmt.Printf("  built:  %s\n", buildTime)
}
```

---

### 3. Repository (Existing - No Changes)

**Package**: `internal/models`

**Status**: No modifications required

**Relevant Fields for Filtering**:

The default filtering logic (show only repos needing attention) reads existing `Repository.GitStatus` fields:

| Field Path | Type | Used For |
|------------|------|----------|
| `GitStatus.Branch` | `string` | Check if main/master (FR-008) |
| `GitStatus.HasUncommittedChanges` | `bool` | Check for uncommitted work (FR-008) |
| `GitStatus.Stashes` | `int` | Check for stashed changes (FR-008) |
| `GitStatus.HasRemote` | `bool` | Check for remote tracking (FR-008) |
| `GitStatus.Ahead` | `int` | Check ahead count (FR-008) |
| `GitStatus.Behind` | `int` | Check behind count (FR-008) |
| `GitStatus.IsDetached` | `bool` | Check detached HEAD (FR-009) |
| `GitStatus.Error` | `error` | Fail-safe for status extraction failures (FR-009) |

**Note**: No schema changes needed. The filtering logic operates on the existing `GitStatus` structure.

---

### 4. GitStatus (Existing - Verification Required)

**Package**: `internal/models`

**Status**: Verify all required fields exist

**Required Fields for Clean State Determination**:

Based on FR-008 and FR-009, the following fields must be available in `GitStatus`:

```go
type GitStatus struct {
    Branch                string  // Branch name or "HEAD" if detached
    HasUncommittedChanges bool    // Staged or unstaged changes
    Stashes               int     // Count of stashed changes
    HasRemote             bool    // Remote tracking configured
    Ahead                 int     // Commits ahead of remote
    Behind                int     // Commits behind of remote
    IsDetached            bool    // Detached HEAD state
    Error                 error   // Status extraction error
    // ... other existing fields
}
```

**Verification Step**: During Phase 1, confirm that `internal/models/repository.go` includes all these fields. If any are missing, they must be added to the `GitStatus` struct.

---

## Data Flow

### Version Flag Flow

```
User: gitree --version
  ↓
Cobra: Parse flags, detect --version
  ↓
PreRunE Hook: Execute before main logic
  ↓
Display: Print version, commit, buildTime
  ↓
Exit: Return immediately (no scanning)
```

### No-Color Flag Flow

```
User: gitree --no-color
  ↓
Cobra: Parse flags, set noColor bool
  ↓
Main Init: color.NoColor = true
  ↓
All Output: fatih/color respects global flag
  ↓
Result: No ANSI escape codes in output
```

### Default Filtering Flow (Show Only Changed Repos)

```
User: gitree (no --all flag)
  ↓
Cobra: Parse flags, showAll = false (default)
  ↓
Scan: Collect all repositories (existing logic)
  ↓
Status Extract: Get GitStatus for all repos (existing logic)
  ↓
Filter: cli.FilterRepositories(repos, FilterOptions{ShowAll: false})
  ↓
  For each repo:
    IsClean(repo) → evaluate FR-008 conditions
    If clean: exclude from output
    If needs attention: include in output
  ↓
Build Tree: tree.Build() with filtered list
  ↓
Prune: Remove directory branches with only clean repos
  ↓
Format: tree.Format() on pruned tree
  ↓
Output: Display filtered tree (only repos needing attention)
```

### All Flag Flow (Show All Repos)

```
User: gitree --all
  ↓
Cobra: Parse flags, showAll = true
  ↓
Scan: Collect all repositories (existing logic)
  ↓
Status Extract: Get GitStatus for all repos (existing logic)
  ↓
Filter: cli.FilterRepositories(repos, FilterOptions{ShowAll: true})
  ↓
  Returns unfiltered list (all repos)
  ↓
Build Tree: tree.Build() with complete list
  ↓
Format: tree.Format() on complete tree
  ↓
Output: Display all repositories regardless of state
```

---

## Validation Rules

### FilterOptions

- No validation required (simple boolean flag)
- Default value: `ShowAll = false` (filtering enabled, show only repos needing attention)

### VersionInfo

- Fields are build-time constants
- Validation: Display defaults ("dev", "none", "unknown") clearly indicate non-release builds
- Format: BuildTime should be ISO 8601 (enforced by Makefile date command)

### Clean State Determination (IsClean function)

**Validation Logic** (implements FR-008):

```go
func IsClean(repo *models.Repository) bool {
    // Rule 1: Nil status = unknown = dirty (fail-safe)
    if repo.GitStatus == nil {
        return false
    }

    status := repo.GitStatus

    // Rule 2: Status extraction error = unknown = dirty (fail-safe)
    if status.Error != nil {
        return false
    }

    // Rule 3: Must be on main or master branch
    if status.Branch != "main" && status.Branch != "master" {
        return false
    }

    // Rule 4: No uncommitted changes
    if status.HasUncommittedChanges {
        return false
    }

    // Rule 5: No stashes
    if status.Stashes > 0 {
        return false
    }

    // Rule 6: Must have remote tracking
    if !status.HasRemote {
        return false
    }

    // Rule 7: Not ahead of remote
    if status.Ahead > 0 {
        return false
    }

    // Rule 8: Not behind remote
    if status.Behind > 0 {
        return false
    }

    // Rule 9: Not in detached HEAD state
    if status.IsDetached {
        return false
    }

    // All conditions met: repository is clean
    return true
}
```

**Truth Table** (sample cases):

| Branch | Changes | Stashes | Remote | Ahead | Behind | Detached | Result | Reason |
|--------|---------|---------|--------|-------|--------|----------|--------|--------|
| main | false | 0 | true | 0 | 0 | false | CLEAN | All conditions met |
| feature | false | 0 | true | 0 | 0 | false | DIRTY | Not main/master |
| main | true | 0 | true | 0 | 0 | false | DIRTY | Has changes |
| main | false | 2 | true | 0 | 0 | false | DIRTY | Has stashes |
| main | false | 0 | false | 0 | 0 | false | DIRTY | No remote |
| main | false | 0 | true | 3 | 0 | false | DIRTY | Ahead of remote |
| main | false | 0 | true | 0 | 5 | false | DIRTY | Behind remote |
| HEAD | false | 0 | true | 0 | 0 | true | DIRTY | Detached HEAD |
| main (status.Error != nil) | - | - | - | - | - | - | DIRTY | Status error (fail-safe) |

---

## Impact Analysis

### Modified Packages

1. **cmd/gitree**:
   - New: `root.go` - Cobra command setup
   - Modified: `main.go` - Cobra integration, version variables, flag parsing
   - Add FilterOptions construction

2. **internal/cli** (NEW package):
   - New: `filter.go` with `FilterOptions`, `IsClean()`, `FilterRepositories()`
   - New: `filter_test.go` - unit tests for filtering logic
   - New: `version.go` - version formatting utilities (optional)
   - New: `version_test.go` - version display tests

3. **Makefile**:
   - Add ldflags to build target
   - Inject VERSION, COMMIT, BUILDTIME variables

### No Changes Required

1. **internal/scanner**: No modifications (filtering happens post-scan)
2. **internal/gitstatus**: No modifications (status extraction unchanged)
3. **internal/models**: Verify fields exist, likely no changes needed

---

## Testing Data Requirements

### Unit Tests

1. **FilterOptions**:
   - Test default behavior (ShowAll: false - filtering enabled)
   - Test --all flag behavior (ShowAll: true - show all repos)

2. **IsClean() Function**:
   - Test each FR-008 condition individually
   - Test combination scenarios (truth table cases)
   - Test fail-safe behavior (nil status, error status)
   - Test edge cases (detached HEAD, no remote)

3. **FilterRepositories()**:
   - Test empty repository list
   - Test all clean repositories (default mode shows none, --all shows all)
   - Test all repos needing attention (both modes show all)
   - Test mixed clean/changed repositories (default filters, --all shows all)
   - Test tree structure preservation in both modes

### Integration Tests

1. **Version Flag**:
   - Build with ldflags, verify output
   - Build without ldflags, verify defaults
   - Test version flag precedence (with other flags)

2. **No-Color Flag**:
   - Run with --no-color, scan output for ANSI codes
   - Run without flag, verify colors present
   - Test NO_COLOR environment variable

3. **Filtering Behavior**:
   - Create test repository tree with known clean/changed repos
   - Run without --all flag (default mode)
   - Verify only changed repos shown
   - Run with --all flag
   - Verify all repos shown
   - Verify tree structure maintained in both modes

---

## Summary

**New Entities**: 2

- `FilterOptions` (internal/cli)
- `VersionInfo` (cmd/gitree - package-level variables)

**Modified Entities**: 0 (verify GitStatus has required fields)

**Data Complexity**: Low

- Simple configuration structs
- Stateless filtering logic
- No persistence or state management

**Validation Complexity**: Medium

- IsClean() has 9 conditions to evaluate
- Fail-safe defaults for unknown states
- Truth table testing required

**Next Step**: Generate API contracts (contracts/) defining the filtering function signatures and version display format.

---

## Version Information Model (Extended)

### Structure

Version information is injected at build time via Makefile ldflags:

```go
// Package-level variables in cmd/gitree/main.go
var (
    version   = "dev"       // Default for development builds
    commit    = "none"      // Default when Git metadata unavailable
    buildTime = "unknown"   // Default when build timestamp not injected
)
```

### Build-Time Injection

**Makefile Variables**:

```makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILDTIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS = -X 'main.version=$(VERSION)' \
          -X 'main.commit=$(COMMIT)' \
          -X 'main.buildTime=$(BUILDTIME)'
```

**Build Command**:

```bash
go build -ldflags "$(LDFLAGS)" -o bin/gitree ./cmd/gitree
```

### Default Values (FR-012)

When built without injected metadata (e.g., `go run` or `go build` without ldflags):

| Variable | Default Value | Meaning |
|----------|---------------|---------|
| `version` | `"dev"` | Development build, not from tagged release |
| `commit` | `"none"` | Git commit hash unavailable |
| `buildTime` | `"unknown"` | Build timestamp not recorded |

### Display Format

**Output Template**:

```
gitree version {version}
  commit: {commit}
  built:  {buildTime}
```

**Example (Production Build)**:

```
gitree version v1.2.3
  commit: a1b2c3d
  built:  2025-11-17T14:30:00Z
```

**Example (Development Build)**:

```
gitree version dev
  commit: none
  built:  unknown
```

### Validation (SC-001)

- Version display must complete in <100ms
- All three fields must be present in output
- Format must be consistent (human-readable, not JSON for version command)
- Exit code must be 0 on success

### References

- **FR-001**: Version display requirement
- **FR-002**: Build-time injection requirement
- **FR-012**: Placeholder value requirement
- **T010-T018**: Implementation tasks
