package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate_Render_ValidatesInput(t *testing.T) {
	tpl := CommitTemplateV1
	req, err := tpl.Render(map[string]any{"Diff": "diff", "Context": "ctx", "Files": []string{"a.go"}})
	assert.NoError(t, err)
	assert.Contains(t, req.Messages[1].Content, "diff")
}

func TestTemplate_Render_MissingRequired_Fails(t *testing.T) {
	tpl := CommitTemplateV1
	_, err := tpl.Render(map[string]any{"Diff": "diff"}) // missing Context, Files
	assert.Error(t, err)
}