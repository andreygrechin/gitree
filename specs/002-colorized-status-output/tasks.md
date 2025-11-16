# Tasks: Colorized Status Output

**Input**: Design documents from `/specs/002-colorized-status-output/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: This project follows TDD (Test-First) as mandated by the constitution. All test tasks MUST be completed before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Project uses single project structure:
- Source: `internal/models/` for model modifications
- Tests: `internal/models/` for test files
- Paths are absolute from repository root: `/Users/andrey/repos/gitree/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Dependency upgrade and project preparation

- [X] T001 Upgrade fatih/color dependency to v1.18.0 using `go get github.com/fatih/color@v1.18.0`
- [X] T002 Run `go mod tidy` to clean up dependencies
- [X] T003 Verify fatih/color v1.18.0 in go.mod using `grep fatih/color go.mod`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core color infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Add fatih/color import to internal/models/repository.go
- [X] T005 [P] Create package-level grayColor function using color.New(color.FgHiBlack).SprintFunc() in internal/models/repository.go
- [X] T006 [P] Create package-level yellowColor function using color.New(color.FgYellow).SprintFunc() in internal/models/repository.go
- [X] T007 [P] Create package-level greenColor function using color.New(color.FgGreen).SprintFunc() in internal/models/repository.go
- [X] T008 [P] Create package-level redColor function using color.New(color.FgRed).SprintFunc() in internal/models/repository.go

**Checkpoint**: Color infrastructure ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Distinguish Repository Metadata from Names (Priority: P1) 🎯 MVP

**Goal**: Implement double square brackets with gray color to visually separate metadata from repository names

**Independent Test**: Run gitree and verify that status metadata appears in double square brackets `[[ ]]` with gray-colored bracket characters

### Tests for User Story 1 (TDD - REQUIRED)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T009 [P] [US1] Write test TestGitStatusFormatDoubleBrackets in internal/models/repository_test.go to verify output uses `[[` and `]]` instead of `[` and `]`
- [X] T010 [P] [US1] Write test TestGitStatusFormatGrayBrackets in internal/models/repository_test.go to verify bracket characters contain ANSI code `\033[90m` (gray) when colors enabled
- [X] T011 [P] [US1] Write test TestGitStatusFormatDoubleBracketsNoColor in internal/models/repository_test.go to verify `[[` `]]` present but no ANSI codes when color.NoColor = true
- [X] T012 [US1] Run tests to verify they FAIL (red phase): `go test ./internal/models/ -run TestGitStatusFormat`

### Implementation for User Story 1

- [X] T013 [US1] Modify GitStatus.Format() in internal/models/repository.go to change opening bracket from `[` to grayColor("[[")
- [X] T014 [US1] Modify GitStatus.Format() in internal/models/repository.go to change closing bracket from `]` to grayColor("]]")
- [X] T015 [US1] Update GitStatus.Format() in internal/models/repository.go to add space after opening bracket and before closing bracket for readability
- [X] T016 [US1] Run tests to verify they PASS (green phase): `go test ./internal/models/ -run TestGitStatusFormat`
- [X] T017 [US1] Run all existing tests to ensure no regressions: `make test`

**Checkpoint**: At this point, User Story 1 should be fully functional - metadata visually separated with double gray brackets

---

## Phase 4: User Story 2 - Identify Branch Type at a Glance (Priority: P1)

**Goal**: Implement color-coded branch names (gray for main/master, yellow for feature branches)

**Independent Test**: Run gitree in directories with repositories on different branches and verify that main/master appear in gray and other branches in yellow

### Tests for User Story 2 (TDD - REQUIRED)

- [X] T018 [P] [US2] Write test TestGitStatusFormatMainBranchGray in internal/models/repository_test.go to verify "main" branch contains ANSI code `\033[90m` (gray)
- [X] T019 [P] [US2] Write test TestGitStatusFormatMasterBranchGray in internal/models/repository_test.go to verify "master" branch contains ANSI code `\033[90m` (gray)
- [X] T020 [P] [US2] Write test TestGitStatusFormatFeatureBranchYellow in internal/models/repository_test.go to verify non-main/master branches contain ANSI code `\033[33m` (yellow)
- [X] T021 [P] [US2] Write test TestGitStatusFormatDetachedYellow in internal/models/repository_test.go to verify "DETACHED" contains ANSI code `\033[33m` (yellow)
- [X] T022 [US2] Run tests to verify they FAIL (red phase): `go test ./internal/models/ -run TestGitStatusFormat`

### Implementation for User Story 2

- [X] T023 [US2] Modify GitStatus.Format() in internal/models/repository.go to add branch color logic: if Branch == "main" or "master" use grayColor(g.Branch), else use yellowColor(g.Branch)
- [X] T024 [US2] Run tests to verify they PASS (green phase): `go test ./internal/models/ -run TestGitStatusFormat`
- [X] T025 [US2] Run all existing tests to ensure no regressions: `make test`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - metadata in double gray brackets with color-coded branch names

---

## Phase 5: User Story 3 - Assess Synchronization Status with Visual Cues (Priority: P1)

**Goal**: Implement color-coded ahead/behind indicators (green for ahead, red for behind)

**Independent Test**: Create repositories with various ahead/behind states and verify that ahead counts appear in green and behind counts in red

### Tests for User Story 3 (TDD - REQUIRED)

- [X] T026 [P] [US3] Write test TestGitStatusFormatAheadGreen in internal/models/repository_test.go to verify ahead indicator `↑N` contains ANSI code `\033[32m` (green)
- [X] T027 [P] [US3] Write test TestGitStatusFormatBehindRed in internal/models/repository_test.go to verify behind indicator `↓N` contains ANSI code `\033[31m` (red)
- [X] T028 [P] [US3] Write test TestGitStatusFormatAheadAndBehind in internal/models/repository_test.go to verify ahead is green and behind is red when both present
- [X] T029 [P] [US3] Write test TestGitStatusFormatNoRemoteGray in internal/models/repository_test.go to verify no-remote indicator `○` contains ANSI code `\033[90m` (gray)
- [X] T030 [US3] Run tests to verify they FAIL (red phase): `go test ./internal/models/ -run TestGitStatusFormat`

### Implementation for User Story 3

- [X] T031 [US3] Modify GitStatus.Format() in internal/models/repository.go to wrap ahead indicator in greenColor: greenColor(fmt.Sprintf("↑%d", g.Ahead))
- [X] T032 [US3] Modify GitStatus.Format() in internal/models/repository.go to wrap behind indicator in redColor: redColor(fmt.Sprintf("↓%d", g.Behind))
- [X] T033 [US3] Modify GitStatus.Format() in internal/models/repository.go to wrap no-remote indicator in grayColor: grayColor("○")
- [X] T034 [US3] Run tests to verify they PASS (green phase): `go test ./internal/models/ -run TestGitStatusFormat`
- [X] T035 [US3] Run all existing tests to ensure no regressions: `make test`

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work - color-coded branch types and sync status

---

## Phase 6: User Story 4 - Identify Repositories Requiring Attention (Priority: P2)

**Goal**: Implement red color for attention indicators (stashes, uncommitted changes, bare repositories)

**Independent Test**: Create repositories with stashes, uncommitted changes, and bare repositories, then verify red color is applied

### Tests for User Story 4 (TDD - REQUIRED)

- [X] T036 [P] [US4] Write test TestGitStatusFormatStashRed in internal/models/repository_test.go to verify stash indicator `$` contains ANSI code `\033[31m` (red)
- [X] T037 [P] [US4] Write test TestGitStatusFormatChangesRed in internal/models/repository_test.go to verify uncommitted changes indicator `*` contains ANSI code `\033[31m` (red)
- [X] T038 [P] [US4] Write test TestGitStatusFormatMultipleAttentionIndicators in internal/models/repository_test.go to verify both `$` and `*` are red when present together
- [X] T039 [US4] Run tests to verify they FAIL (red phase): `go test ./internal/models/ -run TestGitStatusFormat`

### Implementation for User Story 4

- [X] T040 [US4] Modify GitStatus.Format() in internal/models/repository.go to wrap stash indicator in redColor: redColor("$")
- [X] T041 [US4] Modify GitStatus.Format() in internal/models/repository.go to wrap uncommitted changes indicator in redColor: redColor("*")
- [X] T042 [US4] Verify "bare" indicator handling in internal/models/repository.go (may be in Repository.Format or elsewhere) and apply redColor("bare") if needed
- [X] T043 [US4] Run tests to verify they PASS (green phase): `go test ./internal/models/ -run TestGitStatusFormat`
- [X] T044 [US4] Run all existing tests to ensure no regressions: `make test`

**Checkpoint**: All user stories should now be independently functional with complete colorization

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, edge cases, and cross-cutting improvements

- [X] T045 [P] Write test TestGitStatusFormatSeparatorGray in internal/models/repository_test.go to verify separator `|` between parts contains ANSI code `\033[90m` (gray)
- [X] T046 [P] Implement separator logic in GitStatus.Format() in internal/models/repository.go: use grayColor("|") between branch name and status indicators only when status indicators are present
- [X] T047 Build binary: `make build`
- [X] T048 Manual test: Run `./bin/gitree` in terminal and verify colored output
- [X] T049 Manual test: Run `./bin/gitree | cat` and verify colors are stripped (no ANSI codes in piped output)
- [X] T050 Manual test: Run `NO_COLOR=1 ./bin/gitree` and verify colors are disabled
- [X] T051 Manual test: Create test repos with various states (per quickstart.md Step 11) and verify all color scenarios
- [X] T052 Run code quality checks: `make lint`
- [X] T053 Fix any linting issues reported
- [X] T054 Run full test suite: `make test`
- [X] T055 Verify all tests pass with 100% success rate

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 stories first: US1, US2, US3; then P2: US4)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Builds on US1's double bracket format but independently testable
- **User Story 3 (P1)**: Can start after Foundational (Phase 2) - Builds on US1-2 but independently testable
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Builds on US1-3 but independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD - constitutional requirement)
- Tests marked [P] can run in parallel (different test functions)
- Implementation tasks run sequentially within a story
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T001-T003 can run sequentially (dependencies between them)
- **Phase 2**: T005-T008 can run in parallel (all create independent color functions)
- **Phase 3 Tests**: T009-T011 can run in parallel (different test functions)
- **Phase 4 Tests**: T018-T021 can run in parallel (different test functions)
- **Phase 5 Tests**: T026-T029 can run in parallel (different test functions)
- **Phase 6 Tests**: T036-T038 can run in parallel (different test functions)
- **Phase 7**: T045-T046 can run together, T047-T055 are sequential
- **User Stories**: Once Phase 2 completes, all user stories (Phase 3-6) can start in parallel by different developers

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
claude: "Write test TestGitStatusFormatDoubleBrackets in internal/models/repository_test.go"
claude: "Write test TestGitStatusFormatGrayBrackets in internal/models/repository_test.go"
claude: "Write test TestGitStatusFormatDoubleBracketsNoColor in internal/models/repository_test.go"
```

---

## Parallel Example: User Story 3

```bash
# Launch all tests for User Story 3 together:
claude: "Write test TestGitStatusFormatAheadGreen in internal/models/repository_test.go"
claude: "Write test TestGitStatusFormatBehindRed in internal/models/repository_test.go"
claude: "Write test TestGitStatusFormatAheadAndBehind in internal/models/repository_test.go"
claude: "Write test TestGitStatusFormatNoRemoteGray in internal/models/repository_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1-3 Only - All P1)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (double brackets)
4. **CHECKPOINT**: Test US1 independently
5. Complete Phase 4: User Story 2 (branch colors)
6. **CHECKPOINT**: Test US1+US2 independently
7. Complete Phase 5: User Story 3 (sync status colors)
8. **CHECKPOINT**: Test US1+US2+US3 independently
9. **STOP and VALIDATE**: Core P1 features complete and testable
10. Deploy/demo if ready

### Full Feature (Including P2)

1. Complete MVP (Phases 1-5)
2. Complete Phase 6: User Story 4 (attention indicators)
3. **CHECKPOINT**: Test all stories
4. Complete Phase 7: Polish
5. Final validation and deployment

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (Phases 1-2)
2. Once Foundational is done:
   - Developer A: User Story 1 (Phase 3) - T009-T017
   - Developer B: User Story 2 (Phase 4) - T018-T025
   - Developer C: User Story 3 (Phase 5) - T026-T035
3. After P1 stories complete:
   - Any developer: User Story 4 (Phase 6) - T036-T044
4. Team completes Polish together (Phase 7)

---

## Notes

- [P] tasks = different files or independent test functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD workflow is MANDATORY: Write tests → Verify FAIL → Implement → Verify PASS
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All changes isolated to internal/models/repository.go and internal/models/repository_test.go
- No changes needed to scanner, gitstatus, tree, or cmd packages
