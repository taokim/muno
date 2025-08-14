# Repo-Claude

Multi-agent orchestration using Repo tool and Claude Code for trunk-based development across multiple repositories.

## Features

- 🚀 **Single Binary**: No runtime dependencies (Python, Node.js, etc.)
- 🔧 **Real Repo Tool Integration**: Uses the actual Repo tool for multi-repository management
- 🤖 **Multi-Agent Orchestration**: Coordinate multiple Claude Code instances across repositories
- 🌳 **Trunk-Based Development**: All agents work directly on main branch
- 📝 **Shared Memory**: Cross-agent coordination through shared memory file
- ⚡ **Fast**: Written in Go for optimal performance

## Prerequisites

- Git
- [Repo tool](https://gerrit.googlesource.com/git-repo/)
- [Claude Code CLI](https://claude.ai/code)

### Installing Repo Tool

```bash
# macOS/Linux
curl https://storage.googleapis.com/git-repo-downloads/repo > ~/bin/repo
chmod a+x ~/bin/repo
export PATH=$PATH:~/bin
```

## Installation

### From Release

```bash
# Download latest release
curl -L https://github.com/taokim/repo-claude/releases/latest/download/repo-claude_$(uname -s)_$(uname -m).tar.gz | tar xz
chmod +x repo-claude
sudo mv repo-claude /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/taokim/repo-claude.git
cd repo-claude
make build
sudo make install
```

### Homebrew (coming soon)

```bash
brew install repo-claude
```

## Quick Start

1. **Initialize a new workspace**:
   ```bash
   repo-claude init my-project
   cd my-project
   ```

2. **Start agents**:
   ```bash
   ./repo-claude start              # Start all auto-start agents
   ./repo-claude start backend-agent # Start specific agent
   ```

3. **Check status**:
   ```bash
   ./repo-claude status
   repo status  # Use repo tool directly
   ```

4. **Stop agents**:
   ```bash
   ./repo-claude stop
   ```

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

```
my-project/
├── repo-claude          # Executable
├── repo-claude.yaml     # Configuration
├── shared-memory.md     # Agent coordination
├── .manifest-repo/      # Local git repo with manifest
│   └── default.xml     # Repo manifest
├── .repo/              # Repo metadata
├── backend/            # Repository (main branch)
│   └── CLAUDE.md      # Agent context
└── frontend/          # Repository (main branch)
    └── CLAUDE.md      # Agent context
```

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

Repo-Claude leverages the Repo tool (used for Android and ChromiumOS) to manage multiple Git repositories. Each agent works in its own repository with awareness of other repositories through relative paths and shared memory.

Key components:
- **Manager**: Core orchestration logic
- **Config**: YAML configuration and state management
- **Manifest**: Repo manifest generation
- **Agent**: Claude Code process management

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details