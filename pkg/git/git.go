package git

// Open opens a git repository using the best available backend.
// It prefers libgit2 if available (with static linking), otherwise falls back to git CLI.
func Open(path string) (Repo, error) {
	// For now, use the fallback implementation
	// TODO: Use libgit2 when static linking is properly configured
	return OpenFallback(path)
}