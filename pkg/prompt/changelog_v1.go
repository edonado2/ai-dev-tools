package prompt

// ChangelogTemplateV1 is the built-in template for generating changelogs.
var ChangelogTemplateV1 = &Template{
	Name:    "changelog",
	Version: "1",
	Description: "Generate changelog from commit history",
	SystemPrompt: `You are a senior engineer writing a changelog.
Follow Keep a Changelog format.
Group commits by type: Added, Changed, Deprecated, Removed, Fixed, Security.
Output in Markdown format.`,
	UserPrompt: `Generate a changelog from these commits:
---
{{.Commits}}
---
Version: {{.Version}}
Date: {{.Date}}
Previous version: {{.PreviousVersion}}`,
	InputSchema:  mustParseSchema(changelogInputSchema),
	OutputSchema: mustParseSchema(changelogOutputSchema),
}

const changelogInputSchema = `{
	"type": "object",
	"properties": {
		"Commits": {"type": "array", "items": {"type": "string"}},
		"Version": {"type": "string"},
		"Date": {"type": "string"},
		"PreviousVersion": {"type": "string"}
	},
	"required": ["Commits", "Version", "Date"]
}`

const changelogOutputSchema = `{
	"type": "object",
	"properties": {
		"changelog": {"type": "string"},
		"version": {"type": "string"}
	},
	"required": ["changelog", "version"]
}`