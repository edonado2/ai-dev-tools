// internal/cli/changelog/changelog.go
package changelog

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "changelog",
		Short: "Generate changelog",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
