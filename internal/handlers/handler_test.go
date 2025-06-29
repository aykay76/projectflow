package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aykay76/projectflow/internal/models"
)

// mockStorage implements storage.Storage interface for testing
type mockStorage struct {
	tasks       map[string]*models.Task
	failNext    bool
	notFoundFor map[string]bool
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		tasks:       make(map[string]*models.Task),
		failNext:    false,
		notFoundFor: make(map[string]bool),
	}
}

func (m *mockStorage) CreateTask(task *models.Task) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("storage error")
	}

	// Generate ID if not set
	if task.ID == "" {
		task.ID = fmt.Sprintf("test-id-%d", len(m.tasks)+1)
	}

	m.tasks[task.ID] = task
	return nil
}

func (m *mockStorage) GetTask(id string) (*models.Task, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	if m.notFoundFor[id] {
		return nil, fmt.Errorf("task not found")
	}

	task, exists := m.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}

	// Return a copy to avoid mutation
	taskCopy := *task
	return &taskCopy, nil
}

func (m *mockStorage) UpdateTask(task *models.Task) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("storage error")
	}

	if m.notFoundFor[task.ID] {
		return fmt.Errorf("task not found")
	}

	if _, exists := m.tasks[task.ID]; !exists {
		return fmt.Errorf("task not found")
	}

	m.tasks[task.ID] = task
	return nil
}

func (m *mockStorage) DeleteTask(id string) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("storage error")
	}

	if m.notFoundFor[id] {
		return fmt.Errorf("task not found")
	}

	if _, exists := m.tasks[id]; !exists {
		return fmt.Errorf("task not found")
	}

	delete(m.tasks, id)
	return nil
}

func (m *mockStorage) ListTasks(projectID string) ([]*models.Task, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		// Filter by project ID if specified, otherwise include all tasks
		if projectID == "" || task.ProjectID == projectID {
			// Return copies to avoid mutation
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}
	return tasks, nil
}

func (m *mockStorage) GetTaskChildren(parentID string) ([]*models.Task, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	if m.notFoundFor[parentID] {
		return nil, fmt.Errorf("parent task not found")
	}

	parent, exists := m.tasks[parentID]
	if !exists {
		return nil, fmt.Errorf("parent task not found")
	}

	children := make([]*models.Task, 0, len(parent.Children))
	for _, childID := range parent.Children {
		if child, exists := m.tasks[childID]; exists {
			childCopy := *child
			children = append(children, &childCopy)
		}
	}

	return children, nil
}

func (m *mockStorage) GetTaskParent(childID string) (*models.Task, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	child, exists := m.tasks[childID]
	if !exists {
		return nil, fmt.Errorf("child task not found")
	}

	if child.ParentID == "" {
		return nil, fmt.Errorf("parent task not found")
	}

	parent, exists := m.tasks[child.ParentID]
	if !exists {
		return nil, fmt.Errorf("parent task not found")
	}

	parentCopy := *parent
	return &parentCopy, nil
}

func (m *mockStorage) GetTaskHierarchy() ([]*models.HierarchyTask, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	// Simple implementation for testing
	hierarchy := make([]*models.HierarchyTask, 0)
	for _, task := range m.tasks {
		if task.ParentID == "" {
			hierarchyTask := &models.HierarchyTask{
				Task:       task,
				ChildTasks: []*models.HierarchyTask{},
			}
			hierarchy = append(hierarchy, hierarchyTask)
		}
	}

	return hierarchy, nil
}

func (m *mockStorage) TaskExists(id string) bool {
	_, exists := m.tasks[id]
	return exists
}

// Project methods for mockStorage
func (m *mockStorage) CreateProject(project *models.Project) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("storage error")
	}

	// For mock, we don't store projects but return success
	return nil
}

func (m *mockStorage) GetProject(id string) (*models.Project, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	if m.notFoundFor[id] {
		return nil, fmt.Errorf("project not found")
	}

	// Return a mock project for testing
	return &models.Project{
		ID:            id,
		Name:          "Test Project",
		Description:   "A test project",
		DisplayPrefix: "TP",
		Settings:      make(map[string]string),
	}, nil
}

func (m *mockStorage) UpdateProject(project *models.Project) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("storage error")
	}

	if m.notFoundFor[project.ID] {
		return fmt.Errorf("project not found")
	}

	return nil
}

func (m *mockStorage) DeleteProject(id string) error {
	if m.failNext {
		m.failNext = false
		return fmt.Errorf("storage error")
	}

	if m.notFoundFor[id] {
		return fmt.Errorf("project not found")
	}

	return nil
}

func (m *mockStorage) ListProjects() ([]*models.Project, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	// Return empty list for tests
	return []*models.Project{}, nil
}

func (m *mockStorage) GetProjectByName(name string) (*models.Project, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	if name == "notfound" || name == "Test Project" {
		return nil, fmt.Errorf("project not found")
	}

	// Return a mock project for testing
	return &models.Project{
		ID:            "test-project-1",
		Name:          name,
		Description:   "A test project",
		DisplayPrefix: "TP",
		Settings:      make(map[string]string),
	}, nil
}

func (m *mockStorage) ProjectExists(id string) bool {
	// For mock purposes, return true unless explicitly set to not exist
	return !m.notFoundFor[id]
}

func (m *mockStorage) Close() error {
	return nil
}

// GetTaskByDisplayID retrieves a task by its display ID
func (m *mockStorage) GetTaskByDisplayID(displayID string) (*models.Task, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	for _, task := range m.tasks {
		if task.DisplayID == displayID {
			taskCopy := *task
			return &taskCopy, nil
		}
	}
	return nil, fmt.Errorf("task not found: %s", displayID)
}

// GetNextDisplayID generates the next display ID for a project
func (m *mockStorage) GetNextDisplayID(projectID string) (string, error) {
	if m.failNext {
		m.failNext = false
		return "", fmt.Errorf("storage error")
	}

	// Find the project to get its prefix
	// For testing, we'll assume PF is the default prefix
	prefix := "PF"

	// Count existing tasks for this project to determine next ID
	count := 0
	for _, task := range m.tasks {
		if task.ProjectID == projectID {
			count++
		}
	}

	return fmt.Sprintf("%s-%d", prefix, count+1), nil
}

// GetProjectByDisplayPrefix retrieves a project by its display prefix
func (m *mockStorage) GetProjectByDisplayPrefix(displayPrefix string) (*models.Project, error) {
	if m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("storage error")
	}

	// For testing, create a mock project
	return &models.Project{
		ID:            "project-1",
		Name:          "Test Project",
		DisplayPrefix: displayPrefix,
	}, nil
}

// Helper function to create a handler with mock storage
func setupHandler() (*Handler, *mockStorage) {
	storage := newMockStorage()

	// Create temp directory for templates (skip template loading for tests)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	handler := &Handler{
		storage:   storage,
		templates: nil, // We'll skip template tests for now
	}

	return handler, storage
}

// Test HandleHierarchy
func TestHandler_HandleHierarchy(t *testing.T) {
	handler, storage := setupHandler()

	t.Run("successful hierarchy retrieval", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/hierarchy", nil)
		w := httptest.NewRecorder()

		handler.HandleHierarchy(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/hierarchy", nil)
		w := httptest.NewRecorder()

		handler.HandleHierarchy(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("storage error", func(t *testing.T) {
		storage.failNext = true
		req := httptest.NewRequest(http.MethodGet, "/api/hierarchy", nil)
		w := httptest.NewRecorder()

		handler.HandleHierarchy(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

// Test HandleTasks
func TestHandler_HandleTasks(t *testing.T) {
	handler, storage := setupHandler()

	t.Run("GET /api/tasks - list tasks", func(t *testing.T) {
		// Add a test task
		task := models.NewTask("Test Task", "Test Description")
		storage.CreateTask(task)

		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tasks []*models.Task
		if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if len(tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(tasks))
		}
	})

	t.Run("GET /api/tasks - storage error", func(t *testing.T) {
		storage.failNext = true
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})

	t.Run("POST /api/tasks - create task", func(t *testing.T) {
		taskData := map[string]interface{}{
			"title":       "New Task",
			"description": "New Description",
			"priority":    "high",
			"status":      "todo",
			"type":        "task",
		}

		body, _ := json.Marshal(taskData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var task models.Task
		if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if task.Title != "New Task" {
			t.Errorf("Expected title 'New Task', got %s", task.Title)
		}
	})

	t.Run("POST /api/tasks - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("POST /api/tasks - missing title", func(t *testing.T) {
		taskData := map[string]interface{}{
			"description": "Description only",
		}

		body, _ := json.Marshal(taskData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("POST /api/tasks - invalid due date", func(t *testing.T) {
		taskData := map[string]interface{}{
			"title":    "Task with bad date",
			"due_date": "invalid-date",
		}

		body, _ := json.Marshal(taskData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("POST /api/tasks - invalid enum values", func(t *testing.T) {
		taskData := map[string]interface{}{
			"title":    "Task with invalid enum",
			"status":   "invalid_status",
			"priority": "invalid_priority",
			"type":     "invalid_type",
		}

		body, _ := json.Marshal(taskData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/tasks", nil)
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// Test HandleTask
func TestHandler_HandleTask(t *testing.T) {
	handler, storage := setupHandler()

	// Create a test task
	task := models.NewTask("Test Task", "Test Description")
	task.ID = "test-task-1"
	storage.CreateTask(task)

	t.Run("GET /api/tasks/{id} - get task", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/test-task-1", nil)
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var responseTask models.Task
		if err := json.NewDecoder(w.Body).Decode(&responseTask); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if responseTask.ID != "test-task-1" {
			t.Errorf("Expected ID 'test-task-1', got %s", responseTask.ID)
		}
	})

	t.Run("GET /api/tasks/{id} - task not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/nonexistent", nil)
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("PUT /api/tasks/{id} - update task", func(t *testing.T) {
		updateData := map[string]interface{}{
			"title":       "Updated Task",
			"description": "Updated Description",
			"status":      "in_progress",
		}

		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/test-task-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var responseTask models.Task
		if err := json.NewDecoder(w.Body).Decode(&responseTask); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if responseTask.Title != "Updated Task" {
			t.Errorf("Expected title 'Updated Task', got %s", responseTask.Title)
		}

		if responseTask.Status != models.StatusInProgress {
			t.Errorf("Expected status 'in_progress', got %s", responseTask.Status)
		}
	})

	t.Run("PUT /api/tasks/{id} - task not found", func(t *testing.T) {
		updateData := map[string]interface{}{
			"title": "Updated Task",
		}

		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/nonexistent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("DELETE /api/tasks/{id} - delete task", func(t *testing.T) {
		// Create another task for deletion
		deleteTask := models.NewTask("Delete Me", "To be deleted")
		deleteTask.ID = "delete-me"
		storage.CreateTask(deleteTask)

		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/delete-me", nil)
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		// Verify task is deleted
		if storage.TaskExists("delete-me") {
			t.Error("Task should have been deleted")
		}
	})

	t.Run("DELETE /api/tasks/{id} - task not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/nonexistent", nil)
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("empty task ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/", nil)
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/tasks/test-task-1", nil)
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// Test HandleTaskChildren
func TestHandler_HandleTaskChildren(t *testing.T) {
	handler, storage := setupHandler()

	// Create parent and child tasks
	parentTask := models.NewTask("Parent Task", "Parent Description")
	parentTask.ID = "parent-task"
	storage.CreateTask(parentTask)

	childTask := models.NewTask("Child Task", "Child Description")
	childTask.ID = "child-task"
	childTask.ParentID = "parent-task"
	parentTask.AddChild("child-task")
	storage.CreateTask(childTask)
	storage.UpdateTask(parentTask)

	t.Run("GET /api/tasks/{id}/children - get children", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/parent-task/children", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildren(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var children []*models.Task
		if err := json.NewDecoder(w.Body).Decode(&children); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if len(children) != 1 {
			t.Errorf("Expected 1 child, got %d", len(children))
		}

		if children[0].ID != "child-task" {
			t.Errorf("Expected child ID 'child-task', got %s", children[0].ID)
		}
	})

	t.Run("GET /api/tasks/{id}/children - parent not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/nonexistent/children", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildren(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("POST /api/tasks/{id}/children - add child", func(t *testing.T) {
		// Create a new child task
		newChild := models.NewTask("New Child", "New Child Description")
		newChild.ID = "new-child"
		storage.CreateTask(newChild)

		requestData := map[string]interface{}{
			"child_id": "new-child",
		}

		body, _ := json.Marshal(requestData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/parent-task/children", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTaskChildren(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("POST /api/tasks/{id}/children - missing child_id", func(t *testing.T) {
		requestData := map[string]interface{}{}

		body, _ := json.Marshal(requestData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/parent-task/children", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTaskChildren(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid URL path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/parent-task/invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildren(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/parent-task/children", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildren(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// Test HandleTaskChildRelation
func TestHandler_HandleTaskChildRelation(t *testing.T) {
	handler, storage := setupHandler()

	// Create parent and child tasks
	parentTask := models.NewTask("Parent Task", "Parent Description")
	parentTask.ID = "parent-task"
	storage.CreateTask(parentTask)

	childTask := models.NewTask("Child Task", "Child Description")
	childTask.ID = "child-task"
	childTask.ParentID = "parent-task"
	parentTask.AddChild("child-task")
	storage.CreateTask(childTask)
	storage.UpdateTask(parentTask)

	t.Run("DELETE /api/tasks/{id}/children/{child_id} - remove child", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/parent-task/children/child-task", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildRelation(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify child relationship was removed
		updatedParent, _ := storage.GetTask("parent-task")
		if len(updatedParent.Children) != 0 {
			t.Error("Child should have been removed from parent")
		}

		updatedChild, _ := storage.GetTask("child-task")
		if updatedChild.ParentID != "" {
			t.Error("Parent ID should have been cleared from child")
		}
	})

	t.Run("DELETE /api/tasks/{id}/children/{child_id} - parent not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/nonexistent/children/child-task", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildRelation(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("DELETE /api/tasks/{id}/children/{child_id} - child not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/parent-task/children/nonexistent", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildRelation(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid URL path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/parent-task/invalid/child-task", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildRelation(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/parent-task/children/child-task", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskChildRelation(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// Test HandleTaskMove
func TestHandler_HandleTaskMove(t *testing.T) {
	handler, storage := setupHandler()

	// Create tasks
	task1 := models.NewTask("Task 1", "Description 1")
	task1.ID = "task-1"
	storage.CreateTask(task1)

	task2 := models.NewTask("Task 2", "Description 2")
	task2.ID = "task-2"
	storage.CreateTask(task2)

	newParent := models.NewTask("New Parent", "New Parent Description")
	newParent.ID = "new-parent"
	storage.CreateTask(newParent)

	t.Run("PUT /api/tasks/{id}/move - move task to new parent", func(t *testing.T) {
		requestData := map[string]interface{}{
			"new_parent_id": "new-parent",
		}

		body, _ := json.Marshal(requestData)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-1/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTaskMove(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify task was moved
		movedTask, _ := storage.GetTask("task-1")
		if movedTask.ParentID != "new-parent" {
			t.Errorf("Expected parent ID 'new-parent', got %s", movedTask.ParentID)
		}

		updatedParent, _ := storage.GetTask("new-parent")
		if len(updatedParent.Children) != 1 || updatedParent.Children[0] != "task-1" {
			t.Error("Task should have been added to new parent's children")
		}
	})

	t.Run("PUT /api/tasks/{id}/move - task not found", func(t *testing.T) {
		requestData := map[string]interface{}{
			"new_parent_id": "new-parent",
		}

		body, _ := json.Marshal(requestData)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/nonexistent/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTaskMove(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("PUT /api/tasks/{id}/move - new parent not found", func(t *testing.T) {
		requestData := map[string]interface{}{
			"new_parent_id": "nonexistent",
		}

		body, _ := json.Marshal(requestData)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-2/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTaskMove(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("PUT /api/tasks/{id}/move - circular reference", func(t *testing.T) {
		// Set up circular reference scenario
		task2.ParentID = "task-1"
		task1.AddChild("task-2")
		storage.UpdateTask(task1)
		storage.UpdateTask(task2)

		requestData := map[string]interface{}{
			"new_parent_id": "task-2",
		}

		body, _ := json.Marshal(requestData)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-1/move", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTaskMove(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid URL path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-1/invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskMove(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/task-1/move", nil)
		w := httptest.NewRecorder()

		handler.HandleTaskMove(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// Test circular reference detection
func TestHandler_CircularReferenceDetection(t *testing.T) {
	handler, storage := setupHandler()

	// Create a chain: A -> B -> C
	taskA := models.NewTask("Task A", "Description A")
	taskA.ID = "task-a"
	storage.CreateTask(taskA)

	taskB := models.NewTask("Task B", "Description B")
	taskB.ID = "task-b"
	taskB.ParentID = "task-a"
	taskA.AddChild("task-b")
	storage.CreateTask(taskB)
	storage.UpdateTask(taskA)

	taskC := models.NewTask("Task C", "Description C")
	taskC.ID = "task-c"
	taskC.ParentID = "task-b"
	taskB.AddChild("task-c")
	storage.CreateTask(taskC)
	storage.UpdateTask(taskB)

	t.Run("wouldCreateCircularReference - direct circular", func(t *testing.T) {
		if !handler.wouldCreateCircularReference("task-a", "task-a") {
			t.Error("Should detect direct circular reference")
		}
	})

	t.Run("wouldCreateCircularReference - indirect circular", func(t *testing.T) {
		if !handler.wouldCreateCircularReference("task-c", "task-a") {
			t.Error("Should detect indirect circular reference")
		}
	})

	t.Run("wouldCreateCircularReference - valid relationship", func(t *testing.T) {
		taskD := models.NewTask("Task D", "Description D")
		taskD.ID = "task-d"
		storage.CreateTask(taskD)

		if handler.wouldCreateCircularReference("task-d", "task-a") {
			t.Error("Should not detect circular reference for valid relationship")
		}
	})
}

// Test edge cases and error scenarios
func TestHandler_EdgeCases(t *testing.T) {
	handler, storage := setupHandler()

	t.Run("POST task with in_progress status auto-sets start date", func(t *testing.T) {
		taskData := map[string]interface{}{
			"title":  "In Progress Task",
			"status": "in_progress",
		}

		body, _ := json.Marshal(taskData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var task models.Task
		if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if task.StartedAt == nil {
			t.Error("StartedAt should be set automatically for in_progress status")
		}
	})

	t.Run("PUT task status to in_progress auto-sets start date", func(t *testing.T) {
		// Create task
		task := models.NewTask("Test Task", "Test Description")
		task.ID = "test-task"
		storage.CreateTask(task)

		updateData := map[string]interface{}{
			"status": "in_progress",
		}

		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/test-task", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTask(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var updatedTask models.Task
		if err := json.NewDecoder(w.Body).Decode(&updatedTask); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if updatedTask.StartedAt == nil {
			t.Error("StartedAt should be set automatically when status changes to in_progress")
		}
	})

	t.Run("storage error during task creation", func(t *testing.T) {
		storage.failNext = true

		taskData := map[string]interface{}{
			"title": "Task that will fail",
		}

		body, _ := json.Marshal(taskData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

// Benchmark tests
func BenchmarkHandler_ListTasks(b *testing.B) {
	handler, storage := setupHandler()

	// Create some test tasks
	for i := 0; i < 100; i++ {
		task := models.NewTask(fmt.Sprintf("Task %d", i), fmt.Sprintf("Description %d", i))
		task.ID = fmt.Sprintf("task-%d", i)
		storage.CreateTask(task)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

func BenchmarkHandler_CreateTask(b *testing.B) {
	handler, _ := setupHandler()

	taskData := map[string]interface{}{
		"title":       "Benchmark Task",
		"description": "Benchmark Description",
		"priority":    "medium",
		"status":      "todo",
		"type":        "task",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		body, _ := json.Marshal(taskData)
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleTasks(w, req)

		if w.Code != http.StatusCreated {
			b.Errorf("Expected status 201, got %d", w.Code)
		}
	}
}

// Project API endpoint tests

func TestHandleProjects_GET(t *testing.T) {
	handler, _ := setupHandler()

	req, err := http.NewRequest("GET", "/api/projects", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.HandleProjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var projects []*models.Project
	if err := json.Unmarshal(w.Body.Bytes(), &projects); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
}

func TestHandleProjects_POST(t *testing.T) {
	handler, _ := setupHandler()

	projectData := map[string]interface{}{
		"name":           "Test Project",
		"description":    "A test project",
		"display_prefix": "TP",
	}

	body, _ := json.Marshal(projectData)
	req, err := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleProjects(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var project models.Project
	if err := json.Unmarshal(w.Body.Bytes(), &project); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if project.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", project.Name)
	}
}

func TestHandleProjects_POST_InvalidJSON(t *testing.T) {
	handler, _ := setupHandler()

	req, err := http.NewRequest("POST", "/api/projects", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleProjects(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleProjects_POST_ValidationError(t *testing.T) {
	handler, _ := setupHandler()

	projectData := map[string]interface{}{
		"name": "", // Invalid: empty name
	}

	body, _ := json.Marshal(projectData)
	req, err := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleProjects(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleProjects_UnsupportedMethod(t *testing.T) {
	handler, _ := setupHandler()

	req, err := http.NewRequest("DELETE", "/api/projects", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.HandleProjects(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleProject_GET(t *testing.T) {
	handler, _ := setupHandler()

	req, err := http.NewRequest("GET", "/api/projects/test-project-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.HandleProject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var project models.Project
	if err := json.Unmarshal(w.Body.Bytes(), &project); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if project.ID != "test-project-1" {
		t.Errorf("Expected ID 'test-project-1', got '%s'", project.ID)
	}
}

func TestHandleProject_GET_NotFound(t *testing.T) {
	handler, storage := setupHandler()
	storage.notFoundFor["nonexistent"] = true

	req, err := http.NewRequest("GET", "/api/projects/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.HandleProject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleProject_PUT(t *testing.T) {
	handler, _ := setupHandler()

	projectData := map[string]interface{}{
		"name":        "Updated Project",
		"description": "Updated description",
	}

	body, _ := json.Marshal(projectData)
	req, err := http.NewRequest("PUT", "/api/projects/test-project-1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleProject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var project models.Project
	if err := json.Unmarshal(w.Body.Bytes(), &project); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if project.Name != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got '%s'", project.Name)
	}
}

func TestHandleProject_PUT_NotFound(t *testing.T) {
	handler, storage := setupHandler()
	storage.notFoundFor["nonexistent"] = true

	projectData := map[string]interface{}{
		"name": "Updated Project",
	}

	body, _ := json.Marshal(projectData)
	req, err := http.NewRequest("PUT", "/api/projects/nonexistent", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleProject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleProject_DELETE(t *testing.T) {
	handler, _ := setupHandler()

	req, err := http.NewRequest("DELETE", "/api/projects/test-project-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.HandleProject(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleProject_DELETE_NotFound(t *testing.T) {
	handler, storage := setupHandler()
	storage.notFoundFor["nonexistent"] = true

	req, err := http.NewRequest("DELETE", "/api/projects/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.HandleProject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleProject_UnsupportedMethod(t *testing.T) {
	handler, _ := setupHandler()

	req, err := http.NewRequest("PATCH", "/api/projects/test-project-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.HandleProject(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
