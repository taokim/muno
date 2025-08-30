#!/bin/bash
# test-workflow.sh - Quick smoke test for muno

set -e

echo "🧪 Running muno smoke test..."

# Build
echo "📦 Building..."
cd "$(dirname "$0")/.."
make build >/dev/null 2>&1

# Create test directory
TEST_DIR=$(mktemp -d)
BINARY="$(pwd)/bin/muno"

# Test basic workflow
cd "$TEST_DIR"
echo "🚀 Testing init..."
"$BINARY" init test-project --non-interactive >/dev/null

echo "📊 Testing status..."
cd test-project
"$BINARY" status >/dev/null

# Quick validation
if [ -f "muno.yaml" ] && [ -d "workspace" ]; then
    echo "✅ Smoke test passed!"
    cd ../..
    rm -rf "$TEST_DIR"
    exit 0
else
    echo "❌ Smoke test failed!"
    exit 1
fi