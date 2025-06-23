package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aykay76/projectflow/internal/storage"
)

// HealthResponse represents the structure of health check responses
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
	Checks    []Check   `json:"checks,omitempty"`
}

// Check represents an individual health check
type Check struct {
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`
	Error    string                 `json:"error,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Duration string                 `json:"duration,omitempty"`
}

// LLMService interface for health checking
type LLMService interface {
	HealthCheck(ctx context.Context) error
	IsEnabled() bool
	GetProviderInfo() map[string]interface{}
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	storage    storage.Storage
	llmService LLMService
	version    string
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(storage storage.Storage, version string) *HealthChecker {
	return &HealthChecker{
		storage:    storage,
		llmService: nil,
		version:    version,
	}
}

// SetLLMService sets the LLM service for health checking
func (h *HealthChecker) SetLLMService(llmService LLMService) {
	h.llmService = llmService
}

// HandleHealth provides basic health status
func (h *HealthChecker) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   h.version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleReady provides readiness status with dependency checks
func (h *HealthChecker) HandleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checks := []Check{}
	overallStatus := "ready"
	statusCode := http.StatusOK

	// Check storage connectivity
	storageCheck := h.checkStorage()
	checks = append(checks, storageCheck)

	if storageCheck.Status != "healthy" {
		overallStatus = "not ready"
		statusCode = http.StatusServiceUnavailable
	}

	// Check LLM service if enabled
	if h.llmService != nil && h.llmService.IsEnabled() {
		llmCheck := h.checkLLM()
		checks = append(checks, llmCheck)

		if llmCheck.Status != "healthy" {
			overallStatus = "not ready"
			statusCode = http.StatusServiceUnavailable
		}
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Version:   h.version,
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// checkStorage verifies storage connectivity
func (h *HealthChecker) checkStorage() Check {
	start := time.Now()
	check := Check{
		Name:    "storage",
		Details: make(map[string]interface{}),
	}

	// Try to list tasks to verify storage is working (use default project "ABC")
	tasks, err := h.storage.ListTasks("ABC")
	duration := time.Since(start)
	check.Duration = duration.String()

	if err != nil {
		check.Status = "unhealthy"
		check.Error = err.Error()
		check.Details["message"] = "Failed to connect to storage backend"
	} else {
		check.Status = "healthy"
		check.Details["message"] = "Storage backend is operational"
		check.Details["tasks_count"] = len(tasks)
	}

	return check
}

// checkLLM verifies LLM service connectivity
func (h *HealthChecker) checkLLM() Check {
	start := time.Now()
	check := Check{
		Name:    "llm",
		Details: make(map[string]interface{}),
	}

	if h.llmService == nil {
		check.Status = "not_configured"
		check.Details["message"] = "LLM service not configured"
		return check
	}

	if !h.llmService.IsEnabled() {
		check.Status = "disabled"
		check.Details["message"] = "LLM service is disabled in configuration"
		return check
	}

	// Get provider information
	providerInfo := h.llmService.GetProviderInfo()
	for key, value := range providerInfo {
		check.Details[key] = value
	}

	// Perform health check with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Longer timeout for Ollama
	defer cancel()

	err := h.llmService.HealthCheck(ctx)
	duration := time.Since(start)
	check.Duration = duration.String()

	if err != nil {
		check.Status = "unhealthy"
		check.Error = err.Error()
		
		// Add specific guidance for common Ollama issues
		if provider, exists := check.Details["provider"]; exists && provider == "ollama" {
			if duration > 10*time.Second {
				check.Details["suggestion"] = "Health check timed out - ensure Ollama is running and responsive"
			} else if err.Error() != "" {
				switch {
				case strings.Contains(err.Error(), "connection refused"):
					check.Details["suggestion"] = "Ollama appears to be offline. Start Ollama with 'ollama serve'"
				case strings.Contains(err.Error(), "model") && strings.Contains(err.Error(), "not found"):
					check.Details["suggestion"] = "The configured model is not available. Run 'ollama pull <model-name>' to download it"
				case strings.Contains(err.Error(), "timeout"):
					check.Details["suggestion"] = "Ollama is responding slowly. Consider increasing timeout or checking system resources"
				default:
					check.Details["suggestion"] = "Check Ollama logs and ensure it's properly configured"
				}
			}
		}
	} else {
		check.Status = "healthy"
		if provider, exists := check.Details["provider"]; exists {
			check.Details["message"] = fmt.Sprintf("LLM provider '%v' is operational", provider)
		}
	}

	return check
}
