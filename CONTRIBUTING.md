# Contributing to ai-dev-tools

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Code of Conduct

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it before contributing.

## How to Contribute

### Reporting Bugs

1. Check if the issue already exists in [Issues](https://github.com/ernedo/ai-dev-tools/issues)
2. Create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment (OS, Go version, ai-dev-tools version)
   - Relevant logs or screenshots

### Suggesting Features

1. Check existing issues and discussions
2. Open a feature request with:
   - Clear use case and motivation
   - Proposed solution (if any)
   - Alternatives considered

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes following the guidelines below
4. Run tests: `go test ./... -short`
5. Commit with conventional commits: `git commit -m "feat(scope): description"`
6. Push and open a PR

## Development Setup

```bash
# Clone and build
git clone https://github.com/ernedo/ai-dev-tools.git
cd ai-dev-tools
go build -o ai-dev ./cmd/ai-dev

# Run tests
go test ./... -short

# Run with race detector
go test ./... -race

# Run linter
golangci-lint run
```

## Coding Standards

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` / `goimports` for formatting
- Run `golangci-lint` before committing
- Write table-driven tests
- Document public APIs with godoc comments

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `revert`

Examples:
- `feat(commit): add gitmoji style support`
- `fix(review): handle empty diff gracefully`
- `docs: update installation instructions`

### Testing

- Unit tests: `*_test.go` alongside code
- Integration tests: `*_integration_test.go` (require API keys, skipped with `-short`)
- Benchmarks: `*_bench_test.go`
- Aim for >80% coverage on new code

### Documentation

- Update relevant docs in `docs/` for user-facing changes
- Update CLI help text for flag changes
- Add examples for new features

## Project Structure

```
├── cmd/                 # CLI commands
│   ├── ai-dev/         # Main entry point
│   ├── commit/         # Commit generator
│   ├── review/         # Code reviewer
│   ├── testgen/        # Test generator
│   ├── changelog/      # Changelog generator
│   ├── config/         # Config management
│   └── doctor/         # Health checks
├── internal/cli/       # CLI wiring
├── pkg/                # Shared packages
│   ├── llm/           # LLM provider abstraction
│   ├── git/           # Git operations
│   ├── github/        # GitHub API
│   ├── config/        # Configuration
│   └── prompt/        # Prompt templates
├── docs/              # VitePress documentation
├── scripts/           # Install scripts
└── .github/           # GitHub workflows
```

## Release Process

Releases are automated via GoReleaser on version tags:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

This triggers:
- Multi-platform binary builds
- GitHub Release with artifacts
- Homebrew formula update
- npm package publish
- Docker image publish

## Getting Help

- [Discussions](https://github.com/ernedo/ai-dev-tools/discussions) for questions
- [Issues](https://github.com/ernedo/ai-dev-tools/issues) for bugs/features
- [Discord](https://discord.gg/ernedo) for real-time chat (if available)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.