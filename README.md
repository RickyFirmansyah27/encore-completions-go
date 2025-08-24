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

1. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your API keys
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run with Encore**:
   ```bash
   encore run
   ```

4. **Or run directly**:
   ```bash
   go run main.go
   ```

## Environment Variables

Configure the following environment variables in your `.env` file:

```env
GROQ_API_KEY=your_groq_api_key_here
OPENROUTER_API_KEY=your_openrouter_api_key_here
GEMINI_API_KEY=your_gemini_api_key_here
ATLASCLOUD_API_KEY=your_atlas_api_key_here
CHUTES_API_KEY=your_chutes_api_key_here
```

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