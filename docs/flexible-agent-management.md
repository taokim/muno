# Flexible Agent Management

This document describes the enhanced `rc start` command that provides flexible ways to start and manage Claude Code agents.

## Overview

The improved start command addresses three key developer experience issues:

1. **Process Attachment**: Ability to see agent output and interact with them
2. **Terminal Management**: Simple ways to manage multiple terminal windows without TMUX complexity
3. **Flexible Selection**: Ad-hoc project grouping beyond rigid agent assignments

## Quick Start

```bash
# Start in foreground (see output)
rc start backend --foreground

# Start in new terminal window
rc start frontend --new-window

# Start multiple agents
rc start frontend backend mobile

# Start by repository selection
rc start --repos frontend,backend

# Interactive selection
rc start --interactive

# Use predefined preset
rc start --preset fullstack
```

## Features

### 1. Process Attachment Options

#### Foreground Mode
Run Claude in the current terminal with full visibility:

```bash
rc start backend --foreground
```

- **Pros**: Full visibility of Claude's output, can interact directly
- **Cons**: Blocks the terminal, need multiple terminals for multiple agents
- **Use case**: Debugging, watching agent behavior, single-agent focus

#### New Window Mode
Open each agent in a separate terminal window:

```bash
rc start frontend backend --new-window
```

- **Pros**: Each agent gets its own window, easy to switch between
- **Cons**: Can clutter desktop with many windows
- **Use case**: Multi-agent development with visual separation

#### Background with Logs
Traditional background mode but with output logged to files:

```bash
rc start --all --log-output
```

Logs are saved to `.logs/agentname-timestamp.log`

### 2. Flexible Agent Selection

#### By Repository
Start agents based on which repositories you want to work with:

```bash
# Single repository
rc start --repos frontend

# Multiple repositories
rc start --repos frontend,backend,mobile

# Pattern matching (future feature)
rc start --repos "*/service"
```

#### By Preset
Use predefined combinations from your configuration:

```bash
# Start fullstack preset (frontend + backend)
rc start --preset fullstack

# Start all microservices
rc start --preset microservices
```

#### Interactive Selection
Choose repositories and agents interactively:

```bash
rc start --interactive

# Shows menu:
# Available repositories and agents:
#   1. frontend     → frontend-dev
#   2. backend      → backend-dev
#   3. mobile       → mobile-dev
# 
# Select repositories (comma-separated numbers or names): 1,2
```

#### Multiple Agents Directly
Start specific agents by name:

```bash
rc start frontend-dev backend-dev mobile-dev
```

### 3. Configuration Presets

Define common development scenarios in `repo-claude.yaml`:

```yaml
workspace:
  presets:
    fullstack:
      description: "Full-stack development"
      repositories: [frontend, backend]
      terminal_layout: "split-horizontal"
    
    backend-all:
      description: "All backend services"
      repositories: [backend, auth-service, data-service]
      terminal_layout: "tabs"
    
    migration:
      description: "Legacy system migration"
      repositories: [legacy-api, backend]
      agents_override:
        legacy-api: analyzer
        backend: migrator
```

### 4. Terminal Layouts (Future Enhancement)

Configure how terminals are arranged:

```yaml
terminal_preferences:
  default_layout: "foreground-single"
  layouts:
    split-horizontal: # Side by side
    split-vertical:   # Top and bottom
    tabs:            # Terminal tabs
    separate-windows: # Individual windows
```

## Examples

### Scenario 1: Frontend and Backend Development

```bash
# Option 1: Direct agent names
rc start frontend-dev backend-dev --new-window

# Option 2: By repositories
rc start --repos frontend,backend --foreground

# Option 3: Using preset
rc start --preset fullstack
```

### Scenario 2: Microservices Debugging

```bash
# Start specific services in foreground for debugging
rc start auth-specialist --foreground

# In another terminal, start related services
rc start --repos data-service,api-gateway
```

### Scenario 3: Ad-hoc Exploration

```bash
# Interactive selection for exploring different parts
rc start --interactive

# Quick look at one service
rc start --repos legacy-api --foreground
```

### Scenario 4: Full Team Simulation

```bash
# Start all auto-start agents in background
rc start

# Or start everything in separate windows
rc start --all --new-window
```

## Platform-Specific Behavior

### macOS
- `--new-window` uses Terminal.app or iTerm2
- AppleScript automation for window management

### Linux
- Tries gnome-terminal, konsole, xterm in order
- Falls back to xterm if none found

### Windows
- Uses Windows Terminal or cmd.exe
- Opens new console windows

## Troubleshooting

### "Agent not found"
- Check agent name in configuration
- Use `rc status` to see available agents

### Terminal window doesn't open
- Check if terminal emulator is installed
- Try `--foreground` mode as fallback
- Check system permissions for automation

### Can't see output
- Use `--foreground` for immediate visibility
- Check `.logs/` directory for `--log-output` mode
- Ensure claude command outputs to stdout/stderr

## Future Enhancements

1. **Pattern Matching**
   ```bash
   rc start --repos "*/service" --tag microservice
   ```

2. **Session Management**
   ```bash
   rc session save my-setup
   rc session load my-setup
   ```

3. **Terminal Layouts**
   ```bash
   rc start --preset fullstack --layout split-horizontal
   ```

4. **Resource Groups**
   ```bash
   rc start --group "memory-limit:8GB" --repos frontend,backend
   ```

5. **Agent Communication Scopes**
   ```bash
   rc start --repos frontend,backend --shared-memory-scope limited
   ```