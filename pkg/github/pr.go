package github

import "time"

type PR struct {
	Number     int
	Title      string
	Body       string
	State      string
	Head       PRBranch
	Base       PRBranch
	Author     string
	URL        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Draft      bool
	Merged     bool
	Mergeable  *bool
	Commits    int
	Additions  int
	Deletions  int
	ChangedFiles int
}

type PRBranch struct {
	Ref  string
	SHA  string
	Repo *Repo
}

type Commit struct {
	SHA     string
	Message string
	Author  string
	Email   string
	Date    time.Time
	URL     string
}

type PRFile struct {
	Filename        string
	Status          string
	Additions       int
	Deletions       int
	Changes         int
	BlobURL         string
	RawURL          string
	ContentsURL     string
	Patch           string
	PreviousFilename string
}