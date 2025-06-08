package health

import (
	"encoding/json"
	"net/http"
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
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	storage storage.Storage
	version string
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(storage storage.Storage, version string) *HealthChecker {
	return &HealthChecker{
		storage: storage,
		version: version,
	}
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
	check := Check{
		Name: "storage",
	}

	// Try to list tasks to verify storage is working
	_, err := h.storage.ListTasks()
	if err != nil {
		check.Status = "unhealthy"
		check.Error = err.Error()
	} else {
		check.Status = "healthy"
	}

	return check
}
