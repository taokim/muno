#!/bin/bash
# test-workflow.sh - Test basic repo-claude workflow

set -e

echo "🧪 Testing repo-claude workflow..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}❌ $1 is not installed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ $1 found${NC}"
}

echo "Checking prerequisites..."
check_tool git
check_tool repo
check_tool go

# Build repo-claude
echo -e "\n📦 Building repo-claude..."
cd "$(dirname "$0")/.."
make clean build

# Create test directory
TEST_DIR=$(mktemp -d)
echo -e "\n📁 Test directory: $TEST_DIR"

# Copy binary
cp bin/repo-claude "$TEST_DIR/"
cd "$TEST_DIR"

# Test init command
echo -e "\n🚀 Testing init command..."
./repo-claude init test-project --non-interactive

# Check created files
echo -e "\n📋 Checking created files..."
files=(
    "test-project/repo-claude"
    "test-project/repo-claude.yaml"
    "test-project/shared-memory.md"
    "test-project/.manifest-repo/default.xml"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $file exists${NC}"
    else
        echo -e "${RED}❌ $file missing${NC}"
        exit 1
    fi
done

# Test status command
echo -e "\n📊 Testing status command..."
cd test-project
./repo-claude status

# Test configuration
echo -e "\n🔧 Checking configuration..."
if grep -q "test-project" repo-claude.yaml; then
    echo -e "${GREEN}✅ Configuration contains project name${NC}"
else
    echo -e "${RED}❌ Configuration missing project name${NC}"
    exit 1
fi

# Check manifest
echo -e "\n📄 Checking manifest..."
if grep -q "<manifest>" .manifest-repo/default.xml; then
    echo -e "${GREEN}✅ Valid manifest created${NC}"
else
    echo -e "${RED}❌ Invalid manifest${NC}"
    exit 1
fi

# Cleanup
cd ..
rm -rf "$TEST_DIR"

echo -e "\n${GREEN}🎉 All tests passed!${NC}"