# MUNO v1.0.1 Patch Release

## Release Date
September 12, 2025

## Overview
This patch release fixes critical issues discovered through regression testing, ensuring 100% test suite pass rate and proper lazy loading behavior.

## Fixed Issues

### 1. Git Status Command Format (CRITICAL)
- **Issue**: Status command failed to detect untracked files
- **Root Cause**: `git status` was called without `--short` flag, causing parser mismatch
- **Fix**: Updated `RealGit.Status()` to use `--short` flag for proper parsing
- **Impact**: Status command now correctly identifies file changes and untracked files

### 2. Lazy Repository Loading Behavior
- **Issue**: Lazy repositories were being auto-cloned when navigating to parent nodes
- **Root Cause**: Manager was proactively cloning lazy children "for better experience"
- **Fix**: Disabled auto-cloning to preserve true lazy loading semantics
- **Impact**: Lazy repositories now only clone when explicitly navigated to

### 3. Repository Path Resolution
- **Issue**: Top-level repositories were created in wrong directory
- **Root Cause**: `ComputeFilesystemPath` not using configured `repos_dir`
- **Fix**: Updated path computation to correctly use `repos_dir` for all repositories
- **Impact**: Repositories now consistently placed in `nodes/` directory

### 4. Directory Pre-creation for Lazy Repos
- **Issue**: Empty directories created for lazy repositories during initialization
- **Root Cause**: `buildTreeFromConfig` creating directories regardless of lazy status
- **Fix**: Only create directories when actually cloning repositories
- **Impact**: Cleaner workspace with no empty placeholder directories

## Testing
- All 36 regression tests passing (100% pass rate)
- Full test suite validates:
  - Configuration persistence
  - Navigation and lazy loading
  - Git operations
  - Error handling
  - Repository management

## Upgrade Instructions
```bash
# For source builds
git pull
make build

# For installed binaries
go install github.com/taokim/muno/cmd/muno@v1.0.1
```

## Compatibility
- Fully backward compatible with v1.0.0
- No configuration changes required
- Existing workspaces will work without modification

## Contributors
- Bug fixes and regression testing improvements

## Next Release
Planning v1.1.0 with API signature management and schema registry features (see roadmap in CLAUDE.md)