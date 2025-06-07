package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

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

// HandleIndex serves the main web interface
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tasks, err := h.storage.ListTasks()
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
	tasks, err := h.storage.ListTasks()
	if err != nil {
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	// Use a temporary struct to handle due_date and started_at as strings
	var taskCreate struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		Type        string `json:"type"`
		ParentID    string `json:"parent_id"`
		DueDate     string `json:"due_date"`
		StartedAt   string `json:"started_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskCreate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create task struct and populate fields
	var task models.Task
	task.Title = taskCreate.Title
	task.Description = taskCreate.Description
	task.ParentID = taskCreate.ParentID

	// Handle due_date
	if taskCreate.DueDate != "" {
		if err := task.SetDueDate(taskCreate.DueDate); err != nil {
			http.Error(w, "Invalid due date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	// Handle started_at
	if taskCreate.StartedAt != "" {
		if err := task.SetStartedAt(taskCreate.StartedAt); err != nil {
			http.Error(w, "Invalid start date format. Use RFC3339", http.StatusBadRequest)
			return
		}
	}

	// Validate required fields
	if task.Title == "" {
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
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request, taskID string) {
	task, err := h.storage.GetTask(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, taskID string) {
	// First get the existing task
	existingTask, err := h.storage.GetTask(taskID)
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

	// Ensure the ID matches the URL and update timestamp
	task.ID = taskID
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
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, taskID string) {
	if err := h.storage.DeleteTask(taskID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		}
		return
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
