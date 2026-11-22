# AI-Powered Developer Tool Suite — Design Spec

**Date**: 2026-11-22  
**Status**: Draft for review  
**Project**: `ai-dev-tools`  
**Repo**: `github.com/ernedo/ai-dev-tools` (to be created)

---

## 1. Problem & Goals

### Problem
Developers waste hours on repetitive tasks: writing commit messages, reviewing PRs, generating tests, maintaining changelogs. Existing tools are either:
- **Single-purpose** (one tool per task, fragmented workflow)
- **Closed-source/SaaS** (expensive, privacy concerns, vendor lock-in)
- **Rigid** (hard-coded prompts, no local model support)

### Goals
1. **Unified CLI** — One tool (`ai-dev`) for all AI-assisted dev workflows
2. **Provider-agnostic** — OpenAI, Anthropic, Ollama, custom endpoints via config
3. **Zero-config defaults** — Works in 30 seconds; deeply configurable when needed
4. **Distribution-ready** — Homebrew, npm, GitHub Action, single binary
5. **Portfolio showcase** — Demonstrates Go systems programming, LLM integration, CLI UX, CI/CD

### Non-Goals (v1)
- Web dashboard / GUI
- Team management / billing
- Fine-tuned models
- IDE extensions (VS Code, JetBrains) — future

---

## 2. Architecture

### 2.1 Monorepo Structure

```
ai-dev-tools/
├── cmd/                          # CLI entry points (Cobra)
│   ├── commit/                   # commit-msg generator
│   ├── review/                   # code reviewer
│   ├── testgen/                  # test generator
│   └── changelog/                # PR summarizer + release notes
├── pkg/
│   ├── llm/                      # Provider abstraction
│   ├── git/                      # Git operations
│   ├── github/                   # GitHub API client
│   ├── prompt/                   # Prompt templates (versioned, testable)
│   ├── config/                   # YAML config + env var parsing
│   └── npm/                      # Node.js bindings (NAPI/Go wrapper)
├── .github/
│   ├── workflows/                # CI: test, build, release (GoReleaser)
│   └── action.yml                # GitHub Action for code review
├── docs/                         # Documentation site (VitePress)
├── scripts/                      # Release, install scripts
├── testdata/                     # Golden files for testing
└── go.mod / package.json
```

### 2.2 Core Design Principles

| Principle | Implementation |
|-----------|----------------|
| **Single binary** | Go compiles to static binary; no runtime deps |
| **Interface-first** | All external deps behind interfaces (testable, swappable) |
| **Prompts as code** | Templates versioned, validated, user-overridable |
| **Structured output** | JSON Schema enforcement via LLM tool calling |
| **Graceful degradation** | Works offline with Ollama; no API key required for local |

---

## 3. Component Specifications

### 3.1 LLM Provider Abstraction (`pkg/llm/`)

```go
// Provider interface - all LLMs implement this
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
    Name() string
    MaxTokens() int
    SupportsTools() bool
    SupportsJSONMode() bool
}

// Unified request/response
type CompletionRequest struct {
    Messages       []Message
    Temperature    float32
    MaxTokens      int
    Tools          []Tool
    ToolChoice     string // "auto", "none", "required", or specific tool
    ResponseFormat *ResponseFormat // JSON schema enforcement
    Stop           []string
}

type CompletionResponse struct {
    Content   string
    ToolCalls []ToolCall
    Usage     Usage
    Model     string
}

type Message struct {
    Role       string // "system", "user", "assistant", "tool"
    Content    string
    ToolCalls  []ToolCall
    ToolCallID string
}

// Built-in providers (extensible via registry)
var Providers = map[string]func(Config) (Provider, error){
    "openai":      NewOpenAIProvider,
    "anthropic":   NewAnthropicProvider,
    "ollama":      NewOllamaProvider,
    "custom":      NewCustomProvider, // OpenAI-compatible endpoints
}

// Factory
func New(provider, apiKey, baseURL string, opts ...Option) (Provider, error)
```

**Provider Details**:

| Provider | Models | Auth | Streaming | Tools | JSON Mode |
|----------|--------|------|-----------|-------|-----------|
| OpenAI | GPT-4o, 4o-mini, o1, o3-mini | API Key | Yes | Yes | Yes |
| Anthropic | Claude 3.5 Sonnet, Haiku, Opus | API Key | Yes | Yes | Yes (via tools) |
| Ollama | llama3, codellama, qwen2.5-coder, custom | None (local) | Yes | Limited | Limited |
| Custom | Any OpenAI-compatible | Configurable | Yes | Varies | Varies |

### 3.2 Prompt Template System (`pkg/prompt/`)

```go
// Template with versioned schemas
type Template struct {
    Name         string
    Version      string
    Description  string
    SystemPrompt string
    UserPrompt   string
    InputSchema  *jsonschema.Schema  // Validates input variables
    OutputSchema *jsonschema.Schema  // Enforces structured output
    Tools        []Tool              // For function calling
}

// Template registry
var Templates = map[string]Template{
    "commit@v1":       CommitTemplateV1,
    "review@v1":       ReviewTemplateV1,
    "testgen@v1":      TestGenTemplateV1,
    "changelog@v1":    ChangelogTemplateV1,
}

// Example: commit@v1
var CommitTemplateV1 = Template{
    Name:    "commit",
    Version: "1",
    SystemPrompt: `You are a senior engineer writing conventional commit messages.
Follow the Conventional Commits spec: <type>(<scope>): <subject>
Types: feat, fix, docs, style, refactor, perf, test, chore, revert.
Subject: imperative, lowercase, no period, max 72 chars.
Body: explain what and why, not how.`,
    UserPrompt: `Generate a commit message for this diff:
---
{{.Diff}}
---
Repository context: {{.Context}}
Files changed: {{.Files}}
`,
    InputSchema:  CommitInputSchema,   // validates Diff, Context, Files
    OutputSchema: CommitOutputSchema,  // enforces {message, type, scope, breaking}
}

// User-overridable: place custom templates in .ai-dev-tools/prompts/
// Template resolution: user dir > built-in
```

### 3.3 Git Operations (`pkg/git/`)

```go
type GitRepo interface {
    // Diff operations
    Diff(base, head string, opts DiffOptions) (string, error)
    DiffCached(opts DiffOptions) (string, error)
    DiffStat(base, head string) (DiffStat, error)

    // Log operations
    Log(opts LogOptions) ([]Commit, error)
    LogSince(tag string) ([]Commit, error)

    // Repository state
    Status() (Status, error)
    GetFile(path, ref string) (string, error)
    ListFiles(ref string, pattern string) ([]string, error)
    RootDir() (string, error)

    // Tags & branches
    LatestTag() (string, error)
    CurrentBranch() (string, error)
    IsDirty() (bool, error)
}

type DiffOptions struct {
    IncludePatterns []string
    ExcludePatterns []string
    MaxSize         int64
}

type Commit struct {
    Hash      string
    ShortHash string
    Message   string
    Author    string
    Email     string
    Date      time.Time
    Files     []string
}
```

**Implementation**: `libgit2` via `github.com/libgit2/git2go` for performance; shell-out (`git` CLI) fallback for edge cases.

### 3.4 GitHub Integration (`pkg/github/`)

```go
type GitHubClient interface {
    // PR operations
    GetPR(owner, repo string, number int) (*PR, error)
    ListPRCommits(owner, repo string, number int) ([]Commit, error)
    ListPRFiles(owner, repo string, number int) ([]PRFile, error)

    // Reviews & comments
    CreateReview(owner, repo string, number int, review ReviewRequest) error
    CreateReviewComment(owner, repo string, number int, comment ReviewComment) error
    UpdateReviewComment(owner, repo string, commentID int, body string) error

    // Repository
    GetRepo(owner, repo string) (*Repo, error)
    GetDefaultBranch(owner, repo string) (string, error)

    // Webhooks (for GitHub App future)
    VerifySignature(payload []byte, signature string) bool
}

type ReviewRequest struct {
    Event     string // "COMMENT", "APPROVE", "REQUEST_CHANGES"
    Body      string
    Comments  []ReviewComment
}

type ReviewComment struct {
    Path     string
    Line     int
    Side     string // "LEFT", "RIGHT"
    Body     string
}
```

**Auth**: GitHub App (future) or `GITHUB_TOKEN` (GitHub Actions) or PAT (CLI).

### 3.5 Configuration (`pkg/config/`)

```go
type Config struct {
    Provider ProviderConfig `yaml:"provider" json:"provider"`
    Tools    ToolsConfig    `yaml:"tools" json:"tools"`
    GitHub   GitHubConfig   `yaml:"github" json:"github"`
    Output   OutputConfig   `yaml:"output" json:"output"`
}

type ProviderConfig struct {
    Name       string `yaml:"name" json:"name"`             // "openai", "anthropic", "ollama", "custom"
    Model      string `yaml:"model" json:"model"`           // model identifier
    APIKeyEnv  string `yaml:"api_key_env" json:"api_key_env"` // env var name
    BaseURL    string `yaml:"base_url" json:"base_url"`     // for custom/Ollama
    Timeout    int    `yaml:"timeout" json:"timeout"`       // seconds
    MaxRetries int    `yaml:"max_retries" json:"max_retries"`
}

type ToolsConfig struct {
    Commit   CommitConfig   `yaml:"commit" json:"commit"`
    Review   ReviewConfig   `yaml:"review" json:"review"`
    TestGen  TestGenConfig  `yaml:"testgen" json:"testgen"`
    Changelog ChangelogConfig `yaml:"changelog" json:"changelog"`
}

type CommitConfig struct {
    Style       string `yaml:"style" json:"style"`         // "conventional", "gitmoji"
    MaxLength   int    `yaml:"max_length" json:"max_length"`
    IncludeBody bool   `yaml:"include_body" json:"include_body"`
}

type ReviewConfig struct {
    SeverityThreshold string   `yaml:"severity_threshold" json:"severity_threshold"` // "info", "warning", "error"
    ExcludePatterns   []string `yaml:"exclude_patterns" json:"exclude_patterns"`
    IncludePatterns   []string `yaml:"include_patterns" json:"include_patterns"`
    MaxFilesPerReview int      `yaml:"max_files_per_review" json:"max_files_per_review"`
}

type TestGenConfig struct {
    Framework       string `yaml:"framework" json:"framework"` // "auto", "jest", "vitest", "pytest", "gotest"
    CoverageTarget  int    `yaml:"coverage_target" json:"coverage_target"`
    IncludeMocks    bool   `yaml:"include_mocks" json:"include_mocks"`
    FollowPatterns  bool   `yaml:"follow_patterns" json:"follow_patterns"`
}

type ChangelogConfig struct {
    Format     string `yaml:"format" json:"format"`       // "keepachangelog", "github"
    GroupBy    string `yaml:"group_by" json:"group_by"`   // "type", "scope", "author"
    IncludePRs bool   `yaml:"include_prs" json:"include_prs"`
}
```

**Config resolution** (priority): CLI flags > env vars > `.ai-dev-tools.yaml` (repo root) > `~/.config/ai-dev-tools/config.yaml` > defaults.

### 3.6 CLI Commands (`cmd/*`)

```bash
# Commit message generator
ai-dev commit [flags]
  --dry-run           # Print message, don't commit
  --style string      # conventional, gitmoji
  --sign-off          # Add Signed-off-by trailer

# Code reviewer
ai-dev review [flags]
  --pr int            # PR number (required in CI)
  --base string       # Base branch (default: main)
  --head string       # Head branch (default: current)
  --output string     # "github" (post comments), "stdout", "json", "sarif"
  --severity string   # Minimum severity to report

# Test generator
ai-dev testgen [path...] [flags]
  --framework string  # auto, jest, vitest, pytest, gotest
  --output string     # stdout, file (creates *_test.ext)
  --coverage int      # Target coverage %

# Changelog / PR summarizer
ai-dev changelog [flags]
  --since string      # Tag or date (default: latest tag)
  --until string      # Tag or date (default: HEAD)
  --format string     # keepachangelog, github, json
  --output string     # stdout, CHANGELOG.md

# Config management
ai-dev config init    # Create config file interactively
ai-dev config show    # Print resolved config
ai-dev config path    # Print config file path

# Utility
ai-dev version        # Version + commit hash
ai-dev doctor         # Check config, auth, connectivity
```

---

## 4. Tool Implementation Details

### 4.1 Commit Message Generator (`cmd/commit/`)

| Aspect | Detail |
|--------|--------|
| **Input** | `git diff --cached` (staged changes) |
| **Context** | Repo root, recent commits (last 5), file types |
| **Prompt** | `commit@v1` template with diff + context |
| **Output** | Structured JSON → formatted commit message |
| **UX** | Prints message, prompts "Commit with this message? [Y/n/e]dit" |
| **Git Hook** | `ai-dev commit --hook` → installs `prepare-commit-msg` hook |

### 4.2 Code Reviewer (`cmd/review/`)

| Aspect | Detail |
|--------|--------|
| **Input** | PR diff (via GitHub API or local `git diff base..head`) |
| **Passes** | 1. Security (secrets, SQLi, XSS, path traversal) 2. Bugs (nil pointers, off-by-one, race conditions) 3. Style (naming, complexity, imports) 4. Performance (N+1, allocations, locks) |
| **Output** | SARIF + GitHub review comments |
| **GitHub Action** | Posts as review with line comments; summary in review body |
| **Local** | Outputs to stdout/JSON/SARIF for CI integration |

### 4.3 Test Generator (`cmd/testgen/`)

| Aspect | Detail |
|--------|--------|
| **Input** | Source file(s) + existing test files (for patterns) |
| **Analysis** | Exports, public functions, structs, interfaces, error paths |
| **Framework Detection** | `*_test.go` → gotest; `*.test.ts` + jest.config → jest; etc. |
| **Output** | Table-driven tests, mock generation, edge cases |
| **Incremental** | Only generates for untested functions (diff against existing) |

### 4.4 Changelog / PR Summarizer (`cmd/changelog/`)

| Aspect | Detail |
|--------|--------|
| **Input** | `git log <since>..<until>` + PR metadata (if GitHub) |
| **Grouping** | By conventional commit type: feat, fix, breaking, docs, etc. |
| **PR Enrichment** | Links PR #, author, labels |
| **Output** | Keep a Changelog format (markdown) |

---

## 5. Distribution Strategy

### 5.1 GoReleaser (`.goreleaser.yaml`)

```yaml
builds:
  - id: ai-dev
    main: ./cmd/ai-dev
    binary: ai-dev
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: "{{.ProjectName}}_{{.Version}}_{{.Os}}_{{.Arch}}"

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: ernedo
    name: ai-dev-tools
  draft: false
  prerelease: auto

homebrew:
  tap:
    owner: ernedo
    name: homebrew-tap
  folder: Formula
  homepage: "https://github.com/ernedo/ai-dev-tools"
```

### 5.2 npm Package (`package.json`)

```json
{
  "name": "@ernedo/ai-dev-tools",
  "version": "0.0.0",
  "description": "AI-powered developer tools CLI",
  "bin": { "ai-dev": "./bin/ai-dev" },
  "scripts": {
    "prepare": "node scripts/download-binary.js",
    "test": "node --test test/**/*.test.js"
  },
  "files": ["bin", "lib"],
  "engines": { "node": ">=18" },
  "os": ["darwin", "linux", "win32"],
  "cpu": ["x64", "arm64"]
}
```

**Binary download**: `scripts/download-binary.js` fetches from GitHub Releases at install time.

### 5.3 GitHub Action (`action.yml`)

```yaml
name: "AI Code Review"
description: "Automated code review using LLMs"
author: "ernedo"
inputs:
  github_token:
    required: true
  provider:
    required: false
    default: "anthropic"
  model:
    required: false
    default: "claude-3-5-sonnet-20241022"
  api_key_env:
    required: false
    default: "ANTHROPIC_API_KEY"
  severity:
    required: false
    default: "warning"
runs:
  using: "docker"
  image: "Dockerfile"
```

**Dockerfile**: Distroless + binary, ~15MB.

### 5.4 Install Script

```bash
# https://get.ai-dev.tools
#!/bin/bash
set -euo pipefail
REPO="ernedo/ai-dev-tools"
BINARY="ai-dev"
# Detect OS/arch, fetch from releases, install to /usr/local/bin or ~/.local/bin
```

---

## 6. Testing Strategy

### 6.1 Unit Tests (`*_test.go`)

| Package | Coverage Target | Key Tests |
|---------|----------------|-----------|
| `pkg/llm` | 90% | Provider factory, request/response marshaling, error mapping, streaming |
| `pkg/prompt` | 95% | Template rendering, schema validation, version resolution, user overrides |
| `pkg/git` | 85% | Diff parsing, log filtering, status, libgit2 vs CLI parity |
| `pkg/github` | 80% | PR fetching, comment posting, pagination, rate limit handling |
| `pkg/config` | 95% | Priority resolution, env var expansion, validation, defaults |

### 6.2 Integration Tests (`*_integration_test.go`)

- **Testcontainers**: PostgreSQL (not needed), Redis (not needed), **Ollama container** for local LLM tests
- **GitHub API**: Mock server (`github.com/gnanderson/mockgh`) for review/client tests
- **Git repos**: Temp directories with fixture repos for git operations

### 6.3 Golden File Tests (`testdata/golden/`)

```
testdata/golden/
├── commit/
│   ├── simple_feat.diff.golden
│   └── multi_file_fix.diff.golden
├── review/
│   ├── security_sqli.go.golden
│   └── style_naming.go.golden
├── testgen/
│   ├── go_http_handler.go.golden
│   └── ts_react_component.tsx.golden
└── changelog/
    ├── v1.2.0_log.golden
    └── breaking_change_log.golden
```

**Workflow**: `go test -update` regenerates; CI fails on diff.

### 6.4 E2E Tests (`e2e_test.go`)

```go
func TestCLI_E2E(t *testing.T) {
    tmpDir := t.TempDir()
    // Init git repo, create files, stage changes
    // Run: ai-dev commit --dry-run
    // Assert: output matches expected format
}
```

### 6.5 Benchmarks (`*_bench_test.go`)

- Latency per tool (p50, p95, p99)
- Token usage per tool (input/output)
- Binary size, startup time
- **Tracked in CI** → plotted over time

---

## 7. Documentation (VitePress)

```
docs/
├── index.md                    # Hero, quick start, features
├── guide/
│   ├── getting-started.md      # Install, config, first run (5 min)
│   ├── configuration.md        # Full config reference
│   ├── providers.md            # OpenAI, Anthropic, Ollama, custom
│   ├── custom-prompts.md       # Overriding templates
│   └── github-action.md        # CI integration
├── tools/
│   ├── commit.md               # Commit generator guide
│   ├── review.md               # Code reviewer guide
│   ├── testgen.md              # Test generator guide
│   └── changelog.md            # Changelog guide
├── reference/
│   ├── cli.md                  # Full CLI reference
│   ├── config-schema.md        # JSON schema for config
│   └── templates.md            # Template API
├── contributing.md
└── changelog.md
```

---

## 8. CI/CD Pipeline (`.github/workflows/`)

| Workflow | Triggers | Jobs |
|----------|----------|------|
| `ci.yml` | Push, PR | lint, test (unit+integration), build, benchmark |
| `release.yml` | Tag `v*` | GoReleaser (binaries, Homebrew, npm, Docker), create release |
| `docs.yml` | Push to main (docs/) | Deploy VitePress to GitHub Pages |
| `bench.yml` | Schedule (weekly) | Run benchmarks, commit results to `benchmarks/` branch |

---

## 9. Security Considerations

| Risk | Mitigation |
|------|------------|
| API keys in config | Never logged; read from env vars only; `doctor` masks values |
| Code sent to LLMs | Local-first (Ollama); `.ai-dev-tools-ignore` patterns; user consent |
| GitHub token scope | Minimal: `repo` for PR comments, `read:org` for repo metadata |
| Supply chain | GoReleaser provenance (SLSA), npm provenance, signed commits |
| Prompt injection | Structured output via JSON schema; input validation; template versioning |

---

## 10. Milestones & Timeline

| Milestone | Scope | Target |
|-----------|-------|--------|
| **M1: Core + Commit** | `pkg/llm`, `pkg/git`, `pkg/config`, `cmd/commit`, CLI framework, config, docs skeleton | Week 1-2 |
| **M2: Review + GitHub Action** | `pkg/github`, `cmd/review`, GitHub Action, SARIF output, CI integration | Week 2-3 |
| **M3: TestGen + Changelog** | `cmd/testgen`, `cmd/changelog`, framework detection, golden tests | Week 3-4 |
| **M4: Polish + Release** | npm package, Homebrew tap, install script, docs, benchmarks, v0.1.0 | Week 4-5 |
| **M5: Launch** | Blog post, HN/Reddit, Twitter, GitHub trending push | Week 5 |

---

## 11. Open Questions

1. **Telemetry?** Anonymous usage stats (opt-in) to prioritize features?
2. **Plugin system?** Allow community templates/tools via `~/.ai-dev-tools/plugins/`?
3. **Team features?** Shared prompts, org-level config, review policies?
4. **VS Code extension?** Separate repo or built-in?

---

## 12. Success Metrics (Portfolio)

| Metric | Target (3 months) |
|--------|-------------------|
| GitHub Stars | 500+ |
| npm Downloads/week | 1,000+ |
| Homebrew Installs | 500+ |
| Contributors | 5+ |
| Blog mentions | 3+ |
| Used in production | 10+ repos |

---

*End of Spec*