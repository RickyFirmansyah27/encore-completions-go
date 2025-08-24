package models

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Prompt      string   `json:"prompt"`
	Model       string   `json:"model,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Stream      *bool    `json:"stream,omitempty"`
	Provider    string   `json:"provider,omitempty"`
}

// ChatMessage represents a single message in a chat conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Choice represents a choice in the completion response
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatResponse represents a chat completion response (OpenAI compatible)
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// ProvidersResponse represents supported providers response
type ProvidersResponse struct {
	Providers []string `json:"providers"`
}

// TestProviderRequest represents a provider test request
type TestProviderRequest struct {
	Provider string `json:"provider"`
}

// TestProviderResponse represents a provider test response
type TestProviderResponse struct {
	Provider string `json:"provider"`
	Status   string `json:"status"`
}
