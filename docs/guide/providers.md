# LLM Providers

ai-dev-tools supports multiple LLM providers through a unified interface.

## Supported Providers

### Anthropic (Default)
```yaml
provider:
  name: anthropic
  model: claude-3-5-sonnet-20241022
  api_key_env: ANTHROPIC_API_KEY
```
Get your API key at [console.anthropic.com](https://console.anthropic.com).

### OpenAI
```yaml
provider:
  name: openai
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
```
Get your API key at [platform.openai.com](https://platform.openai.com).

### Ollama (Local)
```yaml
provider:
  name: ollama
  model: codellama
  base_url: http://localhost:11434
```
Requires [Ollama](https://ollama.ai) running locally. No API key needed.

### Custom (OpenAI-Compatible)
```yaml
provider:
  name: custom
  model: my-model
  base_url: https://api.example.com/v1
  api_key_env: CUSTOM_API_KEY
```
For any OpenAI-compatible endpoint (e.g., LocalAI, vLLM, OpenRouter).

## Provider Comparison

| Provider | Cost | Privacy | Speed | Best For |
|----------|------|---------|-------|----------|
| Anthropic | $$$ | Cloud | Fast | General purpose, coding |
| OpenAI | $$$ | Cloud | Fast | General purpose |
| Ollama | Free | Local | Varies | Privacy-sensitive, offline |
| Custom | Varies | Varies | Varies | Enterprise, specific models |

## Model Recommendations

| Task | Recommended Models |
|------|-------------------|
| Commit messages | claude-3-5-sonnet, gpt-4o, codellama |
| Code review | claude-3-5-sonnet, gpt-4o |
| Test generation | claude-3-5-sonnet, gpt-4o, codellama |
| Changelog | claude-3-5-sonnet, gpt-4o |

## Timeout and Retries

Configure for reliability:

```yaml
provider:
  timeout: 120        # seconds (default: 60)
  max_retries: 5      # default: 3
```

The client implements exponential backoff with jitter for retries on 429/5xx errors.