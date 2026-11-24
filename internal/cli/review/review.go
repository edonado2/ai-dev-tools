// internal/cli/review/review.go
package review

import (
	"github.com/ernedo/ai-dev-tools/cmd/review"
	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/spf13/cobra"
)

// NewCmd creates the review command with all dependencies wired up
func NewCmd() *cobra.Command {
	var repo git.Repo
	var ghClient github.Client
	var llmProvider llm.Provider
	var cfg *config.Config

	// This will be initialized by the root command's PersistentPreRunE
	// For now, return the command structure
	return review.NewCommand(repo, ghClient, llmProvider, cfg)
}
