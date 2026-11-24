# CLI Reference

## Global Flags

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help |
| `--version`, `-v` | Show version |

## Commands

### ai-dev commit

Generate AI-powered commit messages.

```
ai-dev commit [flags]

Flags:
  -d, --dry-run     Print message without committing
  -s, --style       Commit style: conventional, gitmoji (default: conventional)
      --sign-off    Add Signed-off-by trailer
```

### ai-dev review

Multi-pass code review with AI.

```
ai-dev review [flags]

Flags:
      --base        Base branch for local diff (default: main)
      --head        Head branch for local diff (default: HEAD)
      --pr          GitHub PR number to review
      --owner       GitHub repository owner (required with --pr)
      --repo        GitHub repository name (required with --pr)
      --output      Output format: text, json, sarif, github (default: text)
      --severity    Minimum severity: error, warning, info (default: warning)
```

### ai-dev testgen

Generate unit tests for source files.

```
ai-dev testgen [files...] [flags]

Flags:
      --framework   Test framework: gotest, jest, vitest, pytest, auto (default: auto)
      --output      Output mode: file, stdout (default: file)
      --coverage    Generate tests with coverage focus
```

### ai-dev changelog

Generate changelog from git history.

```
ai-dev changelog [flags]

Flags:
      --format      Output format: keepachangelog (default: keepachangelog)
      --since       Start tag/commit (default: latest tag)
      --until       End tag/commit (default: HEAD)
      --output      Output file (default: stdout)
      --include-prs Enrich with GitHub PR data
      --owner       GitHub repository owner (required with --include-prs)
      --repo        GitHub repository name (required with --include-prs)
      --pr          Specific PR number to include
```

### ai-dev config

Manage configuration.

```
ai-dev config [flags]
ai-dev config [command]

Commands:
  init      Initialize configuration file interactively
  show      Show current configuration
  path      Show configuration file path
```

### ai-dev doctor

Run health checks.

```
ai-dev doctor [flags]

Checks:
  - Configuration validity
  - Git repository
  - LLM provider connectivity
  - GitHub authentication
```

### ai-dev version

Show version information.

```
ai-dev version
```

Output:
```
ai-dev-tools v0.1.0
  commit: abc1234
  date:   2024-01-15T10:30:00Z
```

### ai-dev completion

Generate shell completion script.

```
ai-dev completion [bash|zsh|fish|powershell]
```

## Environment Variables

All config values can be overridden via `AI_DEV_*` environment variables:

```bash
AI_DEV_PROVIDER_NAME=anthropic
AI_DEV_PROVIDER_MODEL=claude-3-5-sonnet-20241022
AI_DEV_PROVIDER_API_KEY_ENV=ANTHROPIC_API_KEY
AI_DEV_PROVIDER_BASE_URL=
AI_DEV_PROVIDER_TIMEOUT=60
AI_DEV_PROVIDER_MAX_RETRIES=3
AI_DEV_TOOLS_COMMIT_STYLE=conventional
AI_DEV_TOOLS_REVIEW_SEVERITY_THRESHOLD=warning
AI_DEV_TOOLS_TESTGEN_FRAMEWORK=auto
AI_DEV_TOOLS_TESTGEN_OUTPUT=file
AI_DEV_TOOLS_CHANGELOG_FORMAT=keepachangelog
AI_DEV_GITHUB_TOKEN_ENV=GITHUB_TOKEN
AI_DEV_GITHUB_OWNER=
AI_DEV_GITHUB_REPO=
AI_DEV_OUTPUT_FORMAT=text
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid usage/flags |
| 3 | Configuration error |
| 4 | LLM provider error |
| 5 | Git operation error |
| 6 | GitHub API error |