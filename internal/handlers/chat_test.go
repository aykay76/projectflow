package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aykay76/projectflow/internal/llm"
)

// MockLLMService for testing
type MockLLMService struct {
	enabled  bool
	response *llm.ChatResponse
	err      error
}

func (m *MockLLMService) IsEnabled() bool {
	return m.enabled
}

func (m *MockLLMService) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *MockLLMService) HealthCheck(ctx context.Context) error {
	return nil
}

func TestChatHandler_HandleChat(t *testing.T) {
	// Setup mock storage
	storage := newMockStorage()
	
	// Setup mock LLM service with a response to create a task
	mockLLM := &MockLLMService{
		enabled: true,
		response: &llm.ChatResponse{
			Choices: []llm.ChatResponseChoice{
				{
					Message: llm.Message{
						Content: `{
							"intent": "create_task",
							"confidence": 0.95,
							"parameters": {
								"title": "Fix the login bug",
								"description": "There's an issue with user login",
								"priority": "high",
								"status": "todo",
								"type": "task",
								"project_id": "PF"
							},
							"requires_clarification": false
						}`,
					},
				},
			},
		},
	}

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create chat handler
	chatHandler := NewChatHandler(storage, mockLLM, logger)

	// Test request
	reqBody := ChatRequest{
		Message:        "Create a high priority task to fix the login bug",
		ConversationID: "test-conversation",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Test response
	w := httptest.NewRecorder()
	chatHandler.HandleChat(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response
	if response.ConversationID != "test-conversation" {
		t.Errorf("Expected conversation ID 'test-conversation', got '%s'", response.ConversationID)
	}

	if response.Intent != "create_task" {
		t.Errorf("Expected intent 'create_task', got '%s'", response.Intent)
	}

	if len(response.ActionsTaken) != 1 || response.ActionsTaken[0] != "create_task" {
		t.Errorf("Expected actions taken ['create_task'], got %v", response.ActionsTaken)
	}

	if len(response.TaskIDs) != 1 {
		t.Errorf("Expected 1 task ID, got %v", response.TaskIDs)
	}

	if response.Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", response.Confidence)
	}
}

func TestChatHandler_HandleChatHistory(t *testing.T) {
	storage := newMockStorage()
	mockLLM := &MockLLMService{
		enabled: true,
		response: &llm.ChatResponse{
			Choices: []llm.ChatResponseChoice{
				{
					Message: llm.Message{
						Content: `{
							"intent": "get_help",
							"confidence": 0.8,
							"parameters": {},
							"requires_clarification": false
						}`,
					},
				},
			},
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	chatHandler := NewChatHandler(storage, mockLLM, logger)

	// First, send a chat message to create history
	reqBody := ChatRequest{
		Message:        "Hello",
		ConversationID: "test-conversation",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	chatHandler.HandleChat(w, req)

	// Now test getting the history
	req = httptest.NewRequest(http.MethodGet, "/api/chat/history?conversation_id=test-conversation", nil)
	w = httptest.NewRecorder()
	chatHandler.HandleChatHistory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var conversation Conversation
	err := json.Unmarshal(w.Body.Bytes(), &conversation)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if conversation.ID != "test-conversation" {
		t.Errorf("Expected conversation ID 'test-conversation', got '%s'", conversation.ID)
	}

	if len(conversation.Messages) != 2 { // user message + assistant response
		t.Errorf("Expected 2 messages, got %d", len(conversation.Messages))
	}
}

func TestChatHandler_LLMDisabled(t *testing.T) {
	storage := newMockStorage()
	mockLLM := &MockLLMService{
		enabled: false, // LLM disabled
		response: nil,  // No response needed since LLM is disabled
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	chatHandler := NewChatHandler(storage, mockLLM, logger)

	reqBody := ChatRequest{
		Message:        "Create a task",
		ConversationID: "test-conversation",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	chatHandler.HandleChat(w, req)

	// Should return an error response but still maintain conversation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 (graceful error handling), got %d", w.Code)
	}

	var response ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Intent != "error" {
		t.Errorf("Expected intent 'error', got '%s'", response.Intent)
	}

	if response.Confidence != 0.0 {
		t.Errorf("Expected confidence 0.0, got %f", response.Confidence)
	}
}
