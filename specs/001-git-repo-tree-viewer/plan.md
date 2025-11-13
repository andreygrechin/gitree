# Implementation Plan: Git Repository Tree Viewer

**Branch**: `001-git-repo-tree-viewer` | **Date**: 2025-11-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-git-repo-tree-viewer/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Gitree is a command-line tool that displays an ASCII tree of directories containing Git repositories. It recursively scans directories, detects Git repositories, and displays them in a tree structure with inline Git status information (branch name, ahead/behind commits, stashes, uncommitted changes). The tool processes repositories asynchronously with progress feedback and handles errors gracefully. Built using Go with go-git library for Git operations, briandowns/spinner for progress indication, and stretchr/testify for testing.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: github.com/go-git/go-git/v5 (Git operations), github.com/briandowns/spinner (progress indicator), github.com/stretchr/testify (testing framework)
**Storage**: N/A (reads from file system and Git metadata)
**Testing**: go test with github.com/stretchr/testify
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single CLI application
**Performance Goals**: Process 50+ repositories in under 10 seconds, show progress indicator within 1 second
**Constraints**: 5-10 second timeout per repository for Git operations, graceful degradation on errors, no command-line flags in initial version
**Scale/Scope**: Handle up to hundreds of repositories in nested directory structures, concurrent processing with goroutines

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Library-First ✅

**Status**: PASS
**Assessment**: The feature naturally decomposes into standalone libraries:

- Repository scanner library (filesystem traversal, Git detection)
- Git status library (branch, ahead/behind, stashes, changes detection)
- Tree formatter library (ASCII tree generation)
- CLI wrapper that composes these libraries

Each library has clear, independent utility and can be tested in isolation.

### II. CLI Interface ✅

**Status**: PASS
**Assessment**: The primary interface is CLI-based with text output to stdout and errors to stderr. The tool follows Unix conventions for output. JSON output support will be added in future versions per spec assumptions.

### III. Test-First (NON-NEGOTIABLE) ✅

**Status**: PASS
**Assessment**: The implementation will follow TDD using github.com/stretchr/testify. Tests will be written and approved before implementation begins. Red-Green-Refactor cycle will be enforced throughout development.

### IV. Observability ✅

**Status**: PASS
**Assessment**: The CLI design naturally supports observability:

- Progress indicator (spinner) provides user feedback during operations
- Errors written to stderr keep stdout clean for data
- Structured logging can be added for debugging without affecting output
- Timeout and error indicators provide diagnostic information inline

### V. Simplicity ✅

**Status**: PASS
**Assessment**: The initial version deliberately excludes:

- Command-line flags/arguments
- Configuration files
- Colorized output
- Export formats beyond plain text

YAGNI principles are strictly applied. Complexity will be justified in Complexity Tracking if needed.

**Overall Gate Status**: ✅ PASS - All constitutional principles are satisfied. Proceeding to Phase 0.

---

### Re-evaluation After Phase 1 Design (2025-11-13)

After completing Phase 1 design artifacts (research.md, data-model.md, contracts/, quickstart.md), re-evaluating constitutional compliance:

#### I. Library-First ✅

**Status**: PASS (CONFIRMED)
**Post-Design Assessment**:

- Library structure is well-defined in contracts/library-api.md
- Four independent libraries identified: scanner, gitstatus, tree, models
- Each library has clear API contract with independent utility
- No organizational-only libraries

#### II. CLI Interface ✅

**Status**: PASS (CONFIRMED)
**Post-Design Assessment**:

- CLI contract fully specified in contracts/cli-contract.md
- Text I/O protocol: stdin not used, stdout for data, stderr for progress/errors
- Exit codes defined (0 for success, 1 for fatal errors)
- Human-readable output format specified
- JSON output reserved for future (maintains simplicity principle)

#### III. Test-First (NON-NEGOTIABLE) ✅

**Status**: PASS (CONFIRMED)
**Post-Design Assessment**:

- Testing strategy documented in research.md
- testify framework selected and documented
- Unit test requirements specified for each library
- Integration test approach defined
- Test coverage goals established (>80%)
- Ready for TDD implementation

#### IV. Observability ✅

**Status**: PASS (CONFIRMED)
**Post-Design Assessment**:

- Progress indicator design complete (spinner on stderr)
- Error indicator system specified (error, timeout, bare, no-remote)
- Diagnostic information inline with output
- Stderr/stdout separation maintained throughout
- Structured logging capability reserved for future

#### V. Simplicity ✅

**Status**: PASS (CONFIRMED)
**Post-Design Assessment**:

- Design maintains "zero flags" simplicity for v1.0.0
- No configuration files or complex setup
- Straightforward data model with 4 core entities
- No premature abstractions or framework dependencies
- Future complexity reserved for later versions
- All design complexity justified by requirements

**Overall Re-evaluation**: ✅ PASS - All constitutional principles remain satisfied post-design. Ready for Phase 2 (task generation via /speckit.tasks).

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── gitree/
    └── main.go           # CLI entry point

internal/
├── scanner/
│   ├── scanner.go        # Directory traversal and Git detection
│   └── scanner_test.go
├── gitstatus/
│   ├── status.go         # Git status extraction (branch, ahead/behind, stashes, changes)
│   └── status_test.go
├── tree/
│   ├── formatter.go      # ASCII tree generation
│   └── formatter_test.go
└── models/
    └── repository.go     # Core data structures (Repository, TreeNode, GitStatus)

pkg/
└── gitree/
    └── gitree.go         # Public API for library usage (if needed for external consumers)

go.mod
go.sum
```

**Structure Decision**: Single Go project following standard Go layout conventions. The `cmd/` directory contains the CLI entry point, `internal/` contains implementation libraries (scanner, gitstatus, tree, models) that align with the Library-First principle, and `pkg/` provides a public API if external packages need to use gitree as a library. Tests are co-located with implementation files following Go conventions.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |

---

## Phase 2: Core Implementation

**Prerequisites**: Phase 0 (research.md) and Phase 1 (data-model.md, contracts/, quickstart.md) complete ✅

**Approach**: Test-First Development (TDD) - Write tests before implementation for each component

**Reference Documents**:
- Data structures: `specs/001-git-repo-tree-viewer/data-model.md`
- Library APIs: `specs/001-git-repo-tree-viewer/contracts/library-api.md`
- CLI contract: `specs/001-git-repo-tree-viewer/contracts/cli-contract.md`
- Implementation patterns: `specs/001-git-repo-tree-viewer/research.md`

### Step 1: Project Setup and Dependencies

**Objective**: Initialize Go module and install required dependencies

**Tasks**:
1. Initialize Go module: `go mod init github.com/user/gitree`
2. Add dependencies per `research.md:545-557`:
   - `go get github.com/go-git/go-git/v5@v5.16.3`
   - `go get github.com/briandowns/spinner@v1.23.2`
   - `go get github.com/stretchr/testify@v1.11.1`
3. Create directory structure per `plan.md:153-178`:
   - `cmd/gitree/`
   - `internal/models/`
   - `internal/scanner/`
   - `internal/gitstatus/`
   - `internal/tree/`
4. Create `.gitignore` for Go (binaries, test coverage)

**Verification**: `go mod tidy` succeeds, directories created

---

### Step 2: Implement Models Package

**Objective**: Create core data structures with validation

**Reference**: `data-model.md` (complete), `library-api.md:429-478`

**Test-First Tasks**:

1. **Create test file** `internal/models/repository_test.go`:
   - Write validation tests for Repository struct per `data-model.md:27-32`
   - Write validation tests for GitStatus struct per `data-model.md:79-81`
   - Write validation tests for TreeNode struct per `data-model.md:183-188`
   - Write validation tests for ScanResult struct per `data-model.md:247-253`
   - Write tests for GitStatus.Format() matching examples in `data-model.md:158-165`
   - **Run tests**: Should fail (red phase)

2. **Implement** `internal/models/repository.go`:
   - Implement Repository struct per `data-model.md:34-60`
   - Implement GitStatus struct per `data-model.md:92-156`
   - Implement TreeNode struct per `data-model.md:196-232`
   - Implement ScanResult struct per `data-model.md:256-310`
   - Implement all Validate() methods matching test expectations
   - Implement GitStatus.Format() per `data-model.md:119-156`
   - **Run tests**: Should pass (green phase)

3. **Refactor** if needed:
   - Extract common validation helpers
   - Add constructor functions (NewRepository, NewGitStatus, etc.)
   - Add ScanResult.HasErrors() and SuccessRate() per `data-model.md:292-309`

**Verification**:
- All tests pass
- Test coverage >80% for models package
- Format examples match `cli-contract.md:73-83`

**Cross-Reference Check**:
- Repository.Validate() enforces `data-model.md:27-32`
- GitStatus.Format() produces output matching `cli-contract.md:59-83`
- Invariants from `data-model.md:434-441` are maintained

---

### Step 3: Implement Scanner Library

**Objective**: Directory traversal and Git repository detection

**Reference**: `library-api.md:21-155`, `research.md:127-178` (traversal), `research.md:487-533` (symlinks)

**Test-First Tasks**:

1. **Create test file** `internal/scanner/scanner_test.go`:
   - Test IsGitRepository() detects regular repos (with .git dir)
   - Test IsGitRepository() detects bare repos per `research.md:155-178`
   - Test Scan() finds all repos in test directory tree
   - Test symlink following (once only) per `research.md:487-533`
   - Test symlink loop prevention (visited inodes tracking)
   - Test skipping nested repo contents per spec `FR-018`
   - Test permission denied error handling (non-fatal)
   - Test context cancellation
   - Test MaxDepth option if implemented
   - Create helper function to set up test repos in temp directories
   - **Run tests**: Should fail (red phase)

2. **Implement** `internal/scanner/scanner.go`:
   - Implement ScanOptions struct per `library-api.md:76-83`
   - Implement IsGitRepository() per `library-api.md:91-127` and `research.md:155-178`
   - Implement Scan() using filepath.WalkDir per `research.md:127-178`
   - Implement symlink handling with inode tracking per `research.md:487-533`
   - Implement early exit for nested repos (fs.SkipDir after finding .git)
   - Handle errors per `library-api.md:130-142`:
     - Fatal errors: root path inaccessible → return error
     - Non-fatal errors: individual dir access → collect in ScanResult.Errors
   - **Run tests**: Should pass (green phase)

3. **Refactor**:
   - Extract visitor struct to track state (visited inodes, results)
   - Optimize directory skipping logic
   - Add documentation comments

**Verification**:
- All tests pass
- Test coverage >80% for scanner package
- Correctly detects both regular and bare repos
- Symlink loops don't cause infinite recursion
- Unit test requirements from `library-api.md:145-155` satisfied

**Cross-Reference Check**:
- Symlink handling matches `spec.md:FR-022`
- Bare repo detection matches `spec.md:FR-026`
- Error handling matches `library-api.md:130-142`

---

### Step 4: Implement Git Status Library

**Objective**: Extract Git status information from repositories

**Reference**: `library-api.md:156-304`, `research.md:9-56` (go-git usage), `research.md:298-367` (status extraction)

**Test-First Tasks**:

1. **Create test file** `internal/gitstatus/status_test.go`:
   - Create helper to initialize test repos with known states (using go-git)
   - Test Extract() gets branch name
   - Test Extract() detects detached HEAD per `research.md:313-324`
   - Test Extract() calculates ahead/behind counts per `research.md:326-347`
   - Test Extract() detects no remote (HasRemote=false) per `spec.md:FR-023`
   - Test Extract() detects stashes per `research.md:349-357`
   - Test Extract() detects uncommitted changes per `research.md:359-366`
   - Test Extract() handles bare repositories (HasChanges=false)
   - Test Extract() handles corrupted repos (returns error)
   - Test Extract() respects context timeout
   - Test ExtractBatch() concurrent processing
   - **Run tests**: Should fail (red phase)

2. **Implement** `internal/gitstatus/status.go`:
   - Implement ExtractOptions struct per `library-api.md:217-229`
   - Implement Extract() per `library-api.md:164-213`
   - Use go-git patterns from `research.md:38-55` for opening repos
   - Implement branch detection per `research.md:313-324`
   - Implement ahead/behind calculation per `research.md:326-347`
   - Implement stash detection per `research.md:349-357`
   - Implement uncommitted changes detection per `research.md:359-366`
   - Handle bare repos (skip worktree operations)
   - Respect context timeout from parameter
   - Return partial status with Error field on non-fatal failures
   - **Run tests**: Should pass (green phase)

3. **Implement** ExtractBatch() for concurrent processing:
   - Implement worker pool pattern per `research.md:59-125`
   - Use semaphore to limit concurrent workers (10-20)
   - Collect results in map[string]*GitStatus
   - Handle per-repository timeouts
   - **Run tests**: Concurrent tests should pass

4. **Refactor**:
   - Extract helper functions for each status component
   - Add timeout wrapper function per `research.md:376-430`
   - Optimize go-git API calls (cache repository references if needed)

**Verification**:
- All tests pass
- Test coverage >80% for gitstatus package
- All unit test requirements from `library-api.md:290-303` satisfied
- Timeout handling works correctly per `spec.md:FR-024, FR-025`

**Cross-Reference Check**:
- Branch/detached detection matches `spec.md:FR-005, FR-006`
- Ahead/behind matches `spec.md:FR-007, FR-008`
- No remote indicator matches `spec.md:FR-023`
- Stash detection matches `spec.md:FR-009`
- Uncommitted changes match `spec.md:FR-010`
- Bare repo handling matches `spec.md:FR-026`

---

### Step 5: Implement Tree Formatter Library

**Objective**: Build tree structure and format ASCII output

**Reference**: `library-api.md:305-428`, `research.md:223-296` (tree formatting)

**Test-First Tasks**:

1. **Create test file** `internal/tree/formatter_test.go`:
   - Test Build() creates correct tree structure from flat repo list
   - Test Build() calculates relative paths correctly
   - Test Build() sorts children alphabetically (deterministic)
   - Test Build() sets depth levels correctly
   - Test Build() marks IsLast flags for last children
   - Test Format() with single repository
   - Test Format() with multiple repos at same level
   - Test Format() with nested repos
   - Test Format() uses correct connectors (├──, └──, │)
   - Test Format() includes Git status inline
   - Test Format() matches examples from `cli-contract.md:176-297`
   - Test empty repository list
   - **Run tests**: Should fail (red phase)

2. **Implement** `internal/tree/tree.go`:
   - Implement FormatOptions struct per `library-api.md:356-369`
   - Implement Build() per `library-api.md:371-404`
   - Calculate relative paths from root
   - Sort children alphabetically using TreeNode.SortChildren() from `data-model.md:223-232`
   - Set depth and IsLast flags correctly
   - **Run tests**: Build tests should pass

3. **Implement** Format() function:
   - Implement Format() per `library-api.md:314-352`
   - Use recursive traversal pattern from `research.md:257-296`
   - Generate ASCII connectors: ├──, └──, │
   - Include Git status using GitStatus.Format()
   - Handle error/timeout indicators per `cli-contract.md:70-72`
   - Handle bare repository indicator per `cli-contract.md:69`
   - **Run tests**: Format tests should pass

4. **Refactor**:
   - Extract connector logic into helper function
   - Optimize string building (use strings.Builder)
   - Add edge case handling (nil status, empty tree)

**Verification**:
- All tests pass
- Test coverage >80% for tree package
- Output exactly matches `cli-contract.md` examples
- All unit test requirements from `library-api.md:417-428` satisfied

**Cross-Reference Check**:
- Tree structure matches `spec.md:FR-004`
- Inline status format matches `spec.md:FR-011`
- ASCII style matches `cli-contract.md:54-72`
- Examples match `cli-contract.md:176-297`

---

### Step 6: Implement CLI Entry Point

**Objective**: Command-line interface integrating all libraries

**Reference**: `cli-contract.md` (complete), `library-api.md:480-548` (integration flow)

**Test-First Tasks**:

1. **Create integration test** `cmd/gitree/main_test.go`:
   - Test end-to-end: create test repos → run main → verify output
   - Test "No Git repositories found" case per `cli-contract.md:85-99`
   - Test exit code 0 on success per `cli-contract.md:139-147`
   - Test exit code 1 on fatal error
   - Test progress indicator appears on stderr
   - Test output goes to stdout (not stderr)
   - Mock filesystem for testing (or use temp dirs)
   - **Run tests**: Should fail (red phase)

2. **Implement** `cmd/gitree/main.go`:
   - Get current working directory
   - Initialize spinner per `research.md:209-221`:
     - Use briandowns/spinner
     - Output to stderr
     - Message: "Scanning repositories..."
   - Start spinner
   - Call scanner.Scan() per integration flow `library-api.md:499-507`
   - Handle "no repos found" case per `cli-contract.md:85-99`
   - Extract Git status using gitstatus.ExtractBatch() per `library-api.md:514-529`
   - Apply timeout wrapper (5-10 seconds per repo) per `spec.md:FR-024`
   - Populate repositories with status per `library-api.md:526-529`
   - Build tree using tree.Build() per `library-api.md:531-536`
   - Stop spinner
   - Format output using tree.Format() per `library-api.md:538-544`
   - Print to stdout
   - Handle errors with appropriate exit codes per `cli-contract.md:139-147`
   - **Run tests**: Should pass (green phase)

3. **Refactor**:
   - Extract error handling logic
   - Add defer for cleanup (spinner.Stop)
   - Add documentation comments

**Verification**:
- Integration tests pass
- Manual test: `go run cmd/gitree/main.go` in test directory works
- Output matches `cli-contract.md` examples exactly
- Spinner appears on stderr, output on stdout
- Exit codes correct per `cli-contract.md:139-147`

**Cross-Reference Check**:
- No arguments required per `spec.md:FR-019, FR-021`
- Progress indicator per `spec.md:FR-013, FR-014`
- Output format matches `cli-contract.md:40-83`
- Execution flow matches `cli-contract.md:149-160`

---

### Step 7: Error Handling and Edge Cases

**Objective**: Robust error handling for real-world scenarios

**Reference**: `spec.md` edge cases (lines 105-114), `research.md:376-430` (error handling)

**Test-First Tasks**:

1. **Add edge case tests** across all packages:
   - Scanner: permission denied, symlink to non-existent target
   - GitStatus: timeout, corrupted .git, no HEAD file
   - Tree: nil GitStatus, empty children
   - CLI: empty directory, all repos fail, partial failures
   - **Run tests**: Should fail (red phase)

2. **Implement graceful degradation**:
   - Timeout handling: Apply per-repo timeout, show "timeout" indicator per `spec.md:FR-024, FR-025`
   - Corrupted repo: Show "error" indicator per `spec.md:FR-027, FR-028`
   - Partial failures: Continue processing, collect errors in ScanResult.Errors
   - Non-fatal errors don't cause exit code 1 per `cli-contract.md:147`
   - **Run tests**: Should pass (green phase)

3. **Add error messages to stderr**:
   - Warning messages per `cli-contract.md:119-137`
   - Format: "Error: {description}" or "Warning: {description}"
   - Include context (paths, operation details)

**Verification**:
- All edge case tests pass
- Tool doesn't crash on any error scenario
- Error indicators appear correctly in output
- Non-fatal errors logged to stderr but don't prevent output

**Cross-Reference Check**:
- Graceful degradation per `spec.md:FR-015, FR-016`
- Error indicators match `cli-contract.md:70-72`
- Error messages match `cli-contract.md:119-137`

---

### Step 8: Build and Installation

**Objective**: Create distributable binary

**Tasks**:

1. **Create build script** or document build commands:
   - `go build -o gitree cmd/gitree/main.go`
   - Test binary: `./gitree`
   - Verify version compatibility (Go 1.25+)

2. **Test installation**:
   - `go install ./cmd/gitree`
   - Verify `gitree` in $GOPATH/bin works

3. **Cross-platform build** (if desired):
   - Build for Linux, macOS, Windows
   - Test on each platform if possible

**Verification**:
- Binary runs without Go installed (static linking check)
- `gitree` command works from any directory
- No runtime dependencies beyond system libraries

---

## Phase 3: Refinement and Polish

**Prerequisites**: Phase 2 complete, all core functionality working

### Step 1: Performance Optimization

**Objective**: Meet performance goals from `spec.md:SC-003, SC-005`

**Tasks**:

1. **Performance testing**:
   - Create test with 50+ repositories
   - Measure execution time (target: <10 seconds)
   - Measure spinner startup time (target: <1 second)
   - Profile with `go test -bench` and `pprof`

2. **Optimize bottlenecks**:
   - Adjust worker pool size in ExtractBatch() (test 5, 10, 20, 50 workers)
   - Cache go-git repository objects if multiple operations needed
   - Use buffered channels for result collection
   - Optimize tree sorting (already O(k log k), but verify)

3. **Verify performance goals**:
   - 50 repos processed in <10 seconds per `spec.md:SC-003`
   - Spinner appears within 1 second per `spec.md:SC-005`
   - Document actual performance characteristics

**Verification**:
- Benchmark tests show acceptable performance
- Performance goals met on standard hardware
- No significant memory leaks (verify with pprof)

---

### Step 2: Output Format Verification

**Objective**: Ensure output exactly matches specification

**Tasks**:

1. **Visual testing**:
   - Create test repos with all status combinations
   - Run gitree and compare to `cli-contract.md` examples
   - Verify tree structure matches Unix `tree` command style
   - Check all indicator symbols render correctly: ↑, ↓, ○, $, *

2. **Contract compliance tests**:
   - Write tests for each example in `cli-contract.md:176-297`
   - Verify stdout/stderr separation per `cli-contract.md:101-137`
   - Test exit codes per `cli-contract.md:139-147`

3. **Fix any discrepancies**:
   - Adjust formatting to match examples exactly
   - Fix connector characters if misaligned
   - Verify spacing and indentation

**Verification**:
- Output matches all examples in `cli-contract.md`
- Contract verification tests pass per `cli-contract.md:299-337`
- Visual inspection confirms clean, readable output

---

### Step 3: Documentation Completion

**Objective**: Ensure all documentation is accurate and helpful

**Tasks**:

1. **Update README.md** (if not already created):
   - Add installation instructions per `quickstart.md:13-37`
   - Add usage examples per `quickstart.md:39-68`
   - Link to detailed docs

2. **Verify quickstart.md accuracy**:
   - Test all commands in `quickstart.md`
   - Verify examples produce shown output
   - Update any outdated information

3. **Add code documentation**:
   - Add package-level godoc comments
   - Add function-level comments for exported functions
   - Add examples in godoc format

4. **Create CONTRIBUTING.md** (optional):
   - Testing requirements
   - Code style guidelines
   - Pull request process

**Verification**:
- `go doc` shows helpful documentation
- All examples in quickstart.md work correctly
- Documentation is beginner-friendly

---

### Step 4: Final Testing and Validation

**Objective**: Comprehensive validation of all requirements

**Tasks**:

1. **Run full test suite**:
   - `go test ./...` passes with no failures
   - Coverage report shows >80% per `research.md:480-485`
   - All edge cases covered

2. **Requirements traceability**:
   - Verify each FR in `spec.md:116-147` is implemented
   - Check each success criterion in `spec.md:156-167`
   - Review user stories in `spec.md:18-102` are satisfied

3. **Manual acceptance testing**:
   - Test all scenarios from `quickstart.md:110-158`
   - Test troubleshooting scenarios from `quickstart.md:159-224`
   - Create test directory structures matching examples

4. **Create acceptance checklist** (use `checklists/requirements.md` if exists):
   - Check all functional requirements
   - Check all success criteria
   - Check constitutional compliance
   - Sign off on completion

**Verification**:
- All tests pass
- All requirements verified
- No known bugs or issues
- Ready for release

---

### Step 5: Release Preparation

**Objective**: Prepare for v1.0.0 release

**Tasks**:

1. **Version tagging**:
   - Update version references to 1.0.0
   - Create git tag: `git tag v1.0.0`

2. **Release notes**:
   - Document features
   - List known limitations
   - Acknowledge dependencies

3. **Distribution**:
   - Build binaries for target platforms
   - Create release on GitHub (if applicable)
   - Publish to package manager (if applicable)

**Verification**:
- Version 1.0.0 tagged in git
- Release notes accurate
- Installation process documented and tested

---

## Implementation Checklist

**Phase 2: Core Implementation**
- [ ] Step 1: Project setup and dependencies
- [ ] Step 2: Models package (TDD)
- [ ] Step 3: Scanner library (TDD)
- [ ] Step 4: Git status library (TDD)
- [ ] Step 5: Tree formatter library (TDD)
- [ ] Step 6: CLI entry point (TDD)
- [ ] Step 7: Error handling and edge cases
- [ ] Step 8: Build and installation

**Phase 3: Refinement**
- [ ] Step 1: Performance optimization
- [ ] Step 2: Output format verification
- [ ] Step 3: Documentation completion
- [ ] Step 4: Final testing and validation
- [ ] Step 5: Release preparation

---

## Success Criteria Review

Upon completion of Phase 3, verify all success criteria from `spec.md:156-167`:

- ✅ **SC-001**: Users can visualize all Git repositories without navigating
- ✅ **SC-002**: Users identify branch/status within 3 seconds
- ✅ **SC-003**: Tool scans 50+ repos in <10 seconds
- ✅ **SC-004**: Handles corrupted repos without crashing
- ✅ **SC-005**: Progress indicator appears within 1 second
- ✅ **SC-006**: Output matches tree command visual structure
- ✅ **SC-007**: 100% of Git repos detected
- ✅ **SC-008**: All status indicators accurate

---

## Notes for Implementation

**TDD Workflow**:
1. Write failing test (red)
2. Implement minimal code to pass (green)
3. Refactor for quality (refactor)
4. Repeat

**Cross-Reference Usage**:
- When implementing each component, actively reference the detail files
- Line numbers in references help locate exact specifications
- Verify implementation matches both API contracts and examples

**Testing Strategy**:
- Unit tests: Test each library in isolation
- Integration tests: Test CLI end-to-end
- Manual tests: Real-world directory structures
- Performance tests: Benchmark critical paths

**Constitutional Compliance**:
- Library-First: Each library has independent utility ✅
- CLI Interface: Text I/O, no GUI ✅
- Test-First: TDD enforced throughout ✅
- Observability: Spinner on stderr, errors logged ✅
- Simplicity: No flags, minimal features ✅
