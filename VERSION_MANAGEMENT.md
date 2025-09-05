# MUNO Version Management

This project supports multiple concurrent installations to avoid conflicts between production, development, and local testing versions.

## Quick Start

```bash
# Check status of all versions
make status

# Install development versions
make install-dev    # Installs as 'muno-dev'
make install-local  # Installs as 'muno-local'

# Update development versions with latest code
./scripts/muno-versions.sh update

# Remove development versions (keeps production)
./scripts/muno-versions.sh clean
```

## Version Strategy

### Three Concurrent Versions

1. **`muno` (Production)**
   - Stable version installed via Homebrew or `go install`
   - Used for production workspaces
   - Not affected by local development

2. **`muno-dev` (Development)**
   - Built from your current branch
   - Used for testing new features
   - Updated with `make install-dev`

3. **`muno-local` (Local Testing)**
   - Built from local changes
   - Used for experiments
   - Updated with `make install-local`

## Usage Examples

```bash
# Use production version for stable work
muno tree
muno pull --recursive

# Test new features with development version
muno-dev init my-workspace
muno-dev add git@github.com:org/repo.git

# Experiment with local version
muno-local tree
muno-local use frontend
```

## Common Workflows

### Testing a New Feature

```bash
# 1. Make your changes
vim internal/tree/manager.go

# 2. Install development version
make install-dev

# 3. Test with development version
muno-dev tree  # Your changes
muno tree      # Still production (unchanged)
```

### Updating Development Versions

```bash
# After making changes
./scripts/muno-versions.sh update

# Or manually
make clean
make install-dev
make install-local
```

### Checking What's Installed

```bash
# Visual status display
make status

# Or use the script directly
./scripts/muno-versions.sh status
```

## Installation Paths

- **Production**: `/opt/homebrew/bin/muno` or `/usr/local/bin/muno`
- **Development**: `$GOPATH/bin/muno-dev`
- **Local**: `$GOPATH/bin/muno-local`

## Makefile Targets

| Target | Description | Binary Name |
|--------|-------------|-------------|
| `make install` | Install production version | `muno` |
| `make install-dev` | Install development version | `muno-dev` |
| `make install-local` | Install local version | `muno-local` |
| `make uninstall-dev` | Remove development version | - |
| `make uninstall-local` | Remove local version | - |
| `make status` | Show all versions status | - |

## Script Commands

The `scripts/muno-versions.sh` script provides additional management:

```bash
./scripts/muno-versions.sh status        # Show detailed status
./scripts/muno-versions.sh install       # Install both dev & local
./scripts/muno-versions.sh install-dev   # Install only dev
./scripts/muno-versions.sh install-local # Install only local  
./scripts/muno-versions.sh update        # Update dev & local with latest
./scripts/muno-versions.sh clean         # Remove dev & local (keep prod)
```

## Troubleshooting

### Command Not Found

If `muno-dev` or `muno-local` are not found after installation:

```bash
# Check if $GOPATH/bin is in PATH
echo $PATH | grep -q "$GOPATH/bin" || echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.bashrc

# Or check the actual installation
ls -la $GOPATH/bin/muno*
```

### Version Shows Wrong Commit

The version includes git information. If it shows "dirty":

```bash
# Check uncommitted changes
git status

# Version will show clean after commit
git add . && git commit -m "Update"
make install-dev
```

### Conflict with System Package Manager

If you have `muno` installed via Homebrew:

```bash
# Production (Homebrew)
brew upgrade muno  # Updates production

# Development (local build)
make install-dev   # Separate binary, no conflict
```