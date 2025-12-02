# Implementation Plan: Debug Logging

**Branch**: `004-debug-logging` | **Date**: 2025-11-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-debug-logging/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add a `--debug` CLI flag to enable diagnostic output explaining git status determination and directory scanning behavior. Debug output will list specific files causing non-clean worktree status, timing information for slow operations, and scanning decisions. Output uses simple fmt.Fprintln to stderr, respects --no-color flag, and disables the progress spinner when active.

## Technical Context

**Language/Version**: Go 1.25.5
**Primary Dependencies**:

- `github.com/spf13/cobra v1.10.1` (CLI flag parsing)
- `github.com/go-git/go-git/v5 v5.16.3` (git operations)
- `github.com/fatih/color v1.18.0` (color output)
- `github.com/briandowns/spinner v1.23.2` (progress indication)

**Storage**: N/A (reads git repositories from filesystem)
**Testing**: `go test` with `github.com/stretchr/testify v1.11.1`
**Target Platform**: Cross-platform CLI (darwin, linux, windows)
**Project Type**: Single Go CLI application
**Performance Goals**: Debug output must not add >5% overhead when disabled, <50ms per repository when enabled
**Constraints**:

- Must use fmt.Fprintln/fmt.Fprint only (no logging frameworks per spec FR-006)
- Debug output to stderr only (FR-005)
- Must respect --no-color flag (FR-007)
- Must disable spinner when debug active (FR-010)
- No performance degradation when disabled (Constraint from spec)

**Scale/Scope**: Single CLI tool with ~15 source files across 5 packages

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Principle I: Library-First

**Status**: COMPLIANT
**Rationale**: Debug logging will be implemented as functionality within existing libraries (`internal/scanner`, `internal/gitstatus`). The feature enhances existing modular packages rather than creating organizational-only code.

### ✅ Principle II: CLI Interface

**Status**: COMPLIANT
**Rationale**: Feature exposes debug capability via CLI flag (`--debug`) following existing CLI conventions. Output to stderr maintains text protocol separation (data to stdout, diagnostics to stderr).

### ✅ Principle III: Test-First (NON-NEGOTIABLE)

**Status**: COMPLIANT
**Rationale**: Will write tests before implementation in Phase 2 (tasks.md). Tests will verify debug output appears when flag enabled, is absent when disabled, respects --no-color, disables spinner, and includes expected diagnostic details.

### ✅ Principle IV: Observability

**Status**: COMPLIANT
**Rationale**: This feature IS observability. It adds debuggability through text I/O design (stderr output) and diagnostic information accessible without special tools. Aligns perfectly with constitutional observability mandate.

### ✅ Principle V: Simplicity

**Status**: COMPLIANT
**Rationale**: Uses simplest approach: boolean flag, fmt.Fprintln, no frameworks. Avoids complexity of structured logging, multiple verbosity levels, or configuration files. Follows YAGNI - only on/off debug needed.

**GATE RESULT**: ✅ PASS - All constitutional principles satisfied. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/004-debug-logging/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Go CLI project structure (existing)
cmd/gitree/              # CLI entry point
├── main.go             # Application bootstrap
└── root.go             # Cobra root command (add --debug flag here)

internal/
├── scanner/            # Directory scanning (add debug output)
│   ├── scanner.go
│   └── scanner_test.go
├── gitstatus/          # Git status extraction (add debug output)
│   ├── status.go
│   └── status_test.go
├── tree/               # Tree formatting (no changes needed)
│   ├── formatter.go
│   └── formatter_test.go
├── cli/                # CLI utilities (add debug context passing)
│   ├── filter.go
│   └── filter_test.go
└── models/             # Data models (add debug flag to context)
    ├── repository.go
    └── repository_test.go

tests/                  # Test organization
├── integration/        # Integration tests for debug flag
├── unit/              # Unit tests (in *_test.go files)
└── contract/          # Contract tests (if needed)
```

**Structure Decision**: Gitree uses standard Go single-project layout with `cmd/` for entry points and `internal/` for implementation packages. Debug flag will be added to existing CLI parsing in `cmd/gitree/root.go`, debug output will be added to scanning logic in `internal/scanner/scanner.go` and git status logic in `internal/gitstatus/status.go`. A debug context or configuration struct will be passed through the call chain to control output.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations - table left empty.
