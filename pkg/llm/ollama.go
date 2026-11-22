package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ollamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllama(cfg Config) (Provider, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	model := cfg.Model
	if model == "" {
		model = "codellama"
	}
	return &ollamaProvider{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 5 * time.Minute},
	}, nil
}

func (p *ollamaProvider) Name() string { return "ollama" }

func (p *ollamaProvider) MaxTokens() int { return 8192 }

func (p *ollamaProvider) SupportsTools() bool { return false }

func (p *ollamaProvider) SupportsJSONMode() bool { return true }

func (p *ollamaProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	body := map[string]any{
		"model":    p.model,
		"messages": toOllamaMessages(req.Messages),
		"stream":   false,
		"options": map[string]any{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
		},
	}
	if len(req.Stop) > 0 {
		body["options"].(map[string]any)["stop"] = req.Stop
	}
	if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_schema" {
		body["format"] = req.ResponseFormat.JSONSchema
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("ollama request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama decode error: %w", err)
	}

	return &CompletionResponse{
		Content: result.Message.Content,
		Usage:   Usage{},
		Model:   p.model,
	}, nil
}

func (p *ollamaProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	body := map[string]any{
		"model":    p.model,
		"messages": toOllamaMessages(req.Messages),
		"stream":   true,
		"options": map[string]any{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
		},
	}
	if len(req.Stop) > 0 {
		body["options"].(map[string]any)["stop"] = req.Stop
	}
	if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_schema" {
		body["format"] = req.ResponseFormat.JSONSchema
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("ollama request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	go func() {
		defer close(ch)
		resp, err := p.client.Do(httpReq)
		if err != nil {
			ch <- StreamChunk{Err: err}
			return
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}
			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				ch <- StreamChunk{Err: err}
				return
			}
			if chunk.Message.Content != "" {
				ch <- StreamChunk{Content: chunk.Message.Content}
			}
			if chunk.Done {
				ch <- StreamChunk{Done: true}
				break
			}
		}
	}()

	return ch, nil
}

func toOllamaMessages(msgs []Message) []map[string]string {
	result := make([]map[string]string, len(msgs))
	for i, m := range msgs {
		result[i] = map[string]string{"role": m.Role, "content": m.Content}
	}
	return result
}