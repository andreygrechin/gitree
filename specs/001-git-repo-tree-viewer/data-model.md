# Data Model: Git Repository Tree Viewer

**Feature**: 001-git-repo-tree-viewer
**Date**: 2025-11-13
**Purpose**: Define core entities, their relationships, and validation rules

## Entity Definitions

### 1. Repository

Represents a Git repository discovered during directory scanning.

**Fields**:

- `Path` (string, required): Absolute file system path to the repository directory
- `Name` (string, required): Base name of the repository directory (derived from Path)
- `IsBare` (bool, required): Whether the repository is a bare repository (no working directory)
- `IsSymlink` (bool, required): Whether the repository was reached via a symbolic link
- `GitStatus` (*GitStatus, optional): Current Git status information (nil if error occurred)
- `Error` (error, optional): Error encountered during processing (nil if successful)
- `HasTimeout` (bool, required): Whether Git operations timed out for this repository

**Relationships**:

- Contains one GitStatus (optional, may be nil on error)
- May be a child of a parent TreeNode
- May have sibling repositories in the same directory

**Validation Rules**:

- Path MUST be non-empty and absolute
- Name MUST be non-empty
- If GitStatus is nil, Error or HasTimeout SHOULD be set
- IsBare and GitStatus.HasChanges MUST NOT both be true (bare repos have no working directory)

**Go Implementation**:

```go
type Repository struct {
    Path       string
    Name       string
    IsBare     bool
    IsSymlink  bool
    GitStatus  *GitStatus
    Error      error
    HasTimeout bool
}

func (r *Repository) Validate() error {
    if r.Path == "" {
        return fmt.Errorf("path cannot be empty")
    }
    if !filepath.IsAbs(r.Path) {
        return fmt.Errorf("path must be absolute: %s", r.Path)
    }
    if r.Name == "" {
        return fmt.Errorf("name cannot be empty")
    }
    if r.IsBare && r.GitStatus != nil && r.GitStatus.HasChanges {
        return fmt.Errorf("bare repository cannot have uncommitted changes")
    }
    return nil
}
```

### 2. GitStatus

Represents the Git status information for a repository.

**Fields**:

- `Branch` (string, required): Current branch name or "DETACHED" if HEAD is detached
- `IsDetached` (bool, required): Whether HEAD is in detached state
- `HasRemote` (bool, required): Whether repository has a remote configured
- `Ahead` (int, required): Number of commits ahead of remote (0 if no remote or error)
- `Behind` (int, required): Number of commits behind remote (0 if no remote or error)
- `HasStashes` (bool, required): Whether repository has stashed changes
- `HasChanges` (bool, required): Whether repository has uncommitted changes (modified, added, deleted files)
- `Error` (string, optional): Partial error message if some status info couldn't be retrieved

**Validation Rules**:

- Branch MUST be non-empty
- If IsDetached is true, Branch MUST equal "DETACHED"
- If HasRemote is false, Ahead and Behind MUST be 0
- Ahead and Behind MUST be >= 0

**State Transitions**:

- Normal → Detached: When checking out a specific commit SHA
- Detached → Normal: When checking out a branch
- No changes → Has changes: When files are modified, added, or deleted
- Has changes → No changes: When changes are committed or discarded
- No stashes → Has stashes: When `git stash` is executed
- Has stashes → No stashes: When all stashes are applied or dropped

**Go Implementation**:

```go
type GitStatus struct {
    Branch     string
    IsDetached bool
    HasRemote  bool
    Ahead      int
    Behind     int
    HasStashes bool
    HasChanges bool
    Error      string
}

func (g *GitStatus) Validate() error {
    if g.Branch == "" {
        return fmt.Errorf("branch cannot be empty")
    }
    if g.IsDetached && g.Branch != "DETACHED" {
        return fmt.Errorf("detached HEAD must have branch = 'DETACHED'")
    }
    if !g.HasRemote && (g.Ahead != 0 || g.Behind != 0) {
        return fmt.Errorf("no remote but ahead/behind counts are non-zero")
    }
    if g.Ahead < 0 || g.Behind < 0 {
        return fmt.Errorf("ahead/behind counts cannot be negative")
    }
    return nil
}

func (g *GitStatus) Format() string {
    var parts []string

    // Branch or DETACHED
    parts = append(parts, g.Branch)

    // Ahead/Behind or No Remote indicator
    if g.HasRemote {
        if g.Ahead > 0 {
            parts = append(parts, fmt.Sprintf("↑%d", g.Ahead))
        }
        if g.Behind > 0 {
            parts = append(parts, fmt.Sprintf("↓%d", g.Behind))
        }
    } else {
        parts = append(parts, "○")
    }

    // Stashes
    if g.HasStashes {
        parts = append(parts, "$")
    }

    // Uncommitted changes
    if g.HasChanges {
        parts = append(parts, "*")
    }

    result := "[" + strings.Join(parts, " ") + "]"

    // Append error indicator if present
    if g.Error != "" {
        result += " error"
    }

    return result
}
```

**Format Examples**:

- `[main]` - On main branch, in sync with remote, no changes
- `[main ↑2 ↓1]` - 2 commits ahead, 1 behind
- `[develop $ *]` - Has stashes and uncommitted changes
- `[DETACHED]` - Detached HEAD state
- `[main ○]` - No remote configured
- `[feature bare]` - Bare repository (special case)
- `[main] error` - Partial error retrieving status

### 3. TreeNode

Represents a node in the hierarchical tree structure used for formatting output.

**Fields**:

- `Repository` (*Repository, required): The repository at this tree node
- `Depth` (int, required): Depth level in the tree (0 = root)
- `IsLast` (bool, required): Whether this is the last child of its parent
- `Children` ([]*TreeNode, optional): Child nodes (nested repositories or subdirectories)
- `RelativePath` (string, required): Path relative to scan root, used for display

**Relationships**:

- Contains one Repository
- May have multiple Children (TreeNode instances)
- Has implicit parent relationship (not stored, inferred from tree structure)

**Validation Rules**:

- Repository MUST NOT be nil
- Depth MUST be >= 0
- RelativePath MUST be non-empty
- Children ordering SHOULD be deterministic (alphabetically sorted)

**Tree Properties**:

- Root node has Depth = 0
- Child nodes have Depth = parent.Depth + 1
- Leaf nodes have len(Children) = 0
- Last child of each parent has IsLast = true

**Go Implementation**:

```go
type TreeNode struct {
    Repository   *Repository
    Depth        int
    IsLast       bool
    Children     []*TreeNode
    RelativePath string
}

func (t *TreeNode) Validate() error {
    if t.Repository == nil {
        return fmt.Errorf("repository cannot be nil")
    }
    if t.Depth < 0 {
        return fmt.Errorf("depth cannot be negative")
    }
    if t.RelativePath == "" {
        return fmt.Errorf("relative path cannot be empty")
    }
    return nil
}

func (t *TreeNode) AddChild(child *TreeNode) {
    child.Depth = t.Depth + 1
    t.Children = append(t.Children, child)
}

func (t *TreeNode) SortChildren() {
    sort.Slice(t.Children, func(i, j int) bool {
        return t.Children[i].Repository.Name < t.Children[j].Repository.Name
    })
    // Update IsLast flags
    for i, child := range t.Children {
        child.IsLast = (i == len(t.Children)-1)
    }
}
```

### 4. ScanResult

Represents the complete result of a directory scan operation.

**Fields**:

- `RootPath` (string, required): Absolute path where scan started
- `Repositories` ([]*Repository, required): All repositories found during scan
- `Tree` (*TreeNode, required): Root node of the tree structure
- `TotalScanned` (int, required): Total number of directories scanned
- `TotalRepos` (int, required): Total number of Git repositories found
- `Errors` ([]error, optional): Collection of non-fatal errors encountered during scan
- `Duration` (time.Duration, required): Time taken to complete scan

**Validation Rules**:

- RootPath MUST be non-empty and absolute
- Repositories MUST NOT be nil (can be empty slice)
- Tree MUST NOT be nil
- TotalRepos MUST equal len(Repositories)
- TotalScanned MUST be >= TotalRepos
- Duration MUST be >= 0

**Go Implementation**:

```go
type ScanResult struct {
    RootPath      string
    Repositories  []*Repository
    Tree          *TreeNode
    TotalScanned  int
    TotalRepos    int
    Errors        []error
    Duration      time.Duration
}

func (s *ScanResult) Validate() error {
    if s.RootPath == "" {
        return fmt.Errorf("root path cannot be empty")
    }
    if !filepath.IsAbs(s.RootPath) {
        return fmt.Errorf("root path must be absolute: %s", s.RootPath)
    }
    if s.Repositories == nil {
        return fmt.Errorf("repositories slice cannot be nil")
    }
    if s.Tree == nil {
        return fmt.Errorf("tree cannot be nil")
    }
    if s.TotalRepos != len(s.Repositories) {
        return fmt.Errorf("total repos mismatch: %d != %d", s.TotalRepos, len(s.Repositories))
    }
    if s.TotalScanned < s.TotalRepos {
        return fmt.Errorf("total scanned < total repos: %d < %d", s.TotalScanned, s.TotalRepos)
    }
    if s.Duration < 0 {
        return fmt.Errorf("duration cannot be negative")
    }
    return nil
}

func (s *ScanResult) HasErrors() bool {
    return len(s.Errors) > 0
}

func (s *ScanResult) SuccessRate() float64 {
    if s.TotalRepos == 0 {
        return 1.0
    }

    reposWithErrors := 0
    for _, repo := range s.Repositories {
        if repo.Error != nil || repo.HasTimeout {
            reposWithErrors++
        }
    }

    return float64(s.TotalRepos-reposWithErrors) / float64(s.TotalRepos)
}
```

## Entity Relationships Diagram

```text
┌─────────────────┐
│   ScanResult    │
└────────┬────────┘
         │ contains
         │
    ┌────▼──────────┐
    │   TreeNode    │◄──────┐
    └────┬──────────┘       │
         │ contains         │ children
         │                  │
    ┌────▼──────────┐       │
    │  Repository   │       │
    └────┬──────────┘       │
         │ contains         │
         │                  │
    ┌────▼──────────┐       │
    │   GitStatus   │       │
    └───────────────┘       │
                            │
    ┌───────────────────────┘
    │ Children []*TreeNode
    └───────────────────────►
```

**Cardinality**:

- ScanResult 1 : N Repository (one scan finds many repos)
- ScanResult 1 : 1 TreeNode (one scan produces one tree root)
- TreeNode 1 : 1 Repository (each node represents one repo)
- TreeNode 1 : N TreeNode (each node can have many children)
- Repository 1 : 0..1 GitStatus (one repo has optional status)

## Data Flow

1. **Scan Phase**:
   - Input: Root directory path (string)
   - Process: Directory traversal, Git detection
   - Output: List of Repository instances with paths

2. **Status Extraction Phase**:
   - Input: Repository instances
   - Process: Concurrent Git operations (parallel goroutines)
   - Output: Repository instances populated with GitStatus

3. **Tree Building Phase**:
   - Input: Repository list with status
   - Process: Construct hierarchical TreeNode structure
   - Output: TreeNode root with complete tree

4. **Formatting Phase**:
   - Input: TreeNode tree
   - Process: Recursive tree traversal with formatting
   - Output: Formatted string (ASCII tree with Git info)

5. **Output Phase**:
   - Input: Formatted string
   - Process: Write to stdout
   - Output: Terminal display

## Edge Case Handling

### Repository Edge Cases

1. **Bare Repository**:
   - `IsBare = true`
   - `GitStatus.HasChanges = false` (no working directory)
   - Format shows "bare" indicator: `[main bare]`

2. **Corrupted Repository**:
   - `Error != nil`
   - `GitStatus = nil` or partial
   - Format shows "error" indicator: `repo-name error`

3. **Timeout Repository**:
   - `HasTimeout = true`
   - `GitStatus` may be partial
   - Format shows "timeout" indicator: `repo-name [partial] timeout`

4. **No Remote Repository**:
   - `GitStatus.HasRemote = false`
   - `GitStatus.Ahead = 0`, `GitStatus.Behind = 0`
   - Format shows no-remote indicator: `[main ○]`

5. **Symlinked Repository**:
   - `IsSymlink = true`
   - `Path` is the resolved path (after following symlink)
   - Tracked by inode to prevent duplicates

### GitStatus Edge Cases

1. **Detached HEAD**:
   - `IsDetached = true`
   - `Branch = "DETACHED"`
   - Format: `[DETACHED]`

2. **Conflicted State** (merge/rebase in progress):
   - `HasChanges = true` (conflicts are changes)
   - No special indicator (shown as uncommitted changes)

3. **Empty Repository** (no commits):
   - `Branch = "main"` (or default branch)
   - `HasRemote = false` (typically)
   - `Ahead = 0`, `Behind = 0`

### TreeNode Edge Cases

1. **Single Repository** (scan root is a Git repo):
   - Tree has single root node with no children
   - Output: `.` + GitStatus

2. **Nested Repositories**:
   - Parent repo is one TreeNode
   - Child repo is separate TreeNode (not a child of parent)
   - Both appear at appropriate depth levels

3. **Empty Directory Tree** (no Git repos):
   - `ScanResult.TotalRepos = 0`
   - `ScanResult.Repositories = []` (empty slice)
   - Output: "No Git repositories found"

## Invariants

1. **Tree Consistency**: Every Repository in ScanResult.Repositories MUST appear in exactly one TreeNode in the tree
2. **Depth Consistency**: Child TreeNode depth MUST equal parent depth + 1
3. **Path Consistency**: TreeNode.RelativePath MUST be derivable from Repository.Path relative to ScanResult.RootPath
4. **Error Consistency**: If Repository.Error != nil OR Repository.HasTimeout = true, then GitStatus may be nil or partial
5. **Status Consistency**: If GitStatus != nil, it MUST pass validation rules
6. **Sort Consistency**: TreeNode.Children MUST be sorted deterministically (alphabetically by Repository.Name)
