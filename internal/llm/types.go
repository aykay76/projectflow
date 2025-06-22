package llm

import (
	"context"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // Message content
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents a response from the LLM
type ChatResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Model   string                 `json:"model"`
	Choices []ChatResponseChoice   `json:"choices"`
	Usage   ChatResponseUsage      `json:"usage"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponseChoice represents a single choice in the response
type ChatResponseChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// ChatResponseUsage represents token usage information
type ChatResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Chat sends a chat request and returns the response
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	
	// HealthCheck verifies that the provider is available and configured correctly
	HealthCheck(ctx context.Context) error
	
	// Name returns the name of this provider
	Name() string
	
	// Model returns the default model name for this provider
	Model() string
}

// ProviderConfig holds common configuration for LLM providers
type ProviderConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	Timeout    time.Duration
	MaxTokens  int
}
