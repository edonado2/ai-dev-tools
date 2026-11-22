package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactory_CreatesProviders(t *testing.T) {
	tests := []struct {
		name  string
		model string
		cfg   Config
	}{
		{"openai", "gpt-4o", Config{APIKey: "test-key", Model: "gpt-4o"}},
		{"anthropic", "claude-3-5-sonnet-20241022", Config{APIKey: "test-key", Model: "claude-3-5-sonnet-20241022"}},
		{"ollama", "codellama", Config{Model: "codellama"}},
		{"custom", "my-model", Config{APIKey: "test-key", BaseURL: "http://localhost:8080", Model: "my-model"}},
	}
	for _, tc := range tests {
		p, err := New(tc.name, tc.cfg)
		assert.NoError(t, err)
		assert.Equal(t, tc.name, p.Name())
	}
}

func TestFactory_UnknownProvider_ReturnsError(t *testing.T) {
	_, err := New("unknown", Config{})
	assert.Error(t, err)
}