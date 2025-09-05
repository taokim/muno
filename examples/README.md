# MUNO Configuration Examples

This directory contains example MUNO configurations demonstrating different organizational strategies.

## Available Examples

### 1. Team-Based Organization (`team-based.yaml`)

Organizes repositories by team ownership. Best for:
- Clear team boundaries
- Autonomous team operations
- Conway's Law alignment

```bash
# Use this configuration
muno init my-platform
cp examples/team-based.yaml my-platform/muno.yaml
cd my-platform
muno clone --recursive
```

### 2. Service-Type Organization (`service-type.yaml`)

Groups repositories by architectural layers. Best for:
- Full-stack applications
- Clear separation of concerns
- Technology-focused teams

```bash
# Use this configuration
muno init my-platform
cp examples/service-type.yaml my-platform/muno.yaml
cd my-platform
muno clone --recursive
```

### 3. Domain-Driven Organization (`domain-driven.yaml`)

Follows domain-driven design principles. Best for:
- Microservices architectures
- Business domain alignment
- Large enterprise systems

```bash
# Use this configuration
muno init my-platform
cp examples/domain-driven.yaml my-platform/muno.yaml
cd my-platform
muno clone --recursive
```

## Creating Your Own Structure

These examples can be customized:

1. **Copy an example** that's closest to your needs
2. **Modify the structure** to match your organization
3. **Update repository URLs** to point to your repos
4. **Mark appropriate repos as lazy** for on-demand loading

## Key Configuration Options

### Lazy Loading

Mark repositories as `lazy: true` when they are:
- Documentation or wiki repositories
- Rarely modified shared libraries
- Deprecated or archived services
- Large repositories with binary assets

### Config References

Use external configurations for team autonomy:

```yaml
nodes:
  - name: team-frontend
    config: ../frontend-team/muno.yaml  # Team manages their own structure
```

### Mixed Approaches

Combine strategies as needed:

```yaml
nodes:
  - name: core-services  # Domain grouping
    nodes:
      - name: team-backend  # Team ownership
        nodes:
          - name: payment-api  # Service
```

## Navigation Examples

After setting up your workspace:

```bash
# Navigate to specific service
muno use team-backend/payment-service

# Start Claude at current location
muno claude

# Or start at specific location
muno claude team-frontend/web-app

# Pull all repositories in a subtree
muno pull team-backend --recursive

# Check status across entire workspace
muno status --recursive
```

## For AI Agents

See [AI Agent Guide](../docs/AI_AGENT_GUIDE.md) for comprehensive instructions on:
- Analyzing existing repository structures
- Migrating from other tools (Google Repo, etc.)
- Organizing large collections of repositories
- Generating migration scripts

## Tips

1. **Start Simple**: Begin with a basic structure and evolve
2. **Use Lazy Loading**: Don't clone everything immediately
3. **Document Structure**: Keep a README explaining your organization
4. **Team Boundaries**: Respect ownership and access patterns
5. **Regular Review**: Reorganize as your system grows