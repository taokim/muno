#!/bin/bash

# Simple tag creation script for muno
set -e

# Check if version argument is provided
if [ $# -eq 0 ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh 1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.0.0)"
    exit 1
fi

echo "Creating release tag v$VERSION..."

# Create tag
git tag -a "v$VERSION" -m "Release v$VERSION"

# Push tag
echo "Pushing tag..."
git push origin "v$VERSION"

echo "✅ Tag v$VERSION pushed!"
echo ""
echo "GitHub Actions will now:"
echo "1. Run tests"
echo "2. Build binaries for all platforms" 
echo "3. Create a GitHub release (published immediately)"
echo "4. Update the Homebrew tap automatically"
echo ""
echo "Monitor progress at: https://github.com/taokim/muno/actions"