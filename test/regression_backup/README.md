# MUNO Regression Test Suite

Comprehensive regression testing for MUNO (Multi-repository UNified Orchestration) to ensure all features work correctly before releases.

## Quick Start

```bash
# From project root
cd test/regression
./install.sh          # Install to /tmp
cd /tmp/muno-regression-test
./regression_test.sh  # Run tests
```

## What It Tests

### Core Features
- **Tree Navigation**: `tree`, `use`, `current`, `list` commands
- **Repository Management**: `add`, `remove`, `clone` commands  
- **Git Operations**: `pull`, `status` commands
- **Node Types**: Eager loading, lazy loading, config nodes
- **State Persistence**: Configuration and state files
- **Error Handling**: Invalid commands and non-existent repos

### Node Type Coverage

1. **Eager Repositories** (auto-clone based on naming):
   - `-monorepo`, `-platform`, `-workspace`, `-root-repo`

2. **Lazy Repositories** (clone on-demand):
   - Explicit `lazy: true` flag
   - Clone only when navigated to or explicitly cloned

3. **Config Nodes** (contains muno.yaml):
   - References to child repositories
   - Nested tree structures

4. **Normal Repositories**:
   - Standard git repositories
   - No special behavior

## Test Structure

The test creates a complete test environment under `/tmp/muno-regression-test/`:

```
/tmp/muno-regression-test/
├── test-workspace/       # MUNO workspace
│   ├── muno.yaml        # Configuration
│   ├── .muno-state.json # State tracking
│   └── nodes/           # Repository tree
├── test-repos/          # Source repositories
│   ├── backend-monorepo/
│   ├── frontend-platform/
│   ├── service-lazy/
│   └── ...
└── test_results.txt     # Test results
```

## Running Tests

### Full Test Suite
```bash
# Build and test
make build
cd test/regression
./regression_test.sh
```

### From Any Location
```bash
# The test will find the muno binary
/path/to/muno/test/regression/regression_test.sh
```

### Custom Binary Location
```bash
# Edit regression_test.sh to set MUNO_BIN
MUNO_BIN="/custom/path/to/muno" ./regression_test.sh
```

## Test Results

### Success Output
```
════════════════════════════════════════════════════════
         ✅ ALL TESTS PASSED! READY FOR RELEASE
════════════════════════════════════════════════════════
```

### Partial Success
```
════════════════════════════════════════════════════════
         ⚠️  MOSTLY PASSED - REVIEW FAILURES
════════════════════════════════════════════════════════
```

### Failure
```
════════════════════════════════════════════════════════
         ❌ TESTS FAILED - DO NOT RELEASE
════════════════════════════════════════════════════════
```

## Test Cases

### Essential Tests
1. **Initialize workspace** - `muno init`
2. **Display tree** - `muno tree`
3. **Navigate repositories** - `muno use <repo>`
4. **Clone lazy repos** - `muno clone`
5. **Pull updates** - `muno pull`
6. **Check status** - `muno status`
7. **Add/remove repos** - `muno add/remove`

### Edge Cases
- Non-existent repository navigation
- Already cloned repository
- Empty workspace
- Invalid commands
- State persistence across operations

## Continuous Integration

### GitHub Actions
```yaml
name: Regression Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Run Regression Tests
        run: |
          make build
          ./test/regression/regression_test.sh
```

### Pre-Release Checklist
- [ ] Run regression tests
- [ ] Review any failures
- [ ] Fix critical issues
- [ ] Re-run tests until 100% pass
- [ ] Tag release version

## Troubleshooting

### Common Issues

1. **Binary not found**
   ```bash
   Error: Could not find muno binary
   # Solution: Run 'make build' first
   ```

2. **Permission denied**
   ```bash
   # Solution: Make scripts executable
   chmod +x test/regression/*.sh
   ```

3. **Git not configured**
   ```bash
   # Solution: Configure git
   git config --global user.email "test@example.com"
   git config --global user.name "Test User"
   ```

## Adding New Tests

To add a new test case:

1. Edit `regression_test.sh`
2. Add test using `run_test` function:
   ```bash
   run_test "feature: description" "$MUNO_BIN command" "expected_output"
   ```
3. Update this README

## Files

- `regression_test.sh` - Main test runner
- `install.sh` - Installs test suite to /tmp
- `README.md` - This documentation

## Requirements

- Go 1.21+
- Git 2.0+
- Bash 4.0+
- Make

## Version History

- **v1.0** - Initial regression test suite
  - Core command coverage
  - Node type testing  
  - Error handling
  - State persistence