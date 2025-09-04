# Homebrew Tap for Private/Internal Repositories

This guide explains how to set up Homebrew tap for private or internal repositories.

## Methods for Private Tap Access

### 1. SSH-based Tap (Recommended for Internal Use)

For internal repositories, you can use SSH URLs which leverage existing SSH keys:

```bash
# Tap using SSH (requires SSH key authentication)
brew tap yourorg/yourtap git@github.com:yourorg/homebrew-yourtap.git

# Or for internal GitLab/GitHub Enterprise
brew tap yourorg/yourtap git@gitlab.internal.com:yourorg/homebrew-yourtap.git
```

### 2. Personal Access Token (PAT) with HTTPS

For private GitHub repositories, use a PAT embedded in the URL:

```bash
# Create a PAT with 'repo' scope
# Then tap with the token
brew tap yourorg/yourtap https://<TOKEN>@github.com/yourorg/homebrew-yourtap.git

# For better security, use environment variable
export HOMEBREW_GITHUB_API_TOKEN=your_pat_token
brew tap yourorg/yourtap  # Will use the token automatically
```

### 3. GitHub Enterprise / GitLab Self-Hosted

For enterprise installations:

```bash
# Set enterprise URL
export HOMEBREW_GITHUB_API_DOMAIN=github.enterprise.com

# Then tap normally
brew tap yourorg/yourtap
```

## Formula for Private Repository Downloads

When your formula needs to download from a private repository:

### Option 1: SSH URLs in Formula

```ruby
class Muno < Formula
  desc "Your private tool"
  homepage "https://github.internal.com/yourorg/muno"
  
  # Use git clone with SSH for private repos
  url "git@github.internal.com:yourorg/muno.git",
      tag: "v1.0.0",
      revision: "abc123..."
  
  depends_on "go" => :build
  
  def install
    system "go", "build", "-o", bin/"muno", "./cmd/muno"
  end
end
```

### Option 2: Pre-authenticated URLs

```ruby
class Muno < Formula
  desc "Your private tool"
  homepage "https://github.internal.com/yourorg/muno"
  
  # For private GitHub releases, use token in URL
  if ENV["GITHUB_TOKEN"]
    url "https://#{ENV["GITHUB_TOKEN"]}@github.com/yourorg/muno/archive/v1.0.0.tar.gz"
  else
    odie "GITHUB_TOKEN environment variable is required"
  end
  
  sha256 "..."
end
```

### Option 3: Download Strategy for Private Repos

```ruby
# Custom download strategy for private repositories
class GitHubPrivateRepositoryDownloadStrategy < CurlDownloadStrategy
  require "utils/formatter"
  
  def initialize(url, name, version, **meta)
    super
    @headers = []
    @headers << "Authorization: token #{ENV["GITHUB_TOKEN"]}" if ENV["GITHUB_TOKEN"]
  end
  
  def curl_args
    args = super
    @headers.each do |header|
      args << "-H" << header
    end
    args
  end
end

class Muno < Formula
  desc "Your private tool"
  homepage "https://github.internal.com/yourorg/muno"
  
  url "https://api.github.com/repos/yourorg/muno/tarball/v1.0.0",
      using: GitHubPrivateRepositoryDownloadStrategy
  sha256 "..."
  
  def install
    bin.install "muno"
  end
end
```

## Setting Up Private Tap for MUNO

### For Internal/Private MUNO Deployments

1. **Create private tap repository**:
   ```bash
   # Create private homebrew-yourtap repository on your GitHub/GitLab
   ```

2. **Configure authentication**:
   ```bash
   # Add to ~/.zshrc or ~/.bashrc
   export HOMEBREW_GITHUB_API_TOKEN=your_token_here
   
   # Or use SSH config
   Host github.internal.com
     HostName github.internal.com
     User git
     IdentityFile ~/.ssh/id_rsa_internal
   ```

3. **Tap the private repository**:
   ```bash
   # Using SSH (recommended for internal)
   brew tap yourorg/yourtap git@github.internal.com:yourorg/homebrew-yourtap.git
   
   # Using HTTPS with token
   brew tap yourorg/yourtap https://${GITHUB_TOKEN}@github.internal.com/yourorg/homebrew-yourtap.git
   ```

4. **Install from private tap**:
   ```bash
   brew install yourorg/yourtap/muno
   ```

## Security Considerations

1. **Never commit tokens**: Use environment variables or SSH keys
2. **Rotate tokens regularly**: Set expiration dates on PATs
3. **Limit token scope**: Only grant necessary permissions
4. **Use SSH when possible**: More secure for internal repositories
5. **Audit tap access**: Regularly review who has access to tap repositories

## Troubleshooting

### Authentication Issues
```bash
# Test SSH access
ssh -T git@github.internal.com

# Test token access
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Debug brew tap
brew tap --debug yourorg/yourtap
```

### Clear Cached Credentials
```bash
# Remove cached tap
brew untap yourorg/yourtap

# Clear Homebrew cache
rm -rf "$(brew --cache)/yourorg-yourtap"

# Re-tap with fresh credentials
brew tap yourorg/yourtap git@github.internal.com:yourorg/homebrew-yourtap.git
```

## Example: Private MUNO Deployment

For a company deploying MUNO internally:

```bash
# 1. Set up authentication
export HOMEBREW_GITHUB_API_TOKEN=ghp_xxxxxxxxxxxx

# 2. Tap internal repository
brew tap mycompany/tools git@github.mycompany.com:mycompany/homebrew-tools.git

# 3. Install internal MUNO
brew install mycompany/tools/muno

# 4. Verify installation
muno --version
```

This allows teams to distribute MUNO internally while maintaining security and access control.