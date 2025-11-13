# CLI Contract: gitree

**Version**: 1.0.0
**Date**: 2025-11-13

## Overview

The `gitree` command-line tool provides a tree-structured view of Git repositories in a directory hierarchy with inline Git status information.

## Command Signature

```bash
gitree
```

**Note**: Version 1.0.0 does not accept any command-line arguments or flags.

## Input

### Standard Input (stdin)

- Not used in version 1.0.0

### Command-Line Arguments

- None in version 1.0.0

### Environment Variables

- `PWD`: Current working directory (standard shell variable, determines scan starting point)
- `HOME`: User home directory (used by Git for configuration)
- `GIT_CONFIG_*`: Standard Git environment variables (if set, affects Git operations)

### File System

- Requires read access to directories being scanned
- Requires read access to `.git` directories for repository information
- Current working directory is the starting point for the scan

## Output

### Standard Output (stdout)

#### Success Case - Repositories Found

Format: ASCII tree structure with inline Git status

```text
.
├── project1 [main ↑2 ↓1 *]
├── project2 [develop $ *]
├── subdir/
│   ├── project3 [DETACHED]
│   └── project4 [feature ○]
└── project5 [main bare]
```

**Output Structure**:

- First line: `.` (current directory indicator)
- Subsequent lines: Tree structure with repositories
- Each line format: `{tree-prefix}{repo-name} {git-status}`
- Tree prefixes: `├──`, `└──`, `│` (standard tree command characters)
- Git status format: `[{branch} {indicators}]`

**Git Status Indicators**:

- Branch name: e.g., `main`, `develop`, `feature/foo`
- `DETACHED`: Shown instead of branch name when HEAD is detached
- `↑N`: N commits ahead of remote
- `↓N`: N commits behind remote
- `○`: No remote configured (replaces ahead/behind)
- `$`: Has stashed changes
- `*`: Has uncommitted changes
- `bare`: Bare repository (no working directory)
- `error`: Error retrieving full Git information
- `timeout`: Git operations exceeded timeout

**Status Indicator Combinations** (examples):

- `[main]`: Clean branch, in sync
- `[main ↑2]`: 2 commits ahead
- `[main ↓1]`: 1 commit behind
- `[main ↑2 ↓1]`: Diverged from remote
- `[main $ *]`: Stashes and uncommitted changes
- `[main ○]`: No remote configured
- `[DETACHED]`: Detached HEAD
- `[main bare]`: Bare repository
- `[main] error`: Partial error
- `[main] timeout`: Operation timeout

#### Success Case - No Repositories Found

```text
No Git repositories found
```

**Exit Code**: 0 (success, no error occurred)

#### Empty Directory Case

```text
No Git repositories found
```

**Exit Code**: 0

### Standard Error (stderr)

#### Progress Indicator

During scanning (shown while processing):

```text
⠋ Scanning repositories...
```

**Progress Indicator Characteristics**:

- Animated spinner (various Unicode spinner characters)
- Message: "Scanning repositories..."
- Shown immediately upon start (within 1 second)
- Automatically hidden when scan completes
- Written to stderr to keep stdout clean

#### Error Messages

**Permission Denied**:

```text
Error: permission denied accessing directory: /path/to/dir
```

**Invalid Git Repository**:

```text
Warning: corrupted repository at /path/to/repo: {error-details}
```

**General Errors**:

```text
Error: {error-description}
```

**Error Message Format**:

- Prefix: `Error:` or `Warning:`
- Description: Human-readable error message
- Context: Relevant path or operation details

## Exit Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Scan completed successfully (even if no repos found) |
| 1 | General Error | Fatal error occurred (e.g., cannot access current directory) |
| 2 | Invalid Usage | Invalid command-line usage (reserved for future use with flags) |

**Important**: Non-fatal errors (e.g., single repository errors) do NOT cause non-zero exit code. The tool continues processing other repositories and exits with 0.

## Execution Flow

1. **Start**: Begin execution in current working directory
2. **Show Progress**: Display spinner on stderr
3. **Scan**: Recursively traverse directories
4. **Detect**: Identify Git repositories
5. **Extract**: Get Git status for each repository (parallel)
6. **Build**: Construct tree structure
7. **Hide Progress**: Stop spinner
8. **Output**: Print formatted tree to stdout
9. **Exit**: Return appropriate exit code

## Performance Guarantees

- **Startup Latency**: Progress indicator appears within 1 second
- **Scan Performance**: 50+ repositories processed in under 10 seconds (typical hardware)
- **Timeout**: Individual repository operations timeout after 5-10 seconds
- **Concurrency**: Multiple repositories processed in parallel

## Error Handling Guarantees

- **Non-Fatal Errors**: Tool continues processing other repositories
- **Graceful Degradation**: Shows partial information when full status unavailable
- **No Crashes**: Handles corrupted repos, permission errors, timeouts without crashing
- **Error Indicators**: Inline indicators show which repos had issues

## Examples

### Example 1: Clean Repositories

**Command**:

```bash
gitree
```

**Output** (stdout):

```text
.
├── backend [main]
└── frontend [main]
```

**Exit Code**: 0

### Example 2: Mixed Status

**Command**:

```bash
gitree
```

**Output** (stdout):

```text
.
├── active-dev [feature/new-ui ↑3 *]
├── legacy [master ↓5]
└── experiments [DETACHED $]
```

**Exit Code**: 0

### Example 3: Nested Repositories

**Command**:

```bash
gitree
```

**Output** (stdout):

```text
.
├── monorepo [main]
└── projects/
    ├── project-a [develop ↑1]
    └── project-b [main ○]
```

**Exit Code**: 0

### Example 4: No Repositories

**Command**:

```bash
gitree
```

**Output** (stdout):

```text
No Git repositories found
```

**Exit Code**: 0

### Example 5: Bare Repository

**Command**:

```bash
gitree
```

**Output** (stdout):

```text
.
└── my-repo.git [main bare]
```

**Exit Code**: 0

### Example 6: Error Cases

**Command**:

```bash
gitree
```

**Output** (stdout):

```text
.
├── good-repo [main]
└── broken-repo error
```

**Output** (stderr):

```text
Warning: corrupted repository at /path/to/broken-repo: unable to read HEAD
```

**Exit Code**: 0 (non-fatal error)

### Example 7: Timeout Case

**Command**:

```bash
gitree
```

**Output** (stdout):

```text
.
├── fast-repo [main]
└── slow-repo [main] timeout
```

**Output** (stderr):

```text
Warning: timeout processing repository at /path/to/slow-repo
```

**Exit Code**: 0

## Contract Verification Tests

### Test 1: Output Format Validation

```bash
# Verify output matches tree-like structure
gitree | grep -E '^[├└│\s]+.*\[.*\]$'
```

### Test 2: Exit Code on Success

```bash
gitree
echo $?  # Must be 0
```

### Test 3: Stderr for Progress

```bash
# Verify progress goes to stderr
gitree 2>&1 >/dev/null | grep "Scanning"
```

### Test 4: Stdout Clean (no progress)

```bash
# Verify stdout contains only tree output
gitree 2>/dev/null | grep -v "Scanning"
```

### Test 5: No Repos Found

```bash
cd /tmp/empty_dir
gitree  # Should output "No Git repositories found"
```

### Test 6: Concurrent Processing Performance

```bash
# Create 50 test repos
for i in {1..50}; do git init test$i; done
time gitree  # Should complete in < 10 seconds
```

## Backward Compatibility

**Version 1.0.0**: Initial release, no backward compatibility concerns

**Future Versions**:

- Text output format is part of the contract and will remain stable
- Adding command-line flags will not break flag-less invocation
- Exit codes will remain stable
- Stdout/stderr separation will be maintained

## Constitutional Compliance

### CLI Interface Principle

✅ **Compliance**:

- Text input/output protocol followed
- Output to stdout, errors to stderr
- No GUI or interactive prompts
- Unix-composable design

### Observability Principle

✅ **Compliance**:

- Progress feedback via stderr
- Error messages for failures
- Status indicators for partial failures
- Clean stdout for data piping

## Future Extensions (Not in v1.0.0)

The following are NOT part of the v1.0.0 contract but may be added in future versions:

- **Flags**: `--json`, `--color`, `--depth=N`, `--help`, `--version`
- **JSON Output**: Structured output format for scripting
- **Color Support**: ANSI color codes for terminal display
- **Filtering**: Include/exclude patterns
- **Configuration**: `.gitreerc` configuration file
- **Depth Limiting**: Limit directory traversal depth

These extensions will be designed to maintain backward compatibility with v1.0.0 behavior when no flags are used.
