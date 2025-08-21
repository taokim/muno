# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Repo-Claude Go is a multi-repository orchestration tool that launches Claude Code sessions with flexible scopes spanning multiple repositories. This Go implementation is the active version that replaces the deprecated Python implementation.

**Key Features:**
- Scope-based development across multiple repositories
- Flexible repository grouping with wildcard support
- Runs in current terminal by default, with automatic new windows for multiple sessions
- Shared memory for cross-scope communication
- Trunk-based development workflow
- Direct git management (no Google repo tool dependency)
- Environment variables passed to Claude sessions for context

## Commands

### Running the Tool
```bash
# Build the binary (creates ./bin/rc)
make build
# OR manually:
go build -o bin/rc ./cmd/repo-claude

# Initialize a new workspace
./bin/rc init <workspace-name>

# Start scopes
./bin/rc start              # Start all auto-start scopes
./bin/rc start <scope-name> # Start specific scope
./bin/rc start <repo-name>  # Start scope containing this repo
./bin/rc start --new-window # Force open in new window
./bin/rc start scope1 scope2 # Multiple scopes auto-open in new windows

# List running scopes
./bin/rc ps                 # Shows numbered list for easy kill

# Stop scopes
./bin/rc kill               # Stop all scopes
./bin/rc kill <scope-name>  # Stop specific scope
./bin/rc kill 1 2           # Stop by numbers from ps output

# Check status
./bin/rc status

# Sync repositories
./bin/rc sync
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

### Core Components

1. **RepoClaudeManager** (`pkg/manager/manager.go`)
   - Main orchestration class managing workspace, configuration, and scope lifecycle
   - Handles direct git operations (replacing Google's repo tool)
   - Manages Claude Code session lifecycle and state persistence

2. **Configuration System** (`pkg/config/`)
   - YAML-based configuration (`repo-claude.yaml`) with scope definitions
   - JSON state file (`.repo-claude-state.json`) for runtime scope status
   - Supports both new scope-based and legacy agent-based configs
   - Wildcard support in repository patterns

3. **Git Management** (`pkg/git/`)
   - Direct git operations for cloning and syncing repositories
   - Branch management and status checking
   - Replaces Google's repo tool functionality

4. **Scope Management** (`pkg/manager/scopes.go`)
   - Launches Claude Code instances with multi-repository context
   - Sets environment variables (RC_SCOPE_ID, RC_SCOPE_NAME, RC_SCOPE_REPOS, RC_WORKSPACE_ROOT)
   - Supports scope dependencies and auto-start configuration
   - Creates CLAUDE.md files in each repository for workspace context
   - Terminal tab management with fallback to windows

### Workspace Structure

See [docs/workspace-structure.md](docs/workspace-structure.md) for detailed workspace layout and configuration options.

### Key Design Patterns

1. **Trunk-Based Development**: All scopes work directly on main branch
2. **Shared Memory Pattern**: `shared-memory.md` file for cross-scope coordination
3. **Flexible Scope Mapping**: Scopes can include multiple repositories with wildcards
4. **Direct Git Management**: Uses native git commands instead of Google's repo tool
5. **State Persistence**: JSON-based state tracking for scope lifecycle
6. **Terminal Integration**: Smart tab/window management for better workflow

### Data Flow

1. **Initialization**: 
   - Creates workspace directory
   - Generates configuration from template or interactive input
   - Clones repositories using direct git commands
   - Creates CLAUDE.md files in each repository
   - Initializes shared memory file

2. **Scope Lifecycle**: 
   - Load config → resolve repos → check dependencies → start in order → track state → handle termination
   - Environment variables provide scope context to Claude sessions

3. **Coordination**: 
   - Scopes read/write to shared memory
   - Cross-repository awareness through environment variables
   - Working directory is user's current directory (not locked to single repo)

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

**Coverage Requirement**: Maintain test coverage above 80% for all packages.

```bash
# Unit tests
go test ./pkg/...

# Integration tests
go test ./test/integration/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage percentage
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1
```

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

### Binary Location
- Local builds: `./bin/rc` (in project root)
- Installed: `$GOPATH/bin/rc` (via `make install`)