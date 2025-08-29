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

// GroqProvider implements the Provider interface for Groq API
type GroqProvider struct{}

// NewGroqProvider creates a new Groq provider instance
func NewGroqProvider(cfg *config.Config) *GroqProvider {
	return &GroqProvider{}
}

// GetName returns the provider name
func (g *GroqProvider) GetName() string {
	return "groq"
}

// ChatCompletion calls the Groq API for chat completion
func (g *GroqProvider) ChatCompletion(req *models.ChatRequest, apiKey string) (*models.ChatResponse, error) {
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
		model = "openai/gpt-oss-120b"
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
	response := &models.ChatResponse{
		ID:      groqResponse.ID,
		Object:  groqResponse.Object,
		Created: groqResponse.Created,
		Model:   groqResponse.Model,
		Choices: make([]models.Choice, len(groqResponse.Choices)),
		Usage: models.Usage{
			PromptTokens:     groqResponse.Usage.PromptTokens,
			CompletionTokens: groqResponse.Usage.CompletionTokens,
			TotalTokens:      groqResponse.Usage.TotalTokens,
		},
	}

	for i, choice := range groqResponse.Choices {
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
