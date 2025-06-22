package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aykay76/projectflow/internal/llm"
)

// Intent represents the user's intention
type Intent string

const (
	IntentCreateTask    Intent = "create_task"
	IntentReadTask      Intent = "read_task"
	IntentUpdateTask    Intent = "update_task"
	IntentDeleteTask    Intent = "delete_task"
	IntentListTasks     Intent = "list_tasks"
	IntentCreateProject Intent = "create_project"
	IntentListProjects  Intent = "list_projects"
	IntentGetHelp       Intent = "get_help"
	IntentUnknown       Intent = "unknown"
)

// Priority represents task priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Status represents task status
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
)

// TaskType represents task type
type TaskType string

const (
	TaskTypeEpic    TaskType = "epic"
	TaskTypeStory   TaskType = "story"
	TaskTypeTask    TaskType = "task"
	TaskTypeSubtask TaskType = "subtask"
)

// ParsedRequest represents the structured interpretation of a natural language request
type ParsedRequest struct {
	Intent                Intent                 `json:"intent"`
	Confidence            float64                `json:"confidence"`
	Parameters            map[string]interface{} `json:"parameters"`
	ErrorMessage          string                 `json:"error_message,omitempty"`
	RequiresClarification bool                   `json:"requires_clarification,omitempty"`
	SuggestedActions      []string               `json:"suggested_actions,omitempty"`
}

// MCPCommand represents a command to be sent to the MCP server
type MCPCommand struct {
	Method     string                 `json:"method"`
	Parameters map[string]interface{} `json:"parameters"`
}

// TranslationResult represents the complete translation result
type TranslationResult struct {
	ParsedRequest        ParsedRequest `json:"parsed_request"`
	MCPCommands          []MCPCommand  `json:"mcp_commands"`
	HumanResponse        string        `json:"human_response"`
	RequiresConfirmation bool          `json:"requires_confirmation,omitempty"`
}

// LLMService interface for the translator
type LLMService interface {
	IsEnabled() bool
	Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
}

// Translator handles natural language to MCP translation
type Translator struct {
	llmService LLMService
	logger     *slog.Logger
}

// NewTranslator creates a new translator instance
func NewTranslator(llmService LLMService, logger *slog.Logger) *Translator {
	return &Translator{
		llmService: llmService,
		logger:     logger,
	}
}

// Translate converts natural language input to MCP commands
func (t *Translator) Translate(ctx context.Context, userInput string) (*TranslationResult, error) {
	if !t.llmService.IsEnabled() {
		return nil, fmt.Errorf("LLM service is not enabled")
	}

	t.logger.Debug("Translating user input", "input", userInput)

	// Step 1: Parse the user's intent using LLM
	parsedRequest, err := t.parseIntent(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %w", err)
	}

	// Step 2: Generate MCP commands based on the parsed request
	mcpCommands, err := t.generateMCPCommands(parsedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate MCP commands: %w", err)
	}

	// Step 3: Generate human-friendly response
	humanResponse := t.generateHumanResponse(parsedRequest, mcpCommands)

	result := &TranslationResult{
		ParsedRequest:        *parsedRequest,
		MCPCommands:          mcpCommands,
		HumanResponse:        humanResponse,
		RequiresConfirmation: t.requiresConfirmation(parsedRequest),
	}

	t.logger.Debug("Translation completed",
		"intent", parsedRequest.Intent,
		"confidence", parsedRequest.Confidence,
		"commands_count", len(mcpCommands))

	return result, nil
}

// parseIntent uses LLM to parse the user's natural language input
func (t *Translator) parseIntent(ctx context.Context, userInput string) (*ParsedRequest, error) {
	systemPrompt := t.getSystemPrompt()
	userPrompt := t.getUserPrompt(userInput)

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	req := llm.ChatRequest{
		Messages:    messages,
		Temperature: 0.3, // Lower temperature for more consistent structured output
		MaxTokens:   500,
	}

	resp, err := t.llmService.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM chat failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Parse the JSON response
	var parsedRequest ParsedRequest
	content := resp.Choices[0].Message.Content

	// Clean up the response (remove markdown code blocks if present)
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if err := json.Unmarshal([]byte(content), &parsedRequest); err != nil {
		t.logger.Warn("Failed to parse LLM response as JSON", "content", content, "error", err)

		// Fallback to unknown intent
		return &ParsedRequest{
			Intent:                IntentUnknown,
			Confidence:            0.0,
			ErrorMessage:          fmt.Sprintf("Failed to understand request: %s", userInput),
			RequiresClarification: true,
			SuggestedActions: []string{
				"Try rephrasing your request",
				"Use specific commands like 'create task', 'list tasks', 'update task PF-123'",
			},
		}, nil
	}

	return &parsedRequest, nil
}

// getSystemPrompt returns the system prompt for the LLM
func (t *Translator) getSystemPrompt() string {
	return `You are an intelligent assistant for ProjectFlow, a task management system. Your job is to understand user requests and convert them into structured data.

IMPORTANT: You must respond with valid JSON only. Do not include any text before or after the JSON.

Available intents:
- create_task: Create a new task
- read_task: Get details about a specific task
- update_task: Update an existing task
- delete_task: Delete a task
- list_tasks: List tasks (with optional filters)
- create_project: Create a new project
- list_projects: List all projects
- get_help: User needs help or information
- unknown: Cannot understand the request

Task priorities: low, medium, high, critical
Task statuses: todo, in_progress, done, blocked
Task types: epic, story, task, subtask

For each request, analyze the user's intent and extract relevant parameters. Respond with JSON in this format:
{
  "intent": "create_task",
  "confidence": 0.95,
  "parameters": {
    "title": "Task title",
    "description": "Task description",
    "priority": "high",
    "status": "todo",
    "type": "task",
    "project_id": "PF",
    "parent_id": "PF-123",
    "due_date": "2025-06-30"
  },
  "error_message": "",
  "requires_clarification": false,
  "suggested_actions": []
}

Guidelines:
- Set confidence between 0.0 and 1.0
- Only include parameters that are explicitly mentioned or clearly implied
- If unsure, set requires_clarification to true
- For ambiguous requests, provide suggested_actions
- Default project_id to "PF" if not specified
- Parse dates in YYYY-MM-DD format
- Extract task IDs from formats like "PF-123", "task 123", "issue #123"`
}

// getUserPrompt formats the user input for the LLM
func (t *Translator) getUserPrompt(userInput string) string {
	return fmt.Sprintf("Parse this user request: %s", userInput)
}

// generateMCPCommands converts parsed request to MCP commands
func (t *Translator) generateMCPCommands(parsed *ParsedRequest) ([]MCPCommand, error) {
	if parsed.Intent == IntentUnknown {
		return []MCPCommand{}, nil
	}

	var commands []MCPCommand

	switch parsed.Intent {
	case IntentCreateTask:
		cmd := MCPCommand{
			Method:     "create_task",
			Parameters: make(map[string]interface{}),
		}

		// Extract parameters from parsed request
		if title, ok := parsed.Parameters["title"].(string); ok && title != "" {
			cmd.Parameters["title"] = title
		} else {
			return nil, fmt.Errorf("task title is required")
		}

		if description, ok := parsed.Parameters["description"].(string); ok {
			cmd.Parameters["description"] = description
		}

		if priority, ok := parsed.Parameters["priority"].(string); ok {
			cmd.Parameters["priority"] = priority
		}

		if status, ok := parsed.Parameters["status"].(string); ok {
			cmd.Parameters["status"] = status
		}

		if taskType, ok := parsed.Parameters["type"].(string); ok {
			cmd.Parameters["type"] = taskType
		}

		if projectID, ok := parsed.Parameters["project_id"].(string); ok {
			cmd.Parameters["project_id"] = projectID
		} else {
			cmd.Parameters["project_id"] = "PF" // Default project
		}

		if parentID, ok := parsed.Parameters["parent_id"].(string); ok {
			cmd.Parameters["parent_id"] = parentID
		}

		if dueDate, ok := parsed.Parameters["due_date"].(string); ok {
			cmd.Parameters["due_date"] = dueDate
		}

		commands = append(commands, cmd)

	case IntentListTasks:
		cmd := MCPCommand{
			Method:     "list_tasks",
			Parameters: make(map[string]interface{}),
		}

		if projectID, ok := parsed.Parameters["project_id"].(string); ok {
			cmd.Parameters["project_id"] = projectID
		} else {
			cmd.Parameters["project_id"] = "PF" // Default project
		}

		commands = append(commands, cmd)

	case IntentReadTask:
		cmd := MCPCommand{
			Method:     "get_task",
			Parameters: make(map[string]interface{}),
		}

		if taskID, ok := parsed.Parameters["task_id"].(string); ok && taskID != "" {
			cmd.Parameters["id"] = taskID
		} else {
			return nil, fmt.Errorf("task ID is required")
		}

		commands = append(commands, cmd)

	case IntentUpdateTask:
		cmd := MCPCommand{
			Method:     "update_task",
			Parameters: make(map[string]interface{}),
		}

		if taskID, ok := parsed.Parameters["task_id"].(string); ok && taskID != "" {
			cmd.Parameters["id"] = taskID
		} else {
			return nil, fmt.Errorf("task ID is required")
		}

		// Add any fields to update
		updateFields := []string{"title", "description", "priority", "status", "type", "due_date"}
		for _, field := range updateFields {
			if value, ok := parsed.Parameters[field]; ok {
				cmd.Parameters[field] = value
			}
		}

		commands = append(commands, cmd)

	case IntentDeleteTask:
		cmd := MCPCommand{
			Method:     "delete_task",
			Parameters: make(map[string]interface{}),
		}

		if taskID, ok := parsed.Parameters["task_id"].(string); ok && taskID != "" {
			cmd.Parameters["id"] = taskID
		} else {
			return nil, fmt.Errorf("task ID is required")
		}

		commands = append(commands, cmd)

	case IntentListProjects:
		cmd := MCPCommand{
			Method:     "list_projects",
			Parameters: make(map[string]interface{}),
		}
		commands = append(commands, cmd)

	case IntentCreateProject:
		cmd := MCPCommand{
			Method:     "create_project",
			Parameters: make(map[string]interface{}),
		}

		if name, ok := parsed.Parameters["name"].(string); ok && name != "" {
			cmd.Parameters["name"] = name
		} else {
			return nil, fmt.Errorf("project name is required")
		}

		if description, ok := parsed.Parameters["description"].(string); ok {
			cmd.Parameters["description"] = description
		}

		if prefix, ok := parsed.Parameters["prefix"].(string); ok {
			cmd.Parameters["prefix"] = prefix
		}

		commands = append(commands, cmd)

	default:
		// For help or unknown intents, no MCP commands needed
		return []MCPCommand{}, nil
	}

	return commands, nil
}

// generateHumanResponse creates a human-friendly response
func (t *Translator) generateHumanResponse(parsed *ParsedRequest, commands []MCPCommand) string {
	if parsed.Intent == IntentUnknown {
		response := "I'm not sure what you'd like me to do. "
		if parsed.ErrorMessage != "" {
			response += parsed.ErrorMessage
		}
		if len(parsed.SuggestedActions) > 0 {
			response += "\n\nHere are some things you can try:\n"
			for _, action := range parsed.SuggestedActions {
				response += "• " + action + "\n"
			}
		}
		return response
	}

	if parsed.RequiresClarification {
		response := "I need more information to help you with that request."
		if len(parsed.SuggestedActions) > 0 {
			response += "\n\nCould you please:\n"
			for _, action := range parsed.SuggestedActions {
				response += "• " + action + "\n"
			}
		}
		return response
	}

	switch parsed.Intent {
	case IntentCreateTask:
		if title, ok := parsed.Parameters["title"].(string); ok {
			return fmt.Sprintf("I'll create a new task called '%s' for you.", title)
		}
		return "I'll create a new task for you."

	case IntentListTasks:
		if projectID, ok := parsed.Parameters["project_id"].(string); ok {
			return fmt.Sprintf("Here are the tasks in project %s:", projectID)
		}
		return "Here are your tasks:"

	case IntentReadTask:
		if taskID, ok := parsed.Parameters["task_id"].(string); ok {
			return fmt.Sprintf("Here are the details for task %s:", taskID)
		}
		return "Here are the task details:"

	case IntentUpdateTask:
		if taskID, ok := parsed.Parameters["task_id"].(string); ok {
			return fmt.Sprintf("I'll update task %s for you.", taskID)
		}
		return "I'll update that task for you."

	case IntentDeleteTask:
		if taskID, ok := parsed.Parameters["task_id"].(string); ok {
			return fmt.Sprintf("I'll delete task %s for you.", taskID)
		}
		return "I'll delete that task for you."

	case IntentListProjects:
		return "Here are all your projects:"

	case IntentCreateProject:
		if name, ok := parsed.Parameters["name"].(string); ok {
			return fmt.Sprintf("I'll create a new project called '%s' for you.", name)
		}
		return "I'll create a new project for you."

	case IntentGetHelp:
		return `I can help you manage your tasks and projects! Here are some things you can ask me:

**Tasks:**
• "Create a high priority task to fix the login bug"
• "List all tasks in the PF project"
• "Show me task PF-123"
• "Mark task PF-123 as done"
• "Delete task PF-123"

**Projects:**
• "Create a new project called 'Website Redesign'"
• "List all projects"

Just ask me in natural language and I'll help you get things done!`

	default:
		return "I'm processing your request..."
	}
}

// requiresConfirmation determines if the action needs user confirmation
func (t *Translator) requiresConfirmation(parsed *ParsedRequest) bool {
	// Require confirmation for destructive actions
	switch parsed.Intent {
	case IntentDeleteTask:
		return true
	default:
		return false
	}
}

// ValidateParameters validates the parameters in a parsed request
func (t *Translator) ValidateParameters(parsed *ParsedRequest) error {
	switch parsed.Intent {
	case IntentCreateTask:
		if title, ok := parsed.Parameters["title"].(string); !ok || strings.TrimSpace(title) == "" {
			return fmt.Errorf("task title is required")
		}

		if priority, ok := parsed.Parameters["priority"].(string); ok {
			if !t.isValidPriority(priority) {
				return fmt.Errorf("invalid priority: %s", priority)
			}
		}

		if status, ok := parsed.Parameters["status"].(string); ok {
			if !t.isValidStatus(status) {
				return fmt.Errorf("invalid status: %s", status)
			}
		}

		if taskType, ok := parsed.Parameters["type"].(string); ok {
			if !t.isValidTaskType(taskType) {
				return fmt.Errorf("invalid task type: %s", taskType)
			}
		}

	case IntentUpdateTask, IntentReadTask, IntentDeleteTask:
		if taskID, ok := parsed.Parameters["task_id"].(string); !ok || strings.TrimSpace(taskID) == "" {
			return fmt.Errorf("task ID is required")
		}

	case IntentCreateProject:
		if name, ok := parsed.Parameters["name"].(string); !ok || strings.TrimSpace(name) == "" {
			return fmt.Errorf("project name is required")
		}
	}

	return nil
}

// isValidPriority checks if a priority value is valid
func (t *Translator) isValidPriority(priority string) bool {
	validPriorities := []string{string(PriorityLow), string(PriorityMedium), string(PriorityHigh), string(PriorityCritical)}
	for _, valid := range validPriorities {
		if strings.ToLower(priority) == valid {
			return true
		}
	}
	return false
}

// isValidStatus checks if a status value is valid
func (t *Translator) isValidStatus(status string) bool {
	validStatuses := []string{string(StatusTodo), string(StatusInProgress), string(StatusDone), string(StatusBlocked)}
	for _, valid := range validStatuses {
		if strings.ToLower(status) == valid {
			return true
		}
	}
	return false
}

// isValidTaskType checks if a task type is valid
func (t *Translator) isValidTaskType(taskType string) bool {
	validTypes := []string{string(TaskTypeEpic), string(TaskTypeStory), string(TaskTypeTask), string(TaskTypeSubtask)}
	for _, valid := range validTypes {
		if strings.ToLower(taskType) == valid {
			return true
		}
	}
	return false
}

// ExtractTaskID attempts to extract a task ID from various formats
func ExtractTaskID(text string) string {
	text = strings.TrimSpace(text)

	// Direct format: PF-123
	if strings.Contains(text, "-") && len(text) > 3 {
		parts := strings.Split(text, "-")
		if len(parts) == 2 {
			return text
		}
	}

	// Format: task 123, issue 123, #123
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, "task ", "")
	text = strings.ReplaceAll(text, "issue ", "")
	text = strings.ReplaceAll(text, "#", "")
	text = strings.TrimSpace(text)

	// If it's a number, assume PF project
	if len(text) > 0 && isNumeric(text) {
		return fmt.Sprintf("PF-%s", text)
	}

	return text
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// ParseDueDate attempts to parse various date formats
func ParseDueDate(dateStr string) (string, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return "", nil
	}

	// Handle relative dates
	dateStr = strings.ToLower(dateStr)
	now := time.Now()

	switch dateStr {
	case "today":
		return now.Format("2006-01-02"), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Format("2006-01-02"), nil
	case "next week":
		return now.AddDate(0, 0, 7).Format("2006-01-02"), nil
	case "next month":
		return now.AddDate(0, 1, 0).Format("2006-01-02"), nil
	}

	// Try to parse absolute dates
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02/01/2006",
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan 2006",
		"2 January 2006",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			return parsed.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("unable to parse date: %s", dateStr)
}
