#!/bin/bash

# Script to rename GitHub repository from repo-claude to muno
# This script uses GitHub CLI (gh) to update repository settings

set -e

echo "🐙 MUNO GitHub Repository Rename Script"
echo "========================================"
echo ""
echo "This script will help you rename the GitHub repository from 'repo-claude' to 'muno'"
echo "and update all associated metadata."
echo ""

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is not installed."
    echo "Please install it first: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated with GitHub CLI."
    echo "Please run: gh auth login"
    exit 1
fi

echo "⚠️  WARNING: This will rename your GitHub repository!"
echo "Current: github.com/taokim/repo-claude"
echo "New:     github.com/taokim/muno"
echo ""
read -p "Do you want to continue? (y/N): " confirm

if [[ $confirm != [yY] ]]; then
    echo "Cancelled."
    exit 0
fi

echo ""
echo "📝 Step 1: Renaming repository..."
gh repo rename muno --repo taokim/repo-claude --yes || {
    echo "❌ Failed to rename repository. It might already be renamed or you might not have permissions."
    echo "You can manually rename it at: https://github.com/taokim/repo-claude/settings"
    exit 1
}

echo "✅ Repository renamed successfully!"

echo ""
echo "📝 Step 2: Updating repository description..."
gh repo edit taokim/muno \
    --description "🐙 MUNO - Multi-repository UNified Orchestration. Manage multiple git repositories with monorepo-like convenience." \
    --homepage "https://github.com/taokim/muno" || {
    echo "⚠️  Failed to update description. You can do this manually on GitHub."
}

echo ""
echo "📝 Step 3: Updating repository topics..."
gh repo edit taokim/muno \
    --add-topic "multi-repo" \
    --add-topic "monorepo" \
    --add-topic "git" \
    --add-topic "orchestration" \
    --add-topic "developer-tools" \
    --add-topic "cli" \
    --add-topic "golang" \
    --add-topic "muno" \
    --add-topic "repository-management" \
    --add-topic "devops" || {
    echo "⚠️  Failed to update topics. You can do this manually on GitHub."
}

echo ""
echo "📝 Step 4: Creating/Updating GitHub repository metadata files..."

# Create a temporary directory for GitHub metadata
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Clone just the main branch with minimal depth
git clone --depth 1 git@github.com:taokim/muno.git .

# Update or create .github/FUNDING.yml
mkdir -p .github
cat > .github/FUNDING.yml << 'EOF'
# These are supported funding model platforms

github: [taokim]
# patreon: # Replace with a single Patreon username
# open_collective: # Replace with a single Open Collective username
# ko_fi: # Replace with a single Ko-fi username
# tidelift: # Replace with a single Tidelift platform-name/package-name e.g., npm/babel
# community_bridge: # Replace with a single Community Bridge project-name e.g., cloud-foundry
# liberapay: # Replace with a single Liberapay username
# issuehunt: # Replace with a single IssueHunt username
# otechie: # Replace with a single Otechie username
# lfx_crowdfunding: # Replace with a single LFX Crowdfunding project-name e.g., cloud-foundry
# custom: # Replace with up to 4 custom sponsorship URLs e.g., ['link1', 'link2']
EOF

# Create issue templates
mkdir -p .github/ISSUE_TEMPLATE

cat > .github/ISSUE_TEMPLATE/bug_report.md << 'EOF'
---
name: Bug report
about: Create a report to help us improve MUNO
title: '[BUG] '
labels: 'bug'
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. Navigate to '...'
3. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment:**
 - OS: [e.g. macOS, Ubuntu]
 - MUNO Version: [e.g. v0.6.0]
 - Go Version: [e.g. 1.21]

**Additional context**
Add any other context about the problem here.
EOF

cat > .github/ISSUE_TEMPLATE/feature_request.md << 'EOF'
---
name: Feature request
about: Suggest an idea for MUNO
title: '[FEATURE] '
labels: 'enhancement'
assignees: ''

---

**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is. Ex. I'm always frustrated when [...]

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
EOF

# Create pull request template
cat > .github/pull_request_template.md << 'EOF'
## Description
Brief description of the changes in this PR.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Tests pass locally with my changes
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] I have added necessary documentation (if applicable)

## Checklist
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] Any dependent changes have been merged and published

## Related Issues
Closes #(issue number)

## Screenshots (if applicable)
Add screenshots to help explain your changes.
EOF

# Commit and push if there are changes
if [[ -n $(git status -s) ]]; then
    git add .
    git commit -m "chore: add GitHub metadata files for MUNO

- Add funding configuration
- Add issue templates (bug report, feature request)
- Add pull request template
- Update repository branding for MUNO"
    
    git push origin main || {
        echo "⚠️  Failed to push GitHub metadata files. You may need to do this manually."
    }
    echo "✅ GitHub metadata files added!"
else
    echo "ℹ️  GitHub metadata files already exist."
fi

# Clean up
cd - > /dev/null
rm -rf "$TEMP_DIR"

echo ""
echo "🎉 Repository rename complete!"
echo ""
echo "📋 Summary of changes:"
echo "  ✅ Repository renamed to 'muno'"
echo "  ✅ Description updated"
echo "  ✅ Topics added"
echo "  ✅ GitHub metadata files created/updated"
echo ""
echo "📌 Next steps:"
echo "  1. Update any CI/CD pipelines that reference the old repository"
echo "  2. Update any documentation that links to the old repository"
echo "  3. Notify team members and users about the new repository location"
echo "  4. Consider setting up a redirect from the old repository name"
echo ""
echo "🐙 Your MUNO repository is now ready at: https://github.com/taokim/muno"