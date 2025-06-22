package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aykay76/projectflow/internal/llm"
	"github.com/aykay76/projectflow/internal/models"
	"github.com/aykay76/projectflow/internal/storage"
	"github.com/aykay76/projectflow/internal/translator"
	"github.com/google/uuid"
)

// ChatRequest represents an incoming chat message
type ChatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// ChatResponse represents the response to a chat message
type ChatResponse struct {
	Response             string   `json:"response"`
	ActionsTaken         []string `json:"actions_taken"`
	TaskIDs              []string `json:"task_ids,omitempty"`
	ProjectIDs           []string `json:"project_ids,omitempty"`
	ConversationID       string   `json:"conversation_id"`
	RequiresConfirmation bool     `json:"requires_confirmation,omitempty"`
	Confidence           float64  `json:"confidence"`
	Intent               string   `json:"intent"`
}

// ConversationMessage represents a single message in a conversation
type ConversationMessage struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"` // "user" or "assistant"
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID       string                `json:"id"`
	Messages []ConversationMessage `json:"messages"`
	Created  time.Time             `json:"created"`
	Updated  time.Time             `json:"updated"`
}

// LLMService interface for the chat handler
type LLMService interface {
	IsEnabled() bool
	Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
	HealthCheck(ctx context.Context) error
}

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	storage       storage.Storage
	llmService    LLMService
	translator    *translator.Translator
	logger        *slog.Logger
	conversations map[string]*Conversation // In-memory for MVP, could be moved to storage later
}

// NewChatHandler creates a new chat handler instance
func NewChatHandler(storage storage.Storage, llmService LLMService, logger *slog.Logger) *ChatHandler {
	// Create translator
	translatorService := translator.NewTranslator(llmService, logger)

	return &ChatHandler{
		storage:       storage,
		llmService:    llmService,
		translator:    translatorService,
		logger:        logger,
		conversations: make(map[string]*Conversation),
	}
}

// HandleChat handles POST /api/chat requests
func (ch *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Failed to decode chat request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// Generate conversation ID if not provided
	if req.ConversationID == "" {
		req.ConversationID = uuid.New().String()
	}

	ch.logger.Info("Processing chat message",
		"conversation_id", req.ConversationID,
		"message_length", len(req.Message))

	// Get or create conversation
	conversation := ch.getOrCreateConversation(req.ConversationID)

	// Add user message to conversation
	userMessage := ConversationMessage{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now().UTC(),
	}
	conversation.Messages = append(conversation.Messages, userMessage)

	// Translate natural language to actions
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	translationResult, err := ch.translator.Translate(ctx, req.Message)
	if err != nil {
		ch.logger.Error("Translation failed", "error", err, "conversation_id", req.ConversationID)

		// Return error response but still maintain conversation
		errorResponse := &ChatResponse{
			Response:       "I'm sorry, I'm having trouble understanding your request right now. Please try again.",
			ActionsTaken:   []string{},
			ConversationID: req.ConversationID,
			Confidence:     0.0,
			Intent:         "error",
		}

		ch.respondWithJSON(w, errorResponse)
		return
	}

	// Execute MCP commands and collect results
	actionsTaken, taskIDs, projectIDs, err := ch.executeMCPCommands(translationResult.MCPCommands)
	if err != nil {
		ch.logger.Error("Failed to execute MCP commands", "error", err, "conversation_id", req.ConversationID)

		// Return error response
		errorResponse := &ChatResponse{
			Response:       fmt.Sprintf("I understood what you want to do, but I encountered an error: %s", err.Error()),
			ActionsTaken:   []string{},
			ConversationID: req.ConversationID,
			Confidence:     translationResult.ParsedRequest.Confidence,
			Intent:         string(translationResult.ParsedRequest.Intent),
		}

		ch.respondWithJSON(w, errorResponse)
		return
	}

	// Enhance response with actual data if needed
	enhancedResponse := ch.enhanceResponseWithData(translationResult.HumanResponse, actionsTaken, taskIDs, projectIDs)

	// Create response
	response := &ChatResponse{
		Response:             enhancedResponse,
		ActionsTaken:         actionsTaken,
		TaskIDs:              taskIDs,
		ProjectIDs:           projectIDs,
		ConversationID:       req.ConversationID,
		RequiresConfirmation: translationResult.RequiresConfirmation,
		Confidence:           translationResult.ParsedRequest.Confidence,
		Intent:               string(translationResult.ParsedRequest.Intent),
	}

	// Add assistant message to conversation
	assistantMessage := ConversationMessage{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   response.Response,
		Timestamp: time.Now().UTC(),
		Metadata: map[string]interface{}{
			"actions_taken": actionsTaken,
			"task_ids":      taskIDs,
			"project_ids":   projectIDs,
			"confidence":    response.Confidence,
			"intent":        response.Intent,
		},
	}
	conversation.Messages = append(conversation.Messages, assistantMessage)
	conversation.Updated = time.Now().UTC()

	ch.logger.Info("Chat message processed successfully",
		"conversation_id", req.ConversationID,
		"intent", response.Intent,
		"confidence", response.Confidence,
		"actions_count", len(actionsTaken))

	ch.respondWithJSON(w, response)
}

// HandleChatHistory handles GET /api/chat/history requests
func (ch *ChatHandler) HandleChatHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		http.Error(w, "conversation_id parameter is required", http.StatusBadRequest)
		return
	}

	conversation, exists := ch.conversations[conversationID]
	if !exists {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	ch.respondWithJSON(w, conversation)
}

// getOrCreateConversation gets an existing conversation or creates a new one
func (ch *ChatHandler) getOrCreateConversation(conversationID string) *Conversation {
	if conversation, exists := ch.conversations[conversationID]; exists {
		return conversation
	}

	conversation := &Conversation{
		ID:       conversationID,
		Messages: []ConversationMessage{},
		Created:  time.Now().UTC(),
		Updated:  time.Now().UTC(),
	}

	ch.conversations[conversationID] = conversation
	return conversation
}

// executeMCPCommands executes the translated MCP commands and returns the results
func (ch *ChatHandler) executeMCPCommands(commands []translator.MCPCommand) ([]string, []string, []string, error) {
	var actionsTaken []string
	var taskIDs []string
	var projectIDs []string

	for _, cmd := range commands {
		switch cmd.Method {
		case "create_task":
			taskID, err := ch.executeCreateTask(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to create task: %w", err)
			}
			actionsTaken = append(actionsTaken, "create_task")
			taskIDs = append(taskIDs, taskID)

		case "update_task":
			taskID, err := ch.executeUpdateTask(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to update task: %w", err)
			}
			actionsTaken = append(actionsTaken, "update_task")
			taskIDs = append(taskIDs, taskID)

		case "get_task":
			taskID, err := ch.executeGetTask(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to get task: %w", err)
			}
			actionsTaken = append(actionsTaken, "get_task")
			taskIDs = append(taskIDs, taskID)

		case "delete_task":
			taskID, err := ch.executeDeleteTask(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to delete task: %w", err)
			}
			actionsTaken = append(actionsTaken, "delete_task")
			taskIDs = append(taskIDs, taskID)

		case "list_tasks":
			tasks, projectID, err := ch.executeListTasks(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to list tasks: %w", err)
			}
			actionsTaken = append(actionsTaken, "list_tasks")
			for _, task := range tasks {
				taskIDs = append(taskIDs, task.DisplayID)
			}
			if projectID != "" {
				projectIDs = append(projectIDs, projectID)
			}

		case "create_project":
			projectID, err := ch.executeCreateProject(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to create project: %w", err)
			}
			actionsTaken = append(actionsTaken, "create_project")
			projectIDs = append(projectIDs, projectID)

		case "update_project":
			projectID, err := ch.executeUpdateProject(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to update project: %w", err)
			}
			actionsTaken = append(actionsTaken, "update_project")
			projectIDs = append(projectIDs, projectID)

		case "get_project":
			projectID, err := ch.executeGetProject(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to get project: %w", err)
			}
			actionsTaken = append(actionsTaken, "get_project")
			projectIDs = append(projectIDs, projectID)

		case "delete_project":
			projectID, err := ch.executeDeleteProject(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to delete project: %w", err)
			}
			actionsTaken = append(actionsTaken, "delete_project")
			projectIDs = append(projectIDs, projectID)

		case "list_projects":
			projects, err := ch.executeListProjects(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to list projects: %w", err)
			}
			actionsTaken = append(actionsTaken, "list_projects")
			for _, project := range projects {
				projectIDs = append(projectIDs, project.DisplayPrefix)
			}

		case "get_task_hierarchy":
			err := ch.executeGetTaskHierarchy(cmd.Parameters)
			if err != nil {
				return actionsTaken, taskIDs, projectIDs, fmt.Errorf("failed to get task hierarchy: %w", err)
			}
			actionsTaken = append(actionsTaken, "get_task_hierarchy")

		default:
			ch.logger.Warn("Unknown MCP command", "method", cmd.Method)
			return actionsTaken, taskIDs, projectIDs, fmt.Errorf("unknown command: %s", cmd.Method)
		}
	}

	return actionsTaken, taskIDs, projectIDs, nil
}

// executeCreateTask handles task creation
func (ch *ChatHandler) executeCreateTask(params map[string]interface{}) (string, error) {
	// Extract parameters
	title, ok := params["title"].(string)
	if !ok || title == "" {
		return "", fmt.Errorf("title is required")
	}

	// Optional parameters with defaults
	description := ""
	if d, ok := params["description"].(string); ok {
		description = d
	}

	status := "todo"
	if s, ok := params["status"].(string); ok {
		status = s
	}

	priority := "medium"
	if p, ok := params["priority"].(string); ok {
		priority = p
	}

	taskType := "task"
	if t, ok := params["type"].(string); ok {
		taskType = t
	}

	projectID := "PF" // Default project
	if p, ok := params["project_id"].(string); ok {
		projectID = p
	}

	// Get project to validate it exists
	project, err := ch.storage.GetProjectByDisplayPrefix(projectID)
	if err != nil {
		return "", fmt.Errorf("project not found: %s", projectID)
	}

	// Generate display ID
	displayID, err := ch.storage.GetNextDisplayID(project.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate display ID: %w", err)
	}

	// Parse due date if provided
	var dueDate *time.Time
	if dueDateStr, ok := params["due_date"].(string); ok && dueDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dueDateStr); err == nil {
			dueDate = &parsed
		}
	}

	// Create task
	task := &models.Task{
		ID:          uuid.New().String(),
		DisplayID:   displayID,
		ProjectID:   project.ID,
		Title:       title,
		Description: description,
		Status:      models.TaskStatus(status),
		Priority:    models.TaskPriority(priority),
		Type:        models.TaskType(taskType),
		DueDate:     dueDate,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Handle parent task if specified
	if parentID, ok := params["parent_id"].(string); ok && parentID != "" {
		// Try to find parent by display ID first, then by UUID
		parent, err := ch.storage.GetTaskByDisplayID(parentID)
		if err != nil {
			parent, err = ch.storage.GetTask(parentID)
			if err != nil {
				return "", fmt.Errorf("parent task not found: %s", parentID)
			}
		}
		task.ParentID = parent.ID
	}

	err = ch.storage.CreateTask(task)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	ch.logger.Info("Task created via chat", "task_id", task.DisplayID, "title", task.Title)
	return task.DisplayID, nil
}

// executeUpdateTask handles task updates
func (ch *ChatHandler) executeUpdateTask(params map[string]interface{}) (string, error) {
	// Get task ID (required)
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("task id is required")
	}

	// Try to find task by display ID first, then by UUID
	task, err := ch.storage.GetTaskByDisplayID(id)
	if err != nil {
		task, err = ch.storage.GetTask(id)
		if err != nil {
			return "", fmt.Errorf("task not found: %s", id)
		}
	}

	// Update fields if provided
	if title, ok := params["title"].(string); ok && title != "" {
		task.Title = title
	}
	if description, ok := params["description"].(string); ok {
		task.Description = description
	}
	if status, ok := params["status"].(string); ok && status != "" {
		task.Status = models.TaskStatus(status)
	}
	if priority, ok := params["priority"].(string); ok && priority != "" {
		task.Priority = models.TaskPriority(priority)
	}
	if taskType, ok := params["type"].(string); ok && taskType != "" {
		task.Type = models.TaskType(taskType)
	}

	// Handle due date
	if dueDateStr, ok := params["due_date"].(string); ok {
		if dueDateStr == "" {
			task.DueDate = nil
		} else if parsed, err := time.Parse("2006-01-02", dueDateStr); err == nil {
			task.DueDate = &parsed
		}
	}

	// Handle parent task
	if parentID, ok := params["parent_id"].(string); ok {
		if parentID == "" {
			task.ParentID = ""
		} else {
			parent, err := ch.storage.GetTaskByDisplayID(parentID)
			if err != nil {
				parent, err = ch.storage.GetTask(parentID)
				if err != nil {
					return "", fmt.Errorf("parent task not found: %s", parentID)
				}
			}
			task.ParentID = parent.ID
		}
	}

	task.UpdatedAt = time.Now().UTC()

	err = ch.storage.UpdateTask(task)
	if err != nil {
		return "", fmt.Errorf("failed to update task: %w", err)
	}

	ch.logger.Info("Task updated via chat", "task_id", task.DisplayID, "title", task.Title)
	return task.DisplayID, nil
}

// executeGetTask handles task retrieval
func (ch *ChatHandler) executeGetTask(params map[string]interface{}) (string, error) {
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("task id is required")
	}

	// Try to find task by display ID first, then by UUID
	task, err := ch.storage.GetTaskByDisplayID(id)
	if err != nil {
		task, err = ch.storage.GetTask(id)
		if err != nil {
			return "", fmt.Errorf("task not found: %s", id)
		}
	}

	ch.logger.Debug("Task retrieved via chat", "task_id", task.DisplayID, "title", task.Title)
	return task.DisplayID, nil
}

// executeDeleteTask handles task deletion
func (ch *ChatHandler) executeDeleteTask(params map[string]interface{}) (string, error) {
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("task id is required")
	}

	// Try to find task by display ID first, then by UUID
	task, err := ch.storage.GetTaskByDisplayID(id)
	if err != nil {
		task, err = ch.storage.GetTask(id)
		if err != nil {
			return "", fmt.Errorf("task not found: %s", id)
		}
	}

	err = ch.storage.DeleteTask(task.ID)
	if err != nil {
		return "", fmt.Errorf("failed to delete task: %w", err)
	}

	ch.logger.Info("Task deleted via chat", "task_id", task.DisplayID, "title", task.Title)
	return task.DisplayID, nil
}

// executeListTasks handles task listing
func (ch *ChatHandler) executeListTasks(params map[string]interface{}) ([]*models.Task, string, error) {
	projectID := ""
	if p, ok := params["project_id"].(string); ok && p != "" {
		// Try to get project by display prefix first
		project, err := ch.storage.GetProjectByDisplayPrefix(p)
		if err != nil {
			// Try by UUID
			project, err = ch.storage.GetProject(p)
			if err != nil {
				return nil, "", fmt.Errorf("project not found: %s", p)
			}
		}
		projectID = project.ID
	}

	tasks, err := ch.storage.ListTasks(projectID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list tasks: %w", err)
	}

	ch.logger.Debug("Tasks listed via chat", "project_id", projectID, "count", len(tasks))
	return tasks, projectID, nil
}

// executeCreateProject handles project creation
func (ch *ChatHandler) executeCreateProject(params map[string]interface{}) (string, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("project name is required")
	}

	description := ""
	if d, ok := params["description"].(string); ok {
		description = d
	}

	// Generate prefix from name if not provided
	prefix := ""
	if p, ok := params["prefix"].(string); ok && p != "" {
		prefix = strings.ToUpper(p)
	} else {
		// Generate prefix from name (first 3 chars, uppercase)
		prefix = strings.ToUpper(name)
		if len(prefix) > 3 {
			prefix = prefix[:3]
		}
	}

	project := &models.Project{
		ID:            uuid.New().String(),
		Name:          name,
		Description:   description,
		DisplayPrefix: prefix,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	err := ch.storage.CreateProject(project)
	if err != nil {
		return "", fmt.Errorf("failed to create project: %w", err)
	}

	ch.logger.Info("Project created via chat", "project_prefix", project.DisplayPrefix, "name", project.Name)
	return project.DisplayPrefix, nil
}

// executeUpdateProject handles project updates
func (ch *ChatHandler) executeUpdateProject(params map[string]interface{}) (string, error) {
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("project id is required")
	}

	// Try to find project by display prefix first, then by UUID
	project, err := ch.storage.GetProjectByDisplayPrefix(id)
	if err != nil {
		project, err = ch.storage.GetProject(id)
		if err != nil {
			return "", fmt.Errorf("project not found: %s", id)
		}
	}

	// Update fields if provided
	if name, ok := params["name"].(string); ok && name != "" {
		project.Name = name
	}
	if description, ok := params["description"].(string); ok {
		project.Description = description
	}
	if prefix, ok := params["prefix"].(string); ok && prefix != "" {
		project.DisplayPrefix = strings.ToUpper(prefix)
	}

	project.UpdatedAt = time.Now().UTC()

	err = ch.storage.UpdateProject(project)
	if err != nil {
		return "", fmt.Errorf("failed to update project: %w", err)
	}

	ch.logger.Info("Project updated via chat", "project_prefix", project.DisplayPrefix, "name", project.Name)
	return project.DisplayPrefix, nil
}

// executeGetProject handles project retrieval
func (ch *ChatHandler) executeGetProject(params map[string]interface{}) (string, error) {
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("project id is required")
	}

	// Try to find project by display prefix first, then by UUID
	project, err := ch.storage.GetProjectByDisplayPrefix(id)
	if err != nil {
		project, err = ch.storage.GetProject(id)
		if err != nil {
			return "", fmt.Errorf("project not found: %s", id)
		}
	}

	ch.logger.Debug("Project retrieved via chat", "project_prefix", project.DisplayPrefix, "name", project.Name)
	return project.DisplayPrefix, nil
}

// executeDeleteProject handles project deletion
func (ch *ChatHandler) executeDeleteProject(params map[string]interface{}) (string, error) {
	id, ok := params["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("project id is required")
	}

	// Try to find project by display prefix first, then by UUID
	project, err := ch.storage.GetProjectByDisplayPrefix(id)
	if err != nil {
		project, err = ch.storage.GetProject(id)
		if err != nil {
			return "", fmt.Errorf("project not found: %s", id)
		}
	}

	err = ch.storage.DeleteProject(project.ID)
	if err != nil {
		return "", fmt.Errorf("failed to delete project: %w", err)
	}

	ch.logger.Info("Project deleted via chat", "project_prefix", project.DisplayPrefix, "name", project.Name)
	return project.DisplayPrefix, nil
}

// executeListProjects handles project listing
func (ch *ChatHandler) executeListProjects(params map[string]interface{}) ([]*models.Project, error) {
	projects, err := ch.storage.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	ch.logger.Debug("Projects listed via chat", "count", len(projects))
	return projects, nil
}

// executeGetTaskHierarchy handles task hierarchy retrieval
func (ch *ChatHandler) executeGetTaskHierarchy(params map[string]interface{}) error {
	_, err := ch.storage.GetTaskHierarchy()
	if err != nil {
		return fmt.Errorf("failed to get task hierarchy: %w", err)
	}

	ch.logger.Debug("Task hierarchy retrieved via chat")
	return nil
}

// enhanceResponseWithData improves the response by including actual data
func (ch *ChatHandler) enhanceResponseWithData(response string, actionsTaken, taskIDs, projectIDs []string) string {
	// If projects were listed, enhance with actual project information
	for _, action := range actionsTaken {
		if action == "list_projects" && len(projectIDs) > 0 {
			// Get project details for the listed project IDs
			projects, err := ch.storage.ListProjects()
			if err == nil {
				// Build a nice project list
				var projectList []string
				for _, project := range projects {
					projectList = append(projectList, fmt.Sprintf("• **%s** (%s): %s",
						project.Name, project.DisplayPrefix, project.Description))
				}

				if len(projectList) > 0 {
					enhancedResponse := "Here are your projects:\n\n" + strings.Join(projectList, "\n")
					if len(projects) == 1 {
						enhancedResponse = "Here is your project:\n\n" + strings.Join(projectList, "\n")
					}
					return enhancedResponse
				}
			}
		}

		if action == "list_tasks" && len(taskIDs) > 0 {
			// Get task details for the listed task IDs
			var taskList []string
			for _, taskID := range taskIDs {
				task, err := ch.storage.GetTaskByDisplayID(taskID)
				if err == nil {
					// Format task info
					statusEmoji := "📋"
					switch task.Status {
					case "todo":
						statusEmoji = "📋"
					case "in_progress":
						statusEmoji = "🔄"
					case "done":
						statusEmoji = "✅"
					case "blocked":
						statusEmoji = "🚫"
					}

					priorityInfo := ""
					switch task.Priority {
					case "critical":
						priorityInfo = " 🔴"
					case "high":
						priorityInfo = " 🟡"
					case "low":
						priorityInfo = " 🔵"
					}

					dueDateInfo := ""
					if task.DueDate != nil {
						dueDateInfo = fmt.Sprintf(" (due %s)", task.DueDate.Format("2006-01-02"))
					}

					taskList = append(taskList, fmt.Sprintf("• %s **%s**%s%s: %s",
						statusEmoji, task.DisplayID, priorityInfo, dueDateInfo, task.Title))
				}
			}

			if len(taskList) > 0 {
				enhancedResponse := "Here are your tasks:\n\n" + strings.Join(taskList, "\n")
				if len(taskList) == 1 {
					enhancedResponse = "Here is your task:\n\n" + strings.Join(taskList, "\n")
				}
				return enhancedResponse
			}
		}
	}

	return response
}

// respondWithJSON sends a JSON response
func (ch *ChatHandler) respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		ch.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
