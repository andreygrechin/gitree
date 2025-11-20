# Feature Specification: CLI Framework with Command-Line Flags

**Feature Branch**: `003-cobra-cli-flags`
**Created**: 2025-11-17
**Status**: Draft
**Input**: User description: "I want to add support for commands and flags using spf13/cobra. The app should support these flags: version, no-color, and all. The version flag should show semver, build time, and commit. These variables should be set up at building time using Makefile and its variables. The no-color flag should suppress all color codes. By default, the app should show only repositories with changes, keeping the tree structure. The 'all' flag should reproduce the previous behavior, which shows all git directories in a tree, with and without changes. For detecting changes, you may reuse the same logic used for coloring square double brackets."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Application Version Information (Priority: P2)

As a user running gitree, I want to see the application version, build information, and commit hash so that I can verify which version I'm using and report it when seeking support or filing bug reports.

**Why this priority**: Version information is fundamental for troubleshooting, support, and ensuring users are running the correct version. This is a critical operational requirement that should be available in every CLI tool.

**Independent Test**: Can be fully tested by running the version flag and verifying that semantic version, build timestamp, and commit hash are displayed. Delivers immediate value by providing version transparency.

**Acceptance Scenarios**:

1. **Given** the application is installed, **When** user runs the tool with the version flag, **Then** the output displays semantic version (e.g., "v1.2.3"), build timestamp in ISO format, and Git commit hash
2. **Given** the application is built without version information injected, **When** user runs the version flag, **Then** the output displays placeholder values (e.g., "dev", "unknown", "none") indicating development build
3. **Given** the version flag is invoked, **When** the output is displayed, **Then** the application exits immediately without performing any directory scanning

---

### User Story 2 - Suppress Color Output (Priority: P3)

As a user redirecting gitree output to a file, log system, or non-color-supporting terminal, I want to suppress all color formatting so that the output is clean, readable plain text.

**Why this priority**: Color suppression is essential for automation, logging, and compatibility with various terminal environments. While not critical for basic functionality, it significantly improves usability in non-interactive scenarios.

**Independent Test**: Can be fully tested by running the tool with the no-color flag and verifying that output contains only plain text without color formatting. Delivers value for users in automation and logging scenarios.

**Acceptance Scenarios**:

1. **Given** the application is run with the no-color flag, **When** repository trees are displayed, **Then** all output contains only plain text without any color formatting
2. **Given** the no-color flag is active, **When** git status information is rendered (branch names, ahead/behind indicators, uncommitted changes), **Then** all text is displayed in plain format without color formatting
3. **Given** the application is run in a non-interactive environment (e.g., piped output), **When** the no-color flag is set, **Then** the output is suitable for text processing tools and contains only printable ASCII characters plus standard tree-drawing characters
4. **Given** color output is suppressed, **When** errors or warnings are displayed, **Then** they remain distinguishable through text formatting (symbols, prefixes) rather than color alone

---

### User Story 3 - Show All Repositories Including Clean Ones (Priority: P1)

As a developer managing multiple Git repositories, I want to optionally view all repositories including those in clean state so that I can see the complete repository landscape when needed, even though the default behavior shows only repositories that need attention.

**Why this priority**: While the default filtered view (showing only changed repositories) is the primary workflow for identifying actionable items, users sometimes need to see all repositories for comprehensive inventory, verification, or documentation purposes. This flag provides that capability without changing the default behavior.

**Independent Test**: Can be fully tested by creating a directory structure with both clean and modified repositories, running the tool with the all flag, and verifying that all repositories appear in the output regardless of their state. Delivers value by providing comprehensive repository visibility when needed.

**Acceptance Scenarios**:

1. **Given** the tool is run without the all flag (default mode), **When** a directory contains both clean and changed repositories, **Then** the output displays only repositories with uncommitted changes, non-main branches, stashes, ahead/behind status, or no remote tracking, maintaining the tree structure hierarchy
2. **Given** the tool is run with the all flag, **When** a directory contains both clean and changed repositories, **Then** the output displays all repositories regardless of their state
3. **Given** repositories on main/master branches with clean working trees and synchronized remotes, **When** the tool runs in default mode (without all flag), **Then** these repositories are excluded from the output
4. **Given** repositories on main/master branches with clean working trees and synchronized remotes, **When** the all flag is active, **Then** these repositories are included in the output
5. **Given** repositories on feature branches (not main/master), **When** the tool runs in default mode, **Then** these repositories are included in the output regardless of working tree status
6. **Given** repositories with uncommitted changes, stashes, or ahead/behind status, **When** the all flag is active, **Then** these repositories are displayed with their full status information
7. **Given** all repositories in a directory tree are clean (main/master branch, no changes, no stashes, synchronized with remote), **When** the tool runs in default mode (without all flag), **Then** the output shows the directory structure but indicates that no repositories need attention
8. **Given** nested repository structures where parent directories contain only clean repositories, **When** the tool runs in default mode, **Then** these parent directory nodes are collapsed or hidden from the tree display while preserving the structure for directories containing filtered repositories

---

### Edge Cases

- What happens when the version flag is combined with other flags (all, no-color)? The tool should display version information and exit without processing other flags
- How does the tool handle repositories that are in a detached HEAD state when running in default mode (without all flag)? These should be included as they indicate an abnormal state requiring attention
- What happens when default mode (without all flag) is applied to a directory with no Git repositories? The tool should display a message indicating no repositories were found or all repositories are clean
- How does the no-color flag interact with terminals that don't support color? The flag should have no negative impact; output should remain clean regardless
- What happens when version information cannot be determined (development builds)? The tool should display placeholder values clearly indicating the build environment (e.g., "dev", "unknown")
- How does the default filtered mode handle repositories where Git status extraction times out or fails? These repositories should be included (fail-safe approach) since we cannot determine their clean status
- What happens when a repository is both on main branch AND has uncommitted changes? It should be included in default filtered output (any changed condition triggers inclusion)
- What happens when the all flag is used in a directory with only clean repositories? All repositories should be displayed with their status information

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a version flag that displays semantic version number, build timestamp, and Git commit hash
- **FR-002**: System MUST support injection of version metadata (semver, build time, commit hash) at build time through build system variables
- **FR-003**: System MUST exit immediately after displaying version information without performing any repository scanning operations
- **FR-004**: System MUST provide a no-color flag that suppresses all color formatting in output
- **FR-005**: System MUST apply color suppression globally across all output types (repository names, paths, git status indicators, tree characters, errors, warnings)
- **FR-006**: System MUST display only repositories that need attention by default (without requiring any flag), excluding repositories in clean state from output
- **FR-007**: System MUST provide an all flag that disables the default filtering and shows all repositories regardless of their state
- **FR-008**: System MUST define a repository as "clean" if ALL of the following conditions are met: on main or master branch, working tree has no uncommitted changes (staged or unstaged), no stashes present, synchronized with remote tracking branch (not ahead or behind), and has a configured remote tracking branch
- **FR-009**: System MUST include repositories in default filtered output if ANY of the following conditions are true: on a branch other than main/master, has uncommitted changes, has stashed changes, is ahead of remote, is behind remote, has no remote tracking configured, is in detached HEAD state, or git status extraction failed/timed out
- **FR-010**: System MUST preserve the hierarchical tree structure and visual representation in default mode, showing parent directories that lead to repositories needing attention
- **FR-011**: System MUST collapse or omit directory branches that contain only clean repositories in default mode (without all flag)
- **FR-012**: System MUST allow flags to be combined (e.g., all with no-color), except version flag which causes immediate exit
- **FR-013**: System MUST display placeholder values (e.g., "dev", "unknown", "none") for version information when built without injected metadata
- **FR-014**: System MUST provide clear help text for all flags explaining their purpose and usage
- **FR-015**: System MUST handle invalid flag combinations gracefully with helpful error messages
- **FR-016**: System MUST maintain consistent flag naming conventions following CLI best practices (lowercase, hyphen-separated for multi-word flags)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can retrieve complete version information (semver, build time, commit) in under 100 milliseconds
- **SC-002**: Output with no-color flag contains only plain text without any color formatting codes, verified by text-only validation
- **SC-003**: The default filtered mode correctly identifies and displays only repositories meeting the "needs attention" criteria with 100% accuracy in test scenarios
- **SC-004**: Users can combine all and no-color flags to produce clean, comprehensive output suitable for text processing in automation scripts
- **SC-005**: Version flag causes application to exit immediately without incurring repository scanning overhead (exit time under 100ms regardless of directory size)
- **SC-006**: The default filtered mode reduces visual clutter by at least 50% in typical development environments where majority of repositories are in clean state
- **SC-007**: Users can successfully parse and process gitree output in CI/CD pipelines and automation scripts when no-color flag is used
- **SC-008**: Help documentation for all flags is clear enough that 90% of users can correctly use each flag without referring to external documentation
- **SC-009**: The all flag successfully displays all repositories including clean ones, providing comprehensive visibility when needed

## Assumptions

- The build system has access to Git metadata (tags, commit hash) at build time for version information injection
- The existing color output mechanism can be conditionally disabled
- The existing git status extraction mechanism provides sufficient information to determine repository clean state (branch name, ahead/behind counts, uncommitted changes, stashes, remote tracking status)
- Users expect standard CLI flag conventions (e.g., `--version`, `--no-color`, `--all`)
- The default behavior of showing only repositories that need attention is more useful than showing all repositories for typical development workflows
- The all flag name is clear and follows common CLI conventions for showing complete/unfiltered output
- "Main" and "master" are the standard primary branch names; other naming conventions (e.g., trunk, mainline) are edge cases that can be addressed later
- The tree structure representation should remain visually consistent when filtering is applied (no broken or confusing tree branches)
- Users who prefer seeing all repositories can set the all flag as a default in their shell alias or configuration

## Dependencies

- Existing git status extraction functionality in `internal/gitstatus/` package must provide branch name, ahead/behind counts, uncommitted change indicators, stash count, and remote tracking information
- Existing tree formatting functionality in `internal/tree/` package must support conditional node inclusion/exclusion while maintaining valid tree structure
- Build system (Makefile) must support variable injection at compile time
- Existing color output mechanism must be refactorable to support conditional disabling
