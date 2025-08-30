#!/bin/bash
# Manual test script for new start command features

echo "🧪 Testing improved muno start command"
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
echo "../muno start --help"

echo -e "\n2. Test foreground mode (you'll need to Ctrl+C to exit):"
echo "../muno start frontend-dev --foreground"

echo -e "\n3. Test multiple agents:"
echo "../muno start frontend-dev backend-dev"

echo -e "\n4. Test by repository selection:"
echo "../muno start --repos frontend,backend"

echo -e "\n5. Test interactive selection:"
echo "../muno start --interactive"

echo -e "\n6. Test new window mode (if supported on your OS):"
echo "../muno start frontend-dev --new-window"

echo -e "\n7. Test preset (once implemented in config):"
echo "../muno start --preset fullstack"

echo -e "\n8. Show status after starting:"
echo "../muno status"

echo -e "\n9. Stop all agents:"
echo "../muno stop"

echo -e "\nTest workspace created at: $(pwd)"
echo "Run the commands above to test different features"
echo "Clean up with: rm -rf $(pwd)"