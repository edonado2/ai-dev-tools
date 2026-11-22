# AI-Powered Developer Tool Suite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-ready CLI tool suite (`ai-dev`) with 4 AI-powered commands (commit, review, testgen, changelog) using a provider-agnostic LLM abstraction, distributed as Go binary, npm package, and GitHub Action.

**Architecture:** Modular Go monorepo with `cmd/` entry points, `pkg/` shared libraries (llm, git, github, prompt, config), Cobra CLI, GoReleaser for multi-platform binaries, npm wrapper, and Docker-based GitHub Action. Interface-first design with structured JSON output via LLM tool calling.

**Tech Stack:** Go 1.23+, Cobra (CLI), libgit2 (git2go), GoReleaser, VitePress (docs), GitHub Actions CI, Testcontainers (integration tests), jsonschema (prompt validation), Ollama/OpenAI/Anthropic SDKs.

**Spec:** `docs/superpowers/specs/2026-11-22-ai-dev-tools-design.md`

## Global Constraints

- Go version: 1.23+ (use `go:build` tags, no CGO for libgit2 static linking)
- Zero runtime dependencies in released binary (static linking)
- All LLM providers behind `pkg/llm.Provider` interface
- Prompt templates versioned, schema-validated, user-overridable
- Config priority: CLI flags > env vars > `.ai-dev-tools.yaml` (repo) > `~/.config/ai-dev-tools/config.yaml` > defaults
- Structured output enforced via JSON Schema + tool calling
- Tests: unit (table-driven), integration (Testcontainers), golden files, E2E
- Conventional commits for all commits: `feat(scope): ...`, `fix(scope): ...`, etc.

## Review Focus

| Input / Condition | Expected Behavior | Test Location |
|-------------------|-------------------|---------------|
| Invalid/missing API key | Clear error with `doctor` hint; no crash | Task 3 (config), Task 7 (E2E) |
| LLM returns malformed JSON | Retry with stricter schema; fallback to raw parse | Task 2 (llm provider), Task 5 (tools) |
| Huge diff (>100KB) | Truncate with warning; process in chunks | Task 4 (git), Task 5 (review) |
| Rate limited (429) | Exponential backoff; respect `Retry-After` | Task 2 (llm provider) |
| User overrides prompt template | User template loads, built-in ignored | Task 3 (prompt), Task 5 (tools) |

---

### Task 1: Repository Scaffold & Go Module

**Files:**
- Create: `go.mod`, `go.sum`
- Create: `.gitignore`, `.golangci.yml`, `Makefile`
- Create: `cmd/ai-dev/main.go`
- Create: `internal/cli/root.go`, `internal/cli/version.go`
- Test: `internal/cli/root_test.go`

**Interfaces:**
- Consumes: None (foundation)
- Produces: `cmd/ai-dev` entry point, `internal/cli` command registry

- [ ] **Step 1: Write the failing test**

```go
// internal/cli/root_test.go
func TestRootCommand_HasSubcommands(t *testing.T) {
    cmd := NewRootCmd()
    names := []string{}
    for _, c := range cmd.Commands() {
        names = append(names, c.Name())
    }
    assert.Contains(t, names, "commit")
    assert.Contains(t, names, "review")
    assert.Contains(t, names, "testgen")
    assert.Contains(t, names, "changelog")
    assert.Contains(t, names, "config")
    assert.Contains(t, names, "version")
    assert.Contains(t, names, "doctor")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/... -v -run TestRootCommand_HasSubcommands`
Expected: FAIL — `NewRootCmd` undefined

- [ ] **Step 3: Implement scaffold**

```go
// go.mod
module github.com/ernedo/ai-dev-tools

go 1.23

require (
    github.com/spf13/cobra v1.8.1
    github.com/spf13/viper v1.19.0
    github.com/libgit2/git2go/v1.7.1
    github.com/invopop/jsonschema v0.1.0
    go.yaml.in/yaml/v3 v3.0.1
)

// cmd/ai-dev/main.go
package main
import "github.com/ernedo/ai-dev-tools/internal/cli"
func main() { cli.Execute() }

// internal/cli/root.go
package cli
var rootCmd = &cobra.Command{
    Use:   "ai-dev",
    Short: "AI-powered developer tools",
    PersistentPreRunE: initConfig,
}
func Execute() { rootCmd.Execute() }
func NewRootCmd() *cobra.Command { return rootCmd }
func init() {
    rootCmd.AddCommand(
        commit.NewCmd(),
        review.NewCmd(),
        testgen.NewCmd(),
        changelog.NewCmd(),
        config.NewCmd(),
        versionCmd(),
        doctorCmd(),
    )
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cli/... -v -run TestRootCommand_HasSubcommands`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum .gitignore .golangci.yml Makefile cmd/ai-dev/main.go internal/cli/
git commit -m "chore: scaffold repository with Cobra CLI"
```

---

### Task 2: LLM Provider Abstraction (`pkg/llm`)

**Files:**
- Create: `pkg/llm/types.go`, `pkg/llm/interface.go`, `pkg/llm/factory.go`
- Create: `pkg/llm/openai.go`, `pkg/llm/anthropic.go`, `pkg/llm/ollama.go`, `pkg/llm/custom.go`
- Create: `pkg/llm/retry.go` (backoff, circuit breaker)
- Test: `pkg/llm/*_test.go`, `pkg/llm/*_integration_test.go`

**Interfaces:**
- Consumes: `pkg/config.ProviderConfig`
- Produces: `llm.Provider` interface, `llm.New(name, cfg) (Provider, error)`

- [ ] **Step 1: Write the failing tests**

```go
// pkg/llm/factory_test.go
func TestFactory_CreatesProviders(t *testing.T) {
    tests := []struct{name, model string}{
        {"openai", "gpt-4o"},
        {"anthropic", "claude-3-5-sonnet-20241022"},
        {"ollama", "codellama"},
        {"custom", "my-model"},
    }
    for _, tc := range tests {
        p, err := New(tc.name, Config{Model: tc.model})
        assert.NoError(t, err)
        assert.Equal(t, tc.name, p.Name())
    }
}

func TestFactory_UnknownProvider_ReturnsError(t *testing.T) {
    _, err := New("unknown", Config{})
    assert.Error(t, err)
}

// pkg/llm/openai_test.go
func TestOpenAI_Complete_ReturnsStructuredOutput(t *testing.T) {
    if testing.Short() { t.Skip("requires API key") }
    p := NewOpenAI(Config{APIKey: os.Getenv("OPENAI_API_KEY"), Model: "gpt-4o-mini"})
    resp, err := p.Complete(ctx, CompletionRequest{
        Messages: []Message{{Role: "user", Content: "Return JSON: {\"name\":\"test\"}"}},
        ResponseFormat: &ResponseFormat{Type: "json_schema", JSONSchema: map[string]any{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}},
    })
    assert.NoError(t, err)
    var out map[string]string
    json.Unmarshal([]byte(resp.Content), &out)
    assert.Equal(t, "test", out["name"])
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/llm/... -v -run "TestFactory|TestOpenAI_Complete" -short`
Expected: FAIL — types, interface, factory not implemented

- [ ] **Step 3: Implement types & interface**

```go
// pkg/llm/types.go
type Message struct { Role, Content string; ToolCalls []ToolCall; ToolCallID string }
type ToolCall struct { ID, Type string; Function FunctionCall }
type FunctionCall struct { Name, Arguments string }
type CompletionRequest struct {
    Messages       []Message
    Temperature    float32
    MaxTokens      int
    Tools          []Tool
    ToolChoice     string
    ResponseFormat *ResponseFormat
    Stop           []string
}
type ResponseFormat struct { Type string; JSONSchema map[string]any }
type CompletionResponse struct { Content string; ToolCalls []ToolCall; Usage Usage; Model string }
type Usage struct { PromptTokens, CompletionTokens, TotalTokens int }
type StreamChunk struct { Content string; Done bool; Err error }
type Tool struct { Type string; Function ToolFunction }
type ToolFunction struct { Name, Description string; Parameters map[string]any }

// pkg/llm/interface.go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
    Name() string
    MaxTokens() int
    SupportsTools() bool
    SupportsJSONMode() bool
}
```

- [ ] **Step 4: Implement factory & providers**

```go
// pkg/llm/factory.go
var registry = map[string]func(Config) (Provider, error){
    "openai":    NewOpenAI,
    "anthropic": NewAnthropic,
    "ollama":    NewOllama,
    "custom":    NewCustom,
}
func New(name string, cfg Config) (Provider, error) {
    f, ok := registry[name]
    if !ok { return nil, fmt.Errorf("unknown provider: %s", name) }
    return f(cfg)
}

// pkg/llm/openai.go — use github.com/openai/openai-go (official SDK)
func NewOpenAI(cfg Config) (Provider, error) {
    client := openai.NewClient(option.WithAPIKey(cfg.APIKey), option.WithBaseURL(cfg.BaseURL))
    return &openaiProvider{client: client, model: cfg.Model}, nil
}
type openaiProvider struct { client *openai.Client; model string }
func (p *openaiProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
    // Map CompletionRequest → openai.ChatCompletionNewParams
    // Call client.Chat.Completions.New(ctx, params)
    // Map response → CompletionResponse
}
func (p *openaiProvider) Name() string { return "openai" }
func (p *openaiProvider) MaxTokens() int { return 128000 }
func (p *openaiProvider) SupportsTools() bool { return true }
func (p *openaiProvider) SupportsJSONMode() bool { return true }

// pkg/llm/anthropic.go — use github.com/anthropics/anthropic-sdk-go
// pkg/llm/ollama.go — HTTP to localhost:11434/api/chat
// pkg/llm/custom.go — OpenAI-compatible endpoint
```

- [ ] **Step 5: Implement retry/backoff**

```go
// pkg/llm/retry.go
func WithRetry(p Provider, maxRetries int) Provider {
    return &retryProvider{Provider: p, maxRetries: maxRetries}
}
type retryProvider struct { Provider; maxRetries int }
func (p *retryProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
    var lastErr error
    for i := 0; i <= p.maxRetries; i++ {
        resp, err := p.Provider.Complete(ctx, req)
        if err == nil { return resp, nil }
        if !isRetryable(err) { return nil, err }
        lastErr = err
        backoff := time.Duration(1<<i) * time.Second
        select { case <-ctx.Done(): return nil, ctx.Err(); case <-time.After(backoff): }
    }
    return nil, lastErr
}
func isRetryable(err error) bool {
    // Check for 429, 5xx, network errors
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./pkg/llm/... -v -run "TestFactory|TestOpenAI_Complete" -short`
Expected: PASS (integration test skipped without keys; run without `-short` with keys)

- [ ] **Step 7: Commit**

```bash
git add pkg/llm/
git commit -m "feat(llm): provider abstraction with OpenAI, Anthropic, Ollama, custom"
```

---

### Task 3: Prompt Template System (`pkg/prompt`)

**Files:**
- Create: `pkg/prompt/template.go`, `pkg/prompt/registry.go`, `pkg/prompt/loader.go`
- Create: `pkg/prompt/templates/commit_v1.go`, `pkg/prompt/templates/review_v1.go`, `pkg/prompt/templates/testgen_v1.go`, `pkg/prompt/templates/changelog_v1.go`
- Test: `pkg/prompt/*_test.go`

**Interfaces:**
- Consumes: `pkg/llm.CompletionRequest`, `jsonschema.Schema`
- Produces: `prompt.Template` registry, `prompt.Render(name, version, vars) (CompletionRequest, error)`

- [ ] **Step 1: Write the failing tests**

```go
// pkg/prompt/registry_test.go
func TestRegistry_BuiltInTemplates(t *testing.T) {
    reg := NewRegistry()
    names := reg.List()
    assert.Contains(t, names, "commit@v1")
    assert.Contains(t, names, "review@v1")
    assert.Contains(t, names, "testgen@v1")
    assert.Contains(t, names, "changelog@v1")
}

func TestRegistry_UserOverride_LoadsCustom(t *testing.T) {
    tmp := t.TempDir()
    os.WriteFile(filepath.Join(tmp, "commit@v1.yaml"), []byte(`
name: commit
version: "1"
system_prompt: "Custom: {{.Diff}}"
user_prompt: "Custom prompt"
`), 0644)
    reg := NewRegistry(WithUserDir(tmp))
    tpl, err := reg.Get("commit", "1")
    assert.NoError(t, err)
    assert.Contains(t, tpl.SystemPrompt, "Custom")
}

// pkg/prompt/template_test.go
func TestTemplate_Render_ValidatesInput(t *testing.T) {
    tpl := CommitTemplateV1
    req, err := tpl.Render(map[string]any{"Diff": "diff", "Context": "ctx", "Files": []string{"a.go"}})
    assert.NoError(t, err)
    assert.Contains(t, req.Messages[1].Content, "diff")
}

func TestTemplate_Render_MissingRequired_Fails(t *testing.T) {
    tpl := CommitTemplateV1
    _, err := tpl.Render(map[string]any{"Diff": "diff"}) // missing Context, Files
    assert.Error(t, err)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/prompt/... -v -run "TestRegistry|TestTemplate_Render"`
Expected: FAIL

- [ ] **Step 3: Implement template & registry**

```go
// pkg/prompt/template.go
type Template struct {
    Name         string
    Version      string
    Description  string
    SystemPrompt string
    UserPrompt   string
    InputSchema  *jsonschema.Schema
    OutputSchema *jsonschema.Schema
    Tools        []llm.Tool
}
func (t *Template) Render(vars map[string]any) (*llm.CompletionRequest, error) {
    // Validate vars against InputSchema
    // Execute SystemPrompt & UserPrompt as Go templates
    // Return CompletionRequest with ResponseFormat from OutputSchema
}

// pkg/prompt/registry.go
type Registry struct {
    builtins map[string]*Template
    userDir  string
}
func NewRegistry(opts ...Option) *Registry {
    r := &Registry{builtins: builtinTemplates()}
    for _, opt := range opts { opt(r) }
    return r
}
func (r *Registry) Get(name, version string) (*Template, error) {
    key := name + "@v" + version
    // Check userDir first, then builtins
}
func (r *Registry) List() []string { /* return all keys */ }
```

- [ ] **Step 4: Implement built-in templates**

```go
// pkg/prompt/templates/commit_v1.go
var CommitTemplateV1 = &Template{
    Name: "commit", Version: "1",
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
Files changed: {{.Files}}`,
    InputSchema:  mustParseSchema(commitInputSchema),
    OutputSchema: mustParseSchema(commitOutputSchema),
}
const commitInputSchema = `{"type":"object","properties":{"Diff":{"type":"string"},"Context":{"type":"string"},"Files":{"type":"array","items":{"type":"string"}}},"required":["Diff","Context","Files"]}`
const commitOutputSchema = `{"type":"object","properties":{"message":{"type":"string"},"type":{"type":"string"},"scope":{"type":"string"},"breaking":{"type":"boolean"}},"required":["message","type"]}`

// review_v1.go, testgen_v1.go, changelog_v1.go — similar structure
```

- [ ] **Step 5: Implement user template loader (YAML)**

```go
// pkg/prompt/loader.go
func LoadUserTemplates(dir string) (map[string]*Template, error) {
    entries, _ := os.ReadDir(dir)
    templates := make(map[string]*Template)
    for _, e := range entries {
        if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
            data, _ := os.ReadFile(filepath.Join(dir, e.Name()))
            var tpl Template
            yaml.Unmarshal(data, &tpl)
            // Compile schemas
            templates[tpl.Name+"@v"+tpl.Version] = &tpl
        }
    }
    return templates, nil
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./pkg/prompt/... -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add pkg/prompt/
git commit -m "feat(prompt): versioned template system with JSON Schema validation"
```

---

### Task 4: Git Operations (`pkg/git`)

**Files:**
- Create: `pkg/git/interface.go`, `pkg/git/libgit2.go`, `pkg/git/fallback.go`, `pkg/git/diff.go`
- Test: `pkg/git/*_test.go`, `testdata/git/fixtures/`

**Interfaces:**
- Consumes: None (stdlib + libgit2)
- Produces: `git.Repo` interface, `git.Open(path) (Repo, error)`

- [ ] **Step 1: Write the failing tests**

```go
// pkg/git/libgit2_test.go
func TestLibGit2_Diff_ReturnsUnifiedDiff(t *testing.T) {
    repo := setupTestRepo(t) // Creates temp dir with commits
    diff, err := repo.Diff("HEAD~1", "HEAD", DiffOptions{})
    assert.NoError(t, err)
    assert.Contains(t, diff, "@@")
    assert.Contains(t, diff, "func TestSomething")
}

func TestLibGit2_LogSinceTag_ReturnsCommits(t *testing.T) {
    repo := setupTestRepo(t)
    commits, err := repo.LogSince("v1.0.0")
    assert.NoError(t, err)
    assert.Len(t, commits, 3)
    assert.Equal(t, "feat: add feature", commits[0].Message)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/git/... -v -run "TestLibGit2"`
Expected: FAIL

- [ ] **Step 3: Implement interface & libgit2 backend**

```go
// pkg/git/interface.go
type Repo interface {
    Diff(base, head string, opts DiffOptions) (string, error)
    DiffCached(opts DiffOptions) (string, error)
    DiffStat(base, head string) (DiffStat, error)
    Log(opts LogOptions) ([]Commit, error)
    LogSince(tag string) ([]Commit, error)
    Status() (Status, error)
    GetFile(path, ref string) (string, error)
    ListFiles(ref, pattern string) ([]string, error)
    RootDir() (string, error)
    LatestTag() (string, error)
    CurrentBranch() (string, error)
    IsDirty() (bool, error)
    Close() error
}
type DiffOptions struct { IncludePatterns, ExcludePatterns []string; MaxSize int64 }
type DiffStat struct { FilesChanged, Insertions, Deletions int }
type LogOptions struct { Since, Until string; MaxCount int }
type Commit struct { Hash, ShortHash, Message, Author, Email string; Date time.Time; Files []string }

// pkg/git/libgit2.go
func Open(path string) (Repo, error) {
    r, err := git.OpenRepository(path)
    if err != nil { return nil, err }
    return &libgit2Repo{repo: r}, nil
}
type libgit2Repo struct { repo *git.Repository }
func (r *libgit2Repo) Diff(base, head string, opts DiffOptions) (string, error) {
    // Use git.DiffOptions with callbacks to build unified diff string
    // Respect IncludePatterns/ExcludePatterns via pathspec
}
func (r *libgit2Repo) LogSince(tag string) ([]Commit, error) {
    // Revwalk from tag to HEAD, collect commits
}
// ... implement all methods
```

- [ ] **Step 4: Implement fallback (shell-out to `git` CLI)**

```go
// pkg/git/fallback.go
func OpenFallback(path string) (Repo, error) {
    // Verify git exists, return &fallbackRepo{dir: path}
}
type fallbackRepo struct { dir string }
func (r *fallbackRepo) Diff(base, head string, opts DiffOptions) (string, error) {
    args := []string{"diff", base + ".." + head}
    // Add pathspecs for patterns
    out, _ := exec.Command("git", args...).CombinedOutput()
    return string(out), nil
}
// ... other methods via git CLI
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./pkg/git/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/git/ testdata/git/
git commit -m "feat(git): libgit2-backed repo operations with CLI fallback"
```

---

### Task 5: GitHub Integration (`pkg/github`)

**Files:**
- Create: `pkg/github/client.go`, `pkg/github/pr.go`, `pkg/github/review.go`
- Test: `pkg/github/*_test.go`, `pkg/github/*_integration_test.go` (with mockgh)

**Interfaces:**
- Consumes: `github.com/google/go-github/v60/github`
- Produces: `github.Client` interface, `github.New(token) Client`

- [ ] **Step 1: Write the failing tests**

```go
// pkg/github/client_test.go
func TestClient_GetPR_ReturnsPR(t *testing.T) {
    if testing.Short() { t.Skip("requires GitHub token") }
    c := New(os.Getenv("GITHUB_TOKEN"))
    pr, err := c.GetPR("ernedo", "test-repo", 1)
    assert.NoError(t, err)
    assert.Equal(t, 1, pr.Number)
}

func TestClient_CreateReviewComment_PostsComment(t *testing.T) {
    if testing.Short() { t.Skip("requires GitHub token") }
    c := New(os.Getenv("GITHUB_TOKEN"))
    err := c.CreateReviewComment("ernedo", "test-repo", 1, ReviewComment{
        Path: "file.go", Line: 10, Side: "RIGHT", Body: "Test comment",
    })
    assert.NoError(t, err)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/github/... -v -short`
Expected: FAIL

- [ ] **Step 3: Implement client wrapper**

```go
// pkg/github/client.go
type Client interface {
    GetPR(owner, repo string, number int) (*PR, error)
    ListPRCommits(owner, repo string, number int) ([]Commit, error)
    ListPRFiles(owner, repo string, number int) ([]PRFile, error)
    CreateReview(owner, repo string, number int, review ReviewRequest) error
    CreateReviewComment(owner, repo string, number int, comment ReviewComment) error
    GetRepo(owner, repo string) (*Repo, error)
    GetDefaultBranch(owner, repo string) (string, error)
}
func New(token string) Client {
    return &ghClient{client: github.NewClient(nil).WithAuthToken(token)}
}
type ghClient struct { client *github.Client }
func (c *ghClient) GetPR(owner, repo string, number int) (*PR, error) {
    pr, _, err := c.client.PullRequests.Get(context.Background(), owner, repo, number)
    return toPR(pr), err
}
func (c *ghClient) CreateReviewComment(owner, repo string, number int, comment ReviewComment) error {
    _, _, err := c.client.PullRequests.CreateComment(context.Background(), owner, repo, number, &github.PullRequestComment{
        Path: github.String(comment.Path),
        Line: github.Int(comment.Line),
        Side: github.String(comment.Side),
        Body: github.String(comment.Body),
    })
    return err
}
// ... other methods
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/github/... -v -short` (unit) and without `-short` (integration)
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/github/
git commit -m "feat(github): PR fetching and review comment posting"
```

---

### Task 6: Configuration (`pkg/config`)

**Files:**
- Create: `pkg/config/config.go`, `pkg/config/loader.go`, `pkg/config/validation.go`
- Test: `pkg/config/*_test.go`

**Interfaces:**
- Consumes: Viper, filesystem
- Produces: `config.Load() (Config, error)`, `config.Resolve() Config`

- [ ] **Step 1: Write the failing tests**

```go
// pkg/config/loader_test.go
func TestLoader_PriorityOrder(t *testing.T) {
    tmp := t.TempDir()
    // 1. Default
    cfg := Load()
    assert.Equal(t, "anthropic", cfg.Provider.Name)
    // 2. Config file
    os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(`provider:
  name: openai
`), 0644)
    cfg = Load(WithConfigDir(tmp))
    assert.Equal(t, "openai", cfg.Provider.Name)
    // 3. Env var overrides file
    t.Setenv("AI_DEV_PROVIDER_NAME", "ollama")
    cfg = Load(WithConfigDir(tmp))
    assert.Equal(t, "ollama", cfg.Provider.Name)
    // 4. CLI flag would override env (tested in CLI tasks)
}

func TestLoader_Validation_RejectsInvalid(t *testing.T) {
    _, err := Load(WithConfigYAML(`provider:
  name: invalid
`))
    assert.Error(t, err)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/config/... -v -run TestLoader_PriorityOrder`
Expected: FAIL

- [ ] **Step 3: Implement config structs & loader**

```go
// pkg/config/config.go
type Config struct {
    Provider ProviderConfig `mapstructure:"provider"`
    Tools    ToolsConfig    `mapstructure:"tools"`
    GitHub   GitHubConfig   `mapstructure:"github"`
    Output   OutputConfig   `mapstructure:"output"`
}
type ProviderConfig struct {
    Name       string `mapstructure:"name" validate:"oneof=openai anthropic ollama custom"`
    Model      string `mapstructure:"model"`
    APIKeyEnv  string `mapstructure:"api_key_env"`
    BaseURL    string `mapstructure:"base_url"`
    Timeout    int    `mapstructure:"timeout" validate:"min=1,max=300"`
    MaxRetries int    `mapstructure:"max_retries" validate:"min=0,max=10"`
}
// ... ToolsConfig, GitHubConfig, OutputConfig per spec

// pkg/config/loader.go
func Load(opts ...Option) (Config, error) {
    v := viper.New()
    v.SetConfigName("ai-dev-tools")
    v.SetConfigType("yaml")
    v.AddConfigPath(".")
    v.AddConfigPath("$HOME/.config/ai-dev-tools")
    // Bind env vars with prefix AI_DEV_
    v.SetEnvPrefix("AI_DEV")
    v.AutomaticEnv()
    // Apply defaults
    setDefaults(v)
    // Merge config file
    if err := v.ReadInConfig(); err != nil { /* ignore if not found */ }
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil { return Config{}, err }
    return cfg, validate(cfg)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/config/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/config/
git commit -m "feat(config): layered config with CLI > env > file > defaults priority"
```

---

### Task 7: CLI Commands — Commit Generator (`cmd/commit`)

**Files:**
- Create: `cmd/commit/command.go`
- Test: `cmd/commit/*_test.go`, `e2e/commit_test.go`

**Interfaces:**
- Consumes: `pkg/git.Repo`, `pkg/prompt.Registry`, `pkg/llm.Provider`, `pkg/config.Config`
- Produces: `commit.Command` (Cobra), generates commit message

- [ ] **Step 1: Write the failing tests**

```go
// cmd/commit/command_test.go
func TestCommitCommand_DryRun_PrintsMessage(t *testing.T) {
    repo := setupGitRepo(t) // with staged changes
    llm := &mockProvider{response: `{"message":"feat(cli): add commit cmd","type":"feat","scope":"cli"}`}
    cmd := NewCommand(repo, llm, &Config{Tools: ToolsConfig{Commit: CommitConfig{Style: "conventional"}}})
    out := captureOutput(func() { cmd.RunE(cmd, []string{"--dry-run"}) })
    assert.Contains(t, out, "feat(cli): add commit cmd")
}

func TestCommitCommand_Interactive_CommitsOnConfirm(t *testing.T) {
    repo := setupGitRepo(t)
    llm := &mockProvider{response: `{"message":"fix: bug","type":"fix"}`}
    cmd := NewCommand(repo, llm, defaultConfig())
    // Simulate "y" input
    cmd.SetIn(strings.NewReader("y\n"))
    err := cmd.RunE(cmd, []string{})
    assert.NoError(t, err)
    // Verify commit was created
    commits, _ := repo.Log(LogOptions{MaxCount: 1})
    assert.Contains(t, commits[0].Message, "fix: bug")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/commit/... -v`
Expected: FAIL

- [ ] **Step 3: Implement command**

```go
// cmd/commit/command.go
func NewCommand(repo git.Repo, llm llm.Provider, cfg *Config) *cobra.Command {
    var dryRun, signOff bool
    var style string
    cmd := &cobra.Command{
        Use: "commit",
        Short: "Generate AI-powered commit message",
        RunE: func(cmd *cobra.Command, args []string) error {
            diff, err := repo.DiffCached(DiffOptions{})
            if err != nil { return err }
            if strings.TrimSpace(diff) == "" {
                return errors.New("no staged changes")
            }
            ctx := cmd.Context()
            tpl, _ := prompt.Registry.Get("commit", "1")
            req, _ := tpl.Render(map[string]any{
                "Diff": diff, "Context": getContext(repo), "Files": getStagedFiles(repo),
            })
            resp, err := llm.Complete(ctx, req)
            if err != nil { return err }
            var out CommitOutput
            json.Unmarshal([]byte(resp.Content), &out)
            msg := formatMessage(out, style, signOff)
            if dryRun {
                fmt.Println(msg)
                return nil
            }
            return commitWithMessage(repo, msg)
        },
    }
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print message without committing")
    cmd.Flags().StringVar(&style, "style", "conventional", "conventional|gitmoji")
    cmd.Flags().BoolVar(&signOff, "sign-off", false, "Add Signed-off-by")
    return cmd
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/commit/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/commit/
git commit -m "feat(commit): AI commit message generator with conventional format"
```

---

### Task 8: CLI Commands — Code Reviewer (`cmd/review`)

**Files:**
- Create: `cmd/review/command.go`, `cmd/review/passes.go` (security, bugs, style, perf)
- Test: `cmd/review/*_test.go`, `e2e/review_test.go`

**Interfaces:**
- Consumes: `pkg/git.Repo`, `pkg/github.Client`, `pkg/prompt.Registry`, `pkg/llm.Provider`
- Produces: `review.Command`, posts GitHub review comments or outputs SARIF/JSON

- [ ] **Step 1: Write the failing tests**

```go
// cmd/review/command_test.go
func TestReviewCommand_LocalOutputsSARIF(t *testing.T) {
    repo := setupGitRepo(t) // with changes
    llm := &mockProvider{response: sarifResponse()}
    gh := &mockGitHub{}
    cmd := NewCommand(repo, gh, llm, defaultConfig())
    cmd.SetArgs([]string{"--base", "main", "--head", "feature", "--output", "sarif"})
    out := captureOutput(func() { cmd.Execute() })
    var sarif map[string]any
    json.Unmarshal([]byte(out), &sarif)
    assert.Equal(t, "2.1.0", sarif["version"])
    assert.Len(t, sarif["runs"].([]any), 1)
}

func TestReviewCommand_GitHubAction_PostsComments(t *testing.T) {
    repo := setupGitRepo(t)
    llm := &mockProvider{response: reviewResponse()}
    gh := &mockGitHub{comments: []ReviewComment{}}
    cmd := NewCommand(repo, gh, llm, defaultConfig())
    cmd.SetArgs([]string{"--pr", "42", "--output", "github"})
    err := cmd.Execute()
    assert.NoError(t, err)
    assert.Len(t, gh.(*mockGitHub).comments, 3)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/review/... -v`
Expected: FAIL

- [ ] **Step 3: Implement multi-pass reviewer**

```go
// cmd/review/passes.go
type ReviewPass struct {
    Name        string
    SystemPrompt string
    Severity    string // "error", "warning", "info"
}
var Passes = []ReviewPass{
    {Name: "security", SystemPrompt: securityPrompt, Severity: "error"},
    {Name: "bugs", SystemPrompt: bugsPrompt, Severity: "error"},
    {Name: "style", SystemPrompt: stylePrompt, Severity: "warning"},
    {Name: "performance", SystemPrompt: perfPrompt, Severity: "info"},
}

// cmd/review/command.go
func NewCommand(repo git.Repo, gh github.Client, llm llm.Provider, cfg *Config) *cobra.Command {
    var prNum int; var base, head, output, severity string
    cmd := &cobra.Command{
        Use: "review",
        RunE: func(cmd *cobra.Command, args []string) error {
            var diff string
            if prNum > 0 {
                // Fetch PR diff via GitHub API
                files, _ := gh.ListPRFiles(owner, repo, prNum)
                diff = buildDiffFromFiles(files)
            } else {
                diff, _ = repo.Diff(base, head, DiffOptions{})
            }
            findings := runPasses(cmd.Context(), llm, diff, cfg.Tools.Review)
            switch output {
            case "github":
                return postGitHubReview(gh, owner, repo, prNum, findings)
            case "sarif":
                return writeSARIF(findings)
            case "json":
                return json.NewEncoder(os.Stdout).Encode(findings)
            default:
                return printFindings(findings)
            }
        },
    }
    // flags...
    return cmd
}
func runPasses(ctx context.Context, llm llm.Provider, diff string, cfg ReviewConfig) []Finding {
    var all []Finding
    for _, pass := range Passes {
        if severityBelow(cfg.SeverityThreshold, pass.Severity) { continue }
        req := buildReviewRequest(pass, diff)
        resp, _ := llm.Complete(ctx, req)
        all = append(all, parseFindings(resp)...)
    }
    return filterByPatterns(all, cfg.ExcludePatterns, cfg.IncludePatterns)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/review/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/review/
git commit -m "feat(review): multi-pass code reviewer with GitHub integration"
```

---

### Task 9: CLI Commands — Test Generator (`cmd/testgen`)

**Files:**
- Create: `cmd/testgen/command.go`, `cmd/testgen/generator.go` (per framework)
- Test: `cmd/testgen/*_test.go`, `testdata/testgen/golden/`

**Interfaces:**
- Consumes: `pkg/git.Repo`, `pkg/prompt.Registry`, `pkg/llm.Provider`
- Produces: `testgen.Command`, writes test files

- [ ] **Step 1: Write the failing tests**

```go
// cmd/testgen/command_test.go
func TestTestGenCommand_GoGeneratesTableTests(t *testing.T) {
    repo := setupGoRepo(t) // with handler.go
    llm := &mockProvider{response: goTestResponse()}
    cmd := NewCommand(repo, llm, defaultConfig())
    cmd.SetArgs([]string{"./handler.go", "--output", "file"})
    err := cmd.Execute()
    assert.NoError(t, err)
    // Verify _test.go created with table-driven tests
    content, _ := os.ReadFile("handler_test.go")
    assert.Contains(t, string(content), "func TestHandle")
    assert.Contains(t, string(content), "testCases := []struct")
}

func TestTestGenCommand_FrameworkAutoDetect(t *testing.T) {
    repo := setupTSRepo(t) // with component.tsx + jest.config.js
    llm := &mockProvider{response: tsTestResponse()}
    cmd := NewCommand(repo, llm, defaultConfig())
    cmd.SetArgs([]string{"./component.tsx"})
    err := cmd.Execute()
    assert.NoError(t, err)
    content, _ := os.ReadFile("component.test.tsx")
    assert.Contains(t, string(content), "describe(")
    assert.Contains(t, string(content), "jest.fn()")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/testgen/... -v`
Expected: FAIL

- [ ] **Step 3: Implement generator with framework detection**

```go
// cmd/testgen/generator.go
func DetectFramework(files []string) string {
    for _, f := range files {
        switch {
        case strings.HasSuffix(f, "_test.go"): return "gotest"
        case strings.HasSuffix(f, ".test.ts"), strings.HasSuffix(f, ".test.tsx"):
            if fileExists("jest.config.js") { return "jest" }
            if fileExists("vitest.config.ts") { return "vitest" }
        case strings.HasSuffix(f, "_test.py"): return "pytest"
        }
    }
    return "auto"
}

// cmd/testgen/command.go
func NewCommand(repo git.Repo, llm llm.Provider, cfg *Config) *cobra.Command {
    var framework, output string; var coverage int
    cmd := &cobra.Command{
        Use: "testgen [path...]",
        RunE: func(cmd *cobra.Command, args []string) error {
            files := expandPaths(args)
            fw := framework
            if fw == "auto" { fw = DetectFramework(files) }
            for _, f := range files {
                src, _ := os.ReadFile(f)
                existingTests := findExistingTests(f)
                req := buildTestGenRequest(fw, string(src), existingTests)
                resp, _ := llm.Complete(cmd.Context(), req)
                testCode := extractTestCode(resp.Content)
                if output == "file" {
                    testPath := testFilePath(f, fw)
                    os.WriteFile(testPath, []byte(testCode), 0644)
                } else {
                    fmt.Println(testCode)
                }
            }
            return nil
        },
    }
    // flags...
    return cmd
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/testgen/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/testgen/
git commit -m "feat(testgen): multi-framework test generator with auto-detection"
```

---

### Task 10: CLI Commands — Changelog / PR Summarizer (`cmd/changelog`)

**Files:**
- Create: `cmd/changelog/command.go`
- Test: `cmd/changelog/*_test.go`

**Interfaces:**
- Consumes: `pkg/git.Repo`, `pkg/github.Client`, `pkg/prompt.Registry`, `pkg/llm.Provider`
- Produces: `changelog.Command`, updates CHANGELOG.md

- [ ] **Step 1: Write the failing tests**

```go
// cmd/changelog/command_test.go
func TestChangelogCommand_KeepAChangelogFormat(t *testing.T) {
    repo := setupGitRepo(t) // with conventional commits since v1.0.0
    llm := &mockProvider{response: changelogResponse()}
    cmd := NewCommand(repo, llm, defaultConfig())
    cmd.SetArgs([]string{"--since", "v1.0.0", "--format", "keepachangelog"})
    out := captureOutput(func() { cmd.Execute() })
    assert.Contains(t, out, "## [1.1.0]")
    assert.Contains(t, out, "### Added")
    assert.Contains(t, out, "### Fixed")
    assert.Contains(t, out, "feat: new feature")
}

func TestChangelogCommand_UpdatesFile(t *testing.T) {
    repo := setupGitRepo(t)
    llm := &mockProvider{response: changelogResponse()}
    cmd := NewCommand(repo, llm, defaultConfig())
    cmd.SetArgs([]string{"--output", "CHANGELOG.md"})
    err := cmd.Execute()
    assert.NoError(t, err)
    content, _ := os.ReadFile("CHANGELOG.md")
    assert.Contains(t, string(content), "## [Unreleased]")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/changelog/... -v`
Expected: FAIL

- [ ] **Step 3: Implement command**

```go
// cmd/changelog/command.go
func NewCommand(repo git.Repo, gh github.Client, llm llm.Provider, cfg *Config) *cobra.Command {
    var since, until, format, output string; var includePRs bool
    cmd := &cobra.Command{
        Use: "changelog",
        RunE: func(cmd *cobra.Command, args []string) error {
            if since == "" { since, _ = repo.LatestTag() }
            commits, _ := repo.Log(LogOptions{Since: since, Until: until})
            prData := map[int]PR{}
            if includePRs {
                // Enrich with PR data from GitHub
            }
            req := buildChangelogRequest(format, commits, prData)
            resp, _ := llm.Complete(cmd.Context(), req)
            changelog := resp.Content
            if output != "" {
                return writeChangelog(output, changelog, format)
            }
            fmt.Print(changelog)
            return nil
        },
    }
    // flags...
    return cmd
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/changelog/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/changelog/
git commit -m "feat(changelog): AI-generated changelogs from git history"
```

---

### Task 11: Config CLI & Doctor (`cmd/config`, `cmd/doctor`)

**Files:**
- Create: `cmd/config/command.go`, `cmd/doctor/command.go`
- Test: `cmd/config/*_test.go`, `cmd/doctor/*_test.go`

**Interfaces:**
- Consumes: `pkg/config.Config`, `pkg/llm.Provider`
- Produces: Config management commands, health check

- [ ] **Step 1: Write the failing tests**

```go
// cmd/config/command_test.go
func TestConfigInit_CreatesFile(t *testing.T) {
    tmp := t.TempDir()
    cmd := NewConfigCommand(tmp)
    cmd.SetArgs([]string{"init"})
    cmd.SetIn(strings.NewReader("anthropic\nclaude-3-5-sonnet\n\n"))
    err := cmd.Execute()
    assert.NoError(t, err)
    content, _ := os.ReadFile(filepath.Join(tmp, ".ai-dev-tools.yaml"))
    assert.Contains(t, string(content), "anthropic")
}

// cmd/doctor/command_test.go
func TestDoctor_ChecksConfigAndConnectivity(t *testing.T) {
    cmd := NewDoctorCommand(validConfig())
    out := captureOutput(func() { cmd.Execute() })
    assert.Contains(t, out, "✓ Config valid")
    assert.Contains(t, out, "✓ Anthropic API reachable")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/config/... ./cmd/doctor/... -v`
Expected: FAIL

- [ ] **Step 3: Implement commands**

```go
// cmd/config/command.go
func NewConfigCommand(configDir string) *cobra.Command {
    var show, path bool
    cmd := &cobra.Command{
        Use: "config",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, _ := config.Load(config.WithConfigDir(configDir))
            if show { printYAML(cfg) }
            if path { fmt.Println(configPath(configDir)) }
            return nil
        },
    }
    cmd.AddCommand(initCmd(configDir))
    return cmd
}
func initCmd(dir string) *cobra.Command {
    return &cobra.Command{
        Use: "init",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Interactive prompt for provider, model, api key env
            // Write to .ai-dev-tools.yaml
        },
    }
}

// cmd/doctor/command.go
func NewDoctorCommand(cfg *Config) *cobra.Command {
    return &cobra.Command{
        Use: "doctor",
        RunE: func(cmd *cobra.Command, args []string) error {
            checks := []Check{
                {"Config valid", validateConfig(cfg)},
                {"Git repo", isGitRepo()},
                {"Provider reachable", testProvider(cfg.Provider)},
                {"GitHub auth", testGitHub(cfg.GitHub)},
            }
            for _, c := range checks {
                status := "✓"
                if c.Err != nil { status = "✗" }
                fmt.Printf("%s %s\n", status, c.Name)
                if c.Err != nil { fmt.Printf("  %v\n", c.Err) }
            }
            return nil
        },
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/config/... ./cmd/doctor/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/config/ cmd/doctor/
git commit -m "feat(config,doctor): config management and health check commands"
```

---

### Task 12: GoReleaser, Homebrew Tap, npm Package

**Files:**
- Create: `.goreleaser.yaml`, `.github/workflows/release.yml`
- Create: `homebrew-tap/Formula/ai-dev-tools.rb` (in separate repo)
- Create: `package.json`, `scripts/download-binary.js`, `bin/ai-dev` (wrapper)
- Create: `Dockerfile` (for GitHub Action)
- Create: `.github/action.yml`
- Test: `scripts/download-binary.test.js`

**Interfaces:**
- Consumes: GoReleaser, GitHub Actions, npm
- Produces: Multi-platform binaries, Homebrew formula, npm package, Docker image

- [ ] **Step 1: Write the failing test (download script)**

```javascript
// scripts/download-binary.test.js
const { downloadBinary } = require('./download-binary');
test('downloads correct binary for platform', async () => {
    const path = await downloadBinary('v0.1.0', 'darwin', 'arm64');
    assert.ok(fs.existsSync(path));
    assert.ok(fs.statSync(path).mode & 0o111); // executable
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test scripts/download-binary.test.js`
Expected: FAIL

- [ ] **Step 3: Implement GoReleaser config**

```yaml
# .goreleaser.yaml
version: 2
before:
  hooks:
    - go mod tidy
    - go generate ./...
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
    env:
      - CGO_ENABLED=0
archives:
  - format: tar.gz
    name_template: "{{.ProjectName}}_{{.Version}}_{{.Os}}_{{.Arch}}"
    files:
      - README.md
      - LICENSE
checksum:
  name_template: "checksums.txt"
release:
  github:
    owner: ernedo
    name: ai-dev-tools
  draft: false
homebrew:
  tap:
    owner: ernedo
    name: homebrew-tap
  folder: Formula
  homepage: "https://github.com/ernedo/ai-dev-tools"
  description: "AI-powered developer tools CLI"
  license: "MIT"
```

- [ ] **Step 4: Implement release workflow**

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']
permissions:
  contents: write
  packages: write
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 5: Implement npm package**

```json
// package.json
{
  "name": "@ernedo/ai-dev-tools",
  "version": "0.0.0",
  "description": "AI-powered developer tools CLI",
  "bin": { "ai-dev": "./bin/ai-dev" },
  "scripts": {
    "prepare": "node scripts/download-binary.js",
    "test": "node --test scripts/download-binary.test.js"
  },
  "files": ["bin", "lib"],
  "engines": { "node": ">=18" },
  "os": ["darwin", "linux", "win32"],
  "cpu": ["x64", "arm64"],
  "repository": "github:ernedo/ai-dev-tools"
}
```

```javascript
// scripts/download-binary.js
const { platform, arch } = process;
const version = require('../package.json').version;
const base = `https://github.com/ernedo/ai-dev-tools/releases/download/v${version}`;
const map = { darwin: 'Darwin', linux: 'Linux', win32: 'Windows' };
const archMap = { x64: 'x86_64', arm64: 'arm64' };
const url = `${base}/ai-dev-tools_${version}_${map[platform]}_${archMap[arch]}.tar.gz`;
// Download, extract, chmod +x bin/ai-dev
```

```javascript
// bin/ai-dev (wrapper)
#!/usr/bin/env node
const { execFileSync } = require('child_process');
const path = require('path');
const binary = path.join(__dirname, '../lib/ai-dev');
execFileSync(binary, process.argv.slice(2), { stdio: 'inherit' });
```

- [ ] **Step 6: Implement GitHub Action**

```yaml
# .github/action.yml
name: "AI Code Review"
description: "Automated code review using LLMs"
author: "ernedo"
inputs:
  github_token: { required: true }
  provider: { required: false, default: "anthropic" }
  model: { required: false, default: "claude-3-5-sonnet-20241022" }
  api_key_env: { required: false, default: "ANTHROPIC_API_KEY" }
  severity: { required: false, default: "warning" }
runs:
  using: "docker"
  image: "Dockerfile"
```

```dockerfile
# Dockerfile
FROM gcr.io/distroless/static-debian12
COPY ai-dev /ai-dev
ENTRYPOINT ["/ai-dev"]
CMD ["review"]
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `node --test scripts/download-binary.test.js` (after manual version tag)
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add .goreleaser.yaml .github/workflows/release.yml package.json scripts/ bin/ Dockerfile .github/action.yml
git commit -m "chore(release): GoReleaser, npm package, GitHub Action, Docker"
```

---

### Task 13: Documentation Site (VitePress)

**Files:**
- Create: `docs/` (full VitePress site per spec)
- Create: `.github/workflows/docs.yml`
- Test: `npm run docs:build` (in docs/)

**Interfaces:**
- Consumes: VitePress, markdown content
- Produces: Static site deployed to GitHub Pages

- [ ] **Step 1: Write the failing test**

```bash
# Just verify build works
```

- [ ] **Step 2: Implement docs structure**

```markdown
# docs/index.md
---
layout: home
hero:
  name: ai-dev-tools
  text: AI-Powered Developer Tools
  tagline: Commit, Review, Test, Changelog — one CLI
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: GitHub
      link: https://github.com/ernedo/ai-dev-tools
features:
  - title: 🤖 AI Commit Messages
    details: Conventional commits from your diff in seconds
  - title: 🔍 Code Review
    details: Multi-pass security, bugs, style, perf analysis
  - title: 🧪 Test Generation
    details: Table-driven tests for Go, JS/TS, Python
  - title: 📝 Changelogs
    details: Keep a Changelog format from git history
---

# docs/guide/getting-started.md, configuration.md, providers.md, etc.
# docs/tools/commit.md, review.md, testgen.md, changelog.md
# docs/reference/cli.md, config-schema.md, templates.md
```

- [ ] **Step 3: Configure VitePress**

```javascript
// docs/.vitepress/config.mjs
export default {
    title: "ai-dev-tools",
    description: "AI-Powered Developer Tools",
    themeConfig: {
        nav: [
            { text: "Guide", link: "/guide/getting-started" },
            { text: "Tools", items: [
                { text: "Commit", link: "/tools/commit" },
                { text: "Review", link: "/tools/review" },
                { text: "TestGen", link: "/tools/testgen" },
                { text: "Changelog", link: "/tools/changelog" },
            ]},
            { text: "Reference", link: "/reference/cli" },
            { text: "GitHub", link: "https://github.com/ernedo/ai-dev-tools" },
        ],
        sidebar: { /* auto-generated from structure */ },
    },
}
```

- [ ] **Step 4: Deploy workflow**

```yaml
# .github/workflows/docs.yml
name: Deploy Docs
on:
  push:
    branches: [main]
    paths: ['docs/**']
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd docs && npm ci && npm run build
      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: docs/.vitepress/dist
```

- [ ] **Step 5: Commit**

```bash
git add docs/ .github/workflows/docs.yml
git commit -m "docs: VitePress documentation site with full guides"
```

---

### Task 14: CI Pipeline & Benchmarks

**Files:**
- Create: `.github/workflows/ci.yml`, `.github/workflows/bench.yml`
- Create: `benchmarks/` (tracked in separate branch)
- Test: Push to trigger CI

**Interfaces:**
- Consumes: GitHub Actions, Go test, benchmarks
- Produces: Pass/fail status, benchmark trends

- [ ] **Step 1: Implement CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: go mod download
      - run: go vet ./...
      - run: golangci-lint run
      - run: go test ./... -race -coverprofile=coverage.out
      - uses: codecov/codecov-action@v4
  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: go build -o ai-dev ./cmd/ai-dev
      - run: ./ai-dev version
  integration:
    runs-on: ubuntu-latest
    needs: test
    services:
      ollama:
        image: ollama/ollama:latest
        ports: [11434:11434]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test ./... -tags=integration -v
```

- [ ] **Step 2: Implement benchmark workflow**

```yaml
# .github/workflows/bench.yml
name: Benchmarks
on:
  schedule: [{ cron: '0 0 * * 0' }] # weekly
  workflow_dispatch:
jobs:
  bench:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test ./pkg/... -bench=. -benchtime=5s -benchmem > bench.txt
      - run: |
          git config user.name "github-actions"
          git config user.email "actions@github.com"
          git checkout -b benchmarks
          mkdir -p benchmarks
          mv bench.txt benchmarks/$(date +%Y-%m-%d).txt
          git add benchmarks/
          git commit -m "chore(bench): $(date +%Y-%m-%d)"
          git push origin benchmarks --force
```

- [ ] **Step 3: Add benchmarks to packages**

```go
// pkg/llm/openai_bench_test.go
func BenchmarkOpenAI_Complete(b *testing.B) {
    p := NewOpenAI(Config{APIKey: "test", Model: "gpt-4o-mini"})
    req := CompletionRequest{Messages: []Message{{Role: "user", Content: "Hello"}}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p.Complete(context.Background(), req)
    }
}
```

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/ci.yml .github/workflows/bench.yml
git commit -m "ci: CI pipeline with lint, test, build, integration; weekly benchmarks"
```

---

### Task 15: Install Script & Polish

**Files:**
- Create: `scripts/install.sh`, `scripts/install.ps1`
- Create: `README.md`, `LICENSE`, `CONTRIBUTING.md`, `SECURITY.md`
- Test: `bash scripts/install.sh` (dry-run)

**Interfaces:**
- Consumes: GitHub Releases API
- Produces: One-line install for users

- [ ] **Step 1: Implement install script**

```bash
# scripts/install.sh
#!/bin/bash
set -euo pipefail
REPO="ernedo/ai-dev-tools"
BINARY="ai-dev"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
# Detect OS/arch, fetch latest release asset, install
# Supports: curl | sh, brew, npm, scoop
```

```powershell
# scripts/install.ps1
# PowerShell equivalent for Windows
```

- [ ] **Step 2: Implement README**

```markdown
# ai-dev-tools
AI-powered developer tools: commit, review, testgen, changelog.

## Install
```bash
# macOS/Linux
curl -sfL https://get.ai-dev.tools | sh

# Homebrew
brew tap ernedo/tap && brew install ai-dev-tools

# npm
npm i -g @ernedo/ai-dev-tools

# Windows (Scoop)
scoop bucket add ernedo https://github.com/ernedo/scoop-bucket
scoop install ai-dev-tools
```

## Quick Start
```bash
ai-dev config init    # Configure provider (Anthropic, OpenAI, Ollama)
ai-dev commit         # Generate commit message
ai-dev review --pr 42 # Review PR #42
ai-dev testgen .      # Generate tests
ai-dev changelog      # Update CHANGELOG.md
```

## GitHub Action
```yaml
- uses: ernedo/ai-dev-tools@review
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    provider: anthropic
```

## License
MIT
```

- [ ] **Step 3: Commit**

```bash
git add scripts/install.sh scripts/install.ps1 README.md LICENSE CONTRIBUTING.md SECURITY.md
git commit -m "docs: install script, README, contributing, security policy"
```

---

### Task 16: E2E Tests & Final Integration

**Files:**
- Create: `e2e/*_test.go` (full CLI workflows)
- Test: `go test ./e2e/... -v`

**Interfaces:**
- Consumes: All previous tasks
- Produces: Confidence that suite works end-to-end

- [ ] **Step 1: Write E2E tests**

```go
// e2e/suite_test.go
func TestE2E_FullWorkflow(t *testing.T) {
    tmp := t.TempDir()
    // 1. Init repo, make changes
    // 2. ai-dev config init (with Ollama)
    // 3. ai-dev commit --dry-run
    // 4. ai-dev review --base main --head feature --output json
    // 5. ai-dev testgen ./pkg/... --output file
    // 6. ai-dev changelog --since v1.0.0 --output CHANGELOG.md
    // 7. Verify all outputs correct
}
func TestE2E_GitHubAction_Local(t *testing.T) {
    // Run review via act (GitHub Action locally) or docker
}
```

- [ ] **Step 2: Run E2E tests**

Run: `go test ./e2e/... -v -tags=e2e`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add e2e/
git commit -m "test: E2E suite covering full CLI workflow"
```

---

### Task 17: v0.1.0 Release

**Files:**
- Tag: `v0.1.0`
- Trigger: GoReleaser release workflow

**Interfaces:**
- Consumes: All previous tasks
- Produces: Public release on GitHub, Homebrew, npm

- [ ] **Step 1: Final verification**

Run: `go test ./... -race && go build ./cmd/ai-dev && ./ai-dev doctor`

- [ ] **Step 2: Create release tag**

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

- [ ] **Step 3: Verify release artifacts**

- GitHub Releases: binaries for 6 platforms, checksums
- Homebrew: `brew install ernedo/tap/ai-dev-tools`
- npm: `npm i -g @ernedo/ai-dev-tools`
- GitHub Action: `uses: ernedo/ai-dev-tools@review`

- [ ] **Step 4: Announce**

- Blog post, Twitter, HN, Reddit (r/golang, r/programming, r/devops)

---

## Execution Order Summary

| Phase | Tasks | Dependencies |
|-------|-------|--------------|
| **Foundation** | 1 | — |
| **Core Libraries** | 2, 3, 4, 5, 6 | 1 |
| **CLI Commands** | 7, 8, 9, 10, 11 | 1, 2, 3, 4, 5, 6 |
| **Distribution** | 12 | 1, 7-11 |
| **Documentation** | 13 | 1, 7-11 |
| **CI/Bench** | 14 | 1, 7-11 |
| **Polish** | 15 | 12, 13 |
| **Integration** | 16 | 7-15 |
| **Release** | 17 | 16 |

---

## Recommended Execution Method

**Subagent-driven** — This plan has 17 tasks with clear interfaces between them. Tasks 2-6 (core libraries) can run in parallel. Tasks 7-11 (CLI commands) depend on core libraries but are independent of each other. Tasks 12-15 (distribution, docs, CI, polish) depend on CLI commands. A subagent per task with independent review ensures each interface is validated before the next task consumes it, catching type mismatches early. The final E2E task (16) validates the full integration.