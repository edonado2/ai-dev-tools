package github

import (
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
)

func TestToPR(t *testing.T) {
	now := time.Now()
	ghPR := &github.PullRequest{
		Number:         github.Int(42),
		Title:          github.String("Test PR"),
		Body:           github.String("Test body"),
		State:          github.String("open"),
		HTMLURL:        github.String("https://github.com/owner/repo/pull/42"),
		CreatedAt:      &github.Timestamp{Time: now},
		UpdatedAt:      &github.Timestamp{Time: now},
		Draft:          github.Bool(false),
		Merged:         github.Bool(false),
		Mergeable:      github.Bool(true),
		Commits:        github.Int(3),
		Additions:      github.Int(100),
		Deletions:      github.Int(50),
		ChangedFiles:   github.Int(5),
		User:           &github.User{Login: github.String("testuser")},
		Head: &github.PullRequestBranch{
			Ref: github.String("feature-branch"),
			SHA: github.String("abc123"),
			Repo: &github.Repository{
				Name:        github.String("repo"),
				FullName:    github.String("owner/repo"),
				Owner:       &github.User{Login: github.String("owner")},
				Private:     github.Bool(false),
				HTMLURL:     github.String("https://github.com/owner/repo"),
				DefaultBranch: github.String("main"),
				Description: github.String("Test repo"),
			},
		},
		Base: &github.PullRequestBranch{
			Ref: github.String("main"),
			SHA: github.String("def456"),
			Repo: &github.Repository{
				Name:        github.String("repo"),
				FullName:    github.String("owner/repo"),
				Owner:       &github.User{Login: github.String("owner")},
				Private:     github.Bool(false),
				HTMLURL:     github.String("https://github.com/owner/repo"),
				DefaultBranch: github.String("main"),
				Description: github.String("Test repo"),
			},
		},
	}

	pr := toPR(ghPR)

	assert.NotNil(t, pr)
	assert.Equal(t, 42, pr.Number)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, "Test body", pr.Body)
	assert.Equal(t, "open", pr.State)
	assert.Equal(t, "testuser", pr.Author)
	assert.Equal(t, "https://github.com/owner/repo/pull/42", pr.URL)
	assert.Equal(t, now, pr.CreatedAt)
	assert.Equal(t, now, pr.UpdatedAt)
	assert.False(t, pr.Draft)
	assert.False(t, pr.Merged)
	assert.True(t, *pr.Mergeable)
	assert.Equal(t, 3, pr.Commits)
	assert.Equal(t, 100, pr.Additions)
	assert.Equal(t, 50, pr.Deletions)
	assert.Equal(t, 5, pr.ChangedFiles)
	assert.Equal(t, "feature-branch", pr.Head.Ref)
	assert.Equal(t, "abc123", pr.Head.SHA)
	assert.NotNil(t, pr.Head.Repo)
	assert.Equal(t, "repo", pr.Head.Repo.Name)
	assert.Equal(t, "main", pr.Base.Ref)
}

func TestToPR_Nil(t *testing.T) {
	pr := toPR(nil)
	assert.Nil(t, pr)
}

func TestToCommit(t *testing.T) {
	now := time.Now()
	ghCommit := &github.RepositoryCommit{
		SHA: github.String("abc123"),
		Commit: &github.Commit{
			Message: github.String("Test commit"),
			Author: &github.CommitAuthor{
				Name:  github.String("Test Author"),
				Email: github.String("test@example.com"),
				Date:  &github.Timestamp{Time: now},
			},
		},
		HTMLURL: github.String("https://github.com/owner/repo/commit/abc123"),
	}

	commit := toCommit(ghCommit)

	assert.NotNil(t, &commit)
	assert.Equal(t, "abc123", commit.SHA)
	assert.Equal(t, "Test commit", commit.Message)
	assert.Equal(t, "Test Author", commit.Author)
	assert.Equal(t, "test@example.com", commit.Email)
	assert.Equal(t, now, commit.Date)
	assert.Equal(t, "https://github.com/owner/repo/commit/abc123", commit.URL)
}

func TestToCommit_Nil(t *testing.T) {
	commit := toCommit(nil)
	assert.Equal(t, Commit{}, commit)
}

func TestToPRFile(t *testing.T) {
	ghFile := &github.CommitFile{
		Filename:         github.String("file.go"),
		Status:           github.String("modified"),
		Additions:        github.Int(10),
		Deletions:        github.Int(5),
		Changes:          github.Int(15),
		BlobURL:          github.String("https://github.com/owner/repo/blob/main/file.go"),
		RawURL:           github.String("https://raw.githubusercontent.com/owner/repo/main/file.go"),
		ContentsURL:      github.String("https://api.github.com/repos/owner/repo/contents/file.go"),
		Patch:            github.String("@@ -1,5 +1,10 @@"),
		PreviousFilename: github.String("old_file.go"),
	}

	file := toPRFile(ghFile)

	assert.NotNil(t, &file)
	assert.Equal(t, "file.go", file.Filename)
	assert.Equal(t, "modified", file.Status)
	assert.Equal(t, 10, file.Additions)
	assert.Equal(t, 5, file.Deletions)
	assert.Equal(t, 15, file.Changes)
	assert.Equal(t, "https://github.com/owner/repo/blob/main/file.go", file.BlobURL)
	assert.Equal(t, "https://raw.githubusercontent.com/owner/repo/main/file.go", file.RawURL)
	assert.Equal(t, "https://api.github.com/repos/owner/repo/contents/file.go", file.ContentsURL)
	assert.Equal(t, "@@ -1,5 +1,10 @@", file.Patch)
	assert.Equal(t, "old_file.go", file.PreviousFilename)
}

func TestToPRFile_Nil(t *testing.T) {
	file := toPRFile(nil)
	assert.Equal(t, PRFile{}, file)
}

func TestToRepo(t *testing.T) {
	ghRepo := &github.Repository{
		Name:            github.String("test-repo"),
		FullName:        github.String("owner/test-repo"),
		Owner:           &github.User{Login: github.String("owner")},
		Private:         github.Bool(false),
		HTMLURL:         github.String("https://github.com/owner/test-repo"),
		DefaultBranch:   github.String("main"),
		Description:     github.String("Test repository"),
	}

	repo := toRepo(ghRepo)

	assert.NotNil(t, repo)
	assert.Equal(t, "test-repo", repo.Name)
	assert.Equal(t, "owner/test-repo", repo.FullName)
	assert.Equal(t, "owner", repo.Owner)
	assert.False(t, repo.Private)
	assert.Equal(t, "https://github.com/owner/test-repo", repo.HTMLURL)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.Equal(t, "Test repository", repo.Description)
}

func TestToRepo_Nil(t *testing.T) {
	repo := toRepo(nil)
	assert.Nil(t, repo)
}

func TestToPRBranch(t *testing.T) {
	ghBranch := &github.PullRequestBranch{
		Ref: github.String("feature-branch"),
		SHA: github.String("abc123"),
		Repo: &github.Repository{
			Name:            github.String("repo"),
			FullName:        github.String("owner/repo"),
			Owner:           &github.User{Login: github.String("owner")},
			Private:         github.Bool(false),
			HTMLURL:         github.String("https://github.com/owner/repo"),
			DefaultBranch:   github.String("main"),
			Description:     github.String("Test repo"),
		},
	}

	branch := toPRBranch(ghBranch)

	assert.Equal(t, "feature-branch", branch.Ref)
	assert.Equal(t, "abc123", branch.SHA)
	assert.NotNil(t, branch.Repo)
	assert.Equal(t, "repo", branch.Repo.Name)
}

func TestToPRBranch_Nil(t *testing.T) {
	branch := toPRBranch(nil)
	assert.Equal(t, PRBranch{}, branch)
}