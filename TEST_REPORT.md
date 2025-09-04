# Test Execution Report - MUNO Project

Generated: 2025-09-04

## Test Summary

### Overall Results
- **Status**: ❌ FAILED
- **Total Coverage**: 48.8% of statements
- **Failing Packages**: 3 out of 8

### Package Coverage

| Package | Status | Coverage | Notes |
|---------|--------|----------|-------|
| `cmd/muno` | ✅ PASS | 79.9% | Command-line interface |
| `internal/adapters` | ✅ PASS | 84.2% | Adapter interfaces |
| `internal/config` | ✅ PASS | 87.3% | Configuration management |
| `internal/git` | ✅ PASS | 84.1% | Git operations |
| `internal/manager` | ❌ FAIL | 73.6% | Core manager logic |
| `internal/tree` | ❌ FAIL | 80.3% | Tree navigation system |
| `test` | ❌ FAIL | - | Integration tests |
| `test/e2e` | ✅ PASS | - | End-to-end tests |

## Test Failures

### 1. Manager Package (`internal/manager`)
**Failed Test**: `TestManager_StartAgent`
- **Issue**: Path construction mismatch
- **Expected**: `/workspace/test/repo`
- **Actual**: `/workspace/repos/test/repos/repo`
- **Root Cause**: The test expects a different directory structure than what's being generated

### 2. Tree Package (`internal/tree`)
**Failed Tests**:
- `TestManager_EdgeCases/CloneLazyRepos_non-recursive`
  - Clone count mismatch (expected 1, got 0)
- `TestStatelessManager_EdgeCases`
  - Missing config file error

### 3. Integration Tests (`test`)
**Failed Tests**:
- `TestIntegrationWorkflow/List` - Output format mismatch
- `TestIntegrationWorkflow/Status` - Output format mismatch
- `TestConfigValidation/Command_start_wms` - Command not found (expected config error)

## Coverage Analysis

### High Coverage (>80%)
- `internal/config`: 87.3% - Well-tested configuration logic
- `internal/adapters`: 84.2% - Good interface coverage
- `internal/git`: 84.1% - Solid git operation coverage
- `internal/tree`: 80.3% - Despite failures, good code coverage

### Moderate Coverage (70-80%)
- `cmd/muno`: 79.9% - Command implementation well covered
- `internal/manager`: 73.6% - Core logic needs more coverage

### Low Coverage Areas
- `internal/tree/resolver.go:buildNode`: 27.8% - Critical gap
- `internal/tree/resolver.go:LoadNodeConfig`: 60.0% - Needs improvement

## Recommendations

### Immediate Actions
1. **Fix Path Construction**: Update manager tests to match actual directory structure
2. **Fix Clone Logic**: Investigate why lazy repos aren't cloning in tests
3. **Update Output Assertions**: Integration tests need updated expected outputs

### Coverage Improvements
1. Add tests for `buildNode` function in resolver
2. Increase coverage for config loading edge cases
3. Add more error path testing

### Quality Gates
- **Current**: 48.8% overall coverage
- **Target**: 70-80% per package (as specified in CLAUDE.md)
- **Gap**: Need to increase coverage by ~21.2%

## Test Execution Commands

```bash
# Run all tests with coverage
go test ./... -coverprofile=coverage.out -covermode=atomic

# Run specific package tests
go test ./internal/manager -v
go test ./internal/tree -v
go test ./test -v

# Generate coverage reports
go tool cover -func=coverage.out  # Terminal summary
go tool cover -html=coverage.out -o coverage.html  # HTML report

# Run with race detection
go test -race ./...
```

## Next Steps
1. Fix failing tests in manager, tree, and integration packages
2. Increase test coverage to meet 70-80% target
3. Add missing test cases for low-coverage functions
4. Consider adding more e2e tests for critical workflows