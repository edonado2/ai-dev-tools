// cmd/testgen/generator.go
package testgen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/ernedo/ai-dev-tools/pkg/prompt"
)

// TestGenConfig holds test generation configuration
type TestGenConfig struct {
	Framework string
	Output    string // "file" or "stdout"
	Coverage  bool
}

// TestGenOutput represents the LLM response for test generation
type TestGenOutput struct {
	Tests []TestCase `json:"tests"`
}

type TestCase struct {
	Name  string        `json:"name"`
	Cases []TestSubCase `json:"cases"`
}

type TestSubCase struct {
	Name     string         `json:"name"`
	Input    map[string]any `json:"input,omitempty"`
	Expected map[string]any `json:"expected,omitempty"`
	Setup    string         `json:"setup,omitempty"`
	Expect   string         `json:"expect,omitempty"`
}

// DetectFramework auto-detects the test framework based on file extensions and config files
func DetectFramework(files []string) string {
	hasGo := false
	hasTS := false
	hasPython := false

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		switch ext {
		case ".go":
			hasGo = true
		case ".ts", ".tsx":
			hasTS = true
		case ".py":
			hasPython = true
		}
	}

	// Check for config files that indicate framework
	repoRoot := ""
	if len(files) > 0 {
		repoRoot = filepath.Dir(files[0])
		for {
			parent := filepath.Dir(repoRoot)
			if parent == repoRoot {
				break
			}
			repoRoot = parent
		}
	}

	if hasGo {
		return "gotest"
	}
	if hasTS {
		// Check for jest/vitest config
		if fileExists(filepath.Join(repoRoot, "jest.config.js")) ||
			fileExists(filepath.Join(repoRoot, "jest.config.ts")) ||
			fileExists(filepath.Join(repoRoot, "vitest.config.ts")) {
			return "jest"
		}
		return "jest" // default for TS
	}
	if hasPython {
		return "pytest"
	}

	return "unknown"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Generator handles test generation
type Generator struct {
	repo       git.Repo
	llmProvider llm.Provider
	config     *TestGenConfig
}

// NewGenerator creates a new test generator
func NewGenerator(repo git.Repo, llmProvider llm.Provider, cfg *TestGenConfig) *Generator {
	return &Generator{
		repo:        repo,
		llmProvider: llmProvider,
		config:      cfg,
	}
}

// Generate generates tests for the given files
func (g *Generator) Generate(ctx context.Context, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified")
	}

	framework := g.config.Framework
	if framework == "" {
		framework = DetectFramework(files)
	}

	for _, file := range files {
		if err := g.generateForFile(ctx, file, framework); err != nil {
			return fmt.Errorf("generate for %s: %w", file, err)
		}
	}

	return nil
}

func (g *Generator) generateForFile(ctx context.Context, file, framework string) error {
	// Read source file
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	// Determine package name for Go
	pkg := ""
	if framework == "gotest" {
		pkg = extractGoPackage(string(source))
	}

	// Get focus from config
	focus := "unit tests"
	if g.config.Coverage {
		focus = "unit tests with coverage"
	}

	// Render prompt template
	reg := prompt.NewRegistry()
	tpl, err := reg.Get("testgen", "1")
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	req, err := tpl.Render(map[string]any{
		"Source":  string(source),
		"Package": pkg,
		"File":    file,
		"Focus":   focus,
	})
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	// Call LLM
	resp, err := g.llmProvider.Complete(ctx, *req)
	if err != nil {
		return fmt.Errorf("llm complete: %w", err)
	}

	// Parse response
	var output TestGenOutput
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	// Generate test file content
	testContent := g.generateTestFile(framework, file, pkg, output)

	// Write output
	outputPath := g.getOutputPath(file, framework)
	if g.config.Output == "stdout" {
		fmt.Print(testContent)
	} else {
		if err := os.WriteFile(outputPath, []byte(testContent), 0644); err != nil {
			return fmt.Errorf("write test file: %w", err)
		}
	}

	return nil
}

func extractGoPackage(source string) string {
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "package "))
		}
	}
	return "main"
}

func (g *Generator) getOutputPath(sourceFile, framework string) string {
	ext := filepath.Ext(sourceFile)
	base := strings.TrimSuffix(sourceFile, ext)

	switch framework {
	case "gotest":
		return base + "_test.go"
	case "jest", "vitest":
		if ext == ".tsx" {
			return base + ".test.tsx"
		}
		return base + ".test.ts"
	case "pytest":
		return base + "_test.py"
	default:
		return base + "_test" + ext
	}
}

func (g *Generator) generateTestFile(framework, sourceFile, pkg string, output TestGenOutput) string {
	var sb strings.Builder

	switch framework {
	case "gotest":
		sb.WriteString(fmt.Sprintf("package %s\n\n", pkg))
		sb.WriteString(fmt.Sprintf("// Generated tests for %s\n\n", sourceFile))
		for _, tc := range output.Tests {
			sb.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", tc.Name))
			sb.WriteString("\ttestCases := []struct {\n")
			sb.WriteString("\t\tname     string\n")
			sb.WriteString("\t\tinput    map[string]any\n")
			sb.WriteString("\t\texpected map[string]any\n")
			sb.WriteString("\t}{\n")
			for _, c := range tc.Cases {
				sb.WriteString(fmt.Sprintf("\t\t{\"%s\", %v, %v},\n", c.Name, c.Input, c.Expected))
			}
			sb.WriteString("\t}\n\n")
			sb.WriteString("\tfor _, tc := range testCases {\n")
			sb.WriteString("\t\tt.Run(tc.name, func(t *testing.T) {\n")
			sb.WriteString("\t\t\t// TODO: implement test case\n")
			sb.WriteString("\t\t\t_ = tc.input\n")
			sb.WriteString("\t\t\t_ = tc.expected\n")
			sb.WriteString("\t\t})\n")
			sb.WriteString("\t}\n")
			sb.WriteString("}\n\n")
		}

	case "jest", "vitest":
		sb.WriteString("// Generated tests\n")
		sb.WriteString("import { describe, it, expect, jest } from '@jest/globals';\n\n")
		for _, tc := range output.Tests {
			sb.WriteString(fmt.Sprintf("describe('%s', () => {\n", tc.Name))
			for _, c := range tc.Cases {
				sb.WriteString(fmt.Sprintf("\tit('%s', () => {\n", c.Name))
				if c.Setup != "" {
					sb.WriteString(fmt.Sprintf("\t\t%s\n", c.Setup))
				}
				if c.Expect != "" {
					sb.WriteString(fmt.Sprintf("\t\texpect(%s);\n", c.Expect))
				}
				sb.WriteString("\t});\n")
			}
			sb.WriteString("});\n\n")
		}

	case "pytest":
		sb.WriteString("# Generated tests\n")
		sb.WriteString("import pytest\n\n")
		for _, tc := range output.Tests {
			sb.WriteString(fmt.Sprintf("class Test%s:\n", tc.Name))
			for _, c := range tc.Cases {
				sb.WriteString(fmt.Sprintf("    def test_%s(self):\n", c.Name))
				if c.Setup != "" {
					sb.WriteString(fmt.Sprintf("        %s\n", c.Setup))
				}
				if c.Expect != "" {
					sb.WriteString(fmt.Sprintf("        assert %s\n", c.Expect))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}