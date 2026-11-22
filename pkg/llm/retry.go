package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func WithRetry(p Provider, maxRetries int) Provider {
	return &retryProvider{Provider: p, maxRetries: maxRetries}
}

type retryProvider struct {
	Provider
	maxRetries int
}

func (p *retryProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	var lastErr error
	for i := 0; i <= p.maxRetries; i++ {
		resp, err := p.Provider.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}
		if !isRetryable(err) {
			return nil, err
		}
		lastErr = err
		backoff := time.Duration(1<<i) * time.Second
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, lastErr
}

func (p *retryProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	// For streaming, we don't retry - just pass through
	return p.Provider.Stream(ctx, req)
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for context errors (not retryable)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	// Check for rate limiting (429) or server errors (5xx)
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit") {
		return true
	}
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") || strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		return true
	}
	// Check for network errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") {
		return true
	}
	// Check for HTTP status codes if available
	var httpErr *httpError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 429 || (httpErr.StatusCode >= 500 && httpErr.StatusCode < 600)
	}
	return false
}

type httpError struct {
	StatusCode int
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}