// internal/cli/config/config.go
package config

import (
	cmdconfig "github.com/ernedo/ai-dev-tools/cmd/config"
	pkgconfig "github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/spf13/cobra"
)

// NewCmd creates the config command with all dependencies wired up
func NewCmd() *cobra.Command {
	var cfg *pkgconfig.Config
	var configDir string

	// This will be initialized by the root command's PersistentPreRunE
	// For now, return the command structure with placeholders
	cmd := cmdconfig.NewCommand(cfg, configDir)

	return cmd
}