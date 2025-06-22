package llm

import (
	"context"
	"testing"
	"time"
)

func TestGroqProvider_Name(t *testing.T) {
	provider := NewGroqProvider(ProviderConfig{})
	if provider.Name() != "groq" {
		t.Errorf("Expected provider name to be 'groq', got '%s'", provider.Name())
	}
}

func TestGroqProvider_Model(t *testing.T) {
	// Test with default model
	provider := NewGroqProvider(ProviderConfig{})
	if provider.Model() != GroqDefaultModel {
		t.Errorf("Expected default model to be '%s', got '%s'", GroqDefaultModel, provider.Model())
	}

	// Test with custom model
	customModel := "custom-model"
	provider = NewGroqProvider(ProviderConfig{Model: customModel})
	if provider.Model() != customModel {
		t.Errorf("Expected custom model to be '%s', got '%s'", customModel, provider.Model())
	}
}

func TestGroqProvider_Chat_NoAPIKey(t *testing.T) {
	provider := NewGroqProvider(ProviderConfig{})

	req := ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := provider.Chat(ctx, req)

	if err == nil {
		t.Error("Expected error when API key is not configured")
	}

	if err.Error() != "groq API key not configured" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}
}

func TestGroqProvider_HealthCheck_NoAPIKey(t *testing.T) {
	provider := NewGroqProvider(ProviderConfig{})

	ctx := context.Background()
	err := provider.HealthCheck(ctx)

	if err == nil {
		t.Error("Expected error when API key is not configured")
	}
}

func TestGroqProvider_DefaultValues(t *testing.T) {
	provider := NewGroqProvider(ProviderConfig{})

	// Check that defaults are set correctly
	if provider.config.BaseURL != GroqBaseURL {
		t.Errorf("Expected default base URL to be '%s', got '%s'", GroqBaseURL, provider.config.BaseURL)
	}

	if provider.config.Model != GroqDefaultModel {
		t.Errorf("Expected default model to be '%s', got '%s'", GroqDefaultModel, provider.config.Model)
	}

	if provider.config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", provider.config.Timeout)
	}

	if provider.config.MaxTokens != 1000 {
		t.Errorf("Expected default max tokens to be 1000, got %d", provider.config.MaxTokens)
	}
}

func TestChatRequest_DefaultValues(t *testing.T) {
	// Test that ChatRequest can be created with minimal required fields
	req := ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// Verify the basic structure
	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", req.Messages[0].Role)
	}

	if req.Messages[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", req.Messages[0].Content)
	}
}
