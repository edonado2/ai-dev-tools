// cmd/changelog/command_test.go
package changelog

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements llm.Provider for testing
type mockProvider struct {
	response string
}

func (m *mockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{
		Content: m.response,
		Model:   "test",
		Usage:   llm.Usage{TotalTokens: 100},
	}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req llm.CompletionRequest) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Content: m.response, Done: true}
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
	prs    map[int]*github.PR
	commits map[int][]github.Commit
	files   map[int][]github.PRFile
}

func (m *mockGitHub) GetPR(owner, repo string, number int) (*github.PR, error) {
	if m.prs != nil {
		return m.prs[number], nil
	}
	return &github.PR{Number: number}, nil
}

func (m *mockGitHub) ListPRCommits(owner, repo string, number int) ([]github.Commit, error) {
	if m.commits != nil {
		return m.commits[number], nil
	}
	return nil, nil
}

func (m *mockGitHub) ListPRFiles(owner, repo string, number int) ([]github.PRFile, error) {
	if m.files != nil {
		return m.files[number], nil
	}
	return nil, nil
}

func (m *mockGitHub) CreateReview(owner, repo string, number int, review github.ReviewRequest) error {
	return nil
}

func (m *mockGitHub) CreateReviewComment(owner, repo string, number int, comment github.ReviewComment) error {
	return nil
}

func (m *mockGitHub) GetRepo(owner, repo string) (*github.Repo, error) {
	return &github.Repo{Owner: owner, Name: repo}, nil
}

func (m *mockGitHub) GetDefaultBranch(owner, repo string) (string, error) {
	return "main", nil
}

// setupGitRepo creates a temporary git repo with conventional commits since v1.0.0 for testing
func setupGitRepo(t *testing.T) git.Repo {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo with main branch
	execGit(t, tmpDir, "init", "--initial-branch=main")

	// Configure git user
	execGit(t, tmpDir, "config", "user.email", "test@example.com")
	execGit(t, tmpDir, "config", "user.name", "Test User")

	// Create initial commit and tag v1.0.0
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("initial content"), 0644)
	execGit(t, tmpDir, "add", "test.txt")
	execGit(t, tmpDir, "commit", "-m", "chore: initial commit")
	execGit(t, tmpDir, "tag", "v1.0.0")

	// Add conventional commits after v1.0.0
	os.WriteFile(testFile, []byte("feat: new feature"), 0644)
	execGit(t, tmpDir, "add", "test.txt")
	execGit(t, tmpDir, "commit", "-m", "feat: new feature")

	os.WriteFile(testFile, []byte("fix: bug fix"), 0644)
	execGit(t, tmpDir, "add", "test.txt")
	execGit(t, tmpDir, "commit", "-m", "fix: bug fix")

	// Open the repo
	repo, err := git.Open(tmpDir)
	require.NoError(t, err)

	return repo
}

func execGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("git %s failed: %v: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String()
}

// defaultConfig returns a default configuration for testing
func defaultConfig() *config.Config {
	return &config.Config{
		Tools: config.ToolsConfig{
			Changelog: config.ChangelogConfig{
				Format:      "keepachangelog",
				IncludePRs:  false,
				Output:      "",
			},
		},
	}
}

func changelogResponse() string {
	return `{
    "version": "1.1.0",
    "date": "2024-01-15",
    "sections": {
        "Added": ["feat: new feature"],
        "Fixed": ["fix: bug fix"],
        "Changed": [],
        "Removed": []
    }
}`
}

func TestChangelogCommand_KeepAChangelogFormat(t *testing.T) {
	repo := setupGitRepo(t)
	llm := &mockProvider{response: changelogResponse()}
	cmd := NewCommand(repo, nil, llm, defaultConfig())
	cmd.SetArgs([]string{"--since", "v1.0.0", "--format", "keepachangelog"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "## [1.1.0]")
	assert.Contains(t, output, "### Added")
	assert.Contains(t, output, "### Fixed")
	assert.Contains(t, output, "feat: new feature")
}

func TestChangelogCommand_UpdatesFile(t *testing.T) {
	repo := setupGitRepo(t)
	llm := &mockProvider{response: changelogResponse()}
	cmd := NewCommand(repo, nil, llm, defaultConfig())
	cmd.SetArgs([]string{"--output", "CHANGELOG.md"})
	err := cmd.Execute()
	assert.NoError(t, err)
	content, _ := os.ReadFile("CHANGELOG.md")
	assert.Contains(t, string(content), "## [Unreleased]")
}
