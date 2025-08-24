package route

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Service chat provides AI chat completion functionality
//
//encore:service
type Service struct{}

// getAPIKey returns the API key for the specified provider
func (s *Service) getAPIKey(provider string) string {
	// Load .env file if it exists
	loadEnvFile()

	switch provider {
	case "groq":
		return os.Getenv("GROQ_API_KEY")
	case "openrouter":
		return os.Getenv("OPENROUTER_API_KEY")
	case "gemini":
		return os.Getenv("GEMINI_API_KEY")
	case "atlas":
		return os.Getenv("ATLASCLOUD_API_KEY")
	case "chutes":
		return os.Getenv("CHUTES_API_KEY")
	default:
		return ""
	}
}

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

// ChatCompletion handles chat completion requests
//
//encore:api public method=POST path=/chat/completions
func (s *Service) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Get provider (default to groq if not specified)
	providerName := req.Provider
	if providerName == "" {
		providerName = "groq"
	}

	// Get API key
	apiKey := s.getAPIKey(providerName)
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found for provider: %s", providerName)
	}

	// Call the appropriate provider
	switch providerName {
	case "groq":
		return s.callGroqAPI(req, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

// HealthCheck returns the health status of the service
//
//encore:api public method=GET path=/health
func (s *Service) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	// Check if at least one API key is available
	chatStatus := "healthy"
	if s.getAPIKey("groq") == "" && s.getAPIKey("openrouter") == "" &&
		s.getAPIKey("gemini") == "" && s.getAPIKey("atlas") == "" &&
		s.getAPIKey("chutes") == "" {
		chatStatus = "no_api_keys"
	}

	return &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Services: map[string]string{
			"chat": chatStatus,
		},
	}, nil
}

// GetProviders returns the list of supported AI providers
//
//encore:api public method=GET path=/providers
func (s *Service) GetProviders(ctx context.Context) (*ProvidersResponse, error) {
	return &ProvidersResponse{
		Providers: []string{"groq", "openrouter", "gemini", "atlas", "chutes"},
	}, nil
}

// TestProvider tests if a specific provider is working
//
//encore:api public method=POST path=/providers/test
func (s *Service) TestProvider(ctx context.Context, req *TestProviderRequest) (*TestProviderResponse, error) {
	if req.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	// Check if API key is available
	apiKey := s.getAPIKey(req.Provider)
	if apiKey == "" {
		return &TestProviderResponse{
			Provider: req.Provider,
			Status:   "no_api_key",
		}, nil
	}

	// Validate provider name
	validProviders := []string{"groq", "openrouter", "gemini", "atlas", "chutes"}
	isValid := false
	for _, p := range validProviders {
		if p == req.Provider {
			isValid = true
			break
		}
	}

	if !isValid {
		return &TestProviderResponse{
			Provider: req.Provider,
			Status:   "invalid_provider",
		}, nil
	}

	return &TestProviderResponse{
		Provider: req.Provider,
		Status:   "healthy",
	}, nil
}

// Helper function to get model
func getModel(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	return "llama3-8b-8192" // default
}

// Helper function to generate a simple ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// callGroqAPI calls the Groq API directly
func (s *Service) callGroqAPI(req *ChatRequest, apiKey string) (*ChatResponse, error) {
	// Prepare the request payload
	messages := []map[string]string{
		{
			"role":    "user",
			"content": req.Prompt,
		},
	}

	model := req.Model
	if model == "" {
		model = "llama3-8b-8192"
	}

	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		payload["max_tokens"] = *req.MaxTokens
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var groqResponse struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &groqResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to our response format
	response := &ChatResponse{
		ID:      groqResponse.ID,
		Object:  groqResponse.Object,
		Created: groqResponse.Created,
		Model:   groqResponse.Model,
		Choices: make([]Choice, len(groqResponse.Choices)),
		Usage: Usage{
			PromptTokens:     groqResponse.Usage.PromptTokens,
			CompletionTokens: groqResponse.Usage.CompletionTokens,
			TotalTokens:      groqResponse.Usage.TotalTokens,
		},
	}

	for i, choice := range groqResponse.Choices {
		response.Choices[i] = Choice{
			Index: choice.Index,
			Message: ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return response, nil
}

// Helper function to count tokens (simple approximation)
func countTokens(text string) int {
	// Simple approximation: 1 token per 4 characters
	return len(text) / 4
}
