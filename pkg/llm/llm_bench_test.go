// pkg/llm/llm_bench_test.go
package llm

import (
	"testing"
)

func BenchmarkFactory_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = New("anthropic", Config{Model: "claude-3-5-sonnet", APIKey: "test-key"})
	}
}

func BenchmarkConfig_Struct(b *testing.B) {
	cfg := Config{Model: "test", APIKey: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg
	}
}

func BenchmarkCompletionRequest_Create(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CompletionRequest{
			Messages: []Message{{Role: "user", Content: "Hello world"}},
			Temperature: 0.7,
			MaxTokens: 1000,
		}
	}
}