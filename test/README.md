# MUNO Test Suite

Comprehensive testing for the MUNO multi-repository orchestration tool.

## Test Structure

```
test/
├── README.md                      # This file
├── run_regression_tests.sh        # Main regression test suite
├── run_e2e_tests.sh              # End-to-end test suite
├── Makefile                      # Test automation
│
├── unit/                         # Unit tests (Go)
│   └── (embedded in source)     # Located in internal/*/
│
├── integration/                  # Integration tests (Go)
│   ├── integration_test.go      # Basic integration tests
│   ├── clone_pull_integration_test.go  # Clone/Pull behavior tests
│   └── renamed_node_test.go     # Node renaming tests
│
├── e2e/                         # End-to-end tests (Go)
│   └── workflow_test.go        # Workflow scenarios
│
└── regression/                  # Regression tests (Bash)
    └── README.md               # Regression test documentation
```

## Quick Start

### Run All Tests
```bash
make test-all
```

### Run Specific Test Suites

#### Unit Tests
```bash
make test-unit
# or
go test ./...
```

#### Integration Tests
```bash
make test-integration
# or
go test ./test -run TestIntegration
```

#### End-to-End Tests
```bash
make test-e2e
# or
./test/run_e2e_tests.sh
```

#### Regression Tests
```bash
make test-regression
# or
./test/run_regression_tests.sh
```

## Test Suites

### 1. Unit Tests (`go test`)
- **Location**: Embedded in source code (`internal/*/`)
- **Purpose**: Test individual functions and components
- **Coverage Target**: >70% per package
- **Run**: `go test ./... -short`

### 2. Integration Tests (`integration_test.go`)
- **Purpose**: Test component interactions
- **Key Tests**:
  - Configuration validation
  - Repository management
  - Clone/Pull behavior separation
  - Command integration
- **Run**: `go test ./test -run TestIntegration`

### 3. End-to-End Tests (`run_e2e_tests.sh`)
- **Purpose**: Test complete user workflows
- **Test Categories**:
  - CLI interface tests
  - Real-world scenarios
  - Performance tests
  - Developer workflows
- **Options**:
  - `--verbose`: Show detailed output
  - `--parallel`: Run tests in parallel

### 4. Regression Tests (`run_regression_tests.sh`)
- **Purpose**: Ensure no functionality breaks
- **Test Categories**:
  - Initialization & Configuration
  - Repository Management
  - Clone Command Behavior
  - Pull Command Behavior
  - Git Operations
  - Tree Navigation
  - Error Handling
  - Advanced Features
- **Options**:
  - `--quick`: Run only essential tests
  - `--verbose`: Show detailed output
  - `--keep-dir`: Keep test directory for debugging

## Testing Guidelines

### Writing Tests

#### Go Tests
```go
func TestFeatureName(t *testing.T) {
    t.Run("Scenario", func(t *testing.T) {
        // Arrange
        // Act
        // Assert
    })
}
```

#### Bash Tests
```bash
test_case "Test name" "command to run" "expected_result"
```

### Test Coverage

Monitor test coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Continuous Integration

Tests are automatically run on:
- Pull requests
- Commits to main branch
- Release builds

## Test Commands Reference

### Makefile Targets
```bash
make test           # Run all Go tests
make test-unit      # Run unit tests only
make test-integration # Run integration tests
make test-e2e       # Run E2E tests
make test-regression # Run regression tests
make test-all       # Run everything
make test-coverage  # Generate coverage report
```

### Direct Commands
```bash
# Unit tests with coverage
go test -cover ./...

# Integration tests with verbose output
go test -v ./test -run TestIntegration

# E2E tests with verbose mode
./test/run_e2e_tests.sh --verbose

# Regression tests in quick mode
./test/run_regression_tests.sh --quick

# Specific package tests
go test ./internal/manager -run TestCloneRepos
```

## Key Test Scenarios

### Clone/Pull Separation
- `muno clone`: Clones non-lazy repositories by default
- `muno clone --include-lazy`: Clones all repositories
- `muno pull`: Only updates already cloned repositories

### Repository Management
- Adding repositories with different fetch modes
- Removing repositories safely
- Listing and tree display

### Navigation & Tree Operations
- Tree structure display
- Path resolution
- Status checking

## Troubleshooting

### Test Failures
1. Check MUNO binary is built: `make build`
2. Run with verbose mode: `--verbose`
3. Keep test directory: `--keep-dir`
4. Check specific test: `go test -v -run TestName`

### Common Issues
- **Binary not found**: Run `make build` first
- **Permission denied**: Check file permissions
- **Git errors**: Ensure git is installed and configured
- **Timeout**: Increase timeout with `-timeout 60s`

## Contributing

When adding new features:
1. Write unit tests first (TDD)
2. Add integration tests for feature interactions
3. Include in regression suite for ongoing validation
4. Update this documentation

## Contact

For test-related issues, please file a GitHub issue with:
- Test output
- System information
- Steps to reproduce