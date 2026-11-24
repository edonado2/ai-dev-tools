# Prompt Templates

ai-dev-tools uses versioned prompt templates with JSON Schema validation for structured LLM output.

## Built-in Templates

| Template | Version | Description |
|----------|---------|-------------|
| `commit` | v1 | Conventional commit message generation |
| `review` | v1 | Multi-pass code review |
| `testgen` | v1 | Unit test generation |
| `changelog` | v1 | Keep a Changelog generation |

## Template Structure

Each template is a Go struct with:

```go
type Template struct {
    Name         string
    Version      string
    Description  string
    SystemPrompt string
    UserPrompt   string
    InputSchema  *jsonschema.Schema
    OutputSchema *jsonschema.Schema
    Tools        []llm.Tool
}
```

## User Overrides

Place custom templates in `~/.config/ai-dev-tools/templates/` as YAML files:

```
~/.config/ai-dev-tools/templates/
├── commit@v1.yaml
├── review@v1.yaml
├── testgen@v1.yaml
└── changelog@v1.yaml
```

### Example Override

`~/.config/ai-dev-tools/templates/commit@v1.yaml`:

```yaml
name: commit
version: "1"
description: "Custom commit message generator"
system_prompt: |
  You are a senior engineer writing conventional commit messages.
  Follow the Conventional Commits spec.
  Add emoji prefixes: ✨ feat, 🐛 fix, 📚 docs, ♻️ refactor, ⚡ perf, ✅ test, 🔧 chore
user_prompt: |
  Generate a commit message for this diff:
  ---
  {{.Diff}}
  ---
  Repository context: {{.Context}}
  Files changed: {{.Files}}
input_schema: |
  {
    "type": "object",
    "properties": {
      "Diff": {"type": "string"},
      "Context": {"type": "string"},
      "Files": {"type": "array", "items": {"type": "string"}}
    },
    "required": ["Diff", "Context", "Files"]
  }
output_schema: |
  {
    "type": "object",
    "properties": {
      "message": {"type": "string"},
      "type": {"type": "string"},
      "scope": {"type": "string"},
      "breaking": {"type": "boolean"}
    },
    "required": ["message", "type"]
  }
```

## Template Variables

### commit@v1

| Variable | Type | Description |
|----------|------|-------------|
| `Diff` | string | Git diff of staged changes |
| `Context` | string | Repository context (recent commits) |
| `Files` | []string | List of changed files |

### review@v1

| Variable | Type | Description |
|----------|------|-------------|
| `Diff` | string | Git diff to review |
| `PassName` | string | Current pass (security, bugs, style, performance) |

### testgen@v1

| Variable | Type | Description |
|----------|------|-------------|
| `SourceCode` | string | Source file content |
| `Framework` | string | Target test framework |
| `ExistingTests` | string | Existing test content (if any) |

### changelog@v1

| Variable | Type | Description |
|----------|------|-------------|
| `Commits` | []Commit | List of commits since last tag |
| `PRData` | map[int]PR | PR enrichment data (optional) |
| `Format` | string | Output format (keepachangelog) |
| `PreviousVersion` | string | Previous version tag |
| `Date` | string | Current date |

## Output Schemas

### commit@v1 Output

```json
{
  "message": "feat(cli): add config init command",
  "type": "feat",
  "scope": "cli",
  "breaking": false
}
```

### review@v1 Output (per finding)

```json
{
  "file": "handler.go",
  "line": 42,
  "severity": "error",
  "category": "security",
  "message": "Potential SQL injection",
  "suggestion": "Use parameterized queries",
  "pass": "security"
}
```

### testgen@v1 Output

```json
{
  "test_code": "func TestHandler(t *testing.T) { ... }",
  "framework": "gotest",
  "file_path": "handler_test.go"
}
```

### changelog@v1 Output

```json
{
  "version": "1.1.0",
  "date": "2024-01-15",
  "sections": {
    "Added": ["feat(cli): add config init command"],
    "Fixed": ["fix(review): handle empty diff"]
  }
}
```

## JSON Schema Validation

All templates validate inputs and outputs against JSON Schemas:

1. **Input validation** - Before sending to LLM, variables are validated against `InputSchema`
2. **Output validation** - LLM response is validated against `OutputSchema`
3. **Tool calling** - Structured output enforced via LLM tool/function calling

## Creating Custom Templates

1. Copy a built-in template as starting point
2. Modify `system_prompt` and `user_prompt`
3. Update `input_schema` and `output_schema` if variables change
4. Save as `{name}@v{version}.yaml` in user templates directory
5. Run `ai-dev doctor` to validate template loads correctly

## Versioning

- Templates are versioned independently (v1, v2, etc.)
- Multiple versions can coexist
- Specify version in code: `prompt.Registry.Get("commit", "1")`
- User overrides take precedence over built-ins at same version