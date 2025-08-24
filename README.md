# Encore AI Chat Completion Service (Go)

A Golang implementation of an AI chat completion service using the Encore framework, providing a unified interface for multiple AI providers.

## Project Structure

```
encore-completion-go/
├── main.go                 # Application entry point
├── encore.app             # Encore application configuration
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── .env                   # Environment variables (not in git)
├── .env.example           # Environment variables template
└── src/                   # Source code
    ├── chat.go            # Main service definition
    ├── config/            # Configuration management
    │   └── config.go      # Environment and API key configuration
    ├── controllers/       # HTTP request handlers
    │   └── chat_controller.go  # Chat API endpoints
    ├── models/            # Data structures
    │   └── chat.go        # Request/response models
    ├── providers/         # AI provider implementations
    │   └── groq.go        # Groq API provider
    ├── services/          # Business logic
    │   └── chat_service.go    # Chat completion service
    └── utils/             # Utility functions
        └── helpers.go     # Common helper functions
```

## Features

- **Multiple AI Provider Support**: Currently supports Groq (extensible for OpenRouter, Gemini, Atlas, Chutes)
- **Unified API Interface**: OpenAI-compatible API endpoints
- **Health Monitoring**: Built-in health checks and provider testing
- **Clean Architecture**: Separated concerns with controllers, services, models, and providers
- **Environment Configuration**: Secure API key management through environment variables

## API Endpoints

- `POST /chat/completions` - Chat completion requests
- `GET /health` - Service health check
- `GET /providers` - List supported providers
- `POST /providers/test` - Test specific provider

## Getting Started

### 1. Set up Encore secrets for API keys

This project uses Encore's secure secrets management. You have several options:

**Option A: Using Encore Cloud Dashboard**
1. Go to your app in [Encore Cloud Dashboard](https://app.encore.cloud)
2. Navigate to Settings > Secrets
3. Add the following secrets:
   - `GroqAPIKey`
   - `OpenRouterAPIKey` 
   - `GeminiAPIKey`
   - `AtlasAPIKey`
   - `ChutesAPIKey`

**Option B: Using Encore CLI**
```bash
# Set secrets for local/development environment
encore secret set --type local,dev GroqAPIKey
encore secret set --type local,dev OpenRouterAPIKey
encore secret set --type local,dev GeminiAPIKey
encore secret set --type local,dev AtlasAPIKey
encore secret set --type local,dev ChutesAPIKey

# Set secrets for production
encore secret set --type prod GroqAPIKey
# ... repeat for other secrets
```

**Option C: Local development override**
For local development, copy and edit the secrets file:
```bash
cp .secrets.local.cue.example .secrets.local.cue
# Edit .secrets.local.cue with your actual API keys
```

### 2. Install dependencies and run

```bash
go mod tidy
encore run
```

## Secrets Configuration

This project uses Encore's built-in secrets management for secure API key storage. Secrets are defined in the code and managed through Encore's infrastructure.

### Required Secrets

- `GroqAPIKey` - API key for Groq service
- `OpenRouterAPIKey` - API key for OpenRouter service  
- `GeminiAPIKey` - API key for Gemini service
- `AtlasAPIKey` - API key for Atlas service
- `ChutesAPIKey` - API key for Chutes service

### Security Benefits

- ✅ Secrets are encrypted using Google Cloud KMS
- ✅ Never stored in plain text in your code
- ✅ Environment-specific values (dev, staging, prod)
- ✅ Automatic synchronization across team members
- ✅ No risk of accidentally committing secrets

## Architecture

This project follows a clean architecture pattern with clear separation of concerns:

- **Controllers**: Handle HTTP requests and responses
- **Services**: Contain business logic
- **Models**: Define data structures
- **Providers**: Implement AI provider interfaces
- **Config**: Manage configuration and environment variables
- **Utils**: Provide common utility functions

## Adding New Providers

To add a new AI provider:

1. Implement the `Provider` interface in `src/providers/`
2. Add the provider to the `GetProvider` function
3. Update configuration to include new API keys
4. Add provider name to supported providers list

## Testing

Run tests with:
```bash
go test ./...
```

## Deployment

This service can be deployed using Encore's deployment features or as a standard Go application.