# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Repo-Claude Go is a multi-repository orchestration tool that launches Claude Code agents to work collaboratively across different repositories in a shared workspace. This Go implementation is the active version that replaces the deprecated Python implementation.

**Key Features:**
- Multi-agent coordination across repositories
- Shared memory for inter-agent communication
- Trunk-based development workflow
- Configurable workspace structure
- Agent dependency management
- Direct git management (no Google repo tool dependency)

## Commands

### Running the Tool
```bash
# Build the binary
go build -o repo-claude cmd/main.go

# Initialize a new workspace
./repo-claude init <workspace-name>

# Start agents
./repo-claude start              # Start all auto-start agents
./repo-claude start <agent-name> # Start specific agent

# Stop agents
./repo-claude stop               # Stop all agents
./repo-claude stop <agent-name>  # Stop specific agent

# Check status
./repo-claude status

# Sync repositories
./repo-claude sync
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
   - Main orchestration class managing workspace, configuration, and agent lifecycle
   - Handles direct git operations (replacing Google's repo tool)
   - Manages agent process lifecycle and state persistence

2. **Configuration System** (`pkg/config/`)
   - YAML-based configuration (`repo-claude.yaml`) for workspace and agent settings
   - JSON state file (`.repo-claude-state.json`) for runtime agent status
   - Embedded default configuration template

3. **Git Management** (`pkg/git/`)
   - Direct git operations for cloning and syncing repositories
   - Branch management and status checking
   - Replaces Google's repo tool functionality

4. **Agent Management** (`pkg/agent/`)
   - Launches Claude Code instances with specialized system prompts
   - Supports agent dependencies and auto-start configuration
   - Creates CLAUDE.md files in each repository for agent context

### Workspace Structure

See [docs/workspace-structure.md](docs/workspace-structure.md) for detailed workspace layout and configuration options.

### Key Design Patterns

1. **Trunk-Based Development**: All agents work directly on main branch
2. **Shared Memory Pattern**: `shared-memory.md` file for cross-agent coordination
3. **Repository-Agent Mapping**: Each agent is assigned to specific repositories
4. **Direct Git Management**: Uses native git commands instead of Google's repo tool
5. **State Persistence**: JSON-based state tracking for agent lifecycle

### Data Flow

1. **Initialization**: 
   - Creates workspace directory
   - Generates configuration from template or interactive input
   - Clones repositories using direct git commands
   - Creates CLAUDE.md files in each repository
   - Initializes shared memory file

2. **Agent Lifecycle**: 
   - Load config → check dependencies → start in order → track state → handle termination

3. **Coordination**: 
   - Agents read/write to shared memory
   - Cross-repository awareness through relative paths

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
   - Uses `os/exec` for non-blocking agent execution
   - Signal handling for graceful shutdown
   - Process monitoring and restart capabilities

3. **Dependency Resolution**: 
   - Topological sort for agent startup order
   - Dependency validation before agent launch

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

```bash
# Unit tests
go test ./pkg/...

# Integration tests
go test ./test/integration/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Extension Points

1. **Agent Configuration**: Add new agents in the `agents` section of config
2. **Project Structure**: Modify repository projects in workspace configuration
3. **Coordination Mechanisms**: Extend shared memory format or add new coordination files
4. **Command Extensions**: Add new subcommands in the CLI
5. **Git Strategies**: Implement alternative branching strategies beyond trunk-based

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

# Build for current platform
make build

# Run linters
make lint

# Clean build artifacts
make clean
```