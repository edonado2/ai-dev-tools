package review

import (
	"bytes"
	"context"
	"encoding/json"
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
	callCount int
}

func (m *mockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	m.callCount++
	// Return findings only on first call to simulate single-pass review
	if m.callCount > 1 {
		return &llm.CompletionResponse{
			Content: `{"summary":"","findings":[],"approved":true}`,
			Model:   "test",
			Usage:   llm.Usage{TotalTokens: 10},
		}, nil
	}
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
	comments []github.ReviewComment
	reviews  []github.ReviewRequest
}

func (m *mockGitHub) GetPR(owner, repo string, number int) (*github.PR, error) {
	return &github.PR{Number: number}, nil
}

func (m *mockGitHub) ListPRCommits(owner, repo string, number int) ([]github.Commit, error) {
	return nil, nil
}

func (m *mockGitHub) ListPRFiles(owner, repo string, number int) ([]github.PRFile, error) {
	return []github.PRFile{
		{
			Filename: "test.go",
			Patch:    "@@ -1,3 +1,4 @@\n func Test() {\n+    fmt.Println(\"hello\")\n }\n",
			Status: "modified",
		},
	}, nil
}

func (m *mockGitHub) CreateReview(owner, repo string, number int, review github.ReviewRequest) error {
	m.reviews = append(m.reviews, review)
	return nil
}

func (m *mockGitHub) CreateReviewComment(owner, repo string, number int, comment github.ReviewComment) error {
	m.comments = append(m.comments, comment)
	return nil
}

func (m *mockGitHub) GetRepo(owner, repo string) (*github.Repo, error) {
	return &github.Repo{Owner: owner, Name: repo}, nil
}

func (m *mockGitHub) GetDefaultBranch(owner, repo string) (string, error) {
	return "main", nil
}

// setupGitRepo creates a temporary git repo with changes for testing
func setupGitRepo(t *testing.T) git.Repo {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo with main branch
	execGit(t, tmpDir, "init", "--initial-branch=main")

	// Configure git user
	execGit(t, tmpDir, "config", "user.email", "test@example.com")
	execGit(t, tmpDir, "config", "user.name", "Test User")

	// Create a file and commit it
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte("package main\n\nfunc Test() {\n}\n"), 0644)
	execGit(t, tmpDir, "add", "test.go")
	execGit(t, tmpDir, "commit", "-m", "initial commit")

	// Create a feature branch
	execGit(t, tmpDir, "checkout", "-b", "feature")

	// Create a change
	os.WriteFile(testFile, []byte("package main\n\nfunc Test() {\n\tfmt.Println(\"hello\")\n}\n"), 0644)
	execGit(t, tmpDir, "add", "test.go")
	execGit(t, tmpDir, "commit", "-m", "add hello")

	// Create another change for diff
	os.WriteFile(testFile, []byte("package main\n\nfunc Test() {\n\tfmt.Println(\"hello world\")\n}\n"), 0644)
	execGit(t, tmpDir, "add", "test.go")

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
			Review: config.ReviewConfig{
				SeverityThreshold: "warning",
				ExcludePatterns:   []string{},
				IncludePatterns:   []string{},
			},
		},
	}
}

func sarifResponse() string {
	return `{
		"summary": "Code review completed",
		"findings": [
			{
				"file": "test.go",
				"line": 3,
				"severity": "warning",
				"message": "Missing error handling",
				"suggestion": "Add error handling for fmt.Println"
			}
		],
		"approved": false
	}`
}

func reviewResponse() string {
	return `{
		"summary": "Code review completed",
		"findings": [
			{
				"file": "test.go",
				"line": 3,
				"severity": "error",
				"message": "Security issue: potential injection",
				"suggestion": "Validate input"
			},
			{
				"file": "test.go",
				"line": 4,
				"severity": "warning",
				"message": "Style: missing comment",
				"suggestion": "Add function comment"
			},
			{
				"file": "test.go",
				"line": 5,
				"severity": "info",
				"message": "Performance: could optimize",
				"suggestion": "Use buffered I/O"
			}
		],
		"approved": false
	}`
}

func TestReviewCommand_LocalOutputsSARIF(t *testing.T) {
	repo := setupGitRepo(t)
	llm := &mockProvider{response: sarifResponse()}
	gh := &mockGitHub{}
	cfg := defaultConfig()

	cmd := NewCommand(repo, gh, llm, cfg)
	cmd.SetArgs([]string{"--base", "main", "--head", "HEAD", "--output", "sarif"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	// Don't capture stderr to avoid mixing with JSON output

	err := cmd.Execute()
	require.NoError(t, err)

	var sarif map[string]any
	err = json.Unmarshal(out.Bytes(), &sarif)
	require.NoError(t, err)

	assert.Equal(t, "2.1.0", sarif["version"])
	runs := sarif["runs"].([]any)
	assert.Len(t, runs, 1)
}

func TestReviewCommand_GitHubAction_PostsComments(t *testing.T) {
	repo := setupGitRepo(t)
	llm := &mockProvider{response: reviewResponse()}
	gh := &mockGitHub{comments: []github.ReviewComment{}}
	cfg := defaultConfig()

	cmd := NewCommand(repo, gh, llm, cfg)
	cmd.SetArgs([]string{"--pr", "42", "--owner", "testowner", "--repo", "testrepo", "--output", "github"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Should have created review with comments
	assert.Len(t, gh.reviews, 1)
	assert.Equal(t, "REQUEST_CHANGES", gh.reviews[0].Event)
	assert.Len(t, gh.reviews[0].Comments, 3)
}