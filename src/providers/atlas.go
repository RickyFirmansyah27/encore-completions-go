package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"encore.app/src/models"
)

// AtlasProvider implements the Provider interface for Atlas API
type AtlasProvider struct{}

// NewAtlasProvider creates a new Atlas provider instance
func NewAtlasProvider() Provider {
	return &AtlasProvider{}
}

// GetName returns the provider name
func (a *AtlasProvider) GetName() string {
	return "atlas"
}

// ChatCompletion calls the Atlas API for chat completion
func (a *AtlasProvider) ChatCompletion(req *models.ChatRequest, apiKey string) (*models.ChatResponse, error) {
	// Prepare the request payload
	messages := []map[string]string{
		{
			"role":    "user",
			"content": req.Prompt,
		},
	}

	model := req.Model
	if model == "" {
		model = "openai/gpt-oss-20b"
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

	// Create HTTP request (Note: This is a placeholder URL - adjust as needed for actual Atlas API)
	httpReq, err := http.NewRequest("POST", "https://api.atlascloud.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
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

	// Parse response (assuming OpenAI-compatible format)
	var atlasResponse struct {
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

	if err := json.Unmarshal(body, &atlasResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to our response format
	response := &models.ChatResponse{
		ID:      atlasResponse.ID,
		Object:  atlasResponse.Object,
		Created: atlasResponse.Created,
		Model:   atlasResponse.Model,
		Choices: make([]models.Choice, len(atlasResponse.Choices)),
		Usage: models.Usage{
			PromptTokens:     atlasResponse.Usage.PromptTokens,
			CompletionTokens: atlasResponse.Usage.CompletionTokens,
			TotalTokens:      atlasResponse.Usage.TotalTokens,
		},
	}

	for i, choice := range atlasResponse.Choices {
		response.Choices[i] = models.Choice{
			Index: choice.Index,
			Message: models.ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return response, nil
}
