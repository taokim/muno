# ADR 003: Scope-Based Architecture

Status: Proposed
Date: 2024-01-20

## Context

The current repo-claude implementation uses an "agent" metaphor where each Claude instance is an "agent" assigned to specific repositories. User feedback and analysis reveal this doesn't align with how developers think about their work. Developers think in terms of "working on parts of the codebase" rather than "managing agents."

Additionally, the current implementation has unnecessary complexity with foreground/background/new-window options, and the agent naming scheme adds cognitive overhead without clear value.

## Decision

We will redesign repo-claude around the concept of "scopes" instead of "agents". A scope represents a logical grouping of repositories that a developer wants to work on together.

### Key Changes:

1. **Terminology**: Replace "agents" with "scopes" throughout
2. **Simplification**: Always fork to new terminal window (remove foreground/background options)
3. **Usability**: Numbered process listing for easy reference (`rc kill 1` instead of `rc stop backend-agent`)
4. **Flexibility**: Start command accepts scope names or individual repo names
5. **Dynamic Scopes**: Future support for changing scope within a running Claude session

## Consequences

### Positive
- **Better Mental Model**: "Scopes" align with how developers think about their work
- **Simpler UX**: Consistent behavior (always new window), easier process management
- **More Flexible**: Scopes can overlap, be dynamic, and compose naturally
- **Cleaner Commands**: `rc start backend` vs `rc start backend-agent`

### Negative
- **Breaking Change**: Existing configurations will need migration
- **Relearning**: Users familiar with agent concept need to adjust
- **Implementation Effort**: Significant refactoring required

## Implementation Plan

### Phase 1: Core Redesign (1-2 weeks)
1. Simplify start command - always fork to new window
2. Implement numbered ps/kill commands
3. Redesign configuration schema from agents to scopes
4. Update documentation and examples

### Phase 2: Dynamic Scopes (2-3 weeks)
1. Implement file-based IPC for scope changes
2. Add `/rc:` command support in Claude sessions
3. Update ps command to show current scope (not just initial)
4. Add scope transition logging

## Example Configuration

### Before (Agent-based):
```yaml
agents:
  backend-agent:
    model: "claude-3-sonnet"
    specialization: "API development"
    auto_start: true
  frontend-agent:
    model: "claude-3-sonnet"
    specialization: "UI development"
    dependencies: ["backend-agent"]
```

### After (Scope-based):
```yaml
scopes:
  backend:
    repos: [auth-service, order-service, payment-service]
    description: "Backend services development"
    model: "claude-3-sonnet"
    auto_start: true
    
  fullstack:
    repos: [backend/*, frontend, shared-libs]
    description: "Full-stack development"
    model: "claude-3-sonnet"
    
  order-flow:
    repos: [order-service, payment-service, shipping-service]
    description: "Order processing pipeline"
```

## Migration Strategy

1. Support both configurations temporarily with deprecation warning
2. Provide automated migration tool: `rc migrate-config`
3. Update all documentation and examples
4. Remove agent support in version 2.0