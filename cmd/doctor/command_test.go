package doctor

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements llm.Provider for testing
type mockProvider struct {
	shouldFail bool
}

func (m *mockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return &llm.CompletionResponse{
		Content: "test response",
		Model:   "test",
		Usage:   llm.Usage{TotalTokens: 10},
	}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "test", Done: true}
	close(ch)
	return ch, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) MaxTokens() int {
	return 4096
}

func (m *mockProvider) SupportsTools() bool {
	return false
}

func (m *mockProvider) SupportsJSONMode() bool {
	return true
}

// mockGitHub implements github.Client for testing
type mockGitHub struct {
	shouldFail bool
}

func (m *mockGitHub) GetPR(owner, repo string, number int) (*github.PR, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return &github.PR{Number: number}, nil
}

func (m *mockGitHub) ListPRCommits(owner, repo string, number int) ([]github.Commit, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return nil, nil
}

func (m *mockGitHub) ListPRFiles(owner, repo string, number int) ([]github.PRFile, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return nil, nil
}

func (m *mockGitHub) CreateReview(owner, repo string, number int, review github.ReviewRequest) error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

func (m *mockGitHub) CreateReviewComment(owner, repo string, number int, comment github.ReviewComment) error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

func (m *mockGitHub) GetRepo(owner, repo string) (*github.Repo, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return &github.Repo{Owner: owner, Name: repo}, nil
}

func (m *mockGitHub) GetDefaultBranch(owner, repo string) (string, error) {
	if m.shouldFail {
		return "", assert.AnError
	}
	return "main", nil
}

// mockRepo implements git.Repo for testing
type mockRepo struct {
	shouldFail bool
}

func (m *mockRepo) Diff(base, head string, opts git.DiffOptions) (string, error) {
	if m.shouldFail {
		return "", assert.AnError
	}
	return "", nil
}

func (m *mockRepo) DiffCached(opts git.DiffOptions) (string, error) {
	if m.shouldFail {
		return "", assert.AnError
	}
	return "", nil
}

func (m *mockRepo) DiffStat(base, head string) (git.DiffStat, error) {
	if m.shouldFail {
		return git.DiffStat{}, assert.AnError
	}
	return git.DiffStat{}, nil
}

func (m *mockRepo) Log(opts git.LogOptions) ([]git.Commit, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return nil, nil
}

func (m *mockRepo) LogSince(tag string) ([]git.Commit, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return nil, nil
}

func (m *mockRepo) Status() (git.Status, error) {
	if m.shouldFail {
		return git.Status{}, assert.AnError
	}
	return git.Status{}, nil
}

func (m *mockRepo) GetFile(path, ref string) (string, error) {
	if m.shouldFail {
		return "", assert.AnError
	}
	return "", nil
}

func (m *mockRepo) ListFiles(ref, pattern string) ([]string, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return nil, nil
}

func (m *mockRepo) RootDir() (string, error) {
	if m.shouldFail {
		return "", assert.AnError
	}
	return "/tmp", nil
}

func (m *mockRepo) LatestTag() (string, error) {
	if m.shouldFail {
		return "", assert.AnError
	}
	return "", nil
}

func (m *mockRepo) CurrentBranch() (string, error) {
	if m.shouldFail {
		return "", assert.AnError
	}
	return "main", nil
}

func (m *mockRepo) IsDirty() (bool, error) {
	if m.shouldFail {
		return false, assert.AnError
	}
	return false, nil
}

func (m *mockRepo) Close() error {
	return nil
}

func validConfig() *config.Config {
	return &config.Config{
		Provider: config.ProviderConfig{
			Name:      "anthropic",
			Model:     "claude-3-5-sonnet",
			APIKeyEnv: "ANTHROPIC_API_KEY",
		},
		GitHub: config.GitHubConfig{
			TokenEnv: "GITHUB_TOKEN",
			Owner:    "testowner",
			Repo:     "testrepo",
		},
	}
}

func NewDoctorCommandForTest(cfg *config.Config, configDir string) *cobra.Command {
	repo := &mockRepo{}
	ghClient := &mockGitHub{}
	llmProvider := &mockProvider{}
	return NewCommand(repo, ghClient, llmProvider, cfg, configDir)
}

func TestDoctor_ChecksConfigAndConnectivity(t *testing.T) {
	tmp := t.TempDir()

	// Create a valid config file
	configContent := `provider:
  name: anthropic
  model: claude-3-5-sonnet
  api_key_env: ANTHROPIC_API_KEY
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewDoctorCommandForTest(validConfig(), tmp)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "✓ Config valid")
	assert.Contains(t, output, "✓ Git repository")
	assert.Contains(t, output, "✓ Anthropic API reachable")
	assert.Contains(t, output, "✓ GitHub authentication")
}

func TestDoctor_ConfigInvalid_ShowsError(t *testing.T) {
	tmp := t.TempDir()

	// Create an invalid config file (missing provider name)
	configContent := `provider:
  model: claude-3-5-sonnet
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		Provider: config.ProviderConfig{
			Model: "claude-3-5-sonnet",
		},
	}
	cmd := NewDoctorCommandForTest(cfg, tmp)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "✗ Config invalid")
}

func TestDoctor_ProviderUnreachable_ShowsError(t *testing.T) {
	tmp := t.TempDir()

	configContent := `provider:
  name: anthropic
  model: claude-3-5-sonnet
  api_key_env: ANTHROPIC_API_KEY
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := validConfig()
	// Use a mock provider that fails
	repo := &mockRepo{}
	ghClient := &mockGitHub{}
	llmProvider := &mockProvider{shouldFail: true}
	cmd := NewCommand(repo, ghClient, llmProvider, cfg, tmp)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "✓ Config valid")
	assert.Contains(t, output, "✗ Anthropic API unreachable")
}

func TestDoctor_GitHubAuthFails_ShowsError(t *testing.T) {
	tmp := t.TempDir()

	configContent := `provider:
  name: anthropic
  model: claude-3-5-sonnet
  api_key_env: ANTHROPIC_API_KEY
github:
  token_env: GITHUB_TOKEN
  owner: testowner
  repo: testrepo
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := validConfig()
	cfg.GitHub = config.GitHubConfig{
		TokenEnv: "GITHUB_TOKEN",
		Owner:    "testowner",
		Repo:     "testrepo",
	}

	// Use a mock GitHub client that fails
	repo := &mockRepo{}
	ghClient := &mockGitHub{shouldFail: true}
	llmProvider := &mockProvider{}
	cmd := NewCommand(repo, ghClient, llmProvider, cfg, tmp)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "✓ Config valid")
	assert.Contains(t, output, "✓ Git repository")
	assert.Contains(t, output, "✓ Anthropic API reachable")
	assert.Contains(t, output, "✗ GitHub authentication failed")
}

func TestDoctor_NotInGitRepo_ShowsError(t *testing.T) {
	tmp := t.TempDir()

	configContent := `provider:
  name: anthropic
  model: claude-3-5-sonnet
  api_key_env: ANTHROPIC_API_KEY
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Use a mock repo that fails on RootDir (simulating not in a git repo)
	repo := &mockRepo{shouldFail: true}
	ghClient := &mockGitHub{}
	llmProvider := &mockProvider{}
	cfg := validConfig()
	cmd := NewCommand(repo, ghClient, llmProvider, cfg, tmp)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "✓ Config valid")
	assert.Contains(t, output, "✗ Git repository")
}

func TestDoctor_SkipsGitHubWhenNotConfigured(t *testing.T) {
	tmp := t.TempDir()

	configContent := `provider:
  name: anthropic
  model: claude-3-5-sonnet
  api_key_env: ANTHROPIC_API_KEY
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Config without GitHub settings
	cfg := &config.Config{
		Provider: config.ProviderConfig{
			Name:      "anthropic",
			Model:     "claude-3-5-sonnet",
			APIKeyEnv: "ANTHROPIC_API_KEY",
		},
	}
	repo := &mockRepo{}
	ghClient := &mockGitHub{}
	llmProvider := &mockProvider{}
	cmd := NewCommand(repo, ghClient, llmProvider, cfg, tmp)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "✓ Config valid")
	assert.Contains(t, output, "✓ Git repository")
	assert.Contains(t, output, "✓ Anthropic API reachable")
	// Should NOT check GitHub when not configured
	assert.NotContains(t, output, "GitHub authentication")
}

func TestDoctor_OllamaProvider_ChecksOllama(t *testing.T) {
	tmp := t.TempDir()

	configContent := `provider:
  name: ollama
  model: llama3
  api_key_env: ""
  base_url: http://localhost:11434
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		Provider: config.ProviderConfig{
			Name:      "ollama",
			Model:     "llama3",
			APIKeyEnv: "",
			BaseURL:   "http://localhost:11434",
		},
	}

	repo := &mockRepo{}
	ghClient := &mockGitHub{}
	llmProvider := &mockProvider{}
	cmd := NewCommand(repo, ghClient, llmProvider, cfg, tmp)
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "✓ Config valid")
	assert.Contains(t, output, "✓ Ollama API reachable")
}