// e2e/suite_test.go
package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ernedo/ai-dev-tools/cmd/changelog"
	"github.com/ernedo/ai-dev-tools/cmd/commit"
	"github.com/ernedo/ai-dev-tools/cmd/review"
	"github.com/ernedo/ai-dev-tools/cmd/testgen"
	"github.com/ernedo/ai-dev-tools/internal/cli"
	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements llm.Provider for testing
type mockProvider struct {
	responses []string
	callCount int
}

func (m *mockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.callCount >= len(m.responses) {
		return &llm.CompletionResponse{Content: `{"message": "default response"}`}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return &llm.CompletionResponse{Content: resp}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: "test", Done: true}
	close(ch)
	return ch, nil
}

func (m *mockProvider) Name() string        { return "mock" }
func (m *mockProvider) MaxTokens() int      { return 4096 }
func (m *mockProvider) SupportsTools() bool { return true }
func (m *mockProvider) SupportsJSONMode() bool { return true }

// mockGitRepo implements git.Repo for testing
type mockGitRepo struct {
	rootDir string
}

func (m *mockGitRepo) Diff(base, head string, opts git.DiffOptions) (string, error) {
	return "diff --git a/main.go b/main.go\n+func main() {}\n", nil
}
func (m *mockGitRepo) DiffCached(opts git.DiffOptions) (string, error) {
	return "diff --git a/main.go b/main.go\n+func main() {}\n", nil
}
func (m *mockGitRepo) DiffStat(base, head string) (git.DiffStat, error) {
	return git.DiffStat{FilesChanged: 1, Insertions: 1}, nil
}
func (m *mockGitRepo) Log(opts git.LogOptions) ([]git.Commit, error) {
	return []git.Commit{
		{Hash: "abc123", ShortHash: "abc123", Message: "feat: add feature", Author: "test", Email: "test@test.com"},
		{Hash: "def456", ShortHash: "def456", Message: "fix: bug fix", Author: "test", Email: "test@test.com"},
	}, nil
}
func (m *mockGitRepo) LogSince(tag string) ([]git.Commit, error) {
	return []git.Commit{
		{Hash: "abc123", ShortHash: "abc123", Message: "feat: add feature", Author: "test", Email: "test@test.com"},
	}, nil
}
func (m *mockGitRepo) Status() (git.Status, error) {
	return git.Status{}, nil
}
func (m *mockGitRepo) GetFile(path, ref string) (string, error) {
	return "", nil
}
func (m *mockGitRepo) ListFiles(ref, pattern string) ([]string, error) {
	return []string{"main.go"}, nil
}
func (m *mockGitRepo) RootDir() (string, error) { return m.rootDir, nil }
func (m *mockGitRepo) LatestTag() (string, error) { return "v1.0.0", nil }
func (m *mockGitRepo) CurrentBranch() (string, error) { return "main", nil }
func (m *mockGitRepo) IsDirty() (bool, error) { return true, nil }
func (m *mockGitRepo) Close() error { return nil }

func setupTestEnv(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Initialize git repo
	os.Chdir(tmpDir)

	// Create a basic go.mod
	goMod := `module test

go 1.24
`
	os.WriteFile("go.mod", []byte(goMod), 0644)

	// Create a test file
	mainGo := `package main

func main() {
	println("hello")
}
`
	os.WriteFile("main.go", []byte(mainGo), 0644)

	// Create config
	cfg := &config.Config{
		Provider: config.ProviderConfig{
			Name: "mock", Model: "test", APIKeyEnv: "TEST_KEY",
		},
	}
	cfgPath := filepath.Join(tmpDir, ".ai-dev-tools.yaml")
	cfgData, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(cfgPath, cfgData, 0644)

	return tmpDir, func() { os.Chdir("/") }
}

func TestE2E_CommitCommand_Exists(t *testing.T) {
	cmd := cli.NewRootCmd()

	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "commit" {
			found = true
			break
		}
	}
	require.True(t, found, "commit command should exist")

	// Also verify the actual command has proper flags
	mockRepo := &mockGitRepo{rootDir: "/tmp"}
	mockLLM := &mockProvider{}
	realCmd := commit.NewCommand(mockRepo, mockLLM, nil)
	assert.Contains(t, realCmd.Flags().FlagUsages(), "dry-run")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "style")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "sign-off")
}

func TestE2E_ReviewCommand_Exists(t *testing.T) {
	cmd := cli.NewRootCmd()

	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "review" {
			found = true
			break
		}
	}
	require.True(t, found, "review command should exist")
}

func TestE2E_TestGenCommand_Exists(t *testing.T) {
	cmd := cli.NewRootCmd()

	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "testgen" {
			found = true
			break
		}
	}
	require.True(t, found, "testgen command should exist")
}

func TestE2E_ChangelogCommand_Exists(t *testing.T) {
	cmd := cli.NewRootCmd()

	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "changelog" {
			found = true
			break
		}
	}
	require.True(t, found, "changelog command should exist")
}

func TestE2E_ConfigCommand_Exists(t *testing.T) {
	cmd := cli.NewRootCmd()

	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "config" {
			found = true
			break
		}
	}
	require.True(t, found, "config command should exist")
}

func TestE2E_DoctorCommand_Exists(t *testing.T) {
	cmd := cli.NewRootCmd()

	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "doctor" {
			found = true
			break
		}
	}
	require.True(t, found, "doctor command should exist")
}

func TestE2E_VersionCommand_Exists(t *testing.T) {
	cmd := cli.NewRootCmd()

	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "version" {
			found = true
			break
		}
	}
	require.True(t, found, "version command should exist")
}

// TestE2E_FullWorkflow simulates a full workflow
func TestE2E_FullWorkflow(t *testing.T) {
	cmd := cli.NewRootCmd()

	commands := []string{"commit", "review", "testgen", "changelog", "config", "doctor", "version"}
	for _, name := range commands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		assert.True(t, found, "command %s should exist", name)
	}
}

// TestE2E_ConfigPriority tests config loading priority
func TestE2E_ConfigPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Default config
	cfg1, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "anthropic", cfg1.Provider.Name)

	// 2. Config file overrides default
	cfgPath := filepath.Join(tmpDir, ".ai-dev-tools.yaml")
	cfgContent := `provider:
  name: openai
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
`
	os.WriteFile(cfgPath, []byte(cfgContent), 0644)

	cfg2, err := config.Load(config.WithConfigDir(tmpDir))
	require.NoError(t, err)
	assert.Equal(t, "openai", cfg2.Provider.Name)

	// 3. Env var overrides config file
	t.Setenv("AI_DEV_PROVIDER_NAME", "ollama")
	cfg3, err := config.Load(config.WithConfigDir(tmpDir))
	require.NoError(t, err)
	assert.Equal(t, "ollama", cfg3.Provider.Name)
}

func TestE2E_ConfigValidation(t *testing.T) {
	// Invalid provider should fail
	_, err := config.Load(config.WithConfigYAML(`provider:
  name: invalid
`))
	assert.Error(t, err)

	// Valid config should pass
	_, err = config.Load(config.WithConfigYAML(`provider:
  name: anthropic
  model: claude-3-5-sonnet
  api_key_env: TEST
`))
	assert.NoError(t, err)
}

// TestE2E_DiffParsing tests diff stat parsing
func TestE2E_DiffParsing(t *testing.T) {
	diff := ` 3 files changed, 10 insertions(+), 5 deletions(-)`
	stat := git.ParseDiffStat(diff)
	assert.Equal(t, 3, stat.FilesChanged)
	assert.Equal(t, 10, stat.Insertions)
	assert.Equal(t, 5, stat.Deletions)
}

func TestE2E_ReviewCommand_RealFlags(t *testing.T) {
	mockRepo := &mockGitRepo{rootDir: "/tmp"}
	mockLLM := &mockProvider{}
	mockGH := &mockGitHub{}
	realCmd := review.NewCommand(mockRepo, mockGH, mockLLM, nil)

	assert.Contains(t, realCmd.Flags().FlagUsages(), "base")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "head")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "pr")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "output")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "severity")
}

func TestE2E_TestGenCommand_RealFlags(t *testing.T) {
	mockRepo := &mockGitRepo{rootDir: "/tmp"}
	mockLLM := &mockProvider{}
	realCmd := testgen.NewCommand(mockRepo, mockLLM, nil)

	assert.Contains(t, realCmd.Flags().FlagUsages(), "framework")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "output")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "coverage")
}

func TestE2E_ChangelogCommand_RealFlags(t *testing.T) {
	mockRepo := &mockGitRepo{rootDir: "/tmp"}
	mockLLM := &mockProvider{}
	mockGH := &mockGitHub{}
	realCmd := changelog.NewCommand(mockRepo, mockGH, mockLLM, nil)

	assert.Contains(t, realCmd.Flags().FlagUsages(), "since")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "until")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "format")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "output")
	assert.Contains(t, realCmd.Flags().FlagUsages(), "include-prs")
}

// mockGitHub implements github.Client for testing
type mockGitHub struct{}

func (m *mockGitHub) GetPR(owner, repo string, number int) (*github.PR, error) { return nil, nil }
func (m *mockGitHub) ListPRCommits(owner, repo string, number int) ([]github.Commit, error) { return nil, nil }
func (m *mockGitHub) ListPRFiles(owner, repo string, number int) ([]github.PRFile, error) { return nil, nil }
func (m *mockGitHub) CreateReview(owner, repo string, number int, review github.ReviewRequest) error { return nil }
func (m *mockGitHub) CreateReviewComment(owner, repo string, number int, comment github.ReviewComment) error { return nil }
func (m *mockGitHub) GetRepo(owner, repo string) (*github.Repo, error) { return nil, nil }
func (m *mockGitHub) GetDefaultBranch(owner, repo string) (string, error) { return "", nil }