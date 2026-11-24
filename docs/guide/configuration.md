# Configuration

Configuration priority (highest to lowest):
1. CLI flags
2. Environment variables (`AI_DEV_*`)
3. Repository config (`.ai-dev-tools.yaml`)
4. User config (`~/.config/ai-dev-tools/config.yaml`)
5. Defaults

## Config File Locations

| Location | Use Case |
|----------|----------|
| `.ai-dev-tools.yaml` | Per-repository settings (committed) |
| `~/.config/ai-dev-tools/config.yaml` | User-wide settings |

## Configuration Schema

```yaml
provider:
  name: anthropic           # openai, anthropic, ollama, custom
  model: claude-3-5-sonnet-20241022
  api_key_env: ANTHROPIC_API_KEY
  base_url: ""              # For ollama/custom
  timeout: 60               # seconds
  max_retries: 3

tools:
  commit:
    style: conventional     # conventional, gitmoji
  review:
    severity_threshold: warning  # error, warning, info
    exclude_patterns: []   # glob patterns to exclude
    include_patterns: []   # glob patterns to include
  testgen:
    framework: auto        # gotest, jest, vitest, pytest, auto
    output: file           # file, stdout
    coverage: false
  changelog:
    format: keepachangelog
    include_prs: false

github:
  token_env: GITHUB_TOKEN
  owner: ""                # For PR enrichment
  repo: ""                 # For PR enrichment

output:
  format: text             # text, json, sarif, github
```

## Environment Variables

All config values can be set via environment variables with prefix `AI_DEV_`:

```bash
export AI_DEV_PROVIDER_NAME=anthropic
export AI_DEV_PROVIDER_MODEL=claude-3-5-sonnet-20241022
export AI_DEV_PROVIDER_API_KEY_ENV=ANTHROPIC_API_KEY
export AI_DEV_TOOLS_COMMIT_STYLE=conventional
export AI_DEV_TOOLS_REVIEW_SEVERITY_THRESHOLD=warning
export AI_DEV_GITHUB_TOKEN_ENV=GITHUB_TOKEN
```

## Viewing Configuration

```bash
# Show current config (merged from all sources)
ai-dev config show

# Show config file path
ai-dev config path
```

## Validation

Run `ai-dev doctor` to validate your configuration and check connectivity:

```bash
$ ai-dev doctor
✓ Config valid
✓ Git repository
✓ Anthropic API reachable
✓ GitHub authentication
```