package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type customProvider struct {
	client openai.Client
	model  string
}

func NewCustom(cfg Config) (Provider, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("custom provider requires BaseURL")
	}
	if cfg.APIKey == "" {
		return nil, errors.New("custom provider requires APIKey")
	}
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	)
	model := cfg.Model
	if model == "" {
		return nil, errors.New("custom provider requires Model")
	}
	return &customProvider{client: client, model: model}, nil
}

func (p *customProvider) Name() string { return "custom" }

func (p *customProvider) MaxTokens() int { return 128000 }

func (p *customProvider) SupportsTools() bool { return true }

func (p *customProvider) SupportsJSONMode() bool { return true }

func (p *customProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	params := openai.ChatCompletionNewParams{
		Model:       p.model,
		Messages:    toOpenAIMessages(req.Messages),
		Temperature: openai.Float(float64(req.Temperature)),
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Stop:        toOpenAIStop(req.Stop),
	}

	if len(req.Tools) > 0 {
		params.Tools = toOpenAITools(req.Tools)
	}
	if req.ToolChoice != "" {
		params.ToolChoice = toOpenAIToolChoice(req.ToolChoice)
	}
	if req.ResponseFormat != nil {
		params.ResponseFormat = toOpenAIResponseFormat(req.ResponseFormat)
	}

	resp, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("custom provider completion error: %w", err)
	}

	return &CompletionResponse{
		Content:   resp.Choices[0].Message.Content,
		ToolCalls: fromOpenAIToolCalls(resp.Choices[0].Message.ToolCalls),
		Usage: Usage{
			PromptTokens:     int(resp.Usage.PromptTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		},
		Model: resp.Model,
	}, nil
}

func (p *customProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	params := openai.ChatCompletionNewParams{
		Model:       p.model,
		Messages:    toOpenAIMessages(req.Messages),
		Temperature: openai.Float(float64(req.Temperature)),
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Stop:        toOpenAIStop(req.Stop),
	}

	if len(req.Tools) > 0 {
		params.Tools = toOpenAITools(req.Tools)
	}
	if req.ToolChoice != "" {
		params.ToolChoice = toOpenAIToolChoice(req.ToolChoice)
	}
	if req.ResponseFormat != nil {
		params.ResponseFormat = toOpenAIResponseFormat(req.ResponseFormat)
	}

	go func() {
		defer close(ch)
		stream := p.client.Chat.Completions.NewStreaming(ctx, params)
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					ch <- StreamChunk{Content: delta.Content}
				}
			}
		}
		if err := stream.Err(); err != nil {
			ch <- StreamChunk{Err: err}
		}
		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}