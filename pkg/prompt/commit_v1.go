package prompt

// CommitTemplateV1 is the built-in template for generating conventional commit messages.
var CommitTemplateV1 = &Template{
	Name:    "commit",
	Version: "1",
	Description: "Generate conventional commit messages from git diff",
	SystemPrompt: `You are a senior engineer writing conventional commit messages.
Follow the Conventional Commits spec: <type>(<scope>): <subject>
Types: feat, fix, docs, style, refactor, perf, test, chore, revert.
Subject: imperative, lowercase, no period, max 72 chars.
Body: explain what and why, not how.`,
	UserPrompt: `Generate a commit message for this diff:
---
{{.Diff}}
---
Repository context: {{.Context}}
Files changed: {{.Files}}`,
	InputSchema:  mustParseSchema(commitInputSchema),
	OutputSchema: mustParseSchema(commitOutputSchema),
}

const commitInputSchema = `{
	"type": "object",
	"properties": {
		"Diff": {"type": "string"},
		"Context": {"type": "string"},
		"Files": {"type": "array", "items": {"type": "string"}}
	},
	"required": ["Diff", "Context", "Files"]
}`

const commitOutputSchema = `{
	"type": "object",
	"properties": {
		"message": {"type": "string"},
		"type": {"type": "string"},
		"scope": {"type": "string"},
		"breaking": {"type": "boolean"}
	},
	"required": ["message", "type"]
}`