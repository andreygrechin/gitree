# Feature Specification: Debug Logging

**Feature Branch**: `004-debug-logging`
**Created**: 2025-11-21
**Status**: Implemented
**Input**: User description: "I want to add a flag to enable output of debug information. For example, some directories are marked as having non-clear worktree status. It's not clear why, which makes troubleshooting is very difficult. Use one level of debug logs. Stick with using fmt.Fprintln/fmt.Fprint, no need for structural logging."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Diagnose Worktree Status Issues (Priority: P1)

When a user runs gitree and sees repositories marked with uncommitted changes or unclear status, they need to understand why the tool detected those issues. The debug flag enables visibility into the detection logic and shows exactly which files are causing the status.

**Why this priority**: This is the core use case driving the feature request. Users cannot currently troubleshoot why certain repositories are flagged, making it the most critical capability to deliver.

**Independent Test**: Can be fully tested by running gitree with the debug flag on a directory containing repositories with various git states (dirty worktree, staged changes, etc.) and verifying that debug output lists specific files and explains the status determination.

**Acceptance Scenarios**:

1. **Given** a repository with modified files, **When** user runs gitree with debug flag enabled, **Then** debug output lists each modified file path and marks the repository as needing attention
2. **Given** a repository with untracked files, **When** user runs gitree with debug flag enabled, **Then** debug output lists each untracked file path
3. **Given** a repository with staged changes, **When** user runs gitree with debug flag enabled, **Then** debug output lists each staged file path
4. **Given** a repository on a non-main branch, **When** user runs gitree with debug flag enabled, **Then** debug output shows the current branch name and why it triggered the status indicator
5. **Given** a repository ahead/behind remote, **When** user runs gitree with debug flag enabled, **Then** debug output shows commit counts and remote tracking details

---

### User Story 2 - Troubleshoot Scanning Behavior (Priority: P2)

When gitree takes longer than expected or produces unexpected results, users need to see what directories are being scanned and how the tool is processing the filesystem.

**Why this priority**: Secondary to status diagnosis but still valuable for understanding performance issues and scan behavior.

**Independent Test**: Can be fully tested by running gitree with debug flag on a complex directory structure and verifying that debug output shows directory traversal decisions and timing information.

**Acceptance Scenarios**:

1. **Given** a directory tree with nested repositories, **When** user runs gitree with debug flag enabled, **Then** debug output shows which directories are being scanned and which are being skipped
2. **Given** a slow-responding repository, **When** user runs gitree with debug flag enabled, **Then** debug output shows status extraction timing for each repository
3. **Given** symlinks in the directory structure, **When** user runs gitree with debug flag enabled, **Then** debug output shows symlink detection and loop prevention logic

---

### User Story 3 - Debug Flag Control (Priority: P1)

Users need a simple command-line flag to enable debug output without modifying configuration files or environment variables.

**Why this priority**: This is the interface to the feature and must work for the other stories to be usable.

**Independent Test**: Can be fully tested by verifying that the flag appears in help output, can be enabled/disabled, and controls debug output visibility.

**Acceptance Scenarios**:

1. **Given** no debug flag specified, **When** user runs gitree, **Then** no debug output appears and only standard tree output is shown
2. **Given** debug flag enabled, **When** user runs gitree, **Then** debug output appears alongside standard tree output
3. **Given** user runs gitree with --help, **When** viewing help text, **Then** debug flag is documented with clear description of its purpose

---

### Edge Cases

- What happens when debug output is written to a non-terminal (file, pipe)? Should maintain the same format.
- How does debug output interact with the progress spinner? Spinner is disabled entirely when debug mode is active.
- What happens when debug flag is combined with --no-color? Debug output should also respect color settings.
- How verbose should debug output be for large repository collections? Should not overwhelm the user with excessive detail.
- What happens when a repository has hundreds of modified files? Debug output shows first 20 files, then adds summary line for remaining files.

## Clarifications

### Session 2025-11-21

- Q: When a repository has many modified files (>20), how should debug output handle the file list? → A: Show first 20 files, then add "...and N more files" summary line
- Q: How should debug output interact with the progress spinner? → A: Disable spinner entirely when debug mode is active

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a command-line flag to enable debug output
- **FR-002**: Debug output MUST explain why each repository was marked as needing attention (uncommitted changes, branch status, ahead/behind counts, stashes)
- **FR-003**: Debug output MUST list specific files that cause non-clean worktree status, including:
  - Modified files (changes in working directory)
  - Untracked files (new files not in git)
  - Staged changes (files in index ready to commit)
  - Deleted files
  - When file count exceeds 20 per category, show first 20 files followed by "...and N more files" summary line
- **FR-004**: Debug output MUST show directory scanning decisions (which directories are entered, which are skipped, and why)
- **FR-005**: Debug output MUST be written to stderr to separate it from standard output
- **FR-006**: Debug output MUST use simple print statements (fmt.Fprintln/fmt.Fprint) without structured logging frameworks
- **FR-007**: Debug output MUST respect the --no-color flag and NO_COLOR environment variable
- **FR-008**: Debug output MUST include timing information for git status extraction operations
- **FR-009**: Debug flag MUST be documented in --help output
- **FR-010**: Progress spinner MUST be disabled entirely when debug mode is active
- **FR-011**: When debug is disabled (default), no debug output MUST be produced

### Assumptions

- Debug output will use a simple prefix like "DEBUG:" to distinguish it from regular output
- Debug level will be boolean (on/off) rather than multiple verbosity levels
- Debug output will be human-readable text, not machine-parseable structured logs
- Timing information will be shown in milliseconds for operations taking over 100ms
- File listing applies per category: if a repository has 25 modified and 5 untracked files, debug shows all 5 untracked but only first 20 modified plus "...and 5 more modified files"

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can determine why a specific repository was marked as needing attention by reading debug output
- **SC-002**: Users can identify the exact files causing non-clean worktree status for any repository
- **SC-003**: Debug output includes timing information for all git status operations taking longer than 100 milliseconds
- **SC-004**: Debug flag can be enabled/disabled via command-line argument with immediate effect
- **SC-005**: Debug output does not obscure or interfere with standard tree output readability
- **SC-006**: 100% of git status determination logic includes corresponding debug output explaining the decision

## Scope *(mandatory)*

### In Scope

- Adding command-line flag for debug mode
- Debug output for git status determination (why repos are flagged)
- Debug output listing specific files causing non-clean status
- Debug output for directory scanning behavior
- Timing information for slow operations
- Integration with existing --no-color flag
- Documentation in help text

### Out of Scope

- Multiple debug verbosity levels (only one level: on/off)
- Structured logging frameworks (using simple fmt.Fprint)
- Debug output configuration via config file
- Debug output to file (only stderr)
- Performance profiling or trace output
- Debug output for tree formatting logic (focus on scanning and status)
- Filtering debug output by repository or path

## Dependencies & Constraints *(mandatory)*

### Dependencies

- Existing CLI flag parsing mechanism
- Current stderr/stdout separation
- Existing color handling via --no-color flag
- Progress spinner implementation
- Git status extraction in internal/gitstatus package

### Constraints

- Must use fmt.Fprintln/fmt.Fprint only (no logging frameworks)
- Debug output must go to stderr
- Must not degrade performance when debug is disabled
- Must maintain backward compatibility (default behavior unchanged)

## Open Questions

None. All aspects of the feature are sufficiently specified for implementation.
