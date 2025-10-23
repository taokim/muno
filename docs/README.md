# MUNO Documentation

Welcome to the MUNO documentation. This directory contains comprehensive guides for understanding and using MUNO.

## Quick Start

### For New Users
- **[GUIDE.md](./GUIDE.md)** - Complete user guide for using MUNO effectively
- **[CONFIG_QUICK_START.md](./CONFIG_QUICK_START.md)** - Quick introduction with practical examples

### Core Documentation
- **[CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md)** - Configuration management and node types
- **[ADVANCED_CONFIG.md](./ADVANCED_CONFIG.md)** - Advanced patterns and complex configurations

### Architecture & Design
- **[architecture.md](./architecture.md)** - System architecture and design principles
- **[workspace-structure.md](./workspace-structure.md)** - Understanding MUNO workspaces
- **[features.md](./features.md)** - Complete feature list

## Key Concepts

### Tree-Based Organization
```
workspace/
├── .nodes/
│   ├── team-backend/
│   │   └── .nodes/
│   │       ├── payment-service/
│   │       └── order-service/
│   └── team-frontend/
│       └── .nodes/
│           ├── web-app/
│           └── mobile-app/
```

### Essential Commands
```bash
muno init <workspace>     # Initialize workspace
muno tree                 # Display repository tree
muno add <url>            # Add repository
muno clone --recursive    # Clone lazy repositories
muno pull --recursive     # Update all repos
muno status               # Show status
```

### Node Types
- **Git Repository Nodes** - Direct git repository management
- **Config Reference Nodes** - Delegate to external configurations
- **Hybrid Nodes** - Git repos with their own child configurations

## Common Use Cases

### Single Team
```yaml
workspace:
  name: my-team
nodes:
  - name: service-a
    url: https://github.com/org/service-a.git
  - name: service-b
    url: https://github.com/org/service-b.git
```

### Multi-Team Platform
```yaml
workspace:
  name: platform
nodes:
  - name: team-a
    file: ./teams/a/muno.yaml
  - name: team-b
    file: ./teams/b/muno.yaml
  - name: shared
    url: https://github.com/org/shared.git
```

### Lazy Loading
```yaml
nodes:
  - name: core
    url: https://github.com/org/core.git
    lazy: false  # Always load
  - name: tools
    url: https://github.com/org/tools.git
    lazy: true   # Load on demand
```

## Getting Help

1. Run `muno help <command>` for command-specific help
2. Check the troubleshooting sections in each guide
3. Submit an issue on [GitHub](https://github.com/taokim/muno)

## Additional Resources

- [Main README](../README.md) - Project overview and installation
- [CLAUDE.md](../CLAUDE.md) - AI assistant integration guide
- [Examples](../examples/) - Sample configurations
- [GitHub Repository](https://github.com/taokim/muno)