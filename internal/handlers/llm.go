package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aykay76/projectflow/internal/llm"
)

// LLMHandler handles direct LLM-related HTTP requests
type LLMHandler struct {
	llmService LLMService
	logger     *slog.Logger
}

// LLMInfoResponse represents LLM service information
type LLMInfoResponse struct {
	Enabled   bool                   `json:"enabled"`
	Provider  string                 `json:"provider"`
	Model     string                 `json:"model"`
	Status    string                 `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// LLMChatRequest represents a direct LLM chat request
type LLMChatRequest struct {
	Messages    []llm.Message `json:"messages"`
	Model       string        `json:"model,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// LLMChatResponse represents a direct LLM chat response
type LLMChatResponse struct {
	Response  *llm.ChatResponse `json:"response"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewLLMHandler creates a new LLM handler instance
func NewLLMHandler(llmService LLMService, logger *slog.Logger) *LLMHandler {
	return &LLMHandler{
		llmService: llmService,
		logger:     logger,
	}
}

// HandleLLMInfo handles GET /api/llm/info requests
func (lh *LLMHandler) HandleLLMInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := LLMInfoResponse{
		Enabled:   lh.llmService.IsEnabled(),
		Timestamp: time.Now().UTC(),
	}

	if lh.llmService.IsEnabled() {
		// Get provider info if LLM is enabled
		if providerInfo, ok := lh.llmService.(interface{ GetProviderInfo() map[string]interface{} }); ok {
			providerInfoData := providerInfo.GetProviderInfo()
			if provider, exists := providerInfoData["provider"]; exists {
				if providerStr, ok := provider.(string); ok {
					info.Provider = providerStr
				}
			}
			if model, exists := providerInfoData["model"]; exists {
				if modelStr, ok := model.(string); ok {
					info.Model = modelStr
				}
			}
			info.Metadata = providerInfoData
		}

		// Perform health check to determine status
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := lh.llmService.HealthCheck(ctx); err != nil {
			info.Status = "unhealthy"
			if info.Metadata == nil {
				info.Metadata = make(map[string]interface{})
			}
			info.Metadata["health_error"] = err.Error()
			lh.logger.Warn("LLM health check failed", "error", err)
		} else {
			info.Status = "healthy"
		}
	} else {
		info.Status = "disabled"
		info.Provider = "none"
	}

	lh.respondWithJSON(w, &info)
}

// HandleLLMChat handles POST /api/llm/chat requests (direct LLM interaction)
func (lh *LLMHandler) HandleLLMChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !lh.llmService.IsEnabled() {
		http.Error(w, "LLM service is disabled", http.StatusServiceUnavailable)
		return
	}

	// Parse request
	var req LLMChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lh.logger.Error("Failed to decode LLM chat request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Messages cannot be empty", http.StatusBadRequest)
		return
	}

	lh.logger.Info("Processing direct LLM chat request",
		"messages_count", len(req.Messages),
		"model", req.Model)

	// Create LLM request
	llmReq := llm.ChatRequest{
		Messages:    req.Messages,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      req.Stream,
	}

	// Call LLM service
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Extended timeout for LLM
	defer cancel()

	response, err := lh.llmService.Chat(ctx, llmReq)
	if err != nil {
		lh.logger.Error("LLM chat request failed", "error", err)
		http.Error(w, fmt.Sprintf("LLM request failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	chatResponse := LLMChatResponse{
		Response:  response,
		Timestamp: time.Now().UTC(),
	}

	lh.logger.Info("LLM chat request completed",
		"choices_count", len(response.Choices),
		"total_tokens", response.Usage.TotalTokens)

	lh.respondWithJSON(w, &chatResponse)
}

// HandleLLMHealthCheck handles GET /api/llm/health requests
func (lh *LLMHandler) HandleLLMHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !lh.llmService.IsEnabled() {
		response := map[string]interface{}{
			"status":    "disabled",
			"message":   "LLM service is disabled",
			"timestamp": time.Now().UTC(),
		}
		lh.respondWithJSON(w, response)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := lh.llmService.HealthCheck(ctx); err != nil {
		lh.logger.Error("LLM health check failed", "error", err)
		response := map[string]interface{}{
			"status":    "unhealthy",
			"message":   err.Error(),
			"timestamp": time.Now().UTC(),
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		lh.respondWithJSON(w, response)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"message":   "LLM service is operational",
		"timestamp": time.Now().UTC(),
	}
	lh.respondWithJSON(w, response)
}

// respondWithJSON writes a JSON response to the HTTP response writer
func (lh *LLMHandler) respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		lh.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
