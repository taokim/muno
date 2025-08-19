#!/bin/bash
# Example script showing how to use rc ps for monitoring

echo "🔍 Agent Monitoring Examples"
echo "==========================="

# Check if any agents are running
echo -e "\n1. Quick status check:"
echo "rc ps"
rc ps 2>/dev/null || echo "(Run this in a repo-claude workspace)"

# Monitor resource usage
echo -e "\n2. Resource monitoring (sort by CPU):"
echo "rc ps -d --sort cpu"

# Check agent health
echo -e "\n3. Health check with all agents:"
echo "rc ps -a"

# View agent activity
echo -e "\n4. Agent activity logs:"
echo "rc ps -l"

# Scripting example
echo -e "\n5. Get running agent count:"
echo 'rc ps --format simple | wc -l'

# Watch agents continuously
echo -e "\n6. Continuous monitoring:"
echo "watch -n 5 'rc ps -d'"

# Find high resource users
echo -e "\n7. Find high CPU agents:"
echo "rc ps -d --sort cpu | head -5"

# Check specific agent
echo -e "\n8. Check specific agent:"
echo "rc ps | grep backend-dev"

# Export to JSON for processing
echo -e "\n9. Export for processing:"
echo "rc ps --format json > agents-status.json"

# Kill runaway agents
echo -e "\n10. Find and stop high-CPU agents:"
cat << 'EOF'
# Get agents using >50% CPU
rc ps -d | awk '$4 > 50 {print $1}' | while read agent; do
    echo "High CPU agent: $agent"
    # rc stop $agent  # Uncomment to actually stop
done
EOF

echo -e "\n💡 Pro Tips:"
echo "- Use 'rc ps -d' to see CPU/memory usage"
echo "- Add '-a' to see stopped agents too"
echo "- Use '--sort cpu' or '--sort mem' to find resource hogs"
echo "- Pipe to grep/awk for advanced filtering"
echo "- Use watch command for real-time monitoring"