# Repo-Claude v3 Commands

## Core Commands

### Workspace Management
- `rc init [project-name]` - Initialize a new v3 workspace
- `rc list` - List all available scopes
- `rc status <scope>` - Show detailed status of a scope

### Scope Operations
- `rc start <scope>` - Start a Claude session for a scope
- `rc use <scope>` - Set the active scope
- `rc use --clear` - Clear the active scope
- `rc current` - Show the current active scope

### Scope Management
- `rc scope create <name>` - Create a new scope
- `rc scope delete <name>` - Delete a scope
- `rc scope archive <name>` - Archive a scope

### Git Operations
All git commands are scope-aware and operate within scope context:

- `rc pull [scope]` - Pull repositories in a scope
- `rc commit [scope] -m "message"` - Commit changes
- `rc push [scope]` - Push changes
- `rc branch [scope] <branch-name>` - Switch branches

### Documentation
- `rc docs create <scope|global> <filename>` - Create documentation
- `rc docs list [scope]` - List documentation files
- `rc docs sync` - Sync documentation to Git

### Utility
- `rc version` - Show version information
- `rc help` - Show help for any command

## Commands Removed from v2

### Removed (Not Needed in v3)
- Agent management commands (agents, ps) - v3 uses scopes instead
- Migration commands - v3 is a clean slate, no migration
- PR command - Not implemented, commented out for future

### Simplified
- Tree command - Planned for future implementation with v3 workspace hierarchy
- Sync command - Replaced by scope-specific pull operations

## Key Differences in v3

1. **Scope-Centric**: All operations revolve around scopes, not agents
2. **Smart Loading**: Commands understand meta-repos vs services
3. **Recursive Support**: Commands work with nested workspaces
4. **Simplified Config**: No complex type system, everything is a repository

## Command Examples

```bash
# Initialize a new v3 workspace
rc init my-platform

# List available scopes
rc list

# Start working on backend scope
rc start backend

# Pull latest changes for active scope
rc pull

# Commit and push changes
rc commit -m "feat: add payment processing"
rc push

# Switch to another scope
rc use frontend

# Create documentation
rc docs create global architecture.md
```

## Future Commands (Planned)

- `rc tree` - Display workspace hierarchy with recursive repos
- `rc pr create` - Create pull requests for scope repositories
- `rc diff` - Show differences across scope repositories
- `rc sync` - Advanced synchronization with conflict resolution