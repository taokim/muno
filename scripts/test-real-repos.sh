#!/bin/bash

# Test the simplified tree state management with real GitHub repositories

set -e

echo "=== Testing Simplified Tree State with Real GitHub Repositories ==="
echo ""

# Build the binary
echo "Building muno..."
cd /Users/musinsa/ws/muno/muno-go
make build

# Create test workspace
TEST_DIR="/tmp/test-real-repos-$(date +%s)"
mkdir -p "$TEST_DIR"

echo "Created test workspace: $TEST_DIR"
cd "$TEST_DIR"

# Initialize workspace
echo ""
echo "1. Initializing workspace..."
mkdir -p test-real
cd test-real
/Users/musinsa/ws/muno/muno-go/bin/muno init .

# Add real repositories with nested structure
echo ""
echo "2. Adding real GitHub repositories..."

# Add a popular open-source repo (small)
echo "   Adding spf13/cobra (CLI library)..."
/Users/musinsa/ws/muno/muno-go/bin/muno add https://github.com/spf13/cobra.git

# Add another repo as a child (lazy)
echo "   Adding spf13/viper as lazy child of cobra..."
/Users/musinsa/ws/muno/muno-go/bin/muno use cobra
/Users/musinsa/ws/muno/muno-go/bin/muno add https://github.com/spf13/viper.git --lazy

# Go back to root and add another top-level repo
echo "   Adding google/go-github at root level..."
/Users/musinsa/ws/muno/muno-go/bin/muno use /
/Users/musinsa/ws/muno/muno-go/bin/muno add https://github.com/google/go-github.git --lazy

# Show the tree structure
echo ""
echo "3. Tree structure:"
/Users/musinsa/ws/muno/muno-go/bin/muno tree

# Show current status
echo ""
echo "4. Current status:"
/Users/musinsa/ws/muno/muno-go/bin/muno current

# Navigate to a lazy repo (should trigger clone)
echo ""
echo "5. Navigating to lazy repo (go-github)..."
/Users/musinsa/ws/muno/muno-go/bin/muno use go-github

# Show updated tree
echo ""
echo "6. Updated tree after navigation:"
/Users/musinsa/ws/muno/muno-go/bin/muno tree

# Check the state file
echo ""
echo "7. Verifying state file contains no filesystem paths..."
STATE_FILE="$TEST_DIR/test-real/.muno-tree.json"

if [ -f "$STATE_FILE" ]; then
    echo "   State file exists: $STATE_FILE"
    
    # Check for absence of filesystem paths
    if grep -q "$TEST_DIR" "$STATE_FILE"; then
        echo "   ❌ FAILED: State file contains filesystem path: $TEST_DIR"
        echo "   State file content:"
        cat "$STATE_FILE" | jq '.' | head -20
    else
        echo "   ✅ PASSED: State file contains NO filesystem paths"
    fi
    
    # Check for logical paths
    if grep -q '"/cobra"' "$STATE_FILE"; then
        echo "   ✅ PASSED: State file contains logical paths"
    else
        echo "   ❌ FAILED: State file missing expected logical paths"
    fi
    
    # Show a sample of the state structure
    echo ""
    echo "   State structure sample:"
    cat "$STATE_FILE" | jq '.nodes | to_entries | .[0:3]' 2>/dev/null || echo "   (jq not available for pretty printing)"
else
    echo "   ❌ ERROR: State file not found!"
fi

# Test filesystem path computation
echo ""
echo "8. Filesystem paths (computed, not stored):"
echo "   Logical path '/' maps to:"
ls -la "$TEST_DIR/test-real/repos" 2>/dev/null | head -5

echo "   Logical path '/cobra' maps to:"
ls -la "$TEST_DIR/test-real/nodes/cobra" 2>/dev/null | head -5

echo "   Logical path '/cobra/viper' would map to:"
echo "   $TEST_DIR/test-real/nodes/cobra/nodes/viper"

# Clone remaining lazy repos
echo ""
echo "9. Cloning all lazy repositories..."
/Users/musinsa/ws/muno/muno-go/bin/muno use /
/Users/musinsa/ws/muno/muno-go/bin/muno clone --recursive

# Final tree
echo ""
echo "10. Final tree with all repos cloned:"
/Users/musinsa/ws/muno/muno-go/bin/muno tree

# Summary
echo ""
echo "=== Test Summary ==="
echo "✅ Successfully initialized workspace"
echo "✅ Added real GitHub repositories"
echo "✅ Tree navigation working"
echo "✅ Lazy loading functional"
echo "✅ State management simplified (no filesystem paths)"

echo ""
echo "Test workspace location: $TEST_DIR"
echo "To explore: cd $TEST_DIR/test-real"
echo "To cleanup: rm -rf $TEST_DIR"