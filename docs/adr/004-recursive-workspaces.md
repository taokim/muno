# ADR-003: Recursive Workspaces Architecture

## Status
Proposed

## Context
As repo-claude adoption grows, organizations need to manage 100-500+ repositories. The current flat configuration becomes unwieldy at this scale. Teams need:
- Autonomous workspace management
- Natural organizational boundaries
- Distributed documentation
- Scalable configuration management

## Decision
Implement a recursive, tree-based workspace architecture where:
1. Each node can be a complete repo-claude workspace
2. Workspaces can reference other workspaces as sub-workspaces
3. Documentation lives with each workspace in its git repository
4. Child workspaces are autonomous (don't need to know about parents)

## Consequences

### Positive
- **Scalability**: Naturally handles 500+ repositories through hierarchy
- **Autonomy**: Teams manage their own workspaces independently
- **Versioning**: Each workspace configuration is versioned in git
- **Documentation**: Docs stay close to code with natural scoping
- **Flexibility**: Supports various organizational structures
- **No Central Registry**: Eliminates single point of failure

### Negative
- **Complexity**: More complex than flat structure
- **Learning Curve**: New concepts (path resolution, tree traversal)
- **Performance**: Potential overhead from tree traversal
- **Caching Needs**: Requires sophisticated caching for performance

### Mitigation
- Maintain full backward compatibility with v2 configs
- Implement lazy loading and aggressive caching
- Provide clear documentation and examples
- Use familiar filesystem-like path syntax

## Implementation
See [RECURSIVE_WORKSPACE_PLAN.md](../RECURSIVE_WORKSPACE_PLAN.md) for detailed implementation plan.

## References
- [ADR-001: Scope Isolation](001-scope-isolation.md)
- [ADR-002: Git Integration](002-direct-git-integration.md)