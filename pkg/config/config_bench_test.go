// pkg/config/config_bench_test.go
package config

import (
	"testing"
)

func BenchmarkLoad_Defaults(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Load()
	}
}

func BenchmarkLoad_WithConfigDir(b *testing.B) {
	tmpDir := b.TempDir()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load(WithConfigDir(tmpDir))
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := Config{
		Provider: ProviderConfig{
			Name: "anthropic", Model: "claude-3-5-sonnet", APIKeyEnv: "KEY",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validate(cfg)
	}
}

func BenchmarkDecodeHook(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DecodeHook()
	}
}