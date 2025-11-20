# Implementation Plan: CLI Framework with Command-Line Flags

**Branch**: `003-cobra-cli-flags` | **Date**: 2025-11-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-cobra-cli-flags/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add Cobra CLI framework to gitree with three flags: `--version` (display version info with build-time metadata), `--no-color` (suppress color output), and `--all` (show all repositories). The default behavior will filter to show only repositories needing attention (uncommitted changes, non-main branches, ahead/behind, stashes, no remote). The `--all` flag reverts to showing all repositories including clean ones.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**:

- `github.com/spf13/cobra` v1.10.1 (existing)
- `github.com/go-git/go-git/v5` v5.16.3 (existing)
- `github.com/briandowns/spinner` v1.23.2 (existing)
- `github.com/fatih/color` v1.18.0 (existing - for color suppression)

**Storage**: N/A (CLI tool reads git repositories from filesystem)
**Testing**: Go standard testing (`go test`), existing test suite uses testify/assert
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single project (CLI application)
**Performance Goals**:

- Version flag: exit in <100ms
- CLI parsing overhead: <10ms
- Repository scanning performance unchanged from current implementation

**Constraints**:

- Must maintain backward compatibility for output format (tree structure)
- Color suppression must work across all existing output mechanisms
- Build-time variable injection via Makefile (existing pattern)

**Scale/Scope**:

- Single CLI binary
- 3 new flags (version, no-color, all)
- Refactor main.go to use Cobra command structure
- Add filtering logic for default "changed repos only" behavior

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Library-First ✅ PASS

**Status**: COMPLIANT

**Analysis**: This feature enhances the existing CLI tool without introducing organizational-only libraries. The Cobra framework integration will be contained within `cmd/gitree/` and any shared CLI utilities will have independent utility (flag parsing, version formatting, color suppression). The filtering logic will be added to existing libraries (`internal/tree/` or `internal/scanner/`) where it provides standalone functionality for filtering repository lists.

**Action Required**: None. Standard library-first development applies.

---

### II. CLI Interface ✅ PASS

**Status**: COMPLIANT

**Analysis**: This feature IS the CLI interface enhancement. All functionality (version display, color suppression, repository filtering) follows the text in/out protocol:

- Input: Command-line flags
- Output: Tree structure to stdout
- Errors: To stderr
- Supports both human-readable (tree format) and machine-readable output (current implementation)

**Action Required**: None. Feature directly enhances CLI compliance.

---

### III. Test-First (NON-NEGOTIABLE) ⚠️ REQUIRES ATTENTION

**Status**: MANDATORY TDD WORKFLOW

**Analysis**: This is a NON-NEGOTIABLE principle. Tests MUST be written before implementation and approved by user before coding begins.

**Action Required**:

1. After design phase (Phase 1), generate comprehensive test specifications
2. Submit test specifications to user for approval
3. Only proceed with implementation after user approves tests
4. Follow Red-Green-Refactor cycle strictly

**Test Areas to Cover**:

- CLI flag parsing (version, no-color, all combinations)
- Version flag output format and build-time variable injection
- Color suppression across all output types
- Repository filtering logic (clean vs. needs attention)
- Tree structure preservation with filtering
- Edge cases from spec (detached HEAD, no remotes, timeout failures)

---

### IV. Observability ✅ PASS

**Status**: COMPLIANT

**Analysis**: Current implementation already follows observability principles:

- Spinner and status messages go to stderr
- Data output goes to stdout
- Error messages go to stderr
This feature maintains these patterns and adds version information for debugging.

**Action Required**: None. Maintain existing observability patterns.

---

### V. Simplicity ✅ PASS

**Status**: COMPLIANT

**Analysis**: This feature adds necessary CLI functionality without unnecessary complexity:

- Cobra is a standard Go CLI framework (not over-engineering)
- Three focused flags with clear purposes
- Filtering logic reuses existing git status detection
- No new abstractions or patterns introduced

**Complexity Justification**: Cobra framework adds dependency but provides standard CLI patterns (help text, flag parsing, subcommand structure for future). This is justified by:

1. Industry-standard approach for Go CLIs
2. Better user experience (automatic help, flag validation)
3. Maintainability over manual flag parsing

**Action Required**: None. Complexity is justified and minimal.

---

### Summary

**Overall Status**: ✅ PASS with mandatory TDD workflow requirement

**Blockers**: None, but Test-First principle MUST be followed after design phase.

**Next Steps**:

1. Proceed to Phase 0 (Research)
2. Complete Phase 1 (Design)
3. Generate test specifications
4. **STOP and get user approval of tests**
5. Only then proceed to implementation

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/gitree/
├── main.go              # Main entry point (to be refactored to use Cobra)
└── root.go              # NEW: Cobra root command definition
└── color_test.go
└── version_test.go

internal/
├── cli/                 # NEW: CLI-specific utilities
│   ├── filter.go        # Repository filtering logic
│   └── filter_test.go   # Filter logic tests
├── gitstatus/
│   ├── status.go        # Existing: Git status extraction
│   └── status_test.go
├── models/
│   ├── repository.go    # Existing: Repository data model
│   └── repository_test.go
├── scanner/
│   ├── scanner.go       # Existing: Directory scanning
│   └── scanner_test.go
└── tree/
    ├── formatter.go     # Existing: Tree formatting (to be enhanced for color suppression)
    └── formatter_test.go

Makefile                 # To be updated: Add version variables for build-time injection
go.mod                   # To be updated: Add Cobra dependency
go.sum                   # Auto-generated
```

**Structure Decision**: Single project structure (Go CLI tool). New code organized under:

- `cmd/gitree/`: Cobra command definitions and CLI entry point
- `internal/cli/`: CLI-specific utilities (version, filtering) as standalone libraries
- Existing packages enhanced: `internal/tree/` for color suppression support

This structure follows the Library-First principle by keeping version handling and filtering logic in independently testable `internal/cli/` packages.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitutional violations. All complexity is justified within Simplicity principle (Cobra is industry-standard, minimal new abstractions).
