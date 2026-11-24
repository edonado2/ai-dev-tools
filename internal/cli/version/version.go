// internal/cli/version/version.go
package version

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewCmd(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("ai-dev-tools %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  date:   %s\n", date)
			return nil
		},
	}
}