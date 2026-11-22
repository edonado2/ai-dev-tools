package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type anthropicProvider struct {
	client anthropic.Client
	model  string
}

func NewAnthropic(cfg Config) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("Anthropic API key is required")
	}
	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	client := anthropic.NewClient(opts...)
	model := cfg.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &anthropicProvider{client: client, model: model}, nil
}

func (p *anthropicProvider) Name() string { return "anthropic" }

func (p *anthropicProvider) MaxTokens() int { return 200000 }

func (p *anthropicProvider) SupportsTools() bool { return true }

func (p *anthropicProvider) SupportsJSONMode() bool { return true }

func (p *anthropicProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	params := anthropic.MessageNewParams{
		Model:       p.model,
		Messages:    toAnthropicMessages(req.Messages),
		MaxTokens:   int64(req.MaxTokens),
		Temperature: anthropic.Float(float64(req.Temperature)),
		StopSequences: req.Stop,
	}

	if len(req.Tools) > 0 {
		params.Tools = toAnthropicTools(req.Tools)
	}
	// Handle system message separately
	systemBlocks := extractSystemMessages(req.Messages)
	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Anthropic completion error: %w", err)
	}

	var content string
	var toolCalls []ToolCall
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "tool_use":
			toolCalls = append(toolCalls, ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: FunctionCall{
					Name:      block.Name,
					Arguments: string(block.Input),
				},
			})
		}
	}

	return &CompletionResponse{
		Content:   content,
		ToolCalls: toolCalls,
		Usage: Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
		Model: resp.Model,
	}, nil
}

func (p *anthropicProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	params := anthropic.MessageNewParams{
		Model:       p.model,
		Messages:    toAnthropicMessages(req.Messages),
		MaxTokens:   int64(req.MaxTokens),
		Temperature: anthropic.Float(float64(req.Temperature)),
		StopSequences: req.Stop,
	}

	if len(req.Tools) > 0 {
		params.Tools = toAnthropicTools(req.Tools)
	}
	systemBlocks := extractSystemMessages(req.Messages)
	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}

	go func() {
		defer close(ch)
		stream := p.client.Messages.NewStreaming(ctx, params)
		for stream.Next() {
			event := stream.Current()
			// Check event type using the As methods
			if event.Type == "content_block_delta" {
				deltaEvent := event.AsContentBlockDelta()
				textDelta := deltaEvent.Delta.AsTextDelta()
				ch <- StreamChunk{Content: textDelta.Text}
			} else if event.Type == "message_stop" {
				ch <- StreamChunk{Done: true}
			}
		}
		if err := stream.Err(); err != nil {
			ch <- StreamChunk{Err: err}
		}
	}()

	return ch, nil
}

func extractSystemMessages(msgs []Message) []anthropic.TextBlockParam {
	var result []anthropic.TextBlockParam
	for _, m := range msgs {
		if m.Role == "system" {
			result = append(result, anthropic.TextBlockParam{Text: m.Content})
		}
	}
	return result
}

func toAnthropicMessages(msgs []Message) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "system":
			// System messages are handled separately in the System parameter
			continue
		case "user":
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					result = append(result, anthropic.NewUserMessage(
						anthropic.NewToolResultBlock(tc.ID, tc.Function.Arguments, false),
					))
				}
			} else {
				result = append(result, anthropic.NewUserMessage(
					anthropic.NewTextBlock(m.Content),
				))
			}
		case "assistant":
			if len(m.ToolCalls) > 0 {
				blocks := make([]anthropic.ContentBlockParamUnion, len(m.ToolCalls))
				for i, tc := range m.ToolCalls {
					blocks[i] = anthropic.NewToolUseBlock(tc.ID, tc.Function.Arguments, tc.Function.Name)
				}
				result = append(result, anthropic.NewAssistantMessage(blocks...))
			} else {
				result = append(result, anthropic.NewAssistantMessage(
					anthropic.NewTextBlock(m.Content),
				))
			}
		case "tool":
			result = append(result, anthropic.NewUserMessage(
				anthropic.NewToolResultBlock(m.ToolCallID, m.Content, false),
			))
		default:
			result = append(result, anthropic.NewUserMessage(
				anthropic.NewTextBlock(m.Content),
			))
		}
	}
	return result
}

func toAnthropicTools(tools []Tool) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(tools))
	for i, t := range tools {
		params := map[string]any{}
		if t.Function.Parameters != nil {
			params = t.Function.Parameters
		}
		toolParam := anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Properties: params,
			},
			t.Function.Name,
		)
		if t.Function.Description != "" {
			// Description is in the schema
		}
		result[i] = toolParam
	}
	return result
}