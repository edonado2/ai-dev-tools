// internal/cli/commit/commit.go
package commit

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "Generate commit messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
