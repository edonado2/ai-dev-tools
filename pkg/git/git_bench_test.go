// pkg/git/git_bench_test.go
package git

import (
	"testing"
)

func BenchmarkParseDiffStat(b *testing.B) {
	diff := ` 3 files changed, 10 insertions(+), 5 deletions(-)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseDiffStat(diff)
	}
}

func BenchmarkDiffOptions_Create(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DiffOptions{
			IncludePatterns: []string{"*.go"},
			ExcludePatterns: []string{"*_test.go"},
		}
	}
}