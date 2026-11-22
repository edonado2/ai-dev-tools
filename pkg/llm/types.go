package llm

type Message struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

type ToolCall struct {
	ID       string
	Type     string
	Function FunctionCall
}

type FunctionCall struct {
	Name      string
	Arguments string
}

type CompletionRequest struct {
	Messages       []Message
	Temperature    float32
	MaxTokens      int
	Tools          []Tool
	ToolChoice     string
	ResponseFormat *ResponseFormat
	Stop           []string
}

type ResponseFormat struct {
	Type       string
	JSONSchema map[string]any
}

type CompletionResponse struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
	Model     string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type StreamChunk struct {
	Content string
	Done    bool
	Err     error
}

type Tool struct {
	Type     string
	Function ToolFunction
}

type ToolFunction struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// Config holds provider configuration
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}