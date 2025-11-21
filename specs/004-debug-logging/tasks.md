# Tasks: Debug Logging

**Input**: Design documents from `/specs/004-debug-logging/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Per Constitution Principle III (Test-First), tests are written BEFORE implementation and user approval must be obtained before proceeding with implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Gitree uses standard Go single-project layout:

- Source code: `cmd/` (entry points), `internal/` (implementation packages)
- Tests: Co-located with implementation as `*_test.go` files

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add infrastructure for debug flag passing through the application

- [x] T001 [P] Add Debug field to ScanOptions struct in internal/scanner/scanner.go
- [x] T002 [P] Add Debug field to ExtractOptions struct in internal/gitstatus/status.go
- [x] T003 [P] Add debugPrintf helper function in internal/scanner/scanner.go
- [x] T004 [P] Add debugPrintf helper function in internal/gitstatus/status.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core CLI flag parsing that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Add debugFlag variable in cmd/gitree/root.go alongside existing flags
- [x] T006 Add debug flag definition in cmd/gitree/root.go init() function
- [x] T007 Pass debug flag to ScanOptions in cmd/gitree/root.go runGitree() function
- [x] T008 Pass debug flag to ExtractOptions in cmd/gitree/root.go runGitree() function

**Checkpoint**: Debug flag infrastructure ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 3 - Debug Flag Control (Priority: P1) 🎯 MVP

**Goal**: Users can enable/disable debug output via --debug flag and see it documented in help

**Independent Test**: Run `gitree --help` and verify debug flag is present. Run `gitree --debug` and verify debug messages appear. Run `gitree` without debug and verify no debug messages.

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**
> **GATE: Get user approval of tests before proceeding with implementation**

- [ ] T009 [P] [US3] Unit test: Verify --debug flag is parsed correctly in cmd/gitree/root_test.go
- [ ] T010 [P] [US3] Unit test: Verify no debug output when debug=false in cmd/gitree/root_test.go
- [ ] T011 [P] [US3] Integration test: Verify --help includes debug flag in cmd/gitree/root_test.go
- [ ] T012 [P] [US3] Integration test: Verify debug output appears with --debug in cmd/gitree/root_test.go
- [ ] T013 [P] [US3] Integration test: Verify debug output absent without --debug in cmd/gitree/root_test.go

### Implementation for User Story 3

- [x] T014 [US3] Modify scanner struct to store full ScanOptions (not just rootPath) in internal/scanner/scanner.go
- [x] T015 [US3] Update scanner.Scan() function to pass opts to scanner struct in internal/scanner/scanner.go
- [x] T016 [US3] Conditionally disable spinner start/stop when debugFlag is true in cmd/gitree/root.go (3 locations: line ~126, ~155, ~183)

**Checkpoint**: At this point, --debug flag is functional, appears in help, controls debug output, and disables spinner

---

## Phase 4: User Story 2 - Troubleshoot Scanning Behavior (Priority: P2)

**Goal**: Users can see which directories are scanned, skipped, and why, plus timing information

**Independent Test**: Run `gitree --debug` on a directory with nested repos, symlinks, and permission-denied folders. Verify debug output shows directory traversal decisions and timing.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**
> **GATE: Get user approval of tests before proceeding with implementation**

- [ ] T017 [P] [US2] Unit test: Verify debug output for directory entry in internal/scanner/scanner_test.go
- [ ] T018 [P] [US2] Unit test: Verify debug output for repo detection in internal/scanner/scanner_test.go
- [ ] T019 [P] [US2] Unit test: Verify debug output for permission denied in internal/scanner/scanner_test.go
- [ ] T020 [P] [US2] Unit test: Verify debug output for symlink loop in internal/scanner/scanner_test.go
- [ ] T021 [P] [US2] Unit test: Verify debug output for skipping repo contents in internal/scanner/scanner_test.go
- [ ] T022 [P] [US2] Integration test: Verify timing output for status extraction >100ms in internal/gitstatus/status_test.go

### Implementation for User Story 2

- [x] T023 [P] [US2] Add debug output for entering directory in scanner.walkFunc() in internal/scanner/scanner.go
- [x] T024 [P] [US2] Add debug output for permission denied in scanner.walkFunc() in internal/scanner/scanner.go
- [x] T025 [P] [US2] Add debug output for symlink errors in scanner.walkFunc() in internal/scanner/scanner.go
- [x] T026 [P] [US2] Add debug output for already visited (symlink loop) in scanner.walkFunc() in internal/scanner/scanner.go
- [x] T027 [P] [US2] Add debug output for repo detection with type (regular/bare) in scanner.walkFunc() in internal/scanner/scanner.go
- [x] T028 [P] [US2] Add debug output for skipping repo contents in scanner.walkFunc() in internal/scanner/scanner.go
- [x] T029 [US2] Add timing measurement in extractGitStatus() in internal/gitstatus/status.go
- [x] T030 [US2] Add debug output for timing if >100ms in extractGitStatus() in internal/gitstatus/status.go
- [x] T031 [US2] Update extractGitStatus() function signature to accept opts parameter in internal/gitstatus/status.go
- [x] T032 [US2] Update Extract() function to pass opts to extractGitStatus() in internal/gitstatus/status.go

**Checkpoint**: At this point, User Story 2 should show all directory scanning decisions and timing info

---

## Phase 5: User Story 1 - Diagnose Worktree Status Issues (Priority: P1)

**Goal**: Users can see exactly why repositories are marked as needing attention and which files cause non-clean status

**Independent Test**: Run `gitree --debug` on repos with modified files, untracked files, staged changes, non-main branch, and ahead/behind remote. Verify debug output lists specific files and explains status.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**
> **GATE: Get user approval of tests before proceeding with implementation**

- [ ] T033 [P] [US1] Unit test: Verify debug output shows branch and hasChanges in internal/gitstatus/status_test.go
- [ ] T034 [P] [US1] Unit test: Verify debug output lists modified files in internal/gitstatus/status_test.go
- [ ] T035 [P] [US1] Unit test: Verify debug output lists untracked files in internal/gitstatus/status_test.go
- [ ] T036 [P] [US1] Unit test: Verify debug output lists staged files in internal/gitstatus/status_test.go
- [ ] T037 [P] [US1] Unit test: Verify debug output lists deleted files in internal/gitstatus/status_test.go
- [ ] T038 [P] [US1] Unit test: Verify file truncation at 20 per category in internal/gitstatus/status_test.go
- [ ] T039 [P] [US1] Unit test: Verify ahead/behind counts in debug output in internal/gitstatus/status_test.go
- [ ] T040 [P] [US1] Integration test: Verify debug explains why repo marked needing attention in internal/gitstatus/status_test.go

### Implementation for User Story 1

- [x] T041 [US1] Add debug output for status summary (branch, hasChanges, hasRemote, ahead, behind, hasStashes) in extractGitStatus() in internal/gitstatus/status.go
- [x] T042 [US1] Add file categorization logic in extractUncommittedChanges() (Modified, Untracked, Staged, Deleted) in internal/gitstatus/status.go
- [x] T043 [US1] Add printFileList helper function with 20-file truncation in internal/gitstatus/status.go
- [x] T044 [US1] Add debug output for file lists per category in extractUncommittedChanges() in internal/gitstatus/status.go
- [x] T045 [US1] Update extractUncommittedChanges() function signature to accept opts parameter in internal/gitstatus/status.go
- [x] T046 [US1] Update extractGitStatus() to pass opts to extractUncommittedChanges() in internal/gitstatus/status.go

**Checkpoint**: All user stories should now be independently functional - complete debug logging feature

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Integration testing and cross-flag interactions

- [ ] T047 [P] Integration test: Verify --debug works with --all flag in cmd/gitree/root_test.go
- [ ] T048 [P] Integration test: Verify --debug works with --no-color flag in cmd/gitree/root_test.go
- [ ] T049 [P] Integration test: Verify --debug respects NO_COLOR environment variable in cmd/gitree/root_test.go
- [ ] T050 [P] Integration test: Verify stdout tree output unchanged by debug flag in cmd/gitree/root_test.go
- [ ] T051 [P] Integration test: Verify spinner not shown when debug enabled in cmd/gitree/root_test.go
- [ ] T052 Run make lint to ensure code quality
- [ ] T053 Run make test to verify all tests pass
- [ ] T054 Manual validation: Test debug flag on real repositories with various states
- [ ] T055 Verify quickstart.md testing checklist is complete

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) - BLOCKS all user stories
- **User Stories (Phase 3, 4, 5)**: All depend on Foundational phase completion
  - US3 (Debug Flag Control) - P1 priority, MVP foundation
  - US2 (Scanning Behavior) - P2 priority, can run in parallel with US1 after US3
  - US1 (Worktree Status) - P1 priority, can run in parallel with US2 after US3
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 3 (P1) - Debug Flag Control**: Can start after Foundational (Phase 2) - No dependencies on other stories - MUST complete first as it provides the basic debug infrastructure
- **User Story 2 (P2) - Scanning Behavior**: Can start after US3 complete - Independently testable
- **User Story 1 (P1) - Worktree Status**: Can start after US3 complete - Can run in parallel with US2 - Independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- User approval MUST be obtained after tests written, before implementation
- Tests before implementation within each story
- Parallel-marked tasks [P] within each story can run simultaneously
- Story complete before moving to next

### Parallel Opportunities

- **Phase 1**: All 4 setup tasks marked [P] can run in parallel (different files)
- **Phase 3 Tests**: All 5 US3 test tasks marked [P] can run in parallel
- **Phase 4 Tests**: All 6 US2 test tasks marked [P] can run in parallel
- **Phase 4 Implementation**: Tasks T023-T028 marked [P] can run in parallel (all add debug output to different sections of walkFunc)
- **Phase 5 Tests**: All 8 US1 test tasks marked [P] can run in parallel
- **Phase 6**: All 5 integration test tasks marked [P] can run in parallel
- **After US3 complete**: US2 and US1 can be worked on in parallel by different team members

---

## Parallel Example: Phase 1 Setup

```bash
# Launch all setup tasks together:
Task: "Add Debug field to ScanOptions struct in internal/scanner/scanner.go"
Task: "Add Debug field to ExtractOptions struct in internal/gitstatus/status.go"
Task: "Add debugPrintf helper function in internal/scanner/scanner.go"
Task: "Add debugPrintf helper function in internal/gitstatus/status.go"
```

## Parallel Example: User Story 3 Tests

```bash
# Launch all tests for User Story 3 together:
Task: "Unit test: Verify --debug flag is parsed correctly in cmd/gitree/root_test.go"
Task: "Unit test: Verify no debug output when debug=false in cmd/gitree/root_test.go"
Task: "Integration test: Verify --help includes debug flag in cmd/gitree/root_test.go"
Task: "Integration test: Verify debug output appears with --debug in cmd/gitree/root_test.go"
Task: "Integration test: Verify debug output absent without --debug in cmd/gitree/root_test.go"
```

## Parallel Example: After US3 Complete

```bash
# Developer A works on User Story 2:
Task: "Add debug output for entering directory in scanner.walkFunc()"
Task: "Add debug output for permission denied in scanner.walkFunc()"
# ... etc

# Developer B works on User Story 1 in parallel:
Task: "Add debug output for status summary in extractGitStatus()"
Task: "Add file categorization logic in extractUncommittedChanges()"
# ... etc
```

---

## Implementation Strategy

### MVP First (User Story 3 Only)

1. Complete Phase 1: Setup (~5 minutes)
2. Complete Phase 2: Foundational (~10 minutes)
3. Complete Phase 3: User Story 3 - Write tests, get approval, implement (~20 minutes)
4. **STOP and VALIDATE**: Test that --debug flag works, appears in help, controls output
5. Can deploy minimal debug capability if needed

### Full Feature Delivery (All Stories)

1. Complete Setup + Foundational → Foundation ready (~15 minutes)
2. Complete User Story 3 → Test independently → Basic debug flag works (~20 minutes)
3. Complete User Story 2 → Test independently → Scanning debug complete (~25 minutes)
4. Complete User Story 1 → Test independently → Status debug complete (~30 minutes)
5. Complete Polish → Integration tests → Feature complete (~20 minutes)

**Total Estimated Time**: ~1.5-2 hours for complete feature with tests

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (~15 minutes)
2. Complete User Story 3 together (required foundation) (~20 minutes)
3. Once US3 is done:
   - Developer A: User Story 2 (Scanning Behavior)
   - Developer B: User Story 1 (Worktree Status)
4. Stories complete and integrate independently
5. Team does Polish phase together (~20 minutes)

**Total Time with 2 developers**: ~1 hour

---

## Notes

- Per Constitution Principle III (Test-First), tests are mandatory and must be approved before implementation
- [P] tasks = different files or different sections of same file, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (Red-Green-Refactor)
- Get user approval of tests before proceeding with implementation
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All file paths are explicit to enable direct implementation
- Debug feature uses fmt.Fprintln per spec requirement FR-006
- Debug output to stderr per spec requirement FR-005
- Spinner disabled when debug active per spec requirement FR-010

---

## Test Coverage Summary

**Total Tasks**: 55
**Test Tasks**: 23 (42% - reflects TDD approach)
**Implementation Tasks**: 32 (58%)

**User Story 1 Tests**: 8 tests covering status determination, file listing, truncation, ahead/behind
**User Story 2 Tests**: 6 tests covering directory scanning, timing, skip reasons
**User Story 3 Tests**: 5 tests covering flag parsing, help text, output control
**Integration Tests (Polish)**: 5 tests covering flag interactions, output separation

**Test Distribution**:

- Unit tests: ~18 (testing individual functions)
- Integration tests: ~5 (testing flag interactions and end-to-end behavior)
