# Scope Management Guide

## Overview

Scopes are isolated workspace environments in Repo-Claude v2 that enable parallel development across multiple repositories. Each scope operates independently with its own Git state, allowing multiple features, hotfixes, or experiments to proceed without conflicts.

## Scope Types

### Persistent Scopes
Long-lived development environments for ongoing work:
- Feature development
- Regular maintenance
- Team-specific workspaces
- Domain-focused development (WMS, OMS, Catalog)

### Ephemeral Scopes
Temporary workspaces for short-term tasks:
- Hotfixes
- Experiments
- Proof of concepts
- One-time migrations

**Note**: No TTL (Time-To-Live) is implemented. All scopes are manually managed.

## Commands

### Creating Scopes

#### From Configuration
```bash
# Create a pre-configured scope
rc scope create wms --from-config

# This uses the scope definition from repo-claude.yaml:
# scopes:
#   wms:
#     type: persistent
#     repos: ["wms-*", "shared-libs"]
#     description: "Warehouse Management System"
```

#### Custom Scope
```bash
# Create a custom persistent scope
rc scope create feature-xyz --type persistent \
  --repos "wms-core,wms-inventory,shared-libs" \
  --description "New inventory feature"

# Create an ephemeral hotfix scope
rc scope create hotfix-payment --type ephemeral \
  --repos "oms-payment,oms-core" \
  --description "Fix payment processing bug"
```

#### With Wildcards
```bash
# Create scope with all WMS repositories
rc scope create wms-refactor --type persistent \
  --repos "wms-*,shared-libs"

# Create scope with all backend services
rc scope create backend-upgrade --type persistent \
  --repos "*-core,*-service,api-gateway"
```

### Listing Scopes

```bash
# List all scopes (configured and created)
rc list

# Output example:
# CONFIGURED SCOPES (from repo-claude.yaml):
# - wms         [persistent] Warehouse Management System
# - oms         [persistent] Order Management System  
# - catalog     [persistent] Product Catalog Management
# - hotfix      [ephemeral]  Emergency hotfix template
#
# CREATED SCOPES (in workspaces/):
# - wms-feature-inventory  [active]    Feature: New inventory tracking
# - oms-payment-hotfix     [inactive]  Hotfix: Payment processing
# - search-experiment      [archived]  Experiment: New search algorithm
```

### Starting Scopes

```bash
# Interactive selection
rc start

# Start specific scope
rc start wms-feature-inventory

# Start with specific working directory
rc start wms-feature --dir ./workspaces/wms-feature/wms-core
```

### Managing Scope State

#### Archive a Scope
```bash
# Mark scope as archived (preserves data)
rc scope archive wms-old-feature

# Archived scopes:
# - Are marked in .scope-meta.json
# - Don't appear in active lists by default
# - Can be restored later
```

#### Delete a Scope
```bash
# Permanently delete scope and all its repositories
rc scope delete search-experiment

# Warning: This removes:
# - The entire scope directory
# - All cloned repositories
# - Any uncommitted changes
# - Scope metadata
```

### Git Operations Within Scopes

```bash
# Pull updates for all repos in scope
rc pull --scope wms-feature

# Commit changes across scope repositories
rc commit --scope wms-feature -m "Add inventory tracking"

# Push changes from scope
rc push --scope wms-feature

# Create branch in scope repositories
rc branch --scope wms-feature feature/inventory-v2

# Check status of scope repositories
rc status --scope wms-feature
```

## Scope Lifecycle

### 1. Creation
```bash
rc scope create my-feature --type persistent --repos "..."
```
- Creates directory: `workspaces/my-feature/`
- Generates `.scope-meta.json`
- Clones specified repositories
- Creates CLAUDE.md in each repo

### 2. Development
```bash
rc start my-feature
```
- Launches Claude Code session
- Sets environment variables
- Provides scope context

### 3. Maintenance
```bash
# Regular updates
rc pull --scope my-feature

# Save work
rc commit --scope my-feature -m "Progress update"
rc push --scope my-feature
```

### 4. Completion
```bash
# Option 1: Archive (preserve for reference)
rc scope archive my-feature

# Option 2: Delete (remove completely)
rc scope delete my-feature
```

## E-Commerce Examples

### WMS Feature Development
```bash
# Create scope for new warehouse feature
rc scope create wms-barcode-scanning --type persistent \
  --repos "wms-core,wms-inventory,wms-ui,shared-libs" \
  --description "Implement barcode scanning feature"

# Start development
rc start wms-barcode-scanning

# ... work on feature ...

# Complete feature
rc push --scope wms-barcode-scanning
rc scope archive wms-barcode-scanning
```

### OMS Emergency Hotfix
```bash
# Create ephemeral scope for urgent fix
rc scope create oms-payment-fix-20240115 --type ephemeral \
  --repos "oms-payment,oms-core" \
  --description "Fix critical payment gateway timeout"

# Start fixing
rc start oms-payment-fix-20240115

# ... implement fix ...

# Deploy and cleanup
rc push --scope oms-payment-fix-20240115
rc scope delete oms-payment-fix-20240115  # Clean up ephemeral scope
```

### Search Performance Optimization
```bash
# Create scope for performance work
rc scope create search-perf --type persistent \
  --repos "search-engine,search-indexer,catalog-api" \
  --description "Optimize search query performance"

# Work on optimization
rc start search-perf

# Test different approaches
rc branch --scope search-perf experiment/caching
rc branch --scope search-perf experiment/indexing

# Keep scope for ongoing optimization
```

### Cross-Domain Integration
```bash
# Create scope for integration work
rc scope create integration-api --type persistent \
  --repos "api-gateway,wms-core,oms-core,catalog-service" \
  --description "Implement unified API layer"

# Coordinate across services
rc start integration-api
```

## Scope Metadata

Each scope contains `.scope-meta.json`:

```json
{
  "id": "wms-feature-20240115-093045",
  "name": "wms-feature",
  "type": "persistent",
  "state": "active",
  "repos": [
    "wms-core",
    "wms-inventory",
    "wms-shipping",
    "shared-libs"
  ],
  "description": "New WMS feature development",
  "created_at": "2024-01-15T09:30:45Z",
  "updated_at": "2024-01-15T16:22:30Z",
  "created_by": "user@example.com",
  "model": "claude-3-5-sonnet-20241022"
}
```

## Best Practices

### 1. Naming Conventions
```bash
# Feature scopes
<domain>-feature-<name>      # wms-feature-tracking

# Hotfix scopes  
<domain>-hotfix-<date>        # oms-hotfix-20240115

# Experiment scopes
<domain>-exp-<description>    # search-exp-ml-ranking
```

### 2. Scope Hygiene
- Archive completed persistent scopes
- Delete ephemeral scopes after use
- Regular cleanup of abandoned scopes
- Document scope purpose clearly

### 3. Repository Selection
- Include all dependent repositories
- Use wildcards for related repos
- Don't forget shared libraries
- Consider API dependencies

### 4. Parallel Development
```bash
# Developer A: Feature work
rc scope create wms-feature-a --repos "wms-*"

# Developer B: Different feature
rc scope create wms-feature-b --repos "wms-*"

# Developer C: Hotfix
rc scope create wms-hotfix --repos "wms-core"

# All three can work simultaneously without conflicts
```

## Troubleshooting

### Scope Won't Start
```bash
# Check scope exists
rc list

# Verify repositories are cloned
ls workspaces/<scope-name>/

# Check scope metadata
cat workspaces/<scope-name>/.scope-meta.json
```

### Repository Missing
```bash
# Re-clone repositories for scope
rc pull --scope <scope-name> --clone-missing
```

### Cleanup Stale Scopes
```bash
# List all scopes
rc list

# Archive old scopes
rc scope archive old-feature-1
rc scope archive old-feature-2

# Delete ephemeral scopes
rc scope delete old-hotfix-1
rc scope delete old-experiment-1
```

## Advanced Usage

### Scope Templates
Define template scopes in configuration:
```yaml
scopes:
  feature-template:
    type: ephemeral
    repos: []  # Select at creation
    description: "Feature development template"
  
  hotfix-template:
    type: ephemeral
    repos: []  # Select at creation
    description: "Emergency hotfix template"
```

### Cross-Scope Coordination
Use `shared-memory.md` for coordination:
```markdown
# Shared Memory

## Active Development
- wms-feature-inventory: Implementing new tracking system
- oms-payment-hotfix: Fixing timeout issue (URGENT)

## Blocked
- search-optimization: Waiting for catalog API changes

## Notes
- Database migration scheduled for Sunday
- API freeze starts Monday
```

### Bulk Operations
```bash
# Pull all active scopes
for scope in $(rc list --active --json | jq -r '.[]'); do
  rc pull --scope "$scope"
done

# Archive all scopes older than 30 days
rc list --json | jq -r '.[] | select(.age_days > 30)' | while read scope; do
  rc scope archive "$scope"
done
```