package prompt

// ReviewTemplateV1 is the built-in template for code review.
var ReviewTemplateV1 = &Template{
	Name:    "review",
	Version: "1",
	Description: "Perform code review on a diff",
	SystemPrompt: `You are a senior engineer performing a code review.
Focus on: correctness, security, performance, maintainability, and style.
Be constructive and specific. Reference line numbers when possible.
Output a JSON object with findings.`,
	UserPrompt: `Review this code change:
---
{{.Diff}}
---
Repository context: {{.Context}}
Files changed: {{.Files}}
Review focus: {{.Focus}}`,
	InputSchema:  mustParseSchema(reviewInputSchema),
	OutputSchema: mustParseSchema(reviewOutputSchema),
}

const reviewInputSchema = `{
	"type": "object",
	"properties": {
		"Diff": {"type": "string"},
		"Context": {"type": "string"},
		"Files": {"type": "array", "items": {"type": "string"}},
		"Focus": {"type": "string"}
	},
	"required": ["Diff", "Context", "Files"]
}`

const reviewOutputSchema = `{
	"type": "object",
	"properties": {
		"summary": {"type": "string"},
		"findings": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"file": {"type": "string"},
					"line": {"type": "integer"},
					"severity": {"type": "string", "enum": ["error", "warning", "info"]},
					"message": {"type": "string"},
					"suggestion": {"type": "string"}
				},
				"required": ["file", "severity", "message"]
			}
		},
		"approved": {"type": "boolean"}
	},
	"required": ["summary", "findings", "approved"]
}`