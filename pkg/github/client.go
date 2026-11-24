package github

import (
	"context"

	gh "github.com/google/go-github/v60/github"
)

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
	return &ghClient{client: gh.NewClient(nil).WithAuthToken(token)}
}

type ghClient struct {
	client *gh.Client
}

func (c *ghClient) GetPR(owner, repo string, number int) (*PR, error) {
	pr, _, err := c.client.PullRequests.Get(context.Background(), owner, repo, number)
	if err != nil {
		return nil, err
	}
	return toPR(pr), nil
}

func (c *ghClient) ListPRCommits(owner, repo string, number int) ([]Commit, error) {
	commits, _, err := c.client.PullRequests.ListCommits(context.Background(), owner, repo, number, nil)
	if err != nil {
		return nil, err
	}
	result := make([]Commit, len(commits))
	for i, commit := range commits {
		result[i] = toCommit(commit)
	}
	return result, nil
}

func (c *ghClient) ListPRFiles(owner, repo string, number int) ([]PRFile, error) {
	files, _, err := c.client.PullRequests.ListFiles(context.Background(), owner, repo, number, nil)
	if err != nil {
		return nil, err
	}
	result := make([]PRFile, len(files))
	for i, file := range files {
		result[i] = toPRFile(file)
	}
	return result, nil
}

func toPRFile(f *gh.CommitFile) PRFile {
	if f == nil {
		return PRFile{}
	}
	return PRFile{
		Filename:         f.GetFilename(),
		Status:           f.GetStatus(),
		Additions:        f.GetAdditions(),
		Deletions:        f.GetDeletions(),
		Changes:          f.GetChanges(),
		BlobURL:          f.GetBlobURL(),
		RawURL:           f.GetRawURL(),
		ContentsURL:      f.GetContentsURL(),
		Patch:            f.GetPatch(),
		PreviousFilename: f.GetPreviousFilename(),
	}
}

func (c *ghClient) CreateReview(owner, repo string, number int, review ReviewRequest) error {
	ghReview := &gh.PullRequestReviewRequest{
		Body:    gh.String(review.Body),
		Event:   gh.String(review.Event),
	}
	if len(review.Comments) > 0 {
		comments := make([]*gh.DraftReviewComment, len(review.Comments))
		for i, c := range review.Comments {
			comments[i] = &gh.DraftReviewComment{
				Path: gh.String(c.Path),
				Line: gh.Int(c.Line),
				Side: gh.String(c.Side),
				Body: gh.String(c.Body),
			}
		}
		ghReview.Comments = comments
	}
	_, _, err := c.client.PullRequests.CreateReview(context.Background(), owner, repo, number, ghReview)
	return err
}

func (c *ghClient) CreateReviewComment(owner, repo string, number int, comment ReviewComment) error {
	_, _, err := c.client.PullRequests.CreateComment(context.Background(), owner, repo, number, &gh.PullRequestComment{
		Path: gh.String(comment.Path),
		Line: gh.Int(comment.Line),
		Side: gh.String(comment.Side),
		Body: gh.String(comment.Body),
	})
	return err
}

func (c *ghClient) GetRepo(owner, repo string) (*Repo, error) {
	r, _, err := c.client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return nil, err
	}
	return toRepo(r), nil
}

func (c *ghClient) GetDefaultBranch(owner, repo string) (string, error) {
	r, _, err := c.client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return "", err
	}
	return r.GetDefaultBranch(), nil
}

func toPR(pr *gh.PullRequest) *PR {
	if pr == nil {
		return nil
	}
	return &PR{
		Number:         pr.GetNumber(),
		Title:          pr.GetTitle(),
		Body:           pr.GetBody(),
		State:          pr.GetState(),
		Head:           toPRBranch(pr.Head),
		Base:           toPRBranch(pr.Base),
		Author:         pr.GetUser().GetLogin(),
		URL:            pr.GetHTMLURL(),
		CreatedAt:      pr.GetCreatedAt().Time,
		UpdatedAt:      pr.GetUpdatedAt().Time,
		Draft:          pr.GetDraft(),
		Merged:         pr.GetMerged(),
		Mergeable:      pr.Mergeable,
		Commits:        pr.GetCommits(),
		Additions:      pr.GetAdditions(),
		Deletions:      pr.GetDeletions(),
		ChangedFiles:   pr.GetChangedFiles(),
	}
}

func toPRBranch(b *gh.PullRequestBranch) PRBranch {
	if b == nil {
		return PRBranch{}
	}
	return PRBranch{
		Ref:  b.GetRef(),
		SHA:  b.GetSHA(),
		Repo: toRepo(b.Repo),
	}
}

func toCommit(c *gh.RepositoryCommit) Commit {
	if c == nil {
		return Commit{}
	}
	return Commit{
		SHA:     c.GetSHA(),
		Message: c.GetCommit().GetMessage(),
		Author:  c.GetCommit().GetAuthor().GetName(),
		Email:   c.GetCommit().GetAuthor().GetEmail(),
		Date:    c.GetCommit().GetAuthor().GetDate().Time,
		URL:     c.GetHTMLURL(),
	}
}

func toRepo(r *gh.Repository) *Repo {
	if r == nil {
		return nil
	}
	return &Repo{
		Name:            r.GetName(),
		FullName:        r.GetFullName(),
		Owner:           r.GetOwner().GetLogin(),
		Private:         r.GetPrivate(),
		HTMLURL:         r.GetHTMLURL(),
		DefaultBranch:   r.GetDefaultBranch(),
		Description:     r.GetDescription(),
	}
}