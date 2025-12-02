# Implementation Plan: Colorized Status Output

**Branch**: `002-colorized-status-output` | **Date**: 2025-11-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-colorized-status-output/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add colorization to the Git status output in gitree to improve visual hierarchy and readability. The feature modifies the status display format from single brackets `[branch]` to double brackets `[[branch]]` with ANSI color codes applied to different status components: gray for main/master branches and brackets, yellow for feature branches, green for ahead commits, red for behind commits/stashes/uncommitted changes/bare indicator. Terminal capability detection ensures colors are only applied when supported and disabled for non-TTY output.

## Technical Context

**Language/Version**: Go 1.25.5
**Primary Dependencies**: fatih/color (already present via go-git dependencies) or golang.org/x/term for color support
**Storage**: N/A
**Testing**: Go testing framework (existing test structure in `internal/*/`)
**Target Platform**: CLI tool for Linux, macOS, Windows
**Project Type**: Single project (CLI tool)
**Performance Goals**: No performance degradation from baseline - colorization should add <1ms overhead per repository
**Constraints**: Must maintain backward compatibility with existing output format (double brackets instead of single); must gracefully degrade when colors not supported
**Scale/Scope**: Affects output formatting only - single package modification (internal/models for GitStatus.Format)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Pre-Research)

✅ **Library-First**: This feature modifies the existing `internal/models` library to add colorization capability to the GitStatus.Format() method. The colorization logic will be self-contained within the models package.

✅ **CLI Interface**: The gitree CLI already follows text protocol (stdout for output, stderr for errors). This feature enhances the output formatting while maintaining the same protocol.

✅ **Test-First (NON-NEGOTIABLE)**: Tests will be written before implementation. User approval of tests will be obtained before proceeding with implementation. Tests will verify color codes in output and terminal detection logic.

✅ **Observability**: Colorization does not affect logging or debugging. The feature maintains clean separation - structured output to stdout, errors to stderr. Color codes are stripped for non-TTY output.

✅ **Simplicity**: The implementation is straightforward - add ANSI color codes to the GitStatus.Format() method with terminal capability detection. No new abstractions or complex patterns needed.

**Status**: ✅ PASSED - No constitutional violations

### Post-Design Check

✅ **Library-First**: Design maintains library-first approach. The `internal/models` package remains self-contained with colorization logic encapsulated in GitStatus.Format(). No organizational-only abstractions introduced.

✅ **CLI Interface**: No changes to CLI interface. The tool still follows text in/out protocol. Color codes are automatically stripped for non-TTY output, maintaining clean stdout for data.

✅ **Test-First**: Design includes comprehensive test strategy (see quickstart.md). Tests will cover color code presence/absence, terminal detection, and all status variations. TDD workflow will be enforced during implementation.

✅ **Observability**: No impact on observability. Colorization is purely visual enhancement. Structured logging to stderr remains unchanged. Color codes do not affect debugging or log analysis.

✅ **Simplicity**: Design is simple - modify single method (GitStatus.Format), add package-level color functions, leverage existing fatih/color library. No new abstractions, patterns, or architectural changes. YAGNI principle maintained.

**Status**: ✅ PASSED - No constitutional violations. Design aligns with all core principles.

## Project Structure

### Documentation (this feature)

```text
specs/002-colorized-status-output/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (may be empty - no external contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── models/
│   ├── repository.go        # GitStatus.Format() modification
│   └── repository_test.go   # Tests for colorized output
├── gitstatus/
│   ├── extractor.go         # No changes needed
│   └── extractor_test.go    # No changes needed
├── scanner/
│   ├── scanner.go           # No changes needed
│   └── scanner_test.go      # No changes needed
└── tree/
    ├── formatter.go         # No changes needed
    └── formatter_test.go    # No changes needed

cmd/gitree/
└── main.go                  # No changes needed
```

**Structure Decision**: Single project structure maintained. Changes are isolated to the `internal/models` package, specifically the `GitStatus.Format()` method. No new packages or structural changes required. The existing architecture already separates concerns properly - models handle data representation, tree formatter handles tree structure, and models are responsible for their own string formatting.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Not applicable - no constitutional violations detected.
