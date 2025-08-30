#!/bin/bash

# Advanced Testing Framework for repo-claude v3
# Simulates real-world scenarios with complex tree structures

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RC_BIN="${RC_BIN:-$PROJECT_DIR/bin/rc}"
TEST_BASE="${TEST_BASE:-/tmp/rc-advanced-test}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Test scenarios configuration (bash 3 compatible)
SCENARIOS="microservices:Microservices architecture with shared libraries
monorepo:Large monorepo with multiple applications
documentation:Documentation-heavy project with wikis
ml_pipeline:Machine learning pipeline with data repos
enterprise:Enterprise multi-team structure"

# Function to create microservices scenario
create_microservices_scenario() {
    local base="$1"
    echo -e "${CYAN}Creating Microservices Scenario${NC}"
    
    mkdir -p "$base"
    cd "$base"
    
    # Create service repositories
    local services=("auth-service" "user-service" "payment-service" "notification-service" "gateway")
    local libs=("common-lib" "proto-lib" "config-lib")
    local infra=("kubernetes-configs" "terraform-aws" "monitoring")
    
    # Create shared libraries first
    for lib in "${libs[@]}"; do
        create_repo "$base/$lib" "library" <<EOF
# $lib
Shared library for microservices

## Usage
\`\`\`go
import "github.com/company/$lib"
\`\`\`
EOF
        
        # Add go.mod
        cat > "$base/$lib/go.mod" <<EOF
module github.com/company/$lib

go 1.21

require (
    github.com/stretchr/testify v1.8.0
    google.golang.org/grpc v1.50.0
)
EOF
    done
    
    # Create services with dependencies
    for service in "${services[@]}"; do
        create_repo "$base/$service" "service" <<EOF
# $service
Microservice component

## Dependencies
- common-lib
- proto-lib
- config-lib
EOF
        
        # Add service-specific files
        cat > "$base/$service/main.go" <<EOF
package main

import (
    "fmt"
    "github.com/company/common-lib"
    "github.com/company/proto-lib"
)

func main() {
    fmt.Println("Starting $service")
}
EOF
        
        # Add Dockerfile
        cat > "$base/$service/Dockerfile" <<EOF
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o $service

FROM alpine:latest
COPY --from=builder /app/$service /
CMD ["/$service"]
EOF
    done
    
    # Create infrastructure repos
    for infra_repo in "${infra[@]}"; do
        create_repo "$base/$infra_repo" "infra"
        
        if [[ "$infra_repo" == "kubernetes-configs" ]]; then
            # Add k8s manifests
            for service in "${services[@]}"; do
                cat > "$base/$infra_repo/$service-deployment.yaml" <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: $service
  template:
    metadata:
      labels:
        app: $service
    spec:
      containers:
      - name: $service
        image: company/$service:latest
EOF
            done
        fi
    done
    
    # Create repo-claude config for this scenario
    cat > "$base/test-config.yaml" <<EOF
# Microservices Test Configuration
# Use this to test repo-claude with microservices architecture

repos:
  # Core Services
  - url: file://$base/auth-service
    name: auth
    lazy: false
  
  - url: file://$base/user-service
    name: users
    lazy: false
    
  - url: file://$base/payment-service
    name: payments
    lazy: true
    
  - url: file://$base/notification-service
    name: notifications
    lazy: true
    
  - url: file://$base/gateway
    name: gateway
    lazy: false
  
  # Shared Libraries (always eager)
  - url: file://$base/common-lib
    name: common
    lazy: false
    
  - url: file://$base/proto-lib
    name: proto
    lazy: false
    
  - url: file://$base/config-lib
    name: config
    lazy: false
  
  # Infrastructure (lazy by default)
  - url: file://$base/kubernetes-configs
    name: k8s
    lazy: true
    
  - url: file://$base/terraform-aws
    name: terraform
    lazy: true
    
  - url: file://$base/monitoring
    name: monitoring
    lazy: true
EOF
}

# Function to create monorepo scenario
create_monorepo_scenario() {
    local base="$1"
    echo -e "${CYAN}Creating Monorepo Scenario${NC}"
    
    mkdir -p "$base"
    cd "$base"
    
    # Create the monorepo
    create_repo "$base/monorepo" "monorepo"
    
    cd "$base/monorepo"
    
    # Create app directories
    local apps=("web-app" "mobile-app" "admin-portal" "api-server" "worker-service")
    local packages=("ui-components" "utils" "types" "config" "testing")
    
    for app in "${apps[@]}"; do
        mkdir -p "apps/$app"
        cat > "apps/$app/package.json" <<EOF
{
  "name": "@company/$app",
  "version": "1.0.0",
  "private": true,
  "dependencies": {
    "@company/ui-components": "workspace:*",
    "@company/utils": "workspace:*"
  }
}
EOF
    done
    
    for pkg in "${packages[@]}"; do
        mkdir -p "packages/$pkg"
        cat > "packages/$pkg/package.json" <<EOF
{
  "name": "@company/$pkg",
  "version": "1.0.0",
  "main": "dist/index.js"
}
EOF
    done
    
    # Create root package.json with workspaces
    cat > package.json <<EOF
{
  "name": "company-monorepo",
  "private": true,
  "workspaces": [
    "apps/*",
    "packages/*"
  ],
  "scripts": {
    "build": "turbo run build",
    "test": "turbo run test"
  }
}
EOF
    
    # Add turbo.json
    cat > turbo.json <<EOF
{
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**"]
    },
    "test": {
      "dependsOn": ["build"]
    }
  }
}
EOF
    
    git add . && git commit -m "Add monorepo structure"
    
    # Create additional tool repos
    create_repo "$base/dev-tools" "tools"
    create_repo "$base/ci-cd" "infra"
    create_repo "$base/docs" "docs"
}

# Function to create ML pipeline scenario
create_ml_pipeline_scenario() {
    local base="$1"
    echo -e "${CYAN}Creating ML Pipeline Scenario${NC}"
    
    mkdir -p "$base"
    
    # Create data repos
    local data_repos=("raw-data" "processed-data" "features" "training-data")
    for repo in "${data_repos[@]}"; do
        create_repo "$base/$repo" "data"
        
        # Add data pipeline config
        cat > "$base/$repo/pipeline.yaml" <<EOF
name: $repo
type: data-pipeline
steps:
  - name: extract
    source: s3://bucket/$repo
  - name: transform
    script: transform.py
  - name: load
    destination: s3://bucket/processed/
EOF
    done
    
    # Create model repos
    local models=("classification-model" "regression-model" "nlp-model" "vision-model")
    for model in "${models[@]}"; do
        create_repo "$base/$model" "ml"
        
        # Add model files
        cat > "$base/$model/model.py" <<EOF
import tensorflow as tf
import numpy as np

class ${model//-/_}:
    def __init__(self):
        self.model = self.build_model()
    
    def build_model(self):
        return tf.keras.Sequential([
            tf.keras.layers.Dense(128, activation='relu'),
            tf.keras.layers.Dropout(0.2),
            tf.keras.layers.Dense(10, activation='softmax')
        ])
    
    def train(self, X, y):
        self.model.compile(optimizer='adam',
                          loss='categorical_crossentropy',
                          metrics=['accuracy'])
        return self.model.fit(X, y, epochs=10)
EOF
    done
    
    # Create experiment tracking repo
    create_repo "$base/experiments" "ml"
    cat > "$base/experiments/mlflow.yaml" <<EOF
tracking_uri: http://mlflow:5000
artifact_location: s3://mlflow-artifacts
experiments:
  - name: baseline
  - name: hyperparameter-tuning
  - name: feature-engineering
EOF
    
    # Create serving repos
    create_repo "$base/model-serving" "service"
    create_repo "$base/api-gateway" "service"
}

# Function to create enterprise scenario
create_enterprise_scenario() {
    local base="$1"
    echo -e "${CYAN}Creating Enterprise Scenario${NC}"
    
    mkdir -p "$base"
    
    # Create team structures
    local teams=("platform" "frontend" "backend" "data" "security" "devops")
    
    for team in "${teams[@]}"; do
        mkdir -p "$base/$team"
        
        # Each team has multiple repos
        case $team in
            platform)
                create_repo "$base/$team/core-platform" "library"
                create_repo "$base/$team/sdk" "library"
                create_repo "$base/$team/cli-tools" "tools"
                ;;
            frontend)
                create_repo "$base/$team/web-app" "frontend"
                create_repo "$base/$team/mobile-app" "mobile"
                create_repo "$base/$team/design-system" "frontend"
                ;;
            backend)
                create_repo "$base/$team/api-gateway" "backend"
                create_repo "$base/$team/core-services" "backend"
                create_repo "$base/$team/batch-jobs" "backend"
                ;;
            data)
                create_repo "$base/$team/data-lake" "data"
                create_repo "$base/$team/analytics" "data"
                create_repo "$base/$team/reporting" "data"
                ;;
            security)
                create_repo "$base/$team/auth-service" "service"
                create_repo "$base/$team/vault-config" "infra"
                create_repo "$base/$team/security-tools" "tools"
                ;;
            devops)
                create_repo "$base/$team/infrastructure" "infra"
                create_repo "$base/$team/ci-cd" "infra"
                create_repo "$base/$team/monitoring" "infra"
                ;;
        esac
    done
    
    # Create cross-team shared repos
    create_repo "$base/shared/contracts" "docs"
    create_repo "$base/shared/proto" "library"
    create_repo "$base/shared/configs" "infra"
}

# Helper function to create a git repo
create_repo() {
    local path="$1"
    local type="${2:-default}"
    local readme_content="${3:-# $(basename $path)\n\nRepository of type: $type}"
    
    mkdir -p "$path"
    cd "$path"
    
    git init --quiet
    
    # Create README
    echo -e "$readme_content" > README.md
    
    # Add type-specific files
    case $type in
        library|service|backend)
            echo "package $(basename $path)" > main.go
            ;;
        frontend|mobile)
            echo '{"name": "'$(basename $path)'"}' > package.json
            ;;
        data|ml)
            echo "# $(basename $path)" > notebook.ipynb
            ;;
        infra)
            echo "# Infrastructure" > terraform.tf
            ;;
        tools)
            echo "#!/bin/bash" > tool.sh
            chmod +x tool.sh
            ;;
    esac
    
    git add . >/dev/null 2>&1
    git config user.name "Test" >/dev/null 2>&1
    git config user.email "test@test.com" >/dev/null 2>&1
    git commit -m "Initial commit" --quiet
    
    echo -e "${GREEN}  ✓ Created: $(basename $path) ($type)${NC}"
}

# Function to run scenario tests
run_scenario_tests() {
    local scenario_dir="$1"
    local scenario_name="$2"
    
    echo -e "${YELLOW}Testing $scenario_name scenario${NC}"
    
    # Create test workspace
    local workspace="$scenario_dir/rc-workspace"
    mkdir -p "$workspace"
    cd "$workspace"
    
    # Initialize repo-claude
    echo -e "${BLUE}  Initializing workspace...${NC}"
    $RC_BIN init -n "$scenario_name-test" >/dev/null 2>&1 || true
    
    # Change to the actual workspace directory
    cd "$workspace/$scenario_name-test"
    
    # Add repositories based on scenario
    echo -e "${BLUE}  Adding repositories...${NC}"
    
    case $scenario_name in
        microservices)
            # Add services and libraries
            for repo in "$scenario_dir"/{*-service,*-lib}; do
                [ -d "$repo/.git" ] && {
                    name=$(basename "$repo")
                    $RC_BIN add "file://$repo" --name "$name" $([ "${name##*-}" = "lib" ] && echo "" || echo "--lazy") 2>/dev/null || true
                }
            done
            ;;
        monorepo)
            # Add the monorepo and tools
            $RC_BIN add "file://$scenario_dir/monorepo" --name "mono" 2>/dev/null || true
            $RC_BIN add "file://$scenario_dir/dev-tools" --name "tools" --lazy 2>/dev/null || true
            ;;
        ml_pipeline)
            # Add data and model repos
            for repo in "$scenario_dir"/*; do
                [ -d "$repo/.git" ] && {
                    name=$(basename "$repo")
                    lazy_flag=$([ "${name%%-*}" = "raw" ] && echo "" || echo "--lazy")
                    $RC_BIN add "file://$repo" --name "$name" $lazy_flag 2>/dev/null || true
                }
            done
            ;;
        enterprise)
            # Add team repos
            for team in "$scenario_dir"/*; do
                [ -d "$team" ] && {
                    for repo in "$team"/*; do
                        [ -d "$repo/.git" ] && {
                            team_name=$(basename "$team")
                            repo_name=$(basename "$repo")
                            $RC_BIN add "file://$repo" --name "${team_name}-${repo_name}" --lazy 2>/dev/null || true
                        }
                    done
                }
            done
            ;;
    esac
    
    # Run tests
    echo -e "${BLUE}  Running tests...${NC}"
    
    # Test tree display
    echo -n "    Tree display: "
    if $RC_BIN tree >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
    fi
    
    # Test navigation
    echo -n "    Navigation: "
    if $RC_BIN use / >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
    fi
    
    # Test status
    echo -n "    Status check: "
    if $RC_BIN status 2>&1 | grep -q "Status"; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
    fi
    
    echo
}

# Main execution
main() {
    echo -e "${MAGENTA}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${MAGENTA}║     Advanced Testing Framework for RC v3      ║${NC}"
    echo -e "${MAGENTA}╚════════════════════════════════════════════════╝${NC}"
    echo
    
    # Check if rc binary exists
    if [ ! -f "$RC_BIN" ]; then
        echo -e "${RED}Error: rc binary not found at $RC_BIN${NC}"
        echo "Please build it first: make build"
        exit 1
    fi
    
    # Create test directory
    TEST_DIR="$TEST_BASE/$TIMESTAMP"
    mkdir -p "$TEST_DIR"
    
    echo -e "${YELLOW}════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}TEST DIRECTORY: $TEST_DIR${NC}"
    echo -e "${YELLOW}════════════════════════════════════════════════${NC}"
    echo
    
    # Create scenarios
    echo "$SCENARIOS" | while IFS=: read -r scenario description; do
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${CYAN}Scenario: $scenario${NC}"
        echo -e "${CYAN}Description: $description${NC}"
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        
        scenario_dir="$TEST_DIR/$scenario"
        
        # Create scenario
        case $scenario in
            microservices)
                create_microservices_scenario "$scenario_dir"
                ;;
            monorepo)
                create_monorepo_scenario "$scenario_dir"
                ;;
            ml_pipeline)
                create_ml_pipeline_scenario "$scenario_dir"
                ;;
            enterprise)
                create_enterprise_scenario "$scenario_dir"
                ;;
            *)
                echo -e "${YELLOW}  Scenario not implemented yet${NC}"
                continue
                ;;
        esac
        
        # Run tests for this scenario
        run_scenario_tests "$scenario_dir" "$scenario"
    done
    
    # Create summary report
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}Test Summary${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    cat > "$TEST_DIR/SUMMARY.md" <<EOF
# Advanced Testing Results

**Date**: $(date)
**Test Directory**: $TEST_DIR

## Scenarios Tested

$(echo "$SCENARIOS" | while IFS=: read -r scenario description; do
    echo "### $scenario"
    echo "$description"
    echo "- Location: $TEST_DIR/$scenario"
    echo
done)

## Next Steps

1. Manual testing:
   \`\`\`bash
   cd $TEST_DIR/<scenario>/rc-workspace
   $RC_BIN tree
   $RC_BIN use <node>
   \`\`\`

2. Performance testing:
   \`\`\`bash
   cd $TEST_DIR/microservices/rc-workspace
   time $RC_BIN tree
   \`\`\`

3. Stress testing:
   \`\`\`bash
   # Add 100+ repos
   for i in {1..100}; do
     $RC_BIN add "file:///tmp/repo-\$i" --name "repo-\$i" --lazy
   done
   \`\`\`
EOF
    
    echo -e "${GREEN}════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}ALL TESTS COMPLETE!${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════${NC}"
    echo
    echo -e "${YELLOW}TEST DIRECTORY:${NC}"
    echo -e "  ${CYAN}$TEST_DIR${NC}"
    echo
    echo -e "${YELLOW}Results saved to:${NC}"
    echo -e "  ${CYAN}$TEST_DIR/SUMMARY.md${NC}"
    echo
    echo -e "${YELLOW}To explore a specific scenario:${NC}"
    echo -e "  ${CYAN}cd $TEST_DIR/microservices${NC}"
    echo -e "  ${CYAN}cd $TEST_DIR/monorepo${NC}"
    echo -e "  ${CYAN}cd $TEST_DIR/enterprise${NC}"
    echo
    echo -e "${YELLOW}To test repo-claude with a scenario:${NC}"
    echo -e "  ${CYAN}cd $TEST_DIR/<scenario>/rc-workspace/<scenario>-test${NC}"
    echo -e "  ${CYAN}$RC_BIN tree${NC}"
}

# Parse arguments
case "${1:-}" in
    microservices|monorepo|ml_pipeline|enterprise)
        # Run specific scenario
        TEST_DIR="$TEST_BASE/single-$TIMESTAMP"
        mkdir -p "$TEST_DIR"
        scenario_dir="$TEST_DIR/$1"
        
        echo -e "${CYAN}Running single scenario: $1${NC}"
        echo -e "${YELLOW}════════════════════════════════════════════════${NC}"
        echo -e "${YELLOW}TEST DIRECTORY: $TEST_DIR${NC}"
        echo -e "${YELLOW}════════════════════════════════════════════════${NC}"
        echo
        create_${1}_scenario "$scenario_dir"
        run_scenario_tests "$scenario_dir" "$1"
        
        # Show completion message with paths
        echo
        echo -e "${GREEN}════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}TEST COMPLETE!${NC}"
        echo -e "${GREEN}════════════════════════════════════════════════${NC}"
        echo -e "${YELLOW}Test files created at:${NC}"
        echo -e "  ${CYAN}$scenario_dir${NC}"
        echo -e "${YELLOW}Workspace created at:${NC}"
        echo -e "  ${CYAN}$scenario_dir/rc-workspace/$1-test${NC}"
        echo
        echo -e "${YELLOW}To explore the test:${NC}"
        echo -e "  ${CYAN}cd $scenario_dir/rc-workspace/$1-test${NC}"
        echo -e "  ${CYAN}$RC_BIN tree${NC}"
        ;;
    *)
        # Run all scenarios
        main
        ;;
esac