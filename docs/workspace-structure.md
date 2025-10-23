# Workspace Structure

## Overview

MUNO uses a tree-based workspace structure where repositories are organized in a navigable hierarchy. Each workspace contains configuration, state tracking, and a tree of repositories.

## Directory Layout

```
my-platform/                    # Workspace root
├── muno.yaml            # Configuration file
├── .muno-state.json     # Tree state and navigation tracking
├── nodes/                      # Repository tree root
│   ├── team-backend/           # Parent node (can be a git repo)
│   │   ├── .git/              # Optional git repository
│   │   ├── payment-service/   # Child repository
│   │   │   ├── .git/
│   │   │   └── src/
│   │   ├── order-service/     # Child repository
│   │   │   ├── .git/
│   │   │   └── src/
│   │   └── shared-libs/        # Lazy repository (not cloned)
│   └── team-frontend/          # Parent node
│       ├── .git/
│       ├── web-app/           # Child repository
│       │   ├── .git/
│       │   └── src/
│       └── component-lib/      # Lazy repository
└── CLAUDE.md                   # AI context file

```

## File Descriptions

### `muno.yaml`

Main configuration file defining the workspace:

```yaml
workspace:
  name: my-platform
  root_path: repos
tree:
  - name: team-backend
    url: https://github.com/org/backend-team
    children:
      - name: payment-service
        url: https://github.com/org/payment-service
      - name: order-service
        url: https://github.com/org/order-service
      - name: shared-libs
        url: https://github.com/org/shared-libs
        lazy: true
  - name: team-frontend
    url: https://github.com/org/frontend-team
    children:
      - name: web-app
        url: https://github.com/org/web-app
      - name: component-lib
        url: https://github.com/org/component-lib
        lazy: true
```

### `.muno-state.json`

Runtime state tracking:

```json
{
  "current_path": "/team-backend/payment-service",
  "tree_state": {
    "/": {
      "status": "root",
      "children": ["team-backend", "team-frontend"]
    },
    "/team-backend": {
      "status": "cloned",
      "children": ["payment-service", "order-service", "shared-libs"]
    },
    "/team-backend/payment-service": {
      "status": "cloned",
      "branch": "main",
      "last_pulled": "2024-01-15T10:30:00Z"
    },
    "/team-backend/shared-libs": {
      "status": "lazy",
      "url": "https://github.com/org/shared-libs"
    }
  },
  "navigation_history": [
    "/",
    "/team-backend",
    "/team-backend/payment-service"
  ]
}
```

### `nodes/` Directory

The repository tree root where all repositories are organized:

- **Parent Nodes**: Can be regular directories or git repositories
- **Child Repositories**: Always git repositories (unless lazy)
- **Lazy Repositories**: Placeholder entries, not cloned until accessed

### `CLAUDE.md`

Context file for Claude Code sessions, automatically generated and updated:

```markdown
# Claude Context

You are working in a tree-based multi-repository workspace.

## Current Position
You are at: /team-backend/payment-service

## Available Repositories
- payment-service (current)
- order-service (../order-service)
- shared-libs (../shared-libs) [lazy - not cloned]
- web-app (../../team-frontend/web-app)

## Navigation
Navigate using standard filesystem commands.
Current directory operations affect the current node.
```

## Tree Organization

### Node Types

1. **Root Node** (`/`)
   - Top of the tree
   - Contains team or project nodes
   - Usually not a git repository

2. **Parent Nodes**
   - Can contain child repositories
   - May be git repositories themselves
   - Organize related repositories

3. **Leaf Nodes**
   - Always git repositories (when cloned)
   - Cannot contain children
   - Where actual development happens

### Repository States

- **cloned**: Repository is present and ready
- **lazy**: Repository URL stored but not cloned
- **modified**: Repository has uncommitted changes
- **missing**: Expected repository not found

## Navigation Paths

### Absolute Paths
Start from root:
- `/team-backend`
- `/team-backend/payment-service`
- `/team-frontend/web-app`

### Relative Paths
From current position:
- `payment-service` (child)
- `../order-service` (sibling)
- `../../team-frontend` (different branch)
- `..` (parent)
- `.` (current)

### Special Paths
- `/` - Root of tree
- `-` - Previous position
- `~` - Root (alias)

## Lazy Loading

Repositories marked as lazy are not cloned initially:

1. **Storage Efficiency**: Save disk space
2. **Faster Setup**: Quick workspace initialization
3. **On-Demand Access**: Clone when needed

Auto-clone triggers:
- Navigating to a lazy repository
- Executing operations on lazy nodes
- Manual `muno clone` command

## Best Practices

### Tree Organization

1. **Group by Team**: Top-level nodes for teams
2. **Service Grouping**: Related services under same parent
3. **Shared Resources**: Common libraries at appropriate level
4. **Lazy Loading**: Mark optional/large repos as lazy

### Naming Conventions

- Use descriptive, lowercase names
- Separate words with hyphens
- Avoid special characters
- Keep names consistent with repository names

### Repository Management

1. **Regular Pulls**: Keep repositories updated
2. **Clean States**: Commit or stash before navigation
3. **Branch Consistency**: Maintain branch alignment
4. **Lazy Strategy**: Mark infrequently used repos as lazy

## Example Structures

### Microservices Platform
```
platform/
├── backend-services/
│   ├── auth-service/
│   ├── user-service/
│   ├── payment-service/
│   └── notification-service/
├── frontend-apps/
│   ├── customer-web/
│   ├── admin-portal/
│   └── mobile-app/
└── shared/
    ├── common-libs/
    └── proto-definitions/
```

### Monorepo with Services
```
company/
├── core-platform/        # Main monorepo
│   ├── services/
│   └── packages/
├── mobile-apps/          # Separate mobile repos
│   ├── ios-app/
│   └── android-app/
└── infrastructure/       # DevOps repos
    ├── terraform/
    └── kubernetes/
```

### Full-Stack Application
```
app/
├── backend/
│   ├── api-gateway/
│   ├── core-api/
│   └── worker-services/
├── frontend/
│   ├── web-app/
│   └── component-library/
└── tools/
    ├── cli/
    └── sdk/
```