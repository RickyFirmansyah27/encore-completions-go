package config

// Secrets defined for the application
// All fields default to empty strings if not set via Encore secrets
var secrets struct {
	GroqAPIKey       string // API key for Groq service (defaults to "")
	OpenRouterAPIKey string // API key for OpenRouter service (defaults to "")
	GeminiAPIKey     string // API key for Gemini service (defaults to "")
	AtlasAPIKey      string // API key for Atlas service (defaults to "")
	ChutesAPIKey     string // API key for Chutes service (defaults to "")
}

// Config holds application configuration
type Config struct{}

// LoadConfig creates a new configuration instance
func LoadConfig() *Config {
	return &Config{}
}

// GetAPIKey returns the API key for the specified provider
// Returns empty string if provider not found or API key not set
func (c *Config) GetAPIKey(provider string) string {
	switch provider {
	case "groq":
		return secrets.GroqAPIKey
	case "openrouter":
		return secrets.OpenRouterAPIKey
	case "gemini":
		return secrets.GeminiAPIKey
	case "atlas":
		return secrets.AtlasAPIKey
	case "chutes":
		return secrets.ChutesAPIKey
	default:
		return "" // Empty string for unknown providers
	}
}

// HasAPIKey checks if a provider has a non-empty API key configured
func (c *Config) HasAPIKey(provider string) bool {
	return c.GetAPIKey(provider) != ""
}

// GetSupportedProviders returns list of supported providers
func (c *Config) GetSupportedProviders() []string {
	return []string{"groq", "openrouter", "gemini", "atlas", "chutes"}
}

// IsValidProvider checks if a provider is supported
func (c *Config) IsValidProvider(provider string) bool {
	for _, p := range c.GetSupportedProviders() {
		if p == provider {
			return true
		}
	}
	return false
}
