package providers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"encore.app/src/config"
	"encore.app/src/models"
)

// GeminiProvider implements the Provider interface for Gemini API
type GeminiProvider struct{}

// NewGeminiProvider creates a new Gemini provider instance
func NewGeminiProvider(cfg *config.Config) *GeminiProvider {
	return &GeminiProvider{}
}

// GetName returns the provider name
func (g *GeminiProvider) GetName() string {
	return "gemini"
}

// downloadImageToBase64 downloads an image from HTTP/HTTPS URL and returns base64 encoded data with MIME type
func (g *GeminiProvider) downloadImageToBase64(url string) (string, string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP error %d downloading image", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read image data: %v", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // default fallback
	}

	// Validate content type
	if !strings.HasPrefix(contentType, "image/") {
		return "", "", fmt.Errorf("URL does not point to an image (Content-Type: %s)", contentType)
	}

	base64Data := base64.StdEncoding.EncodeToString(imageData)
	return base64Data, contentType, nil
}

// ChatCompletion calls the Gemini API for chat completion
func (g *GeminiProvider) ChatCompletion(req *models.ChatRequest, apiKey string) (*models.ChatResponse, error) {
	// Prepare the request payload for Gemini
	model := req.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}

	parts := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		for _, part := range msg.Content {
			if part.Type == "text" {
				parts = append(parts, map[string]interface{}{
					"text": part.Text,
				})
			} else if part.Type == "image_url" && part.ImageURL != nil {
				// Extract base64 data from the data URI
				dataURI := part.ImageURL.URL
				if strings.HasPrefix(dataURI, "data:image/jpeg;base64,") {
					base64Data := strings.TrimPrefix(dataURI, "data:image/jpeg;base64,")
					parts = append(parts, map[string]interface{}{
						"inline_data": map[string]string{
							"mime_type": "image/jpeg",
							"data":      base64Data,
						},
					})
				} else if strings.HasPrefix(dataURI, "data:image/png;base64,") {
					base64Data := strings.TrimPrefix(dataURI, "data:image/png;base64,")
					parts = append(parts, map[string]interface{}{
						"inline_data": map[string]string{
							"mime_type": "image/png",
							"data":      base64Data,
						},
					})
				} else if strings.HasPrefix(dataURI, "http://") || strings.HasPrefix(dataURI, "https://") {
					// Download image from HTTP/HTTPS URL and convert to base64
					base64Data, mimeType, err := g.downloadImageToBase64(dataURI)
					if err != nil {
						return nil, fmt.Errorf("failed to download image: %v", err)
					}

					parts = append(parts, map[string]interface{}{
						"inline_data": map[string]string{
							"mime_type": mimeType,
							"data":      base64Data,
						},
					})
				} else {
					return nil, fmt.Errorf("unsupported image data URI format: %s", dataURI)
				}
			}
		}
	}

	contents := []map[string]interface{}{
		{"parts": parts},
	}

	payload := map[string]interface{}{
		"contents": contents,
	}

	// Add generation config if temperature or max tokens are specified
	generationConfig := make(map[string]interface{})
	if req.Temperature != nil {
		generationConfig["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		generationConfig["maxOutputTokens"] = *req.MaxTokens
	}
	if len(generationConfig) > 0 {
		payload["generationConfig"] = generationConfig
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

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

	// Parse Gemini response
	var geminiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &geminiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to our response format
	response := &models.ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: make([]models.Choice, len(geminiResponse.Candidates)),
		Usage: models.Usage{
			PromptTokens:     geminiResponse.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResponse.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResponse.UsageMetadata.TotalTokenCount,
		},
	}

	for i, candidate := range geminiResponse.Candidates {
		content := ""
		if len(candidate.Content.Parts) > 0 {
			content = candidate.Content.Parts[0].Text
		}
		response.Choices[i] = models.Choice{
			Index: i,
			Message: models.ChatMessage{
				Role: "assistant",
				Content: []models.ContentPart{
					{
						Type: "text",
						Text: content,
					},
				},
			},
			FinishReason: strings.ToLower(candidate.FinishReason),
		}
	}

	return response, nil
}
