package review

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernedo/ai-dev-tools/pkg/config"
	"github.com/ernedo/ai-dev-tools/pkg/git"
	"github.com/ernedo/ai-dev-tools/pkg/github"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
	"github.com/ernedo/ai-dev-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

// Finding represents a single code review finding
type Finding struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Pass       string `json:"pass"`
}

// ReviewOutput represents the complete review output
type ReviewOutput struct {
	Summary  string    `json:"summary"`
	Findings []Finding `json:"findings"`
	Approved bool      `json:"approved"`
}

// SARIFOutput represents SARIF format output
type SARIFOutput struct {
	Version string    `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	InformationURI  string `json:"informationUri"`
	Rules           []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortDescription SARIFDescription `json:"shortDescription"`
	FullDescription   SARIFDescription `json:"fullDescription"`
	DefaultConfiguration SARIFConfig `json:"defaultConfiguration"`
}

type SARIFDescription struct {
	Text string `json:"text"`
}

type SARIFConfig struct {
	Level string `json:"level"`
}

type SARIFResult struct {
	RuleID    string        `json:"ruleId"`
	Level     string        `json:"level"`
	Message   SARIFMessage  `json:"message"`
	Locations []SARIFLocation `json:"locations"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

type SARIFRegion struct {
	StartLine int `json:"startLine"`
}

// ReviewConfig holds review-specific configuration
type ReviewConfig struct {
	SeverityThreshold string
	ExcludePatterns   []string
	IncludePatterns   []string
}

// NewCommand creates a new review command
func NewCommand(repo git.Repo, ghClient github.Client, llmProvider llm.Provider, cfg *config.Config) *cobra.Command {
	var prNum int
	var base, head, output, severity string
	var owner, repoName string

	// Get default config values
	reviewCfg := ReviewConfig{
		SeverityThreshold: "warning",
		ExcludePatterns:   []string{},
		IncludePatterns:   []string{},
	}
	if cfg != nil && cfg.Tools.Review.SeverityThreshold != "" {
		reviewCfg.SeverityThreshold = cfg.Tools.Review.SeverityThreshold
		reviewCfg.ExcludePatterns = cfg.Tools.Review.ExcludePatterns
		reviewCfg.IncludePatterns = cfg.Tools.Review.IncludePatterns
	}

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review code changes with multi-pass analysis",
		Long: `Review code changes using AI-powered multi-pass analysis.
Supports reviewing local diffs or GitHub PRs.
Output formats: text, json, sarif, github (posts review comments).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var diff string
			var files []string
			var contextStr string

			if prNum > 0 {
				// Fetch PR diff via GitHub API
				if owner == "" || repoName == "" {
					return errors.New("--owner and --repo required when using --pr")
				}
				prFiles, err := ghClient.ListPRFiles(owner, repoName, prNum)
				if err != nil {
					return fmt.Errorf("fetch PR files: %w", err)
				}
				diff = buildDiffFromFiles(prFiles)
				files = getPRFilenames(prFiles)
				contextStr = fmt.Sprintf("PR #%d in %s/%s", prNum, owner, repoName)
			} else {
				// Local diff
				if base == "" {
					base = "main"
				}
				if head == "" {
					head = "HEAD"
				}
				diffOpts := git.DiffOptions{
					IncludePatterns: reviewCfg.IncludePatterns,
					ExcludePatterns: reviewCfg.ExcludePatterns,
				}
				var err error
				diff, err = repo.Diff(base, head, diffOpts)
				if err != nil {
					return fmt.Errorf("get diff: %w", err)
				}
				files = getChangedFiles(repo, base, head)
				branch, _ := repo.CurrentBranch()
				tag, _ := repo.LatestTag()
				if tag != "" {
					contextStr = fmt.Sprintf("branch: %s, latest tag: %s, diff: %s..%s", branch, tag, base, head)
				} else {
					contextStr = fmt.Sprintf("branch: %s, diff: %s..%s", branch, base, head)
				}
			}

			if strings.TrimSpace(diff) == "" {
				return errors.New("no changes to review")
			}

			// Run multi-pass review
			findings := runPasses(cmd.Context(), llmProvider, diff, files, contextStr, reviewCfg)

			// Output results
			out := cmd.OutOrStdout()
			switch output {
			case "github":
				if prNum == 0 {
					return errors.New("--output github requires --pr")
				}
				return postGitHubReview(cmd.Context(), ghClient, owner, repoName, prNum, findings)
			case "sarif":
				return writeSARIF(out, findings)
			case "json":
				return writeJSON(out, findings)
			case "text":
				return printFindings(out, findings)
			default:
				return fmt.Errorf("unknown output format: %s", output)
			}
		},
	}

	cmd.Flags().IntVar(&prNum, "pr", 0, "GitHub PR number to review")
	cmd.Flags().StringVar(&base, "base", "main", "Base branch for local diff")
	cmd.Flags().StringVar(&head, "head", "HEAD", "Head branch for local diff")
	cmd.Flags().StringVar(&output, "output", "text", "Output format: text, json, sarif, github")
	cmd.Flags().StringVar(&severity, "severity", reviewCfg.SeverityThreshold, "Minimum severity: error, warning, info")
	cmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (required with --pr)")
	cmd.Flags().StringVar(&repoName, "repo", "", "GitHub repository name (required with --pr)")

	// Use config severity as default if provided
	if reviewCfg.SeverityThreshold != "" {
		cmd.Flags().Set("severity", reviewCfg.SeverityThreshold)
	}

	return cmd
}

func runPasses(ctx context.Context, llmProvider llm.Provider, diff string, files []string, contextStr string, cfg ReviewConfig) []Finding {
	var allFindings []Finding

	for _, pass := range Passes {
		// Skip if pass severity is below threshold
		if severityBelow(cfg.SeverityThreshold, pass.Severity) {
			continue
		}

		// Build request for this pass
		req := buildReviewRequest(pass, diff, files, contextStr)

		// Call LLM
		resp, err := llmProvider.Complete(ctx, req)
		if err != nil {
			// Log error but continue with other passes
			fmt.Fprintf(os.Stderr, "Warning: pass %s failed: %v\n", pass.Name, err)
			continue
		}

		// Parse findings
		passFindings := parseFindings(resp.Content, pass.Name)
		allFindings = append(allFindings, passFindings...)
	}

	// Filter by patterns
	return filterByPatterns(allFindings, cfg.ExcludePatterns, cfg.IncludePatterns)
}

func buildReviewRequest(pass ReviewPass, diff string, files []string, contextStr string) llm.CompletionRequest {
	reg := prompt.NewRegistry()
	tpl, _ := reg.Get("review", "1")

	filesStr := strings.Join(files, ", ")
	focus := pass.Name

	req, _ := tpl.Render(map[string]any{
		"Diff":    diff,
		"Context": contextStr,
		"Files":   filesStr,
		"Focus":   focus,
	})

	// Override system prompt with pass-specific prompt
	if len(req.Messages) > 0 {
		req.Messages[0].Content = pass.SystemPrompt
	}

	return *req
}

func parseFindings(content, passName string) []Finding {
	var output ReviewOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		// Try to extract JSON from markdown code blocks
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &output); err != nil {
			return nil
		}
	}

	var findings []Finding
	for _, f := range output.Findings {
		f.Pass = passName
		findings = append(findings, f)
	}
	return findings
}

func extractJSON(content string) string {
	// Try to extract JSON from markdown code blocks
	start := strings.Index(content, "```json")
	if start == -1 {
		start = strings.Index(content, "```")
	}
	if start != -1 {
		start = strings.Index(content[start:], "\n")
		if start != -1 {
			start += start
			end := strings.Index(content[start:], "```")
			if end != -1 {
				return strings.TrimSpace(content[start : start+end])
			}
		}
	}
	return content
}

func severityBelow(threshold, severity string) bool {
	levels := map[string]int{"error": 3, "warning": 2, "info": 1}
	return levels[severity] < levels[threshold]
}

func filterByPatterns(findings []Finding, exclude, include []string) []Finding {
	if len(include) == 0 && len(exclude) == 0 {
		return findings
	}

	var filtered []Finding
	for _, f := range findings {
		matched := false

		// Check include patterns
		if len(include) > 0 {
			for _, pattern := range include {
				if matchPattern(f.File, pattern) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Check exclude patterns
		if len(exclude) > 0 {
			matched = false
			for _, pattern := range exclude {
				if matchPattern(f.File, pattern) {
					matched = true
					break
				}
			}
			if matched {
				continue
			}
		}

		filtered = append(filtered, f)
	}
	return filtered
}

func matchPattern(file, pattern string) bool {
	matched, _ := filepath.Match(pattern, file)
	if matched {
		return true
	}
	// Also try matching just the filename
	matched, _ = filepath.Match(pattern, filepath.Base(file))
	return matched
}

func buildDiffFromFiles(files []github.PRFile) string {
	var parts []string
	for _, f := range files {
		if f.Patch != "" {
			parts = append(parts, fmt.Sprintf("--- a/%s\n+++ b/%s\n%s", f.Filename, f.Filename, f.Patch))
		}
	}
	return strings.Join(parts, "\n\n")
}

func getPRFilenames(files []github.PRFile) []string {
	var result []string
	for _, f := range files {
		result = append(result, f.Filename)
	}
	return result
}

func getChangedFiles(repo git.Repo, base, head string) []string {
	status, err := repo.Status()
	if err != nil {
		return nil
	}
	var files []string
	files = append(files, status.Staged...)
	files = append(files, status.Unstaged...)
	return files
}

func writeSARIF(out io.Writer, findings []Finding) error {
	// Build SARIF output
	rules := make(map[string]SARIFRule)
	results := make([]SARIFResult, 0, len(findings))

	for _, f := range findings {
		ruleID := f.Pass + "/" + sanitizeRuleID(f.Message)
		if _, ok := rules[ruleID]; !ok {
			rules[ruleID] = SARIFRule{
				ID:          ruleID,
				Name:        f.Pass + ": " + f.Message[:min(50, len(f.Message))],
				ShortDescription: SARIFDescription{Text: f.Message},
				FullDescription: SARIFDescription{Text: f.Message + "\n\nSuggestion: " + f.Suggestion},
				DefaultConfiguration: SARIFConfig{Level: sarifLevel(f.Severity)},
			}
		}

		results = append(results, SARIFResult{
			RuleID: ruleID,
			Level:  sarifLevel(f.Severity),
			Message: SARIFMessage{Text: f.Message + "\n\nSuggestion: " + f.Suggestion},
			Locations: []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{URI: f.File},
						Region:           SARIFRegion{StartLine: f.Line},
					},
				},
			},
		})
	}

	// Convert rules map to slice
	ruleSlice := make([]SARIFRule, 0, len(rules))
	for _, r := range rules {
		ruleSlice = append(ruleSlice, r)
	}

	sarif := SARIFOutput{
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "ai-dev-tools review",
						Version:        "1.0.0",
						InformationURI: "https://github.com/ernedo/ai-dev-tools",
						Rules:          ruleSlice,
					},
				},
				Results: results,
			},
		},
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sarif)
}

func sarifLevel(severity string) string {
	switch severity {
	case "error":
		return "error"
	case "warning":
		return "warning"
	default:
		return "note"
	}
}

func sanitizeRuleID(msg string) string {
	// Create a simple rule ID from message
	msg = strings.ToLower(msg)
	msg = strings.ReplaceAll(msg, " ", "_")
	msg = strings.ReplaceAll(msg, ".", "")
	msg = strings.ReplaceAll(msg, ",", "")
	if len(msg) > 50 {
		msg = msg[:50]
	}
	return msg
}

func writeJSON(out io.Writer, findings []Finding) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(findings)
}

func printFindings(out io.Writer, findings []Finding) error {
	if len(findings) == 0 {
		fmt.Fprintln(out, "No findings.")
		return nil
	}

	for _, f := range findings {
		severityIcon := map[string]string{
			"error":   "✗",
			"warning": "⚠",
			"info":    "ℹ",
		}
		icon := severityIcon[f.Severity]
		fmt.Fprintf(out, "%s [%s] %s:%d: %s\n", icon, strings.ToUpper(f.Severity), f.File, f.Line, f.Message)
		if f.Suggestion != "" {
			fmt.Fprintf(out, "  Suggestion: %s\n", f.Suggestion)
		}
		fmt.Fprintf(out, "  Pass: %s\n\n", f.Pass)
	}
	return nil
}

func postGitHubReview(ctx context.Context, ghClient github.Client, owner, repo string, prNum int, findings []Finding) error {
	if len(findings) == 0 {
		// Post approval review
		return ghClient.CreateReview(owner, repo, prNum, github.ReviewRequest{
			Body:  "AI Code Review: No issues found. ✓",
			Event: "APPROVE",
		})
	}

	// Build review comments
	var comments []github.ReviewComment
	var body strings.Builder
	body.WriteString("## AI Code Review Results\n\n")

	// Group by severity
	severities := []string{"error", "warning", "info"}
	for _, sev := range severities {
		var sevFindings []Finding
		for _, f := range findings {
			if f.Severity == sev {
				sevFindings = append(sevFindings, f)
			}
		}
		if len(sevFindings) == 0 {
			continue
		}

		icon := map[string]string{"error": "🔴", "warning": "🟡", "info": "🔵"}[sev]
		body.WriteString(fmt.Sprintf("### %s %s (%d)\n\n", icon, strings.Title(sev), len(sevFindings)))

		for _, f := range sevFindings {
			body.WriteString(fmt.Sprintf("- **%s:%d** [%s] %s\n", f.File, f.Line, f.Pass, f.Message))
			if f.Suggestion != "" {
				body.WriteString(fmt.Sprintf("  > *Suggestion: %s*\n", f.Suggestion))
			}
			// Add line comment for GitHub
			comments = append(comments, github.ReviewComment{
				Path: f.File,
				Line: f.Line,
				Side: "RIGHT",
				Body: fmt.Sprintf("[%s] %s\n\n> %s", strings.ToUpper(f.Severity), f.Message, f.Suggestion),
			})
		}
		body.WriteString("\n")
	}

	// Determine overall event
	event := "COMMENT"
	hasErrors := false
	for _, f := range findings {
		if f.Severity == "error" {
			hasErrors = true
			break
		}
	}
	if hasErrors {
		event = "REQUEST_CHANGES"
	} else if len(findings) > 0 {
		event = "COMMENT"
	}

	review := github.ReviewRequest{
		Body:     body.String(),
		Event:    event,
		Comments: comments,
	}

	return ghClient.CreateReview(owner, repo, prNum, review)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}