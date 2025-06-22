package translator

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/aykay76/projectflow/internal/llm"
)

// MockLLMService implements the LLMService interface for testing
type MockLLMService struct {
	enabled bool
	chatResponse *llm.ChatResponse
	chatError error
}

func (m *MockLLMService) IsEnabled() bool {
	return m.enabled
}

func (m *MockLLMService) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.chatError != nil {
		return nil, m.chatError
	}
	return m.chatResponse, nil
}

func TestTranslator_TranslateDisabled(t *testing.T) {
	mockLLM := &MockLLMService{enabled: false}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	translator := NewTranslator(mockLLM, logger)

	_, err := translator.Translate(context.Background(), "create a task")

	if err == nil {
		t.Error("Expected error when LLM service is disabled")
	}
}

func TestTranslator_CreateTaskIntent(t *testing.T) {
	mockResponse := &llm.ChatResponse{
		Choices: []llm.ChatResponseChoice{
			{
				Message: llm.Message{
					Content: `{
						"intent": "create_task",
						"confidence": 0.95,
						"parameters": {
							"title": "Fix login bug",
							"priority": "high",
							"type": "task"
						}
					}`,
				},
			},
		},
	}

	mockLLM := &MockLLMService{
		enabled: true,
		chatResponse: mockResponse,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	translator := NewTranslator(mockLLM, logger)

	result, err := translator.Translate(context.Background(), "Create a high priority task to fix login bug")

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if result.ParsedRequest.Intent != IntentCreateTask {
		t.Errorf("Expected intent %s, got %s", IntentCreateTask, result.ParsedRequest.Intent)
	}

	if result.ParsedRequest.Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", result.ParsedRequest.Confidence)
	}

	if len(result.MCPCommands) != 1 {
		t.Errorf("Expected 1 MCP command, got %d", len(result.MCPCommands))
	}

	if result.MCPCommands[0].Method != "create_task" {
		t.Errorf("Expected MCP method 'create_task', got '%s'", result.MCPCommands[0].Method)
	}

	title, ok := result.MCPCommands[0].Parameters["title"].(string)
	if !ok || title != "Fix login bug" {
		t.Errorf("Expected title 'Fix login bug', got '%v'", result.MCPCommands[0].Parameters["title"])
	}
}

func TestTranslator_ListTasksIntent(t *testing.T) {
	mockResponse := &llm.ChatResponse{
		Choices: []llm.ChatResponseChoice{
			{
				Message: llm.Message{
					Content: `{
						"intent": "list_tasks",
						"confidence": 0.9,
						"parameters": {
							"project_id": "PF"
						}
					}`,
				},
			},
		},
	}

	mockLLM := &MockLLMService{
		enabled: true,
		chatResponse: mockResponse,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	translator := NewTranslator(mockLLM, logger)

	result, err := translator.Translate(context.Background(), "List all tasks in PF project")

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if result.ParsedRequest.Intent != IntentListTasks {
		t.Errorf("Expected intent %s, got %s", IntentListTasks, result.ParsedRequest.Intent)
	}

	if len(result.MCPCommands) != 1 {
		t.Errorf("Expected 1 MCP command, got %d", len(result.MCPCommands))
	}

	if result.MCPCommands[0].Method != "list_tasks" {
		t.Errorf("Expected MCP method 'list_tasks', got '%s'", result.MCPCommands[0].Method)
	}
}

func TestTranslator_UnknownIntent(t *testing.T) {
	mockResponse := &llm.ChatResponse{
		Choices: []llm.ChatResponseChoice{
			{
				Message: llm.Message{
					Content: `{
						"intent": "unknown",
						"confidence": 0.2,
						"error_message": "Could not understand request",
						"requires_clarification": true,
						"suggested_actions": ["Try rephrasing your request"]
					}`,
				},
			},
		},
	}

	mockLLM := &MockLLMService{
		enabled: true,
		chatResponse: mockResponse,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	translator := NewTranslator(mockLLM, logger)

	result, err := translator.Translate(context.Background(), "blah blah nonsense")

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if result.ParsedRequest.Intent != IntentUnknown {
		t.Errorf("Expected intent %s, got %s", IntentUnknown, result.ParsedRequest.Intent)
	}

	if len(result.MCPCommands) != 0 {
		t.Errorf("Expected 0 MCP commands for unknown intent, got %d", len(result.MCPCommands))
	}

	if !result.ParsedRequest.RequiresClarification {
		t.Error("Expected requires_clarification to be true")
	}
}

func TestExtractTaskID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PF-123", "PF-123"},
		{"task 123", "PF-123"},
		{"issue 456", "PF-456"},
		{"#789", "PF-789"},
		{"ABC-999", "ABC-999"},
		{"123", "PF-123"},
		{"", ""},
		{"invalid", "invalid"},
	}

	for _, test := range tests {
		result := ExtractTaskID(test.input)
		if result != test.expected {
			t.Errorf("ExtractTaskID(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestParseDueDate(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"2025-06-30", false},
		{"2025/06/30", false},
		{"today", false},
		{"tomorrow", false},
		{"next week", false},
		{"invalid date", true},
		{"", false}, // Empty should not error
	}

	for _, test := range tests {
		result, err := ParseDueDate(test.input)
		if test.expectError && err == nil {
			t.Errorf("ParseDueDate(%s) expected error but got none", test.input)
		}
		if !test.expectError && err != nil {
			t.Errorf("ParseDueDate(%s) unexpected error: %s", test.input, err)
		}
		if !test.expectError && test.input != "" && result == "" {
			t.Errorf("ParseDueDate(%s) returned empty result", test.input)
		}
	}
}

func TestTranslator_ValidateParameters(t *testing.T) {
	// Create a minimal mock for validation testing
	mockLLM := &MockLLMService{enabled: false}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	translator := NewTranslator(mockLLM, logger)

	tests := []struct {
		name        string
		request     *ParsedRequest
		expectError bool
	}{
		{
			name: "valid create task",
			request: &ParsedRequest{
				Intent: IntentCreateTask,
				Parameters: map[string]interface{}{
					"title":    "Test task",
					"priority": "high",
					"status":   "todo",
					"type":     "task",
				},
			},
			expectError: false,
		},
		{
			name: "create task missing title",
			request: &ParsedRequest{
				Intent:     IntentCreateTask,
				Parameters: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "create task invalid priority",
			request: &ParsedRequest{
				Intent: IntentCreateTask,
				Parameters: map[string]interface{}{
					"title":    "Test task",
					"priority": "invalid",
				},
			},
			expectError: true,
		},
		{
			name: "update task missing ID",
			request: &ParsedRequest{
				Intent:     IntentUpdateTask,
				Parameters: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "valid update task",
			request: &ParsedRequest{
				Intent: IntentUpdateTask,
				Parameters: map[string]interface{}{
					"task_id": "PF-123",
					"status":  "done",
				},
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := translator.ValidateParameters(test.request)
			if test.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Unexpected validation error: %s", err)
			}
		})
	}
}
