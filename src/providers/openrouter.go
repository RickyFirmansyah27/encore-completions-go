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

// OpenRouterProvider implements the Provider interface for OpenRouter API
type OpenRouterProvider struct{}

// NewOpenRouterProvider creates a new OpenRouter provider instance
func NewOpenRouterProvider(cfg *config.Config) *OpenRouterProvider {
	return &OpenRouterProvider{}
}

// GetName returns the provider name
func (o *OpenRouterProvider) GetName() string {
	return "openrouter"
}

// ChatCompletion calls the OpenRouter API for chat completion
func (o *OpenRouterProvider) ChatCompletion(req *models.ChatRequest, apiKey string) (*models.ChatResponse, error) {
	// Prepare the request payload
	var messages []map[string]interface{}

	for _, msg := range req.Messages {
		var contentParts []map[string]interface{}
		for _, part := range msg.Content {
			if part.Type == "text" {
				contentParts = append(contentParts, map[string]interface{}{
					"type": "text",
					"text": part.Text,
				})
			} else if part.Type == "image_url" && part.ImageURL != nil {
				contentParts = append(contentParts, map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]string{
						"url": part.ImageURL.URL,
					},
				})
			}
		}
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": contentParts,
		})
	}

	model := req.Model
	if model == "" {
		if req.WithImage {
			model = "google/gemini-2.5-flash-image-preview:free"
		} else {
			model = "deepseek/deepseek-chat-v3.1:free"
		}
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
	httpReq, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://encore-completion-go")
	httpReq.Header.Set("X-Title", "Encore Chat Completion")

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

	// Parse response (OpenRouter uses OpenAI-compatible format)
	var openRouterResponse struct {
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

	if err := json.Unmarshal(body, &openRouterResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to our response format
	response := &models.ChatResponse{
		ID:      openRouterResponse.ID,
		Object:  openRouterResponse.Object,
		Created: openRouterResponse.Created,
		Model:   openRouterResponse.Model,
		Choices: make([]models.Choice, len(openRouterResponse.Choices)),
		Usage: models.Usage{
			PromptTokens:     openRouterResponse.Usage.PromptTokens,
			CompletionTokens: openRouterResponse.Usage.CompletionTokens,
			TotalTokens:      openRouterResponse.Usage.TotalTokens,
		},
	}

	for i, choice := range openRouterResponse.Choices {
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
