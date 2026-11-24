// cmd/commit/command_test.go
package commit

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/ernedo/ai-dev-tools/pkg/prompt"
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

// setupGitRepo creates a temporary git repo with staged changes for testing
func setupGitRepo(t *testing.T) git.Repo {
	t.Helper()
	t.Logf("setupGitRepo: creating temp dir")
	tmpDir := t.TempDir()
	t.Logf("setupGitRepo: temp dir: %s", tmpDir)

	// Initialize git repo
	t.Logf("setupGitRepo: git init")
	execGit(t, tmpDir, "init")
	t.Logf("setupGitRepo: git init done")

	// Configure git user
	t.Logf("setupGitRepo: git config")
	execGit(t, tmpDir, "config", "user.email", "test@example.com")
	execGit(t, tmpDir, "config", "user.name", "Test User")
	t.Logf("setupGitRepo: git config done")

	// Create a file and commit it
	t.Logf("setupGitRepo: creating file")
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("initial content"), 0644)
	t.Logf("setupGitRepo: git add")
	execGit(t, tmpDir, "add", "test.txt")
	t.Logf("setupGitRepo: git commit")
	execGit(t, tmpDir, "commit", "-m", "initial commit")
	t.Logf("setupGitRepo: git commit done")

	// Create a staged change
	t.Logf("setupGitRepo: modifying file")
	os.WriteFile(testFile, []byte("modified content"), 0644)
	t.Logf("setupGitRepo: git add modified")
	execGit(t, tmpDir, "add", "test.txt")
	t.Logf("setupGitRepo: git add done")

	// Open the repo
	t.Logf("setupGitRepo: opening repo")
	repo, err := git.Open(tmpDir)
	require.NoError(t, err)
	t.Logf("setupGitRepo: repo opened")

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
func defaultConfig() *Config {
	return &Config{
		Tools: ToolsConfig{
			Commit: CommitConfig{
				Style: "conventional",
			},
		},
	}
}

func TestCommitCommand_DryRun_PrintsMessage(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("PANIC in test: %v", r)
		}
	}()

	t.Logf("TEST STARTED")

	// Test prompt.NewRegistry() in isolation
	t.Logf("Creating registry...")
	reg := prompt.NewRegistry()
	t.Logf("Registry created")
	tpl, err := reg.Get("commit", "1")
	if err != nil {
		t.Logf("Error getting template: %v", err)
	} else {
		t.Logf("Got template: %s v%s", tpl.Name, tpl.Version)
	}

	t.Logf("About to setup repo...")
	t.Logf("Calling setupGitRepo now...")
	repo := setupGitRepo(t)
	t.Logf("Repo setup done")

	// Debug: check what the repo sees
	status, _ := repo.Status()
	t.Logf("Status: Staged=%v, Unstaged=%v, Untracked=%v", status.Staged, status.Unstaged, status.Untracked)

	diff, _ := repo.DiffCached(git.DiffOptions{})
	t.Logf("DiffCached length: %d", len(diff))

	llm := &mockProvider{response: `{"message":"feat(cli): add commit cmd","type":"feat","scope":"cli"}`}
	cmd := NewCommand(repo, llm, &Config{Tools: ToolsConfig{Commit: CommitConfig{Style: "conventional"}}})

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--dry-run"})

	t.Logf("BEFORE EXECUTE")
	err = cmd.Execute()
	t.Logf("AFTER EXECUTE")
	t.Logf("Stderr: %s", errBuf.String())
	if err != nil {
		t.Logf("Error: %v", err)
	}
	assert.NoError(t, err)
	t.Logf("Output: %s", out.String())
	assert.Contains(t, out.String(), "feat(cli): add commit cmd")
}

func TestCommitCommand_Interactive_CommitsOnConfirm(t *testing.T) {
	repo := setupGitRepo(t)
	llm := &mockProvider{response: `{"message":"fix: bug","type":"fix"}`}
	cmd := NewCommand(repo, llm, defaultConfig())

	// Simulate "y" input
	cmd.SetIn(strings.NewReader("y\n"))

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	// Verify commit was created
	commits, err := repo.Log(git.LogOptions{MaxCount: 1})
	require.NoError(t, err)
	assert.Contains(t, commits[0].Message, "fix: bug")
}