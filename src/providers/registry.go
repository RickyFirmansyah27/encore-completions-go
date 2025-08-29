package providers

import (
	"fmt"
	"sync"

	"encore.app/src/config"
	"encore.app/src/models"
)

// Provider interface defines the methods that each AI provider must implement.
type Provider interface {
	GetName() string
	ChatCompletion(req *models.ChatRequest, apiKey string) (*models.ChatResponse, error)
}

var (
	providersMu sync.RWMutex
	providers   = make(map[string]Provider)
)

// RegisterProvider registers a new provider.
func RegisterProvider(name string, provider Provider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers[name] = provider
}

// GetProvider retrieves a registered provider by name.
func GetProvider(name string) (Provider, error) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	p, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return p, nil
}

// InitProviders initializes and registers all available providers.
func InitProviders(cfg *config.Config) {
	RegisterProvider("gemini", NewGeminiProvider(cfg))
	RegisterProvider("openrouter", NewOpenRouterProvider(cfg))
	RegisterProvider("groq", NewGroqProvider(cfg))
	RegisterProvider("atlas", NewAtlasProvider(cfg))
	RegisterProvider("chutes", NewChutesProvider(cfg))
}
