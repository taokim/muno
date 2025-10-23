# Advanced Configuration Guide

## Overview

This guide covers advanced configuration patterns, overlay mechanisms, and best practices for complex MUNO workspace setups.

## Configuration Overlay System

### Understanding Overlay Precedence

MUNO uses a layered configuration system where configurations can overlay each other. The precedence order (highest to lowest):

1. **Runtime Overrides** - Command-line flags and environment variables
2. **Local Overrides** - `.muno-local.yaml` (gitignored)
3. **Node Configuration** - Node-specific `muno.yaml`
4. **Parent Configuration** - Inherited from parent nodes
5. **Root Configuration** - Workspace root `muno.yaml`
6. **System Defaults** - Built-in MUNO defaults

### Layer Interaction Example

```yaml
# System Default (built-in)
repos_dir: "repos"
lazy: false

# Root Config (muno.yaml)
workspace:
  repos_dir: "nodes"  # Override default
nodes:
  - name: backend
    lazy: true  # Override default

# Node Config (backend/muno.yaml)
workspace:
  repos_dir: "services"  # Override parent
  
# Local Override (.muno-local.yaml)
nodes:
  - name: backend
    lazy: false  # Override for local development
    
# Clone when needed
cd .nodes/backend && muno clone  # Clone lazy repo
```

## Override Mechanisms

### 1. Full Replacement Override

Completely replace a configuration section:

```yaml
# Parent config
nodes:
  - name: service-a
    url: https://github.com/prod/service-a.git
  - name: service-b
    url: https://github.com/prod/service-b.git

# Child config (full replacement)
nodes: !replace  # Explicit replace directive
  - name: service-c
    url: https://github.com/dev/service-c.git
```

### 2. Merge Override (Default)

Merge configurations with same-name replacement:

```yaml
# Parent config
nodes:
  - name: shared
    url: https://github.com/org/shared.git
    lazy: false
  - name: auth
    url: https://github.com/org/auth.git

# Child config (merge)
nodes:  # Default merge behavior
  - name: shared  # Replaces parent's 'shared'
    url: https://github.com/org/shared-fork.git
    lazy: true
  - name: new-service  # Adds new node
    url: https://github.com/org/new.git
# Result: shared (overridden), auth (inherited), new-service (added)
```

### 3. Additive Override

Add nodes without replacing existing ones:

```yaml
# Parent config
nodes:
  - name: core
    url: https://github.com/org/core.git

# Child config (additive)
nodes: !append  # Explicit append directive
  - name: extension
    url: https://github.com/org/extension.git
# Result: both core and extension
```

### 4. Selective Field Override

Override specific fields while inheriting others:

```yaml
# Parent config
nodes:
  - name: database
    url: https://github.com/org/database.git
    lazy: false
    description: Production database
    options:
      branch: main
      depth: 1

# Child config
nodes:
  - name: database
    lazy: true  # Override only lazy flag
    options:
      branch: develop  # Override only branch
      # depth inherited as 1
    # url and description inherited
```

## Complex Configuration Patterns

### Pattern 1: Environment-Specific Overlays

```yaml
# base.yaml - Shared base configuration
workspace:
  name: platform
  repos_dir: repositories

nodes:
  - name: core
    url: https://github.com/org/core.git
    lazy: false

# dev-overlay.yaml - Development overlay
include: base.yaml  # Include base config

workspace:
  name: platform-dev  # Override name

nodes:
  - name: core
    lazy: true  # Override for faster dev startup
  - name: dev-tools
    url: https://github.com/org/dev-tools.git

# prod-overlay.yaml - Production overlay  
include: base.yaml

nodes:
  - name: monitoring
    url: https://github.com/org/monitoring.git
    lazy: false  # Always load in production
```

### Pattern 2: Team-Based Configuration

```yaml
# platform/muno.yaml - Root configuration
workspace:
  name: enterprise-platform

nodes:
  # Shared infrastructure
  - name: infrastructure
    url: https://github.com/company/infra.git
    
  # Team nodes with configs
  - name: team-backend
    config: !env BACKEND_CONFIG_URL  # Environment variable
    fallback: ./teams/backend/muno.yaml
    
  - name: team-frontend
    config: 
      primary: https://configs.company.com/frontend.yaml
      fallback: ./teams/frontend/muno.yaml
      cache: 3600  # Cache for 1 hour
```

### Pattern 3: Conditional Configuration

```yaml
# Conditional loading based on environment
nodes:
  - name: debug-tools
    url: https://github.com/org/debug-tools.git
    condition: !env DEBUG_MODE  # Only load if DEBUG_MODE is set
    
  - name: performance-monitor
    url: https://github.com/org/perf-monitor.git
    condition:
      env: ENVIRONMENT
      value: production  # Only in production
      
  - name: test-fixtures
    url: https://github.com/org/test-fixtures.git
    condition:
      not:
        env: ENVIRONMENT
        value: production  # Not in production
```

### Pattern 4: Template Variables

```yaml
# Configuration with variables
variables:
  org: !env GITHUB_ORG
  default_branch: !env DEFAULT_BRANCH || main
  repo_prefix: https://github.com/${org}

workspace:
  name: ${org}-platform
  
nodes:
  - name: service-a
    url: ${repo_prefix}/service-a.git
    options:
      branch: ${default_branch}
      
  - name: service-b
    url: ${repo_prefix}/service-b.git
    options:
      branch: ${default_branch}
```

## Config Reference Node Patterns

### Pattern 1: Versioned Config References

```yaml
nodes:
  - name: stable-services
    config:
      url: https://configs.company.com/services.yaml
      version: v2.1.0  # Specific version
      
  - name: latest-services
    config:
      url: https://configs.company.com/services.yaml
      version: latest  # Always latest
      
  - name: pinned-services
    config:
      url: https://configs.company.com/services.yaml
      sha: abc123def456  # Pinned to specific commit
```

### Pattern 2: Multi-Source Config

```yaml
nodes:
  - name: aggregated-services
    config:
      sources:
        - url: https://configs.company.com/core.yaml
          priority: 1
        - url: https://configs.company.com/extensions.yaml
          priority: 2
        - file: ./local-overrides.yaml
          priority: 3  # Highest priority
      merge_strategy: deep  # deep, shallow, or replace
```

### Pattern 3: Dynamic Config Resolution

```yaml
nodes:
  - name: regional-services
    config:
      resolver: !script ./scripts/get-regional-config.sh
      args:
        - !env AWS_REGION
        - !env DEPLOYMENT_ENV
      cache: 300  # Cache for 5 minutes
```

## Workspace Inheritance

### Multi-Level Workspace Configuration

```yaml
# Level 1: Global (company-wide)
# /company/muno.yaml
workspace:
  name: company-platform
  repos_dir: repositories
  settings:
    clone_depth: 1
    parallel_clones: 4

# Level 2: Division
# /company/engineering/muno.yaml
inherit: ../muno.yaml
workspace:
  name: engineering-platform
  settings:
    clone_depth: 0  # Full clones for engineering
    # parallel_clones inherited as 4

# Level 3: Team
# /company/engineering/backend/muno.yaml
inherit: ../muno.yaml
workspace:
  name: backend-platform
  settings:
    parallel_clones: 8  # More parallel for backend team
    # clone_depth inherited as 0
```

### Workspace Settings Overlay

```yaml
# Default settings
workspace:
  settings:
    auto_clone: true
    verify_ssl: true
    timeout: 30
    retry_count: 3
    
# Override for CI/CD
workspace:
  settings:
    auto_clone: false  # Manual control in CI
    timeout: 120  # Longer timeout for CI
    # verify_ssl and retry_count inherited
```

## Advanced Lazy Loading

### Conditional Lazy Loading

```yaml
nodes:
  - name: large-dataset
    url: https://github.com/org/large-dataset.git
    lazy:
      default: true  # Default lazy
      conditions:
        - env: CI
          value: true
          lazy: false  # Not lazy in CI
        - env: PRELOAD_ALL
          value: true
          lazy: false  # Force eager load
```

### Progressive Loading

```yaml
nodes:
  - name: service-group
    config: ./services/muno.yaml
    loading:
      strategy: progressive
      stages:
        - immediate: [core, auth]  # Load immediately
        - on_demand: [analytics, reporting]  # Load when accessed
        - background: [backup, maintenance]  # Load in background
```

## Security and Access Control

### Authenticated Config References

```yaml
nodes:
  - name: private-services
    config:
      url: https://private-configs.company.com/services.yaml
      auth:
        type: bearer
        token: !env CONFIG_ACCESS_TOKEN
        
  - name: github-private
    config:
      url: https://api.github.com/repos/org/configs/contents/muno.yaml
      auth:
        type: basic
        username: !env GITHUB_USER
        password: !env GITHUB_TOKEN
```

### Config Validation

```yaml
workspace:
  validation:
    schema: https://schemas.company.com/muno-v2.schema.json
    strict: true  # Fail on validation errors
    
nodes:
  - name: validated-service
    url: https://github.com/org/service.git
    validation:
      pre_clone: ./scripts/validate-repo.sh
      post_clone: ./scripts/verify-setup.sh
```

## Performance Optimization

### Caching Strategies

```yaml
workspace:
  cache:
    config:
      enabled: true
      ttl: 3600  # 1 hour
      location: ~/.muno/cache
    
    remote_configs:
      enabled: true
      ttl: 86400  # 24 hours
      refresh_on_error: false
      
nodes:
  - name: frequently-accessed
    url: https://github.com/org/common.git
    cache:
      metadata: true  # Cache repo metadata
      ttl: 600  # 10 minutes
```

### Parallel Operations

```yaml
workspace:
  performance:
    parallel:
      clone: 8  # Clone 8 repos simultaneously
      pull: 4   # Pull 4 repos simultaneously
      config_fetch: 2  # Fetch 2 configs simultaneously
    
    timeout:
      clone: 300  # 5 minutes
      config_fetch: 30  # 30 seconds
      
    retry:
      attempts: 3
      backoff: exponential
      max_delay: 60
```

## Troubleshooting Advanced Configurations

### Debug Mode

```yaml
workspace:
  debug:
    enabled: !env MUNO_DEBUG
    verbose: true
    log_level: trace
    log_file: ./muno-debug.log
    
    trace:
      config_resolution: true
      overlay_merge: true
      variable_substitution: true
```

### Config Resolution Tracing

```bash
# Show config resolution steps
muno config --trace

# Show effective configuration after all overlays
muno config --effective

# Validate configuration without applying
muno config --validate --dry-run

# Show config inheritance tree
muno config --inheritance-tree
```

### Common Issues and Solutions

1. **Overlay Not Applied**
   ```bash
   # Check overlay precedence
   muno config --show-layers
   
   # Verify include paths
   muno config --verify-includes
   ```

2. **Variable Not Resolved**
   ```bash
   # List all variables and their values
   muno config --show-variables
   
   # Debug variable substitution
   MUNO_DEBUG_VARS=true muno tree
   ```

3. **Config Cache Issues**
   ```bash
   # Clear config cache
   muno cache --clear configs
   
   # Force refresh
   muno pull --force-refresh --configs
   ```

## Best Practices

### 1. Configuration Organization

- Keep base configs simple and focused
- Use overlays for environment-specific settings
- Document overlay precedence clearly
- Version control all configuration files

### 2. Security

- Never commit sensitive data in configs
- Use environment variables for secrets
- Implement config validation
- Audit config access logs

### 3. Performance

- Use appropriate cache TTLs
- Implement progressive loading for large trees
- Optimize parallel operation limits
- Monitor config fetch times

### 4. Maintainability

- Use consistent naming conventions
- Document complex overlay logic
- Implement config validation tests
- Regular config cleanup and optimization

## Migration Strategies

### From Simple to Advanced

1. **Phase 1**: Basic node definitions
2. **Phase 2**: Add lazy loading
3. **Phase 3**: Implement config references
4. **Phase 4**: Add overlays and variables
5. **Phase 5**: Implement advanced patterns

### Rollback Strategy

```yaml
# Maintain compatibility with fallbacks
nodes:
  - name: service
    config:
      primary: https://new-config-system.com/service.yaml
      fallback: ./legacy/service-config.yaml
      compatibility_mode: v1  # Use v1 parsing
```

## Future Roadmap

Planned enhancements for advanced configuration:

1. **GraphQL Config API**: Query configurations dynamically
2. **Config Plugins**: Extend config resolution with plugins
3. **Smart Caching**: ML-based cache invalidation
4. **Config Diffing**: Track config changes over time
5. **A/B Config Testing**: Test config changes gradually
6. **Config Marketplace**: Share and discover config patterns