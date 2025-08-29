package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"encore.app/src/config"
	"encore.app/src/models"
)

// ChutesProvider implements the Provider interface for Chutes API
type ChutesProvider struct{}

// NewChutesProvider creates a new Chutes provider instance
func NewChutesProvider(cfg *config.Config) *ChutesProvider {
	return &ChutesProvider{}
}

// GetName returns the provider name
func (c *ChutesProvider) GetName() string {
	return "chutes"
}

// ChatCompletion calls the Chutes API for chat completion
func (c *ChutesProvider) ChatCompletion(req *models.ChatRequest, apiKey string) (*models.ChatResponse, error) {
	// Prepare the request payload
	var messages []interface{}

	var contentParts []models.ContentPart
	contentParts = append(contentParts, models.ContentPart{
		Type: "text",
		Text: req.Prompt,
	})

	if req.WithImage && req.ImageData != "" {
		contentParts = append(contentParts, models.ContentPart{
			Type: "image_url",
			ImageURL: &models.ImageURL{
				URL: "data:image/jpeg;base64," + req.ImageData,
			},
		})
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": contentParts,
	})

	model := req.Model
	if model == "" {
		model = "zai-org/GLM-4.5-FP8"
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

	// Create HTTP request (Note: This is a placeholder URL - adjust as needed for actual Chutes API)
	httpReq, err := http.NewRequest("POST", "https://llm.chutes.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
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
	var chutesResponse struct {
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

	if err := json.Unmarshal(body, &chutesResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to our response format
	response := &models.ChatResponse{
		ID:      chutesResponse.ID,
		Object:  chutesResponse.Object,
		Created: chutesResponse.Created,
		Model:   chutesResponse.Model,
		Choices: make([]models.Choice, len(chutesResponse.Choices)),
		Usage: models.Usage{
			PromptTokens:     chutesResponse.Usage.PromptTokens,
			CompletionTokens: chutesResponse.Usage.CompletionTokens,
			TotalTokens:      chutesResponse.Usage.TotalTokens,
		},
	}

	for i, choice := range chutesResponse.Choices {
		response.Choices[i] = models.Choice{
			Index: choice.Index,
			Message: models.ChatMessage{
				Role: choice.Message.Role,
				Content: []models.ContentPart{
					{
						Type: "text",
						Text: choice.Message.Content,
					},
				},
			},
			FinishReason: choice.FinishReason,
		}
	}

	return response, nil
}
