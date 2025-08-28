# Testing Strategy for Repo-Claude V3

## Overview

The V3 tree-based architecture requires comprehensive testing to ensure it can handle real-world scenarios with complex repository structures. This document outlines our testing approach using local git repositories to simulate production environments.

## Testing Mechanisms

### 1. **Test Tree Generator** (`scripts/generate-test-tree.sh`)

Creates a realistic multi-depth tree structure with actual git repositories.

**Features:**
- Generates 5+ depth tree structures
- Creates real git repositories with commit history
- Supports multiple repository types (frontend, backend, service, library, etc.)
- Configurable depth and breadth
- Creates 50-100+ repositories for stress testing

**Usage:**
```bash
# Generate default test tree (5 depth, 3 repos per level)
./scripts/generate-test-tree.sh

# Custom configuration
./scripts/generate-test-tree.sh /tmp/my-test 7 5
# Creates 7-level deep tree with 5 repos per level

# The script creates:
# - /tmp/repo-claude-test/repo-pool/    # Pool of git repositories
# - /tmp/repo-claude-test/workspace/     # RC workspace for testing
# - /tmp/repo-claude-test/test-repo-claude.sh  # Automated test script
```

### 2. **Advanced Testing Framework** (`scripts/advanced-test-framework.sh`)

Simulates real-world architectural patterns and scenarios.

**Scenarios:**

#### a. Microservices Architecture
- Multiple service repositories
- Shared libraries
- Infrastructure configurations
- Cross-repository dependencies

```bash
./scripts/advanced-test-framework.sh microservices
```

#### b. Monorepo Structure
- Large monolithic repository
- Workspace-based structure
- Shared packages
- Tool repositories

```bash
./scripts/advanced-test-framework.sh monorepo
```

#### c. ML Pipeline
- Data repositories
- Model repositories
- Experiment tracking
- Serving infrastructure

```bash
./scripts/advanced-test-framework.sh ml_pipeline
```

#### d. Enterprise Multi-team
- Team-based repository structure
- Cross-team dependencies
- Shared resources
- Security and compliance repos

```bash
./scripts/advanced-test-framework.sh enterprise
```

### 3. **Local Repository Testing**

The key insight is using `file://` URLs to reference local git repositories:

```yaml
# repo-claude.yaml configuration for local testing
version: 3
workspace:
  name: local-test
  
repositories:
  - url: file:///tmp/test-repos/frontend
    name: frontend
    lazy: false
    
  - url: file:///tmp/test-repos/backend
    name: backend
    lazy: true
    
  - url: file:///tmp/test-repos/shared-lib
    name: shared
    lazy: false
```

This allows:
- **No network dependencies** - All repos are local
- **Fast cloning** - Local file operations
- **Full git functionality** - Real repositories with history
- **Easy modification** - Can modify repos during testing
- **Reproducible tests** - Consistent test environment

## Testing Workflow

### Step 1: Generate Test Environment

```bash
# Build the binary
make build

# Generate test tree
./scripts/generate-test-tree.sh /tmp/rc-test 6 4
# Creates 6-level tree with 4 repos per level
```

### Step 2: Initialize Workspace

```bash
cd /tmp/rc-test/workspace
rc init test-project
```

### Step 3: Add Repositories from Pool

```bash
# Add repositories from the generated pool
rc add file:///tmp/rc-test/repo-pool/frontend-root-L1-1 --name frontend
rc add file:///tmp/rc-test/repo-pool/backend-root-L1-2 --name backend
rc add file:///tmp/rc-test/repo-pool/service-root-L1-3 --name service --lazy

# Add deeper repositories
rc use frontend
rc add file:///tmp/rc-test/repo-pool/frontend-root-L2-1 --name components
rc add file:///tmp/rc-test/repo-pool/frontend-root-L2-2 --name styles --lazy
```

### Step 4: Test Navigation

```bash
# Test tree navigation
rc tree           # Show full tree
rc list          # List current level
rc use /         # Go to root
rc use frontend  # Navigate to frontend
rc use components # Go deeper
rc use ../..     # Navigate up
rc use -         # Go to previous
```

### Step 5: Test Lazy Loading

```bash
# Navigate to lazy repo (should auto-clone if configured)
rc use frontend/styles  # Auto-clones if lazy
rc clone --recursive    # Clone all lazy repos
```

### Step 6: Test Git Operations

```bash
# Status across tree
rc status

# Pull all repos
rc pull /

# Commit changes
rc commit / -m "Test commit"

# Push changes (will fail for local repos, but tests the flow)
rc push /
```

## Performance Testing

### Large-Scale Testing

```bash
# Generate large tree (7 levels, 5 repos per level = ~16,000 repos theoretical max)
./scripts/generate-test-tree.sh /tmp/large-test 7 5

# Add many repos
for i in {1..100}; do
  rc add "file:///tmp/large-test/repo-pool/repo-$i" --name "repo-$i" --lazy
done

# Performance metrics
time rc tree          # Tree display performance
time rc status        # Status check performance
time rc use /        # Navigation performance
```

### Memory Testing

```bash
# Monitor memory usage
/usr/bin/time -v rc tree

# Check with many repos loaded
rc clone --recursive  # Load all repos
/usr/bin/time -v rc status
```

## Integration Testing

### CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: make build
      
      - name: Generate Test Environment
        run: ./scripts/generate-test-tree.sh /tmp/test 5 3
      
      - name: Run Integration Tests
        run: |
          cd /tmp/test/workspace
          ./test-repo-claude.sh
      
      - name: Run Performance Tests
        run: |
          cd /tmp/test/workspace
          ./performance-test.sh
```

### Docker-based Testing

```dockerfile
# Dockerfile.test
FROM golang:1.21-alpine

RUN apk add --no-cache git bash

WORKDIR /app
COPY . .

RUN go build -o rc ./cmd/repo-claude

# Generate test environment
RUN ./scripts/generate-test-tree.sh /test 5 3

# Run tests
CMD ["/test/test-repo-claude.sh"]
```

## Test Scenarios

### 1. Deep Navigation Test
- Create 7+ level deep tree
- Navigate to deepest level
- Test relative navigation
- Test absolute paths

### 2. Lazy Loading Test
- Create tree with 50% lazy repos
- Navigate through tree
- Verify lazy repos clone on access
- Test recursive clone

### 3. Cross-Reference Test
- Create repos with dependencies
- Test cross-repo operations
- Verify shared library handling

### 4. Scale Test
- Add 100+ repositories
- Test performance degradation
- Measure memory usage
- Check UI responsiveness

### 5. Concurrent Operations Test
- Multiple git operations in parallel
- Test file locking
- Verify data consistency

## Validation Checklist

- [ ] Tree displays correctly at all depths
- [ ] Navigation works with relative and absolute paths
- [ ] Lazy repositories clone when accessed
- [ ] Git operations work across the tree
- [ ] Performance acceptable with 100+ repos
- [ ] State persists between sessions
- [ ] CWD resolution works correctly
- [ ] Previous navigation (`-`) works
- [ ] Recursive operations handle depth correctly
- [ ] Error handling for missing/invalid repos

## Benefits of This Approach

1. **Realistic Testing**: Uses actual git repositories, not mocks
2. **Scalable**: Can generate trees of any size
3. **Reproducible**: Scripts create consistent test environments
4. **Flexible**: Multiple scenarios for different use cases
5. **Local**: No external dependencies or network requirements
6. **Comprehensive**: Tests all aspects of the tree structure

## Future Enhancements

1. **Automated Test Suite**: Convert manual tests to automated Go tests
2. **Benchmark Suite**: Add performance benchmarks
3. **Stress Testing**: Push limits with 1000+ repos
4. **Integration Tests**: Test with real remote repositories
5. **UI Testing**: Add tests for any future UI components

## Conclusion

This testing strategy ensures repo-claude V3 can handle real-world scenarios with complex repository structures. The combination of generated test trees and scenario-based testing provides comprehensive coverage of functionality and performance characteristics.