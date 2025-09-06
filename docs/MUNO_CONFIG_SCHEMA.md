# MUNO Configuration Schema

## Complete Configuration Reference

### muno.yaml Structure

```yaml
# Workspace configuration (all fields optional with defaults)
workspace:
  name: "muno-workspace"      # Default: "muno-workspace"
  repos_dir: "repos"          # Default: "repos" - where repos are cloned
  root_repo: ""               # Default: "" - URL if workspace itself is a git repo

# Node definitions (array of child nodes)
nodes:
  - name: "required-field"    # REQUIRED: Node name
    url: "git-url"           # EITHER url OR config (mutually exclusive)
    config: "path/to/config" # EITHER url OR config (mutually exclusive)
    lazy: true               # Optional, intelligent defaults:
                            #   - Default: true (lazy) for regular repos
                            #   - Default: false (eager) for meta-repos
                            #   - Meta-repo patterns: *-monorepo, *-munorepo, 
                            #     *-muno, *-metarepo, *-platform, *-workspace, *-root-repo
```

## Default Values

### Workspace Defaults
- **workspace.name**: `"muno-workspace"`
- **workspace.repos_dir**: `"repos"`
- **workspace.root_repo**: `""` (empty)

### Node Defaults
- **node.lazy**: `true` (except for meta-repos which default to `false`)
  - Meta-repo patterns that default to eager loading:
    - `*-monorepo`
    - `*-munorepo`
    - `*-muno`
    - `*-metarepo`
    - `*-platform`
    - `*-workspace`
    - `*-root-repo`

### Behavior Defaults
- **Auto-clone on navigation**: `true`
- **Git remote**: `"origin"`
- **Git branch**: `"main"`
- **Clone timeout**: `300` seconds
- **Max parallel clones**: `4`
- **Max parallel pulls**: `8`

## Node Types

### 1. Git Repository Nodes (`url` field)
Direct git repositories that can be cloned and managed.

**Fields:**
- `name` (required): Node identifier
- `url` (required): Git repository URL
- `lazy` (optional): Clone behavior
  - Omitted: Uses intelligent defaults
  - `false`: Clone immediately (eager)
  - `true`: Clone on-demand (lazy)

**Example:**
```yaml
nodes:
  - name: frontend
    url: https://github.com/org/frontend.git
    # lazy defaults to true (omitted)
```

### 2. Config Reference Nodes (`config` field)
Delegate subtree management to external configurations.

**Fields:**
- `name` (required): Node identifier
- `config` (required): Path to configuration file
  - Local: `./path/to/muno.yaml`
  - Remote: `https://config.example.com/muno.yaml`

**Example:**
```yaml
nodes:
  - name: team-backend
    config: ./teams/backend/muno.yaml
```

## Configuration Rules

1. **Each node MUST have a `name`**
2. **Each node MUST have EITHER `url` OR `config` (never both)**
3. **The `lazy` field is optional and uses intelligent defaults**
4. **Config nodes cannot be marked as lazy** (config delegation is always immediate)

## Examples

### Good Configuration Examples

#### Minimal Configuration (Uses All Defaults)
```yaml
workspace:
  name: my-platform
nodes:
  - name: frontend
    url: https://github.com/org/frontend.git
  - name: backend
    url: https://github.com/org/backend.git
```

#### Mixed Node Types with Smart Defaults
```yaml
workspace:
  name: engineering
nodes:
  - name: services-monorepo  # Will be eager (ends with -monorepo)
    url: https://github.com/org/services-monorepo.git
  
  - name: utility-lib       # Will be lazy (default)
    url: https://github.com/org/utility.git
  
  - name: team-config       # Config reference node
    config: ./teams/backend/muno.yaml
```

#### Explicit Lazy Configuration
```yaml
nodes:
  - name: large-dataset
    url: https://github.com/org/dataset.git
    lazy: true  # Explicitly lazy (though this is the default)
  
  - name: core-services
    url: https://github.com/org/core.git
    lazy: false  # Explicitly eager (override default)
```

### Bad Configuration Examples (Avoid These)

#### ERROR: Both url and config
```yaml
nodes:
  - name: broken-node
    url: https://github.com/org/repo.git
    config: ./config.yaml  # ERROR: Cannot have both url AND config
```

#### REDUNDANT: Specifying defaults
```yaml
nodes:
  - name: redundant-service
    url: https://github.com/org/service.git
    lazy: true  # REDUNDANT: This is already the default, can be omitted
```

#### ERROR: Config node with lazy
```yaml
nodes:
  - name: invalid-config
    config: ./team/muno.yaml
    lazy: true  # ERROR: Config nodes cannot be lazy
```

## Organization Strategies

### Team-Based Organization
Use config references to let each team manage their subtree:
```yaml
nodes:
  - name: team-frontend
    config: ./teams/frontend/muno.yaml
  - name: team-backend
    config: ./teams/backend/muno.yaml
  - name: team-platform
    config: ./teams/platform/muno.yaml
```

### Service-Type Organization
Group by architectural layers:
```yaml
nodes:
  - name: apis
    config: ./layers/apis/muno.yaml
  - name: frontends
    config: ./layers/frontends/muno.yaml
  - name: libraries
    config: ./layers/libraries/muno.yaml
```

### Domain-Driven Organization
Follow domain boundaries:
```yaml
nodes:
  - name: commerce-domain
    config: ./domains/commerce/muno.yaml
  - name: identity-domain
    config: ./domains/identity/muno.yaml
  - name: payment-domain
    config: ./domains/payment/muno.yaml
```

## Node Type Indicators in Tree Display

When using `muno tree`, nodes are displayed with type indicators:
- **[git: URL]**: Git repository node with its URL
- **[config: PATH]**: Config reference node with its config file path
- **[lazy]**: Repository will be cloned on-demand
- **[not cloned]**: Repository exists in config but hasn't been cloned yet

## File Locations

- **Configuration file**: `muno.yaml` or `.muno.yaml` in workspace root
- **State file**: `.muno-tree.json` (auto-generated, do not edit)
- **Repositories**: `<workspace>/<repos_dir>/<node-name>/`
- **Agent context**: `<workspace>/.muno/agent-context.md` (auto-generated)