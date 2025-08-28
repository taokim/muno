#!/bin/bash

# Generate Test Tree Script for repo-claude v3
# Creates a realistic multi-depth tree structure with local git repositories
# Usage: ./generate-test-tree.sh [base_dir] [depth] [repos_per_level]

set -e

# Configuration
BASE_DIR="${1:-/tmp/repo-claude-test}"
MAX_DEPTH="${2:-5}"
REPOS_PER_LEVEL="${3:-3}"
TOTAL_REPOS=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
    
    echo -e "${BLUE}Creating repository: $repo_name ($repo_type)${NC}"
    
    # Create directory
    mkdir -p "$repo_path"
    cd "$repo_path"
    
    # Initialize git
    git init --quiet
    
    # Create appropriate files based on type
    case "$repo_type" in
        frontend)
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
            mkdir -p src/components
            echo "export const Button = () => 'Button';" > src/components/Button.tsx
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
            mkdir -p internal/api
            echo "package api" > internal/api/server.go
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
            mkdir -p lib
            echo "// Shared utilities" > lib/utils.go
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
            mkdir -p sql
            echo "SELECT * FROM $repo_name;" > sql/query.sql
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
            echo "# Architecture" > architecture.md
            echo "# API Reference" > api.md
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
    git config user.name "Test User" >/dev/null 2>&1
    git config user.email "test@example.com" >/dev/null 2>&1
    git commit -m "Initial commit for $repo_name" --quiet
    
    # Add some history
    echo "# Updates" >> README.md
    git add README.md >/dev/null 2>&1
    git commit -m "Add updates section" --quiet
    
    # Create a feature branch
    git checkout -b feature/test --quiet
    echo "Feature work" > feature.txt
    git add feature.txt >/dev/null 2>&1
    git commit -m "Add feature" --quiet
    
    # Switch back to main
    git checkout main --quiet 2>/dev/null || git checkout master --quiet
    
    ((TOTAL_REPOS++))
}

# Function to generate tree structure recursively
generate_tree_level() {
    local parent_path="$1"
    local current_depth="$2"
    local parent_name="$3"
    
    if [ "$current_depth" -gt "$MAX_DEPTH" ]; then
        return
    fi
    
    echo -e "${GREEN}Generating level $current_depth under $parent_name${NC}"
    
    for i in $(seq 1 $REPOS_PER_LEVEL); do
        # Select random repo type
        local type_index=$((RANDOM % ${#REPO_TYPES[@]}))
        local type_spec="${REPO_TYPES[$type_index]}"
        local type="${type_spec%%:*}"
        
        # Generate repo name
        local repo_name="${type}-${parent_name}-L${current_depth}-${i}"
        local repo_path="$parent_path/$repo_name"
        
        # Create the repository
        create_git_repo "$repo_path" "$type"
        
        # Recursively create children (with decreasing probability)
        if [ $((RANDOM % 100)) -lt $((100 - current_depth * 20)) ]; then
            generate_tree_level "$repo_path" $((current_depth + 1)) "$repo_name"
        fi
    done
}

# Function to create repo-claude configuration
create_repo_claude_config() {
    local workspace_dir="$1"
    
    echo -e "${YELLOW}Creating repo-claude workspace configuration${NC}"
    
    cd "$workspace_dir"
    
    # Initialize repo-claude workspace
    cat > repo-claude.yaml <<EOF
version: 3
workspace:
  name: test-tree
  root_repo: ""
EOF
    
    # Create shared memory
    cat > shared-memory.md <<EOF
# Test Tree Shared Memory

This is a test workspace with $TOTAL_REPOS repositories organized in a tree structure.

## Structure
- Maximum depth: $MAX_DEPTH
- Repositories per level: $REPOS_PER_LEVEL
- Total repositories: $TOTAL_REPOS

## Repository Types
- frontend: React/TypeScript applications
- backend: Go services
- service: Dockerized microservices
- library: Shared libraries
- data: Data pipelines
- ml: Machine learning models
- docs: Documentation
- infra: Infrastructure as Code
- tools: CLI tools and scripts

## Testing Scenarios
1. Navigate deep into the tree
2. Clone lazy repositories
3. Perform git operations across multiple repos
4. Test cross-repository dependencies
EOF
    
    # Create CLAUDE.md
    cat > CLAUDE.md <<EOF
# Test Tree Workspace

This is a test workspace for repo-claude v3 tree-based architecture.

## Key Features to Test

1. **Deep Navigation**: Navigate through $MAX_DEPTH levels
2. **Lazy Loading**: Some repos are marked as lazy
3. **Git Operations**: All repos have git history
4. **Cross-References**: Repos can reference each other

## Commands to Try

\`\`\`bash
# Initialize the tree
rc init test-tree

# Add repositories from the test pool
rc add file://$BASE_DIR/repo-pool/frontend-root-L1-1 --name frontend
rc add file://$BASE_DIR/repo-pool/backend-root-L1-2 --name backend --lazy

# Navigate the tree
rc use frontend
rc tree
rc list

# Clone lazy repos
rc clone --recursive

# Git operations
rc pull /
rc status
\`\`\`
EOF
}

# Function to create a pool of repositories
create_repo_pool() {
    local pool_dir="$BASE_DIR/repo-pool"
    
    echo -e "${YELLOW}Creating repository pool at $pool_dir${NC}"
    mkdir -p "$pool_dir"
    
    # Generate the tree starting from root
    generate_tree_level "$pool_dir" 1 "root"
    
    echo -e "${GREEN}Created $TOTAL_REPOS repositories${NC}"
}

# Function to create test script
create_test_script() {
    local script_path="$BASE_DIR/test-repo-claude.sh"
    
    cat > "$script_path" <<'EOF'
#!/bin/bash

# Test script for repo-claude v3

RC_BIN="${RC_BIN:-rc}"
TEST_DIR="$(dirname "$0")"

echo "Testing repo-claude v3 with generated tree"
echo "========================================="

# Function to run and check command
run_test() {
    local cmd="$1"
    local expected="$2"
    
    echo "Running: $cmd"
    output=$($cmd 2>&1)
    
    if [[ "$output" == *"$expected"* ]]; then
        echo "✓ Success"
    else
        echo "✗ Failed - expected '$expected'"
        echo "Output: $output"
    fi
    echo
}

# Change to workspace
cd "$TEST_DIR/workspace"

# Test 1: Initialize workspace
echo "Test 1: Initialize workspace"
run_test "$RC_BIN init test-tree" "initialized"

# Test 2: Add repositories with different depths
echo "Test 2: Add repositories"
for repo in "$TEST_DIR/repo-pool"/*/; do
    if [ -d "$repo/.git" ]; then
        name=$(basename "$repo")
        # Add first 3 as regular, rest as lazy
        if [[ "$name" == *"L1-1"* ]] || [[ "$name" == *"L1-2"* ]]; then
            run_test "$RC_BIN add file://$repo --name $name" ""
        else
            run_test "$RC_BIN add file://$repo --name $name --lazy" ""
        fi
        
        # Only add first 5 for testing
        count=$((count + 1))
        [ $count -eq 5 ] && break
    fi
done

# Test 3: Navigate tree
echo "Test 3: Tree navigation"
run_test "$RC_BIN tree" "root"
run_test "$RC_BIN list" "Current"

# Test 4: Use node
echo "Test 4: Use node"
run_test "$RC_BIN use /" "Target"

# Test 5: Git status
echo "Test 5: Git operations"
run_test "$RC_BIN status" "Status"

echo "Testing complete!"
EOF
    
    chmod +x "$script_path"
    echo -e "${GREEN}Created test script at $script_path${NC}"
}

# Function to create performance test
create_performance_test() {
    local script_path="$BASE_DIR/performance-test.sh"
    
    cat > "$script_path" <<'EOF'
#!/bin/bash

# Performance test for repo-claude v3

RC_BIN="${RC_BIN:-rc}"
TEST_DIR="$(dirname "$0")"

echo "Performance Testing repo-claude v3"
echo "=================================="

cd "$TEST_DIR/workspace"

# Measure initialization time
echo "1. Initialization performance:"
time $RC_BIN init perf-test

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
    echo -e "${YELLOW}=== Repo-Claude Test Tree Generator ===${NC}"
    echo -e "${YELLOW}Base directory: $BASE_DIR${NC}"
    echo -e "${YELLOW}Max depth: $MAX_DEPTH${NC}"
    echo -e "${YELLOW}Repos per level: $REPOS_PER_LEVEL${NC}"
    echo
    
    # Clean up existing directory
    if [ -d "$BASE_DIR" ]; then
        echo -e "${RED}Removing existing test directory${NC}"
        rm -rf "$BASE_DIR"
    fi
    
    # Create base directories
    mkdir -p "$BASE_DIR/workspace"
    mkdir -p "$BASE_DIR/repo-pool"
    
    # Create repository pool
    create_repo_pool
    
    # Create workspace configuration
    create_repo_claude_config "$BASE_DIR/workspace"
    
    # Create test scripts
    create_test_script
    create_performance_test
    
    # Create usage instructions
    cat > "$BASE_DIR/README.md" <<EOF
# Repo-Claude Test Tree

This directory contains a test environment for repo-claude v3.

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
- Maximum depth: $MAX_DEPTH
- Repositories per level: $REPOS_PER_LEVEL
EOF
    
    echo
    echo -e "${GREEN}=== Test tree generation complete! ===${NC}"
    echo
    echo -e "${BLUE}Test directory: $BASE_DIR${NC}"
    echo -e "${BLUE}Total repositories created: $TOTAL_REPOS${NC}"
    echo
    echo -e "${YELLOW}To start testing:${NC}"
    echo -e "  cd $BASE_DIR/workspace"
    echo -e "  rc init test-tree"
    echo -e "  rc add file://$BASE_DIR/repo-pool/<repo-name> --name <name>"
    echo
    echo -e "${YELLOW}Or run automated tests:${NC}"
    echo -e "  $BASE_DIR/test-repo-claude.sh"
    echo
}

# Run main function
main