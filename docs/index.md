---
title: MUNO Documentation
description: Complete documentation for MUNO - Multi-repository UNified Orchestration
---

# MUNO Documentation

Welcome to the MUNO documentation hub. Find everything you need to manage multiple repositories with monorepo-like convenience.

## 📖 Complete User Guide

### **[MUNO User Guide](GUIDE.md)**
Comprehensive guide for effectively using MUNO for multi-repository management.

This guide covers:
- Repository organization strategies (team-based, service-type, domain-based)
- Migration scenarios from Google Repo, flat structures, and monorepos
- Common workflows and best practices
- Essential commands and configuration examples
- Troubleshooting and validation

**Direct URLs**: 
- Web: `https://taokim.github.io/muno/guide`
- Raw: `https://raw.githubusercontent.com/taokim/muno/main/docs/GUIDE.md`

---

## 📚 Core Documentation

### Getting Started
- [README](https://github.com/taokim/muno#readme) - Overview and quick start guide
- [Installation](homebrew-setup.md) - Installation via Homebrew and other methods
- [Features](features.md) - Complete feature list and capabilities

### Architecture & Design
- [Architecture](architecture.md) - System architecture and design principles
- [Workspace Structure](workspace-structure.md) - Understanding MUNO workspaces
- [Naming](naming.md) - The story behind MUNO's name

### Examples & Templates
- [Example Configurations](https://github.com/taokim/muno/tree/main/examples) - Sample MUNO configurations:
  - [Team-based Organization](https://github.com/taokim/muno/blob/main/examples/team-based.yaml)
  - [Service-type Organization](https://github.com/taokim/muno/blob/main/examples/service-type.yaml)
  - [Domain-driven Organization](https://github.com/taokim/muno/blob/main/examples/domain-driven.yaml)

### Architecture Decision Records
- [ADR Overview](adr/) - Architecture decisions and rationale

---

## 🎯 Quick Links for Specific Use Cases

### For Repository Migration
**If you need to migrate from another tool:**
- From Google Repo → See [User Guide - Migration Scenarios](GUIDE.md#migration-scenarios)
- From flat repository structure → See [Migration from Flat Structure](GUIDE.md#from-flat-repository-structure)
- From monorepo → See [Splitting a Monorepo](GUIDE.md#from-monorepo)

### For Repository Organization
**If you need to organize many repositories:**
- 20-50 repos → [Team-based strategy](GUIDE.md#1-team-based-organization)
- 50-100 repos → [Service-type strategy](GUIDE.md#2-service-type-organization)
- 100+ repos → [Domain-based strategy](GUIDE.md#3-domain-based-organization)

### For Automation Integration
**If you're building automation:**
- Essential commands → [User Guide - Essential Commands](GUIDE.md#essential-commands)
- Validation workflows → [User Guide - Validation](GUIDE.md#validation-checklist)
- Best practices → [User Guide - Best Practices](GUIDE.md#best-practices)

---

## 📖 Documentation Formats

All documentation is available in multiple formats:

### For Humans (Web)
Browse documentation at: [{{ site.url }}{{ site.baseurl }}/]({{ site.url }}{{ site.baseurl }}/)

### For AI Agents (Raw Markdown)
Fetch raw markdown from GitHub:
```
https://raw.githubusercontent.com/{{ site.github.owner_name }}/{{ site.github.repository_name }}/main/docs/[document-name].md
```

### For Programs (JSON API)
Coming soon: JSON-formatted documentation API

---

## 🚀 Getting Help

### In MUNO CLI
```bash
# Show general help
muno --help

# Show command-specific help
muno add --help
muno agent --help

# Show version
muno --version
```

### Community
- [GitHub Issues](https://github.com/taokim/muno/issues) - Report bugs or request features
- [Discussions](https://github.com/taokim/muno/discussions) - Ask questions and share ideas

---

## 📝 Contributing

We welcome contributions! See our [Contributing Guide](https://github.com/taokim/muno/blob/main/CONTRIBUTING.md) for details.

---

## 📄 License

MUNO is released under the [MIT License](https://github.com/taokim/muno/blob/main/LICENSE).

---

*Last updated: {{ site.time | date: '%B %d, %Y' }}*