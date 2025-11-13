# Feature Specification: Git Repository Tree Viewer

**Feature Branch**: `001-git-repo-tree-viewer`
**Created**: 2025-11-13
**Status**: Draft
**Input**: User description: "gitree is a command-line tool written in Go. Its purpose is to display an ASCII tree of directories that contain Git repositories. The experience should be identical to using the Linux tree command, but instead of showing all folders, it only lists directories that are Git repositories. Next to each repository name, gitree should show inline Git information. For each Git repository, gitree must display the current branch name or show 'DETACHED' if the HEAD is detached. It should also display how many commits the local branch is ahead or behind the remote origin. If there are any stashes, a symbol such as $ should appear. If there are uncommitted changes, a symbol such as * should be shown. These pieces of information should appear inline next to the folder name, in a compact and readable format similar to tree output. For example, something like 'repo-name [main ↑2 ↓1 $ *]' but without breaking the visual structure of the tree command. The program should recursively scan directories starting from the current working directory. It needs to detect whether each folder contains a .git directory. Only those directories are included in the output. Nested Git repositories, such as submodules, should appear as their own entries, but the program does not need to display their internal folder structure. Because collecting Git information can take time when scanning many repositories, gitree must perform these operations asynchronously. Each repository should be processed in parallel, using goroutines or similar mechanisms. While gitree is working, a spinner or minimal progress indicator should show that the process is ongoing. Once all results are collected, the complete tree with Git status information should be printed to standard output. The first version of gitree must not include any command-line arguments, flags, or configuration files. Running the program in a directory is enough to trigger the scan and output the results. Non-Git directories should be ignored. The tool should fail gracefully when encountering inaccessible or broken repositories, continuing to display others without crashing. The initial implementation should be simple, with a focus on functionality and performance. The output should be plain ASCII text in the same style as the tree command, suitable for terminal use. Future improvements might include colorized output, filtering options, configurable symbols, or support for exporting data in formats such as JSON, but those are not part of the initial version."

## Clarifications

### Session 2025-11-13

- Q: How does the system handle symbolic links that point to Git repositories? → A: Follow symlinks but only once (no recursive symlink following) to detect repositories
- Q: What happens when a repository has no remote configured (ahead/behind indicators)? → A: Show ○ instead of ahead/behind indicators
- Q: What happens when Git operations timeout or hang? → A: Apply a reasonable timeout (e.g., 5-10 seconds per repository) and show partial info on timeout. Also add status "timeout" at the end of lines for the affected repository
- Q: How are bare repositories displayed (repositories without working directories)? → A: Display bare repositories with branch info and a "bare" indicator (e.g., "[main bare]")
- Q: What happens when a repository's .git directory is corrupted or incomplete? → A: Display repository with an "error" indicator showing what info is available. Use "error" indicator for any non-recoverable errors when full information cannot be retrieved

## User Scenarios & Testing

### User Story 1 - View All Git Repositories in Directory Tree (Priority: P1)

A developer working with multiple projects in a directory hierarchy wants to quickly visualize which directories contain Git repositories and see their current status at a glance without navigating into each one.

**Why this priority**: This is the core value proposition of the tool - providing an overview of all Git repositories in a tree structure. Without this, the tool has no purpose.

**Independent Test**: Can be fully tested by running the tool in a directory containing Git repositories and verifying that all repositories are displayed in a tree format with their paths.

**Acceptance Scenarios**:

1. **Given** a directory with multiple nested Git repositories, **When** the user runs gitree, **Then** all repositories are displayed in a hierarchical tree structure similar to the tree command
2. **Given** a directory with both Git repositories and non-Git directories, **When** the user runs gitree, **Then** only Git repositories are shown in the output
3. **Given** a deeply nested directory structure with repositories at various levels, **When** the user runs gitree, **Then** the tree correctly shows parent-child relationships between directories containing repositories

---

### User Story 2 - View Git Status Information Inline (Priority: P1)

A developer wants to see critical Git status information (branch name, ahead/behind commits, uncommitted changes, stashes) for each repository without running git status in each directory individually.

**Why this priority**: This is the primary differentiator from the standard tree command and provides immediate value for repository management. Without status information, the tool is just a filtered tree view.

**Independent Test**: Can be fully tested by running the tool in directories with repositories in various Git states and verifying that branch names, ahead/behind counts, stash indicators, and uncommitted change indicators appear correctly.

**Acceptance Scenarios**:

1. **Given** a repository on a tracked branch, **When** the user runs gitree, **Then** the current branch name appears next to the repository name
2. **Given** a repository with detached HEAD, **When** the user runs gitree, **Then** "DETACHED" appears instead of a branch name
3. **Given** a repository that is 2 commits ahead and 1 commit behind its remote, **When** the user runs gitree, **Then** indicators showing ↑2 and ↓1 appear next to the repository name
4. **Given** a repository with no remote configured, **When** the user runs gitree, **Then** the ○ symbol appears instead of ahead/behind indicators (e.g., "[main ○]")
5. **Given** a bare repository (without working directory), **When** the user runs gitree, **Then** the repository is displayed with branch information and a "bare" indicator (e.g., "[main bare]")
6. **Given** a repository with uncommitted changes, **When** the user runs gitree, **Then** an asterisk (*) or similar symbol appears next to the repository name
7. **Given** a repository with stashed changes, **When** the user runs gitree, **Then** a dollar sign ($) or similar symbol appears next to the repository name
8. **Given** a repository with multiple status indicators, **When** the user runs gitree, **Then** all relevant indicators appear in a compact, readable format (e.g., "[main ↑2 ↓1 $ *]")

---

### User Story 3 - Fast Processing with Progress Feedback (Priority: P2)

A developer scanning a directory with many repositories wants to see progress feedback while the tool collects Git information, so they know the tool is working and not frozen.

**Why this priority**: This is a usability enhancement that improves user experience, especially for large directory structures. The tool is still functional without it, but user experience suffers.

**Independent Test**: Can be fully tested by running the tool in a directory with multiple repositories and observing that a progress indicator (spinner) appears during processing and disappears when complete.

**Acceptance Scenarios**:

1. **Given** a directory with multiple repositories, **When** the user runs gitree, **Then** a progress indicator appears while Git information is being collected
2. **Given** the tool is collecting Git information, **When** all repositories have been processed, **Then** the progress indicator disappears and the full tree is displayed
3. **Given** a directory with many repositories, **When** the user runs gitree, **Then** multiple repositories are processed concurrently to minimize wait time

---

### User Story 4 - Graceful Error Handling (Priority: P2)

A developer scanning directories that may contain corrupted repositories or permission-restricted folders wants the tool to continue showing results for accessible repositories rather than crashing completely.

**Why this priority**: This ensures robustness in real-world scenarios where not all directories may be accessible or valid. The tool remains useful even when some repositories cannot be read.

**Independent Test**: Can be fully tested by running the tool in directories containing inaccessible or corrupted repositories and verifying that accessible repositories are still displayed and the tool doesn't crash.

**Acceptance Scenarios**:

1. **Given** a directory containing both accessible and inaccessible Git repositories, **When** the user runs gitree, **Then** accessible repositories are displayed normally while inaccessible ones are either skipped or marked with an error indicator
2. **Given** a corrupted or incomplete Git repository, **When** the user runs gitree, **Then** the tool continues processing other repositories without crashing and displays the corrupted repository with partial information and an "error" indicator
3. **Given** a repository where Git operations fail with a non-recoverable error, **When** the user runs gitree, **Then** the repository is shown with available partial information and an "error" indicator
4. **Given** a repository where Git operations exceed the timeout threshold, **When** the user runs gitree, **Then** the repository is displayed with partial information and a "timeout" indicator at the end of the line

---

### User Story 5 - Handle Nested Repositories (Priority: P3)

A developer working with projects that contain Git submodules or nested independent repositories wants to see each repository as a separate entry in the tree.

**Why this priority**: This is an edge case that improves completeness but is not critical for the core functionality. Many users may not have nested repositories.

**Independent Test**: Can be fully tested by creating a directory structure with nested Git repositories (parent repo containing child repos) and verifying that each appears as its own tree node.

**Acceptance Scenarios**:

1. **Given** a Git repository containing another Git repository as a subdirectory, **When** the user runs gitree, **Then** both repositories appear as separate entries in the tree
2. **Given** a nested repository, **When** the user runs gitree, **Then** the tool does not display the internal folder structure of any repository, only the repository directories themselves

---

### Edge Cases

- When a repository has no remote configured, the ○ symbol is displayed instead of ahead/behind indicators (e.g., "[main ○]")
- Symbolic links pointing to directories are followed once (non-recursively) when detecting Git repositories, preventing infinite loops while allowing symlinked repositories to be discovered
- When a repository's .git directory is corrupted, incomplete, or any non-recoverable error prevents retrieving full information, the repository is displayed with available partial information and an "error" indicator (e.g., "repo-name error")
- How does the tool behave in a directory with thousands of nested repositories?
- When Git operations exceed a reasonable timeout (5-10 seconds per repository), partial information is displayed with a "timeout" indicator at the end of the line (e.g., "repo-name [main] timeout")
- Bare repositories (without working directories) are displayed with branch information and a "bare" indicator (e.g., "[main bare]")
- What happens when the user runs the tool from a directory where they lack read permissions for subdirectories?
- How does the tool handle repositories during active Git operations (merge, rebase in progress)?

## Requirements

### Functional Requirements

- **FR-001**: System MUST recursively scan directories starting from the current working directory
- **FR-002**: System MUST detect Git repositories by checking for the presence of a .git directory or bare repository structure
- **FR-022**: System MUST follow symbolic links once (non-recursively) when scanning directories to detect repositories, preventing infinite loops and duplicate traversal
- **FR-026**: System MUST display bare repositories with branch information and a "bare" indicator
- **FR-003**: System MUST display only directories that contain Git repositories, excluding non-Git directories
- **FR-004**: System MUST display repositories in a hierarchical tree structure using ASCII characters similar to the tree command
- **FR-005**: System MUST display the current branch name for each repository next to the repository name
- **FR-006**: System MUST display "DETACHED" when a repository has a detached HEAD instead of a branch name
- **FR-007**: System MUST display the number of commits the local branch is ahead of the remote origin using an up arrow (↑) or similar indicator
- **FR-008**: System MUST display the number of commits the local branch is behind the remote origin using a down arrow (↓) or similar indicator
- **FR-023**: System MUST display a "no remote" indicator (○) when a repository has no remote configured, replacing the ahead/behind indicators
- **FR-009**: System MUST display a stash indicator ($ or similar) when a repository has stashed changes
- **FR-010**: System MUST display an uncommitted changes indicator (* or similar) when a repository has uncommitted changes (modified, added, or deleted files)
- **FR-011**: System MUST display all status information inline next to the repository name in a compact format
- **FR-012**: System MUST process repositories asynchronously in parallel to improve performance
- **FR-013**: System MUST display a progress indicator (spinner or similar) while collecting Git information
- **FR-014**: System MUST hide the progress indicator and display the complete tree once all information is collected
- **FR-015**: System MUST continue processing and displaying other repositories when encountering an inaccessible or broken repository
- **FR-016**: System MUST NOT crash when encountering errors in individual repositories
- **FR-024**: System MUST apply a timeout of 5-10 seconds for Git operations on each repository
- **FR-025**: System MUST display partial repository information with a "timeout" indicator when Git operations exceed the timeout threshold
- **FR-027**: System MUST display repositories with corrupted or incomplete .git directories using available partial information and an "error" indicator
- **FR-028**: System MUST use the "error" indicator for any non-recoverable errors that prevent retrieving full repository information
- **FR-017**: System MUST treat nested Git repositories as separate entries in the tree
- **FR-018**: System MUST NOT display the internal folder structure of repositories, only the repository directory itself
- **FR-019**: System MUST execute without requiring command-line arguments, flags, or configuration files
- **FR-020**: System MUST output results as plain ASCII text to standard output
- **FR-021**: System MUST be executable by simply running the program in any directory

### Key Entities

- **Repository**: A directory containing a .git subdirectory, representing a Git repository with status information including branch name, ahead/behind counts, stash presence, and uncommitted changes flag
- **Tree Node**: A representation of a repository in the hierarchical tree structure, containing the directory path, depth level, and relationship to parent directories
- **Git Status**: The current state of a repository including branch name (or detached HEAD status), commit offset from remote (ahead/behind counts), presence of stashes, and presence of uncommitted changes

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can visualize all Git repositories in a directory tree without navigating into each directory
- **SC-002**: Users can identify the current branch and status of each repository within 3 seconds of viewing the output
- **SC-003**: The tool scans and displays results for at least 50 repositories in under 10 seconds on standard hardware
- **SC-004**: The tool successfully handles and continues processing when encountering at least one inaccessible or corrupted repository without crashing
- **SC-005**: Users see a progress indicator within 1 second of starting the tool when scanning directories with multiple repositories
- **SC-006**: The tree output format closely matches the visual structure of the standard tree command, maintaining readability and familiar structure
- **SC-007**: 100% of Git repositories in the scanned directory tree are detected and displayed
- **SC-008**: All status indicators (branch, ahead/behind, stashes, uncommitted changes) are displayed accurately for repositories where Git information is accessible

### Assumptions

- Users have Git installed and accessible in their system PATH
- Users run the tool with sufficient permissions to read directories and execute Git commands
- The tool will be used primarily in development environments with a moderate number of repositories (10-100)
- Standard ASCII output is sufficient for the initial version; colorization is not required
- Users understand basic Git concepts (branches, commits, stashes, working directory changes)
- The term "uncommitted changes" includes any modifications in the working directory or staging area
- Progress indicators can be simple text-based spinners or dots suitable for terminal output
- Remote tracking is configured for repositories where ahead/behind indicators are expected
- Repository scanning will primarily be limited to file system boundaries (no network-mounted repositories in the initial version)
