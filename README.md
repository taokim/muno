# Repo-Claude

Transform your multi-repository chaos into a unified development experience with AI-powered scopes.

## The Problem

While Claude and other AI coding assistants excel at working within a single repository, modern software development often spans multiple repositories:
- E-commerce platforms with separate services (OMS, WMS, Catalog, Search)
- Microservices architectures with dozens of service repos
- Frontend/backend separation across different repositories  
- Shared libraries and components in separate repos

**The Challenge**: AI agents struggle to understand and work across repository boundaries, missing critical context and dependencies that span multiple repos.

## The Solution

Repo-Claude v2 brings isolated scope-based development to multi-repository management:

### 1. **Isolated Workspaces**
Each scope operates in its own isolated workspace directory:
- Complete separation between different development contexts
- Parallel work on different features without interference
- Clean environment for each scope instance
- Persistent and ephemeral scope types for different workflows

### 2. **Scope-Based AI Development**
Work with flexible scopes that span multiple repositories:
- Define scopes by domain (e.g., "wms", "oms", "search", "catalog")
- Each scope includes all relevant repositories for complete context
- Claude Code sessions understand the full scope of your work
- Isolated workspaces prevent cross-contamination between projects

## Features

- 🚀 **Single Binary**: No runtime dependencies (Python, Node.js, etc.)
- 🗂️ **Multi-Repository Management**: Manage dozens of Git repositories as isolated scopes
- 🤖 **Scope-Based Orchestration**: Launch Claude Code sessions with multi-repository context
- 🔒 **Workspace Isolation**: Each scope has its own isolated workspace directory
- 🎨 **Interactive TUI**: Beautiful terminal UI for selecting scopes and repos (powered by Bubbletea)
- 🔧 **Simple Git Operations**: Direct git commands with parallel execution for speed
- 📝 **Shared Memory**: Cross-scope coordination through shared memory file
- ⚡ **Fast**: Written in Go for optimal performance
- 🎯 **Easy Configuration**: Single YAML file controls everything

## Key Features

### 🚀 Multi-Repository Orchestration
- **Isolated workspaces** for each scope instance
- **Scope-based AI sessions** that see multiple repos as one project
- **Parallel operations** across all repositories
- **Shared memory** for cross-scope coordination

### 🔒 Workspace Isolation (v2 Architecture)
- **Dedicated directories** for each scope in `workspaces/`
- **Scope metadata** tracking in `.scope-meta.json`
- **Repository state management** per scope
- **Clean separation** between different development contexts

### 🌿 Scope Management
- **Create scopes** from templates or custom configurations
- **Archive inactive scopes** for later reference
- **Delete obsolete scopes** to free up space
- **List all scopes** with status information

### 📋 Git Operations
- **Per-scope operations** with isolated Git state
- **Parallel execution** for fast repository updates
- **Conflict detection** before operations
- **Clean history** with rebase by default

## Prerequisites

- Git
- [Claude Code CLI](https://claude.ai/code)

## Installation

### From Source

```bash
git clone https://github.com/taokim/repo-claude.git
cd repo-claude
make build
sudo make install
```

## Quick Start

1. **Initialize a new workspace**:
   ```bash
   rc init my-ecommerce-platform
   cd my-ecommerce-platform
   ```

2. **Create and start scopes**:
   ```bash
   # Create a new scope for WMS development
   rc scope create wms-dev --type persistent --repos "wms-*,shared-libs"
   
   # Start the scope
   rc start wms-dev
   
   # Or use interactive mode
   rc start              # Interactive selection UI
   ```

3. **Manage scopes**:
   ```bash
   rc list               # List all scopes
   rc status             # Show workspace status
   rc scope archive wms-dev  # Archive when done
   ```

4. **Git operations within scope**:
   ```bash
   rc pull --scope wms-dev           # Pull updates for scope repos
   rc commit --scope wms-dev -m "Update inventory logic"
   rc push --scope wms-dev
   ```

## Example: E-Commerce Platform

Imagine you're building an e-commerce platform with:
- **WMS Services**: `wms-core`, `wms-inventory`, `wms-shipping`, `wms-ui`
- **OMS Services**: `oms-core`, `oms-payment`, `oms-fulfillment`, `oms-ui`
- **Search Services**: `search-engine`, `search-indexer`, `search-ui`
- **Catalog Services**: `catalog-service`, `catalog-admin`, `catalog-api`
- **Shared Components**: `shared-libs`, `api-gateway`, `web-storefront`

With Repo-Claude v2's isolated workspaces:
- **WMS Scope**: Complete warehouse management context in `workspaces/wms-feature-123/`
- **OMS Scope**: Order processing workflow in `workspaces/oms-payment-fix/`
- **Search Scope**: Search optimization work in `workspaces/search-perf-improvement/`

Each scope is completely isolated, allowing parallel development without conflicts.

## Configuration

The `repo-claude.yaml` file defines your workspace (v2 structure):

```yaml
version: 2
workspace:
  name: my-ecommerce-platform
  isolation_mode: true      # Default: true (v2 behavior)
  base_path: workspaces     # Where isolated scopes are created

repositories:
  # WMS (Warehouse Management System) components
  wms-core:
    url: https://github.com/yourorg/wms-core.git
    default_branch: main
    groups: [wms, backend, core]
  
  wms-inventory:
    url: https://github.com/yourorg/wms-inventory.git
    default_branch: main
    groups: [wms, backend, inventory]
  
  wms-shipping:
    url: https://github.com/yourorg/wms-shipping.git
    default_branch: main
    groups: [wms, backend, logistics]
    
  wms-ui:
    url: https://github.com/yourorg/wms-ui.git
    default_branch: main
    groups: [wms, frontend, ui]
  
  # OMS (Order Management System) components
  oms-core:
    url: https://github.com/yourorg/oms-core.git
    default_branch: main
    groups: [oms, backend, core]
    
  oms-payment:
    url: https://github.com/yourorg/oms-payment.git
    default_branch: main
    groups: [oms, backend, payment]
    
  # Search components
  search-engine:
    url: https://github.com/yourorg/search-engine.git
    default_branch: main
    groups: [search, backend, core]
    
  # Catalog components
  catalog-service:
    url: https://github.com/yourorg/catalog-service.git
    default_branch: main
    groups: [catalog, backend, core]
    
  # Shared components
  shared-libs:
    url: https://github.com/yourorg/shared-libs.git
    default_branch: main
    groups: [shared, core]
    
  api-gateway:
    url: https://github.com/yourorg/api-gateway.git
    default_branch: main
    groups: [shared, backend, gateway]

scopes:
  # Persistent scopes for long-term development
  wms:
    type: persistent
    repos: ["wms-*", "shared-libs"]
    description: "Warehouse Management System - inventory, shipping, logistics"
    model: claude-3-5-sonnet-20241022
    auto_start: false
  
  oms:
    type: persistent
    repos: ["oms-*", "shared-libs", "api-gateway"]
    description: "Order Management System - processing, payments, fulfillment"
    model: claude-3-5-sonnet-20241022
    auto_start: false
  
  search:
    type: persistent
    repos: ["search-*", "catalog-api", "shared-libs"]
    description: "Search and Discovery - engine, indexing, relevance"
    model: claude-3-5-sonnet-20241022
    auto_start: false
  
  catalog:
    type: persistent
    repos: ["catalog-*", "search-indexer", "shared-libs"]
    description: "Product Catalog Management - products, categories, attributes"
    model: claude-3-5-sonnet-20241022
    auto_start: false
  
  # Ephemeral scope templates
  hotfix:
    type: ephemeral
    repos: []  # Select repos at creation time
    description: "Emergency hotfix - select affected repos"
    model: claude-3-5-sonnet-20241022
    auto_start: false
  
  feature:
    type: ephemeral
    repos: []  # Select repos at creation time
    description: "Feature development - cross-service implementation"
    model: claude-3-5-sonnet-20241022
    auto_start: false

documentation:
  path: docs
  sync_to_git: true
```

### Configuration Reference (v2)

#### Workspace Configuration
| Key | Type | Required | Default | Description |
|-----|------|----------|---------|-------------|
| `version` | int | Yes | 2 | Configuration version |
| `workspace.name` | string | Yes | - | Name of your workspace |
| `workspace.isolation_mode` | bool | No | true | Enable isolated workspaces |
| `workspace.base_path` | string | No | "workspaces" | Directory for isolated scopes |

#### Repository Configuration
| Key | Type | Required | Default | Description |
|-----|------|----------|---------|-------------|
| `url` | string | Yes | - | Git repository URL |
| `default_branch` | string | Yes | - | Default branch to use |
| `groups` | array | No | [] | Repository groups for filtering |

#### Scope Configuration
| Key | Type | Required | Default | Description |
|-----|------|----------|---------|-------------|
| `type` | string | Yes | - | "persistent" or "ephemeral" |
| `repos` | array | Yes | - | Repository names or patterns (wildcards) |
| `description` | string | Yes | - | Human-readable description |
| `model` | string | Yes | - | Claude model to use |
| `auto_start` | boolean | No | false | Start automatically |

### Scope Types

**Persistent Scopes**: Long-lived development environments
- Remain available across sessions
- Ideal for ongoing feature development
- Can be archived when complete

**Ephemeral Scopes**: Temporary workspaces
- Created for specific tasks (hotfixes, experiments)
- Easy cleanup when done
- No long-term maintenance

## Workspace Structure (v2)

After initialization and scope creation:

```
my-ecommerce-platform/
├── repo-claude.yaml         # Configuration
├── .repo-claude-state.json  # State tracking
├── shared-memory.md         # Cross-scope coordination
├── docs/                    # Global documentation
│   ├── global/             # Project-wide docs
│   └── scopes/             # Scope-specific docs
└── workspaces/             # Isolated scope directories
    ├── wms-feature-123/    # Isolated WMS scope
    │   ├── .scope-meta.json  # Scope metadata
    │   ├── wms-core/         # Cloned repository
    │   │   └── CLAUDE.md     # Context for AI
    │   ├── wms-inventory/    # Cloned repository
    │   │   └── CLAUDE.md
    │   ├── wms-shipping/     # Cloned repository
    │   │   └── CLAUDE.md
    │   └── shared-libs/      # Cloned repository
    │       └── CLAUDE.md
    └── oms-payment-fix/    # Isolated OMS scope
        ├── .scope-meta.json
        ├── oms-core/
        │   └── CLAUDE.md
        ├── oms-payment/
        │   └── CLAUDE.md
        └── shared-libs/
            └── CLAUDE.md
```

## Commands Reference

### Scope Management
```bash
rc scope create <name> [options]  # Create new scope
rc scope delete <name>            # Delete scope
rc scope archive <name>           # Archive scope
rc list                           # List all scopes
```

### Scope Operations
```bash
rc start [scope]                  # Start scope (interactive if no args)
rc status                         # Show workspace status
rc pull --scope <name>            # Pull repos in scope
rc commit --scope <name> -m "msg" # Commit in scope
rc push --scope <name>            # Push scope changes
```

### Documentation
```bash
rc docs create --global <name>    # Create global doc
rc docs create --scope <scope>    # Create scope doc
rc docs list                      # List all docs
rc docs sync                      # Sync to Git
```

## Migration from v1

If you have an existing v1 workspace:

1. **Backup your current workspace**
2. **Update configuration**:
   - Add `version: 2` at the top
   - Add `workspace.isolation_mode: true`
   - Add `workspace.base_path: "workspaces"`
3. **Create scopes** for your existing workflows
4. **Clone repositories** into the new isolated scopes

See [Migration Guide](docs/migration-guide.md) for detailed instructions.

## Development

### Building
```bash
make build          # Build binary
make test           # Run tests
make coverage       # Generate coverage report
make lint           # Run linter
```

### Testing
```bash
go test ./...                    # Run all tests
go test -v ./internal/scope/...  # Test specific package
make coverage                    # Coverage report
```

## Architecture

Repo-Claude v2 uses a three-level architecture:

1. **Project Level**: Configuration and shared resources
2. **Scope Level**: Isolated workspace directories
3. **Repository Level**: Individual Git repositories

Key components:
- **ScopeManager**: Creates and manages isolated scopes
- **GitManager**: Handles Git operations within scopes
- **ConfigManager**: Manages v2 configuration schema
- **DocsManager**: Handles documentation system

See [Architecture Documentation](docs/architecture.md) for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details