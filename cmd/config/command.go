package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewCommand creates a new config command with subcommands
func NewCommand(cfg *config.Config, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `Manage ai-dev-tools configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(
		newInitCommand(cfg, configDir),
		newShowCommand(cfg, configDir),
		newPathCommand(configDir),
	)

	return cmd
}

// newInitCommand creates the config init subcommand
func newInitCommand(cfg *config.Config, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		Long:  `Interactively create a new .ai-dev-tools.yaml configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, cfg, configDir)
		},
	}
	return cmd
}

// newShowCommand creates the config show subcommand
func newShowCommand(cfg *config.Config, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current configuration as YAML.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(cmd, cfg, configDir)
		},
	}
	return cmd
}

// newPathCommand creates the config path subcommand
func newPathCommand(configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  `Display the path to the configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPath(cmd, configDir)
		},
	}
	return cmd
}

func runInit(cmd *cobra.Command, cfg *config.Config, configDir string) error {
	configPath := filepath.Join(configDir, ".ai-dev-tools.yaml")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprint(cmd.OutOrStdout(), "Config file already exists. Overwrite? [y/N]: ")
		reader := bufio.NewReader(cmd.InOrStdin())
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
	}

	// Get provider
	fmt.Fprint(cmd.OutOrStdout(), "Provider (anthropic/openai/ollama/custom) [anthropic]: ")
	provider := readInput(cmd.InOrStdin())
	if provider == "" {
		provider = "anthropic"
	}

	// Validate provider
	validProviders := map[string]bool{"anthropic": true, "openai": true, "ollama": true, "custom": true}
	if !validProviders[provider] {
		return fmt.Errorf("invalid provider: %s", provider)
	}

	// Get model
	defaultModel := getDefaultModel(provider)
	fmt.Fprintf(cmd.OutOrStdout(), "Model [%s]: ", defaultModel)
	model := readInput(cmd.InOrStdin())
	if model == "" {
		model = defaultModel
	}

	// Get API key env var
	fmt.Fprint(cmd.OutOrStdout(), "API Key Environment Variable (optional): ")
	apiKeyEnv := readInput(cmd.InOrStdin())

	// Get base URL for custom/ollama
	var baseURL string
	if provider == "custom" || provider == "ollama" {
		defaultBaseURL := getDefaultBaseURL(provider)
		fmt.Fprintf(cmd.OutOrStdout(), "Base URL [%s]: ", defaultBaseURL)
		baseURL = readInput(cmd.InOrStdin())
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
	}

	// Build config
	newCfg := config.Config{
		Provider: config.ProviderConfig{
			Name:      provider,
			Model:     model,
			APIKeyEnv: apiKeyEnv,
			BaseURL:   baseURL,
			Timeout:   60,
			MaxRetries: 3,
		},
	}

	// Write config file
	return writeConfigFile(configPath, newCfg, cmd.OutOrStdout())
}

func runShow(cmd *cobra.Command, cfg *config.Config, configDir string) error {
	// Use the loaded config (which includes file, env, and defaults)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	fmt.Fprint(cmd.OutOrStdout(), string(data))
	return nil
}

func runPath(cmd *cobra.Command, configDir string) error {
	configPath := filepath.Join(configDir, ".ai-dev-tools.yaml")
	fmt.Fprintln(cmd.OutOrStdout(), configPath)
	return nil
}

func readInput(in io.Reader) string {
	reader := bufio.NewReader(in)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func getDefaultModel(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-3-5-sonnet"
	case "openai":
		return "gpt-4o"
	case "ollama":
		return "llama3"
	case "custom":
		return ""
	default:
		return "claude-3-5-sonnet"
	}
}

func getDefaultBaseURL(provider string) string {
	switch provider {
	case "ollama":
		return "http://localhost:11434"
	case "custom":
		return ""
	default:
		return ""
	}
}

func writeConfigFile(path string, cfg config.Config, out io.Writer) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Fprintf(out, "Configuration written to %s\n", path)
	return nil
}