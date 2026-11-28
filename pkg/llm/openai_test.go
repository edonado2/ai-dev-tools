package llm

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenAI_Complete_ReturnsStructuredOutput(t *testing.T) {
	if testing.Short() || os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("requires OPENAI_API_KEY; set RUN_INTEGRATION=1 to run")
	}
	ctx := context.Background()
	p, err := NewOpenAI(Config{APIKey: os.Getenv("OPENAI_API_KEY"), Model: "gpt-4o-mini"})
	assert.NoError(t, err)
	resp, err := p.Complete(ctx, CompletionRequest{
		Messages: []Message{{Role: "user", Content: "Return JSON: {\"name\":\"test\"}"}},
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
				"required": []string{"name"},
			},
		},
	})
	assert.NoError(t, err)
	var out map[string]string
	json.Unmarshal([]byte(resp.Content), &out)
	assert.Equal(t, "test", out["name"])
}