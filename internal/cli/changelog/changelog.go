// internal/cli/changelog/changelog.go
package changelog

import (
	"github.com/ernedo/ai-dev-tools/cmd/changelog"
	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/spf13/cobra"
)

// NewCmd creates the changelog command with all dependencies wired up
func NewCmd() *cobra.Command {
	var repo git.Repo
	var ghClient github.Client
	var llmProvider llm.Provider
	var cfg *config.Config

	// This will be initialized by the root command's PersistentPreRunE
	// For now, return the command structure
	return changelog.NewCommand(repo, ghClient, llmProvider, cfg)
}
