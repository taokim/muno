# Repo-Claude v3

Transform your multi-repository development into a unified tree-based workspace with AI-powered navigation.

## The Evolution

Repo-Claude v3 introduces a revolutionary **tree-based architecture** that treats your entire codebase as a navigable filesystem, eliminating the complexity of scope management while providing intuitive, CWD-first operations.

## Key Innovation: Tree-Based Navigation

### 🌳 **Workspace as a Tree**
Your repositories form a natural tree structure:
```
workspace/
├── team-backend/           # Also a git repo
│   ├── payment-service/    # Child repo
│   ├── order-service/      # Child repo
│   └── shared-libs/        # Lazy-loaded repo
└── team-frontend/          # Also a git repo
    ├── web-app/            # Child repo
    └── component-lib/      # Lazy-loaded repo
```

### 📍 **CWD-First Resolution**
Commands operate based on your current location:
```bash
cd workspaces/team-backend
rc pull                    # Pulls backend repos (CWD-based)
rc add https://...         # Adds repo to backend team
rc tree                    # Shows tree from current position
```

### 💤 **Smart Lazy Loading**
Repositories clone on-demand:
```bash
rc use team-backend        # Auto-clones lazy repos
rc use --no-clone frontend # Navigate without cloning
rc clone --recursive       # Manual clone when needed
```

## Core Features

- 🌳 **Tree Navigation**: Navigate your workspace like a filesystem
- 📍 **CWD-First**: Current directory determines operation target
- 🎯 **Clear Targeting**: Every command shows what it affects
- 💤 **Lazy Loading**: Repos clone only when needed
- 🚀 **Single Binary**: No runtime dependencies
- ⚡ **Fast**: Written in Go for optimal performance

## Installation

### From Source

```bash
git clone https://github.com/taokim/repo-claude.git
cd repo-claude/repo-claude-go
make build
sudo make install
```

## Quick Start

### 1. Initialize Workspace

```bash
rc init my-platform
cd my-platform
```

### 2. Build Your Tree

```bash
# Add team repositories (these become parent nodes)
rc add https://github.com/org/backend-team --name team-backend
rc add https://github.com/org/frontend-team --name team-frontend

# Navigate and add child repositories
rc use team-backend
rc add https://github.com/org/payment-service
rc add https://github.com/org/order-service
rc add https://github.com/org/shared-libs --lazy  # Won't clone until needed

# Navigate to frontend
rc use ../team-frontend
rc add https://github.com/org/web-app
rc add https://github.com/org/component-lib --lazy
```

### 3. Work with the Tree

```bash
# View structure
rc tree                    # Full tree from current position
rc list                    # List immediate children
rc status --recursive      # Status of entire subtree

# Navigate (changes CWD)
rc use /                   # Go to root
rc use team-backend        # Navigate to backend (auto-clones lazy repos)
rc use payment-service     # Go deeper
rc use ..                  # Go up one level
rc use -                   # Previous position

# Git operations (CWD-based)
rc pull                    # Pull at current node
rc pull --recursive        # Pull entire subtree
rc commit -m "Update"      # Commit at current node
rc push --recursive        # Push entire subtree
```

### 4. Start Claude Session

```bash
rc use team-backend/payment-service
rc start                   # Claude session at payment-service

# Or start at specific location
rc start team-frontend     # Start session at frontend
```

## Command Reference

### Navigation Commands
- `rc use <path>` - Navigate to node (changes CWD)
- `rc current` - Show current position
- `rc tree [--depth N]` - Display tree structure
- `rc list [--recursive]` - List child nodes

### Repository Management
- `rc add <url> [--name X] [--lazy]` - Add child repository
- `rc remove <name>` - Remove child repository
- `rc clone [--recursive]` - Clone lazy repositories

### Git Operations
All git commands operate relative to current position:
- `rc pull [path] [--recursive]` - Pull repositories
- `rc push [path] [--recursive]` - Push changes
- `rc commit -m "msg" [--recursive]` - Commit changes
- `rc status [--recursive]` - Show git status

### Session Management
- `rc start [path]` - Start Claude session
- `rc init <name>` - Initialize new workspace

## Target Resolution

Every command clearly shows its target:

```bash
$ rc pull
🎯 Target: team/backend/payment (from CWD)
Pulling 3 repositories...

$ rc pull team/frontend
🎯 Target: team/frontend (explicit)
Pulling 2 repositories...

$ cd /tmp
$ rc pull
🎯 Target: team/backend (from stored current)
⚠️  Not in workspace, using last position
```

## Resolution Priority

1. **Explicit path** - If you specify a path
2. **CWD mapping** - Your current directory location
3. **Stored current** - Last `rc use` position (when outside workspace)
4. **Root fallback** - Default to workspace root

## Key Differences from v2

| v2 (Scope-Based) | v3 (Tree-Based) |
|------------------|-----------------|
| `rc scope create backend` | `rc add https://...backend` |
| `rc start backend` | `rc use backend && rc start` |
| `rc pull --scope backend` | `cd backend && rc pull` |
| Complex scope management | Natural tree navigation |
| Hidden state confusion | CWD-first clarity |

## Philosophy

- **No hidden state**: Your location determines behavior
- **Natural navigation**: Works like a filesystem
- **Clear feedback**: Always shows what will be affected
- **Lazy by default**: Clone only what you need
- **Simple commands**: One `add` command for everything

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.