// cmd/ai-dev/main.go
package main

import "github.com/ernedo/ai-dev-tools/internal/cli"

// Version information set by GoReleaser at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.Execute(version, commit, date)
}