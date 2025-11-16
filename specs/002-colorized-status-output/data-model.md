# Data Model: Colorized Status Output

**Feature**: 002-colorized-status-output
**Date**: 2025-11-16
**Status**: Complete

## Overview

This feature does not introduce new data entities. It enhances the existing `GitStatus` model's `Format()` method to include ANSI color codes. The data model changes are limited to the formatting logic, not the underlying data structures.

## Existing Entities

### GitStatus

**Location**: `internal/models/repository.go`

**Current Structure**:
```go
type GitStatus struct {
    Branch     string // Current branch name or "DETACHED" if HEAD is detached
    IsDetached bool   // Whether HEAD is in detached state
    HasRemote  bool   // Whether repository has a remote configured
    Ahead      int    // Number of commits ahead of remote
    Behind     int    // Number of commits behind remote
    HasStashes bool   // Whether repository has stashed changes
    HasChanges bool   // Whether repository has uncommitted changes
    Error      string // Partial error message if some status info couldn't be retrieved
}
```

**No Structural Changes Required**: The GitStatus entity remains unchanged. All modifications are in the `Format()` method behavior.

### Repository

**Location**: `internal/models/repository.go`

**Current Structure**:
```go
type Repository struct {
    Path       string     // Absolute file system path to the repository directory
    Name       string     // Base name of the repository directory
    IsBare     bool       // Whether the repository is a bare repository
    IsSymlink  bool       // Whether the repository was reached via a symbolic link
    GitStatus  *GitStatus // Current Git status information (nil if error occurred)
    Error      error      // Error encountered during processing
    HasTimeout bool       // Whether Git operations timed out
}
```

**No Structural Changes Required**: The Repository entity is unaffected.

## Format Method Modifications

### Current Format() Output

**Method**: `GitStatus.Format() string`

**Current Behavior**:
- Returns formatted status string like `[main ↑2 ↓1 $ *]`
- Single brackets `[` and `]`
- No color codes (plain text)
- Conditional parts based on status flags

**Example Outputs**:
- `[main]` - On main, in sync, no changes
- `[main ↑2 ↓1]` - 2 ahead, 1 behind
- `[develop $ *]` - Has stashes and uncommitted changes
- `[DETACHED]` - Detached HEAD
- `[main ○]` - No remote configured

### Enhanced Format() Output

**New Behavior**:
- Returns formatted status string like `[[ main | ↑2 ↓1 $ * ]]` with ANSI color codes
- Double brackets `[[` and `]]` in gray
- Separator `|` in gray
- Color-coded components:
  - Branch: gray (main/master) or yellow (others)
  - Ahead: green
  - Behind: red
  - Stashes: red
  - Uncommitted changes: red
  - No remote indicator: gray
  - "bare" indicator: red

**Example Outputs** (with ANSI codes):
- `\033[90m[[\033[0m \033[90mmain\033[0m \033[90m]]\033[0m`
- `\033[90m[[\033[0m \033[33mfeature/login\033[0m \033[90m|\033[0m \033[32m↑2\033[0m \033[31m↓1\033[0m \033[90m]]\033[0m`
- `\033[90m[[\033[0m \033[90mmain\033[0m \033[90m|\033[0m \033[31m$\033[0m \033[31m*\033[0m \033[90m]]\033[0m`

**When Colors Disabled** (TTY detection or NO_COLOR):
- `[[ main ]]`
- `[[ feature/login | ↑2 ↓1 ]]`
- Double brackets preserved, color codes omitted

## Color Function Definitions

**Location**: `internal/models/repository.go` (package-level)

**New Package-Level Variables**:
```go
var (
    grayColor   = color.New(color.FgHiBlack).SprintFunc()
    yellowColor = color.New(color.FgYellow).SprintFunc()
    greenColor  = color.New(color.FgGreen).SprintFunc()
    redColor    = color.New(color.FgRed).SprintFunc()
)
```

**Rationale**:
- Package-level initialization (created once, reused across all Format() calls)
- SprintFunc() returns `func(...interface{}) string` for clean API
- Color functions automatically respect terminal detection and NO_COLOR

## Validation Rules

**No Changes to Existing Validation**:
- `GitStatus.Validate()` remains unchanged
- Color codes do not affect validation logic
- Validation occurs before formatting

**Existing Validation** (preserved):
- Branch cannot be empty
- Detached HEAD must have Branch = "DETACHED"
- No remote implies Ahead/Behind = 0
- Ahead/Behind cannot be negative

## State Transitions

**No State Changes**: This is a presentation-layer change only. GitStatus state and lifecycle remain unchanged.

## Relationships

**No Relationship Changes**:
- Repository → GitStatus relationship (1:0..1) unchanged
- GitStatus is still nullable in Repository
- Error handling flow unchanged

## Implementation Impact

### Modified Files
- `internal/models/repository.go`: GitStatus.Format() method

### Unmodified Files
- `internal/models/repository.go`: All struct definitions, validation methods
- `internal/gitstatus/extractor.go`: Status extraction logic
- `internal/scanner/scanner.go`: Repository discovery logic
- `internal/tree/formatter.go`: Tree building logic
- `cmd/gitree/main.go`: CLI entry point

### New Dependencies
- `github.com/fatih/color` (upgrade from v1.7.0 to v1.18.0)

## Backward Compatibility

**Breaking Changes**:
- Output format changes from `[branch]` to `[[ branch ]]` (double brackets with spaces)
- ANSI color codes added to output (when terminal supports color)

**Non-Breaking**:
- Data structures unchanged
- API signatures unchanged
- Validation logic unchanged
- All functional requirements from spec 001 preserved

**Mitigation**:
- Users relying on exact output format parsing will need to update parsers
- Users piping output to tools that don't support ANSI codes will see color codes stripped automatically
- Double bracket format provides better visual separation (improvement)

## Summary

The data model for this feature is minimal - no new entities, no structural changes to existing entities. All modifications are isolated to the `GitStatus.Format()` method's output formatting logic. The feature leverages package-level color functions to colorize different status components based on semantic meaning (branch type, sync status, attention indicators). Terminal capability detection is handled automatically by the fatih/color library.
