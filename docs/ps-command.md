# rc ps - Agent Process Status

The `rc ps` command displays information about running Claude Code agents, modeled after the Unix `ps` command.

## Overview

View live agents with process information including CPU usage, memory consumption, and activity status - just like the standard Unix `ps` command but specifically for repo-claude agents.

## Usage

```bash
# Show running agents (default)
rc ps

# Unix-style usage (no dashes needed!)
rc ps aux      # All processes, user-oriented format, extended info
rc ps ef       # All processes, full format

# Individual flags
rc ps -a       # Show all agents (including stopped)
rc ps -x       # Show extended information (CPU, memory)
rc ps -u       # User-oriented format with resource usage
rc ps -f       # Full format listing
rc ps -l       # Long format (detailed)

# Additional options
rc ps --logs           # Show with recent log entries
rc ps --sort cpu       # Sort by CPU usage
rc ps --sort mem       # Sort by memory usage
rc ps aux --sort cpu   # Combine Unix style with sorting
```

## Unix Compatibility

The command supports common `ps` usage patterns:

```bash
# These all work the same way:
rc ps aux
rc ps -aux
rc ps -a -u -x

# Common combinations:
rc ps          # Just running agents (like 'ps' without args)
rc ps aux      # All agents with extended info (like 'ps aux')
rc ps -ef      # All agents, full format (like 'ps -ef')
```

## Output Examples

### Basic View (`rc ps`)

```
NAME            STATUS          PID     REPOSITORY      SPECIALIZATION
----            ------          ---     ----------      --------------
frontend-dev    🟢 running      12345   frontend        React, TypeScript, responsive des...
backend-dev     🟢 running      12346   backend         Node.js, Express, PostgreSQL, RES...
```

### Extended View (`rc ps aux`)

```
==========================================================================================
 CLAUDE CODE AGENTS
==========================================================================================
 Workspace: my-project
 Time: 2024-01-15 14:32:15
------------------------------------------------------------------------------------------
NAME            STATUS          PID     CPU%    MEM(MB) TIME    REPO        COMMAND
----            ------          ---     ----    ------- ----    ----        -------
frontend-dev    🟢 running      12345   2.3     245.2   1h23m   frontend    claude --model...
backend-dev     🟢 running      12346   1.5     189.7   1h23m   backend     claude --model...
mobile-dev      ⚫ stopped      0       -       -       -       mobile      -
==========================================================================================
 Summary: 2 running, 3 total

💡 Tips:
  rc ps aux           # Show all agents with details
  rc ps -ef           # Full format listing
  rc ps --logs        # Show with recent logs
  rc start <agent>    # Start a specific agent
  rc stop <agent>     # Stop a specific agent
```

### With Logs (`rc ps --logs`)

```
[Process listing as above]

📋 Recent logs for frontend-dev:
----------------------------------------------------------------------
  2024-01-15 14:30:45 - Analyzing component structure...
  2024-01-15 14:31:02 - Created new Button component
  2024-01-15 14:31:15 - Running tests...

📋 Recent logs for backend-dev:
----------------------------------------------------------------------
  2024-01-15 14:29:30 - Implementing user authentication endpoint
  2024-01-15 14:30:15 - Added validation middleware
  2024-01-15 14:31:50 - Database schema updated
```

## Flag Reference

### Unix-Compatible Flags

| Flag | Description | Unix ps equivalent |
|------|-------------|-------------------|
| `-a` | Show all agents (including stopped) | Show all processes |
| `-x` | Extended info (CPU, memory) | Show processes without tty |
| `-u` | User-oriented format | User format |
| `-f` | Full format listing | Full listing |
| `-l` | Long format | Long listing |
| `aux` | Combination: -a -u -x | Common Unix pattern |
| `ef` | Combination: -e -f | System V style |

### Additional Options

| Option | Description |
|--------|-------------|
| `--logs` | Show recent log entries |
| `--sort cpu` | Sort by CPU usage (highest first) |
| `--sort mem` | Sort by memory usage |
| `--sort time` | Sort by last activity |
| `--sort pid` | Sort by process ID |
| `--sort name` | Sort by agent name (default) |

## Examples

### Daily Development

```bash
# Quick check
rc ps

# See what's using resources
rc ps aux --sort cpu

# Check all agents including stopped
rc ps -a
```

### Debugging

```bash
# See what agents are doing
rc ps --logs

# Find high CPU usage
rc ps aux | grep -E "[0-9]{2,}\.[0-9].*claude"

# Monitor continuously
watch -n 5 'rc ps aux'
```

### Scripting

```bash
# Count running agents
rc ps | grep "running" | wc -l

# Get agent PIDs
rc ps -x | awk '/running/ {print $3}'

# Find and stop high-CPU agents
rc ps aux | awk '$4 > 50 {print $1}' | xargs -I {} rc stop {}
```

## Platform Support

### macOS & Linux
- Full process information using native `ps` command
- Accurate CPU and memory statistics
- Process elapsed time tracking

### Windows
- Uses WMI (Windows Management Instrumentation)
- Basic process information available
- Some metrics may be approximated

## Comparison with Unix ps

| Unix ps | rc ps | Description |
|---------|-------|-------------|
| `ps` | `rc ps` | Show your processes |
| `ps aux` | `rc ps aux` | Show all processes with details |
| `ps -ef` | `rc ps -ef` | Full format, all processes |
| `ps -p PID` | N/A | Not implemented (use grep) |
| `ps -C name` | N/A | Not implemented (use grep) |

## Integration

```bash
# Start agents and monitor
rc start backend --foreground &  # Start in background
rc ps aux                        # Check status

# Stop high-resource agents
rc ps aux --sort cpu | head -n 3
rc stop <high-cpu-agent>

# Health monitoring
while true; do
    rc ps aux
    sleep 60
done
```

## Tips

1. **Muscle Memory**: If you're used to `ps aux`, it works exactly the same way
2. **No Dash Required**: Both `rc ps aux` and `rc ps -aux` work
3. **Sorting**: Use `--sort` for quick resource analysis
4. **Logs**: Add `--logs` to see what agents are actively doing
5. **Scripting**: Output is grep-friendly for automation