---
layout: default
title: MUNO Documentation
description: Complete documentation for MUNO - Multi-repository UNified Orchestration
---

# MUNO Documentation

Welcome to the MUNO documentation hub. Find everything you need to manage multiple repositories with monorepo-like convenience.

## 🤖 For AI Agents

### **⚠️ If you are an AI Agent, READ THIS FIRST** 
**→ [AI Agent Guide](AI_AGENT_GUIDE.md)** - Essential guide for Claude, Gemini, Cursor, and other AI agents to effectively organize repositories using MUNO

This comprehensive guide teaches AI agents how to:
- Analyze existing repository structures
- Recommend optimal organization strategies
- Generate migration scripts
- Execute repository reorganization
- Validate successful migrations

**Direct URL for agents to fetch**: 
- HTML: `{{ site.url }}{{ site.baseurl }}/AI_AGENT_GUIDE`
- Raw Markdown: `https://raw.githubusercontent.com/{{ site.github.owner_name }}/{{ site.github.repository_name }}/main/docs/AI_AGENT_GUIDE.md`

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
- From Google Repo → See [AI Agent Guide - Migration Scenarios](AI_AGENT_GUIDE.md#migration-scenarios)
- From flat repository structure → See [Scenario 4](AI_AGENT_GUIDE.md#scenario-4-migrating-from-flat-structure)
- From monorepo → See [Scenario 3](AI_AGENT_GUIDE.md#scenario-3-splitting-a-monorepo)

### For Repository Organization
**If you need to organize many repositories:**
- 20-50 repos → [Team-based strategy](AI_AGENT_GUIDE.md#1-team-based-organization)
- 50-100 repos → [Service-type strategy](AI_AGENT_GUIDE.md#2-service-type-organization)
- 100+ repos → [Domain-based strategy](AI_AGENT_GUIDE.md#3-domain-based-organization)

### For AI/Automation Integration
**If you're building automation:**
- Command patterns → [AI Agent Guide - Command Patterns](AI_AGENT_GUIDE.md#command-patterns)
- Validation workflows → [AI Agent Guide - Validation](AI_AGENT_GUIDE.md#validation-and-verification)
- Best practices → [AI Agent Guide - Best Practices](AI_AGENT_GUIDE.md#best-practices-for-ai-agents)

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