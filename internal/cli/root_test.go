// internal/cli/root_test.go
package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand_HasSubcommands(t *testing.T) {
	cmd := NewRootCmd()
	names := []string{}
	for _, c := range cmd.Commands() {
		names = append(names, c.Name())
	}
	assert.Contains(t, names, "commit")
	assert.Contains(t, names, "review")
	assert.Contains(t, names, "testgen")
	assert.Contains(t, names, "changelog")
	assert.Contains(t, names, "config")
	assert.Contains(t, names, "version")
	assert.Contains(t, names, "doctor")
}
