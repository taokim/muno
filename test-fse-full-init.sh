#!/bin/bash
# Test full init workflow with FSE repository

set -e

echo "🧪 Testing full repo-claude init with FSE repository..."

# Build
go build -o repo-claude ./cmd/repo-claude

# Clean up
rm -rf fse-full-test

# Test non-interactive init
echo "Testing non-interactive init..."
./repo-claude init fse-full-test --non-interactive

echo ""
echo "Updating configuration for FSE repository..."
cd fse-full-test

# Update the config to use FSE repository
cat > repo-claude.yaml << EOF
workspace:
  name: fse-full-test
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

# Update manifest
cd .manifest-repo
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
git commit -m "Update manifest for FSE repository"

cd ..

echo ""
echo "🔄 Syncing with FSE repository..."
./repo-claude sync

echo ""
echo "📊 Final status:"
./repo-claude status

echo ""
echo "📁 Checking synced content:"
echo "Backend files:"
ls -la backend/ | head -5
echo ""
echo "Frontend files:"
ls -la frontend/ | head -5

echo ""
echo "✅ Full init test complete!"