// internal/cli/testgen/testgen.go
package testgen

import (
	"github.com/ernedo/ai-dev-tools/cmd/testgen"
	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/spf13/cobra"
)

// NewCmd creates the testgen command with all dependencies wired up
func NewCmd() *cobra.Command {
	var repo git.Repo
	var llmProvider llm.Provider
	var cfg *config.Config

	// This will be initialized by the root command's PersistentPreRunE
	// For now, return the command structure
	return testgen.NewCommand(repo, llmProvider, cfg)
}