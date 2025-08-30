# MUNO Architecture

## Overview

MUNO is a multi-repository orchestration tool for Claude Code that provides **tree-based navigation and workspace management**. The architecture treats your entire codebase as a navigable filesystem where repositories form parent-child relationships.

## Tree-Based Architecture

The core innovation is organizing repositories in a tree structure with filesystem-like navigation:

```
Workspace Root
    ↓
Repository Tree (repos/)
    ↓  
Node Level (team/service/component)
```

### Key Principles

1. **Tree Navigation**: Repositories organized in parent-child hierarchy
2. **CWD-First Resolution**: Current directory determines operation target
3. **Lazy Loading**: Repositories clone only when accessed
4. **Clear Targeting**: Every operation shows what it will affect
5. **Direct Git Management**: Native git operations with tree awareness

## Core Components

### 1. Tree Manager (`internal/tree`)

The heart of the navigation system:

#### `manager.go` - Tree Operations
- **Navigate**: Move through the tree changing current position
- **Resolve**: Map CWD to tree nodes for operations
- **Add/Remove**: Manage tree structure dynamically
- **Clone**: Handle lazy repository loading

#### `node.go` - Node Structure
- **Hierarchy**: Parent-child relationships
- **State**: Track repository status (cloned, lazy, modified)
- **Metadata**: Store node information and configuration
- **Operations**: Execute commands at node level

#### `state.go` - Persistence
- **Tree State**: Save/load tree structure
- **Current Position**: Track navigation position
- **History**: Maintain navigation history

### 2. Configuration System (`internal/config`)

Manages workspace configuration:

#### `config.go` - Configuration
- **Tree Structure**: Define initial tree layout
- **Repository Settings**: URLs, branches, lazy flags
- **Workspace Metadata**: Name, root path, settings

#### `types.go` - Data Types
- **Config**: Main configuration structure
- **NodeConfig**: Repository node configuration
- **TreeState**: Runtime tree state

### 3. Manager (`internal/manager`)

Orchestrates all operations:

- **InitWorkspace**: Creates tree-based workspace
- **TreeOperations**: Coordinates tree navigation
- **GitIntegration**: Manages git operations
- **SessionManagement**: Launches Claude Code sessions

### 4. Git Integration (`internal/git`)

Handles version control:
- **NodeOperations**: Git commands at specific nodes
- **RecursiveOps**: Operations across subtrees
- **StatusTracking**: Monitor repository states
- **ParallelExecution**: Concurrent git operations

## Data Flow

### Initialization
```
muno init → Create Workspace → Setup Tree Root → Initialize Config → Create State File
```

### Tree Building
```
muno add <url> → Create Node → Update Parent → Clone/Mark Lazy → Save State
```

### Navigation
```
muno use <path> → Resolve Path → Change Position → Auto-Clone Lazy → Update CWD
```

### Operations
```
Command → Resolve Target (CWD/Explicit) → Execute at Node → Update State → Show Feedback
```

## Workspace Structure

```
my-platform/
├── muno.yaml          # Configuration
├── .muno-state.json   # Tree state
├── repos/                    # Tree root
│   ├── team-backend/         # Parent node (git repo)
│   │   ├── .git/
│   │   ├── payment-service/  # Child repo
│   │   ├── order-service/    # Child repo
│   │   └── shared-libs/      # Lazy repo
│   └── team-frontend/        # Parent node (git repo)
│       ├── .git/
│       ├── web-app/          # Child repo
│       └── component-lib/    # Lazy repo
└── CLAUDE.md                 # AI context
```

## Resolution System

### Target Resolution Priority

1. **Explicit Path**: User-specified target
2. **CWD Mapping**: Current directory location
3. **Stored Position**: Last navigation position
4. **Root Fallback**: Default to workspace root

### Resolution Feedback

Every command shows its target clearly:
```
🎯 Target: team/backend/payment (from CWD)
🎯 Target: team/frontend (explicit)
🎯 Target: / (root fallback)
```

## Key Features

### Lazy Loading

Repositories clone on-demand:
- Mark repositories as lazy during addition
- Auto-clone when navigating to them
- Manual clone with `muno clone` command
- Recursive clone for entire subtrees

### Recursive Operations

Commands can operate on entire subtrees:
- `--recursive` flag for git operations
- Parallel execution for performance
- Progress tracking across repositories
- Aggregated status reporting

### State Management

Persistent state across sessions:
- Tree structure saved to JSON
- Current position tracking
- Navigation history
- Repository status cache

## Design Decisions

### Why Tree-Based?

1. **Natural Mental Model**: Developers think in hierarchies
2. **Filesystem Familiarity**: Leverages existing navigation knowledge
3. **Clear Relationships**: Parent-child structure is intuitive
4. **Scalability**: Trees handle large repository counts well

### Why CWD-First?

1. **No Hidden State**: Your location determines behavior
2. **Predictability**: Commands work where you are
3. **Simplicity**: No complex scope management
4. **Transparency**: Always know what will be affected

### Why Lazy Loading?

1. **Performance**: Don't clone until needed
2. **Storage**: Save disk space
3. **Flexibility**: Add repositories without immediate overhead
4. **Speed**: Faster workspace setup

## Performance Considerations

- **Parallel Git Operations**: Execute across multiple repos simultaneously
- **Lazy Loading**: Defer expensive clones until necessary
- **State Caching**: Minimize filesystem operations
- **Efficient Tree Traversal**: Optimized path resolution algorithms

## Security Considerations

- **Git Credentials**: Leverages existing git credential management
- **No Stored Secrets**: No passwords or tokens in configuration
- **Filesystem Permissions**: Respects OS-level access controls
- **Safe Operations**: Confirmation prompts for destructive actions

## Future Enhancements

- **Plugin System**: Extensible command architecture
- **Remote Trees**: Distributed workspace support
- **Advanced Patterns**: Tree templates and presets
- **Performance Metrics**: Operation timing and optimization
- **Enhanced UI**: Interactive tree visualization