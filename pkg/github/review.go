package github

type ReviewComment struct {
	Path string
	Line int
	Side string
	Body string
}

type ReviewRequest struct {
	Body    string
	Event   string
	Comments []ReviewComment
}

type Repo struct {
	Name        string
	FullName    string
	Owner       string
	Private     bool
	HTMLURL     string
	DefaultBranch string
	Description string
}