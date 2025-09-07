#!/bin/bash

# MUNO Comprehensive Regression Test
# Tests all major features in an integrated manner

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
TEST_DIR="/tmp/muno-regression-test"
WORKSPACE_DIR="$TEST_DIR/test-workspace"
REPOS_DIR="$TEST_DIR/test-repos"

# Find MUNO binary - try multiple locations
if [[ -f "/Users/musinsa/ws/muno/bin/muno" ]]; then
    MUNO_BIN="/Users/musinsa/ws/muno/bin/muno"
elif [[ -f "$(pwd)/bin/muno" ]]; then
    MUNO_BIN="$(pwd)/bin/muno"
elif [[ -f "../../bin/muno" ]]; then
    MUNO_BIN="$(cd ../.. && pwd)/bin/muno"
else
    echo -e "${RED}Error: Could not find muno binary${NC}"
    echo "Please build muno first: make build"
    exit 1
fi

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test result tracking
declare -a FAILED_TESTS

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_pattern="$3"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -ne "${YELLOW}[${TESTS_RUN}] Testing ${test_name}...${NC} "
    
    # Run command and capture output
    set +e
    OUTPUT=$($test_command 2>&1)
    EXIT_CODE=$?
    set -e
    
    # Check if pattern exists in output or exit code is 0
    if [[ -n "$expected_pattern" ]]; then
        if echo "$OUTPUT" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}PASSED${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        fi
    elif [[ $EXIT_CODE -eq 0 ]]; then
        echo -e "${GREEN}PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    fi
    
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    FAILED_TESTS+=("$test_name")
    echo "  Command: $test_command"
    echo "  Exit Code: $EXIT_CODE"
    echo "  Output: $(echo "$OUTPUT" | head -3)"
    return 1
}

# Function to verify file exists
verify_exists() {
    local path="$1"
    local description="$2"
    
    if [[ -e "$path" ]]; then
        echo -e "  ${GREEN}✓${NC} $description exists"
        return 0
    else
        echo -e "  ${RED}✗${NC} $description missing"
        return 1
    fi
}

# Header
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}        MUNO Comprehensive Regression Test${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo ""

# Step 1: Build MUNO
echo -e "${BLUE}Step 1: Building MUNO...${NC}"
MUNO_DIR=$(dirname $(dirname "$MUNO_BIN"))
cd "$MUNO_DIR"
if make build > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Build successful${NC}"
    echo "  Binary: $MUNO_BIN"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Step 2: Setup test environment
echo -e "${BLUE}Step 2: Setting up test environment...${NC}"

# Clean and recreate
rm -rf "$TEST_DIR"
mkdir -p "$REPOS_DIR"
mkdir -p "$WORKSPACE_DIR"

# Create test repositories
echo "Creating test repositories..."
for repo in backend-monorepo frontend-platform service-lazy normal-repo team-config; do
    mkdir -p "$REPOS_DIR/$repo"
    cd "$REPOS_DIR/$repo"
    git init --quiet
    git config user.email "test@example.com"
    git config user.name "Test User"
    echo "# $repo" > README.md
    
    # Add special files for certain repos
    if [[ "$repo" == *"monorepo"* ]] || [[ "$repo" == *"platform"* ]]; then
        echo "eager" > type.txt
    fi
    
    if [[ "$repo" == "team-config" ]]; then
        cat > muno.yaml << EOF
workspace:
  name: team-config
  repos_dir: nodes
nodes:
  - name: child1
    url: file://$REPOS_DIR/child1
    lazy: true
EOF
    fi
    
    git add -A
    git commit -m "Initial commit" --quiet
done

# Create child repos for config node
for repo in child1 child2; do
    mkdir -p "$REPOS_DIR/$repo"
    cd "$REPOS_DIR/$repo"
    git init --quiet
    git config user.email "test@example.com"
    git config user.name "Test User"
    echo "# $repo" > README.md
    git add -A
    git commit -m "Initial commit" --quiet
done

echo -e "${GREEN}✓ Test repositories created${NC}"

# Step 3: Initialize workspace
echo -e "${BLUE}Step 3: Initializing MUNO workspace...${NC}"
cd "$WORKSPACE_DIR"

run_test "muno init" "$MUNO_BIN init test-project --non-interactive" "" || true

# Create comprehensive muno.yaml
cat > muno.yaml << EOF
workspace:
  name: test-project
  repos_dir: nodes

nodes:
  - name: backend-monorepo
    url: file://$REPOS_DIR/backend-monorepo
    
  - name: frontend-platform
    url: file://$REPOS_DIR/frontend-platform
    
  - name: service-lazy
    url: file://$REPOS_DIR/service-lazy
    lazy: true
    
  - name: normal-repo
    url: file://$REPOS_DIR/normal-repo
    lazy: false
    
  - name: team-config
    url: file://$REPOS_DIR/team-config
EOF

echo -e "${GREEN}✓ Workspace initialized${NC}"

# Step 4: Test commands
echo ""
echo -e "${BLUE}Step 4: Testing MUNO commands...${NC}"
echo ""

# Test tree command
echo -e "${CYAN}Testing 'tree' command...${NC}"
run_test "tree: works without error" "$MUNO_BIN tree" "" || true
run_test "tree: shows workspace name" "$MUNO_BIN tree 2>&1" "test-project\|Repository Tree" || true

# Test list command
echo -e "${CYAN}Testing 'list' command...${NC}"
run_test "list: shows children" "$MUNO_BIN list" "backend-monorepo\|frontend-platform" || true

# Test current command
echo -e "${CYAN}Testing 'current' command...${NC}"
run_test "current: shows position" "$MUNO_BIN current" "" || true

# Test use command
echo -e "${CYAN}Testing 'use' command...${NC}"
run_test "use: navigate to repo" "$MUNO_BIN use backend-monorepo" "" || true
verify_exists "$WORKSPACE_DIR/nodes/backend-monorepo" "backend-monorepo directory" || true

run_test "use: navigate to root" "$MUNO_BIN use /" "" || true

# Test clone command
echo -e "${CYAN}Testing 'clone' command...${NC}"
run_test "clone: lazy repository" "$MUNO_BIN clone service-lazy" "" || true
verify_exists "$WORKSPACE_DIR/nodes/service-lazy" "service-lazy directory" || true

# Test status command
echo -e "${CYAN}Testing 'status' command...${NC}"

# Create some changes
if [[ -d "$WORKSPACE_DIR/nodes/normal-repo" ]]; then
    cd "$WORKSPACE_DIR/nodes/normal-repo"
    echo "test" > test.txt
    cd "$WORKSPACE_DIR"
fi

run_test "status: from root" "$MUNO_BIN status" "" || true

# Test pull command
echo -e "${CYAN}Testing 'pull' command...${NC}"

# Add commits to source repos
cd "$REPOS_DIR/backend-monorepo"
echo "update" >> README.md
git add README.md
git commit -m "Update" --quiet

cd "$WORKSPACE_DIR"
run_test "pull: from root" "$MUNO_BIN pull" "" || true

# Test add/remove commands
echo -e "${CYAN}Testing 'add/remove' commands...${NC}"

# Create new test repo
mkdir -p "$REPOS_DIR/new-repo"
cd "$REPOS_DIR/new-repo"
git init --quiet
git config user.email "test@example.com"
git config user.name "Test User"
echo "# new-repo" > README.md
git add -A
git commit -m "Initial" --quiet

cd "$WORKSPACE_DIR"
run_test "add: new repository" "$MUNO_BIN add file://$REPOS_DIR/new-repo" "" || true
run_test "remove: repository" "$MUNO_BIN remove new-repo" "" || true

# Test special node types
echo -e "${CYAN}Testing special node types...${NC}"

# Test eager loading (monorepo/platform should auto-clone)
run_test "eager: navigate to monorepo" "$MUNO_BIN use backend-monorepo" "" || true
verify_exists "$WORKSPACE_DIR/nodes/backend-monorepo" "eager monorepo" || true

# Test config node
run_test "config: navigate to config node" "$MUNO_BIN use team-config" "" || true
verify_exists "$WORKSPACE_DIR/nodes/team-config" "config node" || true

# Test error handling
echo -e "${CYAN}Testing error handling...${NC}"
set +e
run_test "error: non-existent repo" "$MUNO_BIN use non-existent 2>&1" "not found\|does not exist\|error\|Error" || true
set -e

# Step 5: Verify state persistence
echo ""
echo -e "${BLUE}Step 5: Verifying state persistence...${NC}"

verify_exists "$WORKSPACE_DIR/muno.yaml" "Configuration file" || true
verify_exists "$WORKSPACE_DIR/.muno-state.json" "State file" || true

# Summary
echo ""
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                    TEST SUMMARY${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "Tests Run:    ${BLUE}${TESTS_RUN}${NC}"
echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

# Calculate pass rate
if [[ $TESTS_RUN -gt 0 ]]; then
    PASS_RATE=$((TESTS_PASSED * 100 / TESTS_RUN))
    echo -e "Pass Rate:    ${BLUE}${PASS_RATE}%${NC}"
    
    if [[ $PASS_RATE -eq 100 ]]; then
        echo ""
        echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}         ✅ ALL TESTS PASSED! READY FOR RELEASE${NC}"
        echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
    elif [[ $PASS_RATE -ge 80 ]]; then
        echo ""
        echo -e "${YELLOW}════════════════════════════════════════════════════════${NC}"
        echo -e "${YELLOW}         ⚠️  MOSTLY PASSED - REVIEW FAILURES${NC}"
        echo -e "${YELLOW}════════════════════════════════════════════════════════${NC}"
    else
        echo ""
        echo -e "${RED}════════════════════════════════════════════════════════${NC}"
        echo -e "${RED}         ❌ TESTS FAILED - DO NOT RELEASE${NC}"
        echo -e "${RED}════════════════════════════════════════════════════════${NC}"
    fi
fi

# List failed tests
if [[ ${#FAILED_TESTS[@]} -gt 0 ]]; then
    echo ""
    echo -e "${RED}Failed Tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗${NC} $test"
    done
fi

echo ""

# Write results to file
RESULTS_FILE="$TEST_DIR/test_results.txt"
cat > "$RESULTS_FILE" << EOF
MUNO Regression Test Results
============================
Date: $(date)
Binary: $MUNO_BIN

Results:
--------
Tests Run: $TESTS_RUN
Tests Passed: $TESTS_PASSED
Tests Failed: $TESTS_FAILED
Pass Rate: ${PASS_RATE}%

Failed Tests:
EOF

if [[ ${#FAILED_TESTS[@]} -gt 0 ]]; then
    for test in "${FAILED_TESTS[@]}"; do
        echo "- $test" >> "$RESULTS_FILE"
    done
else
    echo "None" >> "$RESULTS_FILE"
fi

echo -e "${CYAN}Results saved to: $RESULTS_FILE${NC}"

# Exit with appropriate code
if [[ $TESTS_FAILED -eq 0 ]]; then
    exit 0
else
    exit 1
fi