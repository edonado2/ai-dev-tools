# Security Policy

## Supported Versions

We release security updates for the latest stable version only.

| Version | Supported |
|---------|-----------|
| Latest  | ✓         |
| Older   | ✗         |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please email security@ernedo.dev with:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes

We will:
1. Acknowledge receipt within 48 hours
2. Provide a preliminary assessment within 7 days
3. Work on a fix and coordinate disclosure timeline

## Security Considerations

### API Keys

- Never commit API keys to version control
- Use environment variables: `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, etc.
- Use `ai-dev config init` which stores only the env var name, not the key
- Rotate keys if accidentally exposed

### GitHub Token

- Use a fine-grained PAT with minimal scopes:
  - `repo` (for private repos)
  - `pull_requests: write` (for posting reviews)
- Store in `GITHUB_TOKEN` environment variable
- Never commit tokens

### Local Models (Ollama)

- Ollama runs locally on `localhost:11434`
- No network exposure by default
- Ensure firewall blocks external access to port 11434

### Custom Providers

- Verify the endpoint is trusted
- Use HTTPS in production
- Consider network isolation for sensitive code

### Code Review Output

- Review output may contain code snippets from your codebase
- SARIF/JSON output may be sent to external tools
- GitHub review comments are posted to the PR (visible to collaborators)

### Configuration Files

- `.ai-dev-tools.yaml` may contain:
  - Provider names and model identifiers
  - GitHub owner/repo (not secrets)
  - Pattern exclusions
- Safe to commit (no secrets stored)

## Supply Chain

- Dependencies managed via Go modules
- GoReleaser builds with `CGO_ENABLED=0` for static binaries
- Docker images use `gcr.io/distroless/static-debian12` base
- npm package downloads signed binaries from GitHub Releases

## Verification

### Binary Verification

```bash
# Verify checksums
curl -sL https://github.com/ernedo/ai-dev-tools/releases/download/vX.Y.Z/checksums.txt | sha256sum -c

# Verify GoReleaser signature (if enabled)
cosign verify-blob --certificate=... ai-dev
```

### Dependency Scanning

```bash
# Check for vulnerabilities
govulncheck ./...

# Update dependencies
go get -u ./...
go mod tidy
```

## Disclosure Policy

- We follow coordinated vulnerability disclosure
- Fixes released as patch versions
- Credit given to reporters (unless anonymity requested)
- CVE requested for significant vulnerabilities

## Contact

- Security email: security@ernedo.dev
- PGP key: [Available on request](https://ernedo.dev/pgp)