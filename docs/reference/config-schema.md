# Config Schema

Complete configuration schema for `.ai-dev-tools.yaml` and `~/.config/ai-dev-tools/config.yaml`.

## Full Schema

```yaml
provider:
  name: anthropic              # Required: openai, anthropic, ollama, custom
  model: claude-3-5-sonnet-20241022  # Required: model identifier
  api_key_env: ANTHROPIC_API_KEY     # Required: env var name for API key
  base_url: ""                        # Optional: for ollama/custom
  timeout: 60                         # Optional: request timeout (seconds)
  max_retries: 3                      # Optional: max retry attempts

tools:
  commit:
    style: conventional           # conventional, gitmoji
  review:
    severity_threshold: warning   # error, warning, info
    exclude_patterns: []          # Glob patterns to exclude from review
    include_patterns: []          # Glob patterns to include (empty = all)
  testgen:
    framework: auto               # gotest, jest, vitest, pytest, auto
    output: file                  # file, stdout
    coverage: false               # Generate coverage-focused tests
  changelog:
    format: keepachangelog        # keepachangelog
    include_prs: false            # Enrich with PR data

github:
  token_env: GITHUB_TOKEN         # Env var for GitHub token
  owner: ""                       # Repo owner (for PR enrichment)
  repo: ""                        # Repo name (for PR enrichment)

output:
  format: text                    # text, json, sarif, github
```

## Field Details

### provider

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Provider: `openai`, `anthropic`, `ollama`, `custom` |
| `model` | string | Yes | - | Model identifier |
| `api_key_env` | string | Yes* | - | Environment variable containing API key |
| `base_url` | string | No | `` | Base URL for ollama/custom providers |
| `timeout` | int | No | `60` | Request timeout in seconds (1-300) |
| `max_retries` | int | No | `3` | Max retry attempts (0-10) |

*Not required for ollama

### tools.commit

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `style` | string | No | `conventional` | Commit style: `conventional`, `gitmoji` |

### tools.review

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `severity_threshold` | string | No | `warning` | Minimum severity: `error`, `warning`, `info` |
| `exclude_patterns` | []string | No | `[]` | Glob patterns to exclude |
| `include_patterns` | []string | No | `[]` | Glob patterns to include |

### tools.testgen

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `framework` | string | No | `auto` | `gotest`, `jest`, `vitest`, `pytest`, `auto` |
| `output` | string | No | `file` | `file`, `stdout` |
| `coverage` | bool | No | `false` | Generate coverage-focused tests |

### tools.changelog

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `format` | string | No | `keepachangelog` | Output format |
| `include_prs` | bool | No | `false` | Enrich with PR data |

### github

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `token_env` | string | No | `GITHUB_TOKEN` | Env var for GitHub token |
| `owner` | string | No | `` | Repository owner |
| `repo` | string | No | `` | Repository name |

### output

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `format` | string | No | `text` | Default output format: `text`, `json`, `sarif`, `github` |

## Example Configurations

### Minimal (Anthropic)
```yaml
provider:
  name: anthropic
  model: claude-3-5-sonnet-20241022
  api_key_env: ANTHROPIC_API_KEY
```

### Full Configuration
```yaml
provider:
  name: anthropic
  model: claude-3-5-sonnet-20241022
  api_key_env: ANTHROPIC_API_KEY
  timeout: 120
  max_retries: 5

tools:
  commit:
    style: conventional
  review:
    severity_threshold: warning
    exclude_patterns:
      - "*.pb.go"
      - "vendor/**"
    include_patterns:
      - "*.go"
  testgen:
    framework: auto
    output: file
    coverage: true
  changelog:
    format: keepachangelog
    include_prs: true

github:
  token_env: GITHUB_TOKEN
  owner: myorg
  repo: myrepo

output:
  format: text
```

### Ollama (Local)
```yaml
provider:
  name: ollama
  model: codellama
  base_url: http://localhost:11434
```

### Custom Provider
```yaml
provider:
  name: custom
  model: my-model
  base_url: https://api.example.com/v1
  api_key_env: CUSTOM_API_KEY
```

## Validation Rules

- `provider.name` must be one of: `openai`, `anthropic`, `ollama`, `custom`
- `provider.timeout` must be 1-300 seconds
- `provider.max_retries` must be 0-10
- `tools.review.severity_threshold` must be one of: `error`, `warning`, `info`
- `tools.testgen.framework` must be one of: `gotest`, `jest`, `vitest`, `pytest`, `auto`
- `tools.testgen.output` must be one of: `file`, `stdout`
- `tools.changelog.format` must be `keepachangelog`
- `output.format` must be one of: `text`, `json`, `sarif`, `github`
- When `tools.changelog.include_prs` is true, `github.owner` and `github.repo` must be set
- When `tools.review` uses `--pr`, `github.owner` and `github.repo` must be set