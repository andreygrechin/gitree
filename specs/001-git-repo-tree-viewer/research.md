# Research: Git Repository Tree Viewer

**Feature**: 001-git-repo-tree-viewer
**Date**: 2025-11-13
**Purpose**: Document technical decisions, best practices, and architectural choices for gitree implementation

## Research Questions & Decisions

### 1. Go-Git Library Usage for Git Operations

**Decision**: Use `github.com/go-git/go-git/v5` for all Git operations

**Rationale**:

- Pure Go implementation - no external Git binary dependency required
- Cross-platform support (Linux, macOS, Windows) without additional installation
- Rich API for accessing repository information (branches, remotes, status, stashes)
- Active maintenance and good documentation
- Type-safe access to Git objects and metadata
- Built-in support for handling bare repositories

**Alternatives Considered**:

- **Exec-based approach (git CLI)**: Would require Git installation, harder to parse output, less type-safe, more error-prone
- **git2go (libgit2 bindings)**: Requires CGO and libgit2 installation, adds build complexity and portability issues
- **dugite-go**: Less mature, smaller community

**Best Practices**:

- Use `PlainOpen()` for regular repositories and handle bare repositories with `PlainOpenWithOptions()`
- Set context timeouts (5-10 seconds) for all Git operations to prevent hanging
- Use `StatusOptions` to optimize status checks (only check what's needed)
- Cache repository references when multiple operations are needed
- Handle `ErrRepositoryNotExists` and other common errors gracefully

**Key APIs**:

```go
// Open repository
repo, err := git.PlainOpen(path)

// Get HEAD reference
head, err := repo.Head()

// Get remote tracking info
remote, err := repo.Remote("origin")

// Get status (uncommitted changes)
worktree, err := repo.Worktree()
status, err := worktree.Status()

// Count stashes
storer := repo.Storer
iter, err := storer.IterReferences()
// Filter for stash refs
```

### 2. Concurrent Repository Processing with Goroutines

**Decision**: Process repositories concurrently using worker pool pattern with goroutines

**Rationale**:

- Git operations are I/O bound (disk reads, potential network for remote tracking)
- Scanning dozens of repositories sequentially would be slow
- Go's goroutines provide lightweight concurrency
- Worker pool prevents spawning unlimited goroutines for large repository counts
- Enables meeting performance goal of 50+ repos in <10 seconds

**Alternatives Considered**:

- **Sequential processing**: Simple but too slow for multiple repositories
- **Unlimited goroutines**: Could overwhelm system with thousands of repositories
- **External concurrency library**: Unnecessary complexity, Go's standard library sufficient

**Best Practices**:

- Use semaphore or buffered channel to limit concurrent workers (e.g., 10-20 workers)
- Use `sync.WaitGroup` to wait for all goroutines to complete
- Use channels to collect results from goroutines
- Implement context with timeout for graceful shutdown
- Protect shared data structures with mutexes if needed (prefer channels)

**Implementation Pattern**:

```go
type result struct {
    path string
    status *GitStatus
    err error
}

func processRepositories(paths []string) []result {
    results := make(chan result, len(paths))
    sem := make(chan struct{}, 10) // Limit to 10 concurrent workers
    var wg sync.WaitGroup

    for _, path := range paths {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()
            sem <- struct{}{}        // Acquire
            defer func() { <-sem }() // Release

            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()

            status, err := getGitStatus(ctx, p)
            results <- result{path: p, status: status, err: err}
        }(path)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    var collected []result
    for r := range results {
        collected = append(collected, r)
    }
    return collected
}
```

### 3. Directory Traversal and Git Detection

**Decision**: Use `filepath.WalkDir` with custom logic to detect Git repositories and handle symlinks

**Rationale**:

- `filepath.WalkDir` is efficient (uses ReadDir internally, better than Walk)
- Provides full control over traversal logic
- Can skip non-repository directories early
- Easy to implement symlink handling (follow once, track visited paths)

**Alternatives Considered**:

- **filepath.Walk**: Less efficient than WalkDir (stat calls for each entry)
- **Manual recursive traversal**: More code, reinventing the wheel
- **Third-party traversal library**: Unnecessary dependency

**Best Practices**:

- Check for `.git` directory to identify repositories
- For bare repos, check for `HEAD`, `refs/`, `objects/` directories
- Follow symlinks once using `filepath.EvalSymlinks()` and track visited inodes to prevent loops
- Skip nested repository contents after detecting a Git root (per spec FR-018)
- Handle permission errors gracefully with `os.IsPermission(err)`
- Use `fs.SkipDir` to skip directories that don't need traversal

**Git Detection Logic**:

```go
func isGitRepository(path string) (bool, bool) {
    // Check for regular repository
    gitDir := filepath.Join(path, ".git")
    if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
        return true, false // regular repo
    }

    // Check for bare repository
    headFile := filepath.Join(path, "HEAD")
    refsDir := filepath.Join(path, "refs")
    objsDir := filepath.Join(path, "objects")

    if _, err1 := os.Stat(headFile); err1 == nil {
        if info2, err2 := os.Stat(refsDir); err2 == nil && info2.IsDir() {
            if info3, err3 := os.Stat(objsDir); err3 == nil && info3.IsDir() {
                return true, true // bare repo
            }
        }
    }

    return false, false
}
```

### 4. Spinner/Progress Indicator Implementation

**Decision**: Use `github.com/briandowns/spinner` for progress indication

**Rationale**:

- Mature, widely-used library (9k+ stars)
- Simple API for showing/hiding spinner
- Multiple spinner styles available
- Automatically handles terminal detection
- Thread-safe for concurrent usage
- Minimal footprint

**Alternatives Considered**:

- **Manual spinner implementation**: More code, need to handle terminal control
- **github.com/schollz/progressbar**: Overkill for simple progress indication (shows percentages, not needed)
- **github.com/vbauerster/mpb**: Multi-progress bar library, too complex for simple spinner

**Best Practices**:

- Start spinner before launching goroutines
- Write spinner output to stderr to keep stdout clean (per constitution)
- Stop spinner before printing results to stdout
- Use simple spinner style that works across terminals (e.g., dots, line)
- Set appropriate message (e.g., "Scanning repositories...")

**Usage Pattern**:

```go
import "github.com/briandowns/spinner"

s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
s.Suffix = " Scanning repositories..."
s.Writer = os.Stderr // Keep stdout clean
s.Start()

// Process repositories...

s.Stop()
// Print results to stdout
```

### 5. ASCII Tree Formatting

**Decision**: Implement custom tree formatter matching Unix `tree` command style

**Rationale**:

- Tree command uses specific box-drawing characters: `├──`, `└──`, `│`
- Need to track depth and whether node is last child for correct formatting
- Must integrate Git status inline with repository names
- Custom implementation allows exact control over formatting

**Alternatives Considered**:

- **Third-party tree library**: No Go library matches exact tree command output with custom inline info
- **Template-based rendering**: Overkill, tree structure is simple

**Best Practices**:

- Use Unicode box-drawing characters for visual tree structure
- Track parent-child relationships to determine connector characters
- Format: `{prefix}{connector}{repo-name} {git-status}`
- Example output:

  ```text
  .
  ├── project1 [main ↑2 ↓1 *]
  ├── project2 [develop $]
  └── nested/
      └── project3 [DETACHED]
  ```

**Formatting Algorithm**:

```go
type TreeNode struct {
    Path     string
    GitInfo  *GitStatus
    Children []*TreeNode
    IsLast   bool
}

func formatTree(node *TreeNode, prefix string) string {
    var result strings.Builder

    connector := "├── "
    if node.IsLast {
        connector = "└── "
    }

    result.WriteString(prefix)
    result.WriteString(connector)
    result.WriteString(filepath.Base(node.Path))

    if node.GitInfo != nil {
        result.WriteString(" ")
        result.WriteString(formatGitStatus(node.GitInfo))
    }
    result.WriteString("\n")

    childPrefix := prefix
    if node.IsLast {
        childPrefix += "    "
    } else {
        childPrefix += "│   "
    }

    for i, child := range node.Children {
        child.IsLast = (i == len(node.Children)-1)
        result.WriteString(formatTree(child, childPrefix))
    }

    return result.String()
}
```

### 6. Git Status Information Extraction

**Decision**: Extract branch, ahead/behind, stashes, and uncommitted changes using go-git API

**Rationale**:

- All required information is accessible via go-git API
- Type-safe access to Git objects
- No string parsing needed (unlike exec-based approach)
- Can implement precise error handling per status component

**Key Implementation Details**:

**Branch Name / Detached HEAD**:

```go
head, err := repo.Head()
if err != nil {
    return "error", nil
}

if !head.Name().IsBranch() {
    return "DETACHED", nil
}

branchName := head.Name().Short() // e.g., "main"
```

**Ahead/Behind Counts**:

```go
// Get local HEAD commit
localRef, _ := repo.Head()
localCommit, _ := repo.CommitObject(localRef.Hash())

// Get remote tracking branch
remote, err := repo.Remote("origin")
if err != nil {
    return 0, 0, fmt.Errorf("no remote")
}

refs, _ := remote.List(&git.ListOptions{})
// Find matching remote branch
remoteRef := findRemoteRef(refs, branchName)
remoteCommit, _ := repo.CommitObject(remoteRef.Hash())

// Count commits
ahead := countCommitsBetween(localCommit, remoteCommit)
behind := countCommitsBetween(remoteCommit, localCommit)
```

**Stash Detection**:

```go
stashRef, err := repo.Storer.Reference(plumbing.ReferenceName("refs/stash"))
if err != nil {
    return false // No stashes
}
return true
```

**Uncommitted Changes**:

```go
worktree, err := repo.Worktree()
status, err := worktree.Status()

hasChanges := !status.IsClean()
```

**Best Practices**:

- Implement timeout context for each Git operation
- Handle "no remote" case with ○ symbol
- Gracefully degrade on partial failures (show what's available + error indicator)
- Format status compactly: `[branch ↑ahead ↓behind $ *]`

### 7. Error Handling and Timeouts

**Decision**: Implement graceful degradation with timeout handling and error indicators

**Rationale**:

- Git operations can hang on network issues or corrupted repos
- Tool should remain useful even when some repos fail
- Users need to know which repos had issues

**Error Handling Strategy**:

- **Timeout**: 5-10 second timeout per repository, show partial info + "timeout" indicator
- **Corrupted .git**: Show available info + "error" indicator
- **Permission denied**: Skip silently or show "error" indicator (configurable)
- **No remote**: Show ○ symbol instead of ahead/behind
- **Detached HEAD**: Show "DETACHED" instead of branch name

**Implementation Pattern**:

```go
func getGitStatusWithTimeout(path string) (*GitStatus, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resultChan := make(chan *GitStatus, 1)
    errorChan := make(chan error, 1)

    go func() {
        status, err := extractGitStatus(path)
        if err != nil {
            errorChan <- err
            return
        }
        resultChan <- status
    }()

    select {
    case status := <-resultChan:
        return status, nil
    case err := <-errorChan:
        return nil, err
    case <-ctx.Done():
        return nil, fmt.Errorf("timeout")
    }
}
```

**Error Indicator Formatting**:

- Timeout: `repo-name [partial-info] timeout`
- Error: `repo-name error` or `repo-name [partial-info] error`
- Bare repo: `repo-name [branch bare]`
- No remote: `repo-name [branch ○]`

### 8. Testing Strategy with Testify

**Decision**: Use `github.com/stretchr/testify` for assertions and mocking

**Rationale**:

- Most popular Go testing library (20k+ stars)
- Rich assertion library (`assert`, `require`)
- Mock generation capabilities (`mock` package)
- Suite support for setup/teardown
- Compatible with standard `go test`

**Testing Approach**:

**Unit Tests** (per library):

- `scanner/scanner_test.go`: Test directory traversal, Git detection, symlink handling
- `gitstatus/status_test.go`: Test status extraction for various repo states
- `tree/formatter_test.go`: Test tree formatting with different structures
- `models/repository_test.go`: Test data structure validation

**Integration Tests**:

- Create temporary test repositories with known states
- Test end-to-end workflow: scan → extract status → format → output
- Test error scenarios: corrupted repos, permission issues, timeouts

**Test Utilities**:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestScanRepository(t *testing.T) {
    // Setup test repo
    tempDir := createTestRepo(t)
    defer os.RemoveAll(tempDir)

    // Execute
    repos, err := scanner.Scan(tempDir)

    // Assert
    require.NoError(t, err)
    assert.Len(t, repos, 1)
    assert.Equal(t, tempDir, repos[0].Path)
}
```

**Test Coverage Goals**:

- Unit test coverage: >80%
- All error paths tested
- Edge cases covered (bare repos, detached HEAD, no remote, timeouts)

### 9. Symlink Handling

**Decision**: Follow symlinks once using inode tracking to prevent cycles

**Rationale**:

- Users may have symlinked repositories
- Need to prevent infinite loops from circular symlinks
- Following once is sufficient for most use cases (per spec clarification)

**Implementation**:

```go
type scanner struct {
    visited map[uint64]bool // Track visited inodes
}

func (s *scanner) shouldVisit(path string) (bool, error) {
    info, err := os.Lstat(path)
    if err != nil {
        return false, err
    }

    // Get inode
    stat, ok := info.Sys().(*syscall.Stat_t)
    if !ok {
        return true, nil
    }

    inode := stat.Ino
    if s.visited[inode] {
        return false, nil // Already visited
    }

    s.visited[inode] = true

    // Follow symlink if it's a symlink
    if info.Mode()&os.ModeSymlink != 0 {
        realPath, err := filepath.EvalSymlinks(path)
        if err != nil {
            return false, err
        }
        path = realPath
    }

    return true, nil
}
```

## Implementation Recommendations

### Phase Approach

1. **Phase 1**: Core libraries (scanner, gitstatus, tree formatter)
2. **Phase 2**: CLI integration with spinner
3. **Phase 3**: Concurrent processing
4. **Phase 4**: Error handling and timeouts

### Dependencies to Add

```go
// go.mod
module github.com/user/gitree

go 1.25

require (
    github.com/go-git/go-git/v5 v5.16.3
    github.com/briandowns/spinner v1.23.2
    github.com/stretchr/testify v1.11.1
)
```

### Performance Considerations

- Use worker pool to limit concurrent goroutines (10-20 workers)
- Implement early exit for nested repos (skip directory traversal once Git repo found)
- Cache repository objects when multiple operations needed
- Use buffered channels for result collection
- Consider memoization for repeated path operations if profiling shows benefit

## References

- [go-git documentation](https://pkg.go.dev/github.com/go-git/go-git/v5)
- [briandowns/spinner](https://github.com/briandowns/spinner)
- [testify documentation](https://github.com/stretchr/testify)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Unix tree command reference](https://linux.die.net/man/1/tree)
