//go:build libgit2
// +build libgit2

package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/libgit2/git2go"
)

func Open(path string) (Repo, error) {
	r, err := git.OpenRepository(path)
	if err != nil {
		return nil, err
	}
	return &libgit2Repo{repo: r}, nil
}

type libgit2Repo struct {
	repo *git.Repository
}

func (r *libgit2Repo) Diff(base, head string, opts DiffOptions) (string, error) {
	// Parse the revision specs
	baseCommit, err := r.repo.RevparseSingle(base)
	if err != nil {
		return "", fmt.Errorf("invalid base revision %q: %w", base, err)
	}
	defer baseCommit.Free()

	headCommit, err := r.repo.RevparseSingle(head)
	if err != nil {
		return "", fmt.Errorf("invalid head revision %q: %w", head, err)
	}
	defer headCommit.Free()

	baseTree, err := baseCommit.Peel(git.ObjectTree).AsTree()
	if err != nil {
		return "", fmt.Errorf("failed to get base tree: %w", err)
	}
	defer baseTree.Free()

	headTree, err := headCommit.Peel(git.ObjectTree).AsTree()
	if err != nil {
		return "", fmt.Errorf("failed to get head tree: %w", err)
	}
	defer headTree.Free()

	// Create diff options
	diffOpts, err := git.DefaultDiffOptions()
	if err != nil {
		return "", fmt.Errorf("failed to create diff options: %w", err)
	}

	// Add pathspecs for include patterns
	if len(opts.IncludePatterns) > 0 {
		paths := make([]string, len(opts.IncludePatterns))
		for i, p := range opts.IncludePatterns {
			paths[i] = p
		}
		diffOpts.Pathspec = paths
	}

	// Generate diff
	diff, err := r.repo.DiffTreeToTree(baseTree, headTree, &diffOpts)
	if err != nil {
		return "", fmt.Errorf("failed to generate diff: %w", err)
	}
	defer diff.Free()

	// Convert diff to string
	var diffOutput strings.Builder
	err = diff.ForEach(func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
		return func(hunk git.DiffHunk) error {
			diffOutput.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
				hunk.OldStart, hunk.OldLines, hunk.NewStart, hunk.NewLines))
			return nil
		}, nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate diff: %w", err)
	}

	// Also get the patch format for the actual content
	patch, err := diff.Patch()
	if err != nil {
		return "", fmt.Errorf("failed to get patch: %w", err)
	}
	defer patch.Free()

	patchStr, err := patch.String()
	if err != nil {
		return "", fmt.Errorf("failed to convert patch to string: %w", err)
	}

	// If we got nothing from ForEach, use the patch directly
	if diffOutput.Len() == 0 {
		return patchStr, nil
	}

	return patchStr, nil
}

func (r *libgit2Repo) DiffCached(opts DiffOptions) (string, error) {
	// Get the index
	index, err := r.repo.Index()
	if err != nil {
		return "", fmt.Errorf("failed to get index: %w", err)
	}
	defer index.Free()

	// Get head tree
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	defer head.Free()

	headCommit, err := r.repo.LookupCommit(head.Target())
	if err != nil {
		return "", fmt.Errorf("failed to lookup HEAD commit: %w", err)
	}
	defer headCommit.Free()

	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD tree: %w", err)
	}
	defer headTree.Free()

	diffOpts, err := git.DefaultDiffOptions()
	if err != nil {
		return "", fmt.Errorf("failed to create diff options: %w", err)
	}

	if len(opts.IncludePatterns) > 0 {
		paths := make([]string, len(opts.IncludePatterns))
		for i, p := range opts.IncludePatterns {
			paths[i] = p
		}
		diffOpts.Pathspec = paths
	}

	diff, err := r.repo.DiffTreeToIndex(headTree, index, &diffOpts)
	if err != nil {
		return "", fmt.Errorf("failed to generate diff: %w", err)
	}
	defer diff.Free()

	patch, err := diff.Patch()
	if err != nil {
		return "", fmt.Errorf("failed to get patch: %w", err)
	}
	defer patch.Free()

	return patch.String()
}

func (r *libgit2Repo) DiffStat(base, head string) (DiffStat, error) {
	baseCommit, err := r.repo.RevparseSingle(base)
	if err != nil {
		return DiffStat{}, fmt.Errorf("invalid base revision %q: %w", base, err)
	}
	defer baseCommit.Free()

	headCommit, err := r.repo.RevparseSingle(head)
	if err != nil {
		return DiffStat{}, fmt.Errorf("invalid head revision %q: %w", head, err)
	}
	defer headCommit.Free()

	baseTree, err := baseCommit.Peel(git.ObjectTree).AsTree()
	if err != nil {
		return DiffStat{}, fmt.Errorf("failed to get base tree: %w", err)
	}
	defer baseTree.Free()

	headTree, err := headCommit.Peel(git.ObjectTree).AsTree()
	if err != nil {
		return DiffStat{}, fmt.Errorf("failed to get head tree: %w", err)
	}
	defer headTree.Free()

	diffOpts, err := git.DefaultDiffOptions()
	if err != nil {
		return DiffStat{}, fmt.Errorf("failed to create diff options: %w", err)
	}

	diff, err := r.repo.DiffTreeToTree(baseTree, headTree, &diffOpts)
	if err != nil {
		return DiffStat{}, fmt.Errorf("failed to generate diff: %w", err)
	}
	defer diff.Free()

	stats, err := diff.Stats()
	if err != nil {
		return DiffStat{}, fmt.Errorf("failed to get diff stats: %w", err)
	}

	return DiffStat{
		FilesChanged: stats.FilesChanged(),
		Insertions:   stats.Insertions(),
		Deletions:    stats.Deletions(),
	}, nil
}

func (r *libgit2Repo) Log(opts LogOptions) ([]Commit, error) {
	walker, err := r.repo.Walk()
	if err != nil {
		return nil, fmt.Errorf("failed to create walker: %w", err)
	}
	defer walker.Free()

	// Push HEAD to start walking from
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}
	defer head.Free()

	walker.Push(head.Target())

	// If Since is specified, we need to hide commits before that tag
	if opts.Since != "" {
		sinceCommit, err := r.repo.RevparseSingle(opts.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid since revision %q: %w", opts.Since, err)
		}
		defer sinceCommit.Free()
		walker.Hide(sinceCommit.Id())
	}

	if opts.Until != "" {
		untilCommit, err := r.repo.RevparseSingle(opts.Until)
		if err != nil {
			return nil, fmt.Errorf("invalid until revision %q: %w", opts.Until, err)
		}
		defer untilCommit.Free()
		// We'll just stop when we reach this commit
		_ = untilCommit.Id()
	}

	var commits []Commit
	count := 0
	err = walker.Iterate(func(commit *git.Commit) bool {
		if opts.MaxCount > 0 && count >= opts.MaxCount {
			return false
		}

		// Get commit files
		tree, err := commit.Tree()
		if err != nil {
			return true
		}
		defer tree.Free()

		var files []string
		tree.Walk(func(path string, entry *git.TreeEntry) int {
			files = append(files, path)
			return 0
		})

		commits = append(commits, Commit{
			Hash:      commit.Id().String(),
			ShortHash: commit.Id().String()[:7],
			Message:   strings.TrimSpace(commit.Message()),
			Author:    commit.Author().Name,
			Email:     commit.Author().Email,
			Date:      commit.Author().When,
			Files:     files,
		})
		count++
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

func (r *libgit2Repo) LogSince(tag string) ([]Commit, error) {
	return r.Log(LogOptions{Since: tag})
}

func (r *libgit2Repo) Status() (Status, error) {
	statusOpts, err := git.DefaultStatusOptions()
	if err != nil {
		return Status{}, fmt.Errorf("failed to create status options: %w", err)
	}

	statusList, err := r.repo.StatusList(&statusOpts)
	if err != nil {
		return Status{}, fmt.Errorf("failed to get status list: %w", err)
	}
	defer statusList.Free()

	count, err := statusList.EntryCount()
	if err != nil {
		return Status{}, fmt.Errorf("failed to get status count: %w", err)
	}

	var staged, unstaged, untracked []string
	for i := 0; i < count; i++ {
		entry, err := statusList.ByIndex(i)
		if err != nil {
			continue
		}

		status := entry.Status
		path := entry.IndexToWorkdir
		if path == "" {
			path = entry.HeadToIndex
		}

		if status&git.StatusIndexNew != 0 || status&git.StatusIndexModified != 0 || status&git.StatusIndexDeleted != 0 {
			staged = append(staged, path)
		}
		if status&git.StatusWtModified != 0 || status&git.StatusWtDeleted != 0 || status&git.StatusWtNew != 0 {
			unstaged = append(unstaged, path)
		}
		if status&git.StatusWtNew != 0 && status&git.StatusIndexNew == 0 {
			untracked = append(untracked, path)
		}
	}

	return Status{
		Staged:     staged,
		Unstaged:   unstaged,
		Untracked:  untracked,
	}, nil
}

func (r *libgit2Repo) GetFile(path, ref string) (string, error) {
	commit, err := r.repo.RevparseSingle(ref)
	if err != nil {
		return "", fmt.Errorf("invalid revision %q: %w", ref, err)
	}
	defer commit.Free()

	tree, err := commit.Peel(git.ObjectTree).AsTree()
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}
	defer tree.Free()

	entry, err := tree.EntryByPath(path)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	blob, err := r.repo.LookupBlob(entry.Id)
	if err != nil {
		return "", fmt.Errorf("failed to lookup blob: %w", err)
	}
	defer blob.Free()

	return string(blob.Contents()), nil
}

func (r *libgit2Repo) ListFiles(ref, pattern string) ([]string, error) {
	commit, err := r.repo.RevparseSingle(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid revision %q: %w", ref, err)
	}
	defer commit.Free()

	tree, err := commit.Peel(git.ObjectTree).AsTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}
	defer tree.Free()

	var files []string
	tree.Walk(func(path string, entry *git.TreeEntry) int {
		if pattern == "" || matchPattern(path, pattern) {
			files = append(files, path)
		}
		return 0
	})

	return files, nil
}

func matchPattern(path, pattern string) bool {
	// Simple glob matching
	if pattern == "*" {
		return true
	}
	// For now, just check if pattern is contained in path
	return strings.Contains(path, pattern)
}

func (r *libgit2Repo) RootDir() (string, error) {
	path := r.repo.Path()
	// Remove .git suffix if present
	return strings.TrimSuffix(path, "/.git"), nil
}

func (r *libgit2Repo) LatestTag() (string, error) {
	tags, err := r.repo.TagNames("")
	if err != nil {
		return "", fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
	}

	// Sort tags by commit date (newest first)
	var latestTag string
	var latestTime time.Time

	for _, tagName := range tags {
		tagRef, err := r.repo.References.Lookup("refs/tags/" + tagName)
		if err != nil {
			continue
		}

		obj, err := r.repo.Lookup(tagRef.Target())
		if err != nil {
			continue
		}

		var commit *git.Commit
		if obj.Type() == git.ObjectTag {
			tagObj := obj.AsTag()
			target, err := tagObj.Target()
			if err != nil {
				continue
			}
			commit, err = target.AsCommit()
			if err != nil {
				continue
			}
		} else if obj.Type() == git.ObjectCommit {
			commit = obj.AsCommit()
		} else {
			continue
		}

		if commit.Author().When.After(latestTime) {
			latestTime = commit.Author().When
			latestTag = tagName
		}
	}

	if latestTag == "" {
		return "", fmt.Errorf("no valid tags found")
	}

	return latestTag, nil
}

func (r *libgit2Repo) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	defer head.Free()

	if head.IsBranch() {
		return head.Shorthand(), nil
	}

	// Detached HEAD - return the commit hash
	return head.Target().String()[:7], nil
}

func (r *libgit2Repo) IsDirty() (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}
	return len(status.Staged) > 0 || len(status.Unstaged) > 0 || len(status.Untracked) > 0, nil
}

func (r *libgit2Repo) Close() error {
	r.repo.Free()
	return nil
}

// Ensure libgit2Repo implements Repo interface
var _ Repo = (*libgit2Repo)(nil)