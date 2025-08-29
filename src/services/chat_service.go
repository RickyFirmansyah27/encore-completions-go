package services

import (
	"fmt"
	"time"

	"encore.app/src/config"
	"encore.app/src/models"
	"encore.app/src/providers"
)

// Constants for default values and status messages
const (
	DefaultProvider       = "groq"
	DefaultTemperature    = 0.5
	DefaultMaxTokens      = 1000
	StatusHealthy         = "healthy"
	StatusNoAPIKeys       = "no_api_keys"
	StatusInvalidRequest  = "invalid_request"
	StatusInvalidProvider = "invalid_provider"
)

// ChatService handles chat completion business logic
type ChatService struct {
	config *config.Config
}

// NewChatService creates a new chat service instance
func NewChatService(cfg *config.Config) *ChatService {
	providers.InitProviders(cfg)
	return &ChatService{
		config: cfg,
	}
}

// setDefaults applies default values to the request if not provided
func setDefaults(req *models.ChatRequest) {
	if req.Temperature == nil {
		defaultTemp := DefaultTemperature
		req.Temperature = &defaultTemp
	}
	if req.MaxTokens == nil {
		defaultMaxTokens := DefaultMaxTokens
		req.MaxTokens = &defaultMaxTokens
	}
}

// getProviderName returns the provider name with default fallback
func getProviderName(providerName string) string {
	if providerName == "" {
		return DefaultProvider
	}
	return providerName
}

// ProcessChatCompletion processes a chat completion request
func (cs *ChatService) ProcessChatCompletion(req *models.ChatRequest) (*models.ChatResponse, error) {
	if req.Prompt == "" && !req.WithImage {
		return nil, fmt.Errorf("invalid request: prompt cannot be empty if not sending an image")
	}

	// Apply default values
	setDefaults(req)

	// Construct ContentParts
	var contentParts []models.ContentPart
	if req.Prompt != "" {
		contentParts = append(contentParts, models.ContentPart{
			Type: "text",
			Text: req.Prompt,
		})
	}
	if req.WithImage && req.ImageData != "" {
		contentParts = append(contentParts, models.ContentPart{
			Type: "image_url",
			ImageURL: &models.ImageURL{
				URL: req.ImageData,
			},
		})
	}

	// Create ChatMessage
	chatMessage := models.ChatMessage{
		Role:    "user",
		Content: contentParts,
	}

	// Assign the constructed message to the original request's Messages field
	req.Messages = []models.ChatMessage{chatMessage}

	// Get provider name with default
	providerName := getProviderName(req.Provider)

	// Get API key
	apiKey := cs.config.GetAPIKey(providerName)
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found for provider: %s", providerName)
	}

	// Get provider instance
	provider, err := providers.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Call the provider
	return provider.ChatCompletion(req, apiKey)
}

// GetHealthStatus returns the health status of the service
func (cs *ChatService) GetHealthStatus() *models.HealthResponse {
	// Check if at least one API key is available
	chatStatus := StatusHealthy
	hasAnyKey := false

	for _, provider := range cs.config.GetSupportedProviders() {
		if cs.config.GetAPIKey(provider) != "" {
			hasAnyKey = true
			break
		}
	}

	if !hasAnyKey {
		chatStatus = StatusNoAPIKeys
	}

	return &models.HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now().Format(time.RFC3339),
		Services: map[string]string{
			"chat": chatStatus,
		},
	}
}

// GetSupportedProviders returns the list of supported providers
func (cs *ChatService) GetSupportedProviders() *models.ProvidersResponse {
	return &models.ProvidersResponse{
		Providers: cs.config.GetSupportedProviders(),
	}
}

// TestProvider tests if a specific provider is working
func (cs *ChatService) TestProvider(req *models.TestProviderRequest) *models.TestProviderResponse {
	if req.Provider == "" {
		return &models.TestProviderResponse{
			Provider: req.Provider,
			Status:   StatusInvalidRequest,
		}
	}

	// Validate provider name first
	if !cs.config.IsValidProvider(req.Provider) {
		return &models.TestProviderResponse{
			Provider: req.Provider,
			Status:   StatusInvalidProvider,
		}
	}

	// Check if API key is available
	apiKey := cs.config.GetAPIKey(req.Provider)
	if apiKey == "" {
		return &models.TestProviderResponse{
			Provider: req.Provider,
			Status:   StatusNoAPIKeys,
		}
	}

	return &models.TestProviderResponse{
		Provider: req.Provider,
		Status:   StatusHealthy,
	}
}
