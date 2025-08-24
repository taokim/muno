# Changelog

All notable changes to Repo-Claude will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Branch Management Commands** - Complete branch lifecycle management across repositories
  - `rc branch create` - Create same branch in multiple repositories
  - `rc branch list` - Visual overview of branch status across repos
  - `rc branch checkout` - Synchronized branch switching
  - `rc branch delete` - Clean up branches with remote option
  
- **Batch Pull Request Creation** - Create PRs across multiple repositories
  - `rc pr batch-create` - Create PRs for all feature branches with one command
  - Safety checks prevent PRs from main branch
  - Automatic detection of uncommitted changes
  - Auto-push for unpushed branches
  - Detailed results reporting per repository
  
- **Enhanced PR Management** - Full GitHub PR workflow integration
  - `rc pr list` - List PRs across all repositories
  - `rc pr create` - Create individual PRs
  - `rc pr status` - Check PR status with CI/CD results
  - `rc pr checkout` - Review PRs locally
  - `rc pr merge` - Merge with multiple strategies

### Changed
- **Sync uses rebase by default** - `rc sync` now uses `git pull --rebase` instead of `--ff-only`
  - Maintains cleaner, linear commit history
  - Better for trunk-based development workflows
  - Avoids unnecessary merge commits

### Documentation
- Added comprehensive features overview (`docs/features.md`)
- Created detailed PR management guide (`docs/pr-management.md`)
- Updated README with key features section
- Added workflow examples for common scenarios

## [0.2.0] - 2024-01-15

### Added
- Scope-based architecture for better repository grouping
- Terminal tab support for multiple Claude sessions
- Environment variables for scope context
- PS command with numbered output for easy kill targeting

### Changed
- Migrated from Python to Go for better performance
- Removed dependency on Google's repo tool
- Simplified configuration to single YAML file

### Fixed
- Terminal window management on macOS
- State persistence across restarts

## [0.1.0] - 2024-01-01

### Added
- Initial release with basic multi-repository support
- Agent-based Claude Code orchestration
- Shared memory for cross-agent coordination
- Basic git operations (clone, sync, status)

[Unreleased]: https://github.com/taokim/repo-claude/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/taokim/repo-claude/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/taokim/repo-claude/releases/tag/v0.1.0