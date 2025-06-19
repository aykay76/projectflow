package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aykay76/projectflow/internal/models"
)

// handleToolsCall handles tool call requests
func (s *MCPServer) handleToolsCall(request JSONRPCRequest) JSONRPCResponse {
	var toolCallReq ToolCallRequest

	// Parse the params
	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		return s.createErrorResponse(request.ID, -32602, "Invalid params", nil)
	}

	if err := json.Unmarshal(paramsBytes, &toolCallReq); err != nil {
		return s.createErrorResponse(request.ID, -32602, "Invalid params", nil)
	}

	// Handle the specific tool
	var result ToolCallResult
	var callErr error

	switch toolCallReq.Name {
	case "list_tasks":
		result, callErr = s.handleListTasks(toolCallReq.Arguments)
	case "create_task":
		result, callErr = s.handleCreateTask(toolCallReq.Arguments)
	case "get_task":
		result, callErr = s.handleGetTask(toolCallReq.Arguments)
	case "update_task":
		result, callErr = s.handleUpdateTask(toolCallReq.Arguments)
	case "delete_task":
		result, callErr = s.handleDeleteTask(toolCallReq.Arguments)
	case "get_task_hierarchy":
		result, callErr = s.handleGetTaskHierarchy(toolCallReq.Arguments)
	case "list_projects":
		result, callErr = s.handleListProjects(toolCallReq.Arguments)
	case "create_project":
		result, callErr = s.handleCreateProject(toolCallReq.Arguments)
	case "get_project":
		result, callErr = s.handleGetProject(toolCallReq.Arguments)
	case "update_project":
		result, callErr = s.handleUpdateProject(toolCallReq.Arguments)
	case "delete_project":
		result, callErr = s.handleDeleteProject(toolCallReq.Arguments)
	default:
		return s.createErrorResponse(request.ID, -32601, "Unknown tool", nil)
	}

	if callErr != nil {
		result = ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: %s", callErr.Error()),
			}},
			IsError: true,
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleListTasks handles the list_tasks tool call
func (s *MCPServer) handleListTasks(args map[string]interface{}) (ToolCallResult, error) {
	// project_id is now required
	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return ToolCallResult{}, fmt.Errorf("project_id is required")
	}

	tasks, err := s.storage.ListTasks(projectID)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to list tasks for project %s: %w", projectID, err)
	}

	tasksJSON, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Found %d tasks in project %s:\n\n%s", len(tasks), projectID, string(tasksJSON)),
		}},
	}, nil
}

// handleCreateTask handles the create_task tool call
func (s *MCPServer) handleCreateTask(args map[string]interface{}) (ToolCallResult, error) {
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return ToolCallResult{}, fmt.Errorf("title is required and must be a string")
	}

	description, _ := args["description"].(string)

	// Create new task
	task := models.NewTask(title, description)

	// Set project ID if provided, otherwise default to "PF"
	if projectID, ok := args["project_id"].(string); ok && projectID != "" {
		task.ProjectID = projectID
	} else {
		task.ProjectID = "PF" // Default to PF project
	}

	// Set optional fields
	if status, ok := args["status"].(string); ok && status != "" {
		if !models.IsValidStatus(status) {
			return ToolCallResult{}, fmt.Errorf("invalid status: %s", status)
		}
		task.Status = models.TaskStatus(status)
	}

	if priority, ok := args["priority"].(string); ok && priority != "" {
		if !models.IsValidPriority(priority) {
			return ToolCallResult{}, fmt.Errorf("invalid priority: %s", priority)
		}
		task.Priority = models.TaskPriority(priority)
	}

	if taskType, ok := args["type"].(string); ok && taskType != "" {
		if !models.IsValidType(taskType) {
			return ToolCallResult{}, fmt.Errorf("invalid type: %s", taskType)
		}
		task.Type = models.TaskType(taskType)
	}

	if parentID, ok := args["parent_id"].(string); ok && parentID != "" {
		task.ParentID = parentID
	}

	if dueDate, ok := args["due_date"].(string); ok && dueDate != "" {
		if err := task.SetDueDate(dueDate); err != nil {
			return ToolCallResult{}, fmt.Errorf("invalid due date format: %w", err)
		}
	}

	if err := s.storage.CreateTask(task); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to create task: %w", err)
	}

	taskJSON, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to marshal task: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Successfully created task:\n\n%s", string(taskJSON)),
		}},
	}, nil
}

// handleGetTask handles the get_task tool call
func (s *MCPServer) handleGetTask(args map[string]interface{}) (ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return ToolCallResult{}, fmt.Errorf("id is required and must be a string")
	}

	task, err := s.storage.GetTask(id)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to get task: %w", err)
	}

	taskJSON, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to marshal task: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(taskJSON),
		}},
	}, nil
}

// handleUpdateTask handles the update_task tool call
func (s *MCPServer) handleUpdateTask(args map[string]interface{}) (ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return ToolCallResult{}, fmt.Errorf("id is required and must be a string")
	}

	// Get existing task
	existingTask, err := s.storage.GetTask(id)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to get existing task: %w", err)
	}

	// Update fields
	task := *existingTask

	if title, ok := args["title"].(string); ok && title != "" {
		task.Title = title
	}

	if description, ok := args["description"].(string); ok {
		task.Description = description
	}

	if status, ok := args["status"].(string); ok && status != "" {
		if !models.IsValidStatus(status) {
			return ToolCallResult{}, fmt.Errorf("invalid status: %s", status)
		}
		task.Status = models.TaskStatus(status)

		// Auto-set start date if status changes to in_progress
		if task.Status == models.StatusInProgress && task.StartedAt == nil {
			now := time.Now()
			task.StartedAt = &now
		}
	}

	if priority, ok := args["priority"].(string); ok && priority != "" {
		if !models.IsValidPriority(priority) {
			return ToolCallResult{}, fmt.Errorf("invalid priority: %s", priority)
		}
		task.Priority = models.TaskPriority(priority)
	}

	if taskType, ok := args["type"].(string); ok && taskType != "" {
		if !models.IsValidType(taskType) {
			return ToolCallResult{}, fmt.Errorf("invalid type: %s", taskType)
		}
		task.Type = models.TaskType(taskType)
	}

	if parentID, ok := args["parent_id"].(string); ok {
		task.ParentID = parentID
	}

	if dueDate, ok := args["due_date"].(string); ok {
		if dueDate == "" {
			task.DueDate = nil
		} else {
			if err := task.SetDueDate(dueDate); err != nil {
				return ToolCallResult{}, fmt.Errorf("invalid due date format: %w", err)
			}
		}
	}

	task.UpdatedAt = time.Now()

	if err := s.storage.UpdateTask(&task); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to update task: %w", err)
	}

	taskJSON, err := json.MarshalIndent(&task, "", "  ")
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to marshal task: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Successfully updated task:\n\n%s", string(taskJSON)),
		}},
	}, nil
}

// handleDeleteTask handles the delete_task tool call
func (s *MCPServer) handleDeleteTask(args map[string]interface{}) (ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return ToolCallResult{}, fmt.Errorf("id is required and must be a string")
	}

	// Get task info before deletion for confirmation
	task, err := s.storage.GetTask(id)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to get task: %w", err)
	}

	if err := s.storage.DeleteTask(id); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to delete task: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Successfully deleted task: %s (%s)", task.Title, task.ID),
		}},
	}, nil
}

// handleGetTaskHierarchy handles the get_task_hierarchy tool call
func (s *MCPServer) handleGetTaskHierarchy(args map[string]interface{}) (ToolCallResult, error) {
	hierarchy, err := s.storage.GetTaskHierarchy()
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to get task hierarchy: %w", err)
	}

	hierarchyJSON, err := json.MarshalIndent(hierarchy, "", "  ")
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to marshal hierarchy: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Task hierarchy:\n\n%s", string(hierarchyJSON)),
		}},
	}, nil
}

// handleListProjects handles the list_projects tool call
func (s *MCPServer) handleListProjects(args map[string]interface{}) (ToolCallResult, error) {
	projects, err := s.storage.ListProjects()
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to list projects: %w", err)
	}

	projectsJSON, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to marshal projects: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Found %d projects:\n\n%s", len(projects), string(projectsJSON)),
		}},
	}, nil
}

// handleCreateProject handles the create_project tool call
func (s *MCPServer) handleCreateProject(args map[string]interface{}) (ToolCallResult, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return ToolCallResult{}, fmt.Errorf("name is required and must be a string")
	}

	description, _ := args["description"].(string)
	prefix, _ := args["prefix"].(string)

	// Default prefix if not provided
	if prefix == "" {
		prefix = strings.ToUpper(name[:2]) // Use first two letters of name as default
	}

	// Create new project
	project := models.NewProject(name, description, prefix)

	if err := s.storage.CreateProject(project); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to create project: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Successfully created project: %s (ID: %s, Prefix: %s)", project.Name, project.ID, project.DisplayPrefix),
		}},
	}, nil
}

// handleGetProject handles the get_project tool call
func (s *MCPServer) handleGetProject(args map[string]interface{}) (ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return ToolCallResult{}, fmt.Errorf("id is required and must be a string")
	}

	project, err := s.storage.GetProject(id)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to get project: %w", err)
	}

	projectJSON, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to marshal project: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Project details:\n\n%s", string(projectJSON)),
		}},
	}, nil
}

// handleUpdateProject handles the update_project tool call
func (s *MCPServer) handleUpdateProject(args map[string]interface{}) (ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return ToolCallResult{}, fmt.Errorf("id is required and must be a string")
	}

	// Get existing project
	project, err := s.storage.GetProject(id)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to get project: %w", err)
	}

	// Update fields if provided
	if name, ok := args["name"].(string); ok && name != "" {
		project.Name = name
	}
	if description, ok := args["description"].(string); ok {
		project.Description = description
	}
	if prefix, ok := args["prefix"].(string); ok {
		project.DisplayPrefix = prefix
	}

	project.UpdatedAt = time.Now()

	if err := s.storage.UpdateProject(project); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to update project: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Successfully updated project: %s (ID: %s)", project.Name, project.ID),
		}},
	}, nil
}

// handleDeleteProject handles the delete_project tool call
func (s *MCPServer) handleDeleteProject(args map[string]interface{}) (ToolCallResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return ToolCallResult{}, fmt.Errorf("id is required and must be a string")
	}

	// Get project details before deletion for confirmation message
	project, err := s.storage.GetProject(id)
	if err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.storage.DeleteProject(id); err != nil {
		return ToolCallResult{}, fmt.Errorf("failed to delete project: %w", err)
	}

	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("Successfully deleted project: %s (ID: %s)", project.Name, project.ID),
		}},
	}, nil
}
