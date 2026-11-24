// cmd/testgen/command_test.go
package testgen

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

// setupGoRepo creates a temporary git repo with a Go handler file for testing
func setupGoRepo(t *testing.T) git.Repo {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo
	execGit(t, tmpDir, "init", "--initial-branch=main")

	// Configure git user
	execGit(t, tmpDir, "config", "user.email", "test@example.com")
	execGit(t, tmpDir, "config", "user.name", "Test User")

	// Create a Go handler file
	handlerCode := `package main

import (
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Path == "/api/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	http.NotFound(w, r)
}
`
	handlerFile := filepath.Join(tmpDir, "handler.go")
	os.WriteFile(handlerFile, []byte(handlerCode), 0644)
	execGit(t, tmpDir, "add", "handler.go")
	execGit(t, tmpDir, "commit", "-m", "add handler")

	// Open the repo
	repo, err := git.Open(tmpDir)
	require.NoError(t, err)

	return repo
}

// setupTSRepo creates a temporary git repo with a TypeScript component for testing
func setupTSRepo(t *testing.T) git.Repo {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo
	execGit(t, tmpDir, "init", "--initial-branch=main")

	// Configure git user
	execGit(t, tmpDir, "config", "user.email", "test@example.com")
	execGit(t, tmpDir, "config", "user.name", "Test User")

	// Create a TypeScript component
	componentCode := `import React from 'react';

interface Props {
  title: string;
}

export function Component({ title }: Props) {
  return <h1>{title}</h1>;
}

export default Component;
`
	componentFile := filepath.Join(tmpDir, "component.tsx")
	os.WriteFile(componentFile, []byte(componentCode), 0644)

	// Create jest.config.js for framework detection
	jestConfig := `module.exports = {
  testEnvironment: 'jsdom',
};
`
	jestConfigFile := filepath.Join(tmpDir, "jest.config.js")
	os.WriteFile(jestConfigFile, []byte(jestConfig), 0644)

	execGit(t, tmpDir, "add", "component.tsx", "jest.config.js")
	execGit(t, tmpDir, "commit", "-m", "add component")

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

func goTestResponse() string {
	return `{"tests":[{"name":"TestHandle","cases":[{"name":"success","input":{"method":"GET","path":"/api/health"},"expected":{"status":200,"body":"OK"}}]}]}`
}

func tsTestResponse() string {
	return `{"tests":[{"name":"Component","cases":[{"name":"renders","setup":"render(<Component />)","expect":"toBeInTheDocument()"}]}]}`
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

func TestTestGenCommand_GoGeneratesTableTests(t *testing.T) {
	repo := setupGoRepo(t)
	llm := &mockProvider{response: goTestResponse()}
	cmd := NewCommand(repo, llm, defaultConfig())
	cmd.SetArgs([]string{"./handler.go", "--output", "file"})

	err := cmd.Execute()
	assert.NoError(t, err)

	rootDir, err := repo.RootDir()
	require.NoError(t, err)
	content, _ := os.ReadFile(filepath.Join(rootDir, "handler_test.go"))
	assert.Contains(t, string(content), "func TestHandle")
	assert.Contains(t, string(content), "testCases := []struct")
}

func TestTestGenCommand_FrameworkAutoDetect(t *testing.T) {
	repo := setupTSRepo(t)
	llm := &mockProvider{response: tsTestResponse()}
	cmd := NewCommand(repo, llm, defaultConfig())
	cmd.SetArgs([]string{"./component.tsx"})

	err := cmd.Execute()
	assert.NoError(t, err)

	rootDir, err := repo.RootDir()
	require.NoError(t, err)
	content, _ := os.ReadFile(filepath.Join(rootDir, "component.test.tsx"))
	assert.Contains(t, string(content), "describe(")
	assert.Contains(t, string(content), "toBeInTheDocument()")
}