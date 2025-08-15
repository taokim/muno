# Repo-Claude Architecture

## Overview

Repo-Claude is a multi-agent orchestration tool for Claude Code that manages multiple Git repositories with trunk-based development. It enables multiple AI agents to work together across different repositories while maintaining coordination through shared memory.

## Inspiration from Android Repo Tool

This project was heavily inspired by Google's [Android Repo tool](https://gerrit.googlesource.com/git-repo/), which has been successfully managing thousands of Git repositories for the Android Open Source Project since 2008. We adopted several key concepts from Repo:

### Concepts Borrowed from Repo

1. **Multi-Repository Management**: Like Repo, we manage multiple Git repositories as a cohesive workspace
2. **Manifest-Based Configuration**: Our `repo-claude.yaml` serves a similar purpose to Repo's manifest XML
3. **Unified Commands**: Commands like `sync`, `status`, and `forall` are directly inspired by Repo
4. **Workspace Structure**: The idea of a workspace root containing multiple project repositories
5. **Parallel Operations**: Repo's `-j` flag for parallel operations inspired our concurrent Git operations

### Key Differences

While inspired by Repo, we made several deliberate design decisions to simplify the tool for AI agent orchestration:

1. **Single Configuration File**: Instead of XML manifests in a separate Git repository, we use a single `repo-claude.yaml`
2. **Direct Git Operations**: We use native Git commands rather than Repo's Python-based abstraction layer
3. **No Branch Management**: We focus on trunk-based development (main branch only)
4. **Agent Integration**: Added AI agent management capabilities not present in Repo
5. **Shared Memory**: Introduced a coordination mechanism specific to multi-agent collaboration

## Architecture Components

### 1. Configuration (`repo-claude.yaml`)

The configuration file defines:
- **Workspace metadata**: Name and settings
- **Repository definitions**: URL, branch, groups, and agent assignments
- **Agent specifications**: Model, specialization, and dependencies

```yaml
workspace:
  name: my-project
  manifest:
    remote_fetch: https://github.com/myorg/
    default_revision: main
    projects:
      - name: backend
        groups: core,services
        agent: backend-agent

agents:
  backend-agent:
    model: claude-sonnet-4
    specialization: API development
    auto_start: true
```

### 2. Git Manager (`internal/git`)

Handles all Git operations:
- **Clone**: Parallel cloning of repositories
- **Sync**: Pulling latest changes from remotes
- **Status**: Checking Git status across all repositories
- **ForAll**: Running commands in each repository

### 3. Agent Manager (`internal/manager`)

Manages Claude Code instances:
- **Start/Stop**: Lifecycle management of AI agents
- **State Tracking**: Persistent state of running agents
- **Dependency Resolution**: Starting agents in the correct order
- **Process Management**: Handling of Claude Code subprocesses

### 4. Coordination

Multi-agent coordination through:
- **Shared Memory** (`shared-memory.md`): Central file for agent communication
- **CLAUDE.md Files**: Per-repository context for each agent
- **Cross-Repository Awareness**: Agents know about other repositories via relative paths

## Command Flow

### Initialize Workspace
```
repo-claude init → Create/Load Config → Clone Repositories → Setup Coordination Files
```

### Start Agents
```
repo-claude start → Load Config → Resolve Dependencies → Start Claude Code → Track State
```

### Sync Repositories
```
repo-claude sync → Load Config → Parallel Git Pull → Update Status
```

## Design Principles

1. **Simplicity First**: Avoid unnecessary complexity
2. **Git Native**: Use Git directly rather than abstractions
3. **Parallel by Default**: Maximize performance through concurrency
4. **Fail Gracefully**: Continue operations even if some repositories fail
5. **Trunk-Based**: All work happens on the main branch

## Why Not Just Use Android Repo?

We initially used the Android Repo tool directly but found it added unnecessary complexity for AI agent orchestration:

1. **Manifest Management**: Required understanding of XML manifests and manifest repositories
2. **Python Runtime**: Downloaded large Python runtime for each workspace
3. **Feature Mismatch**: Many Repo features (upload, cherry-pick, etc.) weren't relevant
4. **User Confusion**: The relationship between our config and Repo's manifest was unclear

See [ADR-001](adr/001-simplify-git-management.md) for the detailed decision record.

## Future Directions

1. **Plugin System**: Allow custom commands and agent types
2. **Cloud Sync**: Optional cloud-based shared memory
3. **Web UI**: Visual monitoring of agent activity
4. **Enhanced Coordination**: More sophisticated inter-agent communication
5. **Template Library**: Pre-built configurations for common architectures