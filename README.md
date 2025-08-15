# Repo-Claude

Transform your multi-repository chaos into a monorepo-like development experience with AI agents.

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

### 2. **Cross-Repository AI Agents**
Enable AI agents to work seamlessly across repository boundaries:
- Agents understand the full system context, not just individual repos
- Shared memory enables coordination between agents working on different repos
- Agents can reference code and dependencies across the entire workspace

> **Inspired by [Google's Repo tool](https://gerrit.googlesource.com/git-repo/)** - We've adapted Repo's battle-tested multi-repository management concepts for the AI era, creating a monorepo-like experience without the complexity of actual monorepos.

## Features

- 🚀 **Single Binary**: No runtime dependencies (Python, Node.js, etc.)
- 🗂️ **Multi-Repository Management**: Manage dozens of Git repositories as one unified workspace
- 🤖 **Multi-Agent Orchestration**: Coordinate multiple Claude Code instances across repositories
- 🔧 **Simple Git Operations**: Direct git commands with parallel execution for speed
- 🌳 **Trunk-Based Development**: All agents work directly on main branch
- 📝 **Shared Memory**: Cross-agent coordination through shared memory file
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

2. **Start agents**:
   ```bash
   rc start              # Start all auto-start agents
   rc start backend-agent # Start specific agent
   ```

3. **Check status**:
   ```bash
   rc status
   ```

4. **Stop agents**:
   ```bash
   rc stop
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

With Repo-Claude, agents work across all repositories simultaneously:
- **Frontend Agent**: Updates web and mobile UI, understanding the full API structure
- **Backend Agent**: Modifies services with awareness of all consumers
- **API Agent**: Maintains gateway with knowledge of all services and clients

All repositories remain separate (different teams can own them), but AI agents see the complete system - just like developers working in a monorepo.

## Configuration

The `repo-claude.yaml` file defines your workspace:

```yaml
workspace:
  name: my-project
  manifest:
    remote_name: origin
    remote_fetch: https://github.com/yourorg/
    default_revision: main
    projects:
      - name: backend
        groups: core,services
        agent: backend-agent
      - name: frontend
        groups: core,ui
        agent: frontend-agent

agents:
  backend-agent:
    model: claude-sonnet-4
    specialization: API development, database design
    auto_start: true
  frontend-agent:
    model: claude-sonnet-4
    specialization: React/Vue development, UI/UX
    auto_start: true
    dependencies: [backend-agent]
```

## Workspace Structure

After running `rc init my-project`, you'll have:

```
my-project/
├── repo-claude.yaml        # Configuration
├── .repo-claude-state.json # Agent state tracking
└── workspace/              # Default workspace directory
    ├── shared-memory.md    # Agent coordination
    ├── .repo/              # Repo metadata
    │   └── manifests/      # Local git repo with manifest
    │       └── default.xml # Repo manifest
    ├── backend/            # Repository (main branch)
    │   └── CLAUDE.md       # Agent context
    └── frontend/           # Repository (main branch)
        └── CLAUDE.md       # Agent context
```

Note: 
- The `rc` command is installed system-wide via Homebrew or manual installation
- Repositories are cloned into the `workspace/` subdirectory by default
- You can customize the workspace path in `repo-claude.yaml` if needed

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
- **Agent**: Claude Code process management

### Inspiration and History

This project was inspired by Google's Android Repo tool, adopting its multi-repository management concepts while simplifying for AI agent orchestration. We initially used Repo directly but found it added unnecessary complexity for our use case.

See our [detailed architecture documentation](docs/architecture.md) and [design decisions](docs/adr/) for more information.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details