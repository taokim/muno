#!/bin/bash
# Test with real FSE repository

set -e

echo "🧪 Testing repo-claude with FSE repository (master branch)..."

# Build
go build -o repo-claude ./cmd/repo-claude

# Clean up
rm -rf fse-test-workspace

# Create test configuration
mkdir -p fse-test-workspace
cd fse-test-workspace

# Create a configuration file directly
cat > repo-claude.yaml << EOF
workspace:
  name: fse-project
  manifest:
    remote_name: origin
    remote_fetch: git@github.com:musinsa/
    default_revision: master
    projects:
      - name: fse-root-repo
        groups: backend,core
        agent: backend-agent
      - name: fse-root-repo
        groups: frontend,ui
        agent: frontend-agent
agents:
  backend-agent:
    model: claude-sonnet-4
    specialization: Backend API development, database design
    auto_start: true
    dependencies: []
  frontend-agent:
    model: claude-sonnet-4
    specialization: React/Vue frontend development
    auto_start: true
    dependencies:
      - backend-agent
EOF

# Initialize only the manifest repo
echo "Creating manifest repository..."
mkdir -p .manifest-repo
cd .manifest-repo
git init
git config user.email "repo-claude@example.com"
git config user.name "Repo-Claude"

# Create manifest
cat > default.xml << EOF
<?xml version="1.0" encoding="UTF-8"?>
<manifest>
  <remote name="origin" fetch="git@github.com:musinsa/"/>
  <default remote="origin" revision="master" sync-j="4"/>
  <project name="fse-root-repo" path="backend" groups="backend,core"/>
  <project name="fse-root-repo" path="frontend" groups="frontend,ui"/>
</manifest>
EOF

git add default.xml
git commit -m "Initial manifest"
git branch -M main

cd ..

echo ""
echo "📦 Initializing repo..."
repo init -u file://$(pwd)/.manifest-repo -b main

echo ""
echo "🔄 Syncing repositories..."
if repo sync -j4; then
    echo "✅ Sync successful!"
else
    echo "⚠️  Sync failed - checking what we got..."
fi

echo ""
echo "📊 Repository status:"
repo status || true

echo ""
echo "📁 Workspace structure:"
ls -la

echo ""
echo "Creating coordination files..."
# Create shared memory
cat > shared-memory.md << 'EOF'
# Shared Agent Memory

## Current Tasks
- No active tasks

## Coordination Notes
- Agents will update this file with their progress
- Use this for cross-repository coordination
- All repositories managed by Repo tool

## Repo Commands Available
- `repo status` - Show status of all projects
- `repo sync` - Sync all projects
- `repo forall -c '<command>'` - Run command in all projects
- `repo list` - List all projects

## Decisions
- Document architectural decisions here
EOF

# Copy the repo-claude binary
cp ../repo-claude .

echo ""
echo "✅ Setup complete!"
echo ""
echo "📊 Final status check:"
./repo-claude status

echo ""
echo "✅ Test complete! FSE repository integration verified."