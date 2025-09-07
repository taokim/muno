#!/bin/bash

# Install regression test suite
# This script copies the regression test suite to /tmp for execution

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="/tmp/muno-regression-test"

echo "Installing MUNO regression test suite to $TEST_DIR..."

# Clean and create directory
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Copy test files
cp "$SCRIPT_DIR"/*.sh "$TEST_DIR/" 2>/dev/null || true
cp "$SCRIPT_DIR"/*.md "$TEST_DIR/" 2>/dev/null || true

# Make scripts executable
chmod +x "$TEST_DIR"/*.sh

echo "Installation complete!"
echo ""
echo "To run the regression tests:"
echo "  cd $TEST_DIR"
echo "  ./regression_test.sh"
echo ""
echo "For more information, see:"
echo "  $TEST_DIR/README.md"