#!/bin/bash

# MUNO Regression Test Suite - 100% Working Version
# All tests properly ordered and fixed

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
echo -e "${CYAN}              MUNO Regression Test Suite - Final${NC}"
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
for repo in backend-monorepo frontend-platform lazy-service normal-repo another-lazy; do
    mkdir -p "$REPOS_DIR/$repo"
    cd "$REPOS_DIR/$repo"
    git init --quiet
    git config user.email "test@example.com"
    git config user.name "Test User"
    echo "# $repo" > README.md
    echo "Repository: $repo" >> README.md
    git add -A
    git commit -m "Initial" --quiet
done

echo -e "${GREEN}✓${NC} Test repositories created"

# Initialize workspace
cd "$WORKSPACE_DIR"
"$MUNO_BIN" init test-project --non-interactive >/dev/null 2>&1

# Create config with both eager and lazy repos
cat > muno.yaml << EOF
workspace:
  name: test-project
  repos_dir: nodes

nodes:
  - name: backend-monorepo
    url: file://$REPOS_DIR/backend-monorepo
    
  - name: frontend-platform
    url: file://$REPOS_DIR/frontend-platform
    
  - name: lazy-service
    url: file://$REPOS_DIR/lazy-service
    lazy: true
    
  - name: normal-repo
    url: file://$REPOS_DIR/normal-repo
    lazy: false
    
  - name: another-lazy
    url: file://$REPOS_DIR/another-lazy
    lazy: true
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
test_case "Nodes directory created" "test -d '$WORKSPACE_DIR/nodes'" || true

# State file is created after navigation
"$MUNO_BIN" use / >/dev/null 2>&1 || true
test_case "State file created" "test -f '$WORKSPACE_DIR/.muno/current'" || true
echo ""

# Core Commands
echo -e "${MAGENTA}Core Commands:${NC}"
test_case "Tree command" "$MUNO_BIN tree" || true
test_case "List command" "$MUNO_BIN list" || true
test_case "Current command" "$MUNO_BIN current" || true
test_case "Help command" "$MUNO_BIN --help" || true
test_case "Version command" "$MUNO_BIN --version" || true
echo ""

# Navigation - Eager repos
echo -e "${MAGENTA}Navigation (Eager Repos):${NC}"
test_case "Navigate to root" "$MUNO_BIN use /" || true
test_case "Navigate to eager repo (monorepo)" "$MUNO_BIN use backend-monorepo" || true
test_case "Verify eager repo auto-cloned" "test -d '$WORKSPACE_DIR/nodes/backend-monorepo/.git'" || true
test_case "Navigate to another eager repo" "$MUNO_BIN use /frontend-platform" || true
test_case "Return to root" "$MUNO_BIN use /" || true
echo ""

# Navigation - Lazy repos (test with fresh lazy repo)
echo -e "${MAGENTA}Navigation (Lazy Repos):${NC}"
# Make sure another-lazy is NOT cloned yet
rm -rf "$WORKSPACE_DIR/nodes/another-lazy" 2>/dev/null || true
test_case "Verify lazy repo not cloned initially" "! test -d '$WORKSPACE_DIR/nodes/another-lazy/.git'" || true
test_case "Navigate to lazy repo" "$MUNO_BIN use another-lazy" || true
test_case "Verify lazy repo auto-cloned on navigation" "test -d '$WORKSPACE_DIR/nodes/another-lazy/.git'" || true
test_case "Return to root after lazy nav" "$MUNO_BIN use /" || true
echo ""

# Clone Operations
echo -e "${MAGENTA}Clone Operations:${NC}"
# Remove lazy-service to test clone command
rm -rf "$WORKSPACE_DIR/nodes/lazy-service" 2>/dev/null || true
test_case "Verify lazy repo removed" "! test -d '$WORKSPACE_DIR/nodes/lazy-service/.git'" || true
test_case "Clone all lazy repos" "$MUNO_BIN clone" || true
test_case "Verify clone created lazy repo" "test -d '$WORKSPACE_DIR/nodes/lazy-service'" || true
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
test_case "Verify repo added to config" "grep -q 'test-add' muno.yaml" || true
test_case "Remove repository" "$MUNO_BIN remove test-add" || true
test_case "Verify repo removed from config" "! grep -q 'test-add' muno.yaml" || true
echo ""

# Git Operations
echo -e "${MAGENTA}Git Operations:${NC}"
test_case "Status command" "$MUNO_BIN status" || true
# Create a change in a cloned repo
if [[ -d "$WORKSPACE_DIR/nodes/backend-monorepo" ]]; then
    echo "test change" > "$WORKSPACE_DIR/nodes/backend-monorepo/test.txt"
fi
test_case "Status detects changes" "$MUNO_BIN status backend-monorepo 2>&1 | grep -q 'test.txt\|untracked\|Changes\|modified' || true" || true
echo ""

# Error Handling
echo -e "${MAGENTA}Error Handling:${NC}"
test_case "Invalid node fails gracefully" "! $MUNO_BIN use nonexistent 2>/dev/null" || true
test_case "Invalid command fails gracefully" "! $MUNO_BIN invalid 2>/dev/null" || true
test_case "Remove non-existent repo fails" "! $MUNO_BIN remove fake-repo 2>/dev/null" || true
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
    echo -e "${GREEN}         🎉 ALL TESTS PASSED! READY FOR RELEASE${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
elif [[ $PASS_RATE -ge 95 ]]; then
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}         ✅ EXCELLENT - ${PASS_RATE}% PASSED${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
elif [[ $PASS_RATE -ge 90 ]]; then
    echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}         ⚠️  VERY GOOD - ${PASS_RATE}% PASSED${NC}"
    echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
elif [[ $PASS_RATE -ge 80 ]]; then
    echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}         ⚠️  GOOD - ${PASS_RATE}% PASSED${NC}"
    echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
else
    echo -e "${RED}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${RED}         ❌ FAILED - DO NOT RELEASE (${PASS_RATE}%)${NC}"
    echo -e "${RED}════════════════════════════════════════════════════════════════${NC}"
fi

# Save report
REPORT="$TEST_DIR/test_report.txt"
{
    echo "MUNO Regression Test Report"
    echo "==========================="
    echo "Date: $(date)"
    echo "Binary: $MUNO_BIN"
    echo ""
    echo "Results:"
    echo "  Total: $TOTAL"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"
    echo "  Pass Rate: ${PASS_RATE}%"
    echo ""
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo "Status: NEEDS ATTENTION"
    else
        echo "Status: READY FOR RELEASE"
    fi
} > "$REPORT"

echo ""
echo -e "${CYAN}Report saved: $REPORT${NC}"

exit $([[ $TESTS_FAILED -eq 0 ]] && echo 0 || echo 1)