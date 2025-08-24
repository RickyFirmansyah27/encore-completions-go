package config

// Secrets defined for the application
var secrets struct {
	GroqAPIKey       string // API key for Groq service
	OpenRouterAPIKey string // API key for OpenRouter service
	GeminiAPIKey     string // API key for Gemini service
	AtlasAPIKey      string // API key for Atlas service
	ChutesAPIKey     string // API key for Chutes service
}

// Config holds application configuration
type Config struct{}

// LoadConfig creates a new configuration instance
func LoadConfig() *Config {
	return &Config{}
}

// GetAPIKey returns the API key for the specified provider
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
		return ""
	}
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
