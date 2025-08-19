#!/bin/bash
# Manual test script for new start command features

echo "🧪 Testing improved rc start command"
echo "===================================="

# Build the tool
echo "Building repo-claude..."
cd ../
go build -o rc ./cmd/repo-claude
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"

# Create test workspace
TEST_DIR="test_workspace_$(date +%s)"
echo -e "\nCreating test workspace: $TEST_DIR"
mkdir -p $TEST_DIR
cd $TEST_DIR

# Initialize workspace
echo -e "\nInitializing workspace..."
../rc init test-project << EOF
Test Project
3
frontend
https://github.com/example/frontend
main
frontend-dev
Frontend development
claude-3-5-sonnet-20241022
y

backend
https://github.com/example/backend
main
backend-dev
Backend development
claude-3-5-sonnet-20241022
n

mobile
https://github.com/example/mobile
main
mobile-dev
Mobile development
claude-3-5-sonnet-20241022
n

n
EOF

echo -e "\n📋 Test scenarios:"
echo "=================="

echo -e "\n1. Test help output:"
echo "../rc start --help"

echo -e "\n2. Test foreground mode (you'll need to Ctrl+C to exit):"
echo "../rc start frontend-dev --foreground"

echo -e "\n3. Test multiple agents:"
echo "../rc start frontend-dev backend-dev"

echo -e "\n4. Test by repository selection:"
echo "../rc start --repos frontend,backend"

echo -e "\n5. Test interactive selection:"
echo "../rc start --interactive"

echo -e "\n6. Test new window mode (if supported on your OS):"
echo "../rc start frontend-dev --new-window"

echo -e "\n7. Test preset (once implemented in config):"
echo "../rc start --preset fullstack"

echo -e "\n8. Show status after starting:"
echo "../rc status"

echo -e "\n9. Stop all agents:"
echo "../rc stop"

echo -e "\nTest workspace created at: $(pwd)"
echo "Run the commands above to test different features"
echo "Clean up with: rm -rf $(pwd)"