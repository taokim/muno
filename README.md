# Repo-Claude

Transform your multi-repository chaos into a unified development experience with AI-powered scopes.

## The Problem

While Claude and other AI coding assistants excel at working within a single repository, modern software development often spans multiple repositories:
- Microservices architectures with dozens of service repos
- Frontend/backend separation across different repositories  
- Shared libraries and components in separate repos
- Mobile apps alongside web services

**The Challenge**: AI agents struggle to understand and work across repository boundaries, missing critical context and dependencies that span multiple repos.

## The Solution

Repo-Claude brings Google's proven multi-repository management approach (used for Android's 1000+ repos) to AI agent orchestration:

### 1. **Monorepo-Like Organization**
Transform multiple Git repositories into a unified workspace while keeping them separate:
- Single configuration file defines all repositories and their relationships
- Manage dozens of repos with simple commands (`rc pull --clone-missing`, `rc status`)
- Each repo maintains its independence while being part of a cohesive whole

### 2. **Scope-Based AI Development**
Work with flexible scopes that span multiple repositories:
- Define scopes by feature, team, or domain (e.g., "backend", "order-flow", "mobile")
- Each scope includes all relevant repositories for a complete context
- Claude Code sessions run in current terminal by default, with automatic new windows for multiple sessions
- Dynamic scope switching coming in Phase 2

> **Inspired by [Google's Repo tool](https://gerrit.googlesource.com/git-repo/)** - We've adapted Repo's battle-tested multi-repository management concepts for the AI era, creating a monorepo-like experience without the complexity of actual monorepos.

## Features

- 🚀 **Single Binary**: No runtime dependencies (Python, Node.js, etc.)
- 🗂️ **Multi-Repository Management**: Manage dozens of Git repositories as one unified workspace
- 🤖 **Scope-Based Orchestration**: Launch Claude Code sessions with multi-repository context
- 🎨 **Interactive TUI**: Beautiful terminal UI for selecting scopes and repos (powered by Bubbletea)
- 🔧 **Simple Git Operations**: Direct git commands with parallel execution for speed
- 🌳 **Trunk-Based Development**: All agents work directly on main branch
- 📝 **Shared Memory**: Cross-scope coordination through shared memory file
- ⚡ **Fast**: Written in Go for optimal performance
- 🎯 **Easy Configuration**: Single YAML file controls everything

## Key Features

### 🚀 Multi-Repository Orchestration
- **Unified workspace** for multiple Git repositories
- **Scope-based AI sessions** that see multiple repos as one project
- **Parallel operations** across all repositories
- **Shared memory** for cross-repository coordination

### 🌿 Branch Management
- **Create branches** across all repos with one command
- **Synchronized checkouts** to keep repos in sync
- **Batch operations** for branch lifecycle management
- **Visual status** overview of all repository branches

### 📋 Pull Request Automation
- **Batch PR creation** for coordinated changes
- **Safety checks** to prevent PRs from main branch
- **Centralized PR management** across all repositories
- **GitHub CLI integration** for full PR workflow

### 🔄 Smart Git Operations
- **Rebase by default** for cleaner commit history
- **Parallel sync** for fast repository updates
- **Automatic push** when creating PRs
- **Conflict detection** before operations

## Prerequisites

- Git
- [Claude Code CLI](https://claude.ai/code)
- [GitHub CLI](https://cli.github.com) (optional, for PR features)

## Installation

### Homebrew (Recommended)

```bash
brew tap taokim/tap
brew install repo-claude
```

### From Release

```bash
# Download latest release
# macOS (Intel)
curl -L https://github.com/taokim/repo-claude/releases/latest/download/repo-claude_Darwin_x86_64.tar.gz | tar xz

# macOS (Apple Silicon)
curl -L https://github.com/taokim/repo-claude/releases/latest/download/repo-claude_Darwin_arm64.tar.gz | tar xz

# Linux (x86_64)
curl -L https://github.com/taokim/repo-claude/releases/latest/download/repo-claude_Linux_x86_64.tar.gz | tar xz

# Then install
chmod +x rc
sudo mv rc /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/taokim/repo-claude.git
cd repo-claude
make build
sudo make install
```


## Quick Start

1. **Initialize a new workspace**:
   ```bash
   rc init my-project
   cd my-project
   ```

2. **Start scopes**:
   ```bash
   rc start              # Interactive selection UI (NEW DEFAULT!)
   rc start backend      # Start specific scope directly
   rc start order-service # Start scope containing this repo
   rc start --all        # Start all auto-start scopes (non-interactive)
   rc start -i backend   # Force interactive mode even with args
   rc start --new-window # Force open in new window
   rc start scope1 scope2 # Multiple scopes auto-open in new windows
   ```
   
   The new **interactive mode** (default when no arguments) provides a rich TUI experience:
   - Visual list of all scopes and repositories with status indicators
   - Multi-select with space bar, start with enter
   - Filter by running/stopped/scopes/repos with 'f' key
   - Smart grouping shows which repos belong to each scope
   - Keyboard shortcuts for select all ('a'), clear all ('n'), and help ('?')

3. **Check status**:
   ```bash
   rc status             # Show workspace configuration and repository status
   rc list               # List available scopes
   ```

4. **Manage branches** across repositories:
   ```bash
   rc branch create feature/payment   # Create branch in all repos
   rc branch list                     # Show branch status
   rc branch checkout main            # Switch branches
   rc branch delete feature/old       # Clean up branches
   ```

6. **Manage pull requests** (requires GitHub CLI):
   ```bash
   rc pr list            # List PRs across all repos
   rc pr batch-create --title "Add feature"  # Create PRs for all feature branches
   rc pr create --repo backend --title "Fix"  # Create PR for single repo
   rc pr status          # Show PR status with checks
   rc pr checkout 42 --repo backend  # Review PR locally
   rc pr merge 42 --repo backend     # Merge PR
   ```

## 🚀 Git Commands

Repo-Claude now includes powerful Git commands with **parallel execution** and **root repository support** by default:

### Core Git Operations

All Git commands now:
- ✅ **Include root repository by default** (your main project directory)
- ⚡ **Execute in parallel** for maximum speed (configurable concurrency)
- 🎯 **Support unified options** across all commands
- 📊 **Provide clear progress and summary reporting**

#### **Commit** - Stage and commit changes across all repos
```bash
rc commit -m "Update dependencies"           # Commit in all repos including root
rc commit -m "Fix bug" --exclude-root        # Skip root repository
rc commit -m "Feature X" --sequential        # Process repos one by one
rc commit -m "Refactor" --max-parallel 8     # Use 8 parallel operations
rc commit -m "Update" -v                     # Show detailed output
```

#### **Push** - Push commits to remote repositories
```bash
rc push                          # Push all repos with changes
rc push --exclude-root           # Skip root repository
rc push --sequential             # Push one repository at a time
rc push -v                       # Show detailed git output
rc push --max-parallel 10        # Use 10 parallel operations
```

#### **Pull** - Synchronize repositories (pull existing, optionally clone missing)
```bash
rc pull                    # Pull all existing repos
rc pull --clone-missing    # Clone missing repos, then pull all (replaces 'rc sync')
rc pull --rebase           # Pull with rebase (cleaner history)
rc pull --exclude-root     # Skip root repository
rc pull -v                 # Show detailed git output
rc pull --sequential       # Pull one repo at a time
rc pull --max-parallel 2   # Limit to 2 concurrent pulls
```

#### **Fetch** - Download changes without merging
```bash
rc fetch                   # Fetch default remote for all repos
rc fetch --all             # Fetch from all configured remotes
rc fetch --prune           # Remove deleted remote branches
rc fetch --all --prune     # Complete remote synchronization
rc fetch --exclude-root    # Skip root repository
rc fetch -v                # Show detailed fetch information
```

#### **ForAll** - Run ANY command across all repositories
```bash
rc forall -- git status              # Check status of all repos
rc forall -- git log --oneline -5    # Show recent commits
rc forall -- make test                # Run tests in all repos
rc forall -- npm install              # Install dependencies
rc forall -- rm -rf node_modules     # Clean up artifacts

# Advanced options
rc forall --exclude-root -- git pull # Skip root repository
rc forall --sequential -- make build # Build one at a time
rc forall --max-parallel 2 -- test   # Limit parallel execution
rc forall -v -- git fetch            # Verbose output
rc forall -q -- git gc               # Quiet mode
```

### Common Options for All Git Commands

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--exclude-root` | | Exclude root repository from operation | Include root |
| `--sequential` | | Run operations one at a time | Parallel |
| `--max-parallel N` | | Maximum concurrent operations | 4 |
| `--quiet` | `-q` | Suppress output | Normal output |
| `--verbose` | `-v` | Show detailed command output | Summary only |

### Performance Benefits

The new parallel execution provides significant performance improvements:
- **Sync 10 repos**: ~70% faster (10s → 3s)
- **Commit across 20 repos**: ~80% faster (20s → 4s)
- **Push to remotes**: ~75% faster with parallel execution

### Root Repository Support

By default, all Git commands now operate on the root repository (your main project directory) as well as workspace repositories. This is useful for:
- Monorepo-style projects where the root contains shared configuration
- Documentation repositories with a main README
- Projects where the root repository contains CI/CD configuration

Use `--exclude-root` to skip the root repository when needed.

## Example: E-Commerce Platform

Imagine you're building an e-commerce platform with:
- `api-gateway` - API gateway repository
- `user-service` - User authentication service
- `order-service` - Order management service  
- `inventory-service` - Inventory tracking
- `web-frontend` - React web application
- `mobile-app` - React Native mobile app

Without Repo-Claude, AI agents can only see one repository at a time. They miss critical context like:
- How the frontend calls the API gateway
- Shared data models between services
- Authentication flow across services
- Dependencies between order and inventory services

With Repo-Claude, you define scopes that group related repositories:
- **Backend Scope**: Includes all backend services for comprehensive API work
- **Frontend Scope**: Covers web and mobile apps with shared component awareness
- **Order Flow Scope**: Spans order, inventory, and payment services for feature work

All repositories remain separate (different teams can own them), but Claude Code sessions see all repositories in their scope - providing the exact context needed for each task.

## Configuration

The `repo-claude.yaml` file defines your workspace:

```yaml
workspace:
  name: my-project
  path: workspace              # Optional: custom workspace directory (default: "workspace")
  manifest:
    remote_name: origin
    remote_fetch: https://github.com/yourorg/
    default_revision: main
    projects:
      - name: api-gateway
        path: services/gateway
        groups: core,backend
      - name: user-service
        path: services/user
        groups: core,backend
      - name: order-service
        path: services/order
        groups: core,backend
      - name: inventory-service
        path: services/inventory
        groups: backend
      - name: web-frontend
        groups: frontend,web
      - name: mobile-app
        groups: frontend,mobile

scopes:
  backend:
    repos: ["*-service", "api-gateway"]  # Wildcards supported
    description: "Backend services and API gateway"
    model: claude-sonnet-4
    auto_start: true
  
  frontend:
    repos: ["web-frontend", "mobile-app"]
    description: "Web and mobile frontends"
    model: claude-sonnet-4
    auto_start: true
    dependencies: [backend]  # Backend must be running first
  
  order-flow:
    repos: ["order-service", "inventory-service", "api-gateway"]
    description: "Order processing flow across services"
    model: claude-opus-4     # Use more powerful model for complex work
    auto_start: false
```

### Configuration Reference

#### Workspace Configuration
| Key | Type | Required | Default | Description |
|-----|------|----------|---------|-------------|
| `workspace.name` | string | Yes | - | Name of your workspace |
| `workspace.path` | string | No | "workspace" | Directory where repos are cloned |
| `workspace.manifest.remote_name` | string | Yes | - | Name for the remote (usually "origin") |
| `workspace.manifest.remote_fetch` | string | Yes | - | Base URL for cloning repositories |
| `workspace.manifest.default_revision` | string | Yes | - | Default branch/tag to use (e.g., "main") |
| `workspace.manifest.projects` | array | Yes | - | List of repositories to manage |

#### Project Configuration
| Key | Type | Required | Default | Description |
|-----|------|----------|---------|-------------|
| `name` | string | Yes | - | Repository name |
| `path` | string | No | name | Custom path within workspace |
| `groups` | string | No | - | Comma-separated groups for organization |
| `revision` | string | No | default_revision | Custom branch/tag for this repo |

#### Scope Configuration
| Key | Type | Required | Default | Description |
|-----|------|----------|---------|-------------|
| `repos` | array | Yes | - | Repository names or patterns (supports wildcards) |
| `description` | string | Yes | - | Human-readable description of the scope |
| `model` | string | Yes | - | Claude model (e.g., "claude-sonnet-4") |
| `auto_start` | boolean | No | false | Start automatically with `rc start` |
| `dependencies` | array | No | [] | Scopes that must start before this one |

### How Scope Configuration Works

When you start a scope with `rc start backend`, repo-claude:

1. **Launches Claude Code** in current terminal (or new window with `--new-window` or when starting multiple scopes):
   ```bash
   claude --model claude-sonnet-4 --append-system-prompt "..."
   ```

2. **Sets environment variables** for the Claude session:
   - `RC_SCOPE_ID`: Unique identifier for the scope instance
   - `RC_SCOPE_NAME`: The scope name (e.g., "backend")
   - `RC_SCOPE_REPOS`: Comma-separated list of repositories in scope
   - `RC_WORKSPACE_ROOT`: Path where repositories are cloned (workspace directory)
   - `RC_PROJECT_ROOT`: Path where repo-claude.yaml is located (project root)

3. **Provides context** through:
   - **System prompt**: Includes scope name, description, and repositories
   - **CLAUDE.md files**: Created in each repository with workspace context
   - **Working directory**: Set to current directory (not locked to a single repo)

4. **Enables coordination** via `shared-memory.md` for cross-scope communication

### Repository Pattern Matching

Scopes support wildcards in repository patterns:
- `"*-service"` matches all repositories ending with "-service"
- `"api-*"` matches all repositories starting with "api-"
- `"user-service"` matches exactly "user-service"

### Available Claude Models
- `claude-sonnet-4` - Fast, efficient model for most tasks
- `claude-opus-4` - Most capable model for complex tasks
- `claude-haiku-4` - Fastest model for simple tasks

## Workspace Structure

After running `rc init my-project`, you'll have:

```
my-project/
├── repo-claude.yaml        # Configuration
├── .repo-claude-state.json # Scope state tracking
└── workspace/              # Default workspace directory
    ├── shared-memory.md    # Scope coordination
    ├── api-gateway/        # Repository (main branch)
    │   └── CLAUDE.md       # Workspace context
    ├── user-service/       # Repository (main branch)
    │   └── CLAUDE.md       # Workspace context
    ├── order-service/      # Repository (main branch)
    │   └── CLAUDE.md       # Workspace context
    └── web-frontend/       # Repository (main branch)
        └── CLAUDE.md       # Workspace context
```

Note: 
- The `rc` command is installed system-wide via Homebrew or manual installation
- Repositories are cloned into the `workspace/` subdirectory by default
- You can customize the workspace path in `repo-claude.yaml` if needed
- Claude Code sessions run in current terminal by default
- Multiple sessions automatically open in new windows
- Use `--new-window` to force new window for single session

## Migration from Agent-Based Configuration

Repo-Claude now uses a scope-based architecture instead of agent-based. To migrate:

1. **Update your configuration file**:
   - Replace the `agents:` section with `scopes:`
   - Define scopes that group related repositories
   - Remove `agent:` fields from projects

2. **Example migration**:
   ```yaml
   # Old agent-based config
   agents:
     backend-agent:
       model: claude-sonnet-4
       specialization: Backend development
   
   # New scope-based config
   scopes:
     backend:
       repos: ["*-service", "api-gateway"]
       description: "Backend services and API"
       model: claude-sonnet-4
   ```

3. **Command changes**:
   - `rc stop` → `rc kill`
   - `rc ps` now shows numbered output by default
   - `rc start` runs in current terminal by default

The tool maintains backwards compatibility with agent-based configs but will show "[legacy mode]" in output.

## Development

### Building

```bash
make build          # Build binary
make test           # Run tests
make test-short     # Run short tests
make coverage       # Generate coverage report
make lint           # Run linter
make build-all      # Build for all platforms
```

### Testing

```bash
# Run all tests
make test

# Run integration tests
go test -v ./test/...

# Run with coverage
make coverage
```

### Releasing

```bash
# Create a new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Create release
make release
```

## Architecture

Repo-Claude uses direct Git operations to manage multiple repositories. Each agent works in its own repository with awareness of other repositories through relative paths and shared memory.

Key components:
- **GitManager**: Handles all Git operations (clone, sync, status)
- **Manager**: Core orchestration logic
- **Config**: YAML configuration and state management
- **Scope**: Claude Code session management with multi-repository context
- **ProcessManager**: Handles terminal and window management

### Inspiration and History

This project was inspired by Google's Android Repo tool, adopting its multi-repository management concepts while simplifying for AI agent orchestration. We initially used Repo directly but found it added unnecessary complexity for our use case.

See our [detailed architecture documentation](docs/architecture.md) and [design decisions](docs/adr/) for more information.

## FAQ

### Why not use Google's Repo tool directly?

We initially used the Android Repo tool but found it incompatible with AI agent workflows for several reasons:

1. **No Root Documentation Support**: Repo's manifest repositories are designed purely for repository structure, not documentation. AI agents need global context documentation at the workspace root, which Repo's manifest-only approach doesn't support well.

2. **Complex Manifest Management**: Repo requires XML manifests in a separate git repository, adding complexity that confuses users who already have a simple YAML configuration.

3. **Unnecessary Features**: Repo's advanced features (branch management, code review upload, cherry-picking) aren't needed for AI agent orchestration.

4. **Heavy Dependencies**: Repo downloads ~30MB of Python scripts per workspace, while our Go implementation is a single 10MB binary.

5. **Poor AI Context Flow**: AI agents typically read documentation from the root directory first to understand the system. Repo's structure, with manifests in a separate repository, breaks this natural context flow.

See [ADR-001](docs/adr/001-simplify-git-management.md) for the detailed technical decision.

### Can I use this without AI agents?

Yes! While designed for AI agent orchestration, repo-claude works great for anyone managing multiple related repositories. Use it for:
- Keeping multiple microservices in sync
- Managing frontend/backend/mobile repos together
- Coordinating shared libraries with their consumers

### How does this compare to Git submodules?

Unlike submodules, repositories managed by repo-claude:
- Remain completely independent (no parent-child relationships)
- Can be at different branches/commits
- Don't require special git commands
- Are easier to work with for both humans and AI agents

### Is this a monorepo?

No, this creates a monorepo-like experience while keeping repositories separate. You get the benefits of unified tooling and visibility without the drawbacks of a true monorepo (huge repository size, complex permissions, slow clones).

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details