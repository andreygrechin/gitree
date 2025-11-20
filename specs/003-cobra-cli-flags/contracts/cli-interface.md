# CLI Interface Contract

**Feature**: 003-cobra-cli-flags
**Package**: cmd/gitree
**Date**: 2025-11-17

## Command Line Interface

### Root Command

```
gitree [flags] [path]
```

**Description**: Recursively scan directories for Git repositories and display them in a tree structure with status information.

**Arguments**:

- `path` (optional): Directory to scan. Defaults to current working directory.

**Exit Codes**:

- `0`: Success
- `1`: Error (scanning failed, invalid flags, etc.)

---

## Flags

### --version

**Type**: Boolean flag
**Default**: `false`
**Short**: `-v`

**Behavior**:

- Displays version information and exits immediately
- Does not perform directory scanning
- Precedence: Executes before all other flags

**Output Format** (stdout):

```
gitree version <semver>
  commit: <git-hash>
  built:  <iso8601-timestamp>
```

**Example**:

```bash
$ gitree --version
gitree version v1.2.3
  commit: a1b2c3d
  built:  2025-11-17T14:32:00Z
```

**Default Values** (development builds):

```
gitree version dev
  commit: none
  built:  unknown
```

**Contract**:

- FR-001: MUST display all three fields (version, commit, build time)
- FR-003: MUST exit immediately without scanning
- FR-012: MUST show placeholders when build metadata not injected
- SC-001: MUST complete in <100ms
- SC-005: MUST not incur scanning overhead

**Test Cases**:

1. Run with --version, verify output format
2. Run with --version and other flags, verify version takes precedence
3. Build without ldflags, verify default values displayed
4. Measure execution time, verify <100ms

---

### --no-color

**Type**: Boolean flag
**Default**: `false`

**Behavior**:

- Suppresses all ANSI color codes in output
- Applies globally to all output (tree, status, errors)
- Respects `NO_COLOR` environment variable per [no-color.org](https://no-color.org/) standard

**Output Impact**:

- Repository paths: plain text (no color)
- Branch names: plain text (no color)
- Ahead/behind indicators: plain text (no color)
- Status symbols: plain text (no color)
- Error messages: plain text (no color)
- Tree characters: unchanged (box-drawing, not color)

**Example**:

```bash
# With color (default)
$ gitree
[38;5;2m/path/to/repo[0m [38;5;4m[main ↑2][0m

# Without color
$ gitree --no-color
/path/to/repo [main ↑2]
```

**Environment Variable**:

```bash
# Either flag or env var disables color
$ NO_COLOR=1 gitree
# Same as: gitree --no-color
```

**Contract**:

- FR-004: MUST suppress all ANSI escape codes
- FR-005: MUST apply globally to all output types
- SC-002: Output MUST contain zero ANSI escape sequences
- SC-007: Output MUST be parseable in CI/CD pipelines

**Test Cases**:

1. Run with --no-color, scan output for `\x1b[` sequences (count must be 0)
2. Run with NO_COLOR=1, verify same behavior
3. Combine with default filtering (or --all), verify all output remains plain
4. Verify text formatting (symbols, indentation) preserved

**ANSI Detection Pattern**:

```regex
\x1b\[[0-9;]*m
```

---

### --all

**Type**: Boolean flag
**Default**: `false`

**Behavior**:

- Disables default filtering to show ALL repositories (including clean ones)
- By default (without this flag), only repositories needing attention are shown
- Maintains tree structure hierarchy in both modes

**Default Behavior (without --all)**:
Shows only repositories needing attention, filtering out clean repos

**Clean State Definition** (FR-008):
A repository is "clean" (excluded by default, shown with --all) if ALL of the following are true:

1. Branch is `main` or `master`
2. No uncommitted changes (staged or unstaged)
3. No stashed changes
4. Synchronized with remote (not ahead or behind)
5. Has remote tracking configured

**Needs Attention Criteria** (FR-009):
A repository needs attention (shown by default, always shown with --all) if ANY of the following are true:

1. On branch other than main/master
2. Has uncommitted changes
3. Has stashed changes
4. Ahead of remote
5. Behind remote
6. No remote tracking configured
7. Detached HEAD state
8. Status extraction failed/timed out

**Output Examples**:

**Input Tree** (with --all):

```text
/workspace
├── project-a/    [main] clean
├── project-b/    [feature/new] needs attention
└── nested/
    ├── project-c/ [main ↑2] needs attention
    └── project-d/ [master] clean
```

**Default Output** (without --all):

```text
/workspace
├── project-b/    [feature/new]
└── nested/
    └── project-c/ [main ↑2]
```

**Output with --all**:

```text
/workspace
├── project-a/    [main]
├── project-b/    [feature/new]
└── nested/
    ├── project-c/ [main ↑2]
    └── project-d/ [master]
```

**All Repos Clean** (default mode):

```bash
$ gitree
No repositories need attention. All repositories are clean.

$ gitree --all
/workspace
├── project-a/    [main]
└── project-d/    [master]
```

**Contract**:

- FR-006: By default (without --all), MUST show only repos needing attention
- FR-007: --all flag MUST disable filtering and show all repositories
- FR-008: MUST define clean state precisely (5 conditions, all must be true)
- FR-009: MUST include repos needing attention (fail-safe for errors)
- FR-010: MUST preserve tree hierarchy in default mode
- FR-011: MUST prune empty directory branches in default mode
- SC-003: MUST achieve 100% filtering accuracy
- SC-006: Default mode MUST reduce clutter by ≥50% in typical environments
- SC-009: --all flag MUST show all repositories including clean ones

**Test Cases**:

1. Test default mode (no --all): verify only repos needing attention shown
2. Test --all flag: verify all repos shown
3. Test each clean condition individually (5 conditions)
4. Test each "needs attention" condition individually (8 conditions)
5. Test fail-safe: repo with status error → shown in default mode
6. Test all clean with default mode → special message
7. Test all clean with --all → shows all repos
8. Test tree pruning in default mode: verify empty branches removed
9. Measure filtering accuracy: 100% on test suite

**Edge Cases**:

1. Main branch WITH changes → shown in default mode (any attention condition)
2. Detached HEAD → shown in default mode
3. No remote configured → shown in default mode
4. Status timeout → shown in default mode (fail-safe)
5. All repos clean + --all → shows all clean repos
6. Mixed clean/changed + --all → shows all repos

---

## Flag Combinations

### Valid Combinations

| Flags | Behavior |
|-------|----------|
| (none) | Default: scan current directory, show only repos needing attention, with color |
| --version | Show version, exit (ignore other flags) |
| --no-color | Show only repos needing attention (default filter), no color |
| --all | Show all repos including clean ones, with color |
| --no-color --all | Show all repos, no color |
| --version --all | Show version, exit (--version takes precedence) |

### Precedence Rules

1. **--version**: Always executes first, exits immediately
2. **--no-color**: Applied during initialization (before scanning)
3. **Filtering (default or --all)**: Applied after scanning and status extraction

**Contract**:

- FR-012: MUST allow flag combinations (except --version)
- FR-015: MUST handle invalid combinations gracefully
- SC-004: Combined flags must produce correct output

**Test Cases**:

1. Test all valid combinations in table above
2. Verify --version precedence with each flag
3. Verify --no-color (default filter) produces plain output with only changed repos
4. Verify --no-color --all produces plain output with all repos
5. Test with path argument + flags

---

## Input/Output Protocol

### Standard Streams

**stdin**: Not used (CLI tool)
**stdout**: Tree output, version information
**stderr**: Errors, warnings, spinner messages

**Example Flow**:

```bash
$ gitree 2>errors.log 1>output.txt
# stderr → errors.log (spinner, errors)
# stdout → output.txt (filtered tree output - only repos needing attention)

$ gitree --all 2>errors.log 1>output.txt
# stderr → errors.log (spinner, errors)
# stdout → output.txt (complete tree output - all repos)
```

### Output Format

**Tree Structure** (stdout):

```
[path]
├── [repo-name]/  [git-status]
├── [directory]/
│   └── [repo-name]/  [git-status]
└── [repo-name]/  [git-status]
```

**Git Status Format**:

```
[branch-name ↑ahead ↓behind $ *]

Symbols:
  ↑N  - N commits ahead of remote
  ↓N  - N commits behind remote
  $   - Has stashed changes
  *   - Has uncommitted changes
  ○   - No remote configured
```

**Error Format** (stderr):

```
Error: <error-message>
Warning: <warning-message>
```

### Contract

- CLI Interface principle: Text in/out protocol
- FR-014: Help text must explain all flags
- SC-008: 90% users understand flags without external docs

---

## Help Text

**Command**:

```bash
gitree --help
```

**Output** (generated by Cobra):

```
gitree - Recursively scan directories for Git repositories

Usage:
  gitree [flags] [path]

Flags:
  -v, --version       Display version information (semver, commit, build time)
      --no-color      Disable colored output (useful for logs, automation)
  -a, --all           Show all repositories (by default, only repos needing
                      attention are shown: uncommitted changes, ahead/behind
                      remote, non-main branches, stashes, no remote tracking)
  -h, --help          Display this help message

Environment Variables:
  NO_COLOR            Disable colored output (alternative to --no-color flag)

Examples:
  gitree                      # Scan current directory, show only repos needing attention
  gitree /path/to/workspace   # Scan specific directory
  gitree --version            # Show version information
  gitree --no-color           # Show filtered output without colors
  gitree --all                # Show all repositories including clean ones
  gitree --no-color --all     # Show all repos without colors

Exit Codes:
  0   Success
  1   Error (invalid flags, scan failure, etc.)
```

**Contract**:

- FR-014: MUST provide clear help text
- FR-016: MUST use lowercase, hyphen-separated flag names
- SC-008: Help text must be clear enough for 90% comprehension

---

## Build-Time Variables

### Makefile Integration

**Variables** (existing in Makefile):

```makefile
VERSION    := $(shell git describe --tags --always --dirty)
COMMIT     := $(shell git rev-parse --short HEAD)
BUILDTIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
```

**Build Command** (modified):

```makefile
build:
    CGO_ENABLED=0 \
    go build \
        -ldflags "-s -w \
            -X 'main.version=$(VERSION)' \
            -X 'main.commit=$(COMMIT)' \
            -X 'main.buildTime=$(BUILDTIME)'" \
        -o bin/$(APP_NAME) \
        cmd/gitree/main.go
```

**Go Variables** (cmd/gitree/main.go):

```go
var (
    version   = "dev"      // overridden by -ldflags
    commit    = "none"     // overridden by -ldflags
    buildTime = "unknown"  // overridden by -ldflags
)
```

**Contract**:

- FR-002: Build system MUST inject version metadata
- FR-013: Defaults MUST indicate development builds
- Build time format: ISO 8601 (RFC3339)

---

## Summary

This contract defines the external CLI interface for the three new flags. Key contracts:

1. **Version Flag**: <100ms, shows all metadata, exits immediately
2. **No-Color Flag**: Zero ANSI codes, global application, respects NO_COLOR env var
3. **Default Filtering**: Shows only repos needing attention by default, 100% accuracy, tree preservation, fail-safe defaults
4. **All Flag**: Disables filtering to show all repositories including clean ones

All flags follow POSIX conventions (lowercase, hyphenated), generate comprehensive help text, and maintain the existing text-based I/O protocol.
