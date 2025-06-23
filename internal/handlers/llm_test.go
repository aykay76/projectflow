package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aykay76/projectflow/internal/llm"
)

// MockLLMServiceWithProviderInfo extends MockLLMService with provider info
type MockLLMServiceWithProviderInfo struct {
	MockLLMService
	providerInfo map[string]interface{}
}

func (m *MockLLMServiceWithProviderInfo) GetProviderInfo() map[string]interface{} {
	return m.providerInfo
}

// testLogger creates a logger for testing
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestLLMHandler_HandleLLMInfo_Disabled(t *testing.T) {
	mockLLM := &MockLLMService{enabled: false}
	handler := NewLLMHandler(mockLLM, testLogger())

	req := httptest.NewRequest("GET", "/api/llm/info", nil)
	w := httptest.NewRecorder()

	handler.HandleLLMInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response LLMInfoResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Enabled {
		t.Error("Expected LLM to be disabled")
	}

	if response.Status != "disabled" {
		t.Errorf("Expected status 'disabled', got '%s'", response.Status)
	}

	if response.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", response.Provider)
	}
}

func TestLLMHandler_HandleLLMInfo_Enabled(t *testing.T) {
	mockLLM := &MockLLMServiceWithProviderInfo{
		MockLLMService: MockLLMService{enabled: true},
		providerInfo: map[string]interface{}{
			"enabled":  true,
			"provider": "ollama",
			"model":    "llama3.2",
		},
	}
	handler := NewLLMHandler(mockLLM, testLogger())

	req := httptest.NewRequest("GET", "/api/llm/info", nil)
	w := httptest.NewRecorder()

	handler.HandleLLMInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response LLMInfoResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Enabled {
		t.Error("Expected LLM to be enabled")
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Provider != "ollama" {
		t.Errorf("Expected provider 'ollama', got '%s'", response.Provider)
	}

	if response.Model != "llama3.2" {
		t.Errorf("Expected model 'llama3.2', got '%s'", response.Model)
	}
}

func TestLLMHandler_HandleLLMChat_Disabled(t *testing.T) {
	mockLLM := &MockLLMService{enabled: false}
	handler := NewLLMHandler(mockLLM, testLogger())

	reqBody := LLMChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/llm/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLLMChat(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestLLMHandler_HandleLLMChat_Success(t *testing.T) {
	mockResponse := &llm.ChatResponse{
		ID:     "test-response",
		Object: "chat.completion",
		Model:  "llama3.2",
		Choices: []llm.ChatResponseChoice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
				FinishReason: "stop",
			},
		},
		Usage: llm.ChatResponseUsage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}

	mockLLM := &MockLLMService{
		enabled:  true,
		response: mockResponse,
	}
	handler := NewLLMHandler(mockLLM, testLogger())

	reqBody := LLMChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/llm/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLLMChat(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response LLMChatResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Response == nil {
		t.Fatal("Expected response to contain LLM response")
	}

	if response.Response.ID != mockResponse.ID {
		t.Errorf("Expected response ID '%s', got '%s'", mockResponse.ID, response.Response.ID)
	}

	if len(response.Response.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(response.Response.Choices))
	}

	if response.Response.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("Expected specific content, got '%s'", response.Response.Choices[0].Message.Content)
	}
}

func TestLLMHandler_HandleLLMChat_EmptyMessages(t *testing.T) {
	mockLLM := &MockLLMService{enabled: true}
	handler := NewLLMHandler(mockLLM, testLogger())

	reqBody := LLMChatRequest{
		Messages: []llm.Message{}, // Empty messages
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/llm/chat", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleLLMChat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestLLMHandler_HandleLLMHealthCheck_Disabled(t *testing.T) {
	mockLLM := &MockLLMService{enabled: false}
	handler := NewLLMHandler(mockLLM, testLogger())

	req := httptest.NewRequest("GET", "/api/llm/health", nil)
	w := httptest.NewRecorder()

	handler.HandleLLMHealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "disabled" {
		t.Errorf("Expected status 'disabled', got '%v'", response["status"])
	}
}

func TestLLMHandler_HandleLLMHealthCheck_Healthy(t *testing.T) {
	mockLLM := &MockLLMService{enabled: true}
	handler := NewLLMHandler(mockLLM, testLogger())

	req := httptest.NewRequest("GET", "/api/llm/health", nil)
	w := httptest.NewRecorder()

	handler.HandleLLMHealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}
}
