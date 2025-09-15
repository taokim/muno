# MUNO Documentation

Welcome to the MUNO documentation. This directory contains comprehensive guides for understanding and using MUNO's configuration management system.

## 📚 Documentation Structure

### Getting Started
- **[CONFIG_QUICK_START.md](./CONFIG_QUICK_START.md)** - Quick introduction to MUNO configuration with practical examples

### Core Documentation
- **[CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md)** - Complete guide to configuration management, node types, and inheritance
- **[ADVANCED_CONFIG.md](./ADVANCED_CONFIG.md)** - Advanced patterns, overlay mechanisms, and complex configurations

### Reference Documentation
- **[API Reference](./api/)** - API documentation (coming soon)
- **[CLI Reference](./cli/)** - Command-line interface reference (coming soon)

## 🎯 Quick Navigation

### By Use Case

#### I want to...
- **Set up a simple project** → [CONFIG_QUICK_START.md](./CONFIG_QUICK_START.md#basic-configuration)
- **Manage multiple teams** → [CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md#config-reference-nodes)
- **Use environment-specific configs** → [ADVANCED_CONFIG.md](./ADVANCED_CONFIG.md#pattern-1-environment-specific-overlays)
- **Implement lazy loading** → [CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md#node-types)
- **Override parent configurations** → [CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md#overlay-and-override-mechanisms)
- **Use config templates** → [ADVANCED_CONFIG.md](./ADVANCED_CONFIG.md#pattern-4-template-variables)

### By Topic

#### Configuration Basics
- [Configuration File Structure](./CONFIG_MANAGEMENT.md#configuration-file-structure)
- [Node Types](./CONFIG_MANAGEMENT.md#node-types)
- [Basic Examples](./CONFIG_QUICK_START.md#configuration-scenarios)

#### Advanced Configuration
- [Overlay System](./ADVANCED_CONFIG.md#configuration-overlay-system)
- [Override Mechanisms](./ADVANCED_CONFIG.md#override-mechanisms)
- [Complex Patterns](./ADVANCED_CONFIG.md#complex-configuration-patterns)

#### Best Practices
- [Team Autonomy](./CONFIG_MANAGEMENT.md#best-practices)
- [Security Considerations](./CONFIG_MANAGEMENT.md#security-considerations)
- [Performance Optimization](./ADVANCED_CONFIG.md#performance-optimization)

## 📖 Reading Order

### For New Users
1. Start with [CONFIG_QUICK_START.md](./CONFIG_QUICK_START.md)
2. Review basic concepts in [CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md#overview)
3. Try the examples in [CONFIG_QUICK_START.md](./CONFIG_QUICK_START.md#configuration-scenarios)

### For Team Leads
1. Review [CONFIG_MANAGEMENT.md](./CONFIG_MANAGEMENT.md#config-reference-nodes)
2. Understand [Team Autonomy](./CONFIG_MANAGEMENT.md#best-practices)
3. Explore [Multi-Team Patterns](./ADVANCED_CONFIG.md#pattern-2-team-based-configuration)

### For Advanced Users
1. Deep dive into [ADVANCED_CONFIG.md](./ADVANCED_CONFIG.md)
2. Study [Override Mechanisms](./ADVANCED_CONFIG.md#override-mechanisms)
3. Implement [Complex Patterns](./ADVANCED_CONFIG.md#complex-configuration-patterns)

## 🔑 Key Concepts

### Configuration Hierarchy
```
Root Config (muno.yaml)
    ├── Team Config (team/muno.yaml)
    │   └── Service Config (service/muno.yaml)
    └── Shared Config (shared/muno.yaml)
```

### Node Types
- **Git Repository Nodes** - Direct git repository management
- **Config Reference Nodes** - Delegate to external configurations
- **Hybrid Nodes** - Git repos with their own child configurations

### Override Precedence
1. Runtime overrides (CLI flags)
2. Local overrides (.muno-local.yaml)
3. Node configuration
4. Parent configuration
5. Root configuration
6. System defaults

## 💡 Common Use Cases

### Single Team Setup
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
    config: ./teams/a/muno.yaml
  - name: team-b
    config: ./teams/b/muno.yaml
  - name: shared
    url: https://github.com/org/shared.git
```

### Lazy Loading Strategy
```yaml
nodes:
  - name: core
    url: https://github.com/org/core.git
    lazy: false  # Always load
  - name: tools
    url: https://github.com/org/tools.git
    lazy: true   # Load on demand
```

## 🛠 Tools and Commands

### Essential Commands
```bash
muno init <workspace>     # Initialize workspace
muno tree                  # Display repository tree
muno use <path>           # Navigate to node
muno add <url>            # Add repository
muno pull --all           # Update all repos
```

### Configuration Commands
```bash
muno config --show        # Show effective configuration
muno config --validate    # Validate configuration
muno config --trace       # Trace config resolution
```

## 📝 Configuration Examples

Complete examples can be found in each documentation file:
- [Basic Examples](./CONFIG_QUICK_START.md#configuration-scenarios)
- [Advanced Examples](./CONFIG_MANAGEMENT.md#examples)
- [Complex Patterns](./ADVANCED_CONFIG.md#complex-configuration-patterns)

## 🔍 Troubleshooting

### Quick Fixes
- [Common Issues](./CONFIG_MANAGEMENT.md#troubleshooting)
- [Advanced Debugging](./ADVANCED_CONFIG.md#troubleshooting-advanced-configurations)
- [Configuration Tips](./CONFIG_QUICK_START.md#configuration-tips)

### Getting Help
1. Check the troubleshooting sections in each guide
2. Run `muno help <command>` for command-specific help
3. Review the [FAQ](./FAQ.md) (coming soon)
4. Submit an issue on [GitHub](https://github.com/taokim/muno)

## 📚 Additional Resources

### Related Documentation
- [Main README](../README.md) - Project overview and installation
- [CLAUDE.md](../CLAUDE.md) - AI assistant integration guide
- [Examples](../examples/) - Sample configurations

### External Links
- [GitHub Repository](https://github.com/taokim/muno)
- [Release Notes](https://github.com/taokim/muno/releases)
- [Issue Tracker](https://github.com/taokim/muno/issues)

## 🚀 Contributing

We welcome contributions to improve this documentation:
1. Fork the repository
2. Create a feature branch
3. Make your improvements
4. Submit a pull request

### Documentation Standards
- Use clear, concise language
- Include practical examples
- Maintain consistent formatting
- Test all code examples
- Update the table of contents

## 📄 License

This documentation is part of the MUNO project and follows the same license terms. See [LICENSE](../LICENSE) for details.