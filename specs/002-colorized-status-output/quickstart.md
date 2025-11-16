# Quickstart: Colorized Status Output

**Feature**: 002-colorized-status-output
**Prerequisites**: Completion of spec 001-git-repo-tree-viewer
**Estimated Time**: 2-3 hours

## Overview

This guide walks through implementing colorized Git status output for gitree. The implementation adds ANSI color codes to the status display while maintaining backward compatibility (except for bracket format change from `[` to `[[`).

## Prerequisites

1. Spec 001 (git-repo-tree-viewer) fully implemented and tested
2. Go 1.25+ installed
3. Familiarity with ANSI color codes and terminal capabilities
4. Understanding of existing `GitStatus.Format()` method

## Step 1: Upgrade Dependencies

Add explicit dependency on fatih/color latest version:

```bash
cd /Users/andrey/repos/gitree
go get github.com/fatih/color@v1.18.0
go mod tidy
```

**Verification**:
```bash
grep fatih/color go.mod
# Should show: github.com/fatih/color v1.18.0
```

## Step 2: Understand Current Implementation

**Read**: `internal/models/repository.go`

Current `GitStatus.Format()` method:
- Returns string like `[main ↑2 ↓1 $ *]`
- Single brackets
- No color codes
- Builds string by appending parts

**Key Logic**:
1. Branch name or DETACHED
2. Ahead/Behind indicators (or ○ for no remote)
3. Stash indicator ($)
4. Uncommitted changes indicator (*)
5. Optional "bare" indicator
6. Optional error suffix

## Step 3: Write Tests (TDD - Required)

**File**: `internal/models/repository_test.go`

Add test functions for colorized output:

```go
func TestGitStatusFormatColorizedMainBranch(t *testing.T) {
    // Test main branch appears in gray
}

func TestGitStatusFormatColorizedFeatureBranch(t *testing.T) {
    // Test feature branch appears in yellow
}

func TestGitStatusFormatColorizedAheadBehind(t *testing.T) {
    // Test ahead in green, behind in red
}

func TestGitStatusFormatColorizedAttentionIndicators(t *testing.T) {
    // Test stashes/changes in red
}

func TestGitStatusFormatDoubleBrackets(t *testing.T) {
    // Test output uses [[ ]] instead of [ ]
}

func TestGitStatusFormatNoColor(t *testing.T) {
    // Test with color.NoColor = true
    // Verify no ANSI codes but double brackets preserved
}
```

**Run Tests** (should fail - Red phase of TDD):
```bash
make test
# OR
go test ./internal/models/
```

## Step 4: Implement Color Functions

**File**: `internal/models/repository.go`

Add package-level color function variables after imports:

```go
package models

import (
    "errors"
    "fmt"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "github.com/fatih/color"  // Add this import
)

// Add after import block, before type definitions
var (
    grayColor   = color.New(color.FgHiBlack).SprintFunc()
    yellowColor = color.New(color.FgYellow).SprintFunc()
    greenColor  = color.New(color.FgGreen).SprintFunc()
    redColor    = color.New(color.FgRed).SprintFunc()
)
```

**Why Package-Level**:
- Created once, reused across all Format() calls
- No initialization overhead per call
- Thread-safe (SprintFunc() returns immutable function)

## Step 5: Modify GitStatus.Format() Method

**File**: `internal/models/repository.go`

Replace the existing `Format()` method with colorized version:

```go
func (g *GitStatus) Format() string {
    var parts []string

    // Branch: gray for main/master, yellow otherwise
    if g.Branch == "main" || g.Branch == "master" {
        parts = append(parts, grayColor(g.Branch))
    } else {
        parts = append(parts, yellowColor(g.Branch))
    }

    // Ahead/Behind: green/red, or gray no-remote indicator
    if g.HasRemote {
        if g.Ahead > 0 {
            parts = append(parts, greenColor(fmt.Sprintf("↑%d", g.Ahead)))
        }
        if g.Behind > 0 {
            parts = append(parts, redColor(fmt.Sprintf("↓%d", g.Behind)))
        }
    } else {
        parts = append(parts, grayColor("○"))
    }

    // Stashes: red
    if g.HasStashes {
        parts = append(parts, redColor("$"))
    }

    // Uncommitted changes: red
    if g.HasChanges {
        parts = append(parts, redColor("*"))
    }

    // Join with separator and wrap in gray double brackets
    result := grayColor("[[") + " " + strings.Join(parts, " "+grayColor("|")+" ") + " " + grayColor("]]")

    // Error suffix (if applicable)
    if g.Error != "" {
        result += " error"
    }

    return result
}
```

**Note**: The separator `|` is only added between parts, not before the first part or after the last.

## Step 6: Update Repository Format

**File**: `internal/models/repository.go`

Find the section handling "bare" indicator (if applicable) and update to use red color:

```go
// If GitStatus.Format() is called for bare repos, ensure "bare" appears in red
// This may be in the tree formatter or Repository's display logic
// Check existing implementation and apply redColor("bare")
```

**Check**: Search codebase for "bare" indicator formatting and apply red color.

## Step 7: Run Tests (Green Phase)

```bash
make test
# OR
go test ./internal/models/ -v
```

**Expected**:
- All new colorization tests pass
- All existing tests pass (regression check)
- Color codes present in output when colors enabled
- No color codes when `color.NoColor = true`

## Step 8: Manual Testing

Build and test the binary:

```bash
make build
./bin/gitree
```

**Test Cases**:

1. **Terminal output** (should show colors):
   ```bash
   ./bin/gitree
   ```

2. **Piped output** (should NOT show colors):
   ```bash
   ./bin/gitree | cat
   ```

3. **NO_COLOR environment** (should NOT show colors):
   ```bash
   NO_COLOR=1 ./bin/gitree
   ```

4. **Redirected output** (should NOT show colors):
   ```bash
   ./bin/gitree > output.txt
   cat output.txt  # No ANSI codes
   ```

## Step 9: Verify Color Output

**Check**: Run gitree in a terminal and verify:
- Main/master branches appear in gray
- Feature branches appear in yellow
- Ahead commits (↑) appear in green
- Behind commits (↓) appear in red
- Stashes ($) appear in red
- Uncommitted changes (*) appear in red
- Brackets `[[` and `]]` appear in gray
- Separator `|` appears in gray

**Visual Inspection**: Use a terminal with both dark and light themes to ensure gray color has sufficient contrast.

## Step 10: Code Quality Checks

Run linting and formatting:

```bash
make lint
# OR
make format && go vet ./... && staticcheck ./...
```

**Fix any issues** reported by linters.

## Step 11: Integration Testing

Test with various repository states:

1. Create test repositories with different states:
   ```bash
   # Create test directory
   mkdir /tmp/gitree-test && cd /tmp/gitree-test

   # Repo on main, in sync
   git init repo-main && cd repo-main
   git commit --allow-empty -m "initial"
   cd ..

   # Repo on feature branch with ahead commits
   git init repo-feature && cd repo-feature
   git commit --allow-empty -m "initial"
   git checkout -b feature/test
   git commit --allow-empty -m "change"
   cd ..

   # Repo with uncommitted changes
   git init repo-changes && cd repo-changes
   echo "test" > file.txt
   git add file.txt
   cd ..
   ```

2. Run gitree in test directory:
   ```bash
   cd /tmp/gitree-test
   /path/to/gitree/bin/gitree
   ```

3. Verify colors for each repository state

## Step 12: Update Documentation

**Optional**: Update README.md with color output examples (if showing sample output).

**Note**: CLAUDE.md already documents the feature in "Project Overview" and "Key Implementation Details".

## Common Issues

### Issue: Colors not appearing in terminal

**Cause**: Terminal doesn't support ANSI colors or NO_COLOR is set

**Fix**:
- Verify terminal supports colors: `echo -e "\033[31mRed\033[0m"`
- Check NO_COLOR: `echo $NO_COLOR` (should be empty)
- Test explicitly: `color.NoColor = false` in code

### Issue: Tests fail due to ANSI codes

**Cause**: Tests expecting plain text output

**Fix**: Use `color.NoColor = true` in test setup or check for ANSI codes in assertions

### Issue: Colors appear in piped output

**Cause**: Terminal detection not working

**Fix**: Verify fatih/color version (should be v1.18.0) and check `term.IsTerminal()` logic

### Issue: Gray color not visible

**Cause**: Terminal theme has poor contrast

**Fix**: Consider using `color.FgWhite` instead of `color.FgHiBlack` (test on both light/dark themes)

## Completion Checklist

- [ ] Dependencies upgraded (`fatih/color@v1.18.0`)
- [ ] Tests written (TDD - before implementation)
- [ ] Tests pass (all new and existing)
- [ ] `GitStatus.Format()` returns colorized output
- [ ] Double brackets `[[` `]]` used instead of single
- [ ] Color codes respect terminal detection
- [ ] NO_COLOR environment variable honored
- [ ] Manual testing completed (terminal, pipe, redirect)
- [ ] Code quality checks pass (`make lint`)
- [ ] Integration testing with various repo states
- [ ] Visual verification on dark and light terminal themes

## Next Steps

After completing this feature:

1. Run `/speckit.tasks` to generate task breakdown
2. Implement tasks following TDD workflow
3. Create PR with colorization changes
4. Request review focusing on color accessibility and terminal compatibility

## Reference

- Feature Spec: [spec.md](spec.md)
- Implementation Plan: [plan.md](plan.md)
- Research: [research.md](research.md)
- Data Model: [data-model.md](data-model.md)
