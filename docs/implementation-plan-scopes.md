# Scope-Based Architecture Implementation

## Overview

This document describes the completed implementation of the scope-based architecture for repo-claude. The transition from agent-based to scope-based architecture was successfully completed in August 2025.

## Implemented Features

### Phase 1: Core Scope-Based Architecture ✅

#### 1.1 Configuration Schema Redesign
- Transitioned from agent-based to scope-based configuration
- Maintained backward compatibility with legacy agent configurations
- Introduced flexible repository grouping with wildcard support

**Configuration Structure:**
```yaml
workspace:
  name: my-project
  manifest:
    projects:
      - name: backend-service
        url: https://github.com/myorg/backend
        groups: backend,services

scopes:
  backend:
    repos: ["backend-*", "auth-service"]
    description: "Backend services development"
    model: claude-sonnet-4
    auto_start: true
```

#### 1.2 Enhanced ps and kill Commands
- Added numbered output for easy reference
- Implemented `rc kill <number>` functionality
- Maintained support for `rc kill <scope-name>`

**Example Output:**
```
$ rc ps
#  SCOPE         STATUS  PID    CURRENT SCOPE (REPOS)                          CHANGED
1  backend       🟢      12345  backend (auth-service, order-service)           -
2  frontend      🟢      12346  frontend (web-app, mobile-app)                  -
```

#### 1.3 Terminal Integration
- Default behavior: Opens in new terminal tab
- Fallback: Opens in new window if tab creation fails
- Added `--new-window` flag for explicit window creation

#### 1.4 Environment Variables
Successfully implemented environment variable passing to Claude sessions:
- `RC_SCOPE_ID`: Unique identifier for the scope instance
- `RC_SCOPE_NAME`: Human-readable scope name  
- `RC_SCOPE_REPOS`: Comma-separated list of repositories in scope
- `RC_WORKSPACE_ROOT`: Root directory of the workspace

### Phase 2: Dynamic Scope Tracking ✅

#### 2.1 Extended State Management
- Added `CurrentRepos` and `CurrentScope` fields to track dynamic scope changes
- Implemented `LastChange` timestamp for tracking scope modifications
- Created `ChangeScopeContext` method for future dynamic updates

**State Structure:**
```go
type ScopeStatus struct {
    Name         string
    Status       string
    PID          int
    Repos        []string // Initial repositories
    CurrentRepos []string // Current repositories (can change)
    CurrentScope string   // Current scope name
    LastActivity string
    LastChange   string   // When scope was last changed
}
```

#### 2.2 Enhanced ps Command Display
- Shows current scope with 📍 indicator when changed
- Displays time since last scope change
- Properly formats long repository lists with truncation

### Additional Improvements

#### Version Management
- Local builds now use YYMMDDHHMM format (e.g., `2508211443`)
- Tagged releases use git tags
- Version displayed via `rc --version`

#### Binary Location
- Build target: `./bin/rc`
- Created by `make build`
- Can be installed to `$GOPATH/bin/rc` via `make install`

#### Testing
- Maintained >80% test coverage for core packages
- Added comprehensive tests for new state tracking functionality
- Updated CLAUDE.md with testing guidelines

## Breaking Changes

### Configuration Format
Users must manually migrate from agent-based to scope-based configuration:

**Old Format (agent-based):**
```yaml
agents:
  backend-agent:
    repository: backend
    specialization: "Backend development"
```

**New Format (scope-based):**
```yaml
scopes:
  backend:
    repos: ["backend", "services/*"]
    description: "Backend development"
```

### Command Changes
- `rc stop` → `rc kill`
- Removed foreground/background options from start command
- Terminal tab is now the default (was new window)

## Future Considerations

While the core scope-based architecture is complete, the following features were descoped:
- Dynamic scope switching via `/rc:` commands
- File-based IPC for real-time scope updates
- Command injection into Claude sessions

These features can be added in future releases if needed.

## Migration Guide

1. Back up your existing `repo-claude.yaml`
2. Create a new configuration using scope-based format
3. Update any scripts that use old command syntax
4. Test with `rc start <scope-name>`

## Success Metrics Achieved

- ✅ Simplified command structure
- ✅ More intuitive configuration format
- ✅ Better support for multi-repository workflows
- ✅ Terminal tab integration for improved UX
- ✅ Maintained backward compatibility where possible
- ✅ >80% test coverage maintained