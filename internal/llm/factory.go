package llm

import (
	"fmt"
	"time"

	"github.com/aykay76/projectflow/internal/config"
)

// Factory creates LLM providers based on configuration
type Factory struct {
	config *config.Config
}

// NewFactory creates a new provider factory
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateProvider creates a provider based on the configuration
func (f *Factory) CreateProvider() (Provider, error) {
	if !f.config.IsLLMEnabled() {
		return nil, fmt.Errorf("LLM is disabled in configuration")
	}

	providerConfig := ProviderConfig{
		APIKey:    f.config.LLMAPIKey,
		BaseURL:   f.config.LLMBaseURL,
		Model:     f.config.LLMModel,
		Timeout:   time.Duration(f.config.LLMTimeout) * time.Second,
		MaxTokens: f.config.LLMMaxTokens,
	}

	switch f.config.GetLLMProvider() {
	case "groq":
		return NewGroqProvider(providerConfig), nil
	case "ollama":
		return NewOllamaProvider(providerConfig), nil
	case "openai":
		// TODO: Implement OpenAIProvider
		return nil, fmt.Errorf("openai provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", f.config.LLMProvider)
	}
}

// GetSupportedProviders returns a list of supported provider names
func (f *Factory) GetSupportedProviders() []string {
	return []string{
		"groq",
		"ollama",
		// "openai",    // TODO: Uncomment when implemented
	}
}
