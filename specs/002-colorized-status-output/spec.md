# Feature Specification: Colorized Status Output

**Feature Branch**: `002-colorized-status-output`
**Created**: 2025-11-16
**Status**: Draft
**Input**: User description: "I want to change the visual representation of the app output. Currently, all metadata (like branch name, number of commits behind/ahead, etc.) is shown in the same color and is barely distinguishable from the names of repositories/directories. I want to add colors.

First, use double square brackets [[ main | ↑2 ↓1 $ * ]] to better isolate the text. Use gray text for the brackets. If the branch name is main or master, use gray text; otherwise use yellow. For the number of commits, use green for commits ahead and red for commits behind. For stashes, use red. For 'bare', use red. For uncommitted changes, use red."

## User Scenarios & Testing

### User Story 1 - Distinguish Repository Metadata from Names (Priority: P1)

A developer viewing the gitree output wants to quickly distinguish between repository names and their Git status metadata without straining to read the plain text output, making it easier to scan the tree structure.

**Why this priority**: This is the core value proposition of the feature - improving visual hierarchy and readability of the output. Without this, the metadata blends with repository names and reduces scanning efficiency.

**Independent Test**: Can be fully tested by running gitree and verifying that status metadata appears in double square brackets [[ ]] with gray-colored bracket characters, creating visual separation from repository names.

**Acceptance Scenarios**:

1. **Given** a repository with Git status metadata, **When** the user runs gitree, **Then** the metadata appears enclosed in double square brackets [[ ]] instead of single brackets [ ]
2. **Given** a repository with Git status metadata, **When** the user runs gitree, **Then** the double square bracket characters [[ and ]] are displayed in gray color
3. **Given** multiple repositories in the tree, **When** the user views the output, **Then** the colored metadata is visually distinct from the uncolored repository/directory names

---

### User Story 2 - Identify Branch Type at a Glance (Priority: P1)

A developer scanning multiple repositories wants to quickly identify which repositories are on the main development branches (main/master) versus feature branches, without reading each branch name carefully.

**Why this priority**: This provides immediate context about repository state and workflow position. Main branches typically represent stable code, while other branches indicate work in progress. This visual differentiation is critical for repository management.

**Independent Test**: Can be fully tested by running gitree in directories with repositories on different branches and verifying that main/master appear in gray and other branches in yellow.

**Acceptance Scenarios**:

1. **Given** a repository on the "main" branch, **When** the user runs gitree, **Then** the branch name "main" is displayed in gray color
2. **Given** a repository on the "master" branch, **When** the user runs gitree, **Then** the branch name "master" is displayed in gray color
3. **Given** a repository on any other branch (e.g., "feature/login", "develop"), **When** the user runs gitree, **Then** the branch name is displayed in yellow color

---

### User Story 3 - Assess Synchronization Status with Visual Cues (Priority: P1)

A developer wants to immediately identify repositories that are ahead of or behind their remote counterparts using color-coded indicators, allowing quick assessment of which repositories need pushing or pulling.

**Why this priority**: This is critical for understanding synchronization state across multiple repositories. Green (ahead) indicates work ready to push, while red (behind) indicates updates to pull. This affects workflow decisions.

**Independent Test**: Can be fully tested by creating repositories with various ahead/behind states and verifying that ahead counts appear in green and behind counts in red.

**Acceptance Scenarios**:

1. **Given** a repository that is 2 commits ahead of its remote, **When** the user runs gitree, **Then** the "↑2" indicator is displayed in green color
2. **Given** a repository that is 1 commit behind its remote, **When** the user runs gitree, **Then** the "↓1" indicator is displayed in red color
3. **Given** a repository that is both ahead and behind, **When** the user runs gitree, **Then** the ahead indicator (↑) appears in green and the behind indicator (↓) appears in red
4. **Given** a repository with no remote configured showing the ○ symbol, **When** the user runs gitree, **Then** the ○ symbol is displayed in yellow color

---

### User Story 4 - Identify Repositories Requiring Attention (Priority: P2)

A developer wants to quickly spot repositories with uncommitted changes, stashes, or bare repository indicators using red color coding, making it easy to identify repositories that need attention or cleanup.

**Why this priority**: This enhances workflow efficiency by highlighting repositories in states that typically require action. While not as critical as branch and sync status, it helps with repository hygiene and task management.

**Independent Test**: Can be fully tested by creating repositories with stashes, uncommitted changes, and bare repositories, then verifying red color is applied to the appropriate indicators.

**Acceptance Scenarios**:

1. **Given** a repository with uncommitted changes, **When** the user runs gitree, **Then** the uncommitted changes indicator (*) is displayed in red color
2. **Given** a repository with stashed changes, **When** the user runs gitree, **Then** the stash indicator ($) is displayed in red color
3. **Given** a bare repository, **When** the user runs gitree, **Then** the "bare" indicator is displayed in red color
4. **Given** a repository with multiple attention indicators (stashes and uncommitted changes), **When** the user runs gitree, **Then** all relevant indicators ($, *) are displayed in red color

---

### Edge Cases

- What happens when the terminal does not support colors (color codes should gracefully degrade or be automatically detected)?
- How are colors displayed when output is redirected to a file or piped to another command (should color codes be disabled)?
- What happens when running gitree in terminals with light backgrounds versus dark backgrounds (gray color should remain readable)?
- How are colors handled for repositories with error or timeout indicators from the base feature?
- When a repository shows "DETACHED" instead of a branch name, what color should it be (since it's neither main/master nor a feature branch)?

## Requirements

### Functional Requirements

- **FR-001**: System MUST display Git status metadata enclosed in double square brackets [[ ]] instead of single brackets [ ]
- **FR-002**: System MUST display the double square bracket characters ([[ and ]]) in gray color
- **FR-003**: System MUST display the separator character (|) in gray color between branch name and status indicators only when status indicators are present
- **FR-004**: System MUST display branch names "main" and "master" in gray color
- **FR-005**: System MUST display all other branch names (not "main" or "master") in yellow color
- **FR-006**: System MUST display the ahead indicator (↑) and its count in green color
- **FR-007**: System MUST display the behind indicator (↓) and its count in red color
- **FR-008**: System MUST display the stash indicator ($) in red color
- **FR-009**: System MUST display the uncommitted changes indicator (*) in red color
- **FR-010**: System MUST display the "bare" indicator in red color
- **FR-011**: System MUST detect whether the terminal supports color output
- **FR-012**: System MUST disable color output when output is redirected to a file or pipe (not a TTY)
- **FR-013**: System MUST support standard ANSI color codes for terminal output
- **FR-014**: System MUST display "DETACHED" status in yellow color, treating it similarly to non-main branch names
- **FR-015**: System MUST preserve the existing format of status metadata from FR-011 of spec 001, only adding color and changing bracket style
- **FR-016**: System MUST apply colors consistently across all status elements within a single repository entry
- **FR-017**: System MUST maintain the same information density and layout as the uncolored version (colors only, no layout changes beyond bracket style)
- **FR-018**: System MUST display the no-remote indicator (○) in yellow color

### Key Entities

- **Color Scheme**: A configuration of ANSI color codes mapping Git status components (branch type, sync status, attention indicators) to specific colors (gray, yellow, green, red)
- **Terminal Capabilities**: Detection of color support and TTY status to determine whether to apply color codes to output

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can distinguish repository names from Git status metadata 50% faster when viewing colored output compared to plain text (measured by visual scanning tasks)
- **SC-002**: 90% of users can identify whether a repository is on main/master vs. a feature branch within 1 second of viewing a colored entry
- **SC-003**: Users can identify repositories requiring attention (uncommitted changes, stashes) 40% faster with red color indicators
- **SC-004**: Color output is automatically disabled when output is redirected to files or pipes, preventing ANSI codes in non-terminal contexts
- **SC-005**: Colored output remains readable on both dark and light terminal backgrounds (gray and colors maintain sufficient contrast)
- **SC-006**: The tool correctly detects and respects terminal color support capabilities on major platforms (Linux, macOS, Windows)
- **SC-007**: All status information from the base feature (001-git-repo-tree-viewer) is preserved with identical accuracy when colors are added

### Assumptions

- Users run gitree in terminals that support ANSI color codes (standard on modern Linux, macOS, Windows Terminal, and most terminal emulators)
- Color preferences are aligned with common developer conventions (green = good/ahead, red = attention/behind, yellow = warning/non-standard)
- Gray color has sufficient contrast on both light and dark terminal backgrounds (may need testing across terminal themes)
- Users with color vision deficiencies can still use the tool effectively because information is conveyed through symbols in addition to colors
- The double bracket format [[ ]] provides sufficient visual separation without requiring additional spacing
- Terminal color support detection can be performed using standard methods (TTY detection, environment variable checks)
- When color is disabled, the output should still use double brackets [[ ]] for consistency (only color codes are removed, not formatting)
- The "bare" indicator should be red to align with the attention/warning color scheme, even though it's informational rather than actionable
