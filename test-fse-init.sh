#!/bin/bash
# Test script for FSE repository initialization

set -e

echo "🧪 Testing repo-claude with FSE repository..."

# Build if needed
if [ ! -f ./repo-claude ]; then
    echo "Building repo-claude..."
    go build -o repo-claude ./cmd/repo-claude
fi

# Clean up any previous test
rm -rf test-fse-workspace

# Initialize with interactive mode to configure the actual repository
echo "Creating FSE workspace..."
./repo-claude init fse-workspace << EOF
git@github.com:musinsa/
backend
core,services
backend-agent
frontend
core,ui
frontend-agent
mobile
mobile,ui
mobile-agent


n
API development and backend services

Y
React/Vue frontend development

Y
Mobile app development

n
EOF

echo ""
echo "✅ Initialization complete!"
echo ""
echo "Checking created files..."
cd fse-workspace

# Check manifest
echo "📄 Manifest content:"
cat .manifest-repo/default.xml

echo ""
echo "📊 Status check:"
./repo-claude status

echo ""
echo "🔄 Attempting sync..."
./repo-claude sync

echo ""
echo "📁 Repository structure:"
ls -la

echo ""
echo "✅ Test complete!"