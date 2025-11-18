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

2. **Path Resolution (v1.1.1+ fixes)**:
   - **Config Reference Nodes**: Children of config reference nodes use `repos_dir` from the referenced config file, not the parent
   - **Workspace Root Detection**: Uses `os.Lstat()` to check for regular files only, preventing symlinked muno.yaml from being treated as workspace roots
   - **Special Path Cases**:
     - Path `.` returns current directory without tree computation
     - Path `..` from tree root (.nodes) returns workspace directory
     - Path resolution works correctly from any location in the tree
   - **MCD Shell Function**: Integrates with `muno path` command for navigation

3. **Git Operations**:
   - Direct git commands at any tree node
   - Parallel repository operations for performance
   - Branch tracking and status monitoring

4. **Process Management**: 
   - Uses `os/exec` for Claude Code execution
   - Terminal tab creation via AppleScript on macOS
   - Signal handling for graceful shutdown

5. **Error Handling**: 
   - Graceful degradation when repositories missing
   - Comprehensive error reporting with context
   - Recovery mechanisms for interrupted operations

6. **State Persistence**: 
   - Tree state saved to JSON file
   - Current position tracking
   - Navigation history maintained

## Testing

**MANDATORY COVERAGE REQUIREMENT**: 
- **Minimum 70% test coverage for ALL packages**
- **All code changes MUST maintain or improve test coverage**
- **No package should drop below 70% coverage**
- **Target: 70-80% for sustainable quality**

### Current Coverage Status (Target: ≥70%)
- ✅ internal/adapters: 70.5%
- ✅ internal/config: 80.1%
- ✅ internal/git: 78.1%
- ⚠️ internal/plugin: 66.3% (needs improvement)
- ❌ internal/manager: ~30% (critical - needs major work)
- ⚠️ cmd/muno: 62.7% (needs improvement)

### Test Commands

```bash
# Quick regression tests (essential functionality)
make test-basic

# Full regression test suite (all features)
make test-master

# Go unit tests
make test-go

# All tests (unit + regression)
make test-all

# Coverage report
make test-coverage
```

### Testing Infrastructure

**IMPORTANT: Test Entry Points are FIXED**
- **Main Test Runner**: `test/run_regression_tests.sh` - DO NOT create additional test entry points
- **All new tests MUST be integrated** into the existing test framework as functions within `run_regression_tests.sh`
- **NO standalone test scripts** should be created - they must be part of the main runner
- **Makefile targets are FIXED**: Use existing targets (test-basic, test-master, test-all), do not add new ones

### Test Suite Organization

The regression test suite (`test/run_regression_tests.sh`) includes:
- **Initialization & Configuration**: Workspace setup and config management
- **Repository Management**: Add, remove, list operations
- **Clone Behavior**: Lazy loading, eager cloning, recursive operations
- **Pull Behavior**: Update mechanisms for cloned repos
- **Git Operations**: Status, commit, push, pull integration
- **Tree Navigation**: Tree display and traversal
- **Path Resolution & MCD**: Path command and mcd shell function (v1.1.1+ fixes)
- **Error Handling**: Invalid operations and edge cases
- **Advanced Features**: Nested structures, config references, custom repos_dir

### Testing Requirements for Development

**Before Committing Code**:
1. Run coverage check: `go test -cover ./path/to/package`
2. Ensure package has ≥70% coverage
3. Write tests for all new functions/methods
4. Test both success and failure paths
5. Include edge cases and boundary conditions

**Test Patterns to Use**:
- Table-driven tests for comprehensive coverage
- Duck-type mocking for flexible test doubles
- Parallel test execution where possible
- Clear test names describing the scenario

### Adding New Tests

When adding new test functionality:
1. Add a new test function to `test/run_regression_tests.sh` (e.g., `test_new_feature()`)
2. Use the existing `test_case` helper function for individual assertions
3. Add the function call to the main execution flow in the `main()` function
4. DO NOT create new test scripts or entry points

### Regression Testing
The regression test suite validates:
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

### Version History & Critical Fixes
- **v1.1.1**: Fixed mcd functionality broken by symlink handling removal
  - Issue: Removed symlink special case broke config reference node path resolution
  - Fix: Config nodes now properly read referenced config for repos_dir
  - Fix: findWorkspaceRoot uses os.Lstat to ignore symlinked muno.yaml
  - Fix: Path "." and ".." handle special cases correctly

## Critical Development Notes

### Test Framework Constraints
**ABSOLUTELY NO NEW TEST ENTRY POINTS OR SCRIPTS**
- The test infrastructure is FIXED and must not be extended with new entry points
- All tests MUST integrate into `test/run_regression_tests.sh`
- No standalone test scripts, launchers, or runners should be created
- Makefile test targets are FIXED - do not add new ones
- This constraint ensures maintainability and prevents test framework fragmentation

### Path Resolution Implementation
When modifying path resolution (`internal/manager/manager.go`):
- **computeFilesystemPath**: Must check if PARENT node is a config reference for repos_dir
- **ResolvePath**: Path "." should return current directory without tree computation
- **findWorkspaceRoot**: Must use `os.Lstat()` not `os.Stat()` to avoid symlink issues
- Config reference nodes create symlinks that must not be treated as workspace roots

### Config Reference Node Behavior
Config reference nodes (nodes with `file:` field):
- Children use `repos_dir` from the REFERENCED config file
- The referenced config's `repos_dir` overrides parent's setting
- This enables team-based repository organization with custom layouts
- Example: team-frontend can use `webapp/` while team-backend uses `services/`

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