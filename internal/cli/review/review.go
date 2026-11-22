// internal/cli/review/review.go
package review

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "review",
		Short: "Review code changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
