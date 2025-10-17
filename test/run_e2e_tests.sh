#!/bin/bash

# MUNO End-to-End Test Suite
# Single entry point for all E2E tests
# Tests real-world workflows and user scenarios

set +e  # Don't exit on first failure

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Configuration
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Parse arguments
VERBOSE=false
PARALLEL=false
for arg in "$@"; do
    case $arg in
        --verbose) VERBOSE=true ;;
        --parallel) PARALLEL=true ;;
        --help)
            echo "Usage: $0 [options]"
            echo "  --verbose   Show detailed output"
            echo "  --parallel  Run Go tests in parallel"
            exit 0
            ;;
    esac
done

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MUNO_BIN="$PROJECT_ROOT/bin/muno"

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Helper Functions
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

log() {
    echo -e "$1"
}

log_verbose() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "$1"
    fi
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    ((TOTAL_TESTS++))
    echo -n "  Running: $test_name ... "
    
    local output
    output=$(eval "$test_command" 2>&1)
    local result=$?
    
    if [[ $result -eq 0 ]]; then
        echo -e "${GREEN}✓${NC}"
        ((PASSED_TESTS++))
        log_verbose "    Output: $output"
        return 0
    else
        echo -e "${RED}✗${NC}"
        ((FAILED_TESTS++))
        log_verbose "    Failed with output:"
        log_verbose "$output"
        return 1
    fi
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Build Check
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

check_binary() {
    if [[ ! -f "$MUNO_BIN" ]]; then
        log "${YELLOW}MUNO binary not found. Building...${NC}"
        (cd "$PROJECT_ROOT" && make build)
        if [[ ! -f "$MUNO_BIN" ]]; then
            log "${RED}Failed to build MUNO binary${NC}"
            exit 1
        fi
    fi
    log "${GREEN}✓${NC} Using binary: $MUNO_BIN"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Go E2E Tests
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

run_go_e2e_tests() {
    log ""
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    log "${BLUE}▶ Go E2E Tests${NC}"
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    cd "$PROJECT_ROOT"
    
    # Run integration tests
    run_test "Integration Tests" \
        "go test ./test -run TestIntegration -timeout 30s"
    
    # Run E2E workflow tests
    run_test "E2E Workflow Tests" \
        "go test ./test/e2e -timeout 30s"
    
    # Run clone/pull integration test
    run_test "Clone/Pull Integration" \
        "go test ./test -run TestClonePull -timeout 30s"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Scenario Tests
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

run_scenario_tests() {
    log ""
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    log "${BLUE}▶ Real-World Scenario Tests${NC}"
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    local test_dir="/tmp/muno-e2e-$$"
    
    # Scenario 1: Developer Workflow
    run_test "Developer Workflow" "bash -c '
        mkdir -p $test_dir/dev-workflow
        cd $test_dir/dev-workflow
        
        # Initialize workspace
        $MUNO_BIN init my-project >/dev/null 2>&1 || exit 1
        
        # Add repositories
        $MUNO_BIN add https://github.com/golang/example --lazy >/dev/null 2>&1 || exit 1
        
        # Clone repositories
        $MUNO_BIN clone --include-lazy >/dev/null 2>&1 || exit 1
        
        # Check status
        $MUNO_BIN status >/dev/null 2>&1 || exit 1
        
        # Tree view
        $MUNO_BIN tree >/dev/null 2>&1 || exit 1
        
        exit 0
    '"
    
    # Scenario 2: Team Collaboration
    run_test "Team Collaboration" "bash -c '
        mkdir -p $test_dir/team-collab
        cd $test_dir/team-collab
        
        # Initialize shared workspace
        $MUNO_BIN init team-workspace >/dev/null 2>&1 || exit 1
        
        # Add team repositories
        $MUNO_BIN add https://github.com/golang/tools --name tools --lazy >/dev/null 2>&1 || exit 1
        
        # Pull updates (should not clone lazy repos)
        $MUNO_BIN pull --recursive >/dev/null 2>&1
        
        # List repositories
        $MUNO_BIN list >/dev/null 2>&1 || exit 1
        
        exit 0
    '"
    
    # Cleanup
    rm -rf "$test_dir"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Performance Tests
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

run_performance_tests() {
    if [[ "$VERBOSE" != "true" ]]; then
        return 0  # Skip in non-verbose mode
    fi
    
    log ""
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    log "${BLUE}▶ Performance Tests${NC}"
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    local test_dir="/tmp/muno-perf-$$"
    mkdir -p "$test_dir"
    cd "$test_dir"
    
    # Test with many repositories
    run_test "Handle 10+ repositories" "bash -c '
        $MUNO_BIN init perf-test >/dev/null 2>&1 || exit 1
        
        # Add multiple repos
        for i in {1..10}; do
            echo \"Adding repo \$i\"
            $MUNO_BIN add https://github.com/golang/example --name \"repo-\$i\" --lazy >/dev/null 2>&1
        done
        
        # List should handle many repos
        $MUNO_BIN list >/dev/null 2>&1 || exit 1
        
        # Tree should render properly
        $MUNO_BIN tree >/dev/null 2>&1 || exit 1
        
        exit 0
    '"
    
    # Cleanup
    rm -rf "$test_dir"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# CLI Tests
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

run_cli_tests() {
    log ""
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    log "${BLUE}▶ CLI Interface Tests${NC}"
    log "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Test help commands
    run_test "Main help" "$MUNO_BIN --help >/dev/null 2>&1"
    run_test "Init help" "$MUNO_BIN init --help >/dev/null 2>&1"
    run_test "Add help" "$MUNO_BIN add --help >/dev/null 2>&1"
    run_test "Clone help" "$MUNO_BIN clone --help >/dev/null 2>&1"
    run_test "Pull help" "$MUNO_BIN pull --help >/dev/null 2>&1"
    
    # Test version
    run_test "Version flag" "$MUNO_BIN --version >/dev/null 2>&1"
    
    # Test invalid commands
    run_test "Invalid command fails" "! $MUNO_BIN invalid-cmd >/dev/null 2>&1"
    run_test "Missing arguments fails" "! $MUNO_BIN add >/dev/null 2>&1"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Main Execution
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

main() {
    log ""
    log "${CYAN}╔═══════════════════════════════════════════════════╗${NC}"
    log "${CYAN}║          MUNO End-to-End Test Suite              ║${NC}"
    log "${CYAN}╚═══════════════════════════════════════════════════╝${NC}"
    log ""
    
    # Check binary
    check_binary
    
    # Run test suites
    run_cli_tests
    run_go_e2e_tests
    run_scenario_tests
    run_performance_tests
    
    # Summary
    log ""
    log "${CYAN}═══════════════════════════════════════════════════${NC}"
    log "${CYAN}Test Summary:${NC}"
    log "  Total: $TOTAL_TESTS"
    log "  ${GREEN}Passed: $PASSED_TESTS${NC}"
    log "  ${RED}Failed: $FAILED_TESTS${NC}"
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        log ""
        log "${GREEN}✓ All E2E tests passed!${NC}"
        exit 0
    else
        log ""
        log "${RED}✗ Some E2E tests failed!${NC}"
        exit 1
    fi
}

main