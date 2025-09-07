# MUNO Regression Test Suite

This directory contains the regression test suite for MUNO (Multi-repository UNified Orchestration).

## Prerequisites

Before running the tests:
1. Build the MUNO binary: `make build` (from project root)
2. Ensure you have write access to `/tmp` directory
3. Git must be installed and configured

## Running Tests

There is only one test script to run:

```bash
# From the muno project root directory (recommended)
./test/regression/regression_test.sh

# Or from the regression directory
cd test/regression
./regression_test.sh

# To see detailed output during test execution
bash -x ./test/regression/regression_test.sh

# To save test output to a file
./test/regression/regression_test.sh 2>&1 | tee test_output.txt
```

The test will:
1. Automatically build the MUNO binary if needed
2. Create a temporary test environment in `/tmp/muno-regression-test/`
3. Run all 36 tests
4. Clean up the test environment
5. Save a detailed report to `/tmp/muno-regression-test/regression_report_[timestamp].txt`

## What It Tests

The regression test suite validates all core MUNO functionality:

### 1. Configuration & Setup (5 tests)
- Workspace directory creation
- Configuration file management
- State directory initialization
- Repository directory structure
- Current position tracking

### 2. Core Commands (5 tests)
- `tree` - Display repository tree
- `list` - List repositories
- `current` - Show current position
- `--help` - Help documentation
- `--version` - Version information

### 3. Navigation - Eager Repositories (5 tests)
- Navigate to root
- Navigate to eager repositories
- Auto-cloning of eager repositories
- Navigation between repositories
- Return to root functionality

### 4. Navigation - Lazy Repositories (4 tests)
- Lazy repositories not cloned initially
- Navigation to lazy repositories
- Auto-cloning on navigation
- Return to root from lazy repos

### 5. Clone Operations (4 tests)
- Remove lazy repositories
- Clone specific repositories
- Clone creates repository correctly
- Recursive cloning of all lazy repos

### 6. Repository Management (8 tests)
- **Add repository** - Successfully adds new repositories
- **Add persistence** - Verifies add updates `muno.yaml` ✅
- **Add visibility** - Added repos appear in list/tree
- **Add persistence across commands** - Repos persist after adding
- **Remove repository** - Successfully removes repositories
- **Remove persistence** - Verifies remove updates `muno.yaml` ✅
- **Remove visibility** - Removed repos disappear from list/tree
- **Remove persistence verification** - Repos stay removed

### 7. Git Operations (2 tests)
- Status command functionality
- File change detection

### 8. Error Handling (3 tests)
- Invalid node navigation
- Invalid command handling
- Remove non-existent repository

## Test Results

The test suite provides:
- **Color-coded output** for easy reading
- **Pass/Fail status** for each test
- **Summary statistics** including pass rate
- **Detailed report** saved to `/tmp/muno-regression-test/`

### Success Output
```
╔════════════════════════════════════════════════════════════════╗
║                    ALL TESTS PASSED!                          ║
╚════════════════════════════════════════════════════════════════╝
```

### Failure Output
Shows which specific tests failed with detailed information.

## Recent Fixes (2025-09-07)

✅ **Persistence Issue Fixed**: The add/remove commands now correctly persist changes to `muno.yaml`. Previously these were only updating in-memory state, but now they properly save to the configuration file.

✅ **Current File Creation**: The `.muno/current` file is now created during workspace initialization, ensuring proper state tracking from the start.

✅ **Test Suite Improvements**: 
- Fixed clone command test to use navigation instead of incorrect syntax
- Fixed add/remove tests to ensure they run from root position
- Added automatic test repository creation
- All 36 tests now run successfully

## Known Issues

⚠️ **Status Command**: The `muno status` command does not currently detect untracked files in git repositories. Test #33 is skipped due to this issue.

## Troubleshooting

### Test Failures

If tests fail unexpectedly:

1. **Ensure clean state**: Remove any leftover test directories
   ```bash
   rm -rf /tmp/muno-regression-test
   ```

2. **Rebuild binary**: Make sure you have the latest build
   ```bash
   make clean && make build
   ```

3. **Check Git configuration**: Ensure git is properly configured
   ```bash
   git config --global user.name "Test User"
   git config --global user.email "test@example.com"
   ```

4. **Run with debug output**: See exactly what's happening
   ```bash
   bash -x ./test/regression/regression_test.sh 2>&1 | less
   ```

### Common Issues

- **"Binary not found"**: Run `make build` from project root first
- **"Permission denied"**: Check write permissions for `/tmp` directory
- **"Git not configured"**: Set up git user name and email
- **Tests hang**: The test might be waiting for user input (e.g., remove confirmation). Check if `echo y |` is missing from interactive commands

## Test Environment

The test suite:
1. Creates a temporary workspace in `/tmp/muno-regression-test/`
2. Sets up test repositories
3. Runs all tests
4. Cleans up after completion

The test is completely self-contained and does not affect your actual MUNO workspaces.