package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary copy of the test fixture repo
func setupTestRepo(t *testing.T) Repo {
	t.Helper()

	// Get the project root directory
	projectRoot := "/Users/ernesto/Documents/projects/ai-dev-tools"
	fixturePath := filepath.Join(projectRoot, "testdata", "git", "fixtures", "testrepo")
	tempDir := t.TempDir()

	// Copy the fixture recursively
	err := copyDir(fixturePath, tempDir)
	require.NoError(t, err)

	// Open the repo
	repo, err := Open(tempDir)
	require.NoError(t, err)

	return repo
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestLibGit2_Diff_ReturnsUnifiedDiff(t *testing.T) {
	repo := setupTestRepo(t)
	defer repo.Close()

	diff, err := repo.Diff("HEAD~1", "HEAD", DiffOptions{})
	assert.NoError(t, err)
	assert.Contains(t, diff, "@@")
	assert.Contains(t, diff, "func TestSomething")
}

func TestLibGit2_LogSinceTag_ReturnsCommits(t *testing.T) {
	repo := setupTestRepo(t)
	defer repo.Close()

	commits, err := repo.LogSince("v1.0.0")
	assert.NoError(t, err)
	assert.Len(t, commits, 3)
	assert.Equal(t, "feat: add feature", commits[0].Message)
}