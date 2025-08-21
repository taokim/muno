# Migration Guide: From Agent-Based to Scope-Based Architecture

This guide helps you migrate from the legacy agent-based configuration to the new scope-based architecture in repo-claude.

## Overview of Changes

The main change is a shift from **agent-based** to **scope-based** organization:
- **Agents** were tied to single repositories
- **Scopes** can span multiple repositories, providing better context for AI

## Key Benefits

1. **Better Context**: Claude sees all related repositories at once
2. **Flexible Grouping**: Use wildcards to group repositories dynamically
3. **Improved Workflow**: Sessions open in tabs by default
4. **Cleaner Configuration**: No need to map agents to individual repos

## Configuration Changes

### Old Agent-Based Format

```yaml
workspace:
  manifest:
    projects:
      - name: backend
        agent: backend-agent  # Each repo needs an agent
      - name: frontend
        agent: frontend-agent

agents:
  backend-agent:
    model: claude-sonnet-4
    specialization: Backend development
    auto_start: true
  frontend-agent:
    model: claude-sonnet-4
    specialization: Frontend development
    dependencies: [backend-agent]
```

### New Scope-Based Format

```yaml
workspace:
  manifest:
    projects:
      - name: backend
        # No agent field needed!
      - name: frontend
      - name: user-service
      - name: order-service

scopes:
  backend:
    repos: ["backend", "*-service"]  # Can include multiple repos
    description: "Backend API and microservices"
    model: claude-sonnet-4
    auto_start: true
  
  frontend:
    repos: ["frontend"]
    description: "Frontend web application"
    model: claude-sonnet-4
    dependencies: [backend]
```

## Step-by-Step Migration

### 1. Backup Your Configuration

```bash
cp repo-claude.yaml repo-claude.yaml.backup
```

### 2. Update Configuration Structure

Replace the `agents:` section with `scopes:`:

```yaml
# Remove this:
agents:
  my-agent:
    model: claude-sonnet-4
    specialization: "My specialization"

# Add this:
scopes:
  my-scope:
    repos: ["repo1", "repo2"]  # List repos this scope covers
    description: "What this scope does"
    model: claude-sonnet-4
```

### 3. Remove Agent References from Projects

```yaml
projects:
  - name: my-repo
    # Remove this line:
    agent: my-agent
```

### 4. Group Related Repositories

Think about how to group your repositories by:
- **Feature**: All repos needed for a feature
- **Domain**: Backend, frontend, mobile, etc.
- **Team**: Repos owned by specific teams

Use wildcards for dynamic grouping:
- `"*-service"` matches all service repos
- `"api-*"` matches all API repos

## Command Changes

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| `rc stop` | `rc kill` | Stop all scopes |
| `rc stop agent-name` | `rc kill scope-name` | Stop specific scope |
| `rc ps` | `rc ps` | Now shows numbers by default |
| `rc start agent` | `rc start scope` | Opens in tab by default |
| N/A | `rc kill 1 2` | Kill by numbers from ps |
| N/A | `rc start --new-window` | Force new window |

## Common Migration Patterns

### Pattern 1: Service-Oriented Architecture

**Before:**
```yaml
agents:
  auth-agent:
    model: claude-sonnet-4
    specialization: "Authentication service"
  order-agent:
    model: claude-sonnet-4
    specialization: "Order service"
  payment-agent:
    model: claude-sonnet-4
    specialization: "Payment service"
```

**After:**
```yaml
scopes:
  backend:
    repos: ["*-service", "api-gateway"]
    description: "All backend microservices"
    model: claude-sonnet-4
```

### Pattern 2: Frontend/Backend Split

**Before:**
```yaml
agents:
  frontend-agent:
    model: claude-sonnet-4
    specialization: "React development"
  backend-agent:
    model: claude-sonnet-4
    specialization: "API development"
```

**After:**
```yaml
scopes:
  frontend:
    repos: ["web-app", "mobile-app", "shared-components"]
    description: "All frontend applications"
    model: claude-sonnet-4
  
  backend:
    repos: ["api", "auth-service", "data-service"]
    description: "Backend API and services"
    model: claude-sonnet-4
```

### Pattern 3: Feature-Based Scopes

**New capability** - create scopes for cross-cutting features:

```yaml
scopes:
  checkout-flow:
    repos: ["cart-service", "payment-service", "order-service", "web-checkout", "mobile-checkout"]
    description: "Complete checkout flow implementation"
    model: claude-opus-4  # Use stronger model for complex work
  
  user-management:
    repos: ["auth-service", "user-service", "web-app", "mobile-app"]
    description: "User authentication and profile management"
    model: claude-sonnet-4
```

## Backwards Compatibility

The tool maintains backwards compatibility:
- Old agent-based configs still work
- Output shows "[legacy mode]" when using agents
- You can migrate gradually

## Environment Variables

Claude sessions now receive these environment variables:
- `RC_SCOPE_ID`: Unique scope instance ID
- `RC_SCOPE_NAME`: The scope name
- `RC_SCOPE_REPOS`: Comma-separated repo list
- `RC_WORKSPACE_ROOT`: Path where repositories are cloned (workspace directory)
- `RC_PROJECT_ROOT`: Path where repo-claude.yaml is located (project root)

## Tips for Effective Scopes

1. **Think in Features**: Group repos by feature or workflow
2. **Use Wildcards**: `"*-service"` is cleaner than listing each service
3. **Start Small**: Begin with obvious groupings (frontend/backend)
4. **Iterate**: Adjust scopes based on your workflow
5. **Model Selection**: Use stronger models for complex cross-repo work

## Troubleshooting

### Issue: "agent X not found in configuration"
**Solution**: You're using an old agent name. Check your new scope names with `rc ps`.

### Issue: Sessions open in tabs but I want windows
**Solution**: Use `rc start --new-window` to force new window behavior.

### Issue: Can't find which scope has my repository
**Solution**: Use `rc start <repo-name>` - it will find the right scope automatically.

## Need Help?

- Run `rc --help` for command help
- Check the [README](../README.md) for examples
- File issues at [GitHub](https://github.com/taokim/repo-claude)