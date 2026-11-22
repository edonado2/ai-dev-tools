package git

import (
	"time"
)

type Repo interface {
	Diff(base, head string, opts DiffOptions) (string, error)
	DiffCached(opts DiffOptions) (string, error)
	DiffStat(base, head string) (DiffStat, error)
	Log(opts LogOptions) ([]Commit, error)
	LogSince(tag string) ([]Commit, error)
	Status() (Status, error)
	GetFile(path, ref string) (string, error)
	ListFiles(ref, pattern string) ([]string, error)
	RootDir() (string, error)
	LatestTag() (string, error)
	CurrentBranch() (string, error)
	IsDirty() (bool, error)
	Close() error
}

type DiffOptions struct {
	IncludePatterns []string
	ExcludePatterns []string
	MaxSize         int64
}

type DiffStat struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

type LogOptions struct {
	Since    string
	Until    string
	MaxCount int
}

type Commit struct {
	Hash      string
	ShortHash string
	Message   string
	Author    string
	Email     string
	Date      time.Time
	Files     []string
}

type Status struct {
	// Add fields as needed
	Staged  []string
	Unstaged []string
	Untracked []string
}