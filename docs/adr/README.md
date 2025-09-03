# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the MUNO project.

## What is an ADR?

An Architecture Decision Record captures an important architectural decision made along with its context and consequences.

## ADR Index

- [ADR-001](001-simplify-git-management.md) - Simplify Git Management by Removing Android Repo Tool Dependency (2024-08-15)
- [ADR-002](002-tree-based-architecture.md) - Tree-Based Architecture (2024-12-27)
- [ADR-003](003-manager-interface-abstraction.md) - Manager Interface Abstraction for Testability and Plugin Architecture (2024-12)

## ADR Template

When creating a new ADR, use this template:

```markdown
# ADR-XXX: Title

Date: YYYY-MM-DD

## Status

[Proposed | Accepted | Deprecated | Superseded by ADR-XXX]

## Context

What is the issue that we're seeing that is motivating this decision or change?

## Decision

What is the change that we're proposing and/or doing?

## Consequences

What becomes easier or more difficult to do because of this change?
```

## References

- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) by Michael Nygard
- [Architecture Decision Records](https://adr.github.io/)