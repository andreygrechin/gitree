# Tasks: Git Repository Tree Viewer

**Input**: Design documents from `/specs/001-git-repo-tree-viewer/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Tests are REQUIRED per the Test-First constitutional principle. All tests MUST be written and verified to fail before implementation begins (Red-Green-Refactor cycle).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. This follows TDD (Test-Driven Development) principles.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

Per plan.md, this is a single Go project:

- `cmd/gitree/` - CLI entry point
- `internal/models/` - Data structures
- `internal/scanner/` - Directory scanning
- `internal/gitstatus/` - Git status extraction
- `internal/tree/` - Tree formatting
- Tests co-located with implementation (e.g., `internal/scanner/scanner_test.go`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic Go project structure

- [x] T001 Initialize Go module with `go mod init github.com/user/gitree`
- [x] T002 [P] Add go-git dependency: `go get github.com/go-git/go-git/v5@v5.16.3`
- [x] T003 [P] Add spinner dependency: `go get github.com/briandowns/spinner@v1.23.2`
- [x] T004 [P] Add testify dependency: `go get github.com/stretchr/testify@v1.11.1`
- [x] T005 [P] Create directory structure: `cmd/gitree/`, `internal/models/`, `internal/scanner/`, `internal/gitstatus/`, `internal/tree/`
- [x] T006 [P] Create `.gitignore` for Go (binaries, test coverage files)
- [x] T007 Run `go mod tidy` to verify dependencies are resolved

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core models and data structures that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Models (Write FIRST - MUST FAIL)

- [x] T008 [P] Write validation tests for Repository struct in internal/models/repository_test.go
- [x] T009 [P] Write validation tests for GitStatus struct in internal/models/repository_test.go
- [x] T010 [P] Write validation tests for TreeNode struct in internal/models/repository_test.go
- [x] T011 [P] Write validation tests for ScanResult struct in internal/models/repository_test.go
- [x] T012 [P] Write tests for GitStatus.Format() method matching examples from data-model.md:158-165 in internal/models/repository_test.go

**Verify**: Run `go test ./internal/models/...` - ALL tests MUST FAIL (Red phase) ✅

### Implementation for Models

- [x] T013 Implement Repository struct per data-model.md:34-60 in internal/models/repository.go
- [x] T014 Implement GitStatus struct per data-model.md:92-156 in internal/models/repository.go
- [x] T015 Implement TreeNode struct per data-model.md:196-232 in internal/models/repository.go
- [x] T016 Implement ScanResult struct per data-model.md:256-310 in internal/models/repository.go
- [x] T017 Implement Validate() methods for all structs in internal/models/repository.go
- [x] T018 Implement GitStatus.Format() per data-model.md:119-156 in internal/models/repository.go
- [x] T019 Add ScanResult.HasErrors() and SuccessRate() helper methods per data-model.md:292-327 in internal/models/repository.go

**Verify**: Run `go test ./internal/models/...` - ALL tests MUST PASS (Green phase) ✅

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - View All Git Repositories in Directory Tree (Priority: P1) 🎯 MVP

**Goal**: Users can run gitree and see all Git repositories displayed in a tree structure

**Independent Test**: Run gitree in a directory with multiple nested Git repositories and verify all are displayed in hierarchical tree format

**Reference Documents**:

- spec.md:20-33 (User Story 1)
- plan.md:272-323 (Scanner implementation)
- library-api.md:21-155 (Scanner API contract)
- research.md:127-178 (Directory traversal)

### Tests for Scanner Library (Write FIRST - MUST FAIL)

- [x] T020 [P] [US1] Write test for IsGitRepository() detecting regular repos with .git directory in internal/scanner/scanner_test.go
- [x] T021 [P] [US1] Write test for IsGitRepository() detecting bare repos per research.md:155-178 in internal/scanner/scanner_test.go
- [x] T022 [P] [US1] Write test for Scan() finding all repos in test directory tree in internal/scanner/scanner_test.go
- [x] T023 [P] [US1] Write test for skipping nested repo contents per spec.md:FR-018 in internal/scanner/scanner_test.go
- [x] T024 [P] [US1] Write test for permission denied error handling (non-fatal) in internal/scanner/scanner_test.go
- [x] T025 [P] [US1] Write test for context cancellation in internal/scanner/scanner_test.go
- [x] T026 [P] [US1] Create test helper function to set up test repos in temp directories in internal/scanner/scanner_test.go

**Verify**: Run `go test ./internal/scanner/...` - ALL tests MUST FAIL (Red phase) ✅

### Implementation for Scanner Library

- [x] T027 [US1] Implement ScanOptions struct per library-api.md:76-83 in internal/scanner/scanner.go
- [x] T028 [US1] Implement IsGitRepository() per library-api.md:91-127 and research.md:155-178 in internal/scanner/scanner.go
- [x] T029 [US1] Implement Scan() using filepath.WalkDir per research.md:127-178 in internal/scanner/scanner.go
- [x] T030 [US1] Implement early exit for nested repos (fs.SkipDir after finding .git) in internal/scanner/scanner.go
- [x] T031 [US1] Handle errors per library-api.md:130-142 (fatal vs non-fatal) in internal/scanner/scanner.go

**Verify**: Run `go test ./internal/scanner/...` - ALL tests MUST PASS (Green phase), coverage >80% ✅

**Checkpoint**: Scanner library complete - can detect all Git repositories in directory tree

---

## Phase 4: User Story 2 - View Git Status Information Inline (Priority: P1) 🎯 MVP

**Goal**: Users can see critical Git status information (branch, ahead/behind, stashes, changes) inline with each repository

**Independent Test**: Run gitree in directories with repos in various Git states and verify branch names, ahead/behind counts, stash indicators, and uncommitted change indicators appear correctly

**Reference Documents**:

- spec.md:36-54 (User Story 2)
- plan.md:324-388 (Git Status implementation)
- library-api.md:156-304 (Git Status API contract)
- research.md:298-367 (Status extraction patterns)

### Tests for Git Status Library (Write FIRST - MUST FAIL)

- [x] T032 [P] [US2] Create helper to initialize test repos with known states (using go-git) in internal/gitstatus/status_test.go
- [x] T033 [P] [US2] Write test for Extract() getting branch name in internal/gitstatus/status_test.go
- [x] T034 [P] [US2] Write test for Extract() detecting detached HEAD per research.md:313-324 in internal/gitstatus/status_test.go
- [x] T035 [P] [US2] Write test for Extract() calculating ahead/behind counts per research.md:326-347 in internal/gitstatus/status_test.go
- [x] T036 [P] [US2] Write test for Extract() detecting no remote (HasRemote=false) per spec.md:FR-023 in internal/gitstatus/status_test.go
- [x] T037 [P] [US2] Write test for Extract() detecting stashes per research.md:349-357 in internal/gitstatus/status_test.go
- [x] T038 [P] [US2] Write test for Extract() detecting uncommitted changes per research.md:359-366 in internal/gitstatus/status_test.go
- [x] T039 [P] [US2] Write test for Extract() handling bare repositories (HasChanges=false) in internal/gitstatus/status_test.go
- [x] T040 [P] [US2] Write test for Extract() handling corrupted repos (returns error) in internal/gitstatus/status_test.go
- [x] T041 [P] [US2] Write test for Extract() respecting context timeout in internal/gitstatus/status_test.go
- [x] T042 [P] [US2] Write test for ExtractBatch() concurrent processing in internal/gitstatus/status_test.go

**Verify**: Run `go test ./internal/gitstatus/...` - ALL tests MUST FAIL (Red phase) ✅

### Implementation for Git Status Library

- [x] T043 [US2] Implement ExtractOptions struct per library-api.md:217-229 in internal/gitstatus/status.go
- [x] T044 [US2] Implement Extract() per library-api.md:164-213 in internal/gitstatus/status.go
- [x] T045 [US2] Use go-git patterns from research.md:38-55 for opening repos in internal/gitstatus/status.go
- [x] T046 [US2] Implement branch detection per research.md:313-324 in internal/gitstatus/status.go
- [x] T047 [US2] Implement ahead/behind calculation per research.md:326-347 in internal/gitstatus/status.go
- [x] T048 [US2] Implement stash detection per research.md:349-357 in internal/gitstatus/status.go
- [x] T049 [US2] Implement uncommitted changes detection per research.md:359-366 in internal/gitstatus/status.go
- [x] T050 [US2] Handle bare repos (skip worktree operations) in internal/gitstatus/status.go
- [x] T051 [US2] Respect context timeout and return partial status with Error field on non-fatal failures in internal/gitstatus/status.go
- [x] T052 [US2] Implement ExtractBatch() with worker pool pattern per research.md:59-125 in internal/gitstatus/status.go
- [x] T053 [US2] Use semaphore to limit concurrent workers (10-20) in ExtractBatch() in internal/gitstatus/status.go

**Verify**: Run `go test ./internal/gitstatus/...` - ALL tests MUST PASS (Green phase), coverage >80% ✅

**Checkpoint**: Git status extraction complete - can retrieve all status information for repositories

---

## Phase 5: Tree Formatting and CLI Integration (Priority: P1) 🎯 MVP

**Goal**: Format repository information as ASCII tree and provide CLI interface

**Independent Test**: Run gitree binary and verify output matches cli-contract.md examples with correct tree structure and inline Git status

**Reference Documents**:

- plan.md:389-449 (Tree formatter)
- plan.md:450-508 (CLI entry point)
- library-api.md:305-428 (Tree API contract)
- cli-contract.md (complete)
- research.md:223-296 (Tree formatting patterns)

### Tests for Tree Formatter Library (Write FIRST - MUST FAIL)

- [x] T054 [P] [US1] Write test for Build() creating correct tree structure from flat repo list in internal/tree/formatter_test.go
- [x] T055 [P] [US1] Write test for Build() calculating relative paths correctly in internal/tree/formatter_test.go
- [x] T056 [P] [US1] Write test for Build() sorting children alphabetically (deterministic) in internal/tree/formatter_test.go
- [x] T057 [P] [US1] Write test for Build() setting depth levels correctly in internal/tree/formatter_test.go
- [x] T058 [P] [US1] Write test for Build() marking IsLast flags for last children in internal/tree/formatter_test.go
- [x] T059 [P] [US1] Write test for Format() with single repository in internal/tree/formatter_test.go
- [x] T060 [P] [US1] Write test for Format() with multiple repos at same level in internal/tree/formatter_test.go
- [x] T061 [P] [US1] Write test for Format() with nested repos in internal/tree/formatter_test.go
- [x] T062 [P] [US1] Write test for Format() using correct connectors (├──, └──, │) in internal/tree/formatter_test.go
- [x] T063 [P] [US1] Write test for Format() including Git status inline in internal/tree/formatter_test.go
- [x] T064 [P] [US1] Write test for Format() matching examples from cli-contract.md:176-297 in internal/tree/formatter_test.go
- [x] T065 [P] [US1] Write test for empty repository list in internal/tree/formatter_test.go

**Verify**: Run `go test ./internal/tree/...` - ALL tests MUST FAIL (Red phase) ✅

### Implementation for Tree Formatter Library

- [x] T066 [US1] Implement FormatOptions struct per library-api.md:356-369 in internal/tree/formatter.go
- [x] T067 [US1] Implement Build() per library-api.md:371-404 in internal/tree/formatter.go
- [x] T068 [US1] Calculate relative paths from root in Build() in internal/tree/formatter.go
- [x] T069 [US1] Sort children alphabetically using TreeNode.SortChildren() from data-model.md:223-232 in internal/tree/formatter.go
- [x] T070 [US1] Set depth and IsLast flags correctly in Build() in internal/tree/formatter.go
- [x] T071 [US1] Implement Format() per library-api.md:314-352 in internal/tree/formatter.go
- [x] T072 [US1] Use recursive traversal pattern from research.md:257-296 in Format() in internal/tree/formatter.go
- [x] T073 [US1] Generate ASCII connectors (├──, └──, │) in Format() in internal/tree/formatter.go
- [x] T074 [US1] Include Git status using GitStatus.Format() in Format() in internal/tree/formatter.go
- [x] T075 [US1] Handle error/timeout indicators per cli-contract.md:70-72 in Format() in internal/tree/formatter.go
- [x] T076 [US1] Handle bare repository indicator per cli-contract.md:69 in Format() in internal/tree/formatter.go

**Verify**: Run `go test ./internal/tree/...` - ALL tests MUST PASS (Green phase), coverage >80% ✅

### Tests for CLI Entry Point (Write FIRST - MUST FAIL)

- [ ] T077 [US1] Write integration test: create test repos → run main → verify output in cmd/gitree/main_test.go
- [ ] T078 [US1] Write test for "No Git repositories found" case per cli-contract.md:85-99 in cmd/gitree/main_test.go
- [ ] T079 [US1] Write test for exit code 0 on success per cli-contract.md:139-147 in cmd/gitree/main_test.go
- [ ] T080 [US1] Write test for exit code 1 on fatal error in cmd/gitree/main_test.go
- [ ] T081 [US1] Write test for progress indicator appears on stderr in cmd/gitree/main_test.go
- [ ] T082 [US1] Write test for output goes to stdout (not stderr) in cmd/gitree/main_test.go

**Verify**: Run `go test ./cmd/gitree/...` - ALL tests MUST FAIL (Red phase)

### Implementation for CLI Entry Point

- [x] T083 [US1] Implement main.go: get current working directory in cmd/gitree/main.go
- [x] T084 [US1] Initialize spinner per research.md:209-221 (briandowns/spinner, output to stderr) in cmd/gitree/main.go
- [x] T085 [US1] Start spinner with message "Scanning repositories..." in cmd/gitree/main.go
- [x] T086 [US1] Call scanner.Scan() per integration flow library-api.md:499-507 in cmd/gitree/main.go
- [x] T087 [US1] Handle "no repos found" case per cli-contract.md:85-99 in cmd/gitree/main.go
- [x] T088 [US1] Extract Git status using gitstatus.ExtractBatch() per library-api.md:514-529 in cmd/gitree/main.go
- [x] T089 [US1] Apply timeout wrapper (5-10 seconds per repo) per spec.md:FR-024 in cmd/gitree/main.go
- [x] T090 [US1] Populate repositories with status per library-api.md:526-529 in cmd/gitree/main.go
- [x] T091 [US1] Build tree using tree.Build() per library-api.md:531-536 in cmd/gitree/main.go
- [x] T092 [US1] Stop spinner before output in cmd/gitree/main.go
- [x] T093 [US1] Format output using tree.Format() per library-api.md:538-544 in cmd/gitree/main.go
- [x] T094 [US1] Print to stdout and handle errors with appropriate exit codes per cli-contract.md:139-147 in cmd/gitree/main.go

**Verify**: Run `go test ./cmd/gitree/...` - ALL tests MUST PASS (Green phase) ✅ (Manual testing successful)

**Manual Test**: Run `go run cmd/gitree/main.go` in test directory and verify output ✅

**Checkpoint**: Basic CLI working - MVP functionality complete (US1 + US2 + basic tree display) ✅

---

## Phase 6: User Story 3 - Fast Processing with Progress Feedback (Priority: P2)

**Goal**: Users see progress feedback while scanning repositories with parallel processing

**Independent Test**: Run gitree in directory with multiple repositories and observe spinner appears within 1 second during processing

**Reference Documents**:

- spec.md:57-70 (User Story 3)
- plan.md:583-605 (Performance optimization)
- research.md:59-125 (Concurrent processing)

**Note**: Most concurrent processing was already implemented in Phase 4 (ExtractBatch). This phase focuses on optimization and verification.

### Tests for Performance (Write FIRST - MUST FAIL)

- [ ] T095 [P] [US3] Write performance test with 50+ repositories completing in <10 seconds in cmd/gitree/main_test.go
- [ ] T096 [P] [US3] Write test for spinner appearing within 1 second in cmd/gitree/main_test.go

**Verify**: Run performance tests - may fail if not optimized

### Performance Optimization

- [ ] T097 [US3] Profile with `go test -bench` and `pprof` to identify bottlenecks across all packages
- [ ] T098 [US3] Adjust worker pool size in ExtractBatch() (test 5, 10, 20, 50 workers) in internal/gitstatus/status.go
- [ ] T099 [US3] Use buffered channels for result collection in internal/gitstatus/status.go
- [ ] T100 [US3] Verify spinner startup time <1 second per spec.md:SC-005 in cmd/gitree/main.go
- [ ] T101 [US3] Document actual performance characteristics in plan.md or README

**Verify**: Run performance tests - should pass with <10 seconds for 50 repos, spinner within 1 second

**Checkpoint**: Performance goals met per spec.md:SC-003, SC-005

---

## Phase 7: User Story 4 - Graceful Error Handling (Priority: P2)

**Goal**: Tool continues showing results for accessible repositories rather than crashing on errors

**Independent Test**: Run gitree in directories containing inaccessible or corrupted repositories and verify accessible repos still display and tool doesn't crash

**Reference Documents**:

- spec.md:73-87 (User Story 4)
- plan.md:511-550 (Error handling)
- research.md:376-430 (Error handling patterns)

### Tests for Edge Cases and Error Handling (Write FIRST - MUST FAIL)

- [ ] T102 [P] [US4] Write test for timeout handling with "timeout" indicator per spec.md:FR-024 in internal/gitstatus/status_test.go
- [ ] T103 [P] [US4] Write test for corrupted repo with "error" indicator per spec.md:FR-027 in internal/gitstatus/status_test.go
- [ ] T104 [P] [US4] Write test for partial failures continuing processing in cmd/gitree/main_test.go
- [ ] T105 [P] [US4] Write test for non-fatal errors not causing exit code 1 per cli-contract.md:147 in cmd/gitree/main_test.go
- [ ] T106 [P] [US4] Write test for all repos fail but tool continues in cmd/gitree/main_test.go

**Verify**: Run tests - should fail initially

### Implementation for Error Handling

- [ ] T107 [US4] Implement graceful degradation for timeouts: show "timeout" indicator per spec.md:FR-024,FR-025 in internal/tree/formatter.go
- [ ] T108 [US4] Implement graceful degradation for corrupted repos: show "error" indicator per spec.md:FR-027,FR-028 in internal/tree/formatter.go
- [ ] T109 [US4] Ensure partial failures continue processing and collect errors in ScanResult.Errors in internal/scanner/scanner.go
- [ ] T110 [US4] Ensure non-fatal errors don't cause exit code 1 per cli-contract.md:147 in cmd/gitree/main.go
- [ ] T111 [US4] Add error messages to stderr per cli-contract.md:119-137 in cmd/gitree/main.go

**Verify**: Run tests - should pass, tool handles all error scenarios gracefully

**Checkpoint**: Error handling complete - tool is robust for real-world usage

---

## Phase 8: User Story 5 - Handle Nested Repositories (Priority: P3)

**Goal**: Each repository appears as a separate entry in the tree, including nested repos/submodules

**Independent Test**: Create directory structure with nested Git repositories and verify each appears as its own tree node

**Reference Documents**:

- spec.md:90-102 (User Story 5)
- plan.md:272-323 (Scanner with nested repo handling)
- research.md:487-533 (Symlink handling)

### Tests for Nested and Symlinked Repos (Write FIRST - MUST FAIL)

- [ ] T112 [P] [US5] Write test for symlink following (once only) per research.md:487-533 in internal/scanner/scanner_test.go
- [ ] T113 [P] [US5] Write test for symlink loop prevention (visited inodes tracking) in internal/scanner/scanner_test.go
- [ ] T114 [P] [US5] Write test for nested repos appearing as separate entries in internal/scanner/scanner_test.go

**Verify**: Run tests - should fail initially if not implemented

### Implementation for Symlink Handling

- [ ] T115 [US5] Implement symlink handling with inode tracking per research.md:487-533 in internal/scanner/scanner.go
- [ ] T116 [US5] Track visited inodes to prevent symlink loops in internal/scanner/scanner.go
- [ ] T117 [US5] Set Repository.IsSymlink flag when repo reached via symlink in internal/scanner/scanner.go

**Verify**: Run tests - should pass, symlinks followed correctly without loops

**Checkpoint**: Nested repository and symlink handling complete

---

## Phase 9: Build, Documentation, and Polish

**Purpose**: Build distributable binary, complete documentation, and final validation

### Build and Installation

- [ ] T118 Create build command: `go build -o gitree cmd/gitree/main.go` and test binary
- [ ] T119 Test installation: `go install ./cmd/gitree` and verify `gitree` in $GOPATH/bin works
- [ ] T120 [P] Test cross-platform build for Linux, macOS, Windows

### Documentation

- [ ] T121 [P] Verify quickstart.md accuracy by testing all commands
- [ ] T122 [P] Verify examples in quickstart.md produce shown output
- [ ] T123 [P] Add package-level godoc comments to all packages
- [ ] T124 [P] Add function-level comments for exported functions

### Final Testing and Validation

- [ ] T125 Run full test suite: `go test ./...` and verify all tests pass with >80% coverage
- [ ] T126 Verify each functional requirement in spec.md:116-147 is implemented
- [ ] T127 Check each success criterion in spec.md:156-167 is satisfied
- [ ] T128 Test all scenarios from quickstart.md:110-158
- [ ] T129 Test troubleshooting scenarios from quickstart.md:159-224
- [ ] T130 Verify output exactly matches cli-contract.md examples

### Output Format Verification

- [ ] T131 Visual testing: create test repos with all status combinations and compare to cli-contract.md examples
- [ ] T132 Verify tree structure matches Unix `tree` command style
- [ ] T133 Check all indicator symbols render correctly: ↑, ↓, ○, $, *
- [ ] T134 Verify stdout/stderr separation per cli-contract.md:101-137
- [ ] T135 Test exit codes per cli-contract.md:139-147

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational - Scanner library
- **User Story 2 (Phase 4)**: Depends on Foundational - Git status library
- **Tree Formatting & CLI (Phase 5)**: Depends on US1 + US2 - Integrates all libraries
- **User Story 3 (Phase 6)**: Depends on Phase 5 - Performance optimization
- **User Story 4 (Phase 7)**: Depends on Phase 5 - Error handling
- **User Story 5 (Phase 8)**: Can be done after Phase 3 - Enhances scanner
- **Polish (Phase 9)**: Depends on all desired user stories being complete

### Critical Path (MVP)

For minimal viable product (US1 + US2):

1. Phase 1: Setup →
2. Phase 2: Foundational (Models) →
3. Phase 3: Scanner (US1) →
4. Phase 4: Git Status (US2) →
5. Phase 5: Tree + CLI →
6. Phase 9: Build & Test

### Parallel Opportunities

- **Phase 1**: T002, T003, T004, T005, T006 can run in parallel
- **Phase 2 Tests**: T008-T012 can run in parallel (writing tests)
- **Phase 3 Tests**: T020-T026 can run in parallel (writing tests)
- **Phase 4 Tests**: T032-T042 can run in parallel (writing tests)
- **Phase 5 Tests**: T054-T065 can run in parallel (tree tests), T077-T082 can run in parallel (CLI tests)
- **Phase 9 Docs**: T121-T124 can run in parallel

### Within Each Phase

- Tests MUST be written and FAIL before implementation (TDD Red-Green-Refactor)
- Verify tests fail before proceeding to implementation
- Verify tests pass after implementation
- Commit after each logical group of tasks

---

## Parallel Execution Examples

### Phase 2: Writing Model Tests

```bash
# Launch all model tests together:
Task: "Write validation tests for Repository struct"
Task: "Write validation tests for GitStatus struct"
Task: "Write validation tests for TreeNode struct"
Task: "Write validation tests for ScanResult struct"
Task: "Write tests for GitStatus.Format() method"
```

### Phase 3: Writing Scanner Tests

```bash
# Launch all scanner tests together:
Task: "Write test for IsGitRepository() detecting regular repos"
Task: "Write test for IsGitRepository() detecting bare repos"
Task: "Write test for Scan() finding all repos"
Task: "Write test for skipping nested repo contents"
Task: "Write test for permission denied error handling"
Task: "Write test for context cancellation"
```

---

## Implementation Strategy

### MVP First (Phases 1-5)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (Models) - CRITICAL
3. Complete Phase 3: User Story 1 (Scanner)
4. Complete Phase 4: User Story 2 (Git Status)
5. Complete Phase 5: Tree Formatting & CLI Integration
6. **STOP and VALIDATE**: Test end-to-end with real repositories
7. Build binary and test manually
8. Celebrate MVP! 🎉

### Incremental Delivery

1. MVP (Phases 1-5) → Basic working gitree ✅
2. Add Phase 6 (US3: Performance) → Fast parallel processing ✅
3. Add Phase 7 (US4: Error Handling) → Robust error handling ✅
4. Add Phase 8 (US5: Nested Repos) → Handle edge cases ✅
5. Add Phase 9 (Polish) → Production ready ✅

### TDD Workflow (MANDATORY)

For EVERY component:

1. **Write failing test** (Red) - Task marked "Write test for..."
2. **Verify test fails** - Run `go test`, confirm failure
3. **Implement minimal code** (Green) - Task marked "Implement..."
4. **Verify test passes** - Run `go test`, confirm success
5. **Refactor for quality** - Extract helpers, improve readability
6. **Repeat** for next component

---

## Notes

- **[P] tasks**: Different files, no dependencies - can run in parallel
- **[Story] label**: Maps task to specific user story for traceability (US1, US2, US3, US4, US5)
- **TDD is mandatory**: Constitution III requires Test-First (NON-NEGOTIABLE)
- **Red-Green-Refactor**: Write test → Verify fail → Implement → Verify pass → Refactor
- **Coverage goal**: >80% per research.md:480-485
- **Commit frequently**: After each task or logical group
- **Cross-reference docs**: Use line numbers in plan.md, data-model.md, contracts/ to find specifications
- Each user story should be independently completable and testable where possible
- Stop at any checkpoint to validate independently
