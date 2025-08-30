# Scripts Directory

This directory contains various scripts for testing, demonstration, and release management of the repo-claude project.

## Overview

The scripts in this directory are organized by purpose:
- **Testing Scripts**: Comprehensive testing frameworks and verification tools
- **Demo Scripts**: Quick demonstrations of repo-claude capabilities
- **Utility Scripts**: Helper scripts for development workflow
- **Release Scripts**: Build and release automation

## Scripts Documentation

### 🧪 Testing Scripts

#### `advanced-test-framework.sh`
**Purpose**: Comprehensive testing framework for v3 tree-based architecture  
**Usage**: `./scripts/advanced-test-framework.sh`

Tests multiple scenarios including:
- Single repository management
- Multi-repository workspaces
- Nested repository structures
- Enterprise-scale configurations
- Tree navigation and operations

**Features**:
- Creates isolated test environments in `/tmp`
- Tests all major commands (init, add, use, tree, etc.)
- Validates configuration management
- Includes cleanup functionality

#### `verify-test-framework.sh`
**Purpose**: Verifies the test framework setup and environment  
**Usage**: `./scripts/verify-test-framework.sh`

Validates:
- Test directory structure
- Git repository detection
- Command execution paths
- Framework dependencies

#### `test-smart-init.sh`
**Purpose**: Tests the smart initialization feature that detects existing git repositories  
**Usage**: `./scripts/test-smart-init.sh`

Tests:
- Automatic repository discovery
- Interactive repository selection
- Repository migration to repos directory
- Configuration generation

#### `test-workflow.sh`
**Purpose**: Tests common workflow scenarios  
**Usage**: `./scripts/test-workflow.sh`

Covers:
- Project initialization workflows
- Repository management workflows
- Tree navigation patterns
- Common user scenarios

### 🎯 Demo Scripts

#### `quick-demo.sh`
**Purpose**: Quick demonstration of repo-claude capabilities  
**Usage**: `./scripts/quick-demo.sh`

Demonstrates:
- Basic initialization
- Repository addition
- Tree navigation
- Core commands

**Great for**:
- First-time users
- Quick feature demonstrations
- Validation of installation

### 🔧 Utility Scripts

#### `generate-test-tree.sh`
**Purpose**: Generates test tree structures for development  
**Usage**: `./scripts/generate-test-tree.sh [structure-type]`

Options:
- `simple`: Basic tree with few repos
- `complex`: Multi-level nested structure
- `enterprise`: Large-scale tree structure

#### `use-test-repos.sh`
**Purpose**: Sets up test repositories for development  
**Usage**: `./scripts/use-test-repos.sh`

Creates:
- Sample repositories with commits
- Various repository configurations
- Test data for development

### 📦 Release Scripts

#### `release.sh`
**Purpose**: Automates the release process  
**Usage**: `./scripts/release.sh [version]`

**⚠️ Note**: This script is deprecated. Releases should be done via GitHub Actions by creating a git tag.

Process:
- Version validation
- Changelog generation (if applicable)
- Build creation
- GitHub release creation

## Environment Variables

Scripts respect the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `RC_BIN` | Path to rc binary | `./bin/rc` |
| `TEST_DIR` | Base directory for tests | `/tmp` |
| `VERBOSE` | Enable verbose output | `false` |
| `CLEANUP` | Auto-cleanup after tests | `true` |

## Usage Examples

### Running a Quick Demo
```bash
cd repo-claude-go
./scripts/quick-demo.sh
```

### Running Comprehensive Tests
```bash
# Run the advanced test framework
./scripts/advanced-test-framework.sh

# Verify the test setup
./scripts/verify-test-framework.sh
```

### Testing Smart Initialization
```bash
# Test smart init with existing repositories
./scripts/test-smart-init.sh
```

### Setting Up Test Environment
```bash
# Generate a complex test tree
./scripts/generate-test-tree.sh complex

# Set up test repositories
./scripts/use-test-repos.sh
```

## Script Conventions

All scripts follow these conventions:

1. **Exit Codes**:
   - `0`: Success
   - `1`: General error
   - `2`: Missing dependencies
   - `3`: Invalid arguments

2. **Output**:
   - Info messages to stdout
   - Error messages to stderr
   - Color coding when terminal supports it

3. **Safety**:
   - Scripts use `set -e` to exit on errors
   - Cleanup handlers for temporary resources
   - Confirmation prompts for destructive operations

4. **Logging**:
   - Verbose mode available with `-v` flag
   - Debug output with `DEBUG=1` environment variable

## Contributing

When adding new scripts:

1. Follow the naming convention: `purpose-description.sh`
2. Add comprehensive comments and usage information
3. Include error handling and cleanup
4. Update this README with documentation
5. Test on both macOS and Linux if possible

## Troubleshooting

### Common Issues

**Permission Denied**
```bash
chmod +x scripts/*.sh
```

**Command Not Found**
```bash
# Ensure rc is built
make build

# Or specify path
RC_BIN=/path/to/rc ./scripts/script-name.sh
```

**Test Failures**
```bash
# Run with verbose output
VERBOSE=true ./scripts/advanced-test-framework.sh

# Disable cleanup to inspect state
CLEANUP=false ./scripts/test-workflow.sh
```

## Dependencies

Scripts require:
- Bash 4.0+
- Git 2.0+
- Go 1.19+ (for building)
- Standard Unix tools (grep, sed, awk, etc.)

## License

All scripts are part of the repo-claude project and follow the same license.