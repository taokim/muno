# Changelog

All notable changes to Repo-Claude will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2024-12-24

### Added
- **Interactive TUI for start command** - Beautiful terminal UI using Bubbletea framework
  - Visual scope selection with keyboard navigation
  - Real-time status updates and feedback
  - Improved user experience for scope management
  
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

- **Development Workflow Improvements**
  - Enhanced Makefile with better build targets
  - External dependency abstraction through interfaces for better testing
  - Improved CommandExecutor interface for mocking

### Changed
- **Claude sessions now always start in current terminal window** - Better UX for single session workflows
- **Sync uses rebase by default** - `rc sync` now uses `git pull --rebase` instead of `--ff-only`
  - Maintains cleaner, linear commit history
  - Better for trunk-based development workflows
  - Avoids unnecessary merge commits

### Fixed
- Ensure consistent working directory for Claude sessions
- Correct Makefile variable syntax in install-dev target
- Resolve test failures by using CommandExecutor interface
- Interactive TUI selection modes now work properly

### Documentation
- Added comprehensive features overview (`docs/features.md`)
- Created detailed PR management guide (`docs/pr-management.md`)
- Updated README with key features section
- Added workflow examples for common scenarios

## [0.3.2] - 2024-08-24

### Fixed
- Working directory consistency issues

## [0.3.1] - 2024-08-23

### Fixed
- Minor bug fixes and improvements

## [0.3.0] - 2024-08-22

### Added
- Scope-based development architecture
- Direct git management without Google repo tool dependency

## [0.2.1] - 2024-08-20

### Fixed
- Terminal management improvements

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

[Unreleased]: https://github.com/taokim/repo-claude/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/taokim/repo-claude/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/taokim/repo-claude/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/taokim/repo-claude/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/taokim/repo-claude/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/taokim/repo-claude/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/taokim/repo-claude/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/taokim/repo-claude/releases/tag/v0.1.0