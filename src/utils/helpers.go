package utils

import (
	"fmt"
	"time"
)

// GetModel returns the model name, using default if not provided
func GetModel(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	return "llama3-8b-8192" // default
}

// GenerateID generates a simple ID based on current timestamp
func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// CountTokens provides a simple approximation of token count
func CountTokens(text string) int {
	// Simple approximation: 1 token per 4 characters
	return len(text) / 4
}
