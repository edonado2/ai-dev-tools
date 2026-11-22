package llm

import (
	"fmt"
)

var registry = map[string]func(Config) (Provider, error){
	"openai":    NewOpenAI,
	"anthropic": NewAnthropic,
	"ollama":    NewOllama,
	"custom":    NewCustom,
}

func New(name string, cfg Config) (Provider, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return f(cfg)
}