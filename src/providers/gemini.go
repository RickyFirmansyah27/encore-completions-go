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
		// Use a model that supports vision and has a larger context window
		model = "gemini-2.5-flash"
	}

	// Validate and build contents for the Gemini API
	parts := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		for _, part := range msg.Content {
			if part.Type == "text" {
				parts = append(parts, map[string]interface{}{
					"text": part.Text,
				})
			} else if part.Type == "image_url" && part.ImageURL != nil {
				dataURI := part.ImageURL.URL
				base64Data := ""
				mimeType := ""
				var err error

				if strings.HasPrefix(dataURI, "data:image/") {
					// Extract base64 data and mime type from the data URI
					partsURI := strings.SplitN(dataURI, ",", 2)
					if len(partsURI) != 2 {
						return nil, fmt.Errorf("invalid image data URI format")
					}
					header := partsURI[0]
					base64Data = partsURI[1]

					mimeType = strings.TrimPrefix(header, "data:")
					mimeType = strings.TrimSuffix(mimeType, ";base64")

				} else if strings.HasPrefix(dataURI, "http://") || strings.HasPrefix(dataURI, "https://") {
					// Download image from HTTP/HTTPS URL and convert to base64
					base64Data, mimeType, err = g.downloadImageToBase64(dataURI)
					if err != nil {
						return nil, fmt.Errorf("failed to download image: %v", err)
					}
				} else {
					return nil, fmt.Errorf("unsupported image URL format: %s", dataURI)
				}

				// Append inline data part if base64Data is not empty
				if base64Data != "" && mimeType != "" {
					parts = append(parts, map[string]interface{}{
						"inline_data": map[string]string{
							"mime_type": mimeType,
							"data":      base64Data,
						},
					})
				}
			}
		}
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("no valid content found in the request messages")
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
	} else {
		// Set a reasonable default if not specified to avoid early termination
		generationConfig["maxOutputTokens"] = 2048
	}
	if len(generationConfig) > 0 {
		payload["generationConfig"] = generationConfig
	}

	// Add safety settings to get feedback
	safetySettings := []map[string]interface{}{
		{
			"category":  "HARM_CATEGORY_DANGEROUS_CONTENT",
			"threshold": "BLOCK_NONE",
		},
		{
			"category":  "HARM_CATEGORY_HATE_SPEECH",
			"threshold": "BLOCK_NONE",
		},
		{
			"category":  "HARM_CATEGORY_HARASSMENT",
			"threshold": "BLOCK_NONE",
		},
		{
			"category":  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
			"threshold": "BLOCK_NONE",
		},
	}
	payload["safetySettings"] = safetySettings

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
		PromptFeedback struct {
			SafetyRatings []struct {
				Category    string `json:"category"`
				Probability string `json:"probability"`
				Blocked     bool   `json:"blocked"`
			} `json:"safetyRatings"`
		} `json:"promptFeedback"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &geminiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Check if the prompt was blocked by safety settings
	if geminiResponse.PromptFeedback.SafetyRatings != nil {
		for _, rating := range geminiResponse.PromptFeedback.SafetyRatings {
			if rating.Blocked {
				return nil, fmt.Errorf("prompt was blocked due to safety rating: category=%s, probability=%s", rating.Category, rating.Probability)
			}
		}
	}

	// Handle case where no candidates are returned
	if len(geminiResponse.Candidates) == 0 {
		return nil, fmt.Errorf("gemini API returned no candidates, possibly due to safety filters or lack of generated content")
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
		if len(candidate.Content.Parts) > 0 && candidate.Content.Parts[0].Text != "" {
			content = candidate.Content.Parts[0].Text
		}

		finishReason := strings.ToLower(candidate.FinishReason)
		if finishReason == "max_tokens" && len(candidate.Content.Parts) == 0 {
			content = "The response was terminated early due to the 'max_tokens' limit. Please try increasing the max_tokens parameter."
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
			FinishReason: finishReason,
		}
	}

	return response, nil
}
