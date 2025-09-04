# Distributed Configuration Examples

This directory contains examples of the new distributed configuration system in MUNO.

## Key Concepts

### 1. Two Node Types

Nodes in the configuration can be one of two types (mutually exclusive):

- **Repository nodes**: Have a `url` field, represent actual git repositories
- **Config delegation nodes**: Have a `config` field, delegate to another configuration file

```yaml
# Repository node
- name: my-service
  url: https://github.com/org/my-service.git
  lazy: true  # Optional, controls clone behavior

# Config delegation node  
- name: sub-workspace
  config: path/to/config.yaml  # Delegates configuration
```

### 2. Smart Lazy Defaults

Repositories are lazy (clone on-demand) by default, but meta-repositories are automatically eager:

- Names ending with `-monorepo`, `-muno`, `-metarepo`, `-repo` → `lazy: false` (eager)
- All other repositories → `lazy: true` (on-demand)
- Can be explicitly overridden with `lazy: true/false`

### 3. Auto-Discovery

When a repository is cloned, MUNO automatically looks for `muno.yaml` in its root:

1. Clone `team-backend` repository
2. Find `team-backend/muno.yaml` 
3. Automatically load and process its child nodes
4. Creates a recursive tree structure

### 4. Ownership Boundaries

Each configuration only declares its direct children:

- ✅ Parent declares immediate children
- ❌ Parent does NOT declare grandchildren
- ✅ Each level manages only its own repositories

## Example Structure

```
my-platform/                    # Root workspace
├── muno.yaml                   # Root configuration
├── nodes/
│   ├── team-backend/          # Repository with auto-discovered config
│   │   ├── .git/
│   │   ├── muno.yaml          # Auto-discovered, declares children
│   │   └── services/          # Team's repos directory
│   │       ├── payment-service/
│   │       └── order-service/
│   ├── team-frontend/         # Repository with config
│   │   ├── .git/
│   │   └── muno.yaml
│   └── infrastructure/        # Config-only node (no .git)
│       └── .muno-config       # Marker file for config delegation
```

## File Examples

### root-muno.yaml
The root workspace configuration that sets up the top-level structure.

### team-backend-muno.yaml  
Example of a configuration that would be placed inside the team-backend repository as `muno.yaml`.

### infrastructure-config.yaml
Example of a pure configuration file (no repository at this level) that organizes infrastructure repos.

## Benefits

1. **No State Files**: Everything derived from filesystem and configs
2. **Distributed Ownership**: Each team owns their configuration
3. **Auto-Discovery**: Configs found automatically when repos cloned
4. **Smart Defaults**: Meta-repos eager, services lazy by default
5. **Clear Boundaries**: Each config only knows its direct children

## Usage

```bash
# Initialize workspace with the root config
muno init my-platform

# Navigate the tree (auto-clones as needed)
muno use team-backend
muno use team-backend/payment-service

# The tree reflects the distributed structure
muno tree

# Add new repositories at current level
muno add https://github.com/org/new-service.git

# Status shows the full distributed tree
muno status --recursive
```