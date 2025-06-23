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
	tasks []models.Task
	err   error
}

func (m *MockStorage) ListTasks(projectID string) ([]models.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasks, nil
}

func (m *MockStorage) GetTask(id string) (*models.Task, error) { return nil, nil }
func (m *MockStorage) CreateTask(task *models.Task) error      { return nil }
func (m *MockStorage) UpdateTask(task *models.Task) error      { return nil }
func (m *MockStorage) DeleteTask(id string) error              { return nil }
func (m *MockStorage) GetTaskChildren(parentID string) ([]models.Task, error) {
	return nil, nil
}
func (m *MockStorage) AddTaskChild(parentID, childID string) error       { return nil }
func (m *MockStorage) RemoveTaskChild(parentID, childID string) error    { return nil }
func (m *MockStorage) MoveTask(taskID, newParentID string) error         { return nil }
func (m *MockStorage) GetHierarchy() ([]models.Task, error)              { return nil, nil }
func (m *MockStorage) ListProjects() ([]models.Project, error)           { return nil, nil }
func (m *MockStorage) GetProject(id string) (*models.Project, error)     { return nil, nil }
func (m *MockStorage) CreateProject(project *models.Project) error       { return nil }
func (m *MockStorage) UpdateProject(project *models.Project) error       { return nil }
func (m *MockStorage) DeleteProject(id string) error                     { return nil }
func (m *MockStorage) Close() error                                      { return nil }
func (m *MockStorage) GetNextDisplayID(projectID string) (string, error) { return "1", nil }
func (m *MockStorage) GetProjectByDisplayPrefix(prefix string) (*models.Project, error) {
	return nil, nil
}

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
	storage := &MockStorage{tasks: []models.Task{}}
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
	storage := &MockStorage{tasks: []models.Task{}}
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
	storage := &MockStorage{tasks: []models.Task{}}
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
	storage := &MockStorage{tasks: []models.Task{}}
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
