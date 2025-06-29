package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aykay76/projectflow/internal/models"
)

// MockStorage for testing
type MockStorage struct {
	tasks []*models.Task
	err   error
}

func (m *MockStorage) ListTasks(ctx context.Context, projectID string) ([]*models.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasks, nil
}

func (m *MockStorage) GetTask(ctx context.Context, id string) (*models.Task, error) { return nil, nil }
func (m *MockStorage) GetTaskByDisplayID(ctx context.Context, displayID string) (*models.Task, error) {
	return nil, nil
}
func (m *MockStorage) CreateTask(ctx context.Context, task *models.Task) error { return nil }
func (m *MockStorage) UpdateTask(ctx context.Context, task *models.Task) error { return nil }
func (m *MockStorage) DeleteTask(ctx context.Context, id string) error         { return nil }
func (m *MockStorage) GetTaskChildren(ctx context.Context, parentID string) ([]*models.Task, error) {
	return nil, nil
}
func (m *MockStorage) GetTaskParent(ctx context.Context, childID string) (*models.Task, error) {
	return nil, nil
}
func (m *MockStorage) GetTaskHierarchy(ctx context.Context) ([]*models.HierarchyTask, error) {
	return nil, nil
}
func (m *MockStorage) ListProjects(ctx context.Context) ([]*models.Project, error) { return nil, nil }
func (m *MockStorage) GetProject(ctx context.Context, id string) (*models.Project, error) {
	return nil, nil
}
func (m *MockStorage) CreateProject(ctx context.Context, project *models.Project) error { return nil }
func (m *MockStorage) UpdateProject(ctx context.Context, project *models.Project) error { return nil }
func (m *MockStorage) DeleteProject(ctx context.Context, id string) error               { return nil }
func (m *MockStorage) GetProjectByName(ctx context.Context, name string) (*models.Project, error) {
	return nil, nil
}
func (m *MockStorage) Close() error { return nil }
func (m *MockStorage) GetNextDisplayID(ctx context.Context, projectID string) (string, error) {
	return "1", nil
}
func (m *MockStorage) GetProjectByDisplayPrefix(ctx context.Context, prefix string) (*models.Project, error) {
	return nil, nil
}
func (m *MockStorage) CreateTenant(ctx context.Context, tenant *models.Tenant) error { return nil }
func (m *MockStorage) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	return nil, nil
}
func (m *MockStorage) UpdateTenant(ctx context.Context, tenant *models.Tenant) error { return nil }
func (m *MockStorage) DeleteTenant(ctx context.Context, id string) error             { return nil }
func (m *MockStorage) ListTenants(ctx context.Context, limit, offset int) ([]*models.Tenant, int, error) {
	return nil, 0, nil
}
func (m *MockStorage) TenantExists(ctx context.Context, id string) bool  { return false }
func (m *MockStorage) TaskExists(ctx context.Context, id string) bool    { return false }
func (m *MockStorage) ProjectExists(ctx context.Context, id string) bool { return false }

// MockLLMService for testing
type MockLLMService struct {
	enabled      bool
	healthErr    error
	providerInfo map[string]interface{}
}

func (m *MockLLMService) IsEnabled() bool                         { return m.enabled }
func (m *MockLLMService) HealthCheck(ctx context.Context) error   { return m.healthErr }
func (m *MockLLMService) GetProviderInfo() map[string]interface{} { return m.providerInfo }

func TestHealthChecker_HandleHealth(t *testing.T) {
	storage := &MockStorage{}
	checker := NewHealthChecker(storage, "1.0.0-test")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	checker.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version != "1.0.0-test" {
		t.Errorf("Expected version '1.0.0-test', got '%s'", response.Version)
	}
}

func TestHealthChecker_HandleReady_NoLLM(t *testing.T) {
	storage := &MockStorage{tasks: []*models.Task{}}
	checker := NewHealthChecker(storage, "1.0.0-test")

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	checker.HandleReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", response.Status)
	}

	if len(response.Checks) != 1 {
		t.Errorf("Expected 1 check, got %d", len(response.Checks))
	}

	storageCheck := response.Checks[0]
	if storageCheck.Name != "storage" {
		t.Errorf("Expected storage check, got '%s'", storageCheck.Name)
	}

	if storageCheck.Status != "healthy" {
		t.Errorf("Expected storage status 'healthy', got '%s'", storageCheck.Status)
	}
}

func TestHealthChecker_HandleReady_WithHealthyLLM(t *testing.T) {
	storage := &MockStorage{tasks: []*models.Task{}}
	llmService := &MockLLMService{
		enabled: true,
		providerInfo: map[string]interface{}{
			"provider": "ollama",
			"model":    "llama3.2",
		},
	}

	checker := NewHealthChecker(storage, "1.0.0-test")
	checker.SetLLMService(llmService)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	checker.HandleReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", response.Status)
	}

	if len(response.Checks) != 2 {
		t.Errorf("Expected 2 checks, got %d", len(response.Checks))
	}

	// Find LLM check
	var llmCheck *Check
	for _, check := range response.Checks {
		if check.Name == "llm" {
			llmCheck = &check
			break
		}
	}

	if llmCheck == nil {
		t.Fatal("LLM check not found")
	}

	if llmCheck.Status != "healthy" {
		t.Errorf("Expected LLM status 'healthy', got '%s'", llmCheck.Status)
	}

	if llmCheck.Details == nil {
		t.Fatal("Expected LLM details to be populated")
	}

	if provider, exists := llmCheck.Details["provider"]; !exists || provider != "ollama" {
		t.Errorf("Expected provider 'ollama', got '%v'", provider)
	}

	if model, exists := llmCheck.Details["model"]; !exists || model != "llama3.2" {
		t.Errorf("Expected model 'llama3.2', got '%v'", model)
	}
}

func TestHealthChecker_HandleReady_WithUnhealthyLLM(t *testing.T) {
	storage := &MockStorage{tasks: []*models.Task{}}
	llmService := &MockLLMService{
		enabled:   true,
		healthErr: &unhealthyError{"connection refused"},
		providerInfo: map[string]interface{}{
			"provider": "ollama",
			"model":    "llama3.2",
		},
	}

	checker := NewHealthChecker(storage, "1.0.0-test")
	checker.SetLLMService(llmService)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	checker.HandleReady(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "not ready" {
		t.Errorf("Expected status 'not ready', got '%s'", response.Status)
	}

	// Find LLM check
	var llmCheck *Check
	for _, check := range response.Checks {
		if check.Name == "llm" {
			llmCheck = &check
			break
		}
	}

	if llmCheck == nil {
		t.Fatal("LLM check not found")
	}

	if llmCheck.Status != "unhealthy" {
		t.Errorf("Expected LLM status 'unhealthy', got '%s'", llmCheck.Status)
	}

	if llmCheck.Error != "connection refused" {
		t.Errorf("Expected error 'connection refused', got '%s'", llmCheck.Error)
	}

	// Check that Ollama-specific suggestion is provided
	if suggestion, exists := llmCheck.Details["suggestion"]; !exists {
		t.Error("Expected suggestion to be provided for Ollama connection error")
	} else if !strings.Contains(suggestion.(string), "ollama serve") {
		t.Errorf("Expected suggestion to mention 'ollama serve', got '%v'", suggestion)
	}
}

func TestHealthChecker_HandleReady_WithDisabledLLM(t *testing.T) {
	storage := &MockStorage{tasks: []*models.Task{}}
	llmService := &MockLLMService{enabled: false}

	checker := NewHealthChecker(storage, "1.0.0-test")
	checker.SetLLMService(llmService)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	checker.HandleReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", response.Status)
	}

	// Should only have storage check since LLM is disabled
	if len(response.Checks) != 1 {
		t.Errorf("Expected 1 check (LLM disabled), got %d", len(response.Checks))
	}
}

// unhealthyError is a test error type
type unhealthyError struct {
	msg string
}

func (e *unhealthyError) Error() string {
	return e.msg
}
