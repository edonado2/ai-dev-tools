# Changelog Generator

Generate Keep a Changelog format changelogs from git history.

## Usage

```bash
# Generate from latest tag to HEAD
ai-dev changelog

# Generate from specific tag
ai-dev changelog --since v1.0.0

# Generate with PR enrichment
ai-dev changelog --since v1.0.0 --include-prs --owner your-org --repo your-repo

# Write to file
ai-dev changelog --output CHANGELOG.md

# Custom end point
ai-dev changelog --since v1.0.0 --until v2.0.0
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--since` | Latest tag | Start tag/commit |
| `--until` | `HEAD` | End tag/commit |
| `--format` | `keepachangelog` | Output format |
| `--output` | `stdout` | Output file |
| `--include-prs` | `false` | Enrich with PR data |
| `--owner` | - | GitHub owner (required with --include-prs) |
| `--repo` | - | GitHub repo (required with --include-prs) |
| `--pr` | - | Specific PR to include |

## Keep a Changelog Format

```markdown
# Changelog

## [Unreleased]

### Added
- feat(cli): add config init command (#42)

### Fixed
- fix(review): handle empty diff gracefully (#45)

## [1.1.0] - 2024-01-15

### Added
- feat(testgen): add vitest support

### Changed
- chore: update dependencies

## [1.0.0] - 2024-01-01
...
```

## Sections

| Section | Conventional Commit Types |
|---------|---------------------------|
| Added | `feat` |
| Changed | `refactor`, `perf` |
| Deprecated | - |
| Removed | `revert` |
| Fixed | `fix` |
| Security | `fix` with security scope |

## PR Enrichment

When `--include-prs` is enabled, each entry includes:
- PR number and link
- PR title
- Author

Requires GitHub config:
```yaml
github:
  token_env: GITHUB_TOKEN
  owner: your-org
  repo: your-repo
```

## Configuration

```yaml
tools:
  changelog:
    format: keepachangelog
    include_prs: false
```

## How It Works

1. Gets commits from git log (since tag or since flag)
2. Optionally fetches PR data from GitHub
3. Sends commits + PR data to LLM with changelog prompt
4. Parses structured JSON response
5. Formats as Keep a Changelog markdown
6. Writes to file or stdout