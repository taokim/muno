# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

**MUNO** (문어 in Korean, meaning "octopus") is a multi-repository orchestration tool with a tree-based navigation system. Like an octopus coordinating its eight arms, MUNO orchestrates multiple repositories with intelligent coordination. The name stands for **Multi-repository UNified Orchestration**.

Inspired by Google's Repo tool but designed to overcome its limitations, MUNO introduces true hierarchical tree structures where parent nodes can be documented and managed as first-class citizens. Unlike Repo's flat manifest-based approach, MUNO allows you to build and navigate repository trees that match your team's mental model.

Repositories form a navigable tree structure where every operation is based on your current position (CWD-first resolution).

**Key Features:**
- **Tree-based navigation** - Navigate repositories like a filesystem
- **CWD-first resolution** - Current directory determines operation target
- **Lazy loading** - Repositories clone on-demand when navigating
- **Clear targeting** - Every command shows what it will affect
- **Direct git management** - Native git operations at any tree node

## Commands

### Running the Tool
```bash
# Build the binary (creates ./bin/muno)
make build
# OR manually:
go build -o bin/muno ./cmd/muno

# Initialize a new workspace
./bin/muno init <workspace-name>

# Navigation
./bin/muno use <path>              # Navigate to a node (changes CWD)
./bin/muno current                 # Show current position
./bin/muno tree                    # Display tree structure
./bin/muno list                    # List child nodes

# Repository Management
./bin/muno add <repo-url> [--lazy] # Add child repository
./bin/muno remove <name>           # Remove child repository
./bin/muno clone [--recursive]     # Clone lazy repositories

# Start AI agent sessions
./bin/muno agent [name] [path]     # Start AI agent (claude, gemini, cursor, etc.)
./bin/muno claude [path]           # Start Claude CLI (Anthropic)
./bin/muno gemini [path]           # Start Gemini CLI (Google - requires npm install -g @google/gemini-cli)

# Pass arguments to agents
./bin/muno agent gemini -- --help  # Pass help flag to Gemini CLI
./bin/muno claude backend          # Start Claude at backend directory

# Git operations
./bin/muno pull [--recursive]      # Pull repositories
./bin/muno push [--recursive]      # Push changes
./bin/muno commit -m "msg"         # Commit changes
./bin/muno status [--recursive]    # Show git status
```

### Development Commands
```bash
# Run tests
go test ./...

# Run with verbose output
go test -v ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o muno-linux cmd/main.go
GOOS=darwin GOARCH=amd64 go build -o muno-darwin cmd/main.go
GOOS=windows GOARCH=amd64 go build -o muno.exe cmd/main.go

# Install as a Go tool
go install ./cmd/muno
```

## Architecture

### Core Components

1. **Tree Manager** (`internal/tree/`)
   - Manages tree-based workspace navigation
   - Handles CWD-first resolution for all operations
   - Implements lazy loading for repositories

2. **Node System** (`internal/tree/node.go`)
   - Represents each point in the tree (can be a repository)
   - Manages parent-child relationships
   - Tracks repository state (cloned, lazy, modified)

3. **Manager** (`internal/manager/manager.go`)
   - Main orchestration handling workspace initialization
   - Coordinates tree operations with git commands
   - Manages Claude session launching

4. **Configuration System** (`internal/config/config_v3.go`)
   - Tree-based configuration schema
   - Stores tree state and current position
   - Handles workspace metadata

5. **Git Integration** (`internal/git/`)
   - Direct git operations at any tree node
   - Recursive operations for subtrees
   - Status tracking across the tree

### Workspace Structure

```
my-platform/
├── muno.yaml         # Configuration file
├── .muno-state.json  # Tree state tracking
├── nodes/                   # Tree root
│   ├── team-backend/        # Parent node (also a git repo)
│   │   ├── .git/
│   │   ├── payment-service/ # Child repository
│   │   ├── order-service/   # Child repository
│   │   └── shared-libs/     # Lazy repository (not cloned yet)
│   └── team-frontend/       # Parent node (also a git repo)
│       ├── .git/
│       ├── web-app/         # Child repository
│       └── component-lib/   # Lazy repository
└── CLAUDE.md                # This file
```

### Node Types and Configuration

MUNO supports two primary node types:

1. **Git Repository Nodes** (`url` field):
   ```yaml
   nodes:
     - name: payment-service
       url: https://github.com/org/payment.git
       lazy: true
   ```
   - Clone and manage standard git repositories
   - Can contain muno.yaml for child definitions (hybrid nodes)

2. **Config Reference Nodes** (`file` field):
   ```yaml
   nodes:
     - name: team-frontend
       file: ../frontend/muno.yaml  # Local config delegation
     - name: infrastructure
       file: https://config.company.com/infra.yaml  # Remote config
   ```
   - Delegate subtree management to external configurations
   - Enable distributed, team-based configuration management

**Important**: A node must have EITHER `url` OR `file`, never both.

### Key Design Patterns

1. **Tree-Based Navigation**: Repositories organized in parent-child hierarchy
2. **CWD-First Resolution**: Current directory determines operation target
3. **Lazy Loading**: Repositories clone only when accessed
4. **Clear Targeting**: Every operation shows what it will affect
5. **Direct Git Management**: Native git operations with tree awareness
6. **Distributed Configuration**: Config references enable team autonomy

### Data Flow

1. **Initialization**: 
   - Creates workspace directory
   - Initializes configuration file
   - Creates tree root structure
   - Sets up state tracking

2. **Tree Building**:
   - Add repositories as nodes
   - Create parent-child relationships
   - Mark repositories as lazy if needed
   - Update tree state

3. **Navigation**: 
   - Change current position in tree
   - Auto-clone lazy repositories on access
   - Update CWD to match tree position
   - Track navigation history

4. **Operations**: 
   - Resolve target from CWD or explicit path
   - Execute operation at target node
   - Support recursive operations for subtrees
   - Update tree state after operations

## Dependencies

Required external tools:
- **git**: Version control (>= 2.0)
- **claude**: Claude Code CLI for AI agents

Go dependencies (see `go.mod`):
- `gopkg.in/yaml.v3`: YAML configuration parsing
- Standard library packages for JSON, file I/O, process management

## Important Implementation Details

1. **Tree Operations**:
   - CWD-first resolution for all commands
   - Lazy loading with auto-clone on navigation
   - Recursive operations for subtrees
   - State persistence across sessions

2. **Git Operations**:
   - Direct git commands at any tree node
   - Parallel repository operations for performance
   - Branch tracking and status monitoring

3. **Process Management**: 
   - Uses `os/exec` for Claude Code execution
   - Terminal tab creation via AppleScript on macOS
   - Signal handling for graceful shutdown

4. **Error Handling**: 
   - Graceful degradation when repositories missing
   - Comprehensive error reporting with context
   - Recovery mechanisms for interrupted operations

5. **State Persistence**: 
   - Tree state saved to JSON file
   - Current position tracking
   - Navigation history maintained

## Testing

**Coverage Target**: 70-80% for all packages

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./test/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Regression tests (comprehensive end-to-end testing)
./test/regression/regression_test.sh
```

### Testing Guidelines
- Write unit tests for all new functionality
- Aim for >80% test coverage per package
- Use table-driven tests for multiple test cases
- Mock external dependencies appropriately
- Test both success and error paths

### Regression Testing
For comprehensive end-to-end testing of all MUNO functionality, see the [Regression Test Suite](test/regression/README.md). This test suite validates:
- Configuration persistence (add/remove commands)
- Navigation and lazy loading
- Git operations
- Error handling
- All 36 core MUNO features

## Extension Points

1. **Tree Operations**: Add new navigation patterns or tree algorithms
2. **Repository Types**: Support different VCS systems beyond git
3. **Command Extensions**: Add new subcommands in the CLI
4. **Terminal Support**: Add tab support for more terminal emulators

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

# Build for current platform (creates ./bin/muno)
make build

# Run linters
make lint

# Clean build artifacts
make clean

# Check version
./bin/muno --version
```

### Installation
- Local builds: `./bin/muno` (in project root)
- Installed: `$GOPATH/bin/muno` (via `make install`)

### Production
- Release by tagging new version via GitHub Action
- Verify releases with GitHub API
- IGNORE all backward compatibility and migration, even rollout strategy if the version is lower than 1.0 (based on git tag)

## Roadmap - API & Schema Management (v1.0)

MUNO is planning to evolve beyond repository orchestration to become a comprehensive platform for managing API contracts and message schemas across the entire repository tree.

### Planned Features

**API Signature Management**:
- OpenAPI specifications for REST APIs at any tree level
- Protocol Buffer definitions for gRPC services
- API versioning and evolution tracking
- Automatic API documentation generation from tree structure

**Message Schema Registry**:
- Protocol Buffers and Apache Avro schema management
- Schema evolution and compatibility checking
- Cross-repository schema dependency tracking
- Schema inheritance through tree hierarchy

**Implementation Considerations**:
- Exploring whether to implement as core feature or plugin system
- Commands like `muno schema validate`, `muno api generate-docs`
- Tree-level schema organization (org → team → service)
- Integration with existing API gateways and service meshes

This will enable teams to manage not just code repositories but also the contracts and interfaces between services in a hierarchical, organized manner.