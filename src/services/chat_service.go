package services

import (
	"fmt"
	"time"

	"encore.app/src/config"
	"encore.app/src/models"
	"encore.app/src/providers"
)

// ChatService handles chat completion business logic
type ChatService struct {
	config *config.Config
}

// NewChatService creates a new chat service instance
func NewChatService(cfg *config.Config) *ChatService {
	return &ChatService{
		config: cfg,
	}
}

// ProcessChatCompletion processes a chat completion request
func (cs *ChatService) ProcessChatCompletion(req *models.ChatRequest) (*models.ChatResponse, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Get provider (default to groq if not specified)
	providerName := req.Provider
	if providerName == "" {
		providerName = "groq"
	}

	// Get API key
	apiKey := cs.config.GetAPIKey(providerName)
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found for provider: %s", providerName)
	}

	// Get provider instance
	provider := providers.GetProvider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Call the provider
	return provider.ChatCompletion(req, apiKey)
}

// GetHealthStatus returns the health status of the service
func (cs *ChatService) GetHealthStatus() *models.HealthResponse {
	// Check if at least one API key is available
	chatStatus := "healthy"
	hasAnyKey := false

	for _, provider := range cs.config.GetSupportedProviders() {
		if cs.config.GetAPIKey(provider) != "" {
			hasAnyKey = true
			break
		}
	}

	if !hasAnyKey {
		chatStatus = "no_api_keys"
	}

	return &models.HealthResponse{
		Status:    "healthy",
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
			Status:   "invalid_request",
		}
	}

	// Check if API key is available
	apiKey := cs.config.GetAPIKey(req.Provider)
	if apiKey == "" {
		return &models.TestProviderResponse{
			Provider: req.Provider,
			Status:   "no_api_key",
		}
	}

	// Validate provider name
	if !cs.config.IsValidProvider(req.Provider) {
		return &models.TestProviderResponse{
			Provider: req.Provider,
			Status:   "invalid_provider",
		}
	}

	return &models.TestProviderResponse{
		Provider: req.Provider,
		Status:   "healthy",
	}
}
