package llm

import (
	"log/slog"
	"os"
	"testing"

	"github.com/aykay76/projectflow/internal/config"
)

func TestFactory_CreateProvider_Groq(t *testing.T) {
	cfg := &config.Config{
		LLMProvider:  "groq",
		LLMAPIKey:    "test-key",
		LLMModel:     "test-model",
		LLMTimeout:   30,
		LLMMaxTokens: 1000,
	}

	factory := NewFactory(cfg)
	provider, err := factory.CreateProvider()

	if err != nil {
		t.Fatalf("Expected no error, got: %s", err)
	}

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}

	if provider.Name() != "groq" {
		t.Errorf("Expected provider name 'groq', got '%s'", provider.Name())
	}

	if provider.Model() != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", provider.Model())
	}
}

func TestFactory_CreateProvider_Disabled(t *testing.T) {
	cfg := &config.Config{
		LLMProvider: "disabled",
	}

	factory := NewFactory(cfg)
	_, err := factory.CreateProvider()

	if err == nil {
		t.Error("Expected error when LLM is disabled")
	}

	expectedError := "LLM is disabled in configuration"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestFactory_CreateProvider_Unsupported(t *testing.T) {
	cfg := &config.Config{
		LLMProvider: "unsupported-provider",
		LLMAPIKey:   "test-key",
	}

	factory := NewFactory(cfg)
	_, err := factory.CreateProvider()

	if err == nil {
		t.Error("Expected error for unsupported provider")
	}

	expectedError := "unsupported LLM provider: unsupported-provider"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestFactory_GetSupportedProviders(t *testing.T) {
	cfg := &config.Config{}
	factory := NewFactory(cfg)

	supported := factory.GetSupportedProviders()

	if len(supported) == 0 {
		t.Error("Expected at least one supported provider")
	}

	// Check that groq is supported
	groqSupported := false
	for _, provider := range supported {
		if provider == "groq" {
			groqSupported = true
			break
		}
	}

	if !groqSupported {
		t.Error("Expected 'groq' to be in supported providers")
	}
}

func TestService_NewService_Disabled(t *testing.T) {
	cfg := &config.Config{
		LLMProvider: "disabled",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service, err := NewService(cfg, logger)
	if err != nil {
		t.Fatalf("Expected no error, got: %s", err)
	}

	if service.IsEnabled() {
		t.Error("Expected service to be disabled")
	}
}

func TestService_NewService_InvalidProvider(t *testing.T) {
	cfg := &config.Config{
		LLMProvider: "invalid-provider",
		LLMAPIKey:   "test-key",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	_, err := NewService(cfg, logger)
	if err == nil {
		t.Error("Expected error for invalid provider")
	}
}
