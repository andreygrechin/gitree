# Filter API Contract

**Feature**: 003-cobra-cli-flags
**Package**: internal/tree
**Date**: 2025-11-17

## Overview

This contract defines the internal API for repository filtering functionality used by the --dirty-only flag. The API is library code that can be tested independently and used by the CLI.

---

## Type Definitions

### FilterOptions

**Package**: `internal/tree`

**Definition**:

```go
// FilterOptions configures repository filtering behavior
type FilterOptions struct {
    // DirtyOnly excludes repositories in clean state
    // Clean state definition per FR-007
    DirtyOnly bool
}
```

**Usage**:

```go
opts := tree.FilterOptions{DirtyOnly: true}
filtered := tree.FilterRepositories(repos, opts)
```

**Contract**:

- Thread-safe (immutable after creation)
- Default value (zero value): `DirtyOnly = false` (show all)
- No validation required (boolean flag)

---

## Functions

### IsClean

**Signature**:

```go
// IsClean determines if a repository is in clean state per FR-007
// Returns true only if ALL clean conditions are met
// Returns false (fail-safe) if status is nil or has errors
func IsClean(repo *models.Repository) bool
```

**Parameters**:

- `repo`: Repository to evaluate. Must not be nil.

**Return Value**:

- `true`: Repository is clean (all FR-007 conditions met)
- `false`: Repository is dirty (any FR-008 condition met) or unknown (fail-safe)

**Clean Conditions** (ALL must be true):

1. `repo.GitStatus != nil` (status available)
2. `repo.GitStatus.Error == nil` (no status extraction error)
3. `repo.GitStatus.Branch == "main" || repo.GitStatus.Branch == "master"` (on primary branch)
4. `!repo.GitStatus.HasUncommittedChanges` (no staged/unstaged changes)
5. `repo.GitStatus.Stashes == 0` (no stashed changes)
6. `repo.GitStatus.HasRemote` (remote tracking configured)
7. `repo.GitStatus.Ahead == 0` (not ahead of remote)
8. `repo.GitStatus.Behind == 0` (not behind remote)
9. `!repo.GitStatus.IsDetached` (not in detached HEAD state)

**Fail-Safe Behavior**:

- If `repo == nil` → panic (caller error)
- If `repo.GitStatus == nil` → return `false` (unknown = dirty)
- If `repo.GitStatus.Error != nil` → return `false` (error = dirty)

**Example**:

```go
repo := &models.Repository{
    Path: "/workspace/project",
    GitStatus: &models.GitStatus{
        Branch:                "main",
        HasUncommittedChanges: false,
        Stashes:               0,
        HasRemote:             true,
        Ahead:                 0,
        Behind:                0,
        IsDetached:            false,
        Error:                 nil,
    },
}

if tree.IsClean(repo) {
    // Exclude from dirty-only output
} else {
    // Include in dirty-only output
}
```

**Contract**:

- MUST implement FR-007 precisely (9 conditions)
- MUST use fail-safe defaults (unknown state = dirty)
- MUST be deterministic (same input → same output)
- MUST NOT modify input repository
- MUST panic on nil repository (programming error)

**Test Cases**:

1. Clean repo (all conditions true) → true
2. Each condition false individually (9 tests) → false
3. Nil GitStatus → false
4. Non-nil Error in GitStatus → false
5. Feature branch (not main/master) → false
6. Main branch with changes → false
7. Detached HEAD → false
8. No remote → false

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
- Empty slice if no repositories match
- Original slice if `opts.DirtyOnly == false` (no filtering)
- Never returns nil (returns empty slice instead)

**Behavior**:

**When `opts.DirtyOnly == false`** (default):

```go
// No filtering - return original slice
return repos
```

**When `opts.DirtyOnly == true`**:

```go
filtered := make([]*models.Repository, 0, len(repos))
for _, repo := range repos {
    if !IsClean(repo) {  // Dirty repos only
        filtered = append(filtered, repo)
    }
}
return filtered
```

**Examples**:

**No Filtering**:

```go
repos := []*models.Repository{cleanRepo, dirtyRepo}
opts := tree.FilterOptions{DirtyOnly: false}
result := tree.FilterRepositories(repos, opts)
// result == [cleanRepo, dirtyRepo]  (unchanged)
```

**Dirty-Only Filtering**:

```go
repos := []*models.Repository{cleanRepo, dirtyRepo}
opts := tree.FilterOptions{DirtyOnly: true}
result := tree.FilterRepositories(repos, opts)
// result == [dirtyRepo]  (cleanRepo excluded)
```

**All Clean**:

```go
repos := []*models.Repository{cleanRepo1, cleanRepo2}
opts := tree.FilterOptions{DirtyOnly: true}
result := tree.FilterRepositories(repos, opts)
// result == []  (empty slice, not nil)
```

**Empty Input**:

```go
repos := []*models.Repository{}
opts := tree.FilterOptions{DirtyOnly: true}
result := tree.FilterRepositories(repos, opts)
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
2. All clean repos with DirtyOnly → empty output
3. All dirty repos with DirtyOnly → all included
4. Mixed clean/dirty → only dirty included
5. DirtyOnly=false → all included (no filtering)
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

**Option A: Filter Before Build** (Recommended):

```go
// 1. Scan repositories
scanResult, _ := scanner.Scan(ctx, opts)

// 2. Extract status
statuses, _ := gitstatus.ExtractBatch(ctx, repoMap, statusOpts)

// 3. Filter repositories (NEW)
filterOpts := tree.FilterOptions{DirtyOnly: dirtyOnlyFlag}
filtered := tree.FilterRepositories(scanResult.Repositories, filterOpts)

// 4. Build tree with filtered list
root := tree.Build(cwd, filtered, nil)

// 5. Format
output := tree.Format(root, nil)
```

**Option B: Filter During Build**:

```go
// 3. Build tree with filter options
buildOpts := tree.BuildOptions{
    FilterOptions: tree.FilterOptions{DirtyOnly: dirtyOnlyFlag},
}
root := tree.Build(cwd, scanResult.Repositories, &buildOpts)
```

**Recommendation**: Option A (filter before build)

- Simpler: Filtering and tree building are separate concerns
- Testable: Can unit test FilterRepositories() independently
- Flexible: Can filter and inspect results before tree building
- Library-first: Filtering is a standalone library function

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

**Nil Repository**:

```go
repo := nil
tree.IsClean(repo)  // PANIC: programming error
```

**Nil GitStatus** (fail-safe):

```go
repo := &models.Repository{GitStatus: nil}
tree.IsClean(repo)  // Returns false (unknown = dirty)
```

**Status Extraction Error** (fail-safe):

```go
repo := &models.Repository{
    GitStatus: &models.GitStatus{
        Error: errors.New("timeout"),
    },
}
tree.IsClean(repo)  // Returns false (error = dirty)
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

**IsClean() Tests**:

1. All clean conditions true → true
2. Each clean condition false (9 tests) → false
3. Nil GitStatus → false
4. Non-nil Error → false
5. Nil repository → panic

**FilterRepositories() Tests**:

1. Empty input, any options → empty output
2. All clean, DirtyOnly=true → empty output
3. All dirty, DirtyOnly=true → all included
4. Mixed, DirtyOnly=true → only dirty
5. Any input, DirtyOnly=false → all included
6. Order preservation check
7. Input non-modification check

### Integration Tests Required

**End-to-End Filtering**:

1. Create test repo tree (known clean/dirty)
2. Run scan → status extraction → filter
3. Verify correct repositories filtered
4. Verify tree structure preserved

---

## Summary

This API provides clean separation of concerns:

1. **IsClean()**: Pure function for clean-state determination (FR-007/FR-008)
2. **FilterRepositories()**: Pure function for list filtering
3. **Integration**: Filter before tree building (library-first)

Key contracts:

- Fail-safe defaults (unknown = dirty)
- No input modification
- Thread-safe (pure functions)
- O(n) performance
- Comprehensive test coverage
