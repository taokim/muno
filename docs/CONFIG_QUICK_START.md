# Configuration Quick Start Guide

A practical guide to get started with MUNO configuration management.

## Basic Configuration

### Minimal Configuration

```yaml
# muno.yaml - Simplest possible configuration
workspace:
  name: my-project

nodes:
  - name: backend
    url: https://github.com/myorg/backend.git
```

### Standard Team Configuration

```yaml
# muno.yaml - Typical team setup
workspace:
  name: team-platform
  description: Our team's microservices
  repos_dir: services  # Where to clone repos

nodes:
  # Shared libraries (load immediately)
  - name: common
    url: https://github.com/myorg/common.git
    lazy: false
    
  # Service repos (load on demand)
  - name: api-service
    url: https://github.com/myorg/api.git
    lazy: true
    
  - name: worker-service
    url: https://github.com/myorg/worker.git
    lazy: true
```

## Configuration Scenarios

### Scenario 1: Single Team, Multiple Services

```yaml
workspace:
  name: backend-team

nodes:
  - name: auth-service
    url: https://github.com/company/auth.git
    
  - name: user-service
    url: https://github.com/company/users.git
    
  - name: payment-service
    url: https://github.com/company/payments.git
```

**Usage:**
```bash
muno init backend
muno tree                    # View structure
cd .nodes/auth-service       # Navigate to service
muno pull --recursive        # Update all services
```

### Scenario 2: Multiple Teams with Delegation

**Root Configuration:**
```yaml
# company/muno.yaml
workspace:
  name: company-platform

nodes:
  - name: team-backend
    config: ./backend/muno.yaml
    
  - name: team-frontend
    config: ./frontend/muno.yaml
    
  - name: shared
    url: https://github.com/company/shared.git
```

**Team Configuration:**
```yaml
# company/backend/muno.yaml
workspace:
  name: backend-services

nodes:
  - name: api
    url: https://github.com/company/api.git
    
  - name: workers
    url: https://github.com/company/workers.git
```

**Usage:**
```bash
muno init company
cd .nodes/team-backend/.nodes/api    # Navigate to team's service
muno tree                            # See full hierarchy
```

### Scenario 3: Development vs Production

**Development Configuration:**
```yaml
# dev/muno.yaml
workspace:
  name: dev-environment

nodes:
  - name: app
    url: https://github.com/company/app.git
    lazy: true  # Load on demand in dev
    
  - name: test-data
    url: https://github.com/company/test-data.git
    lazy: false  # Always need test data
```

**Production Configuration:**
```yaml
# prod/muno.yaml
workspace:
  name: prod-environment

nodes:
  - name: app
    url: https://github.com/company/app.git
    lazy: false  # Always load in production
    
  # No test-data in production
```

## Common Patterns

### Pattern 1: Lazy Loading Strategy

```yaml
nodes:
  # Core dependencies - load immediately
  - name: core-lib
    url: https://github.com/org/core.git
    lazy: false
    
  # Optional tools - load when needed
  - name: debug-tools
    url: https://github.com/org/debug.git
    lazy: true
    
  # Large repos - definitely lazy
  - name: dataset
    url: https://github.com/org/large-dataset.git
    lazy: true
```

### Pattern 2: Config References

```yaml
nodes:
  # Local config reference
  - name: subproject
    config: ./subproject/muno.yaml
    
  # Remote config reference
  - name: external-team
    config: https://configs.company.com/external.yaml
```

### Pattern 3: Mixed Node Types

```yaml
nodes:
  # Git repository
  - name: service
    url: https://github.com/org/service.git
    
  # Config reference
  - name: libraries
    config: ./libs/muno.yaml
    
  # Another git repo
  - name: docs
    url: https://github.com/org/docs.git
```

## Quick Command Reference

### Initialize Workspace
```bash
# Create new workspace
muno init my-workspace

# Start with existing config
cd my-project
muno init .
```

### Navigate Tree
```bash
# Show tree structure
muno tree

# Navigate to node
cd .nodes/backend/.nodes/service

# Show current location
pwd

# Return to workspace root
cd ../../..
```

### Manage Repositories
```bash
# Add new repository
muno add https://github.com/org/new-repo.git

# Remove repository
muno remove old-repo

# Clone lazy repositories
muno clone --all
```

### Update Repositories
```bash
# Pull current repository
muno pull

# Pull all repositories recursively
muno pull --recursive

# Pull from specific directory
cd .nodes/team-backend
muno pull --recursive
```

## Configuration Tips

### 1. Start Simple
Begin with basic git repositories, add advanced features as needed:

```yaml
# Start with this
nodes:
  - name: my-service
    url: https://github.com/org/service.git

# Evolve to this
nodes:
  - name: my-service
    url: https://github.com/org/service.git
    lazy: true
    description: Main service application
```

### 2. Use Lazy Loading Wisely
- Set `lazy: false` for frequently used repos
- Set `lazy: true` for large or rarely used repos
- Default is `lazy: false` if not specified

### 3. Organize Hierarchically
Structure your tree to match your mental model:

```yaml
nodes:
  - name: frontend
    config: ./frontend/muno.yaml  # Frontend team
    
  - name: backend
    config: ./backend/muno.yaml   # Backend team
    
  - name: infrastructure
    config: ./infra/muno.yaml     # DevOps team
```

### 4. Document Your Configuration

```yaml
workspace:
  name: platform
  description: |
    Main platform configuration.
    Frontend: Team A
    Backend: Team B
    Contact: platform@company.com

nodes:
  - name: critical-service
    url: https://github.com/org/critical.git
    description: "IMPORTANT: This service requires VPN access"
```

## Troubleshooting

### Issue: "Node cannot have both url and config"

**Wrong:**
```yaml
nodes:
  - name: broken
    url: https://github.com/org/repo.git
    config: ./config.yaml  # ERROR!
```

**Correct:**
```yaml
nodes:
  - name: repo-node
    url: https://github.com/org/repo.git
    
  - name: config-node
    config: ./config.yaml
```

### Issue: "Config file not found"

Check the path is relative to the parent config file:
```yaml
# If this file is at /workspace/muno.yaml
nodes:
  - name: sub
    config: ./sub/muno.yaml  # Looks for /workspace/sub/muno.yaml
```

### Issue: Lazy repositories not cloning

Lazy repositories only clone when explicitly requested:
```bash
cd .nodes/lazy-repo  # Navigate to lazy repo directory
muno clone           # This triggers the clone
# OR from workspace root:
muno clone --recursive    # Clone all lazy repos recursively
```

## Next Steps

1. **Basic Setup**: Start with [CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md)
2. **Advanced Features**: Explore [ADVANCED_CONFIG.md](./ADVANCED_CONFIG.md)
3. **Best Practices**: Review examples in main documentation
4. **Get Help**: Run `muno help` or check [README.md](../README.md)

## Example Configurations

Find more examples in the `examples/` directory:
- `examples/simple-team/` - Basic team setup
- `examples/multi-team/` - Multiple team configuration
- `examples/monorepo/` - Monorepo with sub-projects
- `examples/microservices/` - Microservices architecture