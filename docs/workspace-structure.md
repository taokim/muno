# Workspace Structure

## Overview

Repo-Claude organizes your multi-repository project with a clear separation between project configuration and repository workspace.

## Default Structure

```
my-project/                     # Project root
├── repo-claude                # Executable (copied during init)
├── repo-claude.yaml           # Configuration file
├── .repo-claude-state.json    # Agent runtime state tracking (JSON)
└── workspace/                 # Contains all managed git repositories
    ├── shared-memory.md       # Agent coordination file
    ├── backend/               # Git repository (trunk/main branch)
    │   └── CLAUDE.md         # Agent-specific context for backend-agent
    ├── frontend/              # Git repository (trunk/main branch)
    │   └── CLAUDE.md         # Agent-specific context for frontend-agent
    ├── mobile/                # Git repository (trunk/main branch)
    │   └── CLAUDE.md         # Agent-specific context for mobile-agent
    └── shared-libs/           # Git repository (trunk/main branch)
        └── CLAUDE.md         # Shared context (no dedicated agent)
```

## Custom Workspace Path

You can override the default `workspace` subdirectory in your configuration:

```yaml
workspace:
  name: my-project
  path: code  # Use 'code' instead of 'workspace'
  # or
  # path: /absolute/path/to/workspace  # Absolute path
  manifest:
    # ... rest of config
```

## Benefits

1. **Clean Separation**: Configuration stays at project root, code in workspace
2. **Portable**: Can move workspace without affecting configuration
3. **Organized**: All repositories grouped in one location
4. **Flexible**: Configure workspace path as needed

## Configuration Example

```yaml
workspace:
  name: my-workspace
  path: workspace  # Optional, defaults to 'workspace'
  manifest:
    remote_name: origin
    remote_fetch: git@github.com:myorg
    default_revision: main
    projects:
      - name: backend-repo
        path: services/backend  # Path within workspace
        groups: services,backend
        agent: backend-agent
```

In this example:
- Project root: `./my-workspace/`
- Configuration: `./my-workspace/repo-claude.yaml`
- Repositories: `./my-workspace/workspace/services/backend/`
- Shared memory: `./my-workspace/workspace/shared-memory.md`