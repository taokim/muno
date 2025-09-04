#!/bin/bash
# Manual test script for agent command features

echo "🧪 Testing muno agent/claude/gemini commands"
echo "===================================="

# Build the tool
echo "Building muno..."
cd ../
go build -o muno ./cmd/muno
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
../muno init test-project << EOF
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
echo "../muno agent --help"

echo -e "\n2. Test foreground mode (you'll need to Ctrl+C to exit):"
echo "../muno claude frontend-dev"

echo -e "\n3. Test multiple agents:"
echo "../muno agent claude frontend-dev"

echo -e "\n4. Test by repository selection:"
echo "../muno agent gemini"

echo -e "\n5. Test interactive selection:"
echo "../muno claude --help"

echo -e "\n6. Test new window mode (if supported on your OS):"
echo "../muno gemini frontend-dev"

echo -e "\n7. Test preset (once implemented in config):"
echo "../muno agent cursor"

echo -e "\n8. Show status after starting:"
echo "../muno status"

echo -e "\n9. Stop all agents:"
echo "../muno stop"

echo -e "\nTest workspace created at: $(pwd)"
echo "Run the commands above to test different features"
echo "Clean up with: rm -rf $(pwd)"