package models

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"` // Added to support multi-modal
	Prompt      string        `json:"prompt"`
	Model       string        `json:"model,omitempty"`
	Temperature *float64      `json:"temperature,omitempty" default:"0.5"`
	MaxTokens   *int          `json:"max_tokens,omitempty" default:"1000"`
	Stream      *bool         `json:"stream,omitempty"`
	Provider    string        `json:"provider,omitempty"`
	WithImage   bool          `json:"withImage,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
}

// Tool represents a tool that can be used by the model
type Tool struct {
	GoogleSearch GoogleSearch `json:"google_search,omitempty"`
}

// GoogleSearch represents the Google Search tool
type GoogleSearch struct{}

// ContentPart represents a part of the content (text or image)
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL
type ImageURL struct {
	URL string `json:"url"`
}

// ChatMessage represents a single message in a chat conversation
type ChatMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
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
