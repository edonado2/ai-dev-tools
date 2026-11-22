package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// OpenFallback opens a git repository using the git CLI
func OpenFallback(path string) (Repo, error) {
	// Verify git exists
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git not found in PATH: %w", err)
	}

	// Verify it's a git repo
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	return &fallbackRepo{dir: path}, nil
}

type fallbackRepo struct {
	dir string
}

func (r *fallbackRepo) runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

func (r *fallbackRepo) Diff(base, head string, opts DiffOptions) (string, error) {
	args := []string{"diff", base + ".." + head}
	if len(opts.IncludePatterns) > 0 {
		args = append(args, "--")
		args = append(args, opts.IncludePatterns...)
	}
	if len(opts.ExcludePatterns) > 0 {
		for _, p := range opts.ExcludePatterns {
			args = append(args, "--", ":"+p)
		}
	}
	return r.runGit(args...)
}

func (r *fallbackRepo) DiffCached(opts DiffOptions) (string, error) {
	args := []string{"diff", "--cached"}
	if len(opts.IncludePatterns) > 0 {
		args = append(args, "--")
		args = append(args, opts.IncludePatterns...)
	}
	if len(opts.ExcludePatterns) > 0 {
		for _, p := range opts.ExcludePatterns {
			args = append(args, "--", ":"+p)
		}
	}
	return r.runGit(args...)
}

func (r *fallbackRepo) DiffStat(base, head string) (DiffStat, error) {
	args := []string{"diff", "--stat", base + ".." + head}
	output, err := r.runGit(args...)
	if err != nil {
		return DiffStat{}, err
	}
	return ParseDiffStat(output), nil
}

func (r *fallbackRepo) Log(opts LogOptions) ([]Commit, error) {
	args := []string{"log", "--pretty=format:%H|%h|%s|%an|%ae|%ad", "--date=iso"}
	if opts.Since != "" {
		args = append(args, opts.Since+"..HEAD")
	}
	if opts.Until != "" {
		args = append(args, "--until="+opts.Until)
	}
	if opts.MaxCount > 0 {
		args = append(args, fmt.Sprintf("-%d", opts.MaxCount))
	}

	output, err := r.runGit(args...)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 6 {
			continue
		}

		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[5])

		// Get files for this commit
		files, _ := r.getCommitFiles(parts[0])

		commits = append(commits, Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Message:   parts[2],
			Author:    parts[3],
			Email:     parts[4],
			Date:      date,
			Files:     files,
		})
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return commits, nil
}

func (r *fallbackRepo) getCommitFiles(hash string) ([]string, error) {
	output, err := r.runGit("show", "--name-only", "--pretty=format:", hash)
	if err != nil {
		return nil, err
	}
	var files []string
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func (r *fallbackRepo) LogSince(tag string) ([]Commit, error) {
	return r.Log(LogOptions{Since: tag})
}

func (r *fallbackRepo) Status() (Status, error) {
	output, err := r.runGit("status", "--porcelain")
	if err != nil {
		return Status{}, err
	}

	var staged, unstaged, untracked []string
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		path := strings.TrimSpace(line[3:])

		if status[0] != ' ' && status[0] != '?' {
			staged = append(staged, path)
		}
		if status[1] != ' ' {
			unstaged = append(unstaged, path)
		}
		if status == "??" {
			untracked = append(untracked, path)
		}
	}

	return Status{
		Staged:    staged,
		Unstaged:  unstaged,
		Untracked: untracked,
	}, nil
}

func (r *fallbackRepo) GetFile(path, ref string) (string, error) {
	return r.runGit("show", ref+":"+path)
}

func (r *fallbackRepo) ListFiles(ref, pattern string) ([]string, error) {
	args := []string{"ls-tree", "-r", "--name-only", ref}
	if pattern != "" {
		args = append(args, pattern)
	}
	output, err := r.runGit(args...)
	if err != nil {
		return nil, err
	}

	var files []string
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func (r *fallbackRepo) RootDir() (string, error) {
	output, err := r.runGit("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *fallbackRepo) LatestTag() (string, error) {
	output, err := r.runGit("describe", "--tags", "--abbrev=0")
	if err != nil {
		// No tags, try to get the latest tag by date
		output, err = r.runGit("tag", "--sort=-creatordate")
		if err != nil {
			return "", fmt.Errorf("no tags found")
		}
		tags := strings.Split(strings.TrimSpace(output), "\n")
		if len(tags) > 0 && tags[0] != "" {
			return tags[0], nil
		}
		return "", fmt.Errorf("no tags found")
	}
	return strings.TrimSpace(output), nil
}

func (r *fallbackRepo) CurrentBranch() (string, error) {
	output, err := r.runGit("branch", "--show-current")
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(output)
	if branch == "" {
		// Detached HEAD
		output, err = r.runGit("rev-parse", "--short", "HEAD")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(output), nil
	}
	return branch, nil
}

func (r *fallbackRepo) IsDirty() (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}
	return len(status.Staged) > 0 || len(status.Unstaged) > 0 || len(status.Untracked) > 0, nil
}

func (r *fallbackRepo) Close() error {
	return nil
}

// Ensure fallbackRepo implements Repo interface
var _ Repo = (*fallbackRepo)(nil)