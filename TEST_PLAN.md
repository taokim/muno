# MUNO Path Resolution Test Plan

## Quick Session Guide

### 🚀 Starting a New Session

```bash
# 1. Load the test plan from memory
# Tell Claude: "Load the muno path resolution test plan memory"

# 2. Check current status
make test-master
go test -cover ./internal/manager

# 3. Review what was done
git log --oneline -5
cat TEST_PLAN.md | grep "Session.*✅"

# 4. Start your session
# Tell Claude: "Let's continue with Session X of the test plan"
```

### 📋 Session Checklist

**Session 1: Foundation & Critical Path Tests**
- [ ] Create test_helpers.go
- [ ] Write ResolvePath tests (15-20 tests)
- [ ] Fix failing `path --ensure` test
- [ ] Coverage: 43% → 50%+

**Session 2: Tree Path Conversion Tests**
- [ ] Write GetTreePath tests
- [ ] Write buildTreePathFromFilesystem tests
- [ ] Create test fixtures in testdata/
- [ ] Coverage: 50% → 60%+

**Session 3: Filesystem Path & Regression**
- [ ] Write computeFilesystemPath tests
- [ ] Add missing edge cases
- [ ] Update regression tests
- [ ] Coverage: 60% → 65%+

**Session 4: Edge Cases & Integration**
- [ ] Create integration tests
- [ ] Add edge case tests
- [ ] Test error scenarios
- [ ] Coverage: 65% → 70%+

**Session 5: Performance & Polish**
- [ ] Add benchmarks
- [ ] Optimize slow tests
- [ ] Update documentation
- [ ] Coverage: 70% → 75%+

### 💾 Saving Progress

After each session:
```bash
# Commit your work
git add -A
git commit -m "test: [session X] description of what was done"

# Update this file with progress
# Mark completed items with ✅

# Run tests and save results
go test -coverprofile=coverage.out ./internal/manager
echo "Session X: $(go test -cover ./internal/manager | grep coverage)" >> coverage_log.txt
```

### 🔄 Loading Context in Claude

When starting a new Claude session, say:

> "I'm working on improving test coverage for MUNO path resolution. Please:
> 1. Load the memory 'muno_path_resolution_test_plan'
> 2. Check TEST_PLAN.md for current progress
> 3. Continue with Session [X] of the test plan"

### 📊 Coverage Targets

| Session | Target | Key Functions |
|---------|--------|---------------|
| 1 | 50% | ResolvePath |
| 2 | 60% | GetTreePath, buildTreePathFromFilesystem |
| 3 | 65% | computeFilesystemPath |
| 4 | 70% | Edge cases, integration |
| 5 | 75% | Performance, optimization |

### 🛠️ Useful Commands

```bash
# Run specific tests
go test -v ./internal/manager -run TestManager_ResolvePath
go test -v ./internal/manager -run TestManager_GetTreePath

# Check coverage for specific functions
go test -coverprofile=coverage.out ./internal/manager
go tool cover -func=coverage.out | grep -E "ResolvePath|GetTreePath|buildTree|computeFilesystem"

# Run regression tests
make test-master
bash test/run_regression_tests.sh test_path_and_mcd

# Quick test all
make test-all
```

### 📝 Progress Log

| Date | Session | Coverage | Notes |
|------|---------|----------|-------|
| 2024-11-03 | Setup | 43.1% | Fixed parent navigation bug, planned test coverage |
| - | Session 1 | - | Pending |
| - | Session 2 | - | Pending |
| - | Session 3 | - | Pending |
| - | Session 4 | - | Pending |
| - | Session 5 | - | Pending |

### 🎯 Current Focus

**Next Session**: Session 1 - Foundation & Critical Path Tests
**Primary Goal**: Create test infrastructure and ResolvePath tests
**Expected Time**: 2-3 hours

---

*Last Updated: 2024-11-03*
*Test Plan Memory: `muno_path_resolution_test_plan`*