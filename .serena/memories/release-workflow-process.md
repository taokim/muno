# Release Process for MUNO

## IMPORTANT: Never Release Directly!

**CRITICAL**: All releases MUST be done through GitHub Actions workflow, NOT manually.

## Correct Release Process

1. **Create and push a git tag**:
   ```bash
   git tag -a v0.x.x -m "Release v0.x.x with release notes"
   git push origin v0.x.x
   ```

2. **GitHub Actions automatically handles**:
   - Triggers the release.yml workflow on tag push
   - Builds release binaries for all platforms using GoReleaser
   - Creates GitHub release with changelog
   - Uploads binaries to the release
   - Updates Homebrew tap (if configured)

3. **Monitor the workflow**:
   ```bash
   # Check workflow runs
   gh run list --workflow=release.yml
   
   # Watch specific run
   gh run watch <run-id>
   
   # Check release status
   gh release view v0.x.x
   ```

## What NOT to Do

- ❌ NEVER run `make release` manually for production releases
- ❌ NEVER use `gh release create` manually
- ❌ NEVER upload binaries manually to GitHub releases
- ❌ NEVER create releases outside of the GitHub Actions workflow

## Version Numbering

- **Major (v1.0.0)**: Breaking changes, major features
- **Minor (v0.11.0)**: New features, backwards compatible
- **Patch (v0.10.1)**: Bug fixes, documentation updates

## Why GitHub Actions?

- Ensures consistent build environment
- Automatic changelog generation
- Multi-platform builds in parallel
- Signed binaries (if configured)
- Automated Homebrew tap updates
- Reproducible builds
- No manual errors

## Commands for Checking

```bash
# View latest tags
git describe --tags --abbrev=0

# Check commits since last tag
git log --oneline v0.10.0..HEAD

# View GitHub releases
gh release list
```

Remember: The release workflow is defined in `.github/workflows/release.yml` and uses GoReleaser configuration from `.goreleaser.yml`.