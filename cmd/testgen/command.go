// cmd/testgen/command.go
package testgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/spf13/cobra"
)

// NewCommand creates a new testgen command
func NewCommand(repo git.Repo, llmProvider llm.Provider, cfg *config.Config) *cobra.Command {
	// Get default config values
	testGenCfg := TestGenConfig{
		Framework: "",
		Output:    "file",
		Coverage:  false,
	}

	cmd := &cobra.Command{
		Use:   "testgen [files...]",
		Short: "Generate unit tests for source files",
		Long: `Generate unit tests using AI for Go, TypeScript, or Python files.
Supports auto-detection of test framework (gotest, jest, pytest).
Output can be written to files or stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no files specified")
			}

			// Resolve file paths relative to repo root
			rootDir, err := repo.RootDir()
			if err != nil {
				return fmt.Errorf("get repo root: %w", err)
			}

			var files []string
			for _, arg := range args {
				path := arg
				if !filepath.IsAbs(path) {
					path = filepath.Join(rootDir, path)
				}
				// Verify file exists
				if _, err := os.Stat(path); err != nil {
					return fmt.Errorf("file not found: %s", arg)
				}
				files = append(files, path)
			}

			// Auto-detect framework if not specified
			if testGenCfg.Framework == "" {
				testGenCfg.Framework = DetectFramework(files)
				if testGenCfg.Framework == "unknown" {
					return errors.New("could not detect test framework; specify with --framework")
				}
			}

			gen := NewGenerator(repo, llmProvider, &testGenCfg)

			ctx := cmd.Context()
			if err := gen.Generate(ctx, files); err != nil {
				return fmt.Errorf("generate tests: %w", err)
			}

			// Print summary to stderr if output is file
			if testGenCfg.Output == "file" && cmd.ErrOrStderr() != nil {
				out := cmd.ErrOrStderr()
				fmt.Fprintf(out, "Generated tests for %d file(s) using %s\n", len(files), testGenCfg.Framework)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&testGenCfg.Framework, "framework", "", "Test framework: gotest, jest, vitest, pytest (auto-detected if not specified)")
	cmd.Flags().StringVar(&testGenCfg.Output, "output", "file", "Output mode: file, stdout")
	cmd.Flags().BoolVar(&testGenCfg.Coverage, "coverage", false, "Generate tests with coverage focus")

	// Use config if provided
	if cfg != nil && cfg.Tools.TestGen.Framework != "" {
		cmd.Flags().Set("framework", cfg.Tools.TestGen.Framework)
	}
	if cfg != nil && cfg.Tools.TestGen.Output != "" {
		cmd.Flags().Set("output", cfg.Tools.TestGen.Output)
	}
	if cfg != nil {
		cmd.Flags().Set("coverage", fmt.Sprintf("%v", cfg.Tools.TestGen.Coverage))
	}

	return cmd
}