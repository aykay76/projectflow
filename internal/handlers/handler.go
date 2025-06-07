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

	// Create task struct and populate fields
	var task models.Task
	task.Title = taskUpdate.Title
	task.Description = taskUpdate.Description
	task.Status = models.TaskStatus(taskUpdate.Status)
	task.Priority = models.TaskPriority(taskUpdate.Priority)
	task.Type = models.TaskType(taskUpdate.Type)
	task.ParentID = taskUpdate.ParentID
	task.Children = []string{} // Will be preserved by storage layer

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
	if task.Status == models.StatusInProgress && taskUpdate.StartedAt == "" {
		// Need to check current task to see if it already has a start date
		currentTask, err := h.storage.GetTask(taskID)
		if err == nil && currentTask.StartedAt == nil {
			now := time.Now()
			task.StartedAt = &now
		}
	}

	// Ensure the ID matches the URL
	task.ID = taskID
	task.UpdatedAt = time.Now()

	// Validate enum values if provided
	if task.Status != "" && !models.IsValidStatus(string(task.Status)) {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}
	if task.Priority != "" && !models.IsValidPriority(string(task.Priority)) {
		http.Error(w, "Invalid priority", http.StatusBadRequest)
		return
	}
	if task.Type != "" && !models.IsValidType(string(task.Type)) {
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
