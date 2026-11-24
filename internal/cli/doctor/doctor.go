// internal/cli/doctor/doctor.go
package doctor

import (
	cmddoctor "github.com/ernedo/ai-dev-tools/cmd/doctor"
	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/spf13/cobra"
)

// NewCmd creates the doctor command with all dependencies wired up
func NewCmd() *cobra.Command {
	var repo git.Repo
	var ghClient github.Client
	var llmProvider llm.Provider
	var cfg *config.Config
	var configDir string

	// This will be initialized by the root command's PersistentPreRunE
	// For now, return the command structure
	return cmddoctor.NewCommand(repo, ghClient, llmProvider, cfg, configDir)
}