package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewConfigCommand creates a config command for testing with a custom config directory
func NewConfigCommand(configDir string) *cobra.Command {
	cfg, _ := config.Load(config.WithConfigDir(configDir))
	return NewCommand(&cfg, configDir)
}

func TestConfigInit_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewConfigCommand(tmp)
	cmd.SetArgs([]string{"init"})

	// Simulate user input: provider=anthropic, model=claude-3-5-sonnet, api_key_env= (empty)
	input := "anthropic\nclaude-3-5-sonnet\n\n"
	cmd.SetIn(strings.NewReader(input))

	err := cmd.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmp, ".ai-dev-tools.yaml"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "anthropic")
	assert.Contains(t, string(content), "claude-3-5-sonnet")
}

func TestConfigShow_PrintsConfig(t *testing.T) {
	tmp := t.TempDir()

	// Create a config file first
	configContent := `provider:
  name: anthropic
  model: claude-3-5-sonnet
  api_key_env: ANTHROPIC_API_KEY
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewConfigCommand(tmp)
	cmd.SetArgs([]string{"show"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "anthropic")
	assert.Contains(t, output, "claude-3-5-sonnet")
	assert.Contains(t, output, "ANTHROPIC_API_KEY")
}

func TestConfigPath_PrintsConfigPath(t *testing.T) {
	tmp := t.TempDir()

	// Create a config file first
	configContent := `provider:
  name: anthropic
  model: claude-3-5-sonnet
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewConfigCommand(tmp)
	cmd.SetArgs([]string{"path"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, ".ai-dev-tools.yaml")
	assert.Contains(t, output, tmp)
}

func TestConfigInit_WithExistingFile_PromptsOverwrite(t *testing.T) {
	tmp := t.TempDir()

	// Create an existing config file
	configContent := `provider:
  name: openai
  model: gpt-4
`
	err := os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewConfigCommand(tmp)
	cmd.SetArgs([]string{"init"})

	// Simulate user input: overwrite=y, provider=anthropic, model=claude-3-5-sonnet, api_key_env= (empty)
	input := "y\nanthropic\nclaude-3-5-sonnet\n\n"
	cmd.SetIn(strings.NewReader(input))

	err = cmd.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmp, ".ai-dev-tools.yaml"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "anthropic")
	assert.Contains(t, string(content), "claude-3-5-sonnet")
}

func TestConfigShow_NoConfigFile_ShowsDefaults(t *testing.T) {
	tmp := t.TempDir()

	cmd := NewConfigCommand(tmp)
	cmd.SetArgs([]string{"show"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	// Should show defaults
	assert.Contains(t, output, "anthropic")
}

// Mock reader for testing interactive input
type mockReader struct {
	inputs []string
	index  int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.index >= len(m.inputs) {
		return 0, io.EOF
	}
	input := m.inputs[m.index] + "\n"
	m.index++
	return copy(p, input), nil
}

func TestConfigInit_UsesDefaultsWhenEmpty(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewConfigCommand(tmp)
	cmd.SetArgs([]string{"init"})

	// Provide empty inputs to use defaults
	cmd.SetIn(&mockReader{inputs: []string{"", "", ""}})

	err := cmd.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmp, ".ai-dev-tools.yaml"))
	require.NoError(t, err)

	// Should have defaults
	assert.Contains(t, string(content), "anthropic")
}

func TestConfigCommand_Subcommands(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewConfigCommand(tmp)

	// Test init subcommand exists
	initCmd, _, err := cmd.Find([]string{"init"})
	require.NoError(t, err)
	assert.Equal(t, "init", initCmd.Name())

	// Test show subcommand exists
	showCmd, _, err := cmd.Find([]string{"show"})
	require.NoError(t, err)
	assert.Equal(t, "show", showCmd.Name())

	// Test path subcommand exists
	pathCmd, _, err := cmd.Find([]string{"path"})
	require.NoError(t, err)
	assert.Equal(t, "path", pathCmd.Name())
}