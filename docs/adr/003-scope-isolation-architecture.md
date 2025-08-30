# ADR 004: Scope Isolation Architecture

Status: Proposed
Date: 2024-12-24
Author: System Architect

## Context

The current repo-claude implementation shares a single workspace across all scopes, where all repositories are checked out once and shared by all Claude sessions. This approach has several limitations:

1. **No Isolation**: Changes made in one scope immediately affect all other scopes
2. **Branch Conflicts**: Multiple scopes working on different features can't use different branches
3. **Testing Challenges**: Can't test different versions or configurations simultaneously
4. **State Confusion**: Hard to track which scope made which changes
5. **Cleanup Difficulty**: No clear separation for temporary work vs. persistent changes

## Problem Statement

We need a way to:
- Isolate work between different scopes/sessions
- Support both long-lived and ephemeral workspaces
- Maintain cross-repository documentation and shared resources
- Enable parallel development on different features
- Simplify cleanup and state management

## Decision

Redesign repo-claude to use **isolated scope workspaces** where each scope gets its own directory with independent repository checkouts.

### Core Design Principles

1. **Scope Isolation**: Each scope operates in its own isolated directory
2. **Lazy Checkout**: Repositories are cloned only when a scope is first started
3. **Scope-First Commands**: All commands require scope context
4. **Shared Documentation**: Cross-repo docs stored in hierarchical structure
5. **Flexible Lifecycle**: Support both persistent and ephemeral scopes

### Architecture Overview

```
project-root/
├── repo-claude.yaml          # Global configuration
├── .repo-claude-state.json   # Global state tracking
├── docs/                     # Cross-repository documentation
│   ├── global/              # Shared across all scopes
│   │   ├── architecture.md
│   │   └── standards.md
│   └── scopes/              # Scope-specific docs
│       ├── order-service/
│       │   └── design.md
│       └── frontend/
│           └── ui-patterns.md
└── workspaces/              # Isolated scope workspaces
    ├── order-service/       # Pre-defined scope
    │   ├── .scope-config.yaml
    │   ├── shared-memory.md
    │   ├── auth-service/    # Cloned repository
    │   ├── order-service/   # Cloned repository
    │   └── payment-service/ # Cloned repository
    └── ad-hoc-fix-123/     # Ad-hoc scope
        ├── .scope-config.yaml
        ├── shared-memory.md
        └── order-service/   # Only needed repo
```

## Detailed Design

### 1. Scope Types

#### Pre-defined Scopes
- Configured in `repo-claude.yaml`
- Long-lived workspaces for ongoing development
- Example: `order-service`, `frontend`, `fullstack`

#### Ad-hoc Scopes
- Created interactively or via command
- Short-lived for specific tasks
- Auto-generated names or user-specified
- Example: `hotfix-auth-bug`, `feature-checkout-v2`

### 2. Scope Configuration

Global configuration (`repo-claude.yaml`):
```yaml
workspace:
  name: my-project
  base_path: workspaces  # Base directory for all scope workspaces
  
repositories:
  auth-service:
    url: git@github.com:org/auth-service.git
    default_branch: main
  order-service:
    url: git@github.com:org/order-service.git
    default_branch: main
  payment-service:
    url: git@github.com:org/payment-service.git
    default_branch: main
  frontend:
    url: git@github.com:org/frontend.git
    default_branch: develop

scopes:
  order-service:
    type: persistent
    repos: [order-service, payment-service, auth-service]
    description: "Order processing flow development"
    model: claude-3-sonnet
    auto_start: false
    
  frontend:
    type: persistent
    repos: [frontend, shared-ui]
    description: "Frontend development"
    model: claude-3-sonnet
    
  hotfix:
    type: ephemeral
    repos: []  # Dynamically selected
    description: "Template for hotfixes"
    # Note: No TTL - manual lifecycle management
```

Scope-specific configuration (`.scope-config.yaml`):
```yaml
scope_id: order-service-2024-12-24-001
scope_name: order-service
created_at: 2024-12-24T10:00:00Z
last_accessed: 2024-12-24T15:30:00Z
type: persistent
repos:
  - name: order-service
    branch: feature/new-checkout
    commit: abc123def
  - name: payment-service
    branch: main
    commit: 456789ghi
state: active
```

### 3. Command Structure

All commands now require scope context:

```bash
# Scope management
rc init <project-name>              # Initialize project
rc scope create <name> [--repos]    # Create new scope
rc scope list                       # List all scopes
rc scope delete <name>              # Delete scope and its workspace

# Start/stop operations
rc start <scope> [--new-window]     # Start Claude session for scope
rc start --interactive              # Interactive scope creation
rc stop <scope>                    # Stop scope session

# Git operations (scope-specific)
rc pull <scope> [--clone-missing]   # Pull/clone repos for scope
rc status <scope>                  # Status of repos in scope
rc commit <scope> -m "message"     # Commit changes in scope
rc push <scope>                    # Push changes from scope
rc branch <scope> <branch-name>    # Create/switch branch in scope

# Documentation
rc docs edit <scope> <file>        # Edit scope documentation
rc docs list [<scope>]             # List documentation
rc docs sync                       # Sync docs to git
```

### 4. Scope Lifecycle

#### Creation Flow
1. User runs `rc start order-service` or `rc scope create order-service`
2. System checks if `workspaces/order-service/` exists
3. If not, creates directory and clones specified repositories
4. Creates `.scope-config.yaml` with metadata
5. Initializes `shared-memory.md` for the scope
6. Launches Claude session with scope context

#### Activation Flow
1. User runs `rc start order-service` for existing scope
2. System verifies workspace exists
3. Optionally pulls latest changes (`--pull` flag)
4. Updates last_accessed timestamp
5. Launches Claude session

#### Cleanup Flow
1. Ephemeral scopes: Manual cleanup when needed
2. Manual cleanup: `rc scope delete <name>`
3. Archive option: `rc scope archive <name>` (compress and store)

### 5. Cross-Repository Documentation

To address the challenge of shared documentation:

#### Hierarchical Documentation Structure
```
docs/
├── global/                 # Version-controlled global docs
│   ├── README.md
│   ├── architecture/
│   └── standards/
├── scopes/                # Scope-specific docs
│   ├── order-service/
│   └── frontend/
└── .gitignore            # Ignore ephemeral scope docs
```

#### Documentation Commands
```bash
rc docs create global architecture.md   # Create global doc
rc docs create order-service flow.md    # Create scope doc
rc docs edit order-service flow.md      # Edit scope doc
rc docs sync                            # Commit docs to git
```

#### Access from Claude Sessions
- Environment variable: `RC_DOCS_PATH=/path/to/project/docs`
- Scopes can read both global and their own docs
- Special syntax in shared-memory.md for doc references

### 6. Migration Strategy

#### Phase 1: Backward Compatibility (2 weeks)
- Support both old (shared workspace) and new (isolated) modes
- Add `--isolated` flag for new behavior
- Deprecation warnings for old mode

#### Phase 2: Migration Tools (1 week)
- `rc migrate` command to convert existing workspaces
- Automated scope creation from current workspace state
- Documentation migration utilities

#### Phase 3: Default Switch (1 week)
- Make isolated mode default
- Require `--shared` flag for old behavior
- Update all documentation

#### Phase 4: Cleanup (2 weeks)
- Remove shared workspace code
- Finalize API and commands
- Release v2.0

## Consequences

### Positive
- **True Isolation**: Each scope has independent state and branches
- **Parallel Development**: Multiple features can progress simultaneously
- **Clean State**: Easy to reset or cleanup individual scopes
- **Better Testing**: Can test different configurations in parallel
- **Resource Efficiency**: Only clone repos when needed
- **Clear Ownership**: Each scope tracks its own changes

### Negative
- **Disk Usage**: Multiple clones of repositories
- **Initial Setup Time**: First scope start requires cloning
- **Documentation Complexity**: Need new system for shared docs
- **Learning Curve**: Users must adapt to scope-first commands
- **Migration Effort**: Existing users need to migrate

### Neutral
- **Command Length**: Commands now require scope parameter
- **Mental Model Shift**: From shared to isolated workspaces

## Alternatives Considered

### 1. Git Worktrees
- Use git worktrees instead of separate clones
- Pros: Less disk space, shared git objects
- Cons: Complex management, limited tool support

### 2. Docker Containers
- Each scope runs in a container
- Pros: Complete isolation, reproducible
- Cons: Overhead, complexity, requires Docker

### 3. Branch-based Isolation
- Use branches for isolation in shared workspace
- Pros: Simple, less disk usage
- Cons: Branch conflicts, no true isolation

## Implementation Plan

### Week 1: Core Infrastructure
- [ ] Design scope workspace structure
- [ ] Implement scope creation/deletion
- [ ] Add scope-aware git operations

### Week 2: Command Updates
- [ ] Update all commands for scope context
- [ ] Implement interactive scope creation
- [ ] Add scope lifecycle management

### Week 3: Documentation System
- [ ] Design hierarchical doc structure
- [ ] Implement doc commands
- [ ] Add doc synchronization

### Week 4: Migration & Testing
- [ ] Create migration tools
- [ ] Add comprehensive tests
- [ ] Update user documentation

### Week 5: Polish & Release
- [ ] Performance optimization
- [ ] Error handling improvements
- [ ] Release candidate testing

## Open Questions

1. **Scope Naming**: Should we enforce naming conventions?
2. **Resource Limits**: Should we limit number of concurrent scopes?
3. **Shared Dependencies**: How to handle shared libraries/modules?
4. **Remote Sync**: Should scope state sync to remote?
5. **Scope Templates**: Should we support scope templates?

## Decision

**Status**: Proposed for implementation

This design provides the isolation and flexibility needed while maintaining the ability to share documentation and resources across scopes. The migration path ensures smooth transition for existing users.