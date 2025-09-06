# MUNO AI Agent Context Guide

This comprehensive guide provides AI agents (Claude, Gemini, Cursor, etc.) with complete context for working with MUNO - a multi-repository orchestration tool with tree-based navigation.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Core Concepts](#core-concepts)
3. [Command Reference](#command-reference)
4. [Configuration Schema](#configuration-schema)
5. [Repository Organization Patterns](#repository-organization-patterns)
6. [Workflow Examples](#workflow-examples)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Quick Start

### What is MUNO?

**MUNO** (문어 in Korean, meaning "octopus") orchestrates multiple repositories like an octopus coordinating its eight arms. It's a **Multi-repository UNified Orchestration** tool that:

- Organizes repositories in a navigable tree structure
- Provides filesystem-like navigation (`muno use backend/payment`)
- Lazy-loads repositories only when needed
- Maintains clear workspace organization at scale

### Basic Usage

```bash
# Initialize workspace
muno init my-workspace
cd my-workspace

# Add repositories
muno add https://github.com/org/frontend.git
muno add https://github.com/org/backend.git --lazy

# Navigate
muno use frontend        # Navigate into frontend
muno current             # Show current position
muno use ..              # Go back to parent

# View structure
muno tree                # Display full tree
muno list                # List children of current node

# Work with repos
muno status --recursive  # Check git status
muno pull --recursive    # Update all repos
```

---

## Core Concepts

### 1. Tree-Based Navigation

MUNO treats repositories as nodes in a tree, similar to filesystem directories:

```
workspace/
├── team-backend/           # Parent node
│   ├── payment-service/    # Child repository
│   ├── order-service/      # Child repository  
│   └── shared-libs/        # Lazy repository (not cloned)
└── team-frontend/          # Parent node
    ├── web-app/            # Child repository
    └── mobile-app/         # Lazy repository
```

### 2. Node Types

#### Git Repository Nodes (`url` field)
```yaml
nodes:
  - name: payment-service
    url: https://github.com/org/payment.git
    lazy: true  # Optional, intelligent defaults
```

#### Config Reference Nodes (`config` field)
```yaml
nodes:
  - name: team-backend
    config: ./teams/backend/muno.yaml  # Delegate to another config
```

**Important**: Each node must have EITHER `url` OR `config`, never both.

### 3. Lazy Loading

Repositories marked as `lazy` clone only when accessed:
- Saves disk space and clone time
- Perfect for documentation, archived repos, large datasets
- Auto-clones when you navigate to them

### 4. CWD-First Resolution

Your current position determines command scope:
```bash
muno use backend/payment   # Navigate to payment service
muno pull                  # Pulls only payment service
muno pull --recursive      # Pulls payment and all children
```

---

## Command Reference

### Essential Commands

| Command | Description | Example |
|---------|-------------|---------|
| `init <name>` | Initialize new workspace | `muno init platform` |
| `use <path>` | Navigate to node | `muno use backend/api` |
| `current` | Show current position | `muno current` |
| `tree` | Display tree structure | `muno tree` |
| `list` | List child nodes | `muno list` |
| `add <url>` | Add repository | `muno add https://github.com/org/repo.git` |
| `add --config <name>` | Add config node | `muno add --config team-backend` |
| `remove <name>` | Remove node | `muno remove old-service` |
| `clone` | Clone lazy repos | `muno clone --recursive` |
| `status` | Git status | `muno status --recursive` |
| `pull` | Git pull | `muno pull --recursive` |
| `push` | Git push | `muno push --recursive` |
| `commit -m` | Git commit | `muno commit -m "update"` |

### AI Agent Commands

| Command | Description | Example |
|---------|-------------|---------|
| `agent <name> [path]` | Start any AI agent | `muno agent claude backend` |
| `claude [path]` | Start Claude CLI | `muno claude frontend/web` |
| `gemini [path]` | Start Gemini CLI | `muno gemini services` |

### Command Modifiers

| Flag | Description | Applies To |
|------|-------------|------------|
| `--lazy` | Mark repo as lazy-loaded | `add` |
| `--recursive` | Apply to entire subtree | `clone`, `pull`, `push`, `status` |
| `--` | Pass args to agent | `agent`, `claude`, `gemini` |

### Navigation Examples

```bash
# Absolute navigation
muno use /                      # Go to workspace root
muno use /backend/payment        # Go to specific path

# Relative navigation  
muno use backend                # Go to child 'backend'
muno use ../frontend            # Go to sibling 'frontend'
muno use ..                     # Go to parent

# Check position
muno current                    # Shows: /backend/payment
```

---

## Configuration Schema

### Complete muno.yaml Structure

```yaml
# Workspace configuration (all optional with defaults)
workspace:
  name: "my-workspace"        # Default: "muno-workspace"
  repos_dir: "nodes"          # Default: "nodes" - where repos are cloned
  root_repo: ""               # Default: "" - optional root git URL

# Node definitions
nodes:
  # Git repository node
  - name: "service-name"      # REQUIRED: unique identifier
    url: "git-url"           # Git repository URL
    lazy: true               # Optional, smart defaults:
                            #   true (default) for regular repos
                            #   false for *-monorepo, *-platform, etc.
  
  # Config reference node
  - name: "team-name"        # REQUIRED: unique identifier  
    config: "path/config.yaml" # Path to external config
```

### Smart Defaults

#### Lazy Loading Defaults
Repositories ending with these patterns default to **eager** (non-lazy) loading:
- `-monorepo`, `-munorepo`, `-muno`
- `-metarepo`, `-platform`, `-workspace`
- `-root-repo`

All other repositories default to **lazy** loading.

### Configuration Examples

#### Minimal (Uses All Defaults)
```yaml
workspace:
  name: my-platform
nodes:
  - name: frontend
    url: https://github.com/org/frontend.git
  - name: backend
    url: https://github.com/org/backend.git
```

#### Team-Based with Config Delegation
```yaml
workspace:
  name: engineering
nodes:
  - name: team-backend
    config: ./teams/backend/muno.yaml
  - name: team-frontend
    config: ./teams/frontend/muno.yaml
  - name: shared-platform     # Eager by default (ends with -platform)
    url: https://github.com/org/shared-platform.git
```

#### Explicit Lazy Control
```yaml
nodes:
  - name: core-services
    url: https://github.com/org/core.git
    lazy: false  # Override default, clone immediately
  
  - name: documentation
    url: https://github.com/org/docs.git
    lazy: true   # Redundant but explicit
```

---

## Repository Organization Patterns

### ⏺ MUNO Workspace Organization Guide

This guide provides battle-tested principles for transforming flat repository structures into efficient hierarchical MUNO workspaces.

#### Key Principles

1. **Use Config Nodes for Logical Grouping**
   - Create top-level config nodes for major systems/domains (e.g., moms, mwms, scm, infra)
   - Each config node references its own YAML file in a systems/ directory
   - This creates clean namespace separation: system/repository paths

2. **Leverage Default Lazy Loading**
   - DO NOT specify `lazy: true` or `lazy: false` for terminal nodes
   - MUNO defaults to lazy loading for all terminal nodes
   - Only override when absolutely necessary (rare cases)
   - This is crucial for large workspace trees to maintain efficiency

3. **Keep Configurations Minimal**
   - Each repository needs only:
     - `name`: Repository identifier
     - `url`: Git repository URL
     - Comment describing its purpose (for documentation)
   - Avoid redundant settings that match defaults

4. **Maintain Documentation**
   - Keep descriptive comments for each repository
   - Comments provide context without affecting configuration
   - Essential for team understanding and navigation

#### Implementation Example

**Step 1: Root Configuration**
```yaml
# muno.yaml
workspace:
  name: fse-workspace
  repos_dir: repos

nodes:
  - name: moms
    config: ./systems/moms.yaml

  - name: mwms
    config: ./systems/mwms.yaml

  - name: scm
    config: ./systems/scm.yaml

  - name: infra
    config: ./systems/infra.yaml
```

**Step 2: System Configurations**
```yaml
# systems/moms.yaml
workspace:
  name: moms-system

nodes:
  # Core backend service
  - name: backend
    url: git@github.com:musinsa/moms-be.git

  # API Gateway for external channels
  - name: gateway
    url: git@github.com:musinsa/moms-gateway.git

  # Financial settlement processing
  - name: settlement
    url: git@github.com:musinsa/mass-settlement.git
```

**Step 3: Directory Structure**
```
project-root/
├── muno.yaml              # Root configuration
├── systems/               # System-specific configs
│   ├── moms.yaml
│   ├── mwms.yaml
│   ├── scm.yaml
│   └── infra.yaml
└── repos/                 # MUNO-managed (auto-created)
    ├── moms/repos/...
    ├── mwms/repos/...
    ├── scm/repos/...
    └── infra/repos/...
```

#### Common Pitfalls to Avoid

**❌ DON'T Do This:**
```yaml
# Redundant lazy specifications
- name: backend
  url: git@github.com:org/repo.git
  lazy: true  # UNNECESSARY - this is default

# Overly complex nested configs
- name: connectors
  config: ./connectors/sap/config.yaml  # Too deep

# Missing comments
- name: rfid-connector
  url: git@github.com:org/rfid.git  # What does this do?
```

**✅ DO This Instead:**
```yaml
# Let defaults work
- name: backend
  url: git@github.com:org/repo.git

# Keep configs at one level
- name: sap-connector
  url: git@github.com:org/sap.git

# Include helpful comments
# RFID Tracking System
- name: rfid-connector
  url: git@github.com:org/rfid.git
```

#### Benefits of This Approach

1. **Scalability**: Works efficiently in massive workspace trees
2. **Clarity**: Clear hierarchical organization (system → repository)
3. **Maintainability**: Each system managed in its own config file
4. **Performance**: Default lazy loading reduces initial setup time
5. **Documentation**: Inline comments provide context without complexity

#### For AI Agents: Step-by-Step Process

When organizing workspaces with MUNO:
1. Start by identifying logical system boundaries
2. Create config nodes for major systems
3. Keep repository definitions minimal
4. Trust MUNO's defaults (especially lazy loading)
5. Add descriptive comments for human understanding
6. Test navigation paths before finalizing

### Pattern 1: Team-Based Organization

**Best for**: Clear team ownership, Conway's Law alignment

```yaml
# Root muno.yaml
nodes:
  - name: team-backend
    config: ./teams/backend/muno.yaml
  - name: team-frontend
    config: ./teams/frontend/muno.yaml
  - name: team-platform
    config: ./teams/platform/muno.yaml

# teams/backend/muno.yaml
nodes:
  - name: payment-service
    url: https://github.com/org/payment.git
  - name: order-service
    url: https://github.com/org/order.git
  - name: shared-libs
    url: https://github.com/org/backend-libs.git
    lazy: true
```

### Pattern 2: Service-Type Organization

**Best for**: Technology layers, architectural boundaries

```yaml
nodes:
  - name: apis
    config: ./layers/apis/muno.yaml
  - name: frontends
    config: ./layers/frontends/muno.yaml
  - name: libraries
    config: ./layers/libraries/muno.yaml
  - name: infrastructure
    config: ./layers/infrastructure/muno.yaml
```

### Pattern 3: Domain-Driven Organization

**Best for**: Microservices, bounded contexts

```yaml
nodes:
  - name: commerce
    config: ./domains/commerce/muno.yaml
  - name: identity
    config: ./domains/identity/muno.yaml
  - name: payments
    config: ./domains/payments/muno.yaml
  - name: platform
    config: ./domains/platform/muno.yaml
```

### Pattern 4: Environment-Based Organization

**Best for**: Multi-environment deployments

```yaml
nodes:
  - name: production
    config: ./envs/production/muno.yaml
  - name: staging
    config: ./envs/staging/muno.yaml
  - name: development
    config: ./envs/development/muno.yaml
```

---

## Workflow Examples

### Example 1: Setting Up E-commerce Platform

```bash
# Initialize workspace
muno init ecommerce-platform
cd ecommerce-platform

# Create team structures
muno add --config backend-services
muno add --config frontend-apps
muno add --config infrastructure

# Add backend services
muno use backend-services
muno add https://github.com/shop/catalog-api.git
muno add https://github.com/shop/cart-api.git
muno add https://github.com/shop/payment-api.git
muno add https://github.com/shop/order-api.git

# Add frontend applications
muno use ../frontend-apps
muno add https://github.com/shop/web-store.git
muno add https://github.com/shop/mobile-app.git --lazy
muno add https://github.com/shop/admin-panel.git --lazy

# Add infrastructure
muno use ../infrastructure
muno add https://github.com/shop/terraform.git
muno add https://github.com/shop/k8s-manifests.git
muno add https://github.com/shop/ci-cd.git --lazy

# Verify structure
muno use /
muno tree

# Start working
muno use backend-services/payment-api
muno claude  # Start Claude in payment API
```

### Example 2: Migrating from Flat Structure

```bash
# Analyze existing repos
find . -name ".git" -type d | while read git_dir; do
    repo_path=$(dirname "$git_dir")
    echo "Found: $repo_path"
done

# Initialize MUNO workspace
muno init organized-platform
cd organized-platform

# Create organization structure
muno add --config core-services
muno add --config supporting-services
muno add --config libraries
muno add --config tools

# Migrate core services
muno use core-services
for repo in auth user product order payment; do
    muno add "https://github.com/org/${repo}-service.git"
done

# Migrate supporting services
muno use ../supporting-services
for repo in email notification search analytics; do
    muno add "https://github.com/org/${repo}-service.git" --lazy
done

# Migrate libraries
muno use ../libraries
muno add https://github.com/org/common-utils.git
muno add https://github.com/org/api-contracts.git
muno add https://github.com/org/test-helpers.git --lazy

# Verify and work
muno use /
muno tree
muno status --recursive
```

### Example 3: Team Handoff Workflow

```bash
# Backend team creates their structure
muno init platform
cd platform
muno add --config team-backend
muno use team-backend

# Backend team adds their services
cat > muno.yaml << EOF
workspace:
  name: backend-services
nodes:
  - name: payment
    url: https://github.com/org/payment.git
  - name: order
    url: https://github.com/org/order.git
  - name: inventory
    url: https://github.com/org/inventory.git
EOF

# Commit and share config
git add muno.yaml
git commit -m "Backend team repository structure"
git push

# Frontend team references backend config
muno use /
muno add --config team-frontend
muno use team-frontend

# Frontend creates their config referencing shared libs
cat > muno.yaml << EOF
workspace:
  name: frontend-services
nodes:
  - name: web
    url: https://github.com/org/web.git
  - name: mobile
    url: https://github.com/org/mobile.git
  - name: shared-components
    url: https://github.com/org/components.git
    lazy: true
EOF
```

### Example 4: Bulk Repository Operations

```bash
# Clone all lazy repositories
muno tree | grep "lazy" | while read -r line; do
    path=$(echo $line | awk '{print $1}')
    muno use "$path" && muno clone
done

# Update all repositories
muno use /
muno pull --recursive

# Check status across entire tree
muno status --recursive

# Commit changes in multiple repos
for repo in payment order inventory; do
    muno use "backend/$repo"
    if [ -n "$(git status --porcelain)" ]; then
        muno commit -m "Bulk update: API version bump"
    fi
done

# Push all changes
muno use /
muno push --recursive
```

---

## Best Practices

### 1. Repository Organization

#### Naming Conventions
- **Parent nodes**: Use descriptive categories (`team-backend`, `domain-commerce`)
- **Service repos**: Keep original names for familiarity
- **Config nodes**: Use `.yaml` extension for clarity

#### Tree Depth Guidelines
- **2 levels**: Simple projects (<20 repos)
- **3 levels**: Standard organizations (20-100 repos)  
- **4+ levels**: Large enterprises (100+ repos)

#### Lazy Loading Strategy
Mark as lazy when repositories are:
- Documentation or wikis
- Archived or deprecated
- Large datasets or binaries
- Rarely modified libraries
- Test data repositories

### 2. Configuration Management

#### Use Config References for Team Autonomy
```yaml
# Root delegates to teams
nodes:
  - name: team-backend
    config: https://github.com/backend-team/muno-config/muno.yaml
  - name: team-frontend  
    config: https://github.com/frontend-team/muno-config/muno.yaml
```

#### Keep Configs Version Controlled
```bash
# Each team maintains their config
cd team-configs/backend
git add muno.yaml
git commit -m "Add payment service to backend tree"
git push
```

### 3. Development Workflows

#### Start Claude/Gemini in Context
```bash
# Navigate first, then start agent
muno use backend/payment
muno claude

# Or directly specify path
muno claude backend/payment
```

#### Parallel Operations
```bash
# Clone multiple repos in parallel
muno use backend
for repo in payment order inventory; do
    (muno use $repo && muno clone) &
done
wait
```

#### Scripted Automation
```bash
#!/bin/bash
# Update all services and run tests
muno use /
muno pull --recursive

# Run tests in each service
muno tree | grep -v "lazy" | while read -r path; do
    muno use "$path"
    if [ -f "Makefile" ]; then
        make test
    fi
done
```

### 4. Migration Strategies

#### From Google Repo
1. Parse `manifest.xml` to extract projects
2. Group by project-groups or paths
3. Create MUNO structure matching groups
4. Use lazy loading for inactive projects

#### From Monorepo
1. Use git subtree to split with history
2. Create new repositories for each service
3. Build MUNO structure matching architecture
4. Reference shared code via lazy loading

#### From Chaos (100+ repos)
1. Analyze patterns (language, team, type)
2. Create logical groupings
3. Start with core services (eager)
4. Add supporting services (lazy)
5. Iterate and refine structure

---

## Troubleshooting

### Common Issues and Solutions

| Issue | Solution |
|-------|----------|
| **Too many repos at root** | Create logical groupings, use config nodes |
| **Deep nesting hard to navigate** | Flatten structure or create navigation aliases |
| **Lazy repos not cloning** | Check URLs, network, and permissions |
| **Slow operations** | Use `--recursive` sparingly, leverage lazy loading |
| **Team boundaries unclear** | Use config references for delegation |
| **Can't find repository** | Use `muno tree | grep <name>` to locate |

### Validation Commands

```bash
# Verify structure
muno tree                     # Full tree view
muno list --recursive         # All nodes from current position

# Check repository status  
muno status --recursive       # Git status for all
muno pull --dry-run          # Test pull without executing

# Test navigation
muno use backend/payment
muno current                 # Should show: /backend/payment
muno use ../..              # Should return to root

# Validate configuration
cat muno.yaml | yaml-lint   # Check YAML syntax
muno tree | grep ERROR       # Find configuration errors
```

### Debug Mode

```bash
# Run with verbose output
MUNO_DEBUG=1 muno tree

# Check state file
cat .muno-tree.json | jq .

# Verify git operations
muno use backend/payment
git remote -v
git status
```

---

## Architecture Notes

### File Structure
```
workspace/
├── muno.yaml              # Main configuration
├── .muno-tree.json        # State tracking (auto-generated)
├── .muno/
│   └── agent-context.md   # This file (auto-embedded)
├── nodes/                 # Default repository directory
│   ├── backend/
│   │   ├── payment/
│   │   └── order/
│   └── frontend/
│       └── web/
└── teams/                 # Team-specific configs (optional)
    ├── backend/
    │   └── muno.yaml
    └── frontend/
        └── muno.yaml
```

### State Management
- Configuration: `muno.yaml` or `.muno.yaml`
- Tree state: `.muno-tree.json` (auto-maintained)
- Current position: Tracked in state file
- Repository status: Real-time from git

### Performance Considerations
- Lazy loading reduces initial setup time
- Parallel operations for bulk commands
- Recursive operations can be slow on large trees
- Use targeted operations when possible

---

## Advanced Features

### Custom Agent Integration
```bash
# Register custom agent
export MUNO_AGENTS="cursor:cursor-cmd,copilot:gh-copilot"

# Use custom agent
muno agent cursor backend
muno agent copilot frontend
```

### Repository Templates
```yaml
# Create template for microservices
# templates/microservice.yaml
nodes:
  - name: api
    url: "https://github.com/org/{{SERVICE}}-api.git"
  - name: worker
    url: "https://github.com/org/{{SERVICE}}-worker.git"
  - name: admin
    url: "https://github.com/org/{{SERVICE}}-admin.git"
    lazy: true
```

### Git Hooks Integration
```bash
# Add pre-commit hook to all repos
muno tree | grep -v "lazy" | while read -r path; do
    muno use "$path"
    cp ../../.pre-commit-config.yaml .
    pre-commit install
done
```

### CI/CD Integration
```yaml
# GitHub Actions example
name: MUNO Workspace CI
on: [push]
jobs:
  test-all:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup MUNO
        run: |
          go install github.com/musinsa/muno/cmd/muno@latest
          muno clone --recursive
      - name: Run tests
        run: |
          muno use /
          muno exec --recursive make test
```

---

## Summary

MUNO transforms repository chaos into organized, navigable workspaces. Key benefits:

1. **Organization**: Logical tree structure matching your mental model
2. **Efficiency**: Lazy loading saves time and space
3. **Navigation**: Filesystem-like movement between repositories
4. **Flexibility**: Config delegation enables team autonomy
5. **Integration**: Native support for AI agents and git operations

Start with analysis, design your structure, implement incrementally, and iterate based on team needs.

For technical details, see [CLAUDE.md](../CLAUDE.md) in the repository root.