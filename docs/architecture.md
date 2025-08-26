# Repo-Claude v2 Architecture

## Overview

Repo-Claude v2 is a multi-repository orchestration tool for Claude Code that provides **isolated, scope-based development environments**. Each scope operates in its own workspace directory, enabling parallel development without conflicts while maintaining coordination through shared memory.

## Version 2: Workspace Isolation

The v2 architecture introduces fundamental changes to support isolated development:

### Three-Level Hierarchy
```
Project Level (repo-claude.yaml, shared-memory.md)
    ↓
Scope Level (workspaces/<scope-name>/)
    ↓  
Repository Level (individual Git repos)
```

### Key Architectural Changes from v1

1. **Isolated Workspaces**: Each scope has its own directory in `workspaces/`
2. **Scope Metadata**: `.scope-meta.json` files track scope state
3. **Per-Scope Git State**: Repositories cloned separately for each scope
4. **Documentation System**: Structured docs with global and scope-specific content
5. **No TTL**: Manual lifecycle management without automatic expiration

## Architecture Components

### 1. Configuration System (`internal/config`)

The v2 configuration schema introduces isolation-specific settings:

```yaml
version: 2  # Required for v2 features
workspace:
  name: my-ecommerce-platform
  isolation_mode: true      # Default: true
  base_path: workspaces     # Where scopes are created

repositories:
  wms-core:
    url: https://github.com/yourorg/wms-core.git
    default_branch: main
    groups: [wms, backend, core]

scopes:
  wms:
    type: persistent  # or ephemeral
    repos: ["wms-*", "shared-libs"]
    description: "Warehouse Management System"
    model: claude-3-5-sonnet-20241022
    auto_start: false
```

### 2. Scope Manager (`internal/scope`)

Core component for isolated workspace management:

#### `manager.go` - Scope Lifecycle
- **Create**: Initializes new scope directory with metadata
- **Delete**: Removes scope and its repositories
- **Archive**: Marks scope as archived for later reference
- **List**: Shows all scopes with their status
- **Get**: Retrieves specific scope information

#### `scope.go` - Scope Operations
- **Clone**: Clones repositories into scope directory
- **Pull/Push**: Git operations within scope context
- **Start**: Launches Claude Code session for scope
- **Status**: Reports repository state within scope

#### `types.go` - Data Structures
```go
type Meta struct {
    ID          string
    Name        string
    Type        Type        // persistent or ephemeral
    State       State       // active, inactive, archived
    Repos       []string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    // No TTL field per requirements
}
```

### 3. Manager (`internal/manager`)

Orchestrates the overall system:

- **InitWorkspace**: Creates v2 project structure
- **ListScopes**: Combines configured and created scopes
- **StartScope**: Launches isolated scope sessions
- **Delegates**: To ScopeManager and DocsManager

### 4. Documentation System (`internal/docs`)

Manages structured documentation:

```
docs/
├── global/           # Project-wide documentation
│   ├── architecture.md
│   └── api-design.md
└── scopes/          # Scope-specific docs
    ├── wms/
    │   └── inventory-logic.md
    └── oms/
        └── payment-flow.md
```

### 5. Git Manager (`internal/git`)

Handles Git operations with scope awareness:
- **Clone**: Into scope-specific directories
- **Operations**: Pull, push, commit, branch per scope
- **Status**: Repository state within scope context

## Data Flow

### Workspace Initialization
```
rc init → Create v2 Config → Setup workspaces/ → Initialize docs/ → Create shared-memory.md
```

### Scope Creation
```
rc scope create → Create workspaces/<name>/ → Generate .scope-meta.json → Clone repos → Create CLAUDE.md files
```

### Scope Activation
```
rc start <scope> → Load metadata → Verify repos → Launch Claude session → Update state
```

### Cross-Scope Coordination
```
Scope A → Write to shared-memory.md → Scope B reads → Coordinated action
```

## E-Commerce Example Structure

After initialization with e-commerce configuration:

```
my-ecommerce-platform/
├── repo-claude.yaml              # v2 configuration
├── .repo-claude-state.json       # Runtime state
├── shared-memory.md              # Cross-scope coordination
├── docs/
│   ├── global/
│   │   ├── api-standards.md
│   │   └── deployment-guide.md
│   └── scopes/
│       ├── wms/
│       │   └── warehouse-ops.md
│       └── oms/
│           └── order-flow.md
└── workspaces/
    ├── wms-feature-inventory/    # Isolated WMS scope
    │   ├── .scope-meta.json
    │   ├── wms-core/
    │   │   └── CLAUDE.md
    │   ├── wms-inventory/
    │   │   └── CLAUDE.md
    │   └── shared-libs/
    │       └── CLAUDE.md
    └── oms-payment-hotfix/       # Isolated OMS scope
        ├── .scope-meta.json
        ├── oms-core/
        │   └── CLAUDE.md
        ├── oms-payment/
        │   └── CLAUDE.md
        └── shared-libs/
            └── CLAUDE.md
```

## Design Principles

### v2 Specific Principles

1. **Isolation First**: Complete separation between scopes
2. **Explicit Lifecycle**: Manual scope management (no TTL)
3. **State Tracking**: Comprehensive metadata for each scope
4. **Documentation Integration**: First-class documentation support

### Inherited from v1

1. **Simplicity**: Avoid unnecessary complexity
2. **Git Native**: Direct Git operations
3. **Parallel by Default**: Concurrent operations
4. **Fail Gracefully**: Continue despite failures
5. **Clear Structure**: Three-level hierarchy

## Scope Types

### Persistent Scopes
- Long-lived development environments
- Remain across sessions
- Used for ongoing feature work
- Example: `wms`, `oms`, `catalog`

### Ephemeral Scopes
- Temporary workspaces
- Created for specific tasks
- Easy cleanup when done
- Example: `hotfix`, `feature`, `experiment`

## Benefits of v2 Architecture

1. **Parallel Development**: Multiple scopes without conflicts
2. **Clean Separation**: No cross-contamination between projects
3. **Flexible Workflows**: Mix persistent and ephemeral scopes
4. **Better Organization**: Clear structure with metadata
5. **Documentation Integration**: Structured knowledge management

## Migration from v1

For existing v1 workspaces:

1. **Backup** existing workspace
2. **Update** config to v2 schema:
   - Add `version: 2`
   - Add `workspace.isolation_mode: true`
   - Add `workspace.base_path: "workspaces"`
3. **Create** scopes for existing workflows
4. **Clone** repositories into new scope directories

See [Migration Guide](migration-guide.md) for detailed steps.

## Implementation Status

### Completed
- ✅ Core scope management (create, delete, archive)
- ✅ Isolated workspace directories
- ✅ Scope metadata tracking
- ✅ Git operations per scope
- ✅ Documentation system
- ✅ E-commerce example configuration
- ✅ v2 configuration schema

### Not Implemented
- ❌ TTL for ephemeral scopes (per requirements)
- ❌ Automatic scope cleanup
- ❌ Cross-scope repository sharing

## Future Enhancements

1. **Scope Templates**: Pre-configured scope definitions
2. **Scope Cloning**: Duplicate existing scope setup
3. **Scope Snapshots**: Save and restore scope state
4. **Enhanced Coordination**: Advanced cross-scope communication
5. **Web UI**: Visual scope management interface