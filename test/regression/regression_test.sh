#!/bin/bash

# MUNO Regression Test Suite
# Comprehensive testing for all MUNO functionality

# Don't use set -e as it causes the script to exit on test failures
# We want to run all tests even if some fail
set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="/tmp/muno-regression-test"
WORKSPACE_DIR="$TEST_DIR/test-workspace"
REPOS_DIR="$TEST_DIR/test-repos"
# Determine MUNO binary path
if [[ -n "$MUNO_BIN" ]]; then
    MUNO_BIN="$MUNO_BIN"
elif [[ -f "$(pwd)/bin/muno" ]]; then
    MUNO_BIN="$(pwd)/bin/muno"
elif [[ -f "$(dirname $(dirname $(pwd)))/bin/muno" ]]; then
    MUNO_BIN="$(dirname $(dirname $(pwd)))/bin/muno"
else
    MUNO_BIN="$(pwd)/bin/muno"
fi
REPORT_FILE="$TEST_DIR/regression_report_$(date +%Y%m%d_%H%M%S).txt"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Arrays to track test results
declare -a FAILED_TEST_NAMES=()
declare -a SKIPPED_TEST_NAMES=()

# Cleanup function
cleanup() {
    if [[ -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
    fi
}

# Setup test environment
setup_environment() {
    echo -e "${BLUE}▶ Setting up test environment...${NC}"
    
    # Clean up any previous test run
    cleanup
    
    # Create test directories
    mkdir -p "$TEST_DIR"
    mkdir -p "$REPOS_DIR"
    
    # Create test repositories
    local repos=("backend-monorepo" "frontend-app" "payment-service" "user-service" "auth-service")
    for repo in "${repos[@]}"; do
        local repo_path="$REPOS_DIR/$repo"
        mkdir -p "$repo_path"
        cd "$repo_path"
        git init --quiet
        echo "# $repo" > README.md
        git add . >/dev/null 2>&1
        git commit -m "Initial commit" --quiet
        cd - >/dev/null
    done
    
    echo -e "  ${GREEN}✓${NC} Created ${#repos[@]} test repositories"
    
    # Initialize MUNO workspace
    mkdir -p "$WORKSPACE_DIR"
    cd "$WORKSPACE_DIR"
    "$MUNO_BIN" init test-workspace >/dev/null 2>&1
    
    # Add repositories to workspace config
    cat > muno.yaml << EOF
workspace:
    name: test-workspace
    repos_dir: nodes
nodes:
    - name: backend-monorepo
      url: file://$REPOS_DIR/backend-monorepo
      fetch: eager
    - name: frontend-app
      url: file://$REPOS_DIR/frontend-app
      fetch: eager
    - name: payment-service
      url: file://$REPOS_DIR/payment-service
      fetch: lazy
    - name: user-service
      url: file://$REPOS_DIR/user-service
      fetch: lazy
    - name: auth-service
      url: file://$REPOS_DIR/auth-service
      fetch: lazy
EOF
    
    echo -e "  ${GREEN}✓${NC} Workspace initialized"
}

# Test case function
test_case() {
    local test_name="$1"
    local test_command="$2"
    local skip_test="${3:-false}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    printf "[%2d] %-50s " "$TOTAL_TESTS" "$test_name"
    
    if [[ "$skip_test" == "skip" ]]; then
        echo -e "${YELLOW}⊘ (skipped)${NC}"
        SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
        SKIPPED_TEST_NAMES+=("$test_name")
        echo "[$TOTAL_TESTS] $test_name: SKIPPED" >> "$REPORT_FILE"
        return 0
    fi
    
    if eval "$test_command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo "[$TOTAL_TESTS] $test_name: PASSED" >> "$REPORT_FILE"
        return 0
    else
        echo -e "${RED}✗${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$test_name")
        echo "[$TOTAL_TESTS] $test_name: FAILED" >> "$REPORT_FILE"
        return 1
    fi
}

# Build MUNO
build_muno() {
    echo -e "${BLUE}▶ Building MUNO...${NC}"
    # Find project root (where Makefile exists)
    local project_root
    if [[ -f "$(pwd)/Makefile" ]]; then
        project_root="$(pwd)"
    elif [[ -f "$(dirname $(dirname $(pwd)))/Makefile" ]]; then
        project_root="$(dirname $(dirname $(pwd)))"
    else
        project_root="$(dirname "$MUNO_BIN")/.."
    fi
    cd "$project_root"
    if make build >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} Build complete: $MUNO_BIN"
    else
        echo -e "  ${RED}✗${NC} Build failed"
        exit 1
    fi
}

# Run all tests
run_tests() {
    cd "$WORKSPACE_DIR"
    
    # Initialize report
    echo "MUNO Regression Test Report" > "$REPORT_FILE"
    echo "Generated: $(date)" >> "$REPORT_FILE"
    echo "Binary: $MUNO_BIN" >> "$REPORT_FILE"
    echo "Test Directory: $TEST_DIR" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}                          RUNNING TESTS${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
    echo ""
    
    # Configuration & Setup Tests
    echo -e "${MAGENTA}▶ Configuration & Setup${NC}"
    test_case "Workspace directory exists" "[[ -d '$WORKSPACE_DIR' ]]"
    test_case "Configuration file exists" "[[ -f '$WORKSPACE_DIR/muno.yaml' ]]"
    test_case "State directory exists" "[[ -d '$WORKSPACE_DIR/.muno' ]]"
    # The nodes directory might be "repos" based on config
    test_case "Repository directory created" "[[ -d '$WORKSPACE_DIR/nodes' ]] || [[ -d '$WORKSPACE_DIR/repos' ]]"
    test_case "Current file created" "[[ -f '$WORKSPACE_DIR/.muno/current' ]]"
    echo ""
    
    # Core Commands Tests
    echo -e "${MAGENTA}▶ Core Commands${NC}"
    test_case "Tree command works" "$MUNO_BIN tree"
    test_case "List command works" "$MUNO_BIN list"
    test_case "Current command works" "$MUNO_BIN current"
    test_case "Help displays correctly" "$MUNO_BIN --help"
    test_case "Version displays correctly" "$MUNO_BIN --version"
    echo ""
    
    # Navigation - Eager Repositories
    echo -e "${MAGENTA}▶ Navigation - Eager Repositories${NC}"
    test_case "Navigate to root" "$MUNO_BIN use /"
    test_case "Navigate to eager repo" "$MUNO_BIN use backend-monorepo"
    test_case "Eager repo auto-cloned" "[[ -d '$WORKSPACE_DIR/nodes/backend-monorepo/.git' ]]"
    test_case "Navigate to another eager repo" "$MUNO_BIN use /frontend-app"
    test_case "Return to root" "$MUNO_BIN use /"
    echo ""
    
    # Navigation - Lazy Repositories
    echo -e "${MAGENTA}▶ Navigation - Lazy Repositories${NC}"
    test_case "Lazy repo not cloned initially" "[[ ! -d '$WORKSPACE_DIR/nodes/payment-service/.git' ]]"
    test_case "Navigate to lazy repo" "$MUNO_BIN use payment-service"
    test_case "Lazy repo auto-cloned on nav" "[[ -d '$WORKSPACE_DIR/nodes/payment-service/.git' ]]"
    test_case "Return to root" "$MUNO_BIN use /"
    echo ""
    
    # Clone Operations
    echo -e "${MAGENTA}▶ Clone Operations${NC}"
    # Remove a lazy repo first to test cloning
    if [[ -d "$WORKSPACE_DIR/nodes/user-service" ]]; then
        rm -rf "$WORKSPACE_DIR/nodes/user-service"
    fi
    test_case "Lazy repo successfully removed" "[[ ! -d '$WORKSPACE_DIR/nodes/user-service/.git' ]]"
    test_case "Navigate to trigger clone" "$MUNO_BIN use user-service"
    test_case "Lazy repo cloned on navigation" "[[ -d '$WORKSPACE_DIR/nodes/user-service/.git' ]]"
    test_case "Clone all remaining lazy repos" "$MUNO_BIN clone --recursive"
    echo ""
    
    # Repository Management (ADD/REMOVE with persistence)
    echo -e "${MAGENTA}▶ Repository Management${NC}"
    cd "$WORKSPACE_DIR"
    
    # Navigate to root first for add/remove operations
    $MUNO_BIN use / >/dev/null 2>&1
    
    # Test ADD with persistence
    # First create the test-add repository
    mkdir -p "$REPOS_DIR/test-add"
    cd "$REPOS_DIR/test-add"
    git init --quiet
    echo "# test-add" > README.md
    git add . >/dev/null 2>&1
    git commit -m "Initial commit" --quiet
    cd "$WORKSPACE_DIR"
    
    test_case "Add repository succeeds" "$MUNO_BIN add 'file://$REPOS_DIR/test-add' --name test-add --lazy"
    test_case "Add updates config file" "grep -q 'test-add' muno.yaml"
    test_case "Added repo appears in list" "$MUNO_BIN list | grep -q 'test-add'"
    
    # Test that added repo persists across commands
    test_case "Added repo persists in tree" "$MUNO_BIN tree | grep -q 'test-add'"
    
    # Test REMOVE with persistence
    test_case "Remove repository succeeds" "echo y | $MUNO_BIN remove test-add"
    test_case "Remove updates config file" "! grep -q 'test-add' muno.yaml"
    test_case "Removed repo not in list" "! $MUNO_BIN list | grep -q 'test-add'"
    test_case "Removed repo not in tree" "! $MUNO_BIN tree | grep -q 'test-add'"
    echo ""
    
    # Git Operations
    echo -e "${MAGENTA}▶ Git Operations${NC}"
    test_case "Status command works" "$MUNO_BIN status"
    
    # Create a change to detect
    if [[ -d "$WORKSPACE_DIR/nodes/backend-monorepo" ]]; then
        echo "test change" > "$WORKSPACE_DIR/nodes/backend-monorepo/test.txt"
    fi
    # TODO: Fix status command to detect untracked files
    test_case "Status detects file changes" "$MUNO_BIN status /backend-monorepo 2>&1 | grep -E 'test.txt|untracked|Changes|modified'"
    
    # Clean up the test file
    rm -f "$WORKSPACE_DIR/nodes/backend-monorepo/test.txt"
    echo ""
    
    # Error Handling
    echo -e "${MAGENTA}▶ Error Handling${NC}"
    test_case "Invalid node handled gracefully" "! $MUNO_BIN use non-existent-node"
    test_case "Invalid command handled gracefully" "! $MUNO_BIN invalid-command"
    test_case "Remove non-existent fails properly" "! echo y | $MUNO_BIN remove non-existent"
    echo ""
}

# Display summary
display_summary() {
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}                           TEST SUMMARY${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
    echo ""
    
    local pass_rate=0
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi
    
    echo -e "${BLUE}┌─────────────────────────────────────┐${NC}"
    echo -e "${BLUE}│${NC} Total Tests:     ${CYAN}$(printf "%3d" $TOTAL_TESTS)               ${NC} ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC} Passed:          ${GREEN}$(printf "%3d" $PASSED_TESTS)               ${NC} ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC} Failed:          ${RED}$(printf "%3d" $FAILED_TESTS)               ${NC} ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC} Skipped:         ${YELLOW}$(printf "%3d" $SKIPPED_TESTS)               ${NC} ${BLUE}│${NC}"
    echo -e "${BLUE}│${NC} Pass Rate:       $(printf "%3d%%" $pass_rate)              ${NC} ${BLUE}│${NC}"
    echo -e "${BLUE}└─────────────────────────────────────┘${NC}"
    echo ""
    
    # Display result message
    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║                    ALL TESTS PASSED!                          ║${NC}"
        echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    else
        echo -e "${RED}╔════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${RED}║                    SOME TESTS FAILED                          ║${NC}"
        echo -e "${RED}╚════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        echo -e "${RED}Failed tests:${NC}"
        for test_name in "${FAILED_TEST_NAMES[@]}"; do
            echo -e "  ${RED}✗${NC} $test_name"
        done
    fi
    
    if [[ ${#SKIPPED_TEST_NAMES[@]} -gt 0 ]]; then
        echo ""
        echo -e "${YELLOW}Skipped tests:${NC}"
        for test_name in "${SKIPPED_TEST_NAMES[@]}"; do
            echo -e "  ${YELLOW}⊘${NC} $test_name"
        done
    fi
    
    echo ""
    echo -e "${CYAN}📄 Report saved: $REPORT_FILE${NC}"
    echo ""
    
    # Append summary to report
    echo "" >> "$REPORT_FILE"
    echo "SUMMARY" >> "$REPORT_FILE"
    echo "=======" >> "$REPORT_FILE"
    echo "Total Tests: $TOTAL_TESTS" >> "$REPORT_FILE"
    echo "Passed: $PASSED_TESTS" >> "$REPORT_FILE"
    echo "Failed: $FAILED_TESTS" >> "$REPORT_FILE"
    echo "Skipped: $SKIPPED_TESTS" >> "$REPORT_FILE"
    echo "Pass Rate: ${pass_rate}%" >> "$REPORT_FILE"
}

# Main execution
main() {
    # Clear screen for clean output
    clear
    
    # Display header
    echo -e "${MAGENTA}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${MAGENTA}║                 MUNO Regression Test Suite                    ║${NC}"
    echo -e "${MAGENTA}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # Build MUNO
    build_muno
    echo ""
    
    # Setup environment
    setup_environment
    echo ""
    
    # Run tests
    run_tests
    
    # Display summary
    display_summary
    
    # Cleanup
    cleanup
    
    # Exit with appropriate code
    if [[ $FAILED_TESTS -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"