package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_BuiltInTemplates(t *testing.T) {
	reg := NewRegistry()
	names := reg.List()
	assert.Contains(t, names, "commit@v1")
	assert.Contains(t, names, "review@v1")
	assert.Contains(t, names, "testgen@v1")
	assert.Contains(t, names, "changelog@v1")
}

func TestRegistry_UserOverride_LoadsCustom(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "commit@v1.yaml"), []byte(`
name: commit
version: "1"
system_prompt: "Custom: {{.Diff}}"
user_prompt: "Custom prompt"
`), 0644)
	reg := NewRegistry(WithUserDir(tmp))
	tpl, err := reg.Get("commit", "1")
	assert.NoError(t, err)
	assert.Contains(t, tpl.SystemPrompt, "Custom")
}