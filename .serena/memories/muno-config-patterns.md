# MUNO Configuration Patterns and Technical Guidelines

## Configuration Loading Hierarchy
The configuration system follows this precedence order:
1. **Embedded Defaults** (in binary via `//go:embed`)
2. **Project Config File** (`muno.yaml` or `.muno.yaml`)
3. **Runtime Values** (programmatic overrides)

## Key Configuration Files

### Embedded Defaults
- **Location**: `/internal/config/defaults.yaml`
- **Purpose**: All default values embedded in binary
- **Access**: Via `config.GetDefaults()`, `config.GetDefaultReposDir()`, etc.

### Configuration Types
1. **ConfigTree** - Tree-based v3 configuration (current)
2. **Config** - Legacy v1/v2 configuration (backward compatibility)

## Important Patterns

### Adding New Default Values
1. Add to `/internal/config/defaults.yaml`
2. Update `DefaultConfiguration` struct in `defaults.go`
3. Add getter function if needed
4. Values automatically available to entire codebase

### Accessing Configuration Values
```go
// For repos directory
reposDir := config.GetDefaultReposDir()

// For eager patterns
patterns := config.GetEagerLoadPatterns()

// For config file names
names := config.GetConfigFileNames()
```

### Configuration Override Pattern
```go
// In LoadTree function
cfg = *MergeWithDefaults(&cfg)
```

## Testing Configuration
- Default values test: `TestEmbeddedDefaults`
- Override behavior test: `TestConfigurationOverride`
- Merge logic test: `TestMergeWithDefaults`

## Common Configuration Keys

### Workspace Settings
- `workspace.name` - Workspace/project name
- `workspace.repos_dir` - Directory for repositories (default: "nodes")
- `workspace.root_repo` - Optional root repository URL

### Detection Patterns
- `detection.eager_patterns` - Patterns that trigger eager loading
  - `-monorepo`, `-muno`, `-metarepo`
  - `-platform`, `-workspace`, `-root-repo`
- `detection.ignore_patterns` - Patterns to ignore during scanning

### File Settings
- `files.config_names` - Config files to search for
- `files.state_file` - State file name (`.muno-tree.json`)

## Migration Notes
- Constants package (`/internal/constants`) is deprecated
- Only used by legacy v1/v2 config code
- All new code should use config defaults system

## Configuration Validation
- Workspace name is required
- Repos directory defaults to "nodes" if empty
- Validation happens after merge with defaults

## Custom Configuration Example
```yaml
# Project muno.yaml - only override what you need
workspace:
  name: my-project
  repos_dir: repositories  # Custom directory instead of "nodes"
# Everything else uses defaults from embedded config
```