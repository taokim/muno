# Migration Guide: v1 to v2

## Overview

Repo-Claude v2 introduces **workspace isolation**, where each scope operates in its own directory under `workspaces/`. This guide helps you migrate from the v1 shared workspace model to the v2 isolated architecture.

## Key Changes

### v1 (Shared Workspace)
- Single `workspace/` directory for all repositories
- All scopes share the same Git working directories
- Potential conflicts when working on multiple features
- Agent-based configuration

### v2 (Isolated Workspaces)
- Separate `workspaces/<scope-name>/` for each scope
- Complete isolation between different development contexts
- Parallel development without conflicts
- Scope-based configuration with e-commerce examples
- No TTL implementation (manual lifecycle management)

## Migration Steps

### Step 1: Backup Existing Workspace

```bash
# Create backup of v1 workspace
cp -r my-project my-project-v1-backup

# Or create a tar archive
tar -czf my-project-v1-backup.tar.gz my-project/
```

### Step 2: Update Configuration File

#### Update Version and Workspace Settings

```yaml
# Add at the top of repo-claude.yaml
version: 2

workspace:
  name: my-project
  isolation_mode: true      # Enable isolation (default in v2)
  base_path: workspaces    # New base directory for scopes
```

#### Convert Agents to Scopes

**v1 Agent Configuration:**
```yaml
agents:
  backend-agent:
    model: claude-3-sonnet
    specialization: "Backend API development"
    auto_start: true
    
projects:
  - name: api-service
    path: services/api
    agent: backend-agent
```

**v2 Scope Configuration:**
```yaml
scopes:
  backend:
    type: persistent
    repos: ["api-service", "auth-service", "shared-libs"]
    description: "Backend API development"
    model: claude-3-5-sonnet-20241022
    auto_start: false
```

### Step 3: Create Initial Scopes

```bash
# Navigate to project directory
cd my-project

# Create a main development scope with all repositories
rc scope create main --type persistent --repos "*"

# Create domain-specific scopes
rc scope create backend --type persistent --repos "*-service,shared-libs"
rc scope create frontend --type persistent --repos "*-ui,*-web,shared-ui"
```

### Step 4: Clone Repositories into New Scopes

```bash
# The repositories will be cloned automatically when creating scopes
# Or manually trigger cloning:
rc pull --scope main --clone-missing
rc pull --scope backend --clone-missing
rc pull --scope frontend --clone-missing
```

### Step 5: Migrate Work in Progress

If you have uncommitted changes in v1:

```bash
# In v1 workspace, stash or commit changes
cd workspace/api-service
git stash save "WIP: migration to v2"

# In v2 scope, apply changes
cd workspaces/main/api-service
git stash pop
```

Or use patches:
```bash
# Create patch from v1
cd workspace/api-service
git diff > ~/migration.patch

# Apply to v2 scope
cd workspaces/main/api-service
git apply ~/migration.patch
```

### Step 6: Update Scripts and Documentation

Update any scripts that reference the old workspace structure:

**v1 Path:**
```bash
cd my-project/workspace/api-service
```

**v2 Path:**
```bash
cd my-project/workspaces/main/api-service
# or
cd my-project/workspaces/backend/api-service
```

### Step 7: Clean Up Old Structure (Optional)

Once migration is verified:

```bash
# Remove old workspace directory
rm -rf workspace/

# Keep only v2 structure
ls -la
# Should show:
# - repo-claude.yaml (updated)
# - workspaces/
# - docs/
# - shared-memory.md
```

## Configuration Examples

### Complete v2 E-Commerce Configuration

```yaml
version: 2

workspace:
  name: ecommerce-platform
  isolation_mode: true
  base_path: workspaces

repositories:
  # WMS components
  wms-core:
    url: https://github.com/yourorg/wms-core.git
    default_branch: main
    groups: [wms, backend]
    
  wms-inventory:
    url: https://github.com/yourorg/wms-inventory.git
    default_branch: main
    groups: [wms, backend]
    
  # OMS components
  oms-core:
    url: https://github.com/yourorg/oms-core.git
    default_branch: main
    groups: [oms, backend]
    
  oms-payment:
    url: https://github.com/yourorg/oms-payment.git
    default_branch: main
    groups: [oms, backend]
    
  # Shared
  shared-libs:
    url: https://github.com/yourorg/shared-libs.git
    default_branch: main
    groups: [shared]

scopes:
  wms:
    type: persistent
    repos: ["wms-*", "shared-libs"]
    description: "Warehouse Management System"
    model: claude-3-5-sonnet-20241022
    
  oms:
    type: persistent
    repos: ["oms-*", "shared-libs"]
    description: "Order Management System"
    model: claude-3-5-sonnet-20241022
    
  hotfix:
    type: ephemeral
    repos: []  # Select at creation
    description: "Emergency fixes"
    model: claude-3-5-sonnet-20241022

documentation:
  path: docs
  sync_to_git: true
```

## Command Changes

### Starting Sessions

**v1 Commands:**
```bash
rc start              # Start all auto-start agents
rc start backend-agent # Start specific agent
```

**v2 Commands:**
```bash
rc start              # Interactive scope selection
rc start wms         # Start specific scope
rc scope create hotfix-123 --repos "oms-payment"
rc start hotfix-123
```

### Git Operations

**v1 (Global):**
```bash
rc pull --clone-missing
rc sync
```

**v2 (Per Scope):**
```bash
rc pull --scope wms
rc commit --scope wms -m "Update inventory"
rc push --scope wms
```

## Common Migration Scenarios

### Scenario 1: Simple Project

If your v1 project has a simple structure:

```bash
# Quick migration
echo "version: 2" >> repo-claude.yaml
echo "workspace:" >> repo-claude.yaml
echo "  isolation_mode: true" >> repo-claude.yaml
echo "  base_path: workspaces" >> repo-claude.yaml

# Create single scope with everything
rc scope create main --type persistent --repos "*"
rc start main
```

### Scenario 2: Multi-Team Project

For projects with multiple teams:

```bash
# Create team-specific scopes
rc scope create team-wms --type persistent \
  --repos "wms-*,shared-libs" \
  --description "WMS team workspace"

rc scope create team-oms --type persistent \
  --repos "oms-*,shared-libs" \
  --description "OMS team workspace"

rc scope create team-frontend --type persistent \
  --repos "*-ui,web-*,mobile-*" \
  --description "Frontend team workspace"
```

### Scenario 3: Active Development

If you have active feature branches:

```bash
# Create scope for each active feature
rc scope create feature-payment --type persistent \
  --repos "oms-payment,oms-core,shared-libs"

# Checkout feature branch in scope
cd workspaces/feature-payment/oms-payment
git checkout feature/new-payment-gateway

# Continue development
rc start feature-payment
```

## Troubleshooting

### Issue: Configuration Not Recognized

**Symptom:** v2 features not working

**Solution:** Ensure `version: 2` is at the top of `repo-claude.yaml`

### Issue: Repositories Not Found

**Symptom:** Scopes created but repos missing

**Solution:** 
```bash
rc pull --scope <scope-name> --clone-missing
```

### Issue: Path Conflicts

**Symptom:** Scripts failing with path errors

**Solution:** Update paths from `workspace/` to `workspaces/<scope-name>/`

### Issue: State File Conflicts

**Symptom:** `.repo-claude-state.json` errors

**Solution:**
```bash
# Reset state file
rm .repo-claude-state.json
rc status  # Regenerates state
```

## Benefits After Migration

1. **Parallel Development**: Work on multiple features simultaneously
2. **Clean Isolation**: No conflicts between different work streams
3. **Better Organization**: Clear scope-based structure
4. **Flexible Workflows**: Mix persistent and ephemeral scopes
5. **E-Commerce Ready**: Pre-configured for common e-commerce patterns

## Rollback Plan

If you need to rollback to v1:

```bash
# Restore backup
rm -rf my-project
mv my-project-v1-backup my-project

# Or from archive
tar -xzf my-project-v1-backup.tar.gz
```

## Getting Help

For migration assistance:
- Review [Architecture Documentation](architecture.md)
- Check [Scope Management Guide](scope-management.md)
- See [Workspace Structure](workspace-structure.md)
- Refer to example configuration in `examples/scope-isolation-config.yaml`