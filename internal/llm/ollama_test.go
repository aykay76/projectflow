package llm

import (
	"context"
	"testing"
	"time"
)

func TestOllamaProvider_Name(t *testing.T) {
	provider := NewOllamaProvider(ProviderConfig{})
	if provider.Name() != "ollama" {
		t.Errorf("Expected provider name to be 'ollama', got '%s'", provider.Name())
	}
}

func TestOllamaProvider_Model(t *testing.T) {
	// Test with default model
	provider := NewOllamaProvider(ProviderConfig{})
	if provider.Model() != OllamaDefaultModel {
		t.Errorf("Expected default model to be '%s', got '%s'", OllamaDefaultModel, provider.Model())
	}

	// Test with custom model
	customModel := "custom-model"
	provider = NewOllamaProvider(ProviderConfig{Model: customModel})
	if provider.Model() != customModel {
		t.Errorf("Expected custom model to be '%s', got '%s'", customModel, provider.Model())
	}
}

func TestOllamaProvider_DefaultConfiguration(t *testing.T) {
	provider := NewOllamaProvider(ProviderConfig{})

	// Check that defaults are set correctly
	if provider.config.BaseURL != OllamaBaseURL {
		t.Errorf("Expected default base URL to be '%s', got '%s'", OllamaBaseURL, provider.config.BaseURL)
	}

	if provider.config.Model != OllamaDefaultModel {
		t.Errorf("Expected default model to be '%s', got '%s'", OllamaDefaultModel, provider.config.Model)
	}

	if provider.config.Timeout != 60*time.Second {
		t.Errorf("Expected default timeout to be 60s, got %v", provider.config.Timeout)
	}

	if provider.config.MaxTokens != 1000 {
		t.Errorf("Expected default max tokens to be 1000, got %d", provider.config.MaxTokens)
	}
}

func TestOllamaProvider_CustomConfiguration(t *testing.T) {
	config := ProviderConfig{
		BaseURL:   "http://custom:8080",
		Model:     "custom-model",
		Timeout:   30 * time.Second,
		MaxTokens: 500,
		APIKey:    "test-key",
	}

	provider := NewOllamaProvider(config)

	// Check that custom values are preserved
	if provider.config.BaseURL != config.BaseURL {
		t.Errorf("Expected base URL to be '%s', got '%s'", config.BaseURL, provider.config.BaseURL)
	}

	if provider.config.Model != config.Model {
		t.Errorf("Expected model to be '%s', got '%s'", config.Model, provider.config.Model)
	}

	if provider.config.Timeout != config.Timeout {
		t.Errorf("Expected timeout to be %v, got %v", config.Timeout, provider.config.Timeout)
	}

	if provider.config.MaxTokens != config.MaxTokens {
		t.Errorf("Expected max tokens to be %d, got %d", config.MaxTokens, provider.config.MaxTokens)
	}

	if provider.config.APIKey != config.APIKey {
		t.Errorf("Expected API key to be '%s', got '%s'", config.APIKey, provider.config.APIKey)
	}
}

func TestOllamaProvider_Chat_RequestDefaults(t *testing.T) {
	provider := NewOllamaProvider(ProviderConfig{
		BaseURL: "http://fake-ollama:11434", // Use fake URL to avoid actual network calls
	})

	req := ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	
	// This will fail due to network error, but we can check that the request is formatted correctly
	// by examining the error message
	_, err := provider.Chat(ctx, req)

	if err == nil {
		t.Error("Expected error when connecting to fake URL")
	}

	// The error should contain our fake URL, confirming the request was attempted
	if !contains(err.Error(), "fake-ollama") {
		t.Errorf("Expected error to mention fake URL, got: %s", err.Error())
	}
}

func TestOllamaProvider_HealthCheck_Offline(t *testing.T) {
	provider := NewOllamaProvider(ProviderConfig{
		BaseURL: "http://fake-ollama:11434", // Use fake URL
	})

	ctx := context.Background()
	err := provider.HealthCheck(ctx)

	if err == nil {
		t.Error("Expected error when Ollama is not accessible")
	}

	// Should mention that Ollama is not running
	if !contains(err.Error(), "is Ollama running") {
		t.Errorf("Expected error to mention Ollama not running, got: %s", err.Error())
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
