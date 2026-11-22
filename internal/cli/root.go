// internal/cli/root.go
package cli

import (
	"github.com/ernedo/ai-dev-tools/internal/cli/changelog"
	"github.com/ernedo/ai-dev-tools/internal/cli/commit"
	"github.com/ernedo/ai-dev-tools/internal/cli/config"
	"github.com/ernedo/ai-dev-tools/internal/cli/doctor"
	"github.com/ernedo/ai-dev-tools/internal/cli/review"
	"github.com/ernedo/ai-dev-tools/internal/cli/testgen"
	"github.com/ernedo/ai-dev-tools/internal/cli/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "ai-dev",
	Short:             "AI-powered developer tools",
	PersistentPreRunE: initConfig,
}

func Execute() {
	rootCmd.Execute()
}

func NewRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.AddCommand(
		commit.NewCmd(),
		review.NewCmd(),
		testgen.NewCmd(),
		changelog.NewCmd(),
		config.NewCmd(),
		version.NewCmd(),
		doctor.NewCmd(),
	)
}

func initConfig(cmd *cobra.Command, args []string) error {
	return nil
}
