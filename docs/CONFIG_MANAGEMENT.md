# MUNO Configuration Management

## Overview

MUNO uses a hierarchical configuration system that supports distributed team management, config inheritance, and flexible overlay/override mechanisms. This allows teams to maintain their own configurations while inheriting from parent configurations.

## Table of Contents

1. [Configuration File Structure](#configuration-file-structure)
2. [Node Types](#node-types)
3. [Configuration Inheritance](#configuration-inheritance)
4. [Overlay and Override Mechanisms](#overlay-and-override-mechanisms)
5. [Config Reference Nodes](#config-reference-nodes)
6. [Best Practices](#best-practices)
7. [Examples](#examples)

## Configuration File Structure

### Basic Structure (`muno.yaml`)

```yaml
# Workspace configuration
workspace:
  name: my-platform
  description: Platform-wide repository management
  repos_dir: nodes  # Default directory for repositories

# Node definitions
nodes:
  - name: backend
    url: https://github.com/org/backend.git
    lazy: true
    
  - name: frontend
    config: ./frontend/muno.yaml  # Config reference
    
  - name: shared
    url: https://github.com/org/shared.git
    lazy: false
```

### Field Descriptions

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `workspace.name` | string | Workspace identifier | Yes |
| `workspace.description` | string | Human-readable description | No |
| `workspace.repos_dir` | string | Directory for repositories (default: "nodes") | No |
| `nodes` | array | List of node definitions | No |
| `nodes[].name` | string | Node identifier | Yes |
| `nodes[].url` | string | Git repository URL | No* |
| `nodes[].config` | string | Path/URL to config file | No* |
| `nodes[].lazy` | boolean | Lazy loading flag | No |
| `nodes[].description` | string | Node description | No |

*Note: Each node must have EITHER `url` OR `config`, never both.

## Node Types

### 1. Git Repository Nodes

Standard git repositories that MUNO manages directly:

```yaml
nodes:
  - name: payment-service
    url: https://github.com/org/payment.git
    lazy: true  # Clone on first access
```

### 2. Config Reference Nodes

Nodes that delegate their subtree management to external configurations:

```yaml
nodes:
  - name: team-frontend
    config: ../frontend/muno.yaml  # Local config file
    
  - name: infrastructure
    config: https://config.company.com/infra.yaml  # Remote config
```

### 3. Hybrid Nodes

Git repositories that also contain their own `muno.yaml` for child definitions:

```yaml
# In parent muno.yaml
nodes:
  - name: backend-monorepo
    url: https://github.com/org/backend.git

# In backend-monorepo/muno.yaml (after cloning)
nodes:
  - name: auth-service
    url: https://github.com/org/auth.git
  - name: payment-service
    url: https://github.com/org/payment.git
```

## Configuration Inheritance

### Inheritance Chain

Configurations inherit in a parent-to-child manner:

```
Root Config (muno.yaml)
    ├── Team Config (team-backend/muno.yaml)
    │   ├── Service Config (services/muno.yaml)
    │   └── Shared Config (shared/muno.yaml)
    └── Team Config (team-frontend/muno.yaml)
```

### Property Resolution

Properties are resolved using the following precedence (highest to lowest):

1. **Local Override**: Explicitly set in the current config
2. **Parent Config**: Inherited from parent configuration
3. **Default Values**: System defaults

Example:
```yaml
# Parent config
workspace:
  repos_dir: repositories
  
# Child config (inherits repos_dir unless overridden)
workspace:
  name: child-workspace
  # repos_dir inherited as "repositories"
```

## Overlay and Override Mechanisms

### 1. Complete Override

Replace entire configuration sections:

```yaml
# Parent config
nodes:
  - name: service-a
    url: https://github.com/org/service-a.git
    lazy: true

# Child config (complete override)
nodes:  # Replaces entire nodes array
  - name: service-b
    url: https://github.com/org/service-b.git
```

### 2. Merge Strategy

Merge configurations at the node level:

```yaml
# Parent config
nodes:
  - name: shared
    url: https://github.com/org/shared.git
    lazy: false

# Child config with merge
nodes:
  - name: shared  # Same name, will override parent
    lazy: true    # Override just the lazy flag
  - name: new-service  # Add new node
    url: https://github.com/org/new.git
```

### 3. Additive Configuration

Add new nodes without affecting parent nodes:

```yaml
# Use special syntax (when implemented)
nodes+:  # Additive mode
  - name: additional-service
    url: https://github.com/org/additional.git
```

### 4. Selective Override

Override specific properties while inheriting others:

```yaml
# Parent defines base configuration
workspace:
  repos_dir: nodes
  description: Base workspace
  
# Child overrides selectively
workspace:
  description: Specialized workspace  # Override
  # repos_dir inherited as "nodes"
```

## Config Reference Nodes

### Local Config References

Reference configurations in the local filesystem:

```yaml
nodes:
  - name: subproject
    config: ./subproject/muno.yaml  # Relative path
    
  - name: sibling
    config: ../sibling-project/muno.yaml  # Parent directory
    
  - name: absolute
    config: /opt/configs/project.yaml  # Absolute path
```

### Remote Config References

Reference configurations from URLs:

```yaml
nodes:
  - name: cloud-config
    config: https://config.example.com/cloud.yaml
    
  - name: github-config
    config: https://raw.githubusercontent.com/org/configs/main/platform.yaml
```

### Config Reference Resolution

1. **Load Time**: Config references are resolved when the node is accessed
2. **Caching**: Remote configs are cached locally for performance
3. **Updates**: Use `muno pull` to refresh remote configurations
4. **Security**: HTTPS recommended for remote configs

## Best Practices

### 1. Team Autonomy

Enable teams to manage their own configurations:

```yaml
# Platform root config
nodes:
  - name: team-backend
    config: https://github.com/backend-team/config/muno.yaml
    
  - name: team-frontend
    config: https://github.com/frontend-team/config/muno.yaml
```

### 2. Shared Components

Centralize shared components at appropriate levels:

```yaml
# Root level - platform-wide shared
nodes:
  - name: shared-libs
    url: https://github.com/org/shared-libs.git
    
  # Team level - team-specific shared
  - name: team-backend
    config: ./backend/muno.yaml

# In backend/muno.yaml
nodes:
  - name: backend-commons
    url: https://github.com/org/backend-commons.git
```

### 3. Environment-Specific Configs

Use different configs for different environments:

```yaml
# Production config
nodes:
  - name: services
    config: https://config.company.com/prod/services.yaml

# Development config  
nodes:
  - name: services
    config: https://config.company.com/dev/services.yaml
```

### 4. Version Control

Keep configurations in version control:

```bash
# Track config changes
git add muno.yaml
git commit -m "Update team structure"
git push
```

## Examples

### Example 1: Multi-Team Platform

```yaml
# Root muno.yaml
workspace:
  name: enterprise-platform
  description: Company-wide platform
  repos_dir: repositories

nodes:
  # Shared infrastructure
  - name: infrastructure
    url: https://github.com/company/infrastructure.git
    lazy: false
    
  # Team configurations
  - name: team-payments
    config: https://github.com/payments-team/config/raw/main/muno.yaml
    
  - name: team-accounts
    config: https://github.com/accounts-team/config/raw/main/muno.yaml
    
  - name: team-mobile
    config: ./mobile/muno.yaml  # Local team config
```

### Example 2: Microservices with Shared Libraries

```yaml
# services/muno.yaml
workspace:
  name: microservices
  repos_dir: services

nodes:
  # Shared libraries (eager load)
  - name: common-utils
    url: https://github.com/org/common-utils.git
    lazy: false
    
  - name: proto-definitions
    url: https://github.com/org/proto-definitions.git
    lazy: false
    
  # Service repositories (lazy load)
  - name: auth-service
    url: https://github.com/org/auth-service.git
    lazy: true
    
  - name: user-service
    url: https://github.com/org/user-service.git
    lazy: true
    
  - name: notification-service
    url: https://github.com/org/notification-service.git
    lazy: true
```

### Example 3: Monorepo with Sub-projects

```yaml
# backend-monorepo/muno.yaml (within the cloned repo)
workspace:
  name: backend-monorepo
  description: Backend services monorepo

nodes:
  # Core services
  - name: api-gateway
    url: https://github.com/org/api-gateway.git
    
  - name: auth-service
    url: https://github.com/org/auth-service.git
    
  # Data layer
  - name: data-pipeline
    config: ./data/muno.yaml  # Delegate to data team
    
  # Testing and tools
  - name: integration-tests
    url: https://github.com/org/integration-tests.git
    lazy: true
```

### Example 4: Override Configuration

```yaml
# Base configuration (base.yaml)
workspace:
  name: base-platform
  repos_dir: repos
  
nodes:
  - name: core-service
    url: https://github.com/org/core.git
    lazy: false
    
  - name: optional-service
    url: https://github.com/org/optional.git
    lazy: true

# Override for development (dev.yaml)
workspace:
  name: dev-platform
  # Inherits repos_dir: repos
  
nodes:
  - name: core-service
    url: https://github.com/org/core.git
    lazy: true  # Override to lazy for dev
    
  - name: optional-service
    url: https://github.com/fork/optional.git  # Use fork
    lazy: false  # Load immediately in dev
    
  - name: dev-tools  # Add dev-only node
    url: https://github.com/org/dev-tools.git
```

## Troubleshooting

### Common Issues

1. **Config Not Found**
   ```bash
   Error: Config file not found: ./team/muno.yaml
   ```
   Solution: Ensure the config file path is correct and accessible.

2. **Circular References**
   ```bash
   Error: Circular config reference detected
   ```
   Solution: Check that configs don't reference each other in a loop.

3. **Remote Config Unreachable**
   ```bash
   Error: Failed to fetch remote config
   ```
   Solution: Verify network connectivity and URL correctness.

4. **Node Type Conflict**
   ```bash
   Error: Node cannot have both 'url' and 'config'
   ```
   Solution: Use either `url` for git repos or `config` for references, not both.

### Debugging Commands

```bash
# Validate configuration
muno validate

# Show resolved configuration
muno config --show

# Refresh remote configs
muno pull --configs

# Check config inheritance
muno tree --show-config
```

## Migration Guide

### From Flat to Hierarchical

1. **Identify team boundaries**
2. **Create team-specific configs**
3. **Move repositories to appropriate configs**
4. **Update root config with references**
5. **Test navigation and access**

### From Monolithic to Distributed

1. **Extract team configurations**
2. **Push configs to team repositories**
3. **Update root with remote references**
4. **Implement access controls**
5. **Document team ownership**

## Security Considerations

1. **Access Control**: Use HTTPS and authentication for remote configs
2. **Validation**: Validate all config files before applying
3. **Audit Trail**: Track config changes in version control
4. **Secrets**: Never store credentials in config files
5. **Permissions**: Limit write access to production configs

## Future Enhancements

Planned features for configuration management:

1. **Schema Validation**: YAML schema for config validation
2. **Config Templates**: Reusable configuration templates
3. **Variable Substitution**: Environment variables in configs
4. **Config Versioning**: Support for config version requirements
5. **Merge Strategies**: Configurable merge behaviors
6. **Config Hooks**: Pre/post config load hooks
7. **Encrypted Configs**: Support for encrypted configuration sections