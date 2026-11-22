package git

import (
	"fmt"
	"strings"
)

// ParseDiffStat parses the output of `git diff --stat`
// Example input: " 3 files changed, 10 insertions(+), 5 deletions(-)"
func ParseDiffStat(output string) DiffStat {
	stat := DiffStat{}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "file") && strings.Contains(line, "changed") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				fmt.Sscanf(parts[0], "%d", &stat.FilesChanged)
			}
		}
		if strings.Contains(line, "insertion") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if strings.Contains(p, "insertion") && i > 0 {
					fmt.Sscanf(parts[i-1], "%d", &stat.Insertions)
				}
			}
		}
		if strings.Contains(line, "deletion") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if strings.Contains(p, "deletion") && i > 0 {
					fmt.Sscanf(parts[i-1], "%d", &stat.Deletions)
				}
			}
		}
	}
	return stat
}