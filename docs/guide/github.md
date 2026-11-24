# GitHub Integration

GitHub integration enables:
- PR code reviews (posted as review comments)
- Changelog enrichment with PR data
- Automated reviews via GitHub Action

## Setup

### 1. Create a Personal Access Token

1. Go to [GitHub Settings → Developer settings → Personal access tokens](https://github.com/settings/tokens)
2. Generate a token with these scopes:
   - `repo` (full control of private repositories)
   - `read:org` (read org membership)

### 2. Configure Token

Option A: Environment variable
```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

Option B: Config file
```yaml
github:
  token_env: GITHUB_TOKEN
  owner: your-org
  repo: your-repo
```

### 3. Verify

```bash
$ ai-dev doctor
✓ GitHub authentication
```

## PR Code Review

Review a pull request and post comments:

```bash
# Review PR #42
ai-dev review --pr 42 --owner your-org --repo your-repo --output github
```

Output formats:
- `github` - Posts review comments to PR (default for PR review)
- `sarif` - SARIF format for security tools
- `json` - Structured JSON output
- `text` - Human-readable text

## Changelog with PR Data

Enrich changelog with PR titles, authors, and links:

```yaml
tools:
  changelog:
    include_prs: true
```

```bash
ai-dev changelog --since v1.0.0 --output CHANGELOG.md
```

Requires `github.owner` and `github.repo` in config.

## GitHub Action

Add automated code reviews to your workflow:

```yaml
# .github/workflows/ai-review.yml
name: AI Code Review
on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  review:
    uses: ernedo/ai-dev-tools@review
    with:
      github_token: ${{ secrets.GITHUB_TOKEN }}
      provider: anthropic
      model: claude-3-5-sonnet-20241022
      severity: warning
```

The action runs `ai-dev review` on the PR and posts comments.

### Action Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `github_token` | Yes | - | GitHub token |
| `provider` | No | `anthropic` | LLM provider |
| `model` | No | `claude-3-5-sonnet-20241022` | Model name |
| `api_key_env` | No | `ANTHROPIC_API_KEY` | API key env var |
| `base_url` | No | - | Custom provider URL |
| `severity` | No | `warning` | Min severity |
| `pr_number` | No | From context | PR to review |

## Required Permissions

The GitHub token needs:
- `pull_requests: write` - Post review comments
- `contents: read` - Fetch PR diff
- `metadata: read` - Basic repo info