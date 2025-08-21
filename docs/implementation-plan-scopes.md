# Scope-Based Architecture Implementation Plan

## Overview

This document outlines the implementation plan for transitioning repo-claude from an agent-based to a scope-based architecture. The implementation is divided into two phases to manage complexity and deliver value incrementally.

## Key Design Decisions

1. **Clean Break**: No migration support - users must recreate configurations in the new format
2. **Environment Variables**: Pass context to Claude sessions via environment variables (RC_SCOPE_ID, RC_SCOPE_NAME, RC_SCOPE_REPOS, RC_WORKSPACE_ROOT)
3. **Thin Orchestration**: Leverage Claude CLI directly without complex wrappers
4. **Working Directory**: Users launch rc from their actual working directory, not a separate workspace

## Phase 1: Core Simplification (Target: 1-2 weeks)

### 1.1 Remove Complexity from Start Command
**Files to modify:**
- `internal/manager/start_options.go`
- `cmd/repo-claude/app.go`

**Changes:**
- Remove `Foreground` and `NewWindow` fields from `StartOptions`
- Always use `createNewTerminalCommand()` logic
- Remove related command flags
- Pass environment variables to Claude session:
  - `RC_SCOPE_ID`: Unique identifier for the scope instance
  - `RC_SCOPE_NAME`: Human-readable scope name
  - `RC_SCOPE_REPOS`: Comma-separated list of repositories in scope
  - `RC_WORKSPACE_ROOT`: Root directory of the workspace

**Acceptance Criteria:**
- `rc start <scope>` always opens new terminal window
- No options for foreground/background/new-window
- Environment variables are correctly set in Claude session

### 1.2 Implement Numbered ps and kill Commands
**Files to modify:**
- `cmd/repo-claude/app.go` (rename stop to kill)
- `internal/manager/agents_list.go`
- `internal/manager/agents.go`

**Changes:**
- Add number column to ps output
- Implement `KillByNumber(num int)` method
- Support both `rc kill 1` and `rc kill backend` syntax

**Example Output:**
```
$ rc ps
#  SCOPE         STATUS  PID    REPOS
1  backend       🟢      12345  auth-service, order-service, payment-service
2  frontend      🟢      12346  web-app, mobile-app
3  order-flow    ⚫      -      (not running)
```

### 1.3 Redesign Configuration Schema
**Files to modify:**
- `internal/config/config.go`
- `internal/manager/interactive.go`
- All test files using config

**New Schema:**
```go
type Config struct {
    Workspace WorkspaceConfig `yaml:"workspace"`
    Scopes    map[string]Scope `yaml:"scopes"`
}

type Scope struct {
    Repos          []string `yaml:"repos"`
    Description    string   `yaml:"description"`
    Model          string   `yaml:"model"`
    AutoStart      bool     `yaml:"auto_start"`
    Dependencies   []string `yaml:"dependencies,omitempty"`
}
```

### 1.4 Update Start Command Logic
**Files to modify:**
- `internal/manager/start_options.go`
- `internal/manager/manager.go`

**Changes:**
- Accept scope names or repo names
- Map repos to appropriate working directory
- Generate appropriate system prompts

**Usage Examples:**
```bash
rc start backend        # Start backend scope
rc start order-service  # Start scope containing order-service
rc start                # Start all auto-start scopes
```

### 1.5 Update All User-Facing Messages
**Files to modify:**
- All files with user output

**Changes:**
- Replace "agent" with "scope" in messages
- Update help text and examples
- Update error messages

## Phase 2: Dynamic Scopes (Target: 2-3 weeks)

### 2.1 Design File-Based IPC Protocol
**New files:**
- `internal/ipc/protocol.go`
- `internal/ipc/watcher.go`

**Design:**
```
.repo-claude/
├── commands/
│   ├── scope-{id}.cmd     # Command files
│   └── scope-{id}.result  # Result files
└── state/
    └── scope-{id}.json    # Current scope state
```

**Command Format:**
```json
{
  "command": "change-scope",
  "scope": "order-flow",
  "timestamp": "2024-01-20T10:00:00Z"
}
```

### 2.2 Implement Command Injection
**New files:**
- `internal/claude/commands.go`

**Approach:**
- Generate `.claude/commands/rc-*.md` files
- Include IPC client code in CLAUDE.md
- Support `/rc:scope-name` syntax

**Example Command File:**
```markdown
---
command: rc:order-flow
description: Switch to order-flow scope
---

Switching to order-flow scope...

```bash
echo '{"command":"change-scope","scope":"order-flow"}' > $RC_WORKSPACE_ROOT/.repo-claude/commands/scope-$RC_SCOPE_ID.cmd
```
```

**CLAUDE.md Environment Context:**
```markdown
## Repo-Claude Context

You are working in a repo-claude managed session with the following context:
- Scope ID: ${RC_SCOPE_ID}
- Scope Name: ${RC_SCOPE_NAME}
- Repositories: ${RC_SCOPE_REPOS}
- Workspace Root: ${RC_WORKSPACE_ROOT}

Use these environment variables to understand your working context and coordinate with other scopes.
```

### 2.3 Implement Scope Watcher
**Files to modify:**
- `internal/manager/manager.go`

**Changes:**
- Start file watcher goroutine on launch
- Process commands from Claude sessions
- Update state files
- Handle errors gracefully

### 2.4 Update ps Command for Real-time Status
**Files to modify:**
- `internal/manager/agents_list.go`

**Changes:**
- Read current scope from state files
- Show scope changes with indicator
- Add timestamp of last change

**Example Output:**
```
$ rc ps
#  SCOPE         STATUS  PID    CURRENT SCOPE (REPOS)                          CHANGED
1  backend       🟢      12345  order-flow (order-service, payment-service) 📍  2m ago
2  frontend      🟢      12346  frontend (web-app, mobile-app)                  -
```

## Testing Strategy

### Phase 1 Tests:
1. Unit tests for new config schema
2. Integration tests for simplified start command
3. E2E tests for ps/kill functionality
4. Environment variable passing tests
5. Working directory context tests

### Phase 2 Tests:
1. IPC protocol unit tests
2. File watcher integration tests
3. Command injection tests
4. Scope change E2E tests

## Breaking Changes Guide

### For Users:
1. **No automatic migration** - manually recreate configuration in new format
2. Update any scripts using old commands
3. Review new scope definitions

### Config Format Changes:
```bash
# Old agent-based launch
rc start backend-agent frontend-agent

# New scope-based launch
rc start fullstack

# Launching from working directory
cd ~/my-project
rc start backend  # Uses current directory as workspace root
```

## Success Metrics

### Phase 1:
- Simplified commands work reliably
- Configuration is more intuitive
- No regression in core functionality

### Phase 2:
- Scope changes work within 100ms
- IPC is reliable and debuggable
- Clear documentation for dynamic features

## Rollback Plan

### Phase 1:
- Tag release before changes as v1.x (agent-based)
- Create v2.0 for scope-based architecture
- Document breaking changes clearly
- Users can install specific version if needed

### Phase 2:
- Feature flag for dynamic scopes
- File-based IPC can be disabled
- Fallback to static scopes