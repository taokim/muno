# MUNO User Guide

This comprehensive guide helps developers and AI agents effectively use MUNO for multi-repository management.

## What is MUNO?

**MUNO** (Multi-repository UNified Orchestration) organizes multiple repositories in a tree structure, similar to how an octopus (문어 in Korean) coordinates its arms. It provides intelligent repository orchestration for transforming chaotic repository collections into well-organized workspaces.

### Key Features
- **Tree-based organization** - Repositories organized in hierarchical structure
- **Lazy loading** - Repositories clone only when needed
- **CWD-first resolution** - Commands operate based on current directory
- **Simple management** - Direct git operations at any node

## Quick Start

### Basic Setup

```bash
# Initialize a workspace
muno init my-workspace
cd my-workspace

# Add repositories
muno add https://github.com/org/frontend.git
muno add https://github.com/org/backend.git --lazy

# View structure
muno tree

# Clone lazy repositories when needed
muno clone --recursive

# Update all repos
muno pull --recursive
```

## Repository Organization Strategies

### 1. Team-Based Organization

Organize by team ownership:

```yaml
workspace/
├── .nodes/
│   ├── team-backend/
│   │   ├── .nodes/
│   │   │   ├── payment-service/
│   │   │   ├── order-service/
│   │   │   └── shared-libs/
│   ├── team-frontend/
│   │   ├── .nodes/
│   │   │   ├── web-app/
│   │   │   └── mobile-app/
│   └── team-platform/
│       ├── .nodes/
│       │   ├── auth-service/
│       │   └── api-gateway/
```

**Implementation:**
```bash
# Initialize workspace
muno init my-platform
cd my-platform

# Create backend team structure
mkdir -p .nodes/team-backend
cd .nodes/team-backend
muno init team-backend
muno add https://github.com/org/payment-service.git
muno add https://github.com/org/order-service.git
muno add https://github.com/org/shared-libs.git --lazy

# Create frontend team structure
cd ../..
mkdir -p .nodes/team-frontend
cd .nodes/team-frontend
muno init team-frontend
muno add https://github.com/org/web-app.git
muno add https://github.com/org/mobile-app.git
```

### 2. Service-Type Organization

Organize by architectural layers:

```yaml
workspace/
├── .nodes/
│   ├── apis/
│   │   ├── .nodes/
│   │   │   ├── payment-api/
│   │   │   └── order-api/
│   ├── frontends/
│   │   ├── .nodes/
│   │   │   ├── web-app/
│   │   │   └── mobile-app/
│   └── libraries/
│       ├── .nodes/
│       │   ├── shared-utils/
│       │   └── api-contracts/
```

### 3. Domain-Based Organization

Organize by business domains:

```yaml
workspace/
├── .nodes/
│   ├── commerce/
│   │   ├── .nodes/
│   │   │   ├── payment/
│   │   │   ├── order/
│   │   │   └── inventory/
│   ├── identity/
│   │   ├── .nodes/
│   │   │   ├── auth/
│   │   │   └── users/
│   └── platform/
│       ├── .nodes/
│       │   ├── monitoring/
│       │   └── api-gateway/
```

## Migration Scenarios

### From Google Repo

If you have a `manifest.xml` from Google Repo:

1. Parse the manifest to understand project structure
2. Create MUNO workspace with similar organization
3. Add repositories maintaining relationships

```bash
# Create workspace
muno init migrated-workspace
cd migrated-workspace

# For each project group, create structure
mkdir -p .nodes/backend
cd .nodes/backend
muno init backend
# Add repositories from that group
muno add https://github.com/org/payment.git
muno add https://github.com/org/order.git --lazy
```

### From Flat Repository Structure

When all repositories are in a single directory:

```bash
# Discover existing repos
find . -name ".git" -type d | while read gitdir; do
    echo "$(dirname "$gitdir")"
done

# Create organized workspace
muno init organized-workspace
cd organized-workspace

# Group by patterns and create structure
mkdir -p .nodes/backend-services
cd .nodes/backend-services
muno init backend-services
# Add backend repositories

cd ../..
mkdir -p .nodes/frontend-apps
cd .nodes/frontend-apps
muno init frontend-apps
# Add frontend repositories
```

## Common Workflows

### Setting Up a New Project

```bash
# 1. Initialize workspace
muno init my-project
cd my-project

# 2. Create logical structure
mkdir -p .nodes/services .nodes/libraries .nodes/infrastructure

# 3. Add services
cd .nodes/services
muno init services
muno add https://github.com/org/api.git
muno add https://github.com/org/worker.git

# 4. Add libraries (lazy loaded)
cd ../libraries
muno init libraries
muno add https://github.com/org/common.git --lazy
muno add https://github.com/org/contracts.git --lazy

# 5. Work in specific repository
cd ../services/.nodes/api
# Make changes, commit, push as normal
```

### Managing Large Repository Collections (100+ repos)

1. **Analyze patterns** in repository names:
   - Language suffixes: `-go`, `-java`, `-js`, `-py`
   - Team prefixes: `team-`, ownership patterns
   - Service types: `-api`, `-service`, `-frontend`, `-lib`

2. **Create hierarchical organization**:
   - Use 2 levels for <20 repos
   - Use 3 levels for 20-100 repos
   - Use 4+ levels for 100+ repos

3. **Apply lazy loading strategy**:
   - Documentation repos → lazy
   - Archived/deprecated → lazy
   - Large datasets → lazy
   - Core services → not lazy

## Essential Commands

```bash
# Workspace Management
muno init <workspace>          # Initialize new workspace
muno tree                       # Display repository tree
muno status                     # Show tree and repository status

# Repository Management
muno add <url>                  # Add repository
muno add <url> --lazy           # Add lazy-loaded repository
muno remove <name>              # Remove repository
muno list                       # List child nodes

# Git Operations
muno clone --recursive          # Clone lazy repositories
muno pull --recursive           # Pull all repositories
muno push                       # Push changes
muno commit -m "message"        # Commit changes

# Navigation (using standard shell)
cd .nodes/<path>                # Navigate to repository
pwd                             # Show current location
```

## Configuration Examples

### Basic Configuration

```yaml
# muno.yaml
workspace:
  name: my-project
  repos_dir: .nodes  # Where to place child repos (default)

nodes:
  - name: backend
    url: https://github.com/org/backend.git
    lazy: false
  - name: frontend
    url: https://github.com/org/frontend.git
    lazy: true
```

### Multi-Team Configuration

```yaml
# Root muno.yaml
workspace:
  name: platform

nodes:
  - name: team-a
    file: ./teams/a/muno.yaml  # Delegate to team config
  - name: team-b
    file: ./teams/b/muno.yaml
  - name: shared
    url: https://github.com/org/shared.git
```

## Best Practices

### 1. Repository Analysis
Before organizing, analyze repositories for:
- Naming patterns (team, service type, technology)
- Dependencies (check package files)
- Team ownership (CODEOWNERS files)
- Activity level (recent commits)
- Size (large repos need special handling)

### 2. Lazy Loading Strategy
Mark as lazy when repositories are:
- Documentation or wikis
- Deprecated or archived
- Large binary assets
- Rarely modified libraries
- Test data repositories

### 3. Naming Conventions
- Parent nodes: `team-backend`, `domain-commerce`
- Keep original repository names for familiarity
- Use consistent patterns across the tree

### 4. Tree Depth Guidelines
- **2 levels**: Simple projects (<20 repos)
- **3 levels**: Standard organizations (20-100 repos)
- **4+ levels**: Large enterprises (100+ repos)

## Validation Checklist

After organizing repositories:

1. **Structure Validation**:
```bash
muno tree                      # Verify tree structure
muno list                      # Check immediate children
```

2. **Repository Status**:
```bash
muno status                    # Check all repo status
muno pull --recursive          # Test pull operations
```

3. **Navigation Testing**:
```bash
cd .nodes/team-backend/.nodes/payment  # Test navigation
pwd                                    # Verify location
cd ../../..                           # Return to root
```

## Common Issues and Solutions

| Issue | Solution |
|-------|----------|
| Too many repos at root | Create logical groupings and redistribute |
| Deep nesting difficult | Flatten structure or create shortcuts |
| Lazy repos not cloning | Check URLs and use `muno clone --recursive` |
| Team boundaries unclear | Use separate muno.yaml files per team |

## Summary

MUNO transforms multi-repository chaos into organized, navigable workspaces through:

1. **Tree-based organization** matching your mental model
2. **Lazy loading** for efficient resource usage
3. **CWD-first operations** for intuitive navigation
4. **Flexible configuration** supporting various organizational patterns

Start simple with basic repository additions, then evolve your structure as needs grow. The key is choosing organization strategies that match your team structure and workflow requirements.