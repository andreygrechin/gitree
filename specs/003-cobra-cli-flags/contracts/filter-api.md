# Filter API Contract

**Feature**: 003-cobra-cli-flags
**Package**: internal/cli
**Date**: 2025-11-17
**Last Updated**: 2025-11-21

## Overview

This contract defines the internal API for repository filtering functionality. By default, gitree shows only repositories needing attention (dirty repos). The `--all` flag disables filtering to show all repositories including clean ones. The API is library code that can be tested independently and used by the CLI.

---

## Type Definitions

### FilterOptions

**Package**: `internal/cli`

**Definition**:

```go
// FilterOptions configures repository filtering behavior
type FilterOptions struct {
    // ShowAll when true, disables filtering (shows all repos including clean ones)
    // When false (default), shows only repositories needing attention
    ShowAll bool
}
```

**Usage**:

```go
// Default behavior: show only repos needing attention (filtering enabled)
opts := cli.FilterOptions{ShowAll: false}
filtered := cli.FilterRepositories(repos, opts)

// With --all flag: show all repos including clean ones
opts := cli.FilterOptions{ShowAll: true}
all := cli.FilterRepositories(repos, opts)
```

**Contract**:

- Thread-safe (immutable after creation)
- Default value (zero value): `ShowAll = false` (filtering enabled, show only dirty repos)
- No validation required (boolean flag)

---

## Functions

### IsClean

**Signature**:

```go
// IsClean determines if a repository is in clean state per FR-008
// Returns true only if ALL clean conditions are met
// Returns false (fail-safe) if repo is nil, status is nil, or has errors
func IsClean(repo *models.Repository) bool
```

**Parameters**:

- `repo`: Repository to evaluate. Must not be nil.

**Return Value**:

- `true`: Repository is clean (all FR-008 conditions met)
- `false`: Repository is dirty (any FR-009 condition met) or unknown (fail-safe)

**Clean Conditions** (ALL must be true per FR-008):

1. `repo != nil` (repository exists)
2. `repo.GitStatus != nil` (status available)
3. `repo.GitStatus.Error == ""` (no status extraction error - note: Error is string type)
4. `repo.GitStatus.Branch == "main" || repo.GitStatus.Branch == "master"` (on primary branch)
5. `!repo.GitStatus.HasChanges` (no uncommitted changes - includes staged and unstaged)
6. `!repo.GitStatus.HasStashes` (no stashed changes)
7. `repo.GitStatus.HasRemote` (remote tracking configured)
8. `repo.GitStatus.Ahead == 0` (not ahead of remote)
9. `repo.GitStatus.Behind == 0` (not behind remote)
10. `!repo.GitStatus.IsDetached` (not in detached HEAD state)

**Fail-Safe Behavior**:

- If `repo == nil` → return `false` (defensive, no panic)
- If `repo.GitStatus == nil` → return `false` (unknown = dirty)
- If `repo.GitStatus.Error != ""` → return `false` (error = dirty)

**Example**:

```go
repo := &models.Repository{
    Path: "/workspace/project",
    Name: "project",
    GitStatus: &models.GitStatus{
        Branch:     "main",
        IsDetached: false,
        HasRemote:  true,
        Ahead:      0,
        Behind:     0,
        HasStashes: false,
        HasChanges: false,
        Error:      "", // Empty string = no error
    },
}

if cli.IsClean(repo) {
    // Exclude from default output (show only dirty repos)
} else {
    // Include in default output (repo needs attention)
}
```

**Contract**:

- MUST implement FR-008 precisely (10 conditions including nil checks)
- MUST use fail-safe defaults (unknown state = dirty)
- MUST be deterministic (same input → same output)
- MUST NOT modify input repository
- MUST return false (not panic) on nil repository (defensive programming)

**Test Cases**:

1. Clean repo (all conditions true) → true
2. Each condition false individually (10 tests) → false
3. Nil repository → false
4. Nil GitStatus → false
5. Non-empty Error string in GitStatus → false
6. Feature branch (not main/master) → false
7. Main branch with changes → false
8. Main branch with stashes → false
9. Detached HEAD → false
10. No remote → false
11. Ahead of remote → false
12. Behind remote → false

---

### FilterRepositories

**Signature**:

```go
// FilterRepositories returns filtered repository list based on options
// Applies filtering without modifying input slice
// Maintains original order of repositories
func FilterRepositories(repos []*models.Repository, opts FilterOptions) []*models.Repository
```

**Parameters**:

- `repos`: Slice of repositories to filter. Can be nil or empty.
- `opts`: Filtering configuration

**Return Value**:

- New slice containing only repositories matching filter criteria
- Empty slice if no repositories match (when ShowAll=false and all repos are clean)
- Original slice if `opts.ShowAll == true` (no filtering, show all repos)
- Never returns nil (returns empty slice instead when no matches)

**Behavior**:

**When `opts.ShowAll == true`** (with --all flag):

```go
// No filtering - return original slice
return repos
```

**When `opts.ShowAll == false`** (default - show only dirty repos):

```go
filtered := make([]*models.Repository, 0, len(repos))
for _, repo := range repos {
    if !IsClean(repo) {  // Include repos needing attention
        filtered = append(filtered, repo)
    }
}
return filtered
```

**Examples**:

**No Filtering (--all flag)**:

```go
repos := []*models.Repository{cleanRepo, dirtyRepo}
opts := cli.FilterOptions{ShowAll: true}
result := cli.FilterRepositories(repos, opts)
// result == [cleanRepo, dirtyRepo]  (all repos shown)
```

**Default Filtering (show only dirty repos)**:

```go
repos := []*models.Repository{cleanRepo, dirtyRepo}
opts := cli.FilterOptions{ShowAll: false}
result := cli.FilterRepositories(repos, opts)
// result == [dirtyRepo]  (cleanRepo excluded, only repos needing attention shown)
```

**All Clean Repos (default mode)**:

```go
repos := []*models.Repository{cleanRepo1, cleanRepo2}
opts := cli.FilterOptions{ShowAll: false}
result := cli.FilterRepositories(repos, opts)
// result == []  (empty slice, not nil - no repos need attention)
```

**Empty Input**:

```go
repos := []*models.Repository{}
opts := cli.FilterOptions{ShowAll: false}
result := cli.FilterRepositories(repos, opts)
// result == []  (empty slice)
```

**Contract**:

- MUST NOT modify input slice
- MUST preserve order of repositories
- MUST return empty slice (not nil) when no matches
- MUST call IsClean() for each repository
- MUST be thread-safe (no shared mutable state)
- O(n) time complexity (single pass)

**Test Cases**:

1. Empty input → empty output
2. All clean repos with ShowAll=false (default) → empty output
3. All dirty repos with ShowAll=false → all included
4. Mixed clean/dirty with ShowAll=false → only dirty included
5. ShowAll=true → all included (no filtering)
6. Order preservation: verify output order matches input order
7. Non-modification: verify input slice unchanged after call

---

## Integration with Tree Building

### Current Tree Building Flow

**Existing**:

```go
// 1. Scan repositories
scanResult, _ := scanner.Scan(ctx, opts)

// 2. Extract status
statuses, _ := gitstatus.ExtractBatch(ctx, repoMap, statusOpts)

// 3. Build tree
root := tree.Build(cwd, scanResult.Repositories, nil)

// 4. Format
output := tree.Format(root, nil)
```

### Modified Flow with Filtering

**Filter Before Build (Implemented Approach)**:

```go
// 1. Scan repositories
scanResult, _ := scanner.Scan(ctx, opts)

// 2. Extract status
statuses, _ := gitstatus.ExtractBatch(ctx, repoMap, statusOpts)

// 3. Filter repositories (based on --all flag)
filterOpts := cli.FilterOptions{ShowAll: allFlag}
filtered := cli.FilterRepositories(scanResult.Repositories, filterOpts)

// 4. Build tree with filtered list
root := tree.Build(cwd, filtered, nil)

// 5. Format
output := tree.Format(root, nil)
```

**Implementation Benefits**:

- Simpler: Filtering and tree building are separate concerns
- Testable: Can unit test FilterRepositories() independently (see internal/cli/filter_test.go)
- Flexible: Can filter and inspect results before tree building
- Library-first: Filtering is a standalone library function in internal/cli package

---

## Tree Structure Preservation

### Challenge

When filtering repositories, we must maintain valid tree structure while removing empty branches (FR-009, FR-010).

**Example Problem**:

```text
Original Tree:
/workspace
├── clean-proj/     [main]     ← exclude
└── nested/
    ├── clean-proj/ [master]   ← exclude
    └── dirty-proj/ [feature]  ← include

Desired Output:
/workspace
└── nested/
    └── dirty-proj/ [feature]

Not This:
/workspace
└── nested/         ← empty directory node - should be kept to preserve hierarchy
```

### Solution

**Approach**: Post-filtering tree pruning

1. **Filter repositories** using FilterRepositories()
2. **Build tree** with filtered list (creates intermediate directory nodes)
3. **Prune empty branches** (optional - may keep structure for clarity)

**Tree Pruning** (optional function):

```go
// PruneEmptyBranches removes directory nodes with no repository descendants
// Returns modified tree with empty branches removed
func PruneEmptyBranches(root *TreeNode) *TreeNode
```

**Contract**:

- FR-009: Keep parent directories leading to filtered repos
- FR-010: Remove directory branches containing only excluded repos
- Preserve visual hierarchy (tree structure)

**Decision**: Determine during implementation whether to prune empty branches or keep directory structure for clarity. Spec allows both approaches.

---

## Error Handling

### IsClean Errors

**Nil Repository** (fail-safe):

```go
repo := nil
cli.IsClean(repo)  // Returns false (defensive, no panic)
```

**Nil GitStatus** (fail-safe):

```go
repo := &models.Repository{GitStatus: nil}
cli.IsClean(repo)  // Returns false (unknown = dirty)
```

**Status Extraction Error** (fail-safe):

```go
repo := &models.Repository{
    GitStatus: &models.GitStatus{
        Error: "timeout", // Note: Error is string type
    },
}
cli.IsClean(repo)  // Returns false (error = dirty)
```

### FilterRepositories Errors

**No Panics**: Function is defensive and handles all edge cases gracefully

- Nil input → empty slice
- Empty input → empty slice
- Invalid options → use defaults

---

## Performance Considerations

### Time Complexity

- `IsClean()`: O(1) - constant time condition checks
- `FilterRepositories()`: O(n) - single pass through repository list
- Combined: O(n) - linear in number of repositories

### Space Complexity

- `IsClean()`: O(1) - no additional allocation
- `FilterRepositories()`: O(n) - new slice for filtered results
- Worst case: O(n) when all repositories are dirty (full copy)
- Best case: O(1) when all repositories are clean (empty slice)

### Optimization Opportunities

**Pre-allocation**:

```go
// Allocate for worst case (all dirty)
filtered := make([]*models.Repository, 0, len(repos))
```

**Short-circuit**: No optimization possible - must evaluate all repositories

**Concurrency**: Not needed - filtering is fast (pure CPU, no I/O)

---

## Testing Contract

### Unit Tests Required

**IsClean() Tests** (see internal/cli/filter_test.go):

1. All clean conditions true (main branch) → true
2. All clean conditions true (master branch) → true
3. Feature branch → false
4. Uncommitted changes → false
5. Has stashes → false
6. No remote → false
7. Ahead of remote → false
8. Behind remote → false
9. Detached HEAD → false
10. Status error (non-empty Error string) → false
11. Nil GitStatus → false
12. Nil repository → false (no panic)

**FilterRepositories() Tests** (see internal/cli/filter_test.go):

1. Empty input, any options → empty output
2. All clean, ShowAll=false (default) → empty output
3. All clean, ShowAll=true → all included
4. All dirty, ShowAll=false → all included
5. All dirty, ShowAll=true → all included
6. Mixed clean/dirty, ShowAll=false → only dirty included
7. Mixed clean/dirty, ShowAll=true → all included
8. Order preservation check
9. Input non-modification check

### Integration Tests Required

**End-to-End Filtering**:

1. Create test repo tree (known clean/dirty)
2. Run scan → status extraction → filter
3. Verify correct repositories filtered
4. Verify tree structure preserved

---

## Summary

This API provides clean separation of concerns:

1. **IsClean()**: Pure function for clean-state determination (FR-008)
2. **FilterRepositories()**: Pure function for list filtering based on ShowAll flag
3. **Integration**: Filter before tree building (library-first approach)

Key contracts:

- Fail-safe defaults (unknown = dirty, nil = false)
- No input modification
- Thread-safe (pure functions)
- O(n) performance
- Comprehensive test coverage in internal/cli/filter_test.go

**Implementation Status**: ✅ Fully implemented in internal/cli/filter.go with complete test coverage

**Key Implementation Details**:

- Package: `internal/cli` (not `internal/tree`)
- FilterOptions uses `ShowAll bool` (not `DirtyOnly bool`)
- GitStatus.Error is `string` type (not `error` type)
- GitStatus uses `HasChanges bool` (not `HasUncommittedChanges bool`)
- GitStatus uses `HasStashes bool` (not `Stashes int`)
- IsClean returns false for nil repository (defensive, no panic)
