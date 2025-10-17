# MUNO Regression Test Suite

Comprehensive regression testing to ensure MUNO functionality remains stable across changes.

## Overview

The regression test suite validates all MUNO features through automated bash scripts that simulate real-world usage patterns.

## Running Tests

### Quick Test
```bash
# From test directory
./run_regression_tests.sh --quick

# From project root
make -C test test-regression-quick
```

### Full Test Suite
```bash
# From test directory
./run_regression_tests.sh

# With options
./run_regression_tests.sh --verbose --keep-dir
```

### Options
- `--quick`: Run only essential tests (faster)
- `--verbose`: Show detailed output for debugging
- `--keep-dir`: Don't cleanup test directory after completion

## Test Coverage

### 1. Initialization & Configuration
- Workspace initialization
- Configuration file creation
- Re-initialization protection
- State management

### 2. Repository Management
- Adding repositories (eager, lazy, auto-detect)
- Removing repositories
- Listing repositories
- Configuration persistence

### 3. Clone Command Behavior
- `clone` without `--include-lazy` (non-lazy only)
- `clone --include-lazy` (all repositories)
- Clone idempotency
- Recursive cloning

### 4. Pull Command Behavior  
- Pull updates only cloned repositories
- Pull never clones new repositories
- Recursive pull operations
- Force pull with `--force`

### 5. Git Operations
- Status checking
- Commit operations
- Push operations
- Branch management

### 6. Tree Navigation
- Tree structure display
- Clone status indicators (✅, 💤)
- Path resolution
- Summary statistics

### 7. Error Handling
- Invalid commands
- Missing arguments
- Non-existent paths
- Repository conflicts

### 8. Advanced Features
- Nested repository structures
- Recursive operations
- Config file references
- Lazy loading behavior

## Test Environment

Tests create isolated environments in `/tmp/muno-regression-*` with:
- Test workspace directory
- Mock git repositories
- Temporary configuration

## Expected Results

### Success Criteria
- All tests pass (exit code 0)
- No failed tests in summary
- Report generated successfully

### Failure Handling
- Failed tests are listed
- Test directory preserved with `--keep-dir`
- Detailed output with `--verbose`

## Adding New Tests

### Test Case Function
```bash
test_case "Test name" "command to test" "expected_result"
```

Parameters:
- `Test name`: Descriptive name for the test
- `command`: Shell command to execute
- `expected_result`: "pass" (default), "fail", or "skip"

### Example
```bash
test_case "Clone with include-lazy" \
    "$MUNO_BIN clone --include-lazy" \
    "pass"

test_case "Invalid command fails" \
    "$MUNO_BIN invalid-cmd" \
    "fail"
```

## Test Report

After execution, a report is generated at:
```
/tmp/muno-regression-*/regression_report_YYYYMMDD_HHMMSS.txt
```

Contains:
- Test summary (passed/failed/skipped)
- Failed test names
- Test suites executed
- Binary path and timestamp

## Troubleshooting

### Test Failures
1. Run with `--verbose` for detailed output
2. Use `--keep-dir` to inspect test directory
3. Check specific command output in verbose mode
4. Verify MUNO binary is up-to-date

### Common Issues
- **Binary not found**: Build with `make build` first
- **Git errors**: Ensure git is installed
- **Permission issues**: Check directory permissions
- **Cleanup failures**: Manually remove `/tmp/muno-*` directories

## Integration with CI/CD

The regression suite can be integrated into CI pipelines:

```yaml
# GitHub Actions example
- name: Run regression tests
  run: |
    make build
    ./test/run_regression_tests.sh --quick
```

## Maintenance

Keep tests up-to-date when:
- Adding new features
- Changing command behavior
- Fixing bugs
- Modifying output format

Always run regression tests before:
- Merging pull requests
- Creating releases
- Major refactoring