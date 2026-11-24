package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoader_PriorityOrder(t *testing.T) {
	tmp := t.TempDir()
	// 1. Default
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "anthropic", cfg.Provider.Name)
	// 2. Config file
	os.WriteFile(filepath.Join(tmp, ".ai-dev-tools.yaml"), []byte(`provider:
  name: openai
`), 0644)
	cfg, err = Load(WithConfigDir(tmp))
	assert.NoError(t, err)
	assert.Equal(t, "openai", cfg.Provider.Name)
	// 3. Env var overrides file
	t.Setenv("AI_DEV_PROVIDER_NAME", "ollama")
	cfg, err = Load(WithConfigDir(tmp))
	assert.NoError(t, err)
	assert.Equal(t, "ollama", cfg.Provider.Name)
	// 4. CLI flag would override env (tested in CLI tasks)
}

func TestLoader_Validation_RejectsInvalid(t *testing.T) {
	_, err := Load(WithConfigYAML(`provider:
  name: invalid
`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")
}