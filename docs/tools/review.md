# Code Reviewer

Multi-pass AI code review with security, bugs, style, and performance analysis.

## Usage

```bash
# Review local changes (main vs HEAD)
ai-dev review --base main --head HEAD

# Review a GitHub PR
ai-dev review --pr 42 --owner your-org --repo your-repo --output github

# Output as SARIF for security tools
ai-dev review --base main --output sarif > results.sarif

# Output as JSON
ai-dev review --base main --output json

# Set minimum severity
ai-dev review --base main --severity error
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--base` | `main` | Base branch for local diff |
| `--head` | `HEAD` | Head branch for local diff |
| `--pr` | - | GitHub PR number |
| `--owner` | - | GitHub repository owner |
| `--repo` | - | GitHub repository name |
| `--output` | `text` | `text`, `json`, `sarif`, `github` |
| `--severity` | `warning` | `error`, `warning`, `info` |

## Review Passes

| Pass | Severity | Focus |
|------|----------|-------|
| Security | error | SQL injection, XSS, secrets, crypto issues |
| Bugs | error | Null derefs, race conditions, logic errors |
| Style | warning | Naming, formatting, best practices |
| Performance | info | N+1 queries, inefficient algorithms, allocations |

## Output Formats

### Text (Default)
```
[ERROR] security: Potential SQL injection in user.go:42
  User input directly concatenated into query

[WARNING] style: Variable 'usr' should be 'user' in handler.go:15
  Use descriptive variable names
```

### JSON
```json
{
  "findings": [
    {
      "file": "user.go",
      "line": 42,
      "severity": "error",
      "category": "security",
      "message": "Potential SQL injection...",
      "suggestion": "Use parameterized queries"
    }
  ]
}
```

### SARIF
Compatible with GitHub Code Scanning, VS Code, and other tools.

### GitHub
Posts review comments directly on the PR.

## Configuration

```yaml
tools:
  review:
    severity_threshold: warning  # error, warning, info
    exclude_patterns:
      - "*.pb.go"
      - "vendor/**"
    include_patterns: []
```

## GitHub Action

```yaml
- uses: ernedo/ai-dev-tools@review
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    severity: warning
```