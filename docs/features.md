# Repo-Claude Features

## Overview

Repo-Claude provides comprehensive multi-repository management with integrated CI/CD workflows, making it easy to coordinate changes across microservices, monorepos, and multi-team projects.

## Core Features

### 🔄 Git Synchronization with Rebase

**Default rebase mode for cleaner history**
- `rc sync` uses `git pull --rebase` by default
- Maintains linear commit history
- Avoids unnecessary merge commits
- Ideal for trunk-based development workflows

### 🌿 Branch Management

**Coordinate branches across all repositories**

#### Create Branches
```bash
rc branch create feature/payment-integration
```
- Creates the same branch in all repositories
- Option to specify base branch with `--from`
- Target specific repos with `--repos`

#### List Branch Status
```bash
rc branch list
```
- Visual indicators: 🔵 main branch, 🟢 feature branch
- Shows all repos at a glance
- `--all` flag to see all branches
- Helpful tip when feature branches are detected

#### Checkout Branches
```bash
rc branch checkout develop
```
- Switch branches across all repos simultaneously
- `--create` flag to create if missing
- Ensures consistency across workspace

#### Delete Branches
```bash
rc branch delete feature/old --remote
```
- Clean up branches across all repos
- Optional `--force` for unmerged branches
- `--remote` to also delete from origin

### 📋 Pull Request Management

**Centralized PR workflows using GitHub CLI**

#### Batch PR Creation
```bash
rc pr batch-create --title "Add payment integration"
```

**Key Features:**
- Creates PRs for all repositories on feature branches
- Automatic safety checks:
  - ✅ Skips repositories on main/master branch
  - ✅ Detects uncommitted changes
  - ✅ Auto-pushes unpushed branches
  - ✅ Shows detailed results per repository

**Options:**
- `--base`: Specify target branch
- `--draft`: Create as draft PRs
- `--reviewers`: Request reviews
- `--assignees`: Assign to users
- `--labels`: Add labels
- `--repos`: Target specific repositories
- `--skip-main-check`: Override safety check (use with caution)

#### Individual PR Creation
```bash
rc pr create --repo backend --title "Fix authentication bug"
```
- Interactive or CLI-driven
- Full GitHub PR features support

#### PR Status Monitoring
```bash
rc pr list --author @me
rc pr status --repo backend --pr 42
```
- List PRs across all repositories
- Check CI/CD status and reviews
- Filter by state, author, assignee, labels

#### PR Review Workflow
```bash
rc pr checkout 42 --repo backend
rc pr merge 42 --repo backend --squash
```
- Checkout PR branches locally
- Multiple merge strategies (merge, squash, rebase)
- Auto-delete branches after merge

## Workflow Examples

### Feature Development Workflow

Perfect for coordinated feature development across microservices:

```bash
# 1. Create feature branches
rc branch create feature/payment-api

# 2. Develop and commit changes
# ... make changes in each repo ...
rc forall -- git add -A
rc forall -- git commit -m "Implement payment processing"

# 3. Push all branches
rc forall -- git push -u origin HEAD

# 4. Create PRs for review
rc pr batch-create \
  --title "Add payment processing API" \
  --body "Implements payment gateway integration across services" \
  --reviewers senior-dev \
  --labels enhancement,backend

# 5. Monitor PR status
rc pr list --author @me
rc pr status

# 6. After approval, merge PRs
rc pr merge 123 --repo backend --squash
rc pr merge 124 --repo frontend --squash

# 7. Clean up branches
rc branch delete feature/payment-api --remote
```

### Hotfix Workflow

Quick fixes across multiple services:

```bash
# 1. Create hotfix branches from main
rc branch create hotfix/security-patch --from main

# 2. Apply fixes and commit
rc forall -- git add -A
rc forall -- git commit -m "Fix security vulnerability CVE-2024-XXX"

# 3. Create PRs with urgency
rc pr batch-create \
  --title "URGENT: Security patch CVE-2024-XXX" \
  --body "Critical security fix" \
  --labels security,urgent

# 4. Fast-track review and merge
rc pr list --label urgent
# ... after quick review ...
rc forall -- gh pr merge --auto --squash
```

### Dependency Update Workflow

Update dependencies across all services:

```bash
# 1. Create update branch
rc branch create chore/update-dependencies

# 2. Update dependencies in each repo
rc forall -- npm update  # or appropriate package manager

# 3. Run tests
rc forall -- npm test

# 4. Commit changes
rc forall -- git add -A
rc forall -- git commit -m "chore: update dependencies"

# 5. Create PRs
rc pr batch-create \
  --title "Update dependencies Q4 2024" \
  --body "Quarterly dependency updates for security and performance" \
  --labels dependencies,maintenance
```

## Safety Features

### Main Branch Protection
- Batch PR creation automatically skips repositories on main/master
- Prevents accidental PRs from production branches
- Override available with `--skip-main-check` flag

### Change Detection
- Warns about uncommitted changes before PR creation
- Prevents incomplete PRs
- Ensures clean working directories

### Automatic Branch Management
- Auto-pushes local branches before PR creation
- Handles branch tracking setup
- Manages remote branch lifecycle

### Visual Feedback
- Clear status indicators with emojis
- Detailed success/failure reporting
- Summary statistics for batch operations

## Integration Benefits

### For Trunk-Based Development
While TBD favors direct commits to main, PR features support:
- Code review for significant changes
- External contributor workflows
- Cross-team collaboration
- Changes requiring discussion or approval

### For Microservices Architecture
- Coordinate API changes across services
- Synchronized feature rollouts
- Consistent dependency updates
- Unified security patches

### For Multi-Team Projects
- Clear ownership with PR assignments
- Parallel development with branch isolation
- Controlled integration through PR reviews
- Audit trail of changes

## Requirements

### Git Operations
- Git 2.0 or later
- Configured remotes for each repository

### Pull Request Features
- GitHub CLI (`gh`) installed and authenticated
- GitHub as remote repository host
- Appropriate repository permissions

## Best Practices

### Branch Naming
- Use consistent prefixes: `feature/`, `bugfix/`, `hotfix/`, `chore/`
- Include ticket numbers when applicable: `feature/JIRA-123-payment`
- Keep names short but descriptive

### PR Titles and Descriptions
- Use conventional commit format for titles
- Include context in PR body
- Reference related issues or tickets
- Add appropriate labels for filtering

### Workflow Automation
- Use `rc forall` for repetitive commands
- Leverage batch operations for efficiency
- Set up consistent reviewer groups
- Automate with CI/CD where possible

## Comparison with Alternatives

### vs. Manual Repository Management
- **10x faster** for multi-repo operations
- **Consistency** across all repositories
- **Safety checks** prevent common mistakes
- **Unified view** of entire workspace

### vs. Monorepo
- **Maintains separation** between teams/services
- **Independent versioning** and deployment
- **Flexible ownership** models
- **Gradual migration** possible

### vs. Git Submodules
- **Simpler workflow** without submodule complexity
- **Better PR integration** with GitHub
- **Parallel operations** for performance
- **Cleaner history** with rebase default