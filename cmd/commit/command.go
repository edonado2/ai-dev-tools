// cmd/commit/command.go
package commit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/ernedo/ai-dev-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

// Config represents the configuration for the commit command
type Config struct {
	Tools ToolsConfig `mapstructure:"tools"`
}

type ToolsConfig struct {
	Commit CommitConfig `mapstructure:"commit"`
}

type CommitConfig struct {
	Style string `mapstructure:"style"`
}

// CommitOutput represents the LLM response
type CommitOutput struct {
	Message  string `json:"message"`
	Type     string `json:"type"`
	Scope    string `json:"scope"`
	Breaking bool   `json:"breaking"`
}

// NewCommand creates a new commit command
func NewCommand(repo git.Repo, llmProvider llm.Provider, cfg *Config) *cobra.Command {
	var dryRun, signOff bool
	var style string

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate AI-powered commit message",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic: %v", r)
				}
			}()
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: RunE started\n")
			diff, err := repo.DiffCached(git.DiffOptions{})
			if err != nil {
				return err
			}
			if strings.TrimSpace(diff) == "" {
				return errors.New("no staged changes")
			}

			ctx := cmd.Context()
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: dryRun=%v, style=%v\n", dryRun, style)
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: creating registry\n")
			reg := prompt.NewRegistry()
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: registry created\n")
			tpl, err := reg.Get("commit", "1")
			if err != nil {
				return fmt.Errorf("get template: %w", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: got template\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: about to call getStagedFiles\n")

			files := getStagedFiles(repo)
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: got files: %v\n", files)
			contextStr := getContext(repo)
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: got context: %s\n", contextStr)

			req, err := tpl.Render(map[string]any{
				"Diff":    diff,
				"Context": contextStr,
				"Files":   files,
			})
			if err != nil {
				return fmt.Errorf("render template: %w", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: rendered template\n")

			resp, err := llmProvider.Complete(ctx, *req)
			if err != nil {
				return fmt.Errorf("llm complete: %w", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: got LLM response: %s\n", resp.Content)

			var out CommitOutput
			if err := json.Unmarshal([]byte(resp.Content), &out); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: unmarshaled: %+v\n", out)

			msg := formatMessage(out, style, signOff)
			fmt.Fprintf(cmd.ErrOrStderr(), "DEBUG: formatted msg: %s\n", msg)

			if dryRun {
				cmd.Println(msg)
				return nil
			}

			return commitWithMessage(repo, msg)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print message without committing")
	cmd.Flags().StringVar(&style, "style", "conventional", "conventional|gitmoji")
	cmd.Flags().BoolVar(&signOff, "sign-off", false, "Add Signed-off-by")

	// Use config style as default if provided
	if cfg != nil && cfg.Tools.Commit.Style != "" {
		cmd.Flags().Set("style", cfg.Tools.Commit.Style)
	}

	return cmd
}

func getStagedFiles(repo git.Repo) []string {
	status, err := repo.Status()
	if err != nil {
		return nil
	}
	return status.Staged
}

func getContext(repo git.Repo) string {
	branch, _ := repo.CurrentBranch()
	tag, _ := repo.LatestTag()
	if tag != "" {
		return fmt.Sprintf("branch: %s, latest tag: %s", branch, tag)
	}
	return fmt.Sprintf("branch: %s", branch)
}

func formatMessage(out CommitOutput, style string, signOff bool) string {
	var msg string
	switch style {
	case "gitmoji":
		msg = formatGitmoji(out)
	default:
		msg = out.Message
	}

	if signOff {
		msg += "\n\nSigned-off-by: " // In real use, this would include the user's name/email
	}

	return msg
}

func formatGitmoji(out CommitOutput) string {
	emojiMap := map[string]string{
		"feat":     "✨",
		"fix":      "🐛",
		"docs":     "📝",
		"style":    "💄",
		"refactor": "♻️",
		"perf":     "⚡",
		"test":     "✅",
		"chore":    "🔧",
		"revert":   "⏪",
	}

	emoji := emojiMap[out.Type]
	if emoji == "" {
		emoji = "📦"
	}

	scope := ""
	if out.Scope != "" {
		scope = "(" + out.Scope + ")"
	}

	return fmt.Sprintf("%s %s%s: %s", emoji, out.Type, scope, out.Message)
}

func commitWithMessage(repo git.Repo, msg string) error {
	// Use git CLI to commit
	// This is a fallback implementation - in production we'd use the repo interface
	rootDir, err := repo.RootDir()
	if err != nil {
		return err
	}

	// We need to execute git commit with the message
	// For now, we'll use a simple approach
	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = rootDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w: %s", err, stderr.String())
	}
	return nil
}