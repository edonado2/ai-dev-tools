package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetPR_ReturnsPR(t *testing.T) {
	if testing.Short() {
		t.Skip("requires GitHub token")
	}
	c := New(os.Getenv("GITHUB_TOKEN"))
	pr, err := c.GetPR("ernedo", "test-repo", 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, pr.Number)
}

func TestClient_CreateReviewComment_PostsComment(t *testing.T) {
	if testing.Short() {
		t.Skip("requires GitHub token")
	}
	c := New(os.Getenv("GITHUB_TOKEN"))
	err := c.CreateReviewComment("ernedo", "test-repo", 1, ReviewComment{
		Path: "file.go", Line: 10, Side: "RIGHT", Body: "Test comment",
	})
	assert.NoError(t, err)
}