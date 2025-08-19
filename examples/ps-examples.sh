#!/bin/bash
# Examples showing rc ps command usage (Unix ps style)

echo "🔍 rc ps - Unix-style Process Monitoring"
echo "========================================"

echo -e "\n1. Basic usage (like Unix ps):"
echo "$ rc ps"
echo "# Shows running agents only"

echo -e "\n2. Show all agents (like ps aux):"
echo "$ rc ps aux"
echo "# Shows all agents with CPU, memory, and extended info"

echo -e "\n3. Full format (like ps -ef):"
echo "$ rc ps -ef"
echo "# Shows all agents in full format"

echo -e "\n4. Individual flags:"
echo "$ rc ps -a    # All agents including stopped"
echo "$ rc ps -x    # Extended info (CPU, memory)"
echo "$ rc ps -u    # User-oriented format"
echo "$ rc ps -l    # Long format"

echo -e "\n5. Sorting options:"
echo "$ rc ps aux --sort cpu    # Sort by CPU usage"
echo "$ rc ps aux --sort mem    # Sort by memory"
echo "$ rc ps -a --sort time    # Sort by runtime"

echo -e "\n6. With logs:"
echo "$ rc ps --logs            # Show recent activity"
echo "$ rc ps aux --logs        # Full details + logs"

echo -e "\n7. Common patterns:"
echo "# Find high CPU agents"
echo '$ rc ps aux | awk "$4 > 10.0 {print $1, $4}"'

echo -e "\n# Count running agents"
echo '$ rc ps | grep -c "running"'

echo -e "\n# Get PIDs of running agents"
echo '$ rc ps -x | awk "/running/ {print $3}"'

echo -e "\n# Monitor continuously"
echo '$ watch -n 5 "rc ps aux --sort cpu"'

echo -e "\n8. Scripting examples:"
echo "# Stop agents using too much memory (>500MB)"
echo 'rc ps aux | awk "$5 > 500 {print $1}" | xargs -I {} rc stop {}'

echo -e "\n# Export to CSV"
echo 'rc ps -x | awk "NR>2 {print $1\",\"$2\",\"$3\",\"$4\",\"$5}"'

echo -e "\n💡 Tips:"
echo "- No dash required: 'rc ps aux' = 'rc ps -aux'"
echo "- Combines with grep/awk for powerful filtering"
echo "- Use watch for real-time monitoring"
echo "- Sort by resource usage to find problems"