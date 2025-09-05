# AI Agent Guide for MUNO Repository Organization

This guide helps AI agents (Claude, Gemini, Cursor, etc.) effectively organize and manage repositories using MUNO's tree-based structure. MUNO provides intelligent repository orchestration that can transform chaotic repository collections into well-organized, navigable workspaces.

## Table of Contents

1. [Understanding MUNO's Approach](#understanding-munos-approach)
2. [Repository Organization Strategies](#repository-organization-strategies)
3. [Migration Scenarios](#migration-scenarios)
4. [Step-by-Step Workflows](#step-by-step-workflows)
5. [Best Practices for AI Agents](#best-practices-for-ai-agents)
6. [Command Patterns](#command-patterns)
7. [Validation and Verification](#validation-and-verification)

## Understanding MUNO's Approach

### Core Concepts

MUNO treats repositories as nodes in a navigable tree structure, similar to a filesystem. Key principles:

- **Tree-based Navigation**: Move through repositories like directories (`muno use backend/payment`)
- **Lazy Loading**: Repositories clone only when needed, saving space and time
- **CWD-first Resolution**: Current position determines operation scope
- **Parent Nodes Matter**: Parent directories can be repositories themselves or organizational containers

### Why Tree Structure?

Traditional flat repository lists become unmanageable at scale. MUNO's tree structure provides:
- Logical grouping and hierarchy
- Team ownership boundaries
- Dependency relationships
- Scalable organization

## Repository Organization Strategies

### 1. Team-Based Organization

**When to use**: Clear team ownership, autonomous teams, Conway's Law alignment

```yaml
workspace/
├── team-backend/
│   ├── payment-service/
│   ├── order-service/
│   └── shared-libs/        # lazy loaded
├── team-frontend/
│   ├── web-app/
│   ├── mobile-app/
│   └── design-system/      # lazy loaded
└── team-platform/
    ├── auth-service/
    └── api-gateway/
```

**Implementation**:
```bash
# Initialize workspace
muno init my-platform

# Create team structures
muno add --config team-backend    # Parent node for backend team
muno use team-backend
muno add https://github.com/org/payment-service.git
muno add https://github.com/org/order-service.git
muno add https://github.com/org/shared-libs.git --lazy

# Navigate back and add frontend team
muno use ..
muno add --config team-frontend
muno use team-frontend
muno add https://github.com/org/web-app.git
muno add https://github.com/org/mobile-app.git
```

### 2. Service-Type Organization

**When to use**: Clear architectural layers, technology-focused teams

```yaml
workspace/
├── apis/
│   ├── payment-api/
│   ├── order-api/
│   └── auth-api/
├── frontends/
│   ├── web-app/
│   ├── mobile-app/
│   └── admin-dashboard/
├── libraries/
│   ├── shared-utils/
│   └── api-contracts/
└── infrastructure/
    ├── terraform-modules/
    └── k8s-manifests/
```

### 3. Domain-Based Organization

**When to use**: Domain-driven design, microservices architecture

```yaml
workspace/
├── commerce/
│   ├── payment/
│   ├── order/
│   └── inventory/
├── identity/
│   ├── auth/
│   ├── users/
│   └── permissions/
└── platform/
    ├── monitoring/
    └── api-gateway/
```

### 4. Multi-Cloud Organization

**When to use**: Multi-cloud deployments, provider-specific services

```yaml
workspace/
├── aws/
│   ├── production/
│   │   └── api-services/
│   └── staging/
│       └── test-services/
├── gcp/
│   └── ml-platform/
└── azure/
    └── legacy-services/
```

## Migration Scenarios

### Scenario 1: Migrating from Google Repo

**Context**: You have a Google Repo `manifest.xml` file with project definitions.

**Analysis Phase**:
1. Parse the manifest to understand project structure
2. Identify project groups and relationships
3. Map remotes to actual repository URLs

**Migration Strategy**:
```bash
# Create workspace
muno init migrated-workspace
cd migrated-workspace

# For each project group in manifest
# Example: backend group
muno add --config backend
muno use backend

# Add projects from that group
muno add https://github.com/org/payment.git
muno add https://github.com/org/order.git --lazy

# Continue for other groups...
```

### Scenario 2: Organizing 100+ Repositories

**Context**: Large collection of repositories with various characteristics.

**Analysis Phase**:
1. Categorize repositories by patterns:
   - Language: `-go`, `-java`, `-js`, `-py` suffixes
   - Team: `team-` prefixes, CODEOWNERS files
   - Service type: `-api`, `-service`, `-frontend`, `-lib`
   - Infrastructure: `terraform-`, `k8s-`, `helm-`

2. Detect relationships:
   - Shared libraries used by multiple services
   - API contracts between services
   - Build dependencies

**Organization Strategy**:
```bash
# Initialize workspace
muno init large-platform

# Create primary categorization (e.g., by team)
muno add --config team-payments
muno add --config team-identity
muno add --config team-platform

# Add repositories to appropriate teams
muno use team-payments
muno add https://github.com/org/payment-api.git
muno add https://github.com/org/payment-frontend.git
muno add https://github.com/org/payment-admin.git

# Mark rarely-used repos as lazy
muno add https://github.com/org/payment-docs.git --lazy
```

### Scenario 3: Splitting a Monorepo

**Context**: Breaking up a monolithic repository into services.

**Preparation**:
1. Identify service boundaries in the monorepo
2. Plan git history preservation strategy
3. Set up new repository locations

**Split Strategy**:
```bash
# In monorepo directory
# Extract services with history using git subtree
git subtree split --prefix=services/payment -b payment-service
git subtree split --prefix=services/order -b order-service

# Create and push to new repositories
cd ..
git clone --branch payment-service ./monorepo payment-service
cd payment-service
git remote set-url origin https://github.com/org/payment-service.git
git push -u origin main

# Set up MUNO workspace
cd ..
muno init platform
cd platform
muno add https://github.com/org/payment-service.git
muno add https://github.com/org/order-service.git
```

### Scenario 4: Migrating from Flat Structure

**Context**: All repositories cloned in a single directory.

**Discovery**:
```bash
# Find all git repositories
find . -type d -name ".git" | while read gitdir; do
    repo_path=$(dirname "$gitdir")
    remote_url=$(cd "$repo_path" && git remote get-url origin 2>/dev/null)
    echo "$repo_path: $remote_url"
done
```

**Organization**:
```bash
muno init organized-workspace
cd organized-workspace

# Group by detected patterns
muno add --config backend-services
muno use backend-services
# Add backend repositories...

muno use ..
muno add --config frontend-apps
muno use frontend-apps
# Add frontend repositories...
```

## Step-by-Step Workflows

### Workflow 1: Setting Up a New Project

```bash
# 1. Initialize workspace
muno init my-project
cd my-project

# 2. Create logical structure
muno add --config services
muno add --config libraries
muno add --config infrastructure

# 3. Add repositories to services
muno use services
muno add https://github.com/org/api.git
muno add https://github.com/org/worker.git

# 4. Add shared libraries (lazy loaded)
muno use ../libraries
muno add https://github.com/org/common.git --lazy
muno add https://github.com/org/contracts.git --lazy

# 5. Add infrastructure
muno use ../infrastructure
muno add https://github.com/org/terraform.git
muno add https://github.com/org/k8s.git

# 6. Navigate and work
muno use ../services/api
muno agent claude  # Start Claude in API service
```

### Workflow 2: Reorganizing Existing Repositories

```bash
# 1. Analyze current structure
ls -la  # List current repos
# Identify patterns and groupings

# 2. Create new MUNO workspace
muno init reorganized
cd reorganized

# 3. Create target structure
muno add --config core-services
muno add --config support-services
muno add --config infrastructure

# 4. Move repositories into structure
muno use core-services
muno add [core service URLs]

muno use ../support-services
muno add [support service URLs]

# 5. Verify organization
muno tree
muno status --recursive
```

## Best Practices for AI Agents

### 1. Repository Analysis

Before organizing, analyze repositories for:

- **Naming Patterns**: Detect team, service type, technology
- **Dependencies**: Check package.json, go.mod, requirements.txt
- **Team Ownership**: Look for CODEOWNERS, maintainers
- **Activity Level**: Recent commits indicate active development
- **Size**: Large repos might need different treatment

### 2. Lazy Loading Strategy

Mark repositories as lazy when they are:
- Documentation or wiki repositories
- Deprecated or archived
- Large binary assets or datasets
- Rarely modified shared libraries
- Test data repositories

```bash
# Examples of lazy loading patterns
muno add https://github.com/org/docs.git --lazy
muno add https://github.com/org/archived-service.git --lazy
muno add https://github.com/org/test-data.git --lazy
```

### 3. Tree Depth Guidelines

- **2 levels**: Simple projects (<20 repos)
- **3 levels**: Standard organizations (20-100 repos)
- **4+ levels**: Large enterprises (100+ repos)

Example 3-level structure:
```
Level 1: Team/Domain
Level 2: Service Category
Level 3: Individual Repository
```

### 4. Naming Conventions

Use consistent, descriptive names:
- Parent nodes: `team-backend`, `domain-commerce`, `layer-frontend`
- Service repos: Keep original names for familiarity
- Config nodes: Use `.yaml` extension for clarity

## Command Patterns

### Essential Commands for Organization

```bash
# Initialize and navigate
muno init <workspace>          # Create workspace
muno use <path>                 # Navigate to node
muno current                    # Show current position
muno tree                       # Display structure

# Add repositories
muno add <url>                  # Add repository
muno add <url> --lazy           # Add lazy-loaded repo
muno add --config <name>        # Add parent node

# Manage structure
muno remove <name>              # Remove node
muno list                       # List children
muno status --recursive         # Check git status

# Work with repos
muno clone                      # Clone lazy repos
muno pull --recursive           # Update all repos
muno agent claude [path]        # Start AI agent
```

### Advanced Patterns

```bash
# Bulk operations
muno use backend
for repo in payment order inventory; do
    muno add "https://github.com/org/${repo}-service.git"
done

# Conditional lazy loading
for repo in $(cat repos.txt); do
    if [[ $repo == *"-docs" ]] || [[ $repo == *"-deprecated" ]]; then
        muno add "$repo" --lazy
    else
        muno add "$repo"
    fi
done

# Parallel cloning
muno tree | grep "lazy" | while read -r path; do
    muno use "$path" && muno clone &
done
wait
```

## Validation and Verification

### Post-Migration Checklist

1. **Structure Validation**:
```bash
muno tree                      # Verify tree structure
muno list --lazy               # Check lazy repos
```

2. **Navigation Testing**:
```bash
muno use team-backend/payment  # Test navigation
muno current                   # Verify position
muno use ../..                 # Test parent navigation
```

3. **Repository Status**:
```bash
muno status --recursive        # Check all repo status
muno pull --dry-run           # Test pull operations
```

4. **Agent Integration**:
```bash
muno agent claude payment     # Test AI agent launch
muno gemini frontend/web      # Test different agents
```

### Common Issues and Solutions

**Issue**: Too many repositories at root level
**Solution**: Create logical groupings and redistribute

**Issue**: Deep nesting makes navigation difficult
**Solution**: Flatten structure or create shortcuts

**Issue**: Lazy repos not cloning when needed
**Solution**: Check repo URLs and network access

**Issue**: Team boundaries unclear
**Solution**: Use config nodes to delegate ownership

## Integration Tips

### For Claude Users
```bash
# Start Claude in specific repo
muno use backend/payment
muno claude

# Or directly
muno claude backend/payment
```

### For Gemini Users
```bash
# With Gemini CLI installed
muno gemini frontend/web

# Pass arguments
muno agent gemini -- --help
```

### For Custom Agents
```bash
# Use the agent command
muno agent <agent-name> [path]

# Examples
muno agent cursor backend
muno agent copilot frontend
```

## Examples

### Example 1: E-commerce Platform

```bash
muno init ecommerce
cd ecommerce

# Core commerce services
muno add --config commerce
muno use commerce
muno add https://github.com/shop/catalog.git
muno add https://github.com/shop/cart.git
muno add https://github.com/shop/checkout.git
muno add https://github.com/shop/payment.git

# Customer-facing apps
muno use ..
muno add --config storefront
muno use storefront
muno add https://github.com/shop/web.git
muno add https://github.com/shop/mobile.git
muno add https://github.com/shop/pwa.git --lazy

# Supporting services
muno use ..
muno add --config platform
muno use platform
muno add https://github.com/shop/auth.git
muno add https://github.com/shop/notifications.git
muno add https://github.com/shop/search.git --lazy
```

### Example 2: Microservices Platform

```bash
muno init microservices
cd microservices

# Domain services
for domain in user product order payment; do
    muno add --config "domain-$domain"
    muno use "domain-$domain"
    muno add "https://github.com/platform/${domain}-api.git"
    muno add "https://github.com/platform/${domain}-worker.git"
    muno add "https://github.com/platform/${domain}-admin.git" --lazy
    muno use ..
done

# Shared infrastructure
muno add --config shared
muno use shared
muno add https://github.com/platform/api-gateway.git
muno add https://github.com/platform/service-mesh.git
muno add https://github.com/platform/observability.git
```

## Conclusion

MUNO transforms repository chaos into organized, navigable workspaces. By following these guidelines, AI agents can help users:

1. Analyze existing repository structures
2. Design optimal organization strategies
3. Execute migrations systematically
4. Validate the final structure
5. Maintain organization over time

The key is choosing the right organization strategy based on team structure, architecture patterns, and workflow requirements. Start with analysis, plan the structure, execute incrementally, and validate thoroughly.

For more information, see:
- [MUNO README](../README.md)
- [MUNO Commands](./COMMANDS.md)
- [Examples](../examples/)