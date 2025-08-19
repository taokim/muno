# rc agents - Agent Process Monitoring

The `rc agents` command provides detailed information about running Claude Code agents, similar to the `ps` command but specifically designed for repo-claude.

## Overview

View live agents with process information including CPU usage, memory consumption, and activity status.

## Usage

```bash
# Show running agents (default)
rc agents

# Show all agents including stopped ones
rc agents -a
rc agents --all

# Show detailed process information
rc agents -d
rc agents --details

# Show with recent log entries
rc agents -l
rc agents --logs

# Combine options
rc agents -d -l     # Details with logs
rc agents -a -d     # All agents with details

# Sort output
rc agents --sort cpu    # Sort by CPU usage
rc agents --sort mem    # Sort by memory usage
rc agents --sort time   # Sort by last activity
rc agents --sort pid    # Sort by process ID

# Output formats
rc agents --format table    # Default table view
rc agents --format simple   # Simple list
rc agents --format json     # JSON output (for scripting)
```

## Output Examples

### Basic View

```
==========================================================================================
 CLAUDE CODE AGENTS
==========================================================================================
 Workspace: my-project
 Time: 2024-01-15 14:32:15
------------------------------------------------------------------------------------------
NAME            STATUS          PID     REPOSITORY      SPECIALIZATION
----            ------          ---     ----------      --------------
frontend-dev    🟢 running      12345   frontend        React, TypeScript, responsive des...
backend-dev     🟢 running      12346   backend         Node.js, Express, PostgreSQL, RES...
mobile-dev      ⚫ stopped      0       mobile          React Native, iOS, Android, mobil...
==========================================================================================
 Summary: 2 running, 3 total

💡 Tips:
  rc agents -d        # Show detailed process info
  rc agents -a        # Show all agents (including stopped)
  rc agents -l        # Show with recent logs
  rc start <agent>    # Start a specific agent
  rc stop <agent>     # Stop a specific agent
```

### Detailed View (`-d` flag)

```
==========================================================================================
 CLAUDE CODE AGENTS
==========================================================================================
 Workspace: my-project
 Time: 2024-01-15 14:32:15
------------------------------------------------------------------------------------------
NAME            STATUS          PID     CPU%    MEM(MB) TIME    REPO        COMMAND
----            ------          ---     ----    ------- ----    ----        -------
frontend-dev    🟢 running      12345   2.3     245.2   1h23m   frontend    claude --model claude-3-5-sonnet...
backend-dev     🟢 running      12346   1.5     189.7   1h23m   backend     claude --model claude-3-5-sonnet...
mobile-dev      ⚫ stopped      0       -       -       -       mobile      -
==========================================================================================
 Summary: 2 running, 3 total
```

### With Logs (`-l` flag)

```
[Agent listing as above]

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

## Features

### Process Monitoring
- **PID**: Process ID for system-level tracking
- **CPU%**: Current CPU usage percentage
- **Memory**: Memory usage in MB
- **Elapsed Time**: How long the agent has been running
- **Command**: The actual command being executed

### Health Checking
- Automatically detects dead processes and updates status
- Verifies process existence before displaying
- Updates state file when processes are found dead

### Log Integration
- Shows recent log entries when using `-l` flag
- Helps understand what agents are currently doing
- Useful for debugging and monitoring activity

### Sorting Options
- Sort by name (alphabetical)
- Sort by CPU usage (highest first)
- Sort by memory usage (highest first)
- Sort by last activity time (most recent first)
- Sort by process ID

## Platform Support

### macOS & Linux
- Full process information using `ps` command
- Accurate CPU and memory statistics
- Process elapsed time tracking

### Windows
- Uses WMI (Windows Management Instrumentation)
- Basic process information available
- Some metrics may be approximated

## Integration with Other Commands

```bash
# Check agent status
rc agents

# Start stopped agents
rc start mobile-dev

# Stop running agents
rc stop frontend-dev

# Start agents and monitor
rc start --foreground frontend-dev  # In one terminal
rc agents -d                         # In another terminal

# Check logs for debugging
rc agents -l | grep ERROR
```

## Use Cases

### Development Monitoring
Monitor resource usage during development:
```bash
# Watch for high CPU usage
rc agents -d --sort cpu

# Check memory consumption
rc agents -d --sort mem
```

### Debugging
See what agents are doing:
```bash
# View recent activity
rc agents -l

# Check specific agent
rc agents | grep frontend-dev
```

### Automation
Use simple format for scripts:
```bash
# Get list of running agents
rc agents --format simple

# Parse JSON output
rc agents --format json | jq '.agents[] | select(.status=="running")'
```

## Tips

1. **Regular Monitoring**: Run `rc agents -d` periodically to check resource usage
2. **Log Review**: Use `rc agents -l` to understand agent behavior
3. **Health Checks**: The command automatically updates status for dead processes
4. **Sorting**: Use sort options to quickly identify resource-heavy agents
5. **Scripting**: Use `--format simple` or `--format json` for automation