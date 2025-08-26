# Workspace Structure (v2)

## Overview

Repo-Claude v2 introduces an **isolated workspace architecture** where each scope operates in its own directory under `workspaces/`. This provides complete separation between different development contexts while maintaining project-level coordination.

## v2 Structure with Isolation

```
my-ecommerce-platform/              # Project root
├── repo-claude.yaml                # v2 configuration (version: 2)
├── .repo-claude-state.json         # Runtime state tracking
├── shared-memory.md                # Cross-scope coordination
├── docs/                           # Documentation system
│   ├── global/                    # Project-wide documentation
│   │   ├── architecture.md
│   │   ├── api-standards.md
│   │   └── deployment-guide.md
│   └── scopes/                    # Scope-specific documentation
│       ├── wms/
│       │   └── warehouse-logic.md
│       ├── oms/
│       │   └── order-flow.md
│       └── catalog/
│           └── product-model.md
└── workspaces/                     # Isolated scope directories
    ├── wms-feature-inventory/      # WMS feature development scope
    │   ├── .scope-meta.json       # Scope metadata
    │   ├── wms-core/              # Git repository
    │   │   └── CLAUDE.md
    │   ├── wms-inventory/         # Git repository
    │   │   └── CLAUDE.md
    │   ├── wms-shipping/          # Git repository
    │   │   └── CLAUDE.md
    │   └── shared-libs/           # Git repository
    │       └── CLAUDE.md
    ├── oms-payment-hotfix/        # OMS hotfix scope
    │   ├── .scope-meta.json
    │   ├── oms-core/
    │   │   └── CLAUDE.md
    │   ├── oms-payment/
    │   │   └── CLAUDE.md
    │   └── shared-libs/
    │       └── CLAUDE.md
    └── search-perf-improvement/   # Search optimization scope
        ├── .scope-meta.json
        ├── search-engine/
        │   └── CLAUDE.md
        ├── search-indexer/
        │   └── CLAUDE.md
        └── catalog-api/
            └── CLAUDE.md
```

## Key Files and Directories

### Project Level

#### `repo-claude.yaml`
v2 configuration file with isolation settings:
```yaml
version: 2                    # Required for v2 features
workspace:
  name: my-ecommerce-platform
  isolation_mode: true        # Default: true
  base_path: workspaces      # Where scopes are created
```

#### `.repo-claude-state.json`
Runtime state tracking for active scopes and processes.

#### `shared-memory.md`
Cross-scope coordination file for AI agents to communicate.

#### `docs/`
Structured documentation:
- `global/`: Project-wide documentation
- `scopes/`: Scope-specific documentation

### Scope Level

#### `workspaces/<scope-name>/`
Isolated directory for each scope instance.

#### `.scope-meta.json`
Metadata for the scope:
```json
{
  "id": "wms-feature-inventory-20240115-123456",
  "name": "wms-feature-inventory",
  "type": "persistent",
  "state": "active",
  "repos": ["wms-core", "wms-inventory", "wms-shipping", "shared-libs"],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T14:45:00Z"
}
```

### Repository Level

#### Individual Git Repositories
Each repository is cloned into the scope directory with its own:
- Git state (branches, commits, working directory)
- `CLAUDE.md` file for AI context

## Configuration Examples

### E-Commerce Platform Configuration

```yaml
version: 2
workspace:
  name: ecommerce-platform
  isolation_mode: true
  base_path: workspaces

repositories:
  # WMS repositories
  wms-core:
    url: https://github.com/yourorg/wms-core.git
    default_branch: main
    groups: [wms, backend]
  
  wms-inventory:
    url: https://github.com/yourorg/wms-inventory.git
    default_branch: main
    groups: [wms, backend]
  
  # OMS repositories
  oms-core:
    url: https://github.com/yourorg/oms-core.git
    default_branch: main
    groups: [oms, backend]
  
  oms-payment:
    url: https://github.com/yourorg/oms-payment.git
    default_branch: main
    groups: [oms, backend, payment]

scopes:
  # Persistent scopes
  wms:
    type: persistent
    repos: ["wms-*", "shared-libs"]
    description: "Warehouse Management System"
    model: claude-3-5-sonnet-20241022
  
  oms:
    type: persistent
    repos: ["oms-*", "shared-libs", "api-gateway"]
    description: "Order Management System"
    model: claude-3-5-sonnet-20241022
  
  # Ephemeral scope templates
  hotfix:
    type: ephemeral
    repos: []  # Select at creation time
    description: "Emergency hotfix"
    model: claude-3-5-sonnet-20241022

documentation:
  path: docs
  sync_to_git: true
```

## Benefits of v2 Structure

### 1. Complete Isolation
Each scope has its own:
- Git working directories
- Branch states
- Uncommitted changes
- Repository versions

### 2. Parallel Development
Multiple developers can work on:
- Different features in the same repositories
- Hotfixes without disrupting feature work
- Experiments without affecting main development

### 3. Clean Organization
- Clear separation of concerns
- Easy to identify scope purpose
- Simple cleanup when work is complete

### 4. Flexible Workflows

#### Feature Development
```bash
rc scope create wms-new-feature --type persistent \
  --repos "wms-*,shared-libs"
rc start wms-new-feature
# Work on feature...
rc scope archive wms-new-feature  # When done
```

#### Hotfix
```bash
rc scope create oms-critical-fix --type ephemeral \
  --repos "oms-payment,oms-core"
rc start oms-critical-fix
# Fix issue...
rc scope delete oms-critical-fix  # Cleanup
```

#### Experimentation
```bash
rc scope create search-experiment --type ephemeral \
  --repos "search-*"
rc start search-experiment
# Try ideas...
rc scope delete search-experiment  # Discard
```

## Scope States

### Active
- Currently being worked on
- Claude Code session may be running
- Repositories are up-to-date

### Inactive
- Not currently in use
- Can be reactivated anytime
- Preserves all work

### Archived
- Work complete or abandoned
- Marked for potential deletion
- Can be restored if needed

## Custom Paths

You can customize the base path for workspaces:

```yaml
workspace:
  base_path: /absolute/path/to/workspaces
  # or
  base_path: custom-workspaces  # Relative to project root
```

## Migration from v1

If you have a v1 workspace structure:

### v1 Structure (shared workspace)
```
my-project/
└── workspace/
    ├── backend/
    ├── frontend/
    └── shared-libs/
```

### Migration Steps

1. Update `repo-claude.yaml`:
   ```yaml
   version: 2
   workspace:
     isolation_mode: true
     base_path: workspaces
   ```

2. Create scopes for your workflows:
   ```bash
   rc scope create main-dev --repos "*"
   ```

3. Repositories will be cloned into:
   ```
   my-project/
   └── workspaces/
       └── main-dev/
           ├── backend/
           ├── frontend/
           └── shared-libs/
   ```

See [Migration Guide](migration-guide.md) for detailed instructions.