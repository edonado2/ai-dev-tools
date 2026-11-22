package llm

import (
	"context"
)

type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	Name() string
	MaxTokens() int
	SupportsTools() bool
	SupportsJSONMode() bool
}