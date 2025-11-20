# gitree

A CLI tool that recursively scans directories for Git repositories and displays them in a tree structure with inline status information.

## About

This project serves as an exploration of [spec-kit](https://github.com/github/spec-kit), a specification-driven development framework. All design artifacts,
implementation plans, and project governance are maintained in the `.specify/` directory.

## Features

- **Tree visualization**: Displays Git repositories in an ASCII tree structure (similar to the `tree` command)
- **Inline Git status**: Shows branch name, ahead/behind counts, stashes, and uncommitted changes
- **Concurrent scanning**: Asynchronously extracts Git status for multiple repositories in parallel
- **Bare repository support**: Detects and displays both regular and bare repositories
- **Graceful error handling**: Continues operation when encountering inaccessible repositories

Example output:

```text
.
├── project-a [[ main | ↑2 ↓1 $ * ]]
├── project-b [[ develop | ○ ]]
└── libs
    ├── lib-core [[ main ]]
    └── lib-utils [[ DETACHED | * ]]
```

**Status symbols**:

- Branch name or `DETACHED` for detached HEAD
- `↑N` - commits ahead of remote
- `↓N` - commits behind remote
- `○` - no remote configured
- `$` - has stashes
- `*` - has uncommitted changes
- `bare` - bare repository

## Installation

### Pre-built Binaries

Download the latest release from the [releases page](https://github.com/andreygrechin/gitree/releases).

**Linux/macOS:**
```bash
tar -xzf gitree_*.tar.gz
sudo mv gitree /usr/local/bin/
```

**Windows:** Extract the `.zip` file and add the directory to your PATH.

<details>
<summary>Install a specific version</summary>

Replace the download URL with the specific version tag (e.g., `v1.0.0`):
```bash
wget https://github.com/andreygrechin/gitree/releases/download/v1.0.0/gitree_*.tar.gz
```
</details>

<details>
<summary>Upgrade/Uninstall</summary>

**Upgrade:** Download the new version and replace the existing binary.

**Uninstall:**
```bash
sudo rm /usr/local/bin/gitree
```
</details>

### Using Go Install

If you have Go 1.25+ installed:

```bash
go install github.com/andreygrechin/gitree/cmd/gitree@latest
```

Installs to `$GOPATH/bin/gitree` (typically `$HOME/go/bin/gitree`). Ensure `$GOPATH/bin` is in your PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

<details>
<summary>Install a specific version</summary>

Use the version tag instead of `@latest`:
```bash
go install github.com/andreygrechin/gitree/cmd/gitree@v1.0.0
```
</details>

<details>
<summary>Upgrade/Uninstall</summary>

**Upgrade:** Run the same command with `@latest` or a newer version tag.

**Uninstall:**
```bash
rm $(go env GOPATH)/bin/gitree
```
</details>

### From Source

```bash
git clone https://github.com/andreygrechin/gitree.git
cd gitree
make build
sudo mv bin/gitree /usr/local/bin/  # or run directly: ./bin/gitree
```

<details>
<summary>Build a specific version</summary>

Checkout the specific tag before building:
```bash
git clone https://github.com/andreygrechin/gitree.git
cd gitree
git checkout v1.0.0
make build
```
</details>

<details>
<summary>Upgrade/Uninstall</summary>

**Upgrade:**
```bash
git pull origin main
make build
```

**Uninstall:** `sudo rm /usr/local/bin/gitree`
</details>

## Build

```bash
make build              # Build binary to bin/gitree
```

## Usage

Run from any directory to scan for Git repositories:

```bash
./bin/gitree
```

The tool will recursively scan the current directory and display all Git repositories in a tree format with their status.

## Development

See [CLAUDE.md](CLAUDE.md) for build commands, architecture details, and development conventions.

### Testing

```bash
make test               # Run all tests
make lint               # Format code and run linters
```

### Project Governance

This project follows a constitution-driven development approach. See `.specify/memory/constitution.md` for core principles including:

- **Library-First**: Features start as standalone libraries
- **CLI Interface**: Text in/out protocol with JSON and human-readable formats
- **Test-First**: Mandatory TDD with Red-Green-Refactor cycle
- **Observability**: Structured logging and debuggability
- **Simplicity**: YAGNI principles with complexity tracking

## License

MIT
