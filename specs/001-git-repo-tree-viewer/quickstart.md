# Quickstart Guide: Gitree

**Version**: 1.0.0
**Date**: 2025-11-13

## What is Gitree?

Gitree is a command-line tool that displays an ASCII tree of directories containing Git repositories. It shows inline Git status information (branch, ahead/behind commits, stashes, uncommitted changes) for each repository, making it easy to get an overview of multiple projects at a glance.

Think of it as the `tree` command, but only showing Git repositories with their status inline.

## Installation

### Prerequisites

- Go 1.25 or higher
- Git (for repositories being scanned, not required for gitree itself)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/user/gitree.git
cd gitree

# Build the binary
go build -o gitree cmd/gitree/main.go

# Install to $GOPATH/bin (optional)
go install ./cmd/gitree
```

### Using go install

```bash
go install github.com/user/gitree/cmd/gitree@latest
```

## Quick Start

### Basic Usage

Navigate to any directory and run:

```bash
gitree
```

That's it! No flags or arguments needed for basic usage.

### Example Output

```text
.
├── backend [main ↑2 *]
├── frontend [develop $ *]
└── mobile/
    ├── ios [feature/new-ui ↑5]
    └── android [main ○]
```

**What this shows**:

- `backend`: On `main` branch, 2 commits ahead of remote, has uncommitted changes (*)
- `frontend`: On `develop` branch, has stashes ($) and uncommitted changes (*)
- `mobile/ios`: On `feature/new-ui` branch, 5 commits ahead of remote
- `mobile/android`: On `main` branch, no remote configured (○)

## Understanding the Output

### Tree Structure

The output uses the same ASCII characters as the Unix `tree` command:

- `├──` : Branch connector (not last child)
- `└──` : Branch connector (last child)
- `│` : Vertical line (indentation for children)

### Git Status Indicators

Status information appears in brackets after each repository name: `[branch indicators]`

| Indicator | Meaning |
|-----------|---------|
| `main` | Current branch name |
| `DETACHED` | Detached HEAD state |
| `↑N` | N commits ahead of remote |
| `↓N` | N commits behind remote |
| `○` | No remote configured |
| `$` | Has stashed changes |
| `*` | Has uncommitted changes |
| `bare` | Bare repository (no working directory) |
| `error` | Error retrieving full information |
| `timeout` | Git operations timed out |

### Example Interpretations

| Output | Interpretation |
|--------|----------------|
| `[main]` | On main, in sync with remote, no changes |
| `[main ↑3]` | 3 commits ahead of remote |
| `[main ↓2]` | 2 commits behind remote |
| `[main ↑2 ↓1]` | Diverged: 2 ahead, 1 behind |
| `[develop $ *]` | Has stashes and uncommitted changes |
| `[DETACHED]` | HEAD detached (not on any branch) |
| `[main ○]` | No remote configured |
| `[main bare]` | Bare repository |
| `[main] timeout` | Git operations took too long |

## Common Scenarios

### Scenario 1: Check All Projects

You have multiple projects in a directory:

```bash
cd ~/projects
gitree
```

Instantly see the status of all your projects without navigating into each one.

### Scenario 2: Find Uncommitted Work

Look for repositories with the `*` indicator:

```bash
gitree | grep '\*'
```

Shows only repositories with uncommitted changes.

### Scenario 3: Find Unpushed Commits

Look for repositories with `↑` indicator:

```bash
gitree | grep '↑'
```

Shows only repositories with commits that haven't been pushed.

### Scenario 4: Check Sync Status

Quickly identify which repositories are out of sync with remotes:

```bash
gitree | grep -E '↑|↓'
```

### Scenario 5: Find Stashed Work

Look for repositories with the `$` indicator:

```bash
gitree | grep '\$'
```

## Troubleshooting

### No Repositories Found

**Issue**: Output shows "No Git repositories found"

**Solutions**:

- Ensure you're in a directory that contains Git repositories
- Check that subdirectories contain `.git` folders
- Verify you have read permissions for the directories

### Permission Denied Errors

**Issue**: Warnings about inaccessible directories appear on stderr

**Solutions**:

- Run with appropriate permissions
- Ignore the warnings if those directories aren't important
- Fix directory permissions: `chmod +r /path/to/dir`

### Slow Performance

**Issue**: Gitree takes a long time to complete

**Possible Causes**:

- Large number of repositories
- Repositories with very large commit histories
- Network latency checking remote status

**Solutions**:

- Wait for completion (progress spinner shows it's working)
- Check for repositories with network issues (timeout indicators)
- Ensure good network connectivity for remote queries

### Timeout Indicators

**Issue**: Some repositories show `timeout` indicator

**Meaning**: Git operations took longer than 5-10 seconds

**Solutions**:

- Network issues may be causing remote queries to hang
- Repository may have extremely large history
- Try running `git fetch` manually in those repositories to diagnose

### Error Indicators

**Issue**: Some repositories show `error` indicator

**Possible Causes**:

- Corrupted `.git` directory
- Incomplete repository clone
- Permission issues

**Solutions**:

- Check the repository manually: `cd /path/to/repo && git status`
- Run `git fsck` to check repository integrity
- Re-clone the repository if corrupted

## Tips & Tricks

### Combine with grep

Filter output using standard Unix tools:

```bash
# Only show repositories on feature branches
gitree | grep 'feature/'

# Show only repositories with issues (errors or timeouts)
gitree | grep -E 'error|timeout'

# Count total repositories
gitree | grep -c '^\(├\|└\)'
```

### Save Output

Save the tree to a file:

```bash
gitree > repo-status.txt
```

### Redirect Progress

Hide the progress spinner:

```bash
gitree 2>/dev/null
```

### Use in Scripts

Example script to alert on uncommitted changes:

```bash
#!/bin/bash
if gitree | grep -q '\*'; then
    echo "Warning: Some repositories have uncommitted changes"
    gitree | grep '\*'
    exit 1
fi
```

### Check Before Going Offline

Before disconnecting from network or going offline:

```bash
gitree
```

Quickly verify:

- No unpushed commits (`↑` indicator)
- No unpulled changes (`↓` indicator)
- All work is committed (no `*` indicator)

## Understanding Performance

### Speed Expectations

- **Small projects** (1-10 repos): < 1 second
- **Medium projects** (10-50 repos): 2-5 seconds
- **Large projects** (50+ repos): 5-10 seconds

Performance depends on:

- Number of repositories
- Disk I/O speed
- Network latency (for remote queries)
- Size of repository histories

### Parallel Processing

Gitree processes repositories in parallel (typically 10-20 at once), so having many repositories doesn't linearly increase execution time.

### Progress Feedback

The animated spinner appears within 1 second and indicates gitree is working. If you see the spinner, the tool hasn't frozen – it's processing your repositories.

## Next Steps

### Learn More

- Read the [CLI Contract](contracts/cli-contract.md) for detailed output format
- Review [Library API](contracts/library-api.md) to use gitree as a library
- Check [Data Model](data-model.md) for internal structure details

### Customize (Future)

Version 1.0.0 intentionally has no flags or configuration. Future versions may add:

- Color output
- JSON format
- Depth limiting
- Include/exclude patterns
- Custom indicator symbols

Stay simple for now – it just works!

## Getting Help

### Check Version

```bash
# Version 1.0.0 doesn't have --version flag yet
# Check binary build date
ls -l $(which gitree)
```

### Report Issues

If you encounter bugs or unexpected behavior:

1. Check the troubleshooting section above
2. Verify you're running the latest version
3. Create an issue at: <https://github.com/user/gitree/issues>

Include:

- Your operating system
- Go version (`go version`)
- Example output showing the issue
- Steps to reproduce

## FAQ

**Q: Why doesn't gitree show all my directories?**
A: Gitree only shows directories that contain Git repositories. Non-Git directories are intentionally hidden.

**Q: Can I show only specific repositories?**
A: Version 1.0.0 doesn't support filtering. Future versions may add this feature.

**Q: Does gitree modify my repositories?**
A: No. Gitree is read-only and never modifies repository state.

**Q: Why is ahead/behind info missing for some repos?**
A: If you see `○` instead of ahead/behind, the repository has no remote configured.

**Q: Can I use gitree in CI/CD pipelines?**
A: Yes! Gitree outputs to stdout with clean exit codes, making it suitable for scripts and automation.

**Q: Does gitree require Git to be installed?**
A: No. Gitree uses the go-git library and doesn't require Git binary. However, the repositories being scanned are obviously Git repositories.

**Q: Can I export the output as JSON?**
A: Not in version 1.0.0. This may be added in future versions.

**Q: How do I show only repositories with uncommitted changes?**
A: Use: `gitree | grep '\*'`

**Q: Does gitree work on Windows?**
A: Yes! Gitree is cross-platform and works on Linux, macOS, and Windows.
