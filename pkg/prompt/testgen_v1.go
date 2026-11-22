package prompt

// TestGenTemplateV1 is the built-in template for generating tests.
var TestGenTemplateV1 = &Template{
	Name:    "testgen",
	Version: "1",
	Description: "Generate unit tests for a given source file",
	SystemPrompt: `You are a senior engineer writing comprehensive unit tests.
Follow Go testing conventions. Use table-driven tests where appropriate.
Test edge cases, error paths, and happy paths.
Output the test file content only.`,
	UserPrompt: `Generate unit tests for this Go source file:
---
{{.Source}}
---
Package: {{.Package}}
File: {{.File}}
Focus: {{.Focus}}`,
	InputSchema:  mustParseSchema(testgenInputSchema),
	OutputSchema: mustParseSchema(testgenOutputSchema),
}

const testgenInputSchema = `{
	"type": "object",
	"properties": {
		"Source": {"type": "string"},
		"Package": {"type": "string"},
		"File": {"type": "string"},
		"Focus": {"type": "string"}
	},
	"required": ["Source", "Package", "File"]
}`

const testgenOutputSchema = `{
	"type": "object",
	"properties": {
		"test_file": {"type": "string"},
		"package": {"type": "string"}
	},
	"required": ["test_file", "package"]
}`