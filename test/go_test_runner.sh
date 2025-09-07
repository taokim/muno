#!/bin/bash

# Go Test Runner for MUNO
# Runs all Go unit and integration tests

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                    MUNO Go Test Suite                          ║${NC}"  
echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Run tests with coverage
echo -e "${BLUE}▶ Running Go tests with coverage...${NC}"
echo ""

# Test each package
PACKAGES=(
    "./internal/config"
    "./internal/git"
    "./internal/tree"
    "./internal/manager"
    "./internal/adapters"
    "./internal/interfaces"
    "./internal/plugin"
    "./cmd/muno"
    "./test"
    "./test/e2e"
)

TOTAL_PACKAGES=${#PACKAGES[@]}
PASSED_PACKAGES=0
FAILED_PACKAGES=0
FAILED_LIST=()

for pkg in "${PACKAGES[@]}"; do
    echo -e "${BLUE}Testing package: $pkg${NC}"
    if go test "$pkg" -cover -timeout 30s > /tmp/test_output.txt 2>&1; then
        COVERAGE=$(grep "coverage:" /tmp/test_output.txt | awk '{print $2}')
        echo -e "  ${GREEN}✓ PASS${NC} (coverage: ${COVERAGE:-N/A})"
        PASSED_PACKAGES=$((PASSED_PACKAGES + 1))
    else
        echo -e "  ${RED}✗ FAIL${NC}"
        FAILED_PACKAGES=$((FAILED_PACKAGES + 1))
        FAILED_LIST+=("$pkg")
        # Show error details
        tail -5 /tmp/test_output.txt | sed 's/^/    /'
    fi
    echo ""
done

# Overall coverage
echo -e "${BLUE}▶ Generating overall coverage report...${NC}"
go test ./... -coverprofile=coverage.out > /dev/null 2>&1 || true
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                         GO TEST SUMMARY${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${BLUE}┌─────────────────────────────────────┐${NC}"
echo -e "${BLUE}│${NC} Packages Tested: ${CYAN}$(printf "%2d" $TOTAL_PACKAGES)                ${NC}${BLUE}│${NC}"
echo -e "${BLUE}│${NC} Passed:          ${GREEN}$(printf "%2d" $PASSED_PACKAGES)                ${NC}${BLUE}│${NC}"
echo -e "${BLUE}│${NC} Failed:          ${RED}$(printf "%2d" $FAILED_PACKAGES)                ${NC}${BLUE}│${NC}"
echo -e "${BLUE}│${NC} Coverage:        ${CYAN}${TOTAL_COVERAGE:-N/A}              ${NC}${BLUE}│${NC}"
echo -e "${BLUE}└─────────────────────────────────────┘${NC}"

if [[ ${#FAILED_LIST[@]} -gt 0 ]]; then
    echo ""
    echo -e "${RED}Failed packages:${NC}"
    for pkg in "${FAILED_LIST[@]}"; do
        echo -e "  ${RED}✗${NC} $pkg"
    done
fi

# Coverage analysis
echo ""
echo -e "${CYAN}Coverage Analysis:${NC}"
go tool cover -func=coverage.out 2>/dev/null | grep -E "^github.com/taokim/muno/(internal|cmd)" | \
    awk '$3 ~ /[0-9]+\.[0-9]+%/ {
        coverage = substr($3, 1, length($3)-1)
        if (coverage >= 80) {
            printf "  ✓ %-50s %s (good)\n", $1, $3
        } else if (coverage >= 60) {
            printf "  ⚠ %-50s %s (fair)\n", $1, $3
        } else {
            printf "  ✗ %-50s %s (needs improvement)\n", $1, $3
        }
    }' | head -20

echo ""

# Generate HTML coverage report
echo -e "${BLUE}▶ Generating HTML coverage report...${NC}"
go tool cover -html=coverage.out -o coverage.html 2>/dev/null
echo -e "  ${GREEN}✓${NC} Coverage report saved to coverage.html"

echo ""

# Determine status
if [[ $FAILED_PACKAGES -eq 0 ]]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                  All Go Tests Passed! ✅                       ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                  Some Go Tests Failed ❌                       ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi