package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aykay76/projectflow/internal/llm"
)

// ExampleTranslations demonstrates how the translator works
func ExampleTranslations() {
	// Mock responses for different user inputs
	examples := []struct {
		userInput    string
		mockResponse string
		description  string
	}{
		{
			userInput:   "Create a high priority task to fix the login bug",
			description: "Creating a task with priority",
			mockResponse: `{
				"intent": "create_task",
				"confidence": 0.95,
				"parameters": {
					"title": "Fix the login bug",
					"priority": "high",
					"type": "task"
				}
			}`,
		},
		{
			userInput:   "List all tasks in the PF project",
			description: "Listing tasks in a project",
			mockResponse: `{
				"intent": "list_tasks",
				"confidence": 0.9,
				"parameters": {
					"project_id": "PF"
				}
			}`,
		},
		{
			userInput:   "Mark task PF-123 as done",
			description: "Updating task status",
			mockResponse: `{
				"intent": "update_task",
				"confidence": 0.93,
				"parameters": {
					"task_id": "PF-123",
					"status": "done"
				}
			}`,
		},
		{
			userInput:   "Show me task 456",
			description: "Reading a specific task",
			mockResponse: `{
				"intent": "read_task",
				"confidence": 0.88,
				"parameters": {
					"task_id": "PF-456"
				}
			}`,
		},
		{
			userInput:   "Delete task PF-789",
			description: "Deleting a task (requires confirmation)",
			mockResponse: `{
				"intent": "delete_task",
				"confidence": 0.96,
				"parameters": {
					"task_id": "PF-789"
				}
			}`,
		},
		{
			userInput:   "What can you help me with?",
			description: "Getting help",
			mockResponse: `{
				"intent": "get_help",
				"confidence": 0.99,
				"parameters": {}
			}`,
		},
		{
			userInput:   "Make me a sandwich",
			description: "Handling unclear requests",
			mockResponse: `{
				"intent": "unknown",
				"confidence": 0.1,
				"error_message": "I can only help with task and project management",
				"requires_clarification": true,
				"suggested_actions": [
					"Try asking about creating, updating, or listing tasks",
					"Ask for help to see what I can do"
				]
			}`,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fmt.Println("=== ProjectFlow Natural Language Translation Examples ===")
	fmt.Println()

	for i, example := range examples {
		fmt.Printf("%d. %s\n", i+1, example.description)
		fmt.Printf("   User Input: \"%s\"\n", example.userInput)

		// Create mock service
		mockLLM := &MockExampleService{response: example.mockResponse}
		translator := NewTranslator(mockLLM, logger)

		// Translate
		result, err := translator.Translate(context.Background(), example.userInput)
		if err != nil {
			fmt.Printf("   ❌ Error: %s\n", err)
		} else {
			fmt.Printf("   📝 Intent: %s (confidence: %.2f)\n", result.ParsedRequest.Intent, result.ParsedRequest.Confidence)
			fmt.Printf("   🤖 Response: %s\n", result.HumanResponse)
			fmt.Printf("   🔧 MCP Commands: %d\n", len(result.MCPCommands))

			if len(result.MCPCommands) > 0 {
				for j, cmd := range result.MCPCommands {
					fmt.Printf("      %d. %s", j+1, cmd.Method)
					if len(cmd.Parameters) > 0 {
						paramBytes, _ := json.Marshal(cmd.Parameters)
						fmt.Printf(" %s", string(paramBytes))
					}
					fmt.Println()
				}
			}

			if result.RequiresConfirmation {
				fmt.Printf("   ⚠️  Requires confirmation\n")
			}
		}

		fmt.Println()
	}
}

// MockExampleService for demonstration purposes
type MockExampleService struct {
	response string
}

func (m *MockExampleService) IsEnabled() bool {
	return true
}

func (m *MockExampleService) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{
		Choices: []llm.ChatResponseChoice{
			{
				Message: llm.Message{
					Content: m.response,
				},
			},
		},
	}, nil
}
