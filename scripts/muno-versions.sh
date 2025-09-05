#!/bin/bash
# MUNO Version Manager
# Manages multiple muno installations that work simultaneously

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_usage() {
    echo "MUNO Version Manager"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  status        Show status of all muno versions"
    echo "  install       Install all development versions (dev & local)"
    echo "  install-dev   Install only muno-dev"
    echo "  install-local Install only muno-local"
    echo "  update        Update dev and local versions with latest code"
    echo "  clean         Remove dev and local versions (keeps production)"
    echo "  help          Show this help message"
    echo ""
    echo "Version usage:"
    echo "  muno          Production version (installed via homebrew/go install)"
    echo "  muno-dev      Development version (your current branch)"
    echo "  muno-local    Local testing version (for experiments)"
    echo ""
    echo "Examples:"
    echo "  $0 status"
    echo "  $0 install"
    echo "  $0 update"
}

show_status() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}        MUNO Installation Status${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo ""
    
    # Check production version
    echo -e "${YELLOW}Production Version (muno):${NC}"
    if command -v muno &> /dev/null; then
        local muno_path=$(which muno)
        local muno_version=$(muno --version 2>/dev/null || echo "unknown")
        echo -e "  ${GREEN}✓${NC} Path: $muno_path"
        echo -e "  ${GREEN}✓${NC} Version: $muno_version"
        
        # Check if it's homebrew or go install
        if [[ "$muno_path" == *"homebrew"* ]]; then
            echo -e "  ${CYAN}ℹ${NC} Installed via: Homebrew"
        elif [[ "$muno_path" == *"go/bin"* ]]; then
            echo -e "  ${CYAN}ℹ${NC} Installed via: go install"
        else
            echo -e "  ${CYAN}ℹ${NC} Installed via: Manual"
        fi
    else
        echo -e "  ${RED}✗${NC} Not installed"
    fi
    echo ""
    
    # Check development version
    echo -e "${YELLOW}Development Version (muno-dev):${NC}"
    if command -v muno-dev &> /dev/null; then
        local dev_path=$(which muno-dev)
        local dev_version=$(muno-dev --version 2>/dev/null || echo "unknown")
        echo -e "  ${GREEN}✓${NC} Path: $dev_path"
        echo -e "  ${GREEN}✓${NC} Version: $dev_version"
        
        # Show git info
        cd "$PROJECT_ROOT"
        local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
        local commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        echo -e "  ${CYAN}ℹ${NC} Git Branch: $branch"
        echo -e "  ${CYAN}ℹ${NC} Git Commit: $commit"
    else
        echo -e "  ${RED}✗${NC} Not installed"
        echo -e "  ${CYAN}ℹ${NC} Run: make install-dev"
    fi
    echo ""
    
    # Check local version
    echo -e "${YELLOW}Local Version (muno-local):${NC}"
    if command -v muno-local &> /dev/null; then
        local local_path=$(which muno-local)
        local local_version=$(muno-local --version 2>/dev/null || echo "unknown")
        echo -e "  ${GREEN}✓${NC} Path: $local_path"
        echo -e "  ${GREEN}✓${NC} Version: $local_version"
    else
        echo -e "  ${RED}✗${NC} Not installed"
        echo -e "  ${CYAN}ℹ${NC} Run: make install-local"
    fi
    echo ""
    
    # Show usage recommendations
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}Usage Recommendations:${NC}"
    echo ""
    echo -e "• Use ${GREEN}muno${NC} for stable operations"
    echo -e "• Use ${YELLOW}muno-dev${NC} for testing new features"
    echo -e "• Use ${YELLOW}muno-local${NC} for experiments"
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
}

install_all() {
    echo -e "${BLUE}Installing development versions...${NC}"
    echo ""
    
    cd "$PROJECT_ROOT"
    
    echo -e "${YELLOW}Building and installing muno-dev...${NC}"
    make install-dev
    echo ""
    
    echo -e "${YELLOW}Building and installing muno-local...${NC}"
    make install-local
    echo ""
    
    echo -e "${GREEN}✓ Installation complete${NC}"
    echo ""
    show_status
}

install_dev_only() {
    echo -e "${BLUE}Installing development version...${NC}"
    echo ""
    
    cd "$PROJECT_ROOT"
    make install-dev
    
    echo -e "${GREEN}✓ muno-dev installed${NC}"
    echo ""
    
    if command -v muno-dev &> /dev/null; then
        local version=$(muno-dev --version)
        echo -e "You can now use: ${YELLOW}muno-dev${NC}"
        echo -e "Version: $version"
    fi
}

install_local_only() {
    echo -e "${BLUE}Installing local version...${NC}"
    echo ""
    
    cd "$PROJECT_ROOT"
    make install-local
    
    echo -e "${GREEN}✓ muno-local installed${NC}"
    echo ""
    
    if command -v muno-local &> /dev/null; then
        local version=$(muno-local --version)
        echo -e "You can now use: ${YELLOW}muno-local${NC}"
        echo -e "Version: $version"
    fi
}

update_versions() {
    echo -e "${BLUE}Updating development versions with latest code...${NC}"
    echo ""
    
    cd "$PROJECT_ROOT"
    
    # Show git status
    echo -e "${CYAN}Current git status:${NC}"
    git status --short
    echo ""
    
    # Rebuild and install
    echo -e "${YELLOW}Rebuilding muno-dev...${NC}"
    make clean
    make install-dev
    echo ""
    
    echo -e "${YELLOW}Rebuilding muno-local...${NC}"
    make install-local
    echo ""
    
    echo -e "${GREEN}✓ Update complete${NC}"
    echo ""
    
    # Show new versions
    if command -v muno-dev &> /dev/null; then
        echo -e "muno-dev version: $(muno-dev --version)"
    fi
    if command -v muno-local &> /dev/null; then
        echo -e "muno-local version: $(muno-local --version)"
    fi
}

clean_dev_versions() {
    echo -e "${RED}Removing development versions...${NC}"
    echo -e "${CYAN}Note: Production 'muno' will be preserved${NC}"
    echo ""
    
    cd "$PROJECT_ROOT"
    
    make uninstall-dev 2>/dev/null || true
    make uninstall-local 2>/dev/null || true
    
    # Additional cleanup
    rm -f "$HOME/go/bin/muno-dev" 2>/dev/null || true
    rm -f "$HOME/go/bin/muno-local" 2>/dev/null || true
    rm -f "$GOPATH/bin/muno-dev" 2>/dev/null || true
    rm -f "$GOPATH/bin/muno-local" 2>/dev/null || true
    
    echo -e "${GREEN}✓ Development versions removed${NC}"
    echo -e "${CYAN}ℹ Production 'muno' is still available${NC}"
}

# Main script logic
case "${1:-}" in
    status)
        show_status
        ;;
    install)
        install_all
        ;;
    install-dev)
        install_dev_only
        ;;
    install-local)
        install_local_only
        ;;
    update)
        update_versions
        ;;
    clean)
        clean_dev_versions
        ;;
    help|--help|-h|"")
        print_usage
        ;;
    *)
        echo -e "${RED}Error: Unknown command '$1'${NC}"
        print_usage
        exit 1
        ;;
esac