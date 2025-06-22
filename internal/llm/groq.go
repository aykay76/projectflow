package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	GroqBaseURL      = "https://api.groq.com/openai/v1"
	GroqDefaultModel = "llama-3.1-8b-instant"
)

// GroqProvider implements the Provider interface for Groq AI
type GroqProvider struct {
	config ProviderConfig
	client *http.Client
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider(config ProviderConfig) *GroqProvider {
	if config.BaseURL == "" {
		config.BaseURL = GroqBaseURL
	}
	if config.Model == "" {
		config.Model = GroqDefaultModel
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1000
	}

	return &GroqProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (g *GroqProvider) Name() string {
	return "groq"
}

// Model returns the default model
func (g *GroqProvider) Model() string {
	return g.config.Model
}

// Chat sends a chat request to Groq
func (g *GroqProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if g.config.APIKey == "" {
		return nil, fmt.Errorf("groq API key not configured")
	}

	// Set default values if not provided
	if req.Model == "" {
		req.Model = g.config.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = g.config.MaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7 // Default temperature for good balance
	}

	// Create the HTTP request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", g.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.config.APIKey))

	// Send the request
	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// HealthCheck verifies that Groq is accessible and configured correctly
func (g *GroqProvider) HealthCheck(ctx context.Context) error {
	if g.config.APIKey == "" {
		return fmt.Errorf("groq API key not configured")
	}

	// Send a simple test message
	req := ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Model:     g.config.Model,
		MaxTokens: 10,
	}

	resp, err := g.Chat(ctx, req)
	if err != nil {
		return fmt.Errorf("groq health check failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("groq health check failed: no response choices")
	}

	return nil
}
