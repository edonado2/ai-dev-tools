package doctor

import (
	"context"
	"fmt"
	"strings"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/spf13/cobra"
)

// NewCommand creates a new doctor command
func NewCommand(repo git.Repo, ghClient github.Client, llmProvider llm.Provider, cfg *config.Config, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system health",
		Long:  `Run health checks for configuration, git repository, LLM provider connectivity, and GitHub authentication.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd, repo, ghClient, llmProvider, cfg, configDir)
		},
	}
	return cmd
}

func runDoctor(cmd *cobra.Command, repo git.Repo, ghClient github.Client, llmProvider llm.Provider, cfg *config.Config, configDir string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	// 1. Check Config
	fmt.Fprint(out, "Checking configuration... ")
	if err := checkConfig(cfg); err != nil {
		fmt.Fprintf(out, "✗ Config invalid: %v\n", err)
	} else {
		fmt.Fprintln(out, "✓ Config valid")
	}

	// 2. Check Git Repository
	fmt.Fprint(out, "Checking git repository... ")
	if err := checkGitRepo(repo); err != nil {
		fmt.Fprintf(out, "✗ Git repository: %v\n", err)
	} else {
		fmt.Fprintln(out, "✓ Git repository")
	}

	// 3. Check LLM Provider
	fmt.Fprint(out, "Checking LLM provider... ")
	if err := checkProvider(ctx, llmProvider, cfg); err != nil {
		fmt.Fprintf(out, "✗ %s API unreachable: %v\n", providerDisplayName(cfg.Provider.Name), err)
	} else {
		fmt.Fprintf(out, "✓ %s API reachable\n", providerDisplayName(cfg.Provider.Name))
	}

	// 4. Check GitHub Authentication (only if configured)
	if cfg.GitHub.TokenEnv != "" || cfg.GitHub.Owner != "" || cfg.GitHub.Repo != "" {
		fmt.Fprint(out, "Checking GitHub authentication... ")
		if err := checkGitHub(ctx, ghClient, cfg); err != nil {
			fmt.Fprintf(out, "✗ GitHub authentication failed: %v\n", err)
		} else {
			fmt.Fprintln(out, "✓ GitHub authentication")
		}
	}

	return nil
}

func checkConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.Provider.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	if cfg.Provider.Model == "" {
		return fmt.Errorf("provider model is required")
	}
	// Validate provider name
	validProviders := map[string]bool{"anthropic": true, "openai": true, "ollama": true, "custom": true}
	if !validProviders[cfg.Provider.Name] {
		return fmt.Errorf("invalid provider: %s", cfg.Provider.Name)
	}
	return nil
}

func checkGitRepo(repo git.Repo) error {
	if repo == nil {
		return fmt.Errorf("git repo is nil")
	}
	// Try to get the root directory to verify we're in a git repo
	_, err := repo.RootDir()
	if err != nil {
		return err
	}
	// Also try to get current branch
	_, err = repo.CurrentBranch()
	if err != nil {
		return err
	}
	return nil
}

func checkProvider(ctx context.Context, provider llm.Provider, cfg *config.Config) error {
	if provider == nil {
		return fmt.Errorf("provider is nil")
	}
	// Try a simple completion to test connectivity
	req := llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "ping"},
		},
		MaxTokens: 10,
	}
	_, err := provider.Complete(ctx, req)
	return err
}

func checkGitHub(ctx context.Context, ghClient github.Client, cfg *config.Config) error {
	if ghClient == nil {
		return fmt.Errorf("GitHub client not initialized")
	}
	if cfg.GitHub.Owner == "" || cfg.GitHub.Repo == "" {
		return fmt.Errorf("GitHub owner and repo not configured")
	}
	// Try to get the repo to test authentication
	_, err := ghClient.GetRepo(cfg.GitHub.Owner, cfg.GitHub.Repo)
	return err
}

func providerDisplayName(name string) string {
	switch name {
	case "anthropic":
		return "Anthropic"
	case "openai":
		return "OpenAI"
	case "ollama":
		return "Ollama"
	case "custom":
		return "Custom"
	default:
		return strings.Title(name)
	}
}