# 🐙 MUNO Repository Rename Instructions

## Overview
This document provides instructions for completing the rename of the GitHub repository from `repo-claude` to `muno`.

## ✅ Completed Steps (Local)

1. **Code Rebranding** ✅
   - Binary renamed: `rc` → `muno`
   - Module renamed: `github.com/taokim/repo-claude` → `github.com/taokim/muno`
   - All documentation updated with MUNO branding
   - Octopus logo added as mascot

2. **Local Git Configuration** ✅
   - Remote URL updated to `git@github.com:taokim/muno.git`
   - All changes committed with comprehensive message

3. **GitHub Metadata** ✅
   - Issue templates created
   - Pull request template added
   - Funding configuration added
   - Package.json for better GitHub integration

## 📋 Steps to Complete on GitHub

### Option 1: Using the Provided Script (Recommended)

Run the GitHub rename script:
```bash
./scripts/github-rename.sh
```

This script will:
- Rename the repository on GitHub
- Update description and topics
- Push metadata files

### Option 2: Manual Steps

1. **Rename Repository on GitHub**
   - Go to https://github.com/taokim/repo-claude/settings
   - In "Repository name" field, change `repo-claude` to `muno`
   - Click "Rename"

2. **Update Repository Description**
   ```
   🐙 MUNO - Multi-repository UNified Orchestration. Manage multiple git repositories with monorepo-like convenience.
   ```

3. **Add Repository Topics**
   - multi-repo
   - monorepo
   - git
   - orchestration
   - developer-tools
   - cli
   - golang
   - muno
   - repository-management
   - devops

4. **Push Local Changes**
   ```bash
   git push origin main
   ```

## 🔄 After Rename Checklist

- [ ] Repository successfully renamed on GitHub
- [ ] Can access at https://github.com/taokim/muno
- [ ] Description updated with MUNO branding
- [ ] Topics added for better discoverability
- [ ] All local changes pushed
- [ ] GitHub Pages (if any) still working
- [ ] CI/CD pipelines updated (if any)
- [ ] Update any external links to the repository
- [ ] Notify team members/users of the new location

## 📢 Announcement Template

```markdown
## 🐙 repo-claude is now MUNO!

We're excited to announce that repo-claude has been rebranded to **MUNO** (Multi-repository UNified Orchestration).

### What's Changed?
- New name: MUNO (inspired by "MUsinsa moNOrepo")
- New command: `muno` (previously `rc`)
- New mascot: An intelligent octopus 🐙
- New repository: https://github.com/taokim/muno

### Why the Change?
MUNO better represents our mission: bringing monorepo-like convenience to multi-repository projects. The octopus mascot symbolizes intelligent coordination of multiple repositories.

### For Existing Users
1. Update your remote: `git remote set-url origin git@github.com:taokim/muno.git`
2. Reinstall with new name: `make clean && make install`
3. Use `muno` command instead of `rc`

The functionality remains the same - just with a better name and clearer identity!
```

## 🚀 Next Steps

1. Complete the GitHub rename using one of the options above
2. Test that everything works at the new URL
3. Consider setting up a redirect from old repository (GitHub does this automatically for a while)
4. Update any documentation or wikis that reference the old name
5. Announce the change to users/contributors

## 📝 Notes

- GitHub will automatically redirect `repo-claude` to `muno` for a period of time
- All existing issues, PRs, and stars will be preserved
- Git history remains intact
- Forks will continue to work but should update their upstream remote

---

Generated: $(date)
Ready for: GitHub Repository Rename