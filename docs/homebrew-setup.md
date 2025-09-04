# Homebrew Setup for MUNO

This document describes how to set up and maintain the Homebrew tap for MUNO.

## Creating the Tap Repository

1. Create a new GitHub repository named `homebrew-tap`
2. Create the Formula directory structure:
   ```
   homebrew-tap/
   └── Formula/
       └── muno.rb
   ```

3. Copy the formula from `homebrew-formula.rb` to `Formula/muno.rb`

## Updating the Formula

When releasing a new version:

1. Update the formula with the new version URL and SHA256:
   ```ruby
   url "https://github.com/taokim/muno/archive/refs/tags/vX.Y.Z.tar.gz"
   sha256 "actual_sha256_hash_here"
   ```

2. Calculate SHA256:
   ```bash
   curl -L https://github.com/taokim/muno/archive/refs/tags/vX.Y.Z.tar.gz | shasum -a 256
   ```

3. Test the formula locally:
   ```bash
   brew tap --force taokim/tap
   brew install --build-from-source muno
   brew test muno
   ```

4. Commit and push to the tap repository

## User Installation

Users can install MUNO via:
```bash
brew tap taokim/tap
brew install muno
```

Or in a single command:
```bash
brew install taokim/tap/muno
```

## Updating

Users can update MUNO via:
```bash
brew upgrade muno
```

## Uninstalling

```bash
brew uninstall muno
brew untap taokim/tap  # Optional: remove the tap
```

## GitHub Actions for Automated Updates

Consider adding a GitHub Action to automatically update the formula when new releases are published:

```yaml
name: Update Homebrew Formula
on:
  release:
    types: [published]

jobs:
  update-formula:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          repository: taokim/homebrew-tap
          token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
      
      - name: Update Formula
        run: |
          VERSION="${{ github.event.release.tag_name }}"
          URL="https://github.com/taokim/muno/archive/refs/tags/${VERSION}.tar.gz"
          SHA256=$(curl -L "$URL" | shasum -a 256 | cut -d' ' -f1)
          
          sed -i "s|url \".*\"|url \"$URL\"|" Formula/muno.rb
          sed -i "s|sha256 \".*\"|sha256 \"$SHA256\"|" Formula/muno.rb
      
      - name: Commit and Push
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add Formula/muno.rb
          git commit -m "Update muno to ${{ github.event.release.tag_name }}"
          git push
```

## Testing Installation Methods

Test all installation methods work correctly:

1. **Homebrew**: `brew install taokim/tap/muno`
2. **Go Install**: `go install github.com/taokim/muno/cmd/muno@latest`
3. **From Source**: Clone, build, and install manually
4. **Binary Download**: Download from GitHub releases