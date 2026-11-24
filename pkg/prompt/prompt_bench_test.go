// pkg/prompt/prompt_bench_test.go
package prompt

import (
	"testing"
)

func BenchmarkRegistry_Get(b *testing.B) {
	reg := NewRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.Get("commit", "1")
	}
}

func BenchmarkTemplate_Render(b *testing.B) {
	tpl := CommitTemplateV1
	vars := map[string]any{
		"Diff":    "diff content",
		"Context": "repo context",
		"Files":   []string{"a.go", "b.go"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tpl.Render(vars)
	}
}

func BenchmarkRegistry_List(b *testing.B) {
	reg := NewRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reg.List()
	}
}