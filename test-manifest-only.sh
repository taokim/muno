#!/bin/bash
# Test manifest repository creation

set -e

echo "Testing manifest creation..."

# Create test directory
mkdir -p test-manifest-check
cd test-manifest-check

# Create manifest repo manually
echo "Creating manifest repo..."
mkdir -p .manifest-repo
cd .manifest-repo
git init
git config user.email "test@example.com"
git config user.name "Test"

# Create manifest
cat > default.xml << EOF
<?xml version="1.0" encoding="UTF-8"?>
<manifest>
  <remote name="origin" fetch="git@github.com:musinsa/"/>
  <default remote="origin" revision="main" sync-j="4"/>
  <project name="fse-root-repo" path="backend"/>
  <project name="fse-root-repo" path="frontend"/>
</manifest>
EOF

git add default.xml
git commit -m "Initial manifest"
git branch -M main

echo "Manifest repo created. Checking..."
echo "Current branch: $(git branch --show-current)"
echo "Commits:"
git log --oneline

cd ..

echo ""
echo "Testing repo init..."
repo init -u file://$(pwd)/.manifest-repo -b main

echo ""
echo "✅ Repo init succeeded!"

echo ""
echo "Testing repo sync..."
repo sync || echo "Sync failed (expected if repos don't exist yet)"

echo ""
echo "Checking .repo structure..."
ls -la .repo/

echo "✅ Test complete!"