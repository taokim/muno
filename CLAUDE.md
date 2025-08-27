# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## ✅ V3 Tree-Based Architecture Implemented

**Completed (2024-12-27)**: 
- **V3 tree-based architecture is now fully implemented**
- **All scope concepts have been removed**
- **Tree navigation with CWD-first resolution is working**
- **Lazy loading and auto-clone on navigation implemented**
- **No migration from v2 - v3 is a clean slate**
- The codebase now uses tree-based navigation exclusively

## Overview

Repo-Claude v3 is a multi-repository orchestration tool with a tree-based navigation system. Repositories form a navigable tree structure where every operation is based on your current position (CWD-first resolution).

**Key Features:**
- **Tree-based navigation** - Navigate repositories like a filesystem
- **CWD-first resolution** - Current directory determines operation target
- **Lazy loading** - Repositories clone on-demand when navigating
- **Clear targeting** - Every command shows what it will affect
- **Direct git management** - Native git operations at any tree node
- **No scope concept** - Simple tree navigation replaces complex scope management

## Commands

### Running the Tool
```bash
# Build the binary (creates ./bin/rc)
make build
# OR manually:
go build -o bin/rc ./cmd/repo-claude

# Initialize a new workspace
./bin/rc init <workspace-name>

# Navigation (v3)
./bin/rc use <path>              # Navigate to a node (changes CWD)
./bin/rc current                 # Show current position
./bin/rc tree                    # Display tree structure
./bin/rc list                    # List child nodes

# Repository Management
./bin/rc add <repo-url> [--lazy] # Add child repository
./bin/rc remove <name>           # Remove child repository
./bin/rc clone [--recursive]     # Clone lazy repositories

# Start Claude session
./bin/rc start [path]            # Start at current or specified node

# Check status
./bin/rc status             # Shows workspace configuration and scope status

# Git operations within scopes
./bin/rc pull --scope <name>              # Pull repos in scope
./bin/rc commit --scope <name> -m "msg"   # Commit in scope
./bin/rc push --scope <name>              # Push scope changes
./bin/rc branch --scope <name> <branch>   # Switch branch in scope

# Documentation management
./bin/rc docs create --global <name>      # Create global doc
./bin/rc docs create --scope <scope>      # Create scope doc
./bin/rc docs list                        # List all docs
./bin/rc docs sync                        # Sync to Git

# Pull request management
./bin/rc pr list --scope <name>           # List PRs for scope repos
./bin/rc pr create --scope <name>         # Create PR for scope
./bin/rc pr status --scope <name>         # Show PR status
```

### Development Commands
```bash
# Run tests
go test ./...

# Run with verbose output
go test -v ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o repo-claude-linux cmd/main.go
GOOS=darwin GOARCH=amd64 go build -o repo-claude-darwin cmd/main.go
GOOS=windows GOARCH=amd64 go build -o repo-claude.exe cmd/main.go

# Install as a Go tool
go install ./cmd/repo-claude
```

## Architecture

### Core Components (v2 Architecture)

1. **Manager** (`internal/manager/manager.go`)
   - Main orchestration handling workspace initialization and coordination
   - Delegates to specialized managers (ScopeManager, DocsManager)
   - Manages v2 configuration with isolation mode

2. **ScopeManager** (`internal/scope/manager.go`)
   - Creates and manages isolated scope directories in `workspaces/`
   - Handles scope lifecycle (create, delete, archive)
   - Manages `.scope-meta.json` files for scope metadata

3. **Scope** (`internal/scope/scope.go`)
   - Represents an individual isolated workspace
   - Handles git operations within scope context
   - Creates CLAUDE.md files for AI context
   - Manages repository state per scope

4. **Configuration System** (`internal/config/config.go`)
   - v2 configuration schema with `version: 2`
   - `isolation_mode: true` by default
   - `base_path: "workspaces"` for isolated scopes
   - E-commerce themed default repositories and scopes

5. **Documentation System** (`internal/docs/manager.go`)
   - Manages global docs in `docs/global/`
   - Manages scope-specific docs in `docs/scopes/`
   - Supports Git synchronization

6. **Types** (`internal/scope/types.go`)
   - `Meta`: Scope metadata (id, name, type, state, repos)
   - `RepoState`: Repository status within scope
   - `Type`: persistent or ephemeral
   - `State`: active, inactive, or archived
   - No TTL field per requirements

### Workspace Structure (v2)

```
my-ecommerce-platform/
├── repo-claude.yaml         # v2 configuration
├── .repo-claude-state.json  # State tracking
├── shared-memory.md         # Cross-scope coordination
├── docs/                    # Documentation system
│   ├── global/             # Project-wide docs
│   └── scopes/             # Scope-specific docs
└── workspaces/             # Isolated scope directories
    ├── wms-dev/            # WMS development scope
    │   ├── .scope-meta.json
    │   ├── wms-core/
    │   ├── wms-inventory/
    │   ├── wms-shipping/
    │   └── shared-libs/
    └── oms-hotfix/         # OMS hotfix scope
        ├── .scope-meta.json
        ├── oms-core/
        ├── oms-payment/
        └── shared-libs/
```

See [docs/workspace-structure.md](docs/workspace-structure.md) for details.

### Key Design Patterns (v2)

1. **Workspace Isolation**: Each scope operates in its own `workspaces/<scope-name>/` directory
2. **Scope Metadata**: `.scope-meta.json` tracks scope state and repository information
3. **Three-Level Architecture**: Project → Scope → Repository hierarchy
4. **Persistent vs Ephemeral**: Long-lived vs temporary scope lifecycle management
5. **Direct Git Management**: Native git operations with per-scope state
6. **Shared Memory Pattern**: `shared-memory.md` for cross-scope coordination
7. **Documentation System**: Structured docs with global and scope-specific content
8. **No TTL**: Manual scope lifecycle management without automatic expiration

### Data Flow (v2)

1. **Initialization**: 
   - Creates project directory with v2 structure
   - Generates `repo-claude.yaml` with `version: 2`, `isolation_mode: true`
   - Creates `workspaces/` base directory for isolated scopes
   - Initializes `docs/` structure (global and scopes subdirectories)
   - Creates shared memory file

2. **Scope Creation**:
   - Create isolated directory in `workspaces/<scope-name>/`
   - Generate `.scope-meta.json` with scope metadata
   - Clone repositories into scope directory
   - Create CLAUDE.md in each repository
   - Initialize scope documentation

3. **Scope Lifecycle**: 
   - Load scope metadata → verify repos → start Claude session
   - Track state (active/inactive/archived)
   - Manage repository state per scope
   - Handle cleanup and archival

4. **Coordination**: 
   - Scopes communicate via shared memory
   - Each scope has isolated git state
   - Documentation system provides cross-scope knowledge

## Dependencies

Required external tools:
- **git**: Version control (>= 2.0)
- **claude**: Claude Code CLI for AI agents

Go dependencies (see `go.mod`):
- `gopkg.in/yaml.v3`: YAML configuration parsing
- Standard library packages for JSON, file I/O, process management

## Important Implementation Details

1. **Git Operations**:
   - Direct git clone/pull operations replace repo tool
   - Parallel repository operations for performance
   - Branch tracking and status monitoring

2. **Process Management**: 
   - Uses `os/exec` for non-blocking Claude Code execution
   - Terminal tab creation via AppleScript on macOS
   - Fallback to new window if tab creation fails
   - Signal handling for graceful shutdown
   - Process monitoring with numbered tracking

3. **Dependency Resolution**: 
   - Topological sort for scope startup order
   - Dependency validation before scope launch
   - Repository pattern matching with wildcard support

4. **Error Handling**: 
   - Graceful degradation when repositories or tools are missing
   - Comprehensive error reporting with context

5. **State Persistence**: 
   - Atomic writes for state file updates
   - Recovery mechanisms for interrupted operations

6. **Configuration Management**:
   - Schema validation for YAML configuration
   - Default values and interactive setup support

## Testing

**Coverage Target**: 70-80% for all packages (current: ~57% overall)

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./test/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage percentage
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1
```

### Current Coverage Status
- `internal/config`: 91.2% ✓
- `internal/git`: 65.0%
- `internal/docs`: 66.9%
- `internal/scope`: 46.5%
- `internal/manager`: 43.4%

### Testing Guidelines
- Write unit tests for all new functionality
- Aim for >80% test coverage per package
- Use table-driven tests for multiple test cases
- Mock external dependencies appropriately
- Test both success and error paths

## Extension Points

1. **Scope Configuration**: Add new scopes in the `scopes` section of config
2. **Project Structure**: Modify repository projects in workspace configuration
3. **Coordination Mechanisms**: Extend shared memory format or add new coordination files
4. **Command Extensions**: Add new subcommands in the CLI
5. **Git Strategies**: Implement alternative branching strategies beyond trunk-based
6. **Terminal Support**: Add tab support for more terminal emulators

## Code Style

- Follow standard Go conventions and `gofmt`
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and testable
- Use interfaces for extensibility

## Building and Distribution

```bash
# Build release binaries
make release

# Build for current platform (creates ./bin/rc)
make build

# Run linters
make lint

# Clean build artifacts
make clean

# Check version (format: YYMMDDHHMM for local builds, git tag for releases)
./bin/rc --version
```

### Local
- Local builds: `./bin/rc` (in project root)
- Installed: `$GOPATH/bin/rc` (via `make install`)

### Production
- Release by tagging new version via GitHub Action, do not use goreleaser directly or any release script locally for GitHub releases
- When releasing, verifying the release done with GH API