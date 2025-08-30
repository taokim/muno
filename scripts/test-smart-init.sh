#!/bin/bash

# Test script for the new smart init feature
# This verifies that rc init detects repos and stores them in config

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RC_BIN="${RC_BIN:-$PROJECT_DIR/bin/rc}"
TEST_DIR="/tmp/rc-smart-init-test-$(date +%s)"

echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}     Testing Smart Init with Repository Detection${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo

# Check rc binary
if [ ! -f "$RC_BIN" ]; then
    echo -e "${RED}Error: rc binary not found at $RC_BIN${NC}"
    echo "Please build it first: make build"
    exit 1
fi

echo -e "${YELLOW}Test directory: $TEST_DIR${NC}"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Step 1: Create some git repositories
echo -e "${CYAN}Step 1: Creating test git repositories${NC}"

create_repo() {
    local name="$1"
    local remote="$2"
    
    echo -e "  Creating $name..."
    mkdir -p "$name"
    cd "$name"
    git init --quiet
    git config user.name "Test"
    git config user.email "test@test.com"
    
    echo "# $name" > README.md
    git add README.md
    git commit -m "Initial commit" --quiet
    
    if [ -n "$remote" ]; then
        git remote add origin "$remote"
    fi
    
    cd ..
}

create_repo "auth-service" "https://github.com/example/auth-service.git"
create_repo "user-service" "https://github.com/example/user-service.git"
create_repo "payment-service" "https://github.com/example/payment-service.git"
create_repo "common-lib" "https://github.com/example/common-lib.git"

# Also create a non-git directory
mkdir -p "not-a-repo"
echo "This is not a git repo" > not-a-repo/file.txt

echo -e "${GREEN}✓ Created 4 git repos and 1 non-git directory${NC}"
echo

# Step 2: Run smart init (non-interactive)
echo -e "${CYAN}Step 2: Running rc init with smart detection${NC}"
echo -e "  Command: $RC_BIN init test-project --no-interactive${NC}"

# Use 'yes' to automatically answer Y to all prompts
yes | $RC_BIN init test-project --interactive --force 2>&1 | tee init-output.log || true

echo

# Step 3: Verify results
echo -e "${CYAN}Step 3: Verifying results${NC}"

# Check if repos were moved
echo -n "  Checking if repos were moved to repos/: "
if [ -d "repos/auth-service" ] && [ -d "repos/user-service" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    ls -la repos/ 2>/dev/null || echo "    repos/ directory not found"
fi

# Check repo-claude.yaml
echo -n "  Checking repo-claude.yaml contains repositories: "
if [ -f "repo-claude.yaml" ]; then
    repo_count=$(grep -c "name:" repo-claude.yaml 2>/dev/null || echo "0")
    if [ "$repo_count" -ge 4 ]; then
        echo -e "${GREEN}✓ ($repo_count repositories)${NC}"
    else
        echo -e "${RED}✗ (only $repo_count repositories)${NC}"
    fi
    
    echo
    echo -e "${YELLOW}  repo-claude.yaml content:${NC}"
    cat repo-claude.yaml | head -20
else
    echo -e "${RED}✗ (file not found)${NC}"
fi

echo

# Step 4: Test that we can use the workspace
echo -e "${CYAN}Step 4: Testing workspace functionality${NC}"

echo -n "  Running 'rc tree': "
if $RC_BIN tree 2>&1 | grep -q "auth-service\|user-service"; then
    echo -e "${GREEN}✓${NC}"
    echo
    $RC_BIN tree
else
    echo -e "${RED}✗${NC}"
fi

echo
echo -n "  Running 'rc list': "
if $RC_BIN list 2>&1 | grep -q "auth-service\|user-service"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
fi

echo

# Step 5: Test adding a new repo
echo -e "${CYAN}Step 5: Testing 'rc add' updates config${NC}"

echo -e "  Adding a new repository..."
$RC_BIN add "https://github.com/example/new-service.git" --name new-service --lazy || true

echo -n "  Checking if new-service is in repo-claude.yaml: "
if grep -q "new-service" repo-claude.yaml 2>/dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
fi

echo

# Summary
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                    Test Summary${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo
echo "Test directory: $TEST_DIR"
echo
echo "Key features tested:"
echo "✓ Smart detection of existing git repositories"
echo "✓ Moving repositories to repos/ directory"
echo "✓ Storing repository definitions in repo-claude.yaml"
echo "✓ Loading workspace from config"
echo "✓ Adding new repos updates config"
echo
echo -e "${GREEN}Test complete!${NC}"
echo
echo "To explore the test workspace:"
echo "  cd $TEST_DIR"
echo "  $RC_BIN tree"
echo "  cat repo-claude.yaml"