package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/aykay76/projectflow/internal/logger"
	"github.com/aykay76/projectflow/internal/metrics"
	"github.com/aykay76/projectflow/internal/middleware"
	"github.com/aykay76/projectflow/internal/models"
	"github.com/aykay76/projectflow/internal/storage"
)

// Handler handles HTTP requests
type Handler struct {
	storage   storage.Storage
	templates *template.Template
}

// NewHandler creates a new handler instance
func NewHandler(storage storage.Storage) *Handler {
	// Load templates
	templates := template.Must(template.ParseGlob("web/templates/*.html"))

	return &Handler{
		storage:   storage,
		templates: templates,
	}
}

// getTaskByIdentifier resolves a task by either UUID or display ID
func (h *Handler) getTaskByIdentifier(identifier string) (*models.Task, error) {
	// Check if identifier looks like a UUID (36 characters with dashes)
	if len(identifier) == 36 && strings.Count(identifier, "-") == 4 {
		// Try UUID lookup first for performance
		task, err := h.storage.GetTask(identifier)
		if err == nil {
			return task, nil
		}
	}

	// If not a UUID or UUID lookup failed, and identifier looks like a display ID, try display ID lookup
	if models.IsValidDisplayID(identifier) {
		return h.storage.GetTaskByDisplayID(identifier)
	}

	// If it doesn't look like a display ID, try UUID lookup anyway (in case the UUID check above failed)
	return h.storage.GetTask(identifier)
}

// HandleIndex serves the main web interface
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Default to showing tasks from the "PF" project
	tasks, err := h.storage.ListTasks("PF")
	if err != nil {
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	data := struct {
		Tasks []*models.Task
		Title string
	}{
		Tasks: tasks,
		Title: "ProjectFlow - Task Management",
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// HandleHierarchy handles /api/hierarchy endpoint
func (h *Handler) HandleHierarchy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hierarchyTasks, err := h.storage.GetTaskHierarchy()
	if err != nil {
		http.Error(w, "Failed to get task hierarchy", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(hierarchyTasks)
}

// HandleTasks handles /api/tasks endpoint
func (h *Handler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTask handles /api/tasks/{id} endpoint
func (h *Handler) HandleTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract task ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID := strings.Split(path, "/")[0]

	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, taskID)
	case http.MethodPut:
		h.updateTask(w, r, taskID)
	case http.MethodDelete:
		h.deleteTask(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	// Get project_id from query parameters, default to "PF"
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		projectID = "PF"
	}

	logger.InfoContext(ctx, "Listing tasks", "project_id", projectID, "request_id", requestID)

	tasks, err := h.storage.ListTasks(projectID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list tasks", "error", err, "project_id", projectID, "request_id", requestID)
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Tasks loaded successfully", "count", len(tasks), "project_id", projectID, "request_id", requestID)
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	logger.InfoContext(ctx, "Creating new task", "request_id", requestID)

	// Use a temporary struct to handle due_date and started_at as strings
	var taskCreate struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		Type        string `json:"type"`
		ParentID    string `json:"parent_id"`
		ProjectID   string `json:"project_id"`  // Add project_id field
		DueDate     string `json:"due_date"`
		StartedAt   string `json:"started_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskCreate); err != nil {
		logger.WarnContext(ctx, "Invalid JSON in request body", "error", err, "request_id", requestID)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create task struct and populate fields
	var task models.Task
	task.Title = taskCreate.Title
	task.Description = taskCreate.Description
	task.ParentID = taskCreate.ParentID
	task.ProjectID = taskCreate.ProjectID  // Set project_id from request

	// Handle due_date
	if taskCreate.DueDate != "" {
		if err := task.SetDueDate(taskCreate.DueDate); err != nil {
			logger.WarnContext(ctx, "Invalid due date format", "error", err, "due_date", taskCreate.DueDate, "request_id", requestID)
			http.Error(w, "Invalid due date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	// Handle started_at
	if taskCreate.StartedAt != "" {
		if err := task.SetStartedAt(taskCreate.StartedAt); err != nil {
			logger.WarnContext(ctx, "Invalid start date format", "error", err, "started_at", taskCreate.StartedAt, "request_id", requestID)
			http.Error(w, "Invalid start date format. Use RFC3339", http.StatusBadRequest)
			return
		}
	}

	// Validate required fields
	if task.Title == "" {
		logger.WarnContext(ctx, "Task title is required", "request_id", requestID)
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Set defaults if not provided
	if taskCreate.Status == "" {
		task.Status = models.StatusTodo
	} else {
		task.Status = models.TaskStatus(taskCreate.Status)
		// Auto-set start date if status is in_progress and no start date provided
		if task.Status == models.StatusInProgress && taskCreate.StartedAt == "" {
			task.StartTask()
		}
	}
	if taskCreate.Priority == "" {
		task.Priority = models.PriorityMedium
	} else {
		task.Priority = models.TaskPriority(taskCreate.Priority)
	}
	if taskCreate.Type == "" {
		task.Type = models.TypeTask
	} else {
		task.Type = models.TaskType(taskCreate.Type)
	}

	// Validate enum values
	if !models.IsValidStatus(string(task.Status)) {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}
	if !models.IsValidPriority(string(task.Priority)) {
		http.Error(w, "Invalid priority", http.StatusBadRequest)
		return
	}
	if !models.IsValidType(string(task.Type)) {
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	// Set timestamps
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Initialize children slice
	if task.Children == nil {
		task.Children = []string{}
	}

	if err := h.storage.CreateTask(&task); err != nil {
		// Record failed task creation
		if m, ok := metrics.FromContext(ctx); ok {
			m.RecordTaskOperation("create", "failed")
			m.RecordStorageOperation("create", "failed")
		}
		logger.ErrorContext(ctx, "Failed to create task in storage", "error", err, "request_id", requestID, "task_title", task.Title)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Record successful task creation metrics
	if m, ok := metrics.FromContext(ctx); ok {
		m.RecordTaskOperation("create", "success")
		m.RecordStorageOperation("create", "success")
	}

	logger.InfoContext(ctx, "Task created successfully", "request_id", requestID, "task_id", task.ID, "task_title", task.Title)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request, taskID string) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	logger.DebugContext(ctx, "Getting task", "request_id", requestID, "task_id", taskID)

	task, err := h.getTaskByIdentifier(taskID)
	if err != nil {
		// Record failed task retrieval
		if m, ok := metrics.FromContext(ctx); ok {
			m.RecordTaskOperation("get", "failed")
			m.RecordStorageOperation("get", "failed")
		}
		if strings.Contains(err.Error(), "not found") {
			logger.WarnContext(ctx, "Task not found", "request_id", requestID, "task_id", taskID)
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			logger.ErrorContext(ctx, "Failed to get task from storage", "error", err, "request_id", requestID, "task_id", taskID)
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	// Record successful task retrieval
	if m, ok := metrics.FromContext(ctx); ok {
		m.RecordTaskOperation("get", "success")
		m.RecordStorageOperation("get", "success")
	}

	logger.DebugContext(ctx, "Task retrieved successfully", "request_id", requestID, "task_id", taskID, "task_title", task.Title)
	json.NewEncoder(w).Encode(task)
}

// HandleTaskByDisplayID handles /api/tasks/by-display-id/{display_id} endpoint
func (h *Handler) HandleTaskByDisplayID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	// Extract display ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/by-display-id/")
	displayID := strings.Split(path, "/")[0]

	if displayID == "" {
		http.Error(w, "Display ID required", http.StatusBadRequest)
		return
	}

	logger.DebugContext(ctx, "Getting task by display ID", "request_id", requestID, "display_id", displayID)

	task, err := h.storage.GetTaskByDisplayID(displayID)
	if err != nil {
		// Record failed task retrieval
		if m, ok := metrics.FromContext(ctx); ok {
			m.RecordTaskOperation("get", "failed")
			m.RecordStorageOperation("get", "failed")
		}
		if strings.Contains(err.Error(), "not found") {
			logger.WarnContext(ctx, "Task not found by display ID", "request_id", requestID, "display_id", displayID)
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			logger.ErrorContext(ctx, "Failed to get task by display ID from storage", "error", err, "request_id", requestID, "display_id", displayID)
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	// Record successful task retrieval
	if m, ok := metrics.FromContext(ctx); ok {
		m.RecordTaskOperation("get", "success")
		m.RecordStorageOperation("get", "success")
	}

	logger.DebugContext(ctx, "Task retrieved successfully by display ID", "request_id", requestID, "display_id", displayID, "task_id", task.ID, "task_title", task.Title)
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, taskID string) {
	// First get the existing task
	existingTask, err := h.getTaskByIdentifier(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	// Use a temporary struct to handle due_date and started_at as strings
	var taskUpdate struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		Type        string `json:"type"`
		ParentID    string `json:"parent_id"`
		DueDate     string `json:"due_date"`
		StartedAt   string `json:"started_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Start with existing task and only update provided fields
	task := *existingTask
	if taskUpdate.Title != "" {
		task.Title = taskUpdate.Title
	}
	if taskUpdate.Description != "" {
		task.Description = taskUpdate.Description
	}
	if taskUpdate.Status != "" {
		task.Status = models.TaskStatus(taskUpdate.Status)
	}
	if taskUpdate.Priority != "" {
		task.Priority = models.TaskPriority(taskUpdate.Priority)
	}
	if taskUpdate.Type != "" {
		task.Type = models.TaskType(taskUpdate.Type)
	}
	if taskUpdate.ParentID != "" {
		task.ParentID = taskUpdate.ParentID
	}

	// Handle due_date
	if taskUpdate.DueDate != "" {
		if err := task.SetDueDate(taskUpdate.DueDate); err != nil {
			http.Error(w, "Invalid due date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	// Handle started_at
	if taskUpdate.StartedAt != "" {
		if err := task.SetStartedAt(taskUpdate.StartedAt); err != nil {
			http.Error(w, "Invalid start date format. Use RFC3339", http.StatusBadRequest)
			return
		}
	}

	// Auto-set start date if status changes to in_progress and no start date provided
	if task.Status == models.StatusInProgress && task.StartedAt == nil {
		now := time.Now()
		task.StartedAt = &now
	}

	// Update timestamp but keep the original UUID (don't overwrite with display ID)
	task.UpdatedAt = time.Now()

	// Validate enum values if provided
	if taskUpdate.Status != "" && !models.IsValidStatus(string(task.Status)) {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}
	if taskUpdate.Priority != "" && !models.IsValidPriority(string(task.Priority)) {
		http.Error(w, "Invalid priority", http.StatusBadRequest)
		return
	}
	if taskUpdate.Type != "" && !models.IsValidType(string(task.Type)) {
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateTask(&task); err != nil {
		// Record failed task update
		if m, ok := metrics.FromContext(r.Context()); ok {
			m.RecordTaskOperation("update", "failed")
			m.RecordStorageOperation("update", "failed")
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
		}
		return
	}

	// Record successful task update
	if m, ok := metrics.FromContext(r.Context()); ok {
		m.RecordTaskOperation("update", "success")
		m.RecordStorageOperation("update", "success")
	}

	json.NewEncoder(w).Encode(task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, taskID string) {
	ctx := r.Context()

	// First resolve the task to get the UUID (needed for deletion)
	task, err := h.getTaskByIdentifier(taskID)
	if err != nil {
		// Record failed task deletion
		if m, ok := metrics.FromContext(ctx); ok {
			m.RecordTaskOperation("delete", "failed")
			m.RecordStorageOperation("delete", "failed")
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find task", http.StatusInternalServerError)
		}
		return
	}

	if err := h.storage.DeleteTask(task.ID); err != nil {
		// Record failed task deletion
		if m, ok := metrics.FromContext(ctx); ok {
			m.RecordTaskOperation("delete", "failed")
			m.RecordStorageOperation("delete", "failed")
		}
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	// Record successful task deletion
	if m, ok := metrics.FromContext(ctx); ok {
		m.RecordTaskOperation("delete", "success")
		m.RecordStorageOperation("delete", "success")
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleTaskChildren handles /api/tasks/{id}/children endpoint
func (h *Handler) HandleTaskChildren(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract task ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "children" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	parentID := parts[0]

	if parentID == "" {
		http.Error(w, "Parent task ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTaskChildren(w, r, parentID)
	case http.MethodPost:
		h.addTaskChild(w, r, parentID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTaskChildRelation handles /api/tasks/{id}/children/{child_id} endpoint
func (h *Handler) HandleTaskChildRelation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract parent and child IDs from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "children" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	parentID := parts[0]
	childID := parts[2]

	if parentID == "" || childID == "" {
		http.Error(w, "Parent and child task IDs required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		h.removeTaskChild(w, r, parentID, childID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTaskMove handles /api/tasks/{id}/move endpoint
func (h *Handler) HandleTaskMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract task ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "move" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	taskID := parts[0]

	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		h.moveTask(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getTaskChildren(w http.ResponseWriter, r *http.Request, parentID string) {
	children, err := h.storage.GetTaskChildren(parentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Parent task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task children", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(children)
}

func (h *Handler) addTaskChild(w http.ResponseWriter, r *http.Request, parentID string) {
	var request struct {
		ChildID string `json:"child_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.ChildID == "" {
		http.Error(w, "child_id is required", http.StatusBadRequest)
		return
	}

	// Verify both tasks exist
	parentTask, err := h.storage.GetTask(parentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Parent task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get parent task", http.StatusInternalServerError)
		}
		return
	}

	childTask, err := h.storage.GetTask(request.ChildID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Child task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get child task", http.StatusInternalServerError)
		}
		return
	}

	// Check for circular references
	if h.wouldCreateCircularReference(parentID, request.ChildID) {
		http.Error(w, "Operation would create circular reference", http.StatusBadRequest)
		return
	}

	// Remove child from its current parent if it has one
	if childTask.ParentID != "" {
		currentParent, err := h.storage.GetTask(childTask.ParentID)
		if err == nil {
			currentParent.RemoveChild(request.ChildID)
			h.storage.UpdateTask(currentParent)
		}
	}

	// Add child to new parent
	parentTask.AddChild(request.ChildID)
	childTask.ParentID = parentID
	childTask.UpdatedAt = time.Now()

	// Update both tasks
	if err := h.storage.UpdateTask(parentTask); err != nil {
		http.Error(w, "Failed to update parent task", http.StatusInternalServerError)
		return
	}

	if err := h.storage.UpdateTask(childTask); err != nil {
		http.Error(w, "Failed to update child task", http.StatusInternalServerError)
		return
	}

	response := struct {
		Message string       `json:"message"`
		Parent  *models.Task `json:"parent"`
		Child   *models.Task `json:"child"`
	}{
		Message: "Child relationship created successfully",
		Parent:  parentTask,
		Child:   childTask,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) removeTaskChild(w http.ResponseWriter, r *http.Request, parentID, childID string) {
	// Verify parent task exists
	parentTask, err := h.storage.GetTask(parentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Parent task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get parent task", http.StatusInternalServerError)
		}
		return
	}

	// Verify child task exists
	childTask, err := h.storage.GetTask(childID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Child task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get child task", http.StatusInternalServerError)
		}
		return
	}

	// Check if the relationship actually exists
	childExists := false
	for _, child := range parentTask.Children {
		if child == childID {
			childExists = true
			break
		}
	}

	if !childExists {
		http.Error(w, "Child relationship does not exist", http.StatusBadRequest)
		return
	}

	// Remove the relationship
	parentTask.RemoveChild(childID)
	childTask.ParentID = ""
	childTask.UpdatedAt = time.Now()

	// Update both tasks
	if err := h.storage.UpdateTask(parentTask); err != nil {
		http.Error(w, "Failed to update parent task", http.StatusInternalServerError)
		return
	}

	if err := h.storage.UpdateTask(childTask); err != nil {
		http.Error(w, "Failed to update child task", http.StatusInternalServerError)
		return
	}

	response := struct {
		Message string       `json:"message"`
		Parent  *models.Task `json:"parent"`
		Child   *models.Task `json:"child"`
	}{
		Message: "Child relationship removed successfully",
		Parent:  parentTask,
		Child:   childTask,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) moveTask(w http.ResponseWriter, r *http.Request, taskID string) {
	var request struct {
		NewParentID string `json:"new_parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get the task to move
	task, err := h.storage.GetTask(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	// If new parent ID is provided, verify it exists
	var newParent *models.Task
	if request.NewParentID != "" {
		newParent, err = h.storage.GetTask(request.NewParentID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "New parent task not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to get new parent task", http.StatusInternalServerError)
			}
			return
		}

		// Check for circular references
		if h.wouldCreateCircularReference(request.NewParentID, taskID) {
			http.Error(w, "Operation would create circular reference", http.StatusBadRequest)
			return
		}
	}

	// Remove task from current parent if it has one
	if task.ParentID != "" {
		currentParent, err := h.storage.GetTask(task.ParentID)
		if err == nil {
			currentParent.RemoveChild(taskID)
			h.storage.UpdateTask(currentParent)
		}
	}

	// Add task to new parent if specified
	if newParent != nil {
		newParent.AddChild(taskID)
		if err := h.storage.UpdateTask(newParent); err != nil {
			http.Error(w, "Failed to update new parent task", http.StatusInternalServerError)
			return
		}
	}

	// Update the task's parent ID
	task.ParentID = request.NewParentID
	task.UpdatedAt = time.Now()

	if err := h.storage.UpdateTask(task); err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	response := struct {
		Message   string       `json:"message"`
		Task      *models.Task `json:"task"`
		NewParent *models.Task `json:"new_parent,omitempty"`
	}{
		Message:   "Task moved successfully",
		Task:      task,
		NewParent: newParent,
	}

	json.NewEncoder(w).Encode(response)
}

// wouldCreateCircularReference checks if making childID a child of parentID would create a circular reference
func (h *Handler) wouldCreateCircularReference(parentID, childID string) bool {
	// If parent and child are the same, it's circular
	if parentID == childID {
		return true
	}

	// Check if parentID is already a descendant of childID
	return h.isDescendant(childID, parentID)
}

// isDescendant checks if ancestorID is a descendant of taskID
func (h *Handler) isDescendant(taskID, ancestorID string) bool {
	task, err := h.storage.GetTask(taskID)
	if err != nil {
		return false
	}

	// Check all children recursively
	for _, childID := range task.Children {
		if childID == ancestorID {
			return true
		}
		if h.isDescendant(childID, ancestorID) {
			return true
		}
	}

	return false
}

// Project API Handlers

// HandleProjects handles requests to /api/projects
func (h *Handler) HandleProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.listProjects(w, r)
	case http.MethodPost:
		h.createProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProject handles requests to /api/projects/{id}
func (h *Handler) HandleProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract project ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	if path == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	// Remove any trailing path elements (e.g., /api/projects/123/something)
	projectID := strings.Split(path, "/")[0]

	switch r.Method {
	case http.MethodGet:
		h.getProject(w, r, projectID)
	case http.MethodPut:
		h.updateProject(w, r, projectID)
	case http.MethodDelete:
		h.deleteProject(w, r, projectID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listProjects returns all projects as JSON
func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.storage.ListProjects()
	if err != nil {
		logger.Error("Failed to list projects", "error", err)
		http.Error(w, "Failed to retrieve projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		logger.Error("Failed to encode projects JSON", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// createProject creates a new project from JSON request
func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var projectData struct {
		Name          string            `json:"name"`
		Description   string            `json:"description"`
		DisplayPrefix string            `json:"display_prefix"`
		Settings      map[string]string `json:"settings,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&projectData); err != nil {
		logger.Error("Failed to decode project JSON", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if strings.TrimSpace(projectData.Name) == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(projectData.DisplayPrefix) == "" {
		http.Error(w, "Project display prefix is required", http.StatusBadRequest)
		return
	}

	// Check if project with same name already exists
	existingProject, err := h.storage.GetProjectByName(projectData.Name)
	if err == nil && existingProject != nil {
		http.Error(w, "Project with this name already exists", http.StatusConflict)
		return
	}

	// Create new project
	project := models.NewProject(projectData.Name, projectData.Description, projectData.DisplayPrefix)

	// Add any provided settings
	if projectData.Settings != nil {
		for key, value := range projectData.Settings {
			project.SetSetting(key, value)
		}
	}

	// Validate the project
	if err := project.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save to storage
	if err := h.storage.CreateProject(project); err != nil {
		logger.Error("Failed to create project", "error", err)
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	logger.Info("Project created", "project_id", project.ID, "name", project.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		logger.Error("Failed to encode project JSON", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getProject returns a specific project by ID
func (h *Handler) getProject(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := h.storage.GetProject(projectID)
	if err != nil {
		logger.Error("Failed to get project", "error", err, "project_id", projectID)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		logger.Error("Failed to encode project JSON", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// updateProject updates an existing project
func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// First check if project exists
	existingProject, err := h.storage.GetProject(projectID)
	if err != nil {
		logger.Error("Failed to get project for update", "error", err, "project_id", projectID)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	var updateData struct {
		Name          *string           `json:"name,omitempty"`
		Description   *string           `json:"description,omitempty"`
		DisplayPrefix *string           `json:"display_prefix,omitempty"`
		Settings      map[string]string `json:"settings,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		logger.Error("Failed to decode project update JSON", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if updateData.Name != nil {
		if strings.TrimSpace(*updateData.Name) == "" {
			http.Error(w, "Project name cannot be empty", http.StatusBadRequest)
			return
		}

		// Check for name conflicts (excluding current project)
		if *updateData.Name != existingProject.Name {
			conflictProject, err := h.storage.GetProjectByName(*updateData.Name)
			if err == nil && conflictProject != nil && conflictProject.ID != projectID {
				http.Error(w, "Project with this name already exists", http.StatusConflict)
				return
			}
		}

		existingProject.Name = *updateData.Name
	}

	if updateData.Description != nil {
		existingProject.Description = *updateData.Description
	}

	if updateData.DisplayPrefix != nil {
		if strings.TrimSpace(*updateData.DisplayPrefix) == "" {
			http.Error(w, "Project display prefix cannot be empty", http.StatusBadRequest)
			return
		}
		existingProject.DisplayPrefix = strings.ToUpper(*updateData.DisplayPrefix)
	}

	// Update settings if provided
	if updateData.Settings != nil {
		for key, value := range updateData.Settings {
			existingProject.SetSetting(key, value)
		}
	}

	// Validate the updated project
	if err := existingProject.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update timestamp and save
	existingProject.UpdateTimestamp()
	if err := h.storage.UpdateProject(existingProject); err != nil {
		logger.Error("Failed to update project", "error", err, "project_id", projectID)
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	logger.Info("Project updated", "project_id", projectID, "name", existingProject.Name)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(existingProject); err != nil {
		logger.Error("Failed to encode project JSON", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// deleteProject deletes a project by ID
func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request, projectID string) {
	// First check if project exists
	project, err := h.storage.GetProject(projectID)
	if err != nil {
		logger.Error("Failed to get project for deletion", "error", err, "project_id", projectID)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// TODO: In future versions, check if project has associated tasks
	// For now, we'll allow deletion regardless

	if err := h.storage.DeleteProject(projectID); err != nil {
		logger.Error("Failed to delete project", "error", err, "project_id", projectID)
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	logger.Info("Project deleted", "project_id", projectID, "name", project.Name)

	w.WriteHeader(http.StatusNoContent)
}
