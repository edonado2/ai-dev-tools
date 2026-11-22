package prompt

import (
	"bytes"
	"encoding/json"
	"errors"
	"text/template"

	"github.com/invopop/jsonschema"
	"github.com/ernedo/ai-dev-tools/pkg/llm"
)

// Template represents a versioned prompt template with JSON Schema validation.
type Template struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Description  string `yaml:"description"`
	SystemPrompt string `yaml:"system_prompt"`
	UserPrompt   string `yaml:"user_prompt"`
	InputSchema  *jsonschema.Schema `yaml:"-"`
	OutputSchema *jsonschema.Schema `yaml:"-"`
	Tools        []llm.Tool `yaml:"-"`
}

// Render validates input variables against InputSchema, executes the prompt templates,
// and returns a CompletionRequest.
func (t *Template) Render(vars map[string]any) (*llm.CompletionRequest, error) {
	// Validate vars against InputSchema
	if t.InputSchema != nil {
		if err := validateInput(vars, t.InputSchema); err != nil {
			return nil, err
		}
	}

	// Execute SystemPrompt as Go template
	systemPrompt, err := executeTemplate(t.SystemPrompt, vars)
	if err != nil {
		return nil, err
	}

	// Execute UserPrompt as Go template
	userPrompt, err := executeTemplate(t.UserPrompt, vars)
	if err != nil {
		return nil, err
	}

	// Build CompletionRequest
	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Tools: t.Tools,
	}

	// Set ResponseFormat from OutputSchema if available
	if t.OutputSchema != nil {
		req.ResponseFormat = &llm.ResponseFormat{
			Type:       "json_schema",
			JSONSchema: schemaToMap(t.OutputSchema),
		}
	}

	return req, nil
}

// executeTemplate executes a Go template with the given variables.
func executeTemplate(tmplStr string, vars map[string]any) (string, error) {
	tmpl, err := template.New("prompt").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// validateInput validates the input variables against the JSON schema.
func validateInput(vars map[string]any, schema *jsonschema.Schema) error {
	// Check required fields
	if schema.Required != nil {
		for _, field := range schema.Required {
			if _, ok := vars[field]; !ok {
				return errors.New("missing required field: " + field)
			}
		}
	}
	return nil
}

// schemaToMap converts a jsonschema.Schema to a map for ResponseFormat.
func schemaToMap(schema *jsonschema.Schema) map[string]any {
	// Convert schema to JSON and back to map
	data, _ := json.Marshal(schema)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}

// mustParseSchema parses a JSON schema string and panics on error.
func mustParseSchema(s string) *jsonschema.Schema {
	schema := new(jsonschema.Schema)
	if err := schema.UnmarshalJSON([]byte(s)); err != nil {
		panic("invalid schema: " + err.Error())
	}
	return schema
}