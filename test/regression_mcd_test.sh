#!/bin/bash

# Regression test for mcd functionality
# Tests path resolution and navigation with config reference nodes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Create temporary test directory
TEST_DIR=$(mktemp -d /tmp/muno-mcd-test-XXXXX)
# Get absolute path to muno binary
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
MUNO_BIN="${MUNO_BIN:-$SCRIPT_DIR/../bin/muno}"

echo "Testing mcd functionality with MUNO binary: $MUNO_BIN"
echo "Test directory: $TEST_DIR"

# Cleanup function
cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Test function
run_test() {
    local name="$1"
    local command="$2"
    local expected="$3"
    
    echo -n "Testing: $name ... "
    
    result=$(eval "$command" 2>&1) || true
    if [[ "$result" == *"$expected"* ]]; then
        echo -e "${GREEN}PASS${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected: $expected"
        echo "  Got: $result"
        ((TESTS_FAILED++))
    fi
}

# Initialize workspace
cd "$TEST_DIR"
$MUNO_BIN init test-workspace > /dev/null 2>&1

# Create main config
cat > muno.yaml << 'EOF'
workspace:
    name: test-workspace
    repos_dir: .nodes
nodes:
    - name: regular-repo
      url: https://github.com/example/regular
      fetch: lazy
    - name: config-ref
      file: ./external-config.yaml
EOF

# Create external config with custom repos_dir
cat > external-config.yaml << 'EOF'
workspace:
    name: external
    repos_dir: custom-repos
nodes:
    - name: child1
      url: https://github.com/example/child1
      fetch: lazy
    - name: child2
      url: https://github.com/example/child2
      fetch: lazy
EOF

echo ""
echo "=== Path Resolution Tests ==="

# Test 1: Root path from workspace
run_test "Root (/) from workspace root" \
    "$MUNO_BIN path /" \
    "$TEST_DIR/.nodes"

# Test 2: Current directory from workspace
run_test "Current (.) from workspace root" \
    "$MUNO_BIN path ." \
    "$TEST_DIR"

# Test 3: Regular repo path
run_test "Regular repo path" \
    "$MUNO_BIN path regular-repo" \
    "$TEST_DIR/.nodes/regular-repo"

# Test 4: Config ref node path
run_test "Config ref node path" \
    "$MUNO_BIN path config-ref" \
    "$TEST_DIR/.nodes/config-ref"

# Test 5: Child under config ref with custom repos_dir
run_test "Child under config ref (custom repos_dir)" \
    "$MUNO_BIN path config-ref/child1" \
    "$TEST_DIR/.nodes/config-ref/custom-repos/child1"

# Test 6: Second child under config ref
run_test "Second child under config ref" \
    "$MUNO_BIN path config-ref/child2" \
    "$TEST_DIR/.nodes/config-ref/custom-repos/child2"

# Navigate to nodes directory
cd .nodes
echo ""
echo "=== Path Resolution from .nodes Directory ==="

# Test 7: Root from nodes dir
run_test "Root (/) from nodes dir" \
    "$MUNO_BIN path /" \
    "$TEST_DIR/.nodes"

# Test 8: Current from nodes dir
run_test "Current (.) from nodes dir" \
    "$MUNO_BIN path ." \
    "$TEST_DIR/.nodes"

# Test 9: Relative path from nodes
run_test "Relative path to config-ref/child1 from nodes" \
    "$MUNO_BIN path config-ref/child1" \
    "$TEST_DIR/.nodes/config-ref/custom-repos/child1"

# Create config-ref directory to simulate navigation
mkdir -p config-ref
cd config-ref
echo ""
echo "=== Path Resolution from config-ref Directory ==="

# Test 10: Current from config-ref
run_test "Current (.) from config-ref" \
    "$MUNO_BIN path ." \
    "$TEST_DIR/.nodes/config-ref"

# Test 11: Root from nested location
run_test "Root (/) from config-ref" \
    "$MUNO_BIN path /" \
    "$TEST_DIR/.nodes"

# Test 12: Parent (..) from config-ref
run_test "Parent (..) from config-ref" \
    "$MUNO_BIN path .." \
    "$TEST_DIR/.nodes"

# Test 13: Child from config-ref
run_test "Child (child1) from config-ref" \
    "$MUNO_BIN path child1" \
    "$TEST_DIR/.nodes/config-ref/custom-repos/child1"

echo ""
echo "=== Testing --relative Flag ==="
cd "$TEST_DIR"

# Test 14: Relative path from workspace root
run_test "Relative path from workspace (should fail)" \
    "$MUNO_BIN path . --relative 2>&1 | grep -q 'not in repository tree' && echo 'not in repository tree'" \
    "not in repository tree"

cd .nodes

# Test 15: Relative path from nodes
run_test "Relative path from nodes" \
    "$MUNO_BIN path . --relative" \
    "/"

cd config-ref

# Test 16: Relative path from config-ref
run_test "Relative path from config-ref" \
    "$MUNO_BIN path . --relative" \
    "/config-ref"

echo ""
echo "=== Testing with Regular Repo and Git ==="
cd "$TEST_DIR/.nodes"

# Create a git repo with muno.yaml
mkdir -p git-repo
cd git-repo
git init > /dev/null 2>&1
cat > muno.yaml << 'EOF'
workspace:
    name: git-repo
    repos_dir: subrepos
nodes:
    - name: nested
      url: https://github.com/example/nested
      fetch: lazy
EOF
cd "$TEST_DIR"

# Update main config to include git repo
cat >> muno.yaml << 'EOF'
    - name: git-repo
      url: file://$TEST_DIR/.nodes/git-repo
      fetch: eager
EOF

# Test 17: Nested repo under git repo with custom repos_dir
run_test "Nested under git repo with custom repos_dir" \
    "$MUNO_BIN path git-repo/nested" \
    "$TEST_DIR/.nodes/git-repo/subrepos/nested"

echo ""
echo "=== Test Summary ==="
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi