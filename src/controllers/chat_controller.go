package controllers

import (
	"context"
	"fmt"

	"encore.app/src/config"
	"encore.app/src/models"
	"encore.app/src/services"
)

// Service provides AI chat completion functionality
//
//encore:service
type Service struct {
	chatService *services.ChatService
}

// initService initializes the service with required dependencies
func initService() (*Service, error) {
	cfg := config.LoadConfig()
	chatService := services.NewChatService(cfg)

	return &Service{
		chatService: chatService,
	}, nil
}

// ChatCompletion handles chat completion requests
//
//encore:api public method=POST path=/chat/completions
func (s *Service) ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return s.chatService.ProcessChatCompletion(req)
}

// HealthCheck returns the health status of the service
//
//encore:api public method=GET path=/health
func (s *Service) HealthCheck(ctx context.Context) (*models.HealthResponse, error) {
	response := s.chatService.GetHealthStatus()
	return response, nil
}

// GetProviders returns the list of supported AI providers
//
//encore:api public method=GET path=/providers
func (s *Service) GetProviders(ctx context.Context) (*models.ProvidersResponse, error) {
	response := s.chatService.GetSupportedProviders()
	return response, nil
}

// TestProvider tests if a specific provider is working
//
//encore:api public method=POST path=/providers/test
func (s *Service) TestProvider(ctx context.Context, req *models.TestProviderRequest) (*models.TestProviderResponse, error) {
	if req.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	response := s.chatService.TestProvider(req)
	return response, nil
}
