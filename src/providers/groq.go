package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
	log.Printf("[GROQ] Starting ChatCompletion request for model: %s", req.Model)

	// Prepare the request payload
	messages := make([]map[string]interface{}, 0, len(req.Messages))
	for _, msg := range req.Messages {
		contentParts := make([]map[string]interface{}, 0, len(msg.Content))
		for _, part := range msg.Content {
			if part.Type == "text" {
				contentParts = append(contentParts, map[string]interface{}{
					"type": "text",
					"text": part.Text,
				})
			} else if part.Type == "image_url" && part.ImageURL != nil {
				log.Printf("[GROQ] Processing image URL: %s", part.ImageURL.URL)

				// Validate the image URL format and accessibility
				if strings.TrimSpace(part.ImageURL.URL) == "" {
					log.Printf("[GROQ] Error: Empty image URL")
					return nil, fmt.Errorf("empty image URL provided")
				}

				// Check if it's a reasonable URL format
				if !strings.HasPrefix(part.ImageURL.URL, "http://") && !strings.HasPrefix(part.ImageURL.URL, "https://") {
					log.Printf("[GROQ] Error: Invalid image URL format, not http/https: %s", part.ImageURL.URL)
				}

				contentParts = append(contentParts, map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": part.ImageURL.URL,
					},
				})

				log.Printf("[GROQ] Added image URL to content parts: %s", part.ImageURL.URL)
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
			model = "meta-llama/llama-4-maverick-17b-128e-instruct"
		} else {
			model = "openai/gpt-oss-120b"
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
		log.Printf("[GROQ] API request failed with status %d, response body: %s", resp.StatusCode, string(body))
		log.Printf("[GROQ] Request URL: %s", "https://api.groq.com/openai/v1/chat/completions")
		log.Printf("[GROQ] Model used: %s", model)

		// Check if the error is specifically about media/image access
		if resp.StatusCode == 400 && strings.Contains(string(body), "failed to retrieve media") {

			return nil, fmt.Errorf("image access error: %s", string(body))
		}

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
