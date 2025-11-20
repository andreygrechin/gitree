# Tasks: CLI Framework with Command-Line Flags

**Input**: Design documents from `/specs/003-cobra-cli-flags/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: This feature follows Test-First development (Constitution III). All test tasks must be written and approved by the user BEFORE implementation begins.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is a single Go project with the following structure:

- `cmd/gitree/`: CLI entry point
- `internal/cli/`: CLI-specific utilities (new package)
- `internal/models/`: Data models (existing)
- `internal/tree/`: Tree formatting (existing, may be enhanced)
- `Makefile`: Build configuration

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and Cobra framework integration

- [X] T001 Add github.com/spf13/cobra dependency to go.mod via go get
- [X] T002 [P] Create internal/cli package directory structure
- [X] T003 [P] Create cmd/gitree/root.go for Cobra root command definition

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Verify GitStatus struct in internal/models/repository.go has all required fields (Branch, IsDetached, HasRemote, Ahead, Behind, HasStashes, HasChanges, Error)
- [X] T005 Update Makefile build target to include ldflags for version variable injection (-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.buildTime=$(BUILDTIME)')
- [X] T006 Add package-level version variables in cmd/gitree/main.go (version, commit, buildTime with default values "dev", "none", "unknown")

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - View Application Version Information (Priority: P1) 🎯 MVP

**Goal**: Users can view version, build time, and commit hash to verify their installation and report issues

**Independent Test**: Run `gitree --version` and verify output displays semantic version, build timestamp (ISO format), and Git commit hash. Test with both production builds (with ldflags) and development builds (without ldflags).

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T007 [P] [US1] Create cmd/gitree/version_test.go with test for version flag output format (verify all three fields present)
- [X] T008 [P] [US1] Add test in cmd/gitree/version_test.go for version flag with default values (dev build)
- [X] T009 [P] [US1] Add test in cmd/gitree/version_test.go for version flag execution time (<100ms - SC-001)
- [X] T010 [P] [US1] Add test in cmd/gitree/version_test.go for version flag precedence (exits before scanning)

### Implementation for User Story 1

- [X] T011 [US1] Implement version command in cmd/gitree/root.go using Cobra with PreRunE hook to display version info and exit
- [X] T012 [US1] Add version display formatting function in cmd/gitree/root.go (format: "gitree version {version}\n  commit: {commit}\n  built:  {buildTime}")
- [X] T013 [US1] Wire version flag (--version, -v) to version command in cmd/gitree/root.go
- [X] T014 [US1] Update cmd/gitree/main.go to initialize and execute Cobra root command instead of current main logic

**Checkpoint**: At this point, version flag should be fully functional - users can run `gitree --version` and see build information

---

## Phase 4: User Story 2 - Suppress Color Output (Priority: P2)

**Goal**: Users can suppress color formatting for automation, logging, and non-color-supporting terminals

**Independent Test**: Run `gitree --no-color` and verify output contains zero ANSI escape sequences (use regex `\x1b\[[0-9;]*m` to validate). Also test with NO_COLOR environment variable.

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T015 [P] [US2] Create cmd/gitree/color_test.go with ANSI escape sequence detection test (regex validation per contracts/ansi-validation.md)
- [X] T016 [P] [US2] Add test in cmd/gitree/color_test.go for --no-color flag produces zero ANSI codes
- [X] T017 [P] [US2] Add test in cmd/gitree/color_test.go for NO_COLOR environment variable support
- [X] T018 [P] [US2] Add test in cmd/gitree/color_test.go for color output with flag absent (verify ANSI codes present in normal mode)

### Implementation for User Story 2

- [X] T019 [US2] Add --no-color flag definition in cmd/gitree/root.go
- [X] T020 [US2] Implement color suppression logic in cmd/gitree/root.go PersistentPreRun to set color.NoColor = true when flag or NO_COLOR env var present
- [X] T021 [US2] Verify all output mechanisms in internal/tree/formatter.go respect fatih/color's NoColor setting (audit existing color usage)
- [X] T022 [US2] Test color suppression with actual repository scanning output to stderr and stdout

**Checkpoint**: At this point, --no-color flag should work globally - all output should be plain text without ANSI codes

---

## Phase 5: User Story 3 - Show All Repositories Including Clean Ones (Priority: P2)

**Goal**: Users can optionally view all repositories (clean and changed) while the default shows only repositories needing attention

**Independent Test**: Create test directory with both clean repos (on main/master, no changes, synchronized) and repos needing attention (feature branches, uncommitted changes, ahead/behind). Run without --all flag - verify only changed repos shown. Run with --all flag - verify all repos shown. Verify tree structure preserved in both cases.

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T023 [P] [US3] Create internal/cli/filter_test.go with IsClean() unit tests covering all 9 clean conditions from FR-008
- [X] T024 [P] [US3] Add tests in internal/cli/filter_test.go for IsClean() fail-safe behavior (nil status, error status)
- [X] T025 [P] [US3] Add tests in internal/cli/filter_test.go for each "needs attention" condition from FR-009
- [X] T026 [P] [US3] Create FilterRepositories() unit tests in internal/cli/filter_test.go (empty input, all clean, all dirty, mixed, order preservation, non-modification)
- [X] T027 [P] [US3] Add integration test in cmd/gitree/filter_integration_test.go for default filtering behavior (show only repos needing attention)
- [X] T028 [P] [US3] Add integration test in cmd/gitree/filter_integration_test.go for --all flag behavior (show all repos)
- [X] T029 [P] [US3] Add integration test in cmd/gitree/filter_integration_test.go for tree structure preservation with filtering
- [X] T030 [P] [US3] Add edge case test in cmd/gitree/filter_integration_test.go for all repos clean scenario

### Implementation for User Story 3

- [X] T031 [P] [US3] Create internal/cli/filter.go with FilterOptions struct (ShowAll bool field)
- [X] T032 [US3] Implement IsClean() function in internal/cli/filter.go per FR-008 (9 conditions: main/master branch, no changes, no stashes, has remote, not ahead/behind, not detached)
- [X] T033 [US3] Implement FilterRepositories() function in internal/cli/filter.go (applies ShowAll logic, maintains order, fail-safe for errors)
- [X] T034 [US3] Add --all flag definition in cmd/gitree/root.go with help text explaining default filtering behavior
- [X] T035 [US3] Integrate FilterRepositories() into main scan-status-filter-build-format pipeline in cmd/gitree/root.go (call after status extraction, before tree building)
- [X] T036 [US3] Update default command execution in cmd/gitree/root.go to show only repos needing attention (ShowAll=false by default)
- [X] T037 [US3] Handle edge case in cmd/gitree/root.go where all repos are clean in default mode (display helpful message)
- [X] T038 [US3] Update tree building in cmd/gitree/root.go to handle filtered repository lists while preserving directory hierarchy

**Checkpoint**: At this point, all three user stories are complete - default shows only repos needing attention, --all shows everything, --no-color suppresses colors, --version shows build info

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T039 [P] Add comprehensive help text in cmd/gitree/root.go per FR-014 (explain all flags, default behavior, clean state definition)
- [X] T040 [P] Add usage examples in cmd/gitree/root.go help text (default scan, --version, --no-color, --all, combined flags)
- [X] T041 [P] Update CLAUDE.md project documentation with new flags and filtering behavior
- [X] T042 Verify all flag combinations work correctly (version precedence, no-color + all, all combinations from contracts/cli-interface.md)
- [X] T043 Run make lint and fix any linting issues introduced by new code
- [X] T044 Run make test to verify all tests pass
- [X] T045 Validate against quickstart.md scenarios (version display, color suppression, filtering, all flag)
- [X] T046 [P] Performance verification: version flag <100ms (SC-001), CLI parsing <10ms
- [X] T047 [P] Validate filtering accuracy: 100% correct filtering in test scenarios (SC-003)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User Story 1 (Version): Can start after Foundational - No dependencies on other stories
  - User Story 2 (No-Color): Can start after Foundational - Independent of other stories
  - User Story 3 (Filtering/All): Can start after Foundational - Independent of other stories
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Version**: Can start after Foundational (Phase 2) - Completely independent
- **User Story 2 (P2) - No-Color**: Can start after Foundational (Phase 2) - Completely independent
- **User Story 3 (P2) - Filtering/All**: Can start after Foundational (Phase 2) - Completely independent

### Within Each User Story

- Tests MUST be written and FAIL before implementation (Constitution III)
- Test creation tasks can run in parallel (marked [P])
- Implementation must follow test approval
- Core functionality before integration
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: All 3 tasks can run in parallel

**Phase 2 (Foundational)**: T004 (verify fields) can run parallel with T005-T006 (Makefile/version vars)

**Phase 3 Tests (US1)**: All 4 test creation tasks (T007-T010) can run in parallel

**Phase 3 Implementation (US1)**: After tests pass, T011-T012 can be done together, then T013-T014 sequentially

**Phase 4 Tests (US2)**: All 4 test creation tasks (T015-T018) can run in parallel

**Phase 4 Implementation (US2)**: T019-T020 together, then T021-T022

**Phase 5 Tests (US3)**: All 8 test creation tasks (T023-T030) can run in parallel (different test files)

**Phase 5 Implementation (US3)**: T031-T032 in parallel (different aspects of filter.go), then sequential for integration

**Phase 6 (Polish)**: Tasks T039-T041 can run in parallel, then T042-T047 have some dependencies on earlier completions

**Multiple User Stories**: After Foundational (Phase 2), different developers can work on US1, US2, and US3 completely in parallel

---

## Parallel Example: User Story 3 Tests

```bash
# Launch all test creation tasks for User Story 3 together:
Task: "Create internal/cli/filter_test.go with IsClean() unit tests"
Task: "Add IsClean() fail-safe tests in internal/cli/filter_test.go"
Task: "Add needs attention condition tests in internal/cli/filter_test.go"
Task: "Create FilterRepositories() unit tests in internal/cli/filter_test.go"
Task: "Add integration test for default filtering in cmd/gitree/filter_integration_test.go"
Task: "Add integration test for --all flag in cmd/gitree/filter_integration_test.go"
Task: "Add tree preservation test in cmd/gitree/filter_integration_test.go"
Task: "Add all-clean edge case test in cmd/gitree/filter_integration_test.go"
```

---

## Parallel Example: Multiple User Stories

```bash
# After Foundational phase completes, assign developers:

Developer A: Complete User Story 1 (Version Info) - T007 through T014
Developer B: Complete User Story 2 (Color Suppression) - T015 through T022
Developer C: Complete User Story 3 (Filtering/All Flag) - T023 through T038

# All three can proceed completely independently and integrate at the end
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T006) - CRITICAL blocker
3. Complete Phase 3: User Story 1 (T007-T014)
4. **STOP and VALIDATE**: Test version flag independently
5. **APPROVAL GATE**: Get user approval of tests before implementation
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (Version info works!)
3. Add User Story 2 → Test independently → Deploy/Demo (Color suppression works!)
4. Add User Story 3 → Test independently → Deploy/Demo (Smart filtering works!)
5. Polish phase → Final validation → Production ready

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With 3 developers:

1. Team completes Setup (Phase 1) + Foundational (Phase 2) together
2. Once Foundational is done:
   - Developer A: User Story 1 (Version) - T007-T014
   - Developer B: User Story 2 (No-Color) - T015-T022
   - Developer C: User Story 3 (Filtering) - T023-T038
3. Stories complete independently
4. Team integrates and completes Polish phase together

---

## Test-First Workflow (Constitution III) ⚠️

**MANDATORY**: This project follows strict Test-First development:

1. **Write Tests First**: All test tasks (T007-T010, T015-T018, T023-T030) MUST be written before implementation
2. **Verify Tests Fail**: Run tests and confirm they fail (proving they test something real)
3. **Get User Approval**: Submit test suite to user for approval before proceeding
4. **Implement**: Only after approval, proceed with implementation tasks
5. **Red-Green-Refactor**:
   - Red: Tests fail (written first)
   - Green: Implementation makes tests pass
   - Refactor: Clean up code while keeping tests green

**DO NOT** skip test approval. **DO NOT** write implementation before tests.

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label = maps task to specific user story (US1, US2, US3)
- Each user story is independently completable and testable
- Default behavior is to show only repos needing attention (FR-006)
- --all flag reverts to showing all repos (FR-007)
- Tests MUST be approved before implementation begins
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Existing GitStatus fields: Branch, IsDetached, HasRemote, Ahead, Behind, HasStashes, HasChanges, Error

---

## Summary Statistics

- **Total Tasks**: 47
- **Setup Phase**: 3 tasks
- **Foundational Phase**: 3 tasks (BLOCKS all stories)
- **User Story 1 (Version)**: 8 tasks (4 tests + 4 implementation)
- **User Story 2 (No-Color)**: 8 tasks (4 tests + 4 implementation)
- **User Story 3 (Filtering/All)**: 16 tasks (8 tests + 8 implementation)
- **Polish Phase**: 9 tasks
- **Parallel Opportunities**:
  - Phase 1: 3 tasks in parallel
  - US1 Tests: 4 tasks in parallel
  - US2 Tests: 4 tasks in parallel
  - US3 Tests: 8 tasks in parallel
  - All 3 user stories can proceed in parallel after Foundational phase
- **MVP Scope**: Phases 1-2-3 (Setup + Foundational + User Story 1) = 14 tasks
- **Full Feature**: All 47 tasks

---

## Success Criteria Mapping

- **SC-001**: Version <100ms → T009, T046
- **SC-002**: Zero ANSI codes with --no-color → T015, T016
- **SC-003**: 100% filtering accuracy → T026, T047
- **SC-004**: Combined flags work correctly → T042
- **SC-005**: Version exits immediately → T010
- **SC-006**: Filtering reduces clutter ≥50% → T027, T047
- **SC-007**: Output parseable in CI/CD → T016, T022
- **SC-008**: Help text comprehension 90% → T039
- **SC-009**: --all shows all repos → T028

All success criteria are testable and mapped to specific tasks.
