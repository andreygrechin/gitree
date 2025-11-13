<!--
Sync Impact Report:
- Version: 0.0.0 → 1.0.0
- Modified principles: N/A (initial constitution)
- Added sections:
  * Core Principles (5 principles: Library-First, CLI Interface, Test-First, Observability, Simplicity)
  * Development Standards (Code Quality, Documentation, Testing Standards)
  * Governance (Amendment Process, Versioning Policy, Compliance, Runtime Guidance)
- Removed sections: N/A
- Templates requiring updates:
  ✅ .specify/templates/plan-template.md (reviewed - constitution check section properly references constitution, complexity tracking aligns)
  ✅ .specify/templates/spec-template.md (reviewed - requirements and success criteria align with principles)
  ✅ .specify/templates/tasks-template.md (reviewed - task organization supports TDD principle, phase structure aligns)
  ✅ .specify/templates/checklist-template.md (reviewed - generic template, no updates needed)
  ✅ .specify/templates/agent-file-template.md (reviewed - generic template, no updates needed)
- Follow-up TODOs: None
- Version bump rationale: MAJOR version 1.0.0 as this is the initial ratification of the constitution, establishing the foundational governance structure and core principles for the Gitree project.
-->

# Gitree Constitution

## Core Principles

### I. Library-First

Every feature MUST start as a standalone library. Libraries MUST be self-contained, independently testable, and documented with a clear purpose. Organizational-only libraries (libraries created solely for code organization without independent utility) are prohibited.

**Rationale**: Library-first architecture ensures modularity, reusability, and testability. Each library can be developed, tested, and maintained independently, reducing coupling and enabling better composition of functionality.

### II. CLI Interface

Every library MUST expose its functionality via a command-line interface. All CLI tools MUST follow the text in/out protocol: input from stdin or arguments, output to stdout, errors to stderr. Both JSON and human-readable formats MUST be supported for output.

**Rationale**: CLI interfaces provide universal accessibility, scriptability, and composability. The text protocol ensures tools can be chained together using standard Unix conventions, enabling powerful workflows without complex integrations.

### III. Test-First (NON-NEGOTIABLE)

Test-Driven Development is mandatory. Tests MUST be written before implementation. User approval of tests MUST be obtained before implementation begins. The Red-Green-Refactor cycle MUST be strictly enforced: write tests that fail, implement to make them pass, then refactor.

**Rationale**: TDD ensures requirements are clearly understood before coding begins, produces better design through test-driven API thinking, provides immediate feedback, and creates comprehensive test coverage as a natural byproduct of development.

### IV. Observability

All CLI tools MUST ensure debuggability through their text I/O design. Structured logging MUST be implemented for all significant operations. Logs MUST be written to stderr to keep stdout clean for data output. Diagnostic information MUST be accessible without requiring special debugging tools.

**Rationale**: Observability is not optional for production systems. Text-based protocols and structured logging enable developers and operators to understand system behavior, diagnose issues quickly, and maintain systems effectively over time.

### V. Simplicity

Start simple and apply YAGNI (You Aren't Gonna Need It) principles rigorously. Complexity MUST be justified against the Complexity Tracking table in implementation plans. Prefer straightforward solutions over clever abstractions. Add features only when there is demonstrated need.

**Rationale**: Simplicity reduces cognitive load, minimizes bugs, accelerates development, and improves maintainability. Complex solutions are expensive to build, test, understand, and modify. Simple code is easier to reason about and more resilient to change.

## Development Standards

### Code Quality

- All code MUST pass linting and formatting checks before commit
- Code reviews MUST verify adherence to all constitutional principles
- Breaking changes MUST follow semantic versioning (MAJOR.MINOR.PATCH)
- Public APIs MUST maintain backward compatibility within major versions

### Documentation

- Every library MUST include a README with purpose, installation, usage examples, and API reference
- CLI tools MUST provide `--help` output documenting all flags and arguments
- Complex algorithms or business logic MUST include inline comments explaining the "why"
- Breaking changes MUST be documented in CHANGELOG.md

### Testing Standards

- Unit tests MUST cover individual functions and edge cases
- Integration tests MUST cover CLI contract compliance (input/output protocol)
- Contract tests MUST validate interfaces between libraries
- All tests MUST be automated and run in CI/CD pipelines

## Governance

### Amendment Process

This constitution supersedes all other development practices and guidelines. Amendments require:

1. Written proposal documenting the change and rationale
2. Review and approval from project maintainers
3. Migration plan for existing code if applicable
4. Update to constitution version following semantic versioning rules

### Versioning Policy

Constitution version follows semantic versioning:

- **MAJOR**: Backward incompatible changes (removing/redefining principles)
- **MINOR**: Additions (new principles or materially expanded guidance)
- **PATCH**: Clarifications, wording improvements, non-semantic refinements

### Compliance

All pull requests and code reviews MUST verify compliance with constitutional principles. Violations MUST be documented in the Complexity Tracking section of implementation plans with explicit justification. Unjustified violations MUST be rejected.

### Runtime Guidance

For ongoing development guidance and agent-specific instructions, refer to `.specify/memory/agent-guidance.md` (if it exists). The constitution defines the immutable "what" and "why"; runtime guidance documents evolve the "how" based on project experience.

**Version**: 1.0.0 | **Ratified**: 2025-11-13 | **Last Amended**: 2025-11-13
