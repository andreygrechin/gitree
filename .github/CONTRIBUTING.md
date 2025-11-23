# Contributing to gitree

Contributions are welcome! This document covers the essentials for contributing to gitree.

## Before You Start

- Check [CLAUDE.md](../CLAUDE.md) for project architecture, build commands, and code conventions
- Review existing [specifications](../specs/) to understand how features are designed
- For bugs, check if an issue already exists before opening a new one

## Contributing Code

1. Fork the repository and create a feature branch
2. Make your changes following the project's code conventions (see CLAUDE.md)
3. Run quality checks  locally before committing:

    ```bash
    make lint      # Format and lint (required)
    make test      # Run tests (required)
    make security  # Security checks (required)
    ```

4. Ensure all tests pass and add tests for new functionality
5. Open a pull request with a clear description of changes

## Feature Development

For new features, consider creating a specification in `specs/` using [spec-kit](https://github.com/github/spec-kit). This helps document design decisions and implementation approach.

## Code Quality Expectations

- All linters must pass (golangci-lint with project config)
- New functionality requires test coverage

## Pull Request Guidelines

- Keep PRs focused on a single concern
- Include tests for bug fixes and new features
- Update CLAUDE.md if adding significant architectural changes
- Reference related issues in PR description

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Questions? Open an issue for discussion.
