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
	OllamaBaseURL      = "http://localhost:11434"
	OllamaDefaultModel = "llama3.2"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	config ProviderConfig
	client *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config ProviderConfig) *OllamaProvider {
	if config.BaseURL == "" {
		config.BaseURL = OllamaBaseURL
	}
	if config.Model == "" {
		config.Model = OllamaDefaultModel
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second // Ollama can be slower than cloud APIs
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1000
	}

	return &OllamaProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (o *OllamaProvider) Name() string {
	return "ollama"
}

// Model returns the default model
func (o *OllamaProvider) Model() string {
	return o.config.Model
}

// Chat sends a chat request to Ollama
func (o *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Set default values if not provided
	if req.Model == "" {
		req.Model = o.config.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = o.config.MaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7 // Default temperature for good balance
	}

	// Create the HTTP request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", o.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	// Note: Ollama doesn't require an API key, but if one is configured, we'll use it
	if o.config.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.config.APIKey))
	}

	// Send the request
	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama at %s: %w", o.config.BaseURL, err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// HealthCheck verifies that Ollama is accessible and the model is available
func (o *OllamaProvider) HealthCheck(ctx context.Context) error {
	// First check if Ollama is running by hitting the API endpoint
	url := fmt.Sprintf("%s/api/tags", o.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama health check failed - is Ollama running at %s?: %w", o.config.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama health check failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Check if our configured model is available
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read models list: %w", err)
	}

	var modelsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return fmt.Errorf("failed to parse models list: %w", err)
	}

	// Check if our model is in the list
	modelFound := false
	for _, model := range modelsResp.Models {
		if model.Name == o.config.Model {
			modelFound = true
			break
		}
	}

	if !modelFound {
		availableModels := make([]string, len(modelsResp.Models))
		for i, model := range modelsResp.Models {
			availableModels[i] = model.Name
		}
		return fmt.Errorf("model '%s' not found in Ollama. Available models: %v. Run 'ollama pull %s' to download the model",
			o.config.Model, availableModels, o.config.Model)
	}

	// Finally, test a simple chat completion to ensure everything works
	req := ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Model:     o.config.Model,
		MaxTokens: 10,
	}

	_, err = o.Chat(ctx, req)
	if err != nil {
		return fmt.Errorf("ollama chat test failed: %w", err)
	}

	return nil
}
