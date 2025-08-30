#!/bin/bash

# Generate Test Tree Script for repo-claude
# Creates a flat repository pool and a meta tree structure for testing
# Usage: ./generate-test-tree.sh [num_repos] [tree_depth]

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RC_BIN="${RC_BIN:-$PROJECT_DIR/bin/rc}"

# Configuration - Always use /tmp
BASE_DIR="/tmp/test-tree-$(date +%Y%m%d-%H%M%S)"
NUM_REPOS="${1:-20}"
TREE_DEPTH="${2:-3}"
TOTAL_REPOS=0
TOTAL_NODES=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Announce the test directory at the start
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}     REPO-CLAUDE TEST TREE GENERATOR${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo
echo -e "${YELLOW}📁 Test Directory: ${BLUE}$BASE_DIR${NC}"
echo -e "${YELLOW}📊 Configuration:${NC}"
echo -e "   • Repository Pool Size: ${GREEN}$NUM_REPOS${NC}"
echo -e "   • Tree Depth: ${GREEN}$TREE_DEPTH${NC}"
echo
echo -e "${CYAN}────────────────────────────────────────────────────────────────${NC}"
echo

# Repository templates for realistic testing
declare -a REPO_TYPES=(
    "frontend:react,typescript,webpack"
    "backend:golang,api,database"
    "mobile:flutter,dart,ios,android"
    "service:microservice,docker,kubernetes"
    "library:shared,utils,common"
    "data:etl,pipeline,analytics"
    "ml:tensorflow,pytorch,models"
    "docs:documentation,wiki,guides"
    "infra:terraform,ansible,aws"
    "tools:cli,scripts,automation"
)

# Function to create a git repository with content
create_git_repo() {
    local repo_path="$1"
    local repo_type="$2"
    local repo_name=$(basename "$repo_path")
    
    echo -e "  ${BLUE}→ Creating: $repo_name ($repo_type)${NC}"
    
    # Create directory
    mkdir -p "$repo_path"
    cd "$repo_path" || return 1
    
    # Initialize git
    git init >/dev/null 2>&1 || return 1
    
    # Configure git for this repo
    git config user.name "Test User" || return 1
    git config user.email "test@example.com" || return 1
    
    # Create appropriate files based on type
    case "$repo_type" in
        frontend)
            mkdir -p src 2>/dev/null || true
            cat > package.json <<EOF
{
  "name": "$repo_name",
  "version": "1.0.0",
  "type": "frontend",
  "scripts": {
    "dev": "webpack serve",
    "build": "webpack build"
  }
}
EOF
            cat > src/index.ts <<EOF
// $repo_name - Frontend Application
export function main() {
    console.log('Frontend: $repo_name');
}
EOF
            mkdir -p src/components 2>/dev/null || true
            echo "export const Button = () => 'Button';" > src/components/Button.tsx 2>/dev/null || true
            ;;
            
        backend)
            cat > go.mod <<EOF
module github.com/test/$repo_name

go 1.21
EOF
            cat > main.go <<EOF
package main

import "fmt"

func main() {
    fmt.Println("Backend: $repo_name")
}
EOF
            mkdir -p internal/api 2>/dev/null || true
            echo "package api" > internal/api/server.go 2>/dev/null || true
            ;;
            
        service)
            cat > Dockerfile <<EOF
FROM alpine:latest
LABEL service="$repo_name"
CMD ["echo", "Service: $repo_name"]
EOF
            cat > docker-compose.yml <<EOF
version: '3'
services:
  $repo_name:
    build: .
    ports:
      - "8080:8080"
EOF
            ;;
            
        library)
            cat > README.md <<EOF
# $repo_name

Shared library for common functionality.

## Usage
\`\`\`
import "$repo_name"
\`\`\`
EOF
            mkdir -p lib 2>/dev/null || true
            echo "// Shared utilities" > lib/utils.go 2>/dev/null || true
            ;;
            
        data)
            cat > pipeline.yaml <<EOF
name: $repo_name
type: data-pipeline
steps:
  - extract
  - transform
  - load
EOF
            mkdir -p sql 2>/dev/null || true
            echo "SELECT * FROM $repo_name;" > sql/query.sql 2>/dev/null || true
            ;;
            
        ml)
            cat > model.py <<EOF
#!/usr/bin/env python3
"""$repo_name - ML Model"""

import numpy as np

class Model:
    def __init__(self):
        self.name = "$repo_name"
    
    def train(self, data):
        pass
    
    def predict(self, input):
        return np.random.random()
EOF
            cat > requirements.txt <<EOF
numpy>=1.20.0
tensorflow>=2.0.0
pandas>=1.3.0
EOF
            ;;
            
        docs)
            cat > README.md <<EOF
# $repo_name Documentation

## Overview
This repository contains documentation for the system.

## Contents
- [Architecture](./architecture.md)
- [API Reference](./api.md)
- [User Guide](./guide.md)
EOF
            echo "# Architecture" > architecture.md 2>/dev/null || true
            echo "# API Reference" > api.md 2>/dev/null || true
            ;;
            
        infra)
            cat > main.tf <<EOF
# $repo_name Infrastructure

provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "$repo_name" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
}
EOF
            ;;
            
        tools)
            cat > tool.sh <<EOF
#!/bin/bash
# $repo_name - CLI Tool

echo "Tool: $repo_name"
echo "Args: \$@"
EOF
            chmod +x tool.sh
            ;;
            
        *)
            cat > README.md <<EOF
# $repo_name

Default repository type.
EOF
            ;;
    esac
    
    # Add .gitignore
    cat > .gitignore <<EOF
*.log
*.tmp
.DS_Store
node_modules/
dist/
build/
*.pyc
__pycache__/
.terraform/
EOF
    
    # Commit files
    git add . >/dev/null 2>&1
    git commit -m "Initial commit for $repo_name" --quiet 2>/dev/null || {
        echo -e "    ${YELLOW}Warning: Could not commit in $repo_name${NC}"
        return 1
    }
    
    # Add some history (optional, skip if fails)
    if [ -f README.md ]; then
        echo "# Updates" >> README.md
        git add README.md >/dev/null 2>&1
        git commit -m "Add updates section" --quiet 2>/dev/null || true
    fi
    
    # Create a feature branch (optional)
    git checkout -b feature/test --quiet 2>/dev/null || true
    echo "Feature work" > feature.txt
    git add feature.txt >/dev/null 2>&1
    git commit -m "Add feature" --quiet 2>/dev/null || true
    
    # Switch back to default branch
    if git show-ref --verify --quiet refs/heads/main; then
        git checkout main --quiet 2>/dev/null
    elif git show-ref --verify --quiet refs/heads/master; then
        git checkout master --quiet 2>/dev/null
    fi
    
    # Return to parent directory
    cd - >/dev/null 2>&1
    
    # Increment counter (use || true to avoid exit on 0)
    ((TOTAL_REPOS++)) || true
}

# Function to generate flat repository pool
generate_repo_pool() {
    local pool_dir="$1"
    
    echo -e "${YELLOW}Creating flat repository pool at $pool_dir${NC}"
    mkdir -p "$pool_dir"
    
    for i in $(seq 1 $NUM_REPOS); do
        # Select random repo type
        local type_index=$((RANDOM % ${#REPO_TYPES[@]}))
        local type_spec="${REPO_TYPES[$type_index]}"
        local type="${type_spec%%:*}"
        
        # Generate unique repo name
        local repo_name="${type}-repo-${i}"
        local repo_path="$pool_dir/$repo_name"
        
        # Create the repository
        create_git_repo "$repo_path" "$type"
    done
    
    echo -e "${GREEN}✓ Created $TOTAL_REPOS repositories in pool${NC}"
}

# Function to build actual tree using rc commands with proper depth
build_tree_with_rc() {
    local workspace_dir="$1"
    local pool_dir="$2"
    local max_depth="${TREE_DEPTH:-3}"
    
    echo -e "${YELLOW}Building tree structure with depth: $max_depth${NC}"
    
    # Get list of repos from pool
    local repos=($(ls -1 "$pool_dir" | grep -v "^\." | sort))
    local num_repos=${#repos[@]}
    
    # Change to workspace directory
    cd "$workspace_dir"
    
    # Initialize the workspace
    echo -e "${BLUE}  → Initializing workspace...${NC}"
    "$RC_BIN" init -n test-tree >/dev/null 2>&1 || {
        echo -e "${RED}Failed to initialize workspace${NC}"
        return 1
    }
    
    # Track which repos we've added
    local repo_index=0
    
    echo -e "${BLUE}  → Building hierarchical tree (depth=$max_depth)...${NC}"
    echo
    
    # For depth 1: Just add all repos at root
    if [ $max_depth -eq 1 ]; then
        echo -e "    ${CYAN}Building flat structure (depth=1)${NC}"
        while [ $repo_index -lt $num_repos ]; do
            local repo_name="repo-$((repo_index+1))"
            echo -e "      Adding $repo_name at root"
            if [ $((repo_index % 3)) -eq 0 ]; then
                "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$repo_name" --lazy >/dev/null 2>&1
            else
                "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$repo_name" >/dev/null 2>&1
            fi
            ((repo_index++)) || true
        done
    
    # For depth 2: Create groups at root, add repos inside them
    elif [ $max_depth -eq 2 ]; then
        echo -e "    ${CYAN}Level 1: Creating main groups${NC}"
        
        # Create 3 main groups using first 3 repos
        local groups=("frontend" "backend" "shared")
        for group in "${groups[@]}"; do
            if [ $repo_index -lt $num_repos ]; then
                echo -e "      Creating $group group..."
                "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$group" >/dev/null 2>&1
                ((repo_index++)) || true
            fi
        done
        
        echo -e "    ${CYAN}Level 2: Adding repos to groups${NC}"
        
        # Distribute remaining repos among groups
        for group in "${groups[@]}"; do
            if [ $repo_index -lt $num_repos ]; then
                "$RC_BIN" use "/$group" >/dev/null 2>&1
                
                # Add 2-3 repos to each group
                for ((i=0; i<3 && repo_index<num_repos; i++)); do
                    local repo_name="${group}-app-$((i+1))"
                    echo -e "      Adding $repo_name to $group"
                    
                    if [ $((repo_index % 2)) -eq 0 ]; then
                        "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$repo_name" --lazy >/dev/null 2>&1
                    else
                        "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$repo_name" >/dev/null 2>&1
                    fi
                    ((repo_index++)) || true
                done
            fi
        done
        
    # For depth 3 or more: Create nested hierarchy
    else
        echo -e "    ${CYAN}Level 1: Creating root groups${NC}"
        
        # Level 1: Create main groups
        if [ $repo_index -lt $num_repos ]; then
            echo -e "      Creating frontend..."
            "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "frontend" >/dev/null 2>&1
            ((repo_index++)) || true
        fi
        
        if [ $repo_index -lt $num_repos ]; then
            echo -e "      Creating backend..."
            "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "backend" >/dev/null 2>&1
            ((repo_index++)) || true
        fi
        
        if [ $repo_index -lt $num_repos ]; then
            echo -e "      Creating shared (lazy)..."
            "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "shared" --lazy >/dev/null 2>&1
            ((repo_index++)) || true
        fi
        
        # Level 2: Add sub-groups
        echo -e "    ${CYAN}Level 2: Creating sub-groups${NC}"
        
        # Frontend sub-groups
        "$RC_BIN" use /frontend >/dev/null 2>&1
        if [ $repo_index -lt $num_repos ]; then
            echo -e "      Adding web to frontend..."
            "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "web" >/dev/null 2>&1
            ((repo_index++)) || true
        fi
        
        if [ $repo_index -lt $num_repos ]; then
            echo -e "      Adding mobile to frontend (lazy)..."
            "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "mobile" --lazy >/dev/null 2>&1
            ((repo_index++)) || true
        fi
        
        # Backend sub-groups
        "$RC_BIN" use /backend >/dev/null 2>&1
        if [ $repo_index -lt $num_repos ]; then
            echo -e "      Adding api to backend..."
            "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "api" >/dev/null 2>&1
            ((repo_index++)) || true
        fi
        
        if [ $repo_index -lt $num_repos ]; then
            echo -e "      Adding services to backend (lazy)..."
            "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "services" --lazy >/dev/null 2>&1
            ((repo_index++)) || true
        fi
        
        # Level 3: Add actual project repos
        if [ $max_depth -ge 3 ] && [ $repo_index -lt $num_repos ]; then
            echo -e "    ${CYAN}Level 3: Adding project repos${NC}"
            
            # Add to frontend/web
            "$RC_BIN" use /frontend/web >/dev/null 2>&1
            for ((i=0; i<2 && repo_index<num_repos; i++)); do
                local repo_name="app-$((i+1))"
                echo -e "      Adding $repo_name to frontend/web"
                "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$repo_name" >/dev/null 2>&1
                ((repo_index++)) || true
            done
            
            # Add to backend/api
            "$RC_BIN" use /backend/api >/dev/null 2>&1
            while [ $repo_index -lt $num_repos ]; do
                local api_num=$((repo_index - 8))
                local repo_name="service-$api_num"
                echo -e "      Adding $repo_name to backend/api"
                
                if [ $((repo_index % 2)) -eq 0 ]; then
                    "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$repo_name" >/dev/null 2>&1
                else
                    "$RC_BIN" add "file://$pool_dir/${repos[$repo_index]}" --name "$repo_name" --lazy >/dev/null 2>&1
                fi
                ((repo_index++)) || true
            done
        fi
    fi
    
    # Go back to root
    "$RC_BIN" use / >/dev/null 2>&1
    
    # Display the final tree (with depth limit to avoid clutter)
    echo
    echo -e "${YELLOW}Generated tree structure:${NC}"
    echo -e "${CYAN}────────────────────────────────────────────${NC}"
    
    # Use depth flag to limit display
    if [ $max_depth -le 3 ]; then
        "$RC_BIN" tree -d $((max_depth + 1)) 2>/dev/null || "$RC_BIN" tree 2>/dev/null
    else
        "$RC_BIN" tree 2>/dev/null
    fi
    
    echo
    echo -e "${GREEN}✓ Tree built: $repo_index repos in $max_depth levels${NC}"
    
    # Create a reference file showing the structure
    cat > "$workspace_dir/TREE_STRUCTURE.md" <<EOF
# Generated Tree Structure

This tree was automatically built using rc commands with depth $max_depth.

## Structure Overview

Tree depth: $max_depth levels
Total repositories added: $repo_index

### Level 1 (Root)
- frontend-main/
- backend-main/
- shared-libs/ (lazy)

### Level 2 (Groups)
- frontend-main/
  - web-apps/
  - mobile-apps/ (lazy)
- backend-main/
  - services/
  - data-layer/ (lazy)

### Level 3 (Projects)
- frontend-main/web-apps/
  - web-project-1, web-project-2, etc.
- backend-main/services/
  - api-service-1, api-service-2, etc.

## Navigation Examples

\`\`\`bash
# Show the full tree
rc tree

# Navigate to different levels
rc use frontend-main                    # Level 1
rc use frontend-main/web-apps          # Level 2  
rc use frontend-main/web-apps/web-project-1  # Level 3

# Go back to root
rc use /

# List children at current level
rc list

# Clone all lazy repos
rc clone --recursive
\`\`\`

## Testing Depth

To verify the tree depth:
1. Run \`rc tree\` to see the full hierarchy
2. Navigate deep: \`rc use frontend-main/web-apps/web-project-1\`
3. Check current location: \`rc current\`
EOF
}



# Function to create test script
create_test_script() {
    local script_path="$BASE_DIR/test-repo-claude.sh"
    
    cat > "$script_path" <<EOF
#!/bin/bash

# Test script for repo-claude (tree already built)

# Get absolute path to rc binary
if [ -f "$RC_BIN" ]; then
    RC_BIN="$RC_BIN"
elif [ -f "$PROJECT_DIR/bin/rc" ]; then
    RC_BIN="$PROJECT_DIR/bin/rc"
else
    echo "Error: rc binary not found"
    echo "Please set RC_BIN environment variable or build with 'make build'"
    exit 1
fi

TEST_DIR="\$(dirname "\$0")"

echo "Testing repo-claude with pre-built tree"
echo "==========================================="
echo

# Function to run and check command
run_test() {
    local cmd="\$1"
    local expected="\$2"
    
    echo "Running: \$cmd"
    output=\$(\$cmd 2>&1)
    
    if [[ "\$output" == *"\$expected"* ]]; then
        echo "✓ Success"
    else
        echo "✗ Failed - expected '\$expected'"
        echo "Output: \$output"
    fi
    echo
}

# Change to workspace
cd "\$TEST_DIR/workspace"

# Test 1: Show tree structure
echo "Test 1: Display tree structure"
run_test "\$RC_BIN tree" "/"

# Test 2: Navigate to frontend
echo "Test 2: Navigate to frontend-main"
run_test "\$RC_BIN use frontend-main" "Target"
run_test "\$RC_BIN list" "web-"

# Test 3: Navigate deeper
echo "Test 3: Navigate to nested repo"
first_web=\$(ls -d frontend-main/web-* 2>/dev/null | head -1)
if [ -n "\$first_web" ]; then
    run_test "\$RC_BIN use \$first_web" "Target"
fi

# Test 4: Go back to root
echo "Test 4: Navigate back to root"
run_test "\$RC_BIN use /" "Target: /"

# Test 5: Navigate to backend
echo "Test 5: Navigate to backend-main"
run_test "\$RC_BIN use backend-main" "Target"
run_test "\$RC_BIN list" "service-"

# Test 6: Clone lazy repositories
echo "Test 6: Clone lazy repositories"
run_test "\$RC_BIN clone --recursive" "Cloning"

# Test 7: Git status at root
echo "Test 7: Git operations"
run_test "\$RC_BIN use /" "Target"
run_test "\$RC_BIN status" "Status"

# Test 8: Show current position
echo "Test 8: Show current position"
run_test "\$RC_BIN current" "Current"

echo "Testing complete!"
echo
echo "Manual exploration commands:"
echo "  cd \$TEST_DIR/workspace"
echo "  \$RC_BIN tree           # Show full tree"
echo "  \$RC_BIN use <path>     # Navigate"
echo "  \$RC_BIN list           # List children"
echo "  \$RC_BIN current        # Show position"
EOF
    
    chmod +x "$script_path"
    echo -e "${GREEN}✓ Created test script at $script_path${NC}"
}

# Function to create performance test
create_performance_test() {
    local script_path="$BASE_DIR/performance-test.sh"
    
    cat > "$script_path" <<'EOF'
#!/bin/bash

# Performance test for repo-claude

RC_BIN="${RC_BIN:-rc}"
TEST_DIR="$(dirname "$0")"

echo "Performance Testing repo-claude"
echo "=================================="

cd "$TEST_DIR/workspace"

# Measure initialization time
echo "1. Initialization performance:"
time $RC_BIN init -n perf-test

# Measure adding many repos
echo "2. Adding 50 repositories:"
time for i in {1..50}; do
    $RC_BIN add "file://$TEST_DIR/repo-pool/repo-$i" --name "repo-$i" --lazy 2>/dev/null
done

# Measure tree display
echo "3. Tree display performance:"
time $RC_BIN tree

# Measure navigation
echo "4. Navigation performance (10 navigations):"
time for i in {1..10}; do
    $RC_BIN use / 2>/dev/null
done

# Measure git operations
echo "5. Git status across tree:"
time $RC_BIN status

echo "Performance testing complete!"
EOF
    
    chmod +x "$script_path"
}

# Main execution
main() {
    # Create base directories
    echo -e "${YELLOW}🔧 Creating test environment...${NC}"
    mkdir -p "$BASE_DIR/workspace" || {
        echo -e "${RED}Error: Could not create base directory${NC}"
        exit 1
    }
    mkdir -p "$BASE_DIR/repo-pool"
    
    # Generate flat repository pool
    echo
    echo -e "${YELLOW}📦 Creating repository pool...${NC}"
    generate_repo_pool "$BASE_DIR/repo-pool"
    
    # Build actual tree with rc commands
    echo
    echo -e "${YELLOW}🌳 Building tree structure with rc...${NC}"
    build_tree_with_rc "$BASE_DIR/workspace" "$BASE_DIR/repo-pool"
    
    # Create test scripts
    echo -e "${YELLOW}📝 Creating test scripts...${NC}"
    create_test_script
    create_performance_test
    
    # Create usage instructions
    cat > "$BASE_DIR/README.md" <<EOF
# Repo-Claude Test Tree

This directory contains a test environment for repo-claude.

## Structure

- \`repo-pool/\`: Pool of git repositories ($TOTAL_REPOS total)
- \`workspace/\`: repo-claude workspace
- \`test-repo-claude.sh\`: Basic test script
- \`performance-test.sh\`: Performance test script

## Usage

### Manual Testing

\`\`\`bash
cd workspace
rc init test-tree

# Add repos from pool
rc add file://../repo-pool/frontend-root-L1-1 --name frontend
rc add file://../repo-pool/backend-root-L1-2 --name backend --lazy

# Navigate
rc use frontend
rc tree
\`\`\`

### Automated Testing

\`\`\`bash
./test-repo-claude.sh
\`\`\`

### Performance Testing

\`\`\`bash
./performance-test.sh
\`\`\`

## Repository Types

Each repository is tagged with a type:
- frontend: React/TypeScript apps
- backend: Go services  
- service: Microservices
- library: Shared libraries
- data: Data pipelines
- ml: ML models
- docs: Documentation
- infra: IaC
- tools: CLI tools

## Statistics

- Total repositories: $TOTAL_REPOS
- Repository pool: Flat structure
- Meta tree depth: $TREE_DEPTH
- Architecture: Tree-based navigation
EOF
    
    echo
    echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}✅ TEST TREE GENERATION COMPLETE!${NC}"
    echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
    echo
    echo -e "${YELLOW}📊 Summary:${NC}"
    echo -e "   • Test Directory: ${BLUE}$BASE_DIR${NC}"
    echo -e "   • Total Repositories: ${GREEN}$TOTAL_REPOS${NC}"
    echo -e "   • Repository Pool: ${GREEN}Flat structure${NC}"
    echo -e "   • Meta Tree Depth: ${GREEN}$TREE_DEPTH levels${NC}"
    echo
    echo -e "${YELLOW}🚀 Quick Start:${NC}"
    echo -e "   ${CYAN}cd $BASE_DIR/workspace${NC}"
    echo -e "   ${CYAN}\$RC_BIN init test-tree${NC}"
    echo -e "   ${CYAN}\$RC_BIN add file://$BASE_DIR/repo-pool/<repo-name> --name <name>${NC}"
    echo
    echo -e "${YELLOW}🧪 Run Tests:${NC}"
    echo -e "   ${CYAN}$BASE_DIR/test-repo-claude.sh${NC}"
    echo
    echo -e "${CYAN}────────────────────────────────────────────────────────────────${NC}"
}

# Trap errors to show where it failed
trap 'echo -e "${RED}Error occurred at line $LINENO in function ${FUNCNAME[0]:-main}${NC}"; exit 1' ERR

# Enable debugging mode if DEBUG is set
[ -n "${DEBUG:-}" ] && set -x

# Run main function
main "$@"