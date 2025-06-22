package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aykay76/projectflow/internal/config"
)

// Service manages LLM providers and operations
type Service struct {
	provider Provider
	config   *config.Config
	logger   *slog.Logger
}

// NewService creates a new LLM service
func NewService(cfg *config.Config, logger *slog.Logger) (*Service, error) {
	if !cfg.IsLLMEnabled() {
		return &Service{
			provider: nil,
			config:   cfg,
			logger:   logger,
		}, nil
	}

	factory := NewFactory(cfg)
	provider, err := factory.CreateProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	service := &Service{
		provider: provider,
		config:   cfg,
		logger:   logger,
	}

	// Perform health check
	ctx := context.Background()
	if err := service.HealthCheck(ctx); err != nil {
		logger.Warn("LLM provider health check failed", "error", err)
		// Don't fail service creation, just log the warning
	} else {
		logger.Info("LLM service initialized successfully",
			"provider", provider.Name(),
			"model", provider.Model())
	}

	return service, nil
}

// IsEnabled returns true if LLM functionality is enabled
func (s *Service) IsEnabled() bool {
	return s.provider != nil
}

// Chat sends a chat request to the LLM provider
func (s *Service) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("LLM service is disabled")
	}

	s.logger.Debug("Sending chat request to LLM",
		"provider", s.provider.Name(),
		"model", req.Model,
		"messages_count", len(req.Messages))

	resp, err := s.provider.Chat(ctx, req)
	if err != nil {
		s.logger.Error("LLM chat request failed",
			"provider", s.provider.Name(),
			"error", err)
		return nil, err
	}

	s.logger.Debug("LLM chat request completed",
		"provider", s.provider.Name(),
		"choices_count", len(resp.Choices),
		"total_tokens", resp.Usage.TotalTokens)

	return resp, nil
}

// HealthCheck verifies that the LLM provider is working
func (s *Service) HealthCheck(ctx context.Context) error {
	if !s.IsEnabled() {
		return fmt.Errorf("LLM service is disabled")
	}

	return s.provider.HealthCheck(ctx)
}

// GetProviderInfo returns information about the current provider
func (s *Service) GetProviderInfo() map[string]interface{} {
	if !s.IsEnabled() {
		return map[string]interface{}{
			"enabled":  false,
			"provider": "disabled",
		}
	}

	return map[string]interface{}{
		"enabled":  true,
		"provider": s.provider.Name(),
		"model":    s.provider.Model(),
	}
}
