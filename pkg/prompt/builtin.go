package prompt

// builtinTemplates returns all built-in templates.
func builtinTemplates() map[string]*Template {
	return map[string]*Template{
		"commit@v1":    CommitTemplateV1,
		"review@v1":    ReviewTemplateV1,
		"testgen@v1":   TestGenTemplateV1,
		"changelog@v1": ChangelogTemplateV1,
	}
}