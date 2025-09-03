# MUNO Project Insights
Date: 2025-09-03

## Project Overview
**MUNO** (문어 - "octopus" in Korean) is a multi-repository orchestration tool with tree-based navigation. Like an octopus coordinating its eight arms, MUNO orchestrates multiple repositories with intelligent coordination.

## Architecture
- **Tree-based navigation**: Repositories organized in parent-child hierarchy
- **CWD-first resolution**: Current directory determines operation target
- **Lazy loading**: Repositories clone only when accessed
- **Direct git management**: Native git operations with tree awareness

## Core Components
1. **Tree Manager** (`internal/tree/`) - Tree-based workspace navigation
2. **Node System** (`internal/tree/node.go`) - Repository nodes with parent-child relationships
3. **Manager V2** (`internal/manager/manager_v2.go`) - Main orchestration with dependency injection
4. **Configuration** (`internal/config/config_v3.go`) - Tree-based configuration schema
5. **Git Integration** (`internal/git/`) - Direct git operations at any tree node
6. **Adapters** (`internal/adapters/`) - Interfaces for file system, UI, and tree operations

## Test Coverage Status (2025-09-03)
### Overall: 45.6% (Target: 70%)

### Package Coverage:
- ✅ **internal/config**: 87.3% (exceeds 80% target)
- ✅ **internal/adapters**: 84.2% (exceeds 80% target) 
- ✅ **internal/git**: 84.1% (exceeds 80% target)
- ✅ **cmd/muno**: 81.0% (exceeds 80% target)
- ✅ **internal/tree**: 80.3% (meets 80% target)
- ❌ **internal/manager**: 56.9% (needs +23.1% for 80% target)

### Manager Package - Critical Gap
Low coverage functions blocking 70% total:
- Git operations: CloneRepos, StatusNode, PullNode, PushNode (~10-15% each)
- Tree display: ListNodesRecursive, ShowCurrent, ShowTreeAtPath (~13-28% each)
- Metrics provider: All methods at 0%

## Key Commands
```bash
# Build
make build  # Creates ./bin/muno

# Initialize workspace
./bin/muno init <workspace-name>

# Navigation
./bin/muno use <path>    # Navigate to node
./bin/muno current        # Show current position
./bin/muno tree          # Display tree structure

# Repository Management
./bin/muno add <repo-url> [--lazy]
./bin/muno remove <name>
./bin/muno clone [--recursive]

# Git Operations
./bin/muno pull [--recursive]
./bin/muno push [--recursive]
./bin/muno status [--recursive]
./bin/muno commit -m "message"

# Testing
go test ./...
./test_coverage.sh  # Run coverage report
```

## Project Structure
```
muno/
├── cmd/muno/           # CLI application
├── internal/
│   ├── adapters/       # Interface adapters
│   ├── config/         # Configuration management
│   ├── git/           # Git operations
│   ├── interfaces/    # Core interfaces
│   ├── manager/       # Orchestration logic (V2)
│   ├── mocks/         # Test mocks
│   └── tree/          # Tree navigation
├── test/              # Integration tests
├── bin/muno           # Built binary
└── muno.yaml         # Workspace configuration
```

## Testing Strategy
- Unit tests with mocks for all packages
- Table-driven tests for multiple scenarios
- Integration tests in test/ directory
- Coverage target: 70% total, 80% per package
- Use `./test_coverage.sh` for coverage reports

## Key Design Patterns
1. **Dependency Injection**: Manager V2 uses DI for all providers
2. **Interface-based**: All major components behind interfaces
3. **Tree Navigation**: Repository operations based on tree position
4. **Lazy Loading**: Clone repositories only when needed
5. **CWD Resolution**: Current directory determines operation scope

## Recent Refactoring (2025)
- Migrated from Manager V1 to V2 with dependency injection
- Introduced adapter pattern for better testability
- Achieved 80%+ coverage in 5 of 6 main packages
- Simplified tree navigation with CWD-first resolution

## Next Priority
To reach 70% total coverage, focus on internal/manager package:
1. Add tests for git operations (CloneRepos, PullNode, PushNode)
2. Add tests for tree display functions
3. This would add ~20-25% to manager package coverage
4. Would bring total project coverage to ~70%