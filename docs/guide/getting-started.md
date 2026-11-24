# Getting Started

## Installation

### macOS / Linux (Homebrew)
```bash
brew tap ernedo/tap && brew install ai-dev-tools
```

### npm
```bash
npm i -g @ernedo/ai-dev-tools
```

### Windows (Scoop)
```bash
scoop bucket add ernedo https://github.com/ernedo/scoop-bucket
scoop install ai-dev-tools
```

### Install Script (macOS / Linux)
```bash
curl -sfL https://get.ai-dev.tools | sh
```

### Manual Download
Download the latest release from [GitHub Releases](https://github.com/ernedo/ai-dev-tools/releases).

## Quick Start

```bash
# 1. Initialize configuration (choose provider: anthropic, openai, ollama, custom)
ai-dev config init

# 2. Generate a commit message from staged changes
ai-dev commit --dry-run

# 3. Review a pull request
ai-dev review --pr 42 --owner your-org --repo your-repo

# 4. Generate tests for a file
ai-dev testgen ./handler.go --output file

# 5. Generate a changelog
ai-dev changelog --since v1.0.0 --output CHANGELOG.md
```

## Configuration

Run `ai-dev config init` to interactively create a configuration file:

```bash
$ ai-dev config init
? Provider: anthropic
? Model: claude-3-5-sonnet-20241022
? API Key Env Var: ANTHROPIC_API_KEY
? Base URL (optional):
Configuration written to .ai-dev-tools.yaml
```

This creates a `.ai-dev-tools.yaml` in your repository root:

```yaml
provider:
  name: anthropic
  model: claude-3-5-sonnet-20241022
  api_key_env: ANTHROPIC_API_KEY
  timeout: 60
  max_retries: 3
tools:
  commit:
    style: conventional
  review:
    severity_threshold: warning
    exclude_patterns: []
    include_patterns: []
  testgen:
    framework: auto
    output: file
    coverage: false
  changelog:
    format: keepachangelog
    include_prs: false
github:
  token_env: GITHUB_TOKEN
output:
  format: text
```

## Next Steps

- [Configuration](/guide/configuration) - Detailed configuration options
- [Providers](/guide/providers) - Configure different LLM providers
- [GitHub Integration](/guide/github) - Set up GitHub token for PR reviews
- [CLI Reference](/reference/cli) - Complete command reference