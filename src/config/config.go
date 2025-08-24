package config

import (
	"bufio"
	"os"
	"strings"
)

// Config holds application configuration
type Config struct {
	APIKeys map[string]string
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() *Config {
	loadEnvFile()

	return &Config{
		APIKeys: map[string]string{
			"groq":       os.Getenv("GROQ_API_KEY"),
			"openrouter": os.Getenv("OPENROUTER_API_KEY"),
			"gemini":     os.Getenv("GEMINI_API_KEY"),
			"atlas":      os.Getenv("ATLASCLOUD_API_KEY"),
			"chutes":     os.Getenv("CHUTES_API_KEY"),
		},
	}
}

// GetAPIKey returns the API key for the specified provider
func (c *Config) GetAPIKey(provider string) string {
	return c.APIKeys[provider]
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

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		// .env file doesn't exist, that's okay
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
}
