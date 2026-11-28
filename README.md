# ai-dev-tools

AI-powered developer tools: commit, review, testgen, changelog.

## Features

| Tool | Description |
|------|-------------|
| **commit** | Generate conventional commit messages from staged changes |
| **review** | Multi-pass code review (security, bugs, style, performance) |
| **testgen** | Generate unit tests for Go, TypeScript, Python |
| **changelog** | Keep a Changelog format from git history |

## Install

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
Download from [GitHub Releases](https://github.com/ernedo/ai-dev-tools/releases).

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

This creates a `.ai-dev-tools.yaml` in your repository root. See [Configuration Guide](https://ernedo.github.io/ai-dev-tools/guide/configuration) for all options.

## Providers

| Provider | Models | Setup |
|----------|--------|-------|
| Anthropic | claude-3-5-sonnet, claude-3-opus | API key from [console.anthropic.com](https://console.anthropic.com) |
| OpenAI | gpt-4o, gpt-4o-mini | API key from [platform.openai.com](https://platform.openai.com) |
| Ollama | codellama, llama3, mistral | Run locally: `ollama serve` |
| Custom | Any OpenAI-compatible | Set base_url and api_key_env |

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

## Documentation

- [Getting Started](https://ernedo.github.io/ai-dev-tools/guide/getting-started)
- [Configuration](https://ernedo.github.io/ai-dev-tools/guide/configuration)
- [Providers](https://ernedo.github.io/ai-dev-tools/guide/providers)
- [GitHub Integration](https://ernedo.github.io/ai-dev-tools/guide/github)
- [CLI Reference](https://ernedo.github.io/ai-dev-tools/reference/cli)
- [Config Schema](https://ernedo.github.io/ai-dev-tools/reference/config-schema)
- [Prompt Templates](https://ernedo.github.io/ai-dev-tools/reference/templates)

## Development

```bash
# Build
go build -o ai-dev ./cmd/ai-dev

# Test
go test ./... -short

# Run benchmarks
go test ./pkg/... -bench=. -benchtime=5s -benchmem
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## What I built and what was hard

I built a Go CLI with four commands: commit, review, testgen and changelog. The LLM layer supports Anthropic, OpenAI, Ollama and other OpenAI-compatible endpoints. The review command runs four passes (security, bugs, style, performance) and can post findings to a pull request. It ships as a GitHub Action and through Homebrew, npm, Scoop and a shell installer.

The hard part was reading repository state where libgit2 is unavailable. `pkg/git` uses libgit2 when present and falls back to the git CLI behind one interface, so commands behave the same on a laptop and in CI. Keeping the four review passes returning the same structured output across providers took more work than writing the passes themselves.
