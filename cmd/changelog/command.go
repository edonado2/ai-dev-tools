// cmd/changelog/command.go
package changelog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

	// Get default config values
	changelogCfg := config.ChangelogConfig{
		Format:     "keepachangelog",
		IncludePRs: false,
		Output:     "",
	}
	if cfg != nil && cfg.Tools.Changelog.Format != "" {
		changelogCfg.Format = cfg.Tools.Changelog.Format
		changelogCfg.IncludePRs = cfg.Tools.Changelog.IncludePRs
		changelogCfg.Output = cfg.Tools.Changelog.Output
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

			// Format commits for prompt
			commitStrs := make([]string, len(commits))
			for i, c := range commits {
				commitStrs[i] = fmt.Sprintf("%s %s", c.ShortHash, c.Message)
			}

			// Determine version
			version := "Unreleased"
			if since != "" {
				// Try to parse version from tag
				version = strings.TrimPrefix(since, "v")
			}

			// Get previous version if possible
			previousVersion := ""
			if since != "" {
				// Get tag before since
				// For now, we'll leave this empty
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