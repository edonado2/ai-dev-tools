// cmd/changelog/command.go
package changelog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/ernedo/ai-dev-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

// ChangelogOutput represents the LLM response for changelog
type ChangelogOutput struct {
	Version string            `json:"version"`
	Date    string            `json:"date"`
	Sections map[string][]string `json:"sections"`
}

// NewCommand creates a new changelog command
func NewCommand(repo git.Repo, ghClient github.Client, llmProvider llm.Provider, cfg *config.Config) *cobra.Command {
	var since, until, format, output string
	var includePRs bool
	var owner, repoName string
	var prNum int

	// Get default config values from config (no hardcoded defaults)
	var changelogCfg config.ChangelogConfig
	if cfg != nil {
		changelogCfg = cfg.Tools.Changelog
	}

	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Generate changelog from commit history",
		Long: `Generate a changelog from Git commit history using AI.
Supports Keep a Changelog format with sections for Added, Changed, Deprecated, Removed, Fixed, Security.
Can enrich with GitHub PR data when --include-prs flag is used.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Validate format
			if format != "keepachangelog" {
				return fmt.Errorf("unsupported format: %s (only 'keepachangelog' supported)", format)
			}

			// Determine since tag
			if since == "" {
				tag, err := repo.LatestTag()
				if err != nil {
					return fmt.Errorf("get latest tag: %w", err)
				}
				if tag == "" {
					return errors.New("no tags found; specify --since")
				}
				since = tag
			}

			// Get previous version (tag before since)
			previousVersion := getPreviousTag(repo, since)

			// Get commits since tag
			commits, err := repo.Log(git.LogOptions{
				Since: since,
				Until: until,
			})
			if err != nil {
				return fmt.Errorf("get commit log: %w", err)
			}

			if len(commits) == 0 {
				return errors.New("no commits found since " + since)
			}

			// If includePRs is enabled, enrich commits with PR data
			var commitStrs []string
			if includePRs {
				if ghClient == nil {
					return errors.New("GitHub client not configured; cannot use --include-prs")
				}
				if owner == "" || repoName == "" {
					return errors.New("--owner and --repo are required when using --include-prs")
				}
				commitStrs, err = enrichCommitsWithPRs(ctx, ghClient, owner, repoName, commits, prNum)
				if err != nil {
					return fmt.Errorf("enrich commits with PRs: %w", err)
				}
			} else {
				commitStrs = make([]string, len(commits))
				for i, c := range commits {
					commitStrs[i] = fmt.Sprintf("%s %s", c.ShortHash, c.Message)
				}
			}

			// Determine version
			version := "Unreleased"
			if since != "" {
				// Try to parse version from tag
				version = strings.TrimPrefix(since, "v")
			}

			// Render prompt template
			reg := prompt.NewRegistry()
			tpl, err := reg.Get("changelog", "1")
			if err != nil {
				return fmt.Errorf("get template: %w", err)
			}

			req, err := tpl.Render(map[string]any{
				"Commits":          commitStrs,
				"Version":          version,
				"Date":             time.Now().Format("2006-01-02"),
				"PreviousVersion":  previousVersion,
			})
			if err != nil {
				return fmt.Errorf("render template: %w", err)
			}

			// Call LLM
			resp, err := llmProvider.Complete(ctx, *req)
			if err != nil {
				return fmt.Errorf("llm complete: %w", err)
			}

			// Parse LLM response
			var llmOut ChangelogOutput
			if err := json.Unmarshal([]byte(resp.Content), &llmOut); err != nil {
				// Try to extract JSON from markdown code blocks
				content := extractJSON(resp.Content)
				if err := json.Unmarshal([]byte(content), &llmOut); err != nil {
					return fmt.Errorf("unmarshal changelog output: %w", err)
				}
			}

			// Generate changelog markdown
			changelogMD := formatKeepAChangelog(llmOut, since)

			// Write output
			var out io.Writer
			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return fmt.Errorf("create output file: %w", err)
				}
				defer f.Close()
				out = f
			} else {
				out = cmd.OutOrStdout()
			}

			_, err = fmt.Fprint(out, changelogMD)
			return err
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "Start tag/commit (default: latest tag)")
	cmd.Flags().StringVar(&until, "until", "", "End tag/commit (default: HEAD)")
	cmd.Flags().StringVar(&format, "format", changelogCfg.Format, "Output format: keepachangelog")
	cmd.Flags().StringVar(&output, "output", changelogCfg.Output, "Output file (default: stdout)")
	cmd.Flags().BoolVar(&includePRs, "include-prs", changelogCfg.IncludePRs, "Enrich with GitHub PR data")
	cmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (required with --include-prs)")
	cmd.Flags().StringVar(&repoName, "repo", "", "GitHub repository name (required with --include-prs)")
	cmd.Flags().IntVar(&prNum, "pr", 0, "Specific PR number to include")

	// Use config values as defaults if flags not explicitly set
	if changelogCfg.Format != "" && !cmd.Flags().Changed("format") {
		cmd.Flags().Set("format", changelogCfg.Format)
	}
	if changelogCfg.Output != "" && !cmd.Flags().Changed("output") {
		cmd.Flags().Set("output", changelogCfg.Output)
	}
	if cmd.Flags().Changed("include-prs") {
		// Flag was explicitly set, don't override
	} else if changelogCfg.IncludePRs {
		cmd.Flags().Set("include-prs", fmt.Sprintf("%v", changelogCfg.IncludePRs))
	}

	return cmd
}

// getPreviousTag returns the tag before the given tag
func getPreviousTag(repo git.Repo, since string) string {
	// Use git to get all tags sorted by date, then find the one before 'since'
	rootDir, err := repo.RootDir()
	if err != nil {
		return ""
	}

	cmd := exec.Command("git", "tag", "--sort=-creatordate")
	cmd.Dir = rootDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return ""
	}

	tags := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	found := false
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if found {
			return tag
		}
		if tag == since {
			found = true
		}
	}
	return ""
}

// enrichCommitsWithPRs fetches PR data for commits and returns enriched commit strings
func enrichCommitsWithPRs(ctx context.Context, ghClient github.Client, owner, repoName string, commits []git.Commit, specificPR int) ([]string, error) {
	// Build a map of commit SHA -> git.Commit for quick lookup
	commitMap := make(map[string]git.Commit)
	for _, c := range commits {
		commitMap[c.Hash] = c
	}

	// If specific PR number provided, fetch that PR and enrich matching commits
	if specificPR > 0 {
		pr, err := ghClient.GetPR(owner, repoName, specificPR)
		if err != nil {
			return nil, fmt.Errorf("get PR #%d: %w", specificPR, err)
		}
		prCommits, err := ghClient.ListPRCommits(owner, repoName, specificPR)
		if err != nil {
			return nil, fmt.Errorf("list PR commits: %w", err)
		}
		prFiles, err := ghClient.ListPRFiles(owner, repoName, specificPR)
		if err != nil {
			return nil, fmt.Errorf("list PR files: %w", err)
		}
		return formatEnrichedCommits(pr, prCommits, prFiles, commitMap), nil
	}

	// Without a specific PR, we cannot match commits to PRs
	// (would require ListPRs or SearchPRs which aren't in the interface)
	// Return original commit strings
	var commitStrs []string
	for _, c := range commits {
		commitStrs = append(commitStrs, fmt.Sprintf("%s %s", c.ShortHash, c.Message))
	}
	return commitStrs, nil
}

func formatEnrichedCommits(pr *github.PR, prCommits []github.Commit, prFiles []github.PRFile, commitMap map[string]git.Commit) []string {
	var result []string
	result = append(result, fmt.Sprintf("PR #%d: %s", pr.Number, pr.Title))
	if pr.Body != "" {
		result = append(result, fmt.Sprintf("  Description: %s", pr.Body))
	}
	if len(prFiles) > 0 {
		result = append(result, "  Files changed:")
		for _, f := range prFiles {
			result = append(result, fmt.Sprintf("    %s (%s)", f.Filename, f.Status))
		}
	}
	result = append(result, "  Commits:")
	for _, pc := range prCommits {
		shortSHA := pc.SHA
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}
		// Check if this commit is in our commit map
		if gc, ok := commitMap[pc.SHA]; ok {
			result = append(result, fmt.Sprintf("    %s %s (matched: %s)", shortSHA, pc.Message, gc.ShortHash))
		} else {
			result = append(result, fmt.Sprintf("    %s %s", shortSHA, pc.Message))
		}
	}
	return result
}

func formatKeepAChangelog(out ChangelogOutput, sinceTag string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Changelog\n\n")

	// Unreleased section
	sb.WriteString("## [Unreleased]\n\n")

	// If we have a version from LLM, use it
	version := out.Version
	if version == "" {
		version = "1.0.0"
	}
	date := out.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Add version section
	sb.WriteString(fmt.Sprintf("## [%s] - %s\n\n", version, date))

	// Sections in Keep a Changelog order
	sectionOrder := []string{"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"}
	for _, section := range sectionOrder {
		items, ok := out.Sections[section]
		if !ok || len(items) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("### %s\n\n", section))
		for _, item := range items {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func extractJSON(content string) string {
	// Try to extract JSON from markdown code blocks
	start := strings.Index(content, "```json")
	if start == -1 {
		start = strings.Index(content, "```")
	}
	if start != -1 {
		start = strings.Index(content[start:], "\n")
		if start != -1 {
			start += start
			end := strings.Index(content[start:], "```")
			if end != -1 {
				return strings.TrimSpace(content[start : start+end])
			}
		}
	}
	return content
}
