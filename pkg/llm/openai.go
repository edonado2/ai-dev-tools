package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type openaiProvider struct {
	client openai.Client
	model  string
}

func NewOpenAI(cfg Config) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}
	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	client := openai.NewClient(opts...)
	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}
	return &openaiProvider{client: client, model: model}, nil
}

func (p *openaiProvider) Name() string { return "openai" }

func (p *openaiProvider) MaxTokens() int { return 128000 }

func (p *openaiProvider) SupportsTools() bool { return true }

func (p *openaiProvider) SupportsJSONMode() bool { return true }

func (p *openaiProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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
		return nil, fmt.Errorf("OpenAI completion error: %w", err)
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

func (p *openaiProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
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

func toOpenAIMessages(msgs []Message) []openai.ChatCompletionMessageParamUnion {
	result := make([]openai.ChatCompletionMessageParamUnion, len(msgs))
	for i, m := range msgs {
		switch m.Role {
		case "system":
			result[i] = openai.SystemMessage(m.Content)
		case "user":
			result[i] = openai.UserMessage(m.Content)
		case "assistant":
			if len(m.ToolCalls) > 0 {
				// Create assistant message with tool calls
				toolCalls := toOpenAIToolCalls(m.ToolCalls)
				result[i] = openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content:   openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.String(m.Content)},
						ToolCalls: toolCalls,
					},
				}
			} else {
				result[i] = openai.AssistantMessage(m.Content)
			}
		case "tool":
			result[i] = openai.ToolMessage(m.Content, m.ToolCallID)
		default:
			result[i] = openai.UserMessage(m.Content)
		}
	}
	return result
}

func toOpenAIToolCalls(calls []ToolCall) []openai.ChatCompletionMessageToolCallParam {
	result := make([]openai.ChatCompletionMessageToolCallParam, len(calls))
	for i, c := range calls {
		result[i] = openai.ChatCompletionMessageToolCallParam{
			ID:   c.ID,
			Type: "function",
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      c.Function.Name,
				Arguments: c.Function.Arguments,
			},
		}
	}
	return result
}

func fromOpenAIToolCalls(calls []openai.ChatCompletionMessageToolCall) []ToolCall {
	if calls == nil {
		return nil
	}
	result := make([]ToolCall, len(calls))
	for i, c := range calls {
		result[i] = ToolCall{
			ID:   c.ID,
			Type: string(c.Type),
			Function: FunctionCall{
				Name:      c.Function.Name,
				Arguments: c.Function.Arguments,
			},
		}
	}
	return result
}

func toOpenAITools(tools []Tool) []openai.ChatCompletionToolParam {
	result := make([]openai.ChatCompletionToolParam, len(tools))
	for i, t := range tools {
		params := map[string]any{}
		if t.Function.Parameters != nil {
			params = t.Function.Parameters
		}
		result[i] = openai.ChatCompletionToolParam{
			Type: "function",
			Function: shared.FunctionDefinitionParam{
				Name:        t.Function.Name,
				Description: openai.String(t.Function.Description),
				Parameters:  params,
			},
		}
	}
	return result
}

func toOpenAIStop(stop []string) openai.ChatCompletionNewParamsStopUnion {
	if len(stop) == 0 {
		return openai.ChatCompletionNewParamsStopUnion{}
	}
	if len(stop) == 1 {
		return openai.ChatCompletionNewParamsStopUnion{OfString: openai.String(stop[0])}
	}
	return openai.ChatCompletionNewParamsStopUnion{OfStringArray: stop}
}

func toOpenAIToolChoice(choice string) openai.ChatCompletionToolChoiceOptionUnionParam {
	if choice == "auto" || choice == "none" {
		return openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String(choice)}
	}
	// For specific function, we'd need the function name
	return openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String("auto")}
}

func toOpenAIResponseFormat(rf *ResponseFormat) openai.ChatCompletionNewParamsResponseFormatUnion {
	if rf.Type == "json_schema" && rf.JSONSchema != nil {
		schemaBytes, _ := json.Marshal(rf.JSONSchema)
		var schema map[string]any
		json.Unmarshal(schemaBytes, &schema)
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "response",
					Schema: schema,
					Strict: openai.Bool(true),
				},
				Type: "json_schema",
			},
		}
	}
	return openai.ChatCompletionNewParamsResponseFormatUnion{
		OfText: &shared.ResponseFormatTextParam{Type: "text"},
	}
}