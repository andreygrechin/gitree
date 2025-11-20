# Research: CLI Framework with Command-Line Flags

**Feature**: 003-cobra-cli-flags
**Date**: 2025-11-17
**Status**: Complete

## Overview

This document resolves technical unknowns identified during planning, specifically the choice of CLI framework and best practices for implementing the three core flags: version, no-color, and all.

## Research Items

### 1. CLI Framework Choice: spf13/cobra vs Standard Library flag

**Decision**: Use spf13/cobra

**Rationale**:

1. **Help Generation**: Cobra automatically generates comprehensive help text and usage documentation, which directly addresses FR-013 (clear help text for all flags) and SC-008 (90% user comprehension without external docs)

2. **Flag Parsing**: Cobra provides robust flag parsing with automatic type conversion, validation, and error handling (FR-014)

3. **Naming Conventions**: Cobra enforces POSIX-style long flags (--flag-name) and supports shorthand versions, aligning with FR-015 (lowercase, hyphen-separated conventions)

4. **Ecosystem Maturity**: Cobra is the de facto standard for Go CLI applications (used by kubectl, hugo, github cli). This provides:
   - Extensive documentation and examples
   - Community familiarity reducing learning curve
   - Battle-tested codebase with edge case handling

5. **Version Command Pattern**: Cobra has established patterns for version commands that can exit immediately (FR-003), including precedence handling when combined with other flags (Edge Case #1)

6. **Simplicity Trade-off**: While Cobra adds dependency weight (~200KB), it eliminates the need to write custom:
   - Help text formatting and layout
   - Flag validation and error messages
   - Subcommand infrastructure (future-proofing)
   - Version command boilerplate

**Alternatives Considered**:

| Alternative | Why Rejected |
|-------------|--------------|
| stdlib `flag` package | No automatic help generation, verbose flag definition, no support for POSIX-style long flags (uses single dash), would require significant boilerplate to meet FR-013 and SC-008 |
| `github.com/urfave/cli` | Less established for single-binary CLI tools, different command structure that doesn't match gitree's simple root command pattern |
| `github.com/spf13/pflag` | POSIX-compatible flags without the command framework - would still require manual help text and version handling |

**Implementation Notes**:

- Cobra command will wrap existing main.go logic
- Root command executes current scanning behavior
- Version flag implemented as PreRunE hook to exit before main logic
- No-color and all flags passed as options to existing functions
- Default behavior (without flags) will filter to show only repos needing attention

---

### 2. Build-Time Variable Injection for Version Information

**Decision**: Use Go linker flags (`-ldflags`) with Makefile variables

**Rationale**:

1. **Standard Go Practice**: The `-ldflags "-X"` approach is the canonical method for injecting build-time variables in Go applications

2. **Makefile Integration**: Already using Makefile for build process (verified in existing Makefile). Variables are already defined:

   ```makefile
   VERSION    := $(shell git describe --tags --always --dirty)
   COMMIT     := $(shell git rev-parse --short HEAD)
   BUILDTIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
   ```

3. **Implementation Pattern**:

   ```go
   package main

   var (
       version   = "dev"      // overridden by -ldflags
       commit    = "none"     // overridden by -ldflags
       buildTime = "unknown"  // overridden by -ldflags
   )
   ```

4. **Makefile Build Command Update**:

   ```makefile
   build:
       CGO_ENABLED=0 \
       go build \
           -ldflags "-s -w \
               -X 'main.version=$(VERSION)' \
               -X 'main.commit=$(COMMIT)' \
               -X 'main.buildTime=$(BUILDTIME)'" \
           -o bin/$(APP_NAME) \
           cmd/gitree/main.go
   ```

5. **FR-012 Compliance**: Default values ("dev", "none", "unknown") provide clear indication of development builds when ldflags are not used

**Alternatives Considered**:

| Alternative | Why Rejected |
|-------------|--------------|
| Embed git describe at runtime | Requires git binary in PATH, fails in CI/containers without .git directory, runtime overhead |
| Version file in binary | Requires extra file management, manual updates, no commit hash |
| Environment variables | Not captured at build time, can be changed post-build, not portable with binary |

---

### 3. Color Suppression Implementation

**Decision**: Global color disable via fatih/color package DisableColor() function

**Rationale**:

1. **Existing Dependency**: `github.com/fatih/color` is already in go.mod (transitive dependency via go-git)

2. **Global Control**: fatih/color provides `color.NoColor = true` global variable that disables all color output throughout the application

3. **Implementation**:

   ```go
   if noColor {
       color.NoColor = true
   }
   ```

4. **FR-005 Compliance**: Single global flag ensures color suppression applies to all output (repository names, paths, status indicators, tree characters, errors, warnings)

5. **SC-002 Verification**: Can verify zero ANSI codes by scanning output for escape sequences (pattern: `\x1b\[`)

**Current Color Usage Audit**:

- Existing code uses fatih/color in internal/tree/formatter.go for:
  - Repository paths (colored based on status)
  - Branch names
  - Ahead/behind indicators
  - Status symbols

**Alternatives Considered**:

| Alternative | Why Rejected |
|-------------|--------------|
| Manual ANSI code stripping | Error-prone, doesn't prevent generation (just removes after), performance overhead |
| Pass no-color flag through all functions | Violates DRY, complex refactoring, error-prone |
| Environment variable (NO_COLOR) | Good practice but want explicit flag control, can support both |

**Best Practice Addition**: Also respect `NO_COLOR` environment variable per [no-color.org](https://no-color.org/) standard:

```go
noColorFlag := cmd.Flags().Bool("no-color", false, "Disable color output")
if *noColorFlag || os.Getenv("NO_COLOR") != "" {
    color.NoColor = true
}
```

---

### 4. Repository Filtering Logic (Default: Show Changed Only)

**Decision**: Implement filter function in internal/cli package with multi-condition evaluation. Default behavior filters to show only repos needing attention; `--all` flag disables filtering.

**Rationale**:

1. **Library-First Compliance**: Filtering logic belongs in new internal/cli package as it provides standalone utility for determining repository state

2. **Filter Function Signature**:

   ```go
   // FilterOptions configures repository filtering
   type FilterOptions struct {
       ShowAll bool  // When true, disables filtering (shows all repos)
   }

   // IsClean determines if repository is in clean state per FR-008
   func IsClean(repo *models.Repository) bool {
       // Returns true only if ALL conditions met:
       // - Branch is "main" or "master"
       // - No uncommitted changes
       // - No stashes
       // - Not ahead/behind remote
       // - Has remote tracking configured
   }

   // FilterRepositories returns filtered list maintaining tree structure
   // By default (ShowAll=false), returns only repos needing attention
   // With ShowAll=true, returns all repos
   func FilterRepositories(repos []*models.Repository, opts FilterOptions) []*models.Repository
   ```

3. **FR-008 Clean State Definition**:

   ```go
   func IsClean(repo *models.Repository) bool {
       if repo.GitStatus == nil {
           return false  // Fail-safe: unknown state = not clean (FR-009)
       }

       status := repo.GitStatus

       // Check branch (main or master only)
       if status.Branch != "main" && status.Branch != "master" {
           return false
       }

       // Check for uncommitted changes
       if status.HasUncommittedChanges {
           return false
       }

       // Check for stashes
       if status.Stashes > 0 {
           return false
       }

       // Check remote tracking
       if !status.HasRemote {
           return false  // No remote = not clean (FR-009)
       }

       // Check ahead/behind
       if status.Ahead > 0 || status.Behind > 0 {
           return false
       }

       // Detached HEAD
       if status.IsDetached {
           return false
       }

       return true  // All conditions met
   }
   ```

4. **Tree Structure Preservation (FR-010, FR-011)**:
   - Build full tree first
   - Apply filter to leaf nodes (repositories) - default mode excludes clean repos
   - Prune branches that contain only clean repos (when ShowAll=false)
   - Keep parent directories leading to remaining repos

5. **Edge Case Handling**:
   - Timeout/error in status extraction → include repo (fail-safe per edge case)
   - Detached HEAD → include repo (edge case)
   - Main branch WITH changes → include repo (edge case)
   - All repos clean in default mode → show message (edge case)
   - `--all` flag: show all repos including clean ones

**Alternatives Considered**:

| Alternative | Why Rejected |
|-------------|--------------|
| Filter during scanning | Violates separation of concerns, couples scanner to filtering logic |
| Filter in main.go | Not library code, can't be tested independently |
| Simple dirty flag on Repository | Doesn't capture nuanced clean state definition (FR-008) |
| Keep previous behavior (show all by default) | Spec explicitly requires changed default behavior for improved workflow |

---

### 5. Testing Strategy

**Decision**: Three-tier testing approach

**Test Tiers**:

1. **Unit Tests** (`internal/cli/filter_test.go`):
   - Test IsClean() with all combinations of git status
   - Test FilterRepositories() with various combinations (ShowAll true/false)
   - Use testify assertions (existing pattern)

2. **Integration Tests** (`cmd/gitree/main_test.go`):
   - Test flag parsing and validation
   - Test version output format
   - Test color suppression (verify no ANSI codes)
   - Test default filtering (show only changed repos) with real git repos in temp directories
   - Test `--all` flag (show all repos including clean ones)

3. **Acceptance Tests** (manual verification):
   - Each acceptance scenario from spec
   - Edge cases from spec
   - Success criteria verification (SC-001 through SC-008)

**Test-First Mandate (Constitution III)**:

- Write unit tests for IsClean() covering all FR-008 conditions
- Write integration tests for flag combinations and default behavior
- Get user approval of test suite before implementation
- Follow Red-Green-Refactor cycle

---

## Summary

All NEEDS CLARIFICATION items resolved:

1. ✅ **CLI Framework**: spf13/cobra chosen for help generation, flag parsing, and ecosystem maturity
2. ✅ **Version Injection**: Go ldflags with Makefile variables
3. ✅ **Color Suppression**: fatih/color global disable + NO_COLOR env var support
4. ✅ **Filtering Logic**: Library function in internal/tree with multi-condition IsClean() evaluation

**Complexity Justification**:

- Cobra framework: Justified by automatic help generation (FR-014, SC-008) and robust flag handling (FR-015, FR-016)
- Filtering logic: Required by spec (FR-006, FR-008, FR-009) - inherent complexity in clean-state definition
- Default behavior change: Improves user workflow by focusing attention on repos needing action (FR-006)

**Next Phase**: Proceed to Phase 1 (Design & Contracts) to generate data models and API contracts.
