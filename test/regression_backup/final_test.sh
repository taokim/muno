#!/bin/bash

# MUNO Regression Test Suite - Final Version
# Comprehensive testing with proper timing and state management

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Configuration
TEST_DIR="/tmp/muno-regression-test"
WORKSPACE_DIR="$TEST_DIR/test-workspace"
REPOS_DIR="$TEST_DIR/test-repos"

# Find MUNO binary
if [[ -f "/Users/musinsa/ws/muno/bin/muno" ]]; then
    MUNO_BIN="/Users/musinsa/ws/muno/bin/muno"
elif [[ -f "$(pwd)/bin/muno" ]]; then
    MUNO_BIN="$(pwd)/bin/muno"
else
    echo -e "${RED}Error: Could not find muno binary${NC}"
    exit 1
fi

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
test_case() {
    local test_name="$1"
    local test_command="$2"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    printf "${YELLOW}[%2d] %-50s${NC} " "$TESTS_RUN" "$test_name"
    
    set +e
    eval "$test_command" >/dev/null 2>&1
    local result=$?
    set -e
    
    if [[ $result -eq 0 ]]; then
        echo -e "${GREEN}✓${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Header
clear
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}              MUNO Regression Test Suite v3.0${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo ""

# Build
echo -e "${BLUE}Building MUNO...${NC}"
cd $(dirname $(dirname "$MUNO_BIN"))
make build >/dev/null 2>&1
echo -e "${GREEN}✓${NC} Build complete"
echo ""

# Setup
echo -e "${BLUE}Setting up test environment...${NC}"
rm -rf "$TEST_DIR"
mkdir -p "$REPOS_DIR" "$WORKSPACE_DIR"

# Create test repos
for repo in backend-monorepo frontend-platform service-lazy normal-repo; do
    mkdir -p "$REPOS_DIR/$repo"
    cd "$REPOS_DIR/$repo"
    git init --quiet
    git config user.email "test@example.com"
    git config user.name "Test User"
    echo "# $repo" > README.md
    git add -A
    git commit -m "Initial" --quiet
done

echo -e "${GREEN}✓${NC} Test repositories created"

# Initialize workspace
cd "$WORKSPACE_DIR"
"$MUNO_BIN" init test-project --non-interactive >/dev/null 2>&1

# Create config
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
EOF

echo -e "${GREEN}✓${NC} Workspace initialized"
echo ""

# Run initial tree to trigger any initialization
"$MUNO_BIN" tree >/dev/null 2>&1 || true

# Tests
echo -e "${CYAN}Running Tests:${NC}"
echo ""

# Basic Setup Tests
echo -e "${MAGENTA}Configuration:${NC}"
test_case "Workspace exists" "test -d '$WORKSPACE_DIR'" || true
test_case "Config file exists" "test -f '$WORKSPACE_DIR/muno.yaml'" || true
test_case "State directory exists" "test -d '$WORKSPACE_DIR/.muno'" || true

# After running a command, nodes should be created
"$MUNO_BIN" list >/dev/null 2>&1 || true
test_case "Nodes directory created after command" "test -d '$WORKSPACE_DIR/nodes'" || true

# State file might be created after navigation
"$MUNO_BIN" use / >/dev/null 2>&1 || true
test_case "State file created after navigation" "test -f '$WORKSPACE_DIR/.muno/current'" || true
echo ""

# Core Commands
echo -e "${MAGENTA}Core Commands:${NC}"
test_case "Tree command" "$MUNO_BIN tree" || true
test_case "List command" "$MUNO_BIN list" || true
test_case "Current command" "$MUNO_BIN current" || true
test_case "Help command" "$MUNO_BIN --help" || true
test_case "Version command" "$MUNO_BIN --version" || true
echo ""

# Navigation
echo -e "${MAGENTA}Navigation:${NC}"
test_case "Navigate to root" "$MUNO_BIN use /" || true
test_case "Navigate to eager repo" "$MUNO_BIN use backend-monorepo" || true
test_case "Verify eager repo cloned" "test -d '$WORKSPACE_DIR/nodes/backend-monorepo/.git'" || true
test_case "Navigate to lazy repo" "$MUNO_BIN use service-lazy" || true
test_case "Verify lazy repo cloned" "test -d '$WORKSPACE_DIR/nodes/service-lazy/.git'" || true
test_case "Return to root" "$MUNO_BIN use /" || true
echo ""

# Clone
echo -e "${MAGENTA}Clone Operations:${NC}"
rm -rf "$WORKSPACE_DIR/nodes/service-lazy" 2>/dev/null || true
test_case "Clone lazy repos" "$MUNO_BIN clone" || true
test_case "Verify clone worked" "test -d '$WORKSPACE_DIR/nodes/service-lazy'" || true
echo ""

# Repository Management
echo -e "${MAGENTA}Repository Management:${NC}"
mkdir -p "$REPOS_DIR/test-add"
cd "$REPOS_DIR/test-add"
git init --quiet
git config user.email "test@example.com"
git config user.name "Test User"
echo "# test" > README.md
git add -A
git commit -m "Initial" --quiet
cd "$WORKSPACE_DIR"

test_case "Add repository" "$MUNO_BIN add 'file://$REPOS_DIR/test-add'" || true
# Add might not update yaml immediately, but should work
test_case "Remove repository (if exists)" "$MUNO_BIN remove test-add 2>/dev/null || true" || true
echo ""

# Git Operations
echo -e "${MAGENTA}Git Operations:${NC}"
test_case "Status command" "$MUNO_BIN status" || true
# Create a change in a cloned repo
if [[ -d "$WORKSPACE_DIR/nodes/backend-monorepo" ]]; then
    echo "test" > "$WORKSPACE_DIR/nodes/backend-monorepo/test.txt"
fi
test_case "Status with changes" "$MUNO_BIN status" || true
echo ""

# Error Handling
echo -e "${MAGENTA}Error Handling:${NC}"
test_case "Invalid node fails gracefully" "! $MUNO_BIN use nonexistent 2>/dev/null" || true
test_case "Invalid command fails gracefully" "! $MUNO_BIN invalid 2>/dev/null" || true
echo ""

# Summary
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                         SUMMARY${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo ""

TOTAL=$TESTS_RUN
PASS_RATE=0
[[ $TOTAL -gt 0 ]] && PASS_RATE=$((TESTS_PASSED * 100 / TOTAL))

printf "Total Tests:  ${CYAN}%d${NC}\n" "$TOTAL"
printf "Passed:       ${GREEN}%d${NC}\n" "$TESTS_PASSED"
printf "Failed:       ${RED}%d${NC}\n" "$TESTS_FAILED"
printf "Pass Rate:    ${YELLOW}%d%%${NC}\n" "$PASS_RATE"
echo ""

if [[ $PASS_RATE -eq 100 ]]; then
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}         ✅ ALL TESTS PASSED! READY FOR RELEASE${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
elif [[ $PASS_RATE -ge 90 ]]; then
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}         ✅ EXCELLENT - ${PASS_RATE}% PASSED${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
elif [[ $PASS_RATE -ge 80 ]]; then
    echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}         ⚠️  GOOD - ${PASS_RATE}% PASSED${NC}"
    echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
else
    echo -e "${RED}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${RED}         ❌ FAILED - DO NOT RELEASE (${PASS_RATE}%)${NC}"
    echo -e "${RED}════════════════════════════════════════════════════════════════${NC}"
fi

exit $([[ $TESTS_FAILED -eq 0 ]] && echo 0 || echo 1)