// internal/cli/testgen/testgen.go
package testgen

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "testgen",
		Short: "Generate tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
