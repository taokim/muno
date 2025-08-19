# Deprecated: rc agents

**This command has been renamed to `rc ps` to better align with Unix conventions.**

Please see [ps-command.md](ps-command.md) for the current documentation.

## Migration

All functionality remains the same, just use `ps` instead of `agents`:

```bash
# Old command → New command
rc agents        → rc ps
rc agents -a     → rc ps -a
rc agents -d     → rc ps -x or rc ps aux
rc agents -l     → rc ps --logs
```

The new `rc ps` command also supports Unix-style usage:
- `rc ps aux` - Show all agents with extended info
- `rc ps -ef` - Show all agents with full format

This change makes the command more intuitive for developers familiar with Unix/Linux systems.