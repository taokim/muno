# Config Reference Node Limitation

## Current Behavior

The `muno clone` command currently **does not** process config reference nodes (nodes with a `file` field instead of `url`).

### What Works
- Repository nodes with `url` field are cloned correctly
- Lazy/eager detection works for repository nodes
- `--include-lazy` flag properly includes lazy repositories

### What Doesn't Work
- Config reference nodes (with `file` field) are completely ignored by clone
- Repositories defined in external configuration files are not cloned
- No expansion of distributed configurations during clone operations

## Example

Given this configuration:
```yaml
nodes:
  - name: team-backend
    file: ../backend/muno.yaml    # Config reference - NOT processed by clone
  - name: payment-service  
    url: https://github.com/org/payment.git  # Repository - works with clone
```

Running `muno clone` will:
- ✅ Clone `payment-service` (repository node)
- ❌ Ignore `team-backend` (config reference node)
- ❌ Not clone any repositories defined in `../backend/muno.yaml`

## Workaround

Currently, to clone repositories from config reference nodes:
1. Navigate to the config reference node manually
2. Run clone from within that context
3. Or manually clone the repositories defined in the external config

## Future Enhancement

To properly support config references, the clone command would need to:
1. Detect nodes with `file` field
2. Load and parse the referenced configuration
3. Recursively process repositories from the loaded config
4. Apply the same lazy/eager rules

This would require significant changes to the `CloneRepos` implementation in `internal/manager/manager.go`.

## Impact

This limitation affects:
- Distributed team configurations
- Hierarchical workspace setups
- Config delegation patterns

Teams using config references must manually manage cloning of repositories defined in external configurations.