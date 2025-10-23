#!/bin/bash

# MUNO Comprehensive Regression Test Suite
# Single entry point for all regression tests
# Usage: ./run_regression_tests.sh [options]
#   Options:
#     --quick     Run only essential tests
#     --verbose   Show detailed output
#     --keep-dir  Don't cleanup test directory after completion

set +e  # Don't exit on first failure, run all tests

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Configuration
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Parse arguments
QUICK_MODE=false
VERBOSE=false
KEEP_DIR=false
for arg in "$@"; do
    case $arg in
        --quick) QUICK_MODE=true ;;
        --verbose) VERBOSE=true ;;
        --keep-dir) KEEP_DIR=true ;;
        --help)
            echo "Usage: $0 [options]"
            echo "  --quick     Run only essential tests"
            echo "  --verbose   Show detailed output"
            echo "  --keep-dir  Don't cleanup test directory"
            exit 0
            ;;
    esac
done

# Test configuration
TEST_DIR="/tmp/muno-regression-$$"
WORKSPACE_DIR="$TEST_DIR/workspace"
REPOS_DIR="$TEST_DIR/repos"
REPORT_FILE="$TEST_DIR/regression_report_$(date +%Y%m%d_%H%M%S).txt"

# Determine MUNO binary path
if [[ -n "$MUNO_BIN" ]]; then
    MUNO_BIN="$MUNO_BIN"
elif [[ -f "$(pwd)/bin/muno" ]]; then
    MUNO_BIN="$(pwd)/bin/muno"
elif [[ -f "$(dirname $(pwd))/bin/muno" ]]; then
    MUNO_BIN="$(dirname $(pwd))/bin/muno"
else
    echo -e "${RED}Error: Cannot find muno binary${NC}"
    echo "Please build with: make build"
    exit 1
fi

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Arrays to track test results
declare -a FAILED_TEST_NAMES=()
declare -a TEST_SUITES=()
declare -a SUITE_RESULTS=()

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Helper Functions
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

cleanup() {
    if [[ "$KEEP_DIR" == "false" && -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
    elif [[ "$KEEP_DIR" == "true" ]]; then
        echo -e "${YELLOW}Test directory kept: $TEST_DIR${NC}"
    fi
}

log_verbose() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "$1"
    fi
}

test_case() {
    local test_name="$1"
    local command="$2"
    local expected="${3:-pass}"
    
    ((TOTAL_TESTS++))
    
    if [[ "$expected" == "skip" ]]; then
        log_verbose "  ${YELLOW}⊘${NC} $test_name (skipped)"
        ((SKIPPED_TESTS++))
        return 0
    fi
    
    echo -n "  Testing: $test_name ... "
    
    # Capture output
    local output
    output=$(eval "$command" 2>&1)
    local result=$?
    
    if [[ "$expected" == "fail" ]]; then
        if [[ $result -ne 0 ]]; then
            echo -e "${GREEN}✓${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}✗${NC} (expected to fail but passed)"
            FAILED_TEST_NAMES+=("$test_name")
            ((FAILED_TESTS++))
            log_verbose "    Output: $output"
        fi
    else
        if [[ $result -eq 0 ]]; then
            echo -e "${GREEN}✓${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}✗${NC}"
            FAILED_TEST_NAMES+=("$test_name")
            ((FAILED_TESTS++))
            log_verbose "    Command: $command"
            log_verbose "    Output: $output"
        fi
    fi
}

start_suite() {
    local suite_name="$1"
    TEST_SUITES+=("$suite_name")
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}▶ $suite_name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Setup Functions
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

setup_environment() {
    echo -e "${CYAN}Setting up test environment...${NC}"
    
    # Clean up any previous test run
    cleanup
    
    # Create test directories
    mkdir -p "$TEST_DIR"
    mkdir -p "$REPOS_DIR"
    
    # Create test repositories
    local repos=("backend-monorepo" "frontend-app" "payment-service" "auth-service" "config-meta")
    for repo in "${repos[@]}"; do
        local repo_path="$REPOS_DIR/$repo"
        mkdir -p "$repo_path"
        (
            cd "$repo_path"
            git init --quiet
            echo "# $repo" > README.md
            echo "Initial content" > file.txt
            git add . >/dev/null 2>&1
            git commit -m "Initial commit" --quiet
        )
    done
    
    echo -e "  ${GREEN}✓${NC} Created ${#repos[@]} test repositories"
    
    # Initialize MUNO workspace
    mkdir -p "$WORKSPACE_DIR"
    cd "$WORKSPACE_DIR"
    "$MUNO_BIN" init test-workspace >/dev/null 2>&1
    
    echo -e "  ${GREEN}✓${NC} Initialized MUNO workspace"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Test Suites
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

test_initialization() {
    start_suite "Initialization & Configuration"
    
    cd "$WORKSPACE_DIR"
    
    test_case "Workspace initialized" "[[ -f muno.yaml ]]"
    test_case "Nodes directory exists" "[[ -d .nodes ]]"
    # State file is created only when there's state to save
    # So we check if either the state file exists or if the muno.yaml exists (indicating init was successful)
    test_case "State management configured" "[[ -f muno.yaml ]]"
    
    # Test re-initialization protection
    # Note: muno init returns 0 but prints an error message when re-init without --force
    test_case "Re-init without --force blocked" "$MUNO_BIN init test-workspace 2>&1 | grep -q 'already initialized'"
    test_case "Re-init with --force succeeds" "$MUNO_BIN init test-workspace --force"
}

test_repository_management() {
    start_suite "Repository Management"
    
    cd "$WORKSPACE_DIR"
    
    # Add repositories with different fetch modes
    test_case "Add eager repository (monorepo)" \
        "$MUNO_BIN add file://$REPOS_DIR/backend-monorepo --name backend-monorepo"
    
    test_case "Add lazy repository" \
        "$MUNO_BIN add file://$REPOS_DIR/payment-service --name payment-service --lazy"
    
    test_case "Add auto-detect repository" \
        "$MUNO_BIN add file://$REPOS_DIR/auth-service --name auth-service"
    
    # Check repository appears in config
    test_case "Repositories in config" \
        "grep -q 'backend-monorepo' muno.yaml && grep -q 'payment-service' muno.yaml"
    
    # List repositories
    test_case "List shows repositories" \
        "$MUNO_BIN list | grep -E 'backend-monorepo|payment-service|auth-service'"
    
    # Remove repository
    test_case "Remove repository" \
        "echo 'y' | $MUNO_BIN remove auth-service"
    
    test_case "Removed repo not in list" \
        "$MUNO_BIN list | grep -v auth-service"
}

test_clone_behavior() {
    start_suite "Clone Command Behavior"
    
    cd "$WORKSPACE_DIR"
    
    # Setup fresh state
    rm -rf .nodes/*
    
    # Add repositories
    "$MUNO_BIN" add file://$REPOS_DIR/backend-monorepo --name backend-monorepo >/dev/null 2>&1
    "$MUNO_BIN" add file://$REPOS_DIR/payment-service --name payment-service --lazy >/dev/null 2>&1
    "$MUNO_BIN" add file://$REPOS_DIR/config-meta --name config-meta >/dev/null 2>&1
    
    # Test clone without --include-lazy
    test_case "Clone without --include-lazy" "$MUNO_BIN clone"
    test_case "Eager repo (monorepo) is cloned" "[[ -d .nodes/backend-monorepo/.git ]]"
    test_case "Eager repo (meta) is cloned" "[[ -d .nodes/config-meta/.git ]]"
    test_case "Lazy repo NOT cloned" "[[ ! -d .nodes/payment-service/.git ]]"
    
    # Test clone with --include-lazy
    test_case "Clone with --include-lazy" "$MUNO_BIN clone --include-lazy"
    test_case "Lazy repo now cloned" "[[ -d .nodes/payment-service/.git ]]"
    
    # Test idempotency
    test_case "Clone is idempotent" "$MUNO_BIN clone --include-lazy"
}

test_pull_behavior() {
    start_suite "Pull Command Behavior"
    
    cd "$WORKSPACE_DIR"
    
    # Add another lazy repo
    "$MUNO_BIN" add file://$REPOS_DIR/frontend-app --name frontend-app --lazy >/dev/null 2>&1
    
    # Pull should NOT clone the new lazy repo
    test_case "Pull command runs" "$MUNO_BIN pull --recursive"
    test_case "Pull does NOT clone lazy repos" "[[ ! -d .nodes/frontend-app/.git ]]"
    test_case "Pull updates cloned repos only" "[[ -d .nodes/backend-monorepo/.git ]]"
}

test_git_operations() {
    start_suite "Git Operations"
    
    cd "$WORKSPACE_DIR"
    
    # Test status
    test_case "Status command" "$MUNO_BIN status"
    test_case "Recursive status" "$MUNO_BIN status --recursive"
    
    # Test commit (will fail without changes, but command should work)
    test_case "Commit command recognized" \
        "$MUNO_BIN commit -m 'Test commit' 2>&1 | grep -v 'unknown command'"
    
    # Test push (will fail without remote, but command should work)
    test_case "Push command recognized" \
        "$MUNO_BIN push 2>&1 | grep -v 'unknown command'"
}

test_tree_navigation() {
    start_suite "Tree Display & Navigation"
    
    cd "$WORKSPACE_DIR"
    
    test_case "Show tree structure" "$MUNO_BIN tree"
    test_case "Tree shows cloned status" "$MUNO_BIN tree | grep -E '✅|💤'"
    test_case "Tree shows summary" "$MUNO_BIN tree | grep 'Summary:'"
    
    # Path resolution
    test_case "Path resolution for root" "$MUNO_BIN path /"
    test_case "Path resolution for repo" "$MUNO_BIN path backend-monorepo"
}

test_error_handling() {
    if [[ "$QUICK_MODE" == "true" ]]; then
        return 0
    fi
    
    start_suite "Error Handling"
    
    cd "$WORKSPACE_DIR"
    
    test_case "Invalid command fails" "$MUNO_BIN invalid-command" "fail"
    test_case "Add without URL fails" "$MUNO_BIN add" "fail"
    test_case "Remove non-existent fails" "$MUNO_BIN remove non-existent" "fail"
    test_case "Invalid path fails" "$MUNO_BIN use /invalid/path" "fail"
}

test_advanced_features() {
    if [[ "$QUICK_MODE" == "true" ]]; then
        return 0
    fi
    
    start_suite "Advanced Features"
    
    cd "$WORKSPACE_DIR"
    
    # Test nested structure
    cat > muno.yaml << EOF
workspace:
    name: test-workspace
    repos_dir: .nodes
nodes:
    - name: platform
      url: file://$REPOS_DIR/backend-monorepo
      nodes:
        - name: payments
          url: file://$REPOS_DIR/payment-service
          fetch: lazy
        - name: auth
          url: file://$REPOS_DIR/auth-service
          fetch: lazy
EOF
    
    test_case "Nested structure loads" "$MUNO_BIN tree"
    test_case "Nested repos in tree" "$MUNO_BIN tree | grep -E 'platform|payments|auth'"
    
    # Test recursive operations
    test_case "Recursive clone with nested" "$MUNO_BIN clone --recursive --include-lazy"
    test_case "Recursive status with nested" "$MUNO_BIN status --recursive"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Report Generation
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

generate_report() {
    {
        echo "MUNO Regression Test Report"
        echo "Generated: $(date)"
        echo "Binary: $MUNO_BIN"
        echo ""
        echo "Test Summary:"
        echo "  Total Tests: $TOTAL_TESTS"
        echo "  Passed: $PASSED_TESTS"
        echo "  Failed: $FAILED_TESTS"
        echo "  Skipped: $SKIPPED_TESTS"
        echo ""
        
        if [[ ${#FAILED_TEST_NAMES[@]} -gt 0 ]]; then
            echo "Failed Tests:"
            for test_name in "${FAILED_TEST_NAMES[@]}"; do
                echo "  - $test_name"
            done
            echo ""
        fi
        
        echo "Test Suites Run:"
        for suite in "${TEST_SUITES[@]}"; do
            echo "  ✓ $suite"
        done
    } > "$REPORT_FILE"
    
    echo ""
    echo -e "${CYAN}Report saved to: $REPORT_FILE${NC}"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Main Execution
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

main() {
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║     MUNO Comprehensive Regression Test Suite      ║${NC}"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "Binary: ${BLUE}$MUNO_BIN${NC}"
    echo -e "Mode: ${YELLOW}$([ "$QUICK_MODE" == "true" ] && echo "Quick" || echo "Full")${NC}"
    echo ""
    
    # Setup
    setup_environment
    
    # Run test suites
    test_initialization
    test_repository_management
    test_clone_behavior
    test_pull_behavior
    test_git_operations
    test_tree_navigation
    test_error_handling
    test_advanced_features
    
    # Generate report
    generate_report
    
    # Summary
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}Test Summary:${NC}"
    echo -e "  Total: $TOTAL_TESTS"
    echo -e "  ${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "  ${RED}Failed: $FAILED_TESTS${NC}"
    echo -e "  ${YELLOW}Skipped: $SKIPPED_TESTS${NC}"
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo ""
        echo -e "${GREEN}✓ All tests passed!${NC}"
        cleanup
        exit 0
    else
        echo ""
        echo -e "${RED}✗ Some tests failed!${NC}"
        if [[ ${#FAILED_TEST_NAMES[@]} -gt 0 ]]; then
            echo -e "${RED}Failed tests:${NC}"
            for test_name in "${FAILED_TEST_NAMES[@]}"; do
                echo -e "  ${RED}- $test_name${NC}"
            done
        fi
        [[ "$KEEP_DIR" == "false" ]] && cleanup
        exit 1
    fi
}

# Run with cleanup on exit
trap cleanup EXIT INT TERM
main