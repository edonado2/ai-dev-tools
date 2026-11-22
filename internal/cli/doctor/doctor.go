// internal/cli/doctor/doctor.go
package doctor

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system health",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
