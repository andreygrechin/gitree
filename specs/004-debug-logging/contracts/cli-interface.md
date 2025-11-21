# CLI Interface Contract: Debug Logging

**Feature**: `004-debug-logging`
**Date**: 2025-11-21
**Contract Type**: Command-Line Interface

## Overview

This contract defines the CLI interface for the debug logging feature. It specifies the command-line flag, its behavior, interactions with other flags, and output contracts.

## Command Syntax

```bash
gitree [--debug] [--all] [--no-color] [--version] [--help]
```

## Flag Specification

### --debug Flag

**Name**: `--debug`
**Type**: Boolean flag
**Default**: `false` (debug disabled)
**Short form**: None
**Required**: No
**Position**: Any (flag order doesn't matter)

**Description**: Enable debug output showing diagnostic information about directory scanning and git status determination.

**Behavior**:
- When enabled: Outputs debug messages to stderr with "DEBUG: " prefix
- When disabled (default): No debug output produced
- Can be combined with other flags (--all, --no-color)
- Takes precedence over spinner (spinner disabled when debug enabled)

## Flag Interactions

### --debug with --all

```bash
gitree --debug --all
```

**Behavior**:
- Debug output includes diagnostic info for **all** repositories (not just those needing attention)
- Both flags are independent and combinable
- Debug output volume may be higher since more repositories are processed

**Example output**:
```
DEBUG: Scanning directory: /home/user/projects
DEBUG: Found git repository: /home/user/projects/repo1 (regular)
DEBUG: Repository /home/user/projects/repo1: branch=main, hasChanges=false
DEBUG: Found git repository: /home/user/projects/repo2 (regular)
DEBUG: Repository /home/user/projects/repo2: branch=main, hasChanges=false
/home/user/projects
├── repo1 [[ main ]]
└── repo2 [[ main ]]
```

### --debug with --no-color

```bash
gitree --debug --no-color
```

**Behavior**:
- Debug output respects --no-color flag (no ANSI color codes in debug messages)
- Tree output also has colors disabled (existing behavior)
- Particularly useful for CI/CD or when piping to files

**Example output** (no color codes):
```
DEBUG: Scanning directory: /home/user/projects
DEBUG: Found git repository: /home/user/projects/repo1 (regular)
/home/user/projects
└── repo1 [[ main | * ]]
```

### --debug with --version

```bash
gitree --debug --version
```

**Behavior**:
- Version flag takes precedence (exits after showing version)
- No debug output produced
- Existing behavior maintained (version shows, program exits)

**Output**:
```
gitree version 0.1.0
  commit: abc123
  built:  2025-11-21T10:00:00Z
```

### --debug with NO_COLOR environment variable

```bash
NO_COLOR=1 gitree --debug
```

**Behavior**:
- Equivalent to `gitree --debug --no-color`
- Environment variable respected (existing behavior)

## Output Contracts

### Stdout (Standard Output)

**Purpose**: Program data output (tree structure)

**Content**:
- Git repository tree structure (existing format)
- Repository status information (existing format)
- Informational messages (e.g., "No Git repositories found")

**Guarantee**: **Unchanged by debug flag**
- Stdout output is identical whether --debug is enabled or disabled
- Allows piping stdout to other tools without debug noise

**Example**:
```
/home/user/projects
├── repo1 [[ main | ↑2 * ]]
├── repo2 [[ develop | $ ]]
└── repo3 [[ main ]]
```

### Stderr (Standard Error)

**Purpose**: Diagnostic output and progress indication

**Content when debug disabled** (default):
- Progress spinner (rotating animation with status text)
- Error messages (e.g., "Warning: Some repositories failed...")

**Content when debug enabled**:
- Debug messages with "DEBUG: " prefix
- Error messages (same as default)
- **No spinner** (spinner disabled when debug active per FR-010)

**Format**: Each debug line starts with "DEBUG: " followed by diagnostic message

**Example**:
```
DEBUG: Scanning directory: /home/user/projects
DEBUG: Entering directory: /home/user/projects/repo1
DEBUG: Found git repository: /home/user/projects/repo1 (regular)
DEBUG: Repository /home/user/projects/repo1 status extraction: 123ms
DEBUG: Repository /home/user/projects/repo1: branch=main, hasChanges=true
DEBUG: Modified files (2): src/main.go, README.md
DEBUG: Entering directory: /home/user/projects/repo2
DEBUG: Found git repository: /home/user/projects/repo2 (regular)
DEBUG: Repository /home/user/projects/repo2: branch=develop, hasChanges=false
```

## Debug Output Specification

### Message Format

**Structure**: `DEBUG: <message>`

**Components**:
- Prefix: `DEBUG: ` (uppercase, colon, space) - exactly 7 characters
- Message: Diagnostic information in human-readable format
- Newline: Always ends with `\n`

**No structured format**: Plain text, not JSON or key=value pairs (per spec FR-006)

### Message Categories

#### 1. Directory Scanning Messages

**Format**: `DEBUG: <action> <path> [reason]`

**Examples**:
```
DEBUG: Scanning directory: /home/user/projects
DEBUG: Entering directory: /home/user/projects/subdir
DEBUG: Skipping /home/user/projects/private: permission denied
DEBUG: Skipping /home/user/projects/repo/.git: inside git repository
DEBUG: Skipping /home/user/projects/link: already visited (symlink loop)
```

**Trigger**: Each directory processing decision in scanner.walkFunc()

#### 2. Repository Detection Messages

**Format**: `DEBUG: Found git repository: <path> (regular|bare)`

**Examples**:
```
DEBUG: Found git repository: /home/user/projects/myapp (regular)
DEBUG: Found git repository: /home/user/projects/repo.git (bare)
```

**Trigger**: When IsGitRepository() returns true

#### 3. Status Extraction Messages

**Format**: `DEBUG: Repository <path>: <status_details>`

**Examples**:
```
DEBUG: Repository /home/user/projects/repo: branch=main, hasChanges=true
DEBUG: Repository /home/user/projects/repo: branch=develop, hasRemote=false
DEBUG: Repository /home/user/projects/repo: branch=DETACHED, hasChanges=false
```

**Status details include**:
- `branch=<name>` - Current branch or DETACHED
- `hasChanges=<bool>` - Uncommitted changes present
- `hasRemote=<bool>` - Remote configured
- `ahead=<N>` - Commits ahead of remote (if hasRemote=true and N>0)
- `behind=<N>` - Commits behind remote (if hasRemote=true and N>0)
- `hasStashes=<bool>` - Stashes present (if true)

**Trigger**: After extractGitStatus() completes

#### 4. Timing Messages

**Format**: `DEBUG: Repository <path> status extraction: <duration>ms`

**Examples**:
```
DEBUG: Repository /home/user/projects/slow-repo status extraction: 123ms
DEBUG: Repository /home/user/projects/huge-repo status extraction: 1523ms
```

**Condition**: Only shown when duration > 100ms (per spec SC-003)

**Trigger**: After status extraction completes, if elapsed time exceeds threshold

#### 5. File Listing Messages

**Format**:
```
DEBUG: <category> files (<count>): <file1>, <file2>, ...
DEBUG: ...and <N> more <category> files
```

**Categories**:
- `Modified` - Changes in working directory not staged
- `Untracked` - New files not tracked by git
- `Staged` - Changes added to index
- `Deleted` - Files deleted from working directory

**Examples**:
```
DEBUG: Modified files (3): src/main.go, src/utils.go, README.md
DEBUG: Untracked files (1): newfile.txt
```

**With truncation** (>20 files per category):
```
DEBUG: Modified files (25): file1.go, file2.go, ..., file20.go
DEBUG: ...and 5 more modified files
DEBUG: Untracked files (3): new1.txt, new2.txt, new3.txt
```

**Condition**: Only shown when hasChanges=true

**Trigger**: After extractUncommittedChanges() detects changes

## Help Text Contract

### --help Output

**Location**: `gitree --help` or `gitree -h`

**Required content**: Debug flag must be documented

**Format** (excerpt):
```
Flags:
  -a, --all         Show all repositories including clean ones (default shows only repos needing attention)
      --debug       Enable debug output showing diagnostic information
  -h, --help        help for gitree
      --no-color    Disable color output
  -v, --version     Display version information
```

**Ordering**: Alphabetical by long flag name (existing convention)

**Description**: Must clearly indicate purpose of debug flag

## Exit Codes

**Unchanged**: Debug flag does not affect exit codes

- `0` - Success (repositories scanned successfully)
- `1` - Error (unable to scan directory, fatal error)

**Note**: Debug output failures (stderr write errors) do not cause non-zero exit code (debug is best-effort)

## Environment Variables

### NO_COLOR

**Existing behavior maintained**:
- When `NO_COLOR` is set (any value), color output is disabled
- Affects both stdout tree output and stderr debug output
- Equivalent to --no-color flag

**Contract**:
```bash
NO_COLOR=1 gitree --debug
# Debug output has no ANSI color codes
# Tree output has no ANSI color codes
```

## Performance Guarantees

### Debug Disabled (Default)

**Overhead**: <1% additional execution time
**Memory**: <10 bytes additional memory (boolean flag storage)
**I/O**: No additional I/O

**Contract**: User should not perceive any performance difference when debug is disabled

### Debug Enabled

**Overhead per repository**: <50ms additional time (per spec performance goal)
**Memory**: <2KB per repository (file path strings, max 20 per category)
**I/O**: Stderr writes buffered by OS (non-blocking)

**Contract**: Debug output may slow execution slightly, but remains usable for large repository collections

## Backward Compatibility

**Guarantee**: Fully backward compatible

**Existing behavior preserved**:
- Default behavior unchanged (debug disabled by default)
- All existing flags work identically
- Stdout output format unchanged
- Stderr output (spinner, errors) unchanged when debug disabled

**No breaking changes**:
- No removed flags
- No changed flag meanings
- No changed output formats (when debug disabled)

**Script compatibility**: Existing scripts using gitree continue to work without modification

## Examples

### Basic Usage

**Command**:
```bash
gitree --debug
```

**Stderr output** (debug):
```
DEBUG: Scanning directory: /home/user/projects
DEBUG: Entering directory: /home/user/projects/repo1
DEBUG: Found git repository: /home/user/projects/repo1 (regular)
DEBUG: Repository /home/user/projects/repo1: branch=main, hasChanges=true
DEBUG: Modified files (2): src/main.go, README.md
```

**Stdout output** (unchanged):
```
/home/user/projects
└── repo1 [[ main | * ]]
```

### Debug with All Repositories

**Command**:
```bash
gitree --debug --all
```

**Behavior**: Shows debug info for all repos including clean ones

### Debug to File

**Command**:
```bash
gitree --debug 2>debug.log
```

**Result**:
- Stdout (tree) displayed on terminal
- Stderr (debug) redirected to debug.log file

**Use case**: Capture debug info for later analysis without cluttering terminal

### Debug with Tree to File

**Command**:
```bash
gitree --debug >repos.txt 2>debug.log
```

**Result**:
- Stdout (tree) → repos.txt
- Stderr (debug) → debug.log
- Separate files for data vs diagnostics

## Validation

### Valid Combinations

✅ `gitree --debug`
✅ `gitree --debug --all`
✅ `gitree --debug --no-color`
✅ `gitree --debug --all --no-color`
✅ `NO_COLOR=1 gitree --debug`

### Invalid Combinations

None - debug flag can be combined with any other flag safely

### Edge Cases

**Case**: Debug flag specified multiple times
```bash
gitree --debug --debug
```
**Behavior**: Accepted (Cobra's standard behavior, flag set to true once)

**Case**: No repositories found
```bash
gitree --debug
```
**Stderr**:
```
DEBUG: Scanning directory: /home/user/empty
DEBUG: No repositories found
```
**Stdout**:
```
No Git repositories found in this directory.
```

**Case**: Permission denied on directory
```bash
gitree --debug
```
**Stderr**:
```
DEBUG: Scanning directory: /home/user/projects
DEBUG: Skipping /home/user/projects/private: permission denied
```
**Behavior**: Continue scanning, collect error, show in ScanResult.Errors

## Testing Contract

### Test Requirements

1. **Flag parsing**: Verify --debug flag is recognized and parsed correctly
2. **Help text**: Verify --debug appears in --help output
3. **Debug output presence**: Verify debug messages appear when flag enabled
4. **Debug output absence**: Verify no debug messages when flag disabled
5. **Output separation**: Verify stdout unchanged by debug flag
6. **Color handling**: Verify --no-color affects debug output
7. **Flag combination**: Verify --debug works with --all and --no-color
8. **Spinner suppression**: Verify spinner not shown when debug enabled

### Test Fixtures

- Directory with multiple repositories (various states)
- Repository with >20 modified files (test truncation)
- Repository causing slow status extraction (test timing output)
- Restricted permissions directory (test permission denied debug)
- Symlink directory structure (test loop detection debug)

## Security Considerations

**Path disclosure**: Debug output includes full file paths
- **Risk**: Minimal - user already has filesystem access
- **Mitigation**: Debug is opt-in (user explicitly enables)

**Sensitive file names**: File listings may reveal .env or credential file names
- **Risk**: Minimal - no file content exposed, only names
- **Mitigation**: User controls when debug is enabled

**Stderr capture**: Debug output may be logged by CI/CD systems
- **Risk**: Minimal - no secrets or credentials in debug output
- **Mitigation**: Gitree doesn't read or output file contents

**Conclusion**: Debug feature has no significant security implications

## API Stability

**Stability guarantee**: Debug output format is **NOT** part of stable API

**Rationale**:
- Debug output is for human consumption, not machine parsing (per spec FR-006)
- Message text, format, or verbosity may change in future versions
- Scripts should rely on stdout (tree output), not debug messages

**Recommendation**: Do not parse debug output in scripts or automation
