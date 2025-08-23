# Pull Request Management

Repo-Claude provides centralized pull request management across all repositories in your workspace using the GitHub CLI (`gh`).

## Prerequisites

- **GitHub CLI**: Install `gh` from [cli.github.com](https://cli.github.com)
- **Authentication**: Run `gh auth login` to authenticate with GitHub
- **GitHub Remotes**: Repositories must have GitHub remotes configured

## Commands

### List Pull Requests

View PRs across all repositories in your workspace:

```bash
# List all open PRs
rc pr list

# List all PRs (including closed/merged)
rc pr list --state all

# List your PRs
rc pr list --author @me

# Filter by assignee
rc pr list --assignee username

# Limit results per repo
rc pr list --limit 5
```

Output shows PRs grouped by repository with status indicators:
- 🟢 Open PR
- 📝 Draft PR
- 🔴 Closed PR
- 🟣 Merged PR

### Create Pull Request

Create a new PR in any repository:

```bash
# Interactive mode (prompts for title/body)
rc pr create --repo backend

# With title
rc pr create --repo frontend --title "Add dark mode support"

# With title and body
rc pr create --repo backend --title "Fix auth bug" --body "Fixes #123"

# Create draft PR
rc pr create --repo backend --draft --title "WIP: New feature"

# Specify base branch
rc pr create --repo backend --base develop --title "Feature X"

# Request reviews and assign
rc pr create --repo backend \
  --title "Add payment integration" \
  --reviewers alice,bob \
  --assignees charlie \
  --labels enhancement,backend
```

### Check PR Status

View detailed status of pull requests:

```bash
# Status of all open PRs
rc pr status

# Status of PRs in specific repo
rc pr status --repo backend

# Detailed status of specific PR
rc pr status --repo backend --pr 42
```

Shows:
- PR details (title, author, branch)
- Review status
- CI/CD check results
- Merge conflicts

### Checkout PR Locally

Review or test a PR by checking it out locally:

```bash
# Checkout PR #42 from backend repo
rc pr checkout 42 --repo backend

# Checkout PR #123 from frontend repo
rc pr checkout 123 --repo frontend
```

This fetches the PR branch and switches to it locally.

### Merge Pull Request

Merge a PR when ready:

```bash
# Default merge
rc pr merge 42 --repo backend

# Squash and merge
rc pr merge 42 --repo backend --method squash

# Rebase and merge
rc pr merge 42 --repo backend --method rebase

# Delete branch after merge
rc pr merge 42 --repo backend --delete-branch

# Delete both remote and local branches
rc pr merge 42 --repo backend --delete-branch --delete-local
```

## Workflow Examples

### Review Workflow

1. List open PRs across all repos:
   ```bash
   rc pr list
   ```

2. Check detailed status of a PR:
   ```bash
   rc pr status --repo backend --pr 42
   ```

3. Checkout PR locally for testing:
   ```bash
   rc pr checkout 42 --repo backend
   ```

4. Run tests and review changes:
   ```bash
   rc forall -- git diff main
   ```

5. Merge when approved:
   ```bash
   rc pr merge 42 --repo backend --squash --delete-branch
   ```

### Multi-Repo PR Workflow

When working on features spanning multiple repositories:

1. Create feature branches in affected repos:
   ```bash
   rc forall -- git checkout -b feature/payment-integration
   ```

2. Make changes and commit:
   ```bash
   rc forall -- git add -A
   rc forall -- git commit -m "Add payment integration"
   ```

3. Push branches:
   ```bash
   rc forall -- git push -u origin feature/payment-integration
   ```

4. Create PRs for each repo:
   ```bash
   rc pr create --repo backend --title "Payment API endpoints"
   rc pr create --repo frontend --title "Payment UI components"
   rc pr create --repo shared --title "Payment data models"
   ```

5. Monitor PR status:
   ```bash
   rc pr list --author @me
   rc pr status
   ```

## Integration with Scopes

PR commands work seamlessly with your scope configuration. When you're working in a scope with multiple repositories, you can manage PRs for all of them from one place.

Example with backend scope containing multiple services:

```yaml
scopes:
  - id: backend
    name: Backend Services
    repositories:
      - auth-service
      - order-service
      - payment-service
```

```bash
# List all PRs in backend services
rc pr list

# Create coordinated PRs
rc pr create --repo auth-service --title "Add OAuth support"
rc pr create --repo payment-service --title "Add OAuth integration"
```

## Tips

1. **Use filters effectively**: Combine filters to find specific PRs quickly
   ```bash
   rc pr list --state open --author @me --label bug
   ```

2. **Batch operations**: Use `rc forall` to prepare branches for PRs
   ```bash
   rc forall -- git push -u origin HEAD
   ```

3. **Review dependencies**: Check PR dependencies across repos
   ```bash
   rc pr list --state open | grep -i "depends"
   ```

4. **Clean up branches**: After merging, clean up local branches
   ```bash
   rc forall -- git branch -d feature/old-feature
   ```

## Troubleshooting

### Authentication Issues

If you get authentication errors:
```bash
gh auth status  # Check auth status
gh auth login   # Re-authenticate
```

### Repository Not Found

Ensure the repository name matches exactly:
```bash
rc status  # List all repo names
```

### PR Creation Fails

Common issues:
- No changes between branches
- Base branch doesn't exist
- No push access to repository

### Rate Limiting

GitHub API has rate limits. If you hit them:
- Use `--limit` flag to reduce API calls
- Wait for rate limit reset (usually 1 hour)
- Check limits: `gh api rate_limit`