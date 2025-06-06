package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aykay76/projectflow/internal/models"
	"github.com/google/uuid"
)

// FileStorage implements the Storage interface using the file system
type FileStorage struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileStorage creates a new file-based storage instance
func NewFileStorage(dataDir string) (*FileStorage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create tasks subdirectory
	tasksDir := filepath.Join(dataDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
	}, nil
}

// CreateTask creates a new task and assigns it an ID
func (fs *FileStorage) CreateTask(task *models.Task) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Generate UUID for new task
	task.ID = uuid.New().String()

	// If this task has a parent, add it to parent's children
	if task.ParentID != "" {
		parent, err := fs.getTaskUnsafe(task.ParentID)
		if err != nil {
			return fmt.Errorf("parent task not found: %w", err)
		}
		parent.AddChild(task.ID)
		if err := fs.saveTaskUnsafe(parent); err != nil {
			return fmt.Errorf("failed to update parent task: %w", err)
		}
	}

	return fs.saveTaskUnsafe(task)
}

// GetTask retrieves a task by ID
func (fs *FileStorage) GetTask(id string) (*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getTaskUnsafe(id)
}

// UpdateTask updates an existing task
func (fs *FileStorage) UpdateTask(task *models.Task) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.taskExistsUnsafe(task.ID) {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	return fs.saveTaskUnsafe(task)
}

// DeleteTask deletes a task and removes it from parent's children
func (fs *FileStorage) DeleteTask(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	task, err := fs.getTaskUnsafe(id)
	if err != nil {
		return err
	}

	// Remove from parent's children if it has a parent
	if task.ParentID != "" {
		parent, err := fs.getTaskUnsafe(task.ParentID)
		if err == nil {
			parent.RemoveChild(id)
			fs.saveTaskUnsafe(parent)
		}
	}

	// Delete all children recursively
	for _, childID := range task.Children {
		fs.deleteTaskUnsafe(childID)
	}

	return fs.deleteTaskUnsafe(id)
}

// ListTasks returns all tasks
func (fs *FileStorage) ListTasks() ([]*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.listTasksUnsafe()
}

// listTasksUnsafe returns all tasks (must be called with mutex held)
func (fs *FileStorage) listTasksUnsafe() ([]*models.Task, error) {
	tasksDir := filepath.Join(fs.dataDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	var tasks []*models.Task
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			taskID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
			task, err := fs.getTaskUnsafe(taskID)
			if err == nil {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

// GetTaskChildren returns all direct children of a task
func (fs *FileStorage) GetTaskChildren(parentID string) ([]*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	parent, err := fs.getTaskUnsafe(parentID)
	if err != nil {
		return nil, err
	}

	var children []*models.Task
	for _, childID := range parent.Children {
		child, err := fs.getTaskUnsafe(childID)
		if err == nil {
			children = append(children, child)
		}
	}

	return children, nil
}

// GetTaskParent returns the parent task of a given task
func (fs *FileStorage) GetTaskParent(childID string) (*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	child, err := fs.getTaskUnsafe(childID)
	if err != nil {
		return nil, err
	}

	if child.ParentID == "" {
		return nil, fmt.Errorf("task has no parent")
	}

	return fs.getTaskUnsafe(child.ParentID)
}

// GetTaskHierarchy returns all tasks organized in hierarchical structure
// Returns only top-level tasks (epics without parents) with their nested children
func (fs *FileStorage) GetTaskHierarchy() ([]*models.HierarchyTask, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Get all tasks first
	allTasks, err := fs.listTasksUnsafe()
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup
	taskMap := make(map[string]*models.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Build the hierarchy by finding root tasks (no parent) and recursively building children
	var rootTasks []*models.HierarchyTask
	for _, task := range allTasks {
		if task.ParentID == "" {
			hierarchyTask := fs.buildHierarchyTask(task, taskMap)
			rootTasks = append(rootTasks, hierarchyTask)
		}
	}

	return rootTasks, nil
}

// buildHierarchyTask recursively builds a HierarchyTask with its children
func (fs *FileStorage) buildHierarchyTask(task *models.Task, taskMap map[string]*models.Task) *models.HierarchyTask {
	hierarchyTask := &models.HierarchyTask{
		Task:       task,
		ChildTasks: []*models.HierarchyTask{},
	}

	// Recursively build children
	for _, childID := range task.Children {
		if childTask, exists := taskMap[childID]; exists {
			childHierarchyTask := fs.buildHierarchyTask(childTask, taskMap)
			hierarchyTask.ChildTasks = append(hierarchyTask.ChildTasks, childHierarchyTask)
		}
	}

	return hierarchyTask
}

// TaskExists checks if a task exists
func (fs *FileStorage) TaskExists(id string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.taskExistsUnsafe(id)
}

// Close closes the storage (no-op for file storage)
func (fs *FileStorage) Close() error {
	return nil
}

// Internal unsafe methods (must be called with mutex held)

func (fs *FileStorage) getTaskUnsafe(id string) (*models.Task, error) {
	filePath := filepath.Join(fs.dataDir, "tasks", id+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("task not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

func (fs *FileStorage) saveTaskUnsafe(task *models.Task) error {
	filePath := filepath.Join(fs.dataDir, "tasks", task.ID+".json")
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

func (fs *FileStorage) deleteTaskUnsafe(id string) error {
	filePath := filepath.Join(fs.dataDir, "tasks", id+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete task file: %w", err)
	}
	return nil
}

func (fs *FileStorage) taskExistsUnsafe(id string) bool {
	filePath := filepath.Join(fs.dataDir, "tasks", id+".json")
	_, err := os.Stat(filePath)
	return err == nil
}
