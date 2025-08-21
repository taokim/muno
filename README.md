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
- Manage dozens of repos with simple commands (`rc sync`, `rc status`)
- Each repo maintains its independence while being part of a cohesive whole

### 2. **Scope-Based AI Development**
Work with flexible scopes that span multiple repositories:
- Define scopes by feature, team, or domain (e.g., "backend", "order-flow", "mobile")
- Each scope includes all relevant repositories for a complete context
- Claude Code sessions open in tabs by default for better workflow
- Dynamic scope switching coming in Phase 2

> **Inspired by [Google's Repo tool](https://gerrit.googlesource.com/git-repo/)** - We've adapted Repo's battle-tested multi-repository management concepts for the AI era, creating a monorepo-like experience without the complexity of actual monorepos.

## Features

- 🚀 **Single Binary**: No runtime dependencies (Python, Node.js, etc.)
- 🗂️ **Multi-Repository Management**: Manage dozens of Git repositories as one unified workspace
- 🤖 **Scope-Based Orchestration**: Launch Claude Code sessions with multi-repository context
- 🔧 **Simple Git Operations**: Direct git commands with parallel execution for speed
- 🌳 **Trunk-Based Development**: All agents work directly on main branch
- 📝 **Shared Memory**: Cross-scope coordination through shared memory file
- ⚡ **Fast**: Written in Go for optimal performance
- 🎯 **Easy Configuration**: Single YAML file controls everything

## Prerequisites

- Git
- [Claude Code CLI](https://claude.ai/code)

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
   rc start              # Start all auto-start scopes
   rc start backend      # Start specific scope
   rc start order-service # Start scope containing this repo
   rc start --new-window  # Open in new window instead of tab
   ```

3. **Check status**:
   ```bash
   rc ps                 # List running scopes with numbers
   rc status             # Show detailed workspace status
   ```

4. **Stop scopes**:
   ```bash
   rc kill               # Stop all running scopes
   rc kill backend       # Stop by name
   rc kill 1 2           # Stop by numbers from ps output
   ```

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

1. **Launches Claude Code** in a new terminal tab (or window with `--new-window`):
   ```bash
   claude --model claude-sonnet-4 --append-system-prompt "..."
   ```

2. **Sets environment variables** for the Claude session:
   - `RC_SCOPE_ID`: Unique identifier for the scope instance
   - `RC_SCOPE_NAME`: The scope name (e.g., "backend")
   - `RC_SCOPE_REPOS`: Comma-separated list of repositories in scope
   - `RC_WORKSPACE_ROOT`: Root directory of the workspace

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
- Claude Code sessions open in new tabs by default (use `--new-window` for separate windows)

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
   - `rc start` opens in tabs instead of windows

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
- **ProcessManager**: Handles terminal tab/window creation

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