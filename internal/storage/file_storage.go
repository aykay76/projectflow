package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Create projects subdirectory
	projectsDir := filepath.Join(dataDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create projects directory: %w", err)
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

	// Handle project association and display ID generation
	if task.ProjectID == "" {
		// For backward compatibility, assign to default project
		defaultProject, err := fs.getOrCreateDefaultProjectUnsafe()
		if err != nil {
			return fmt.Errorf("failed to get default project: %w", err)
		}
		task.ProjectID = defaultProject.ID
	}

	// Generate display ID if project exists
	if task.ProjectID != "" {
		displayID, err := fs.getNextDisplayIDUnsafe(task.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to generate display ID: %w", err)
		}
		task.DisplayID = displayID
	}

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

// GetTaskByDisplayID retrieves a task by its display ID (e.g., "PF-1", "PF-2")
func (fs *FileStorage) GetTaskByDisplayID(displayID string) (*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getTaskByDisplayIDUnsafe(displayID)
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

// ListTasks returns tasks for a specific project
func (fs *FileStorage) ListTasks(projectID string) ([]*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.listTasksUnsafe(projectID)
}

// listTasksUnsafe returns tasks for a specific project (must be called with mutex held)
func (fs *FileStorage) listTasksUnsafe(projectID string) ([]*models.Task, error) {
	var tasks []*models.Task

	// If no projectID specified, return empty list (rather than all tasks)
	if projectID == "" {
		return tasks, nil
	}

	// Scan the specific project's tasks directory
	tasksDir := filepath.Join(fs.dataDir, "projects", projectID, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return tasks, nil // Project tasks directory doesn't exist yet, return empty list
		}
		return nil, fmt.Errorf("failed to read tasks directory %s: %w", tasksDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Load the task directly from the file
			task, err := fs.tryLoadTaskFile(tasksDir, entry.Name())
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

	// Get all tasks across all projects
	allTasks, err := fs.listAllTasksUnsafe()
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

// Project CRUD methods

// CreateProject creates a new project and assigns it an ID
func (fs *FileStorage) CreateProject(project *models.Project) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Generate UUID for new project
	project.ID = uuid.New().String()

	return fs.saveProjectUnsafe(project)
}

// GetProject retrieves a project by ID
func (fs *FileStorage) GetProject(id string) (*models.Project, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getProjectUnsafe(id)
}

// UpdateProject updates an existing project
func (fs *FileStorage) UpdateProject(project *models.Project) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.projectExistsUnsafe(project.ID) {
		return fmt.Errorf("project not found: %s", project.ID)
	}

	return fs.saveProjectUnsafe(project)
}

// DeleteProject deletes a project
func (fs *FileStorage) DeleteProject(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.projectExistsUnsafe(id) {
		return fmt.Errorf("project not found: %s", id)
	}

	return fs.deleteProjectUnsafe(id)
}

// ListProjects returns all projects
func (fs *FileStorage) ListProjects() ([]*models.Project, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.listProjectsUnsafe()
}

// GetProjectByName retrieves a project by name
func (fs *FileStorage) GetProjectByName(name string) (*models.Project, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.Name == name {
			return project, nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", name)
}

// ProjectExists checks if a project exists
func (fs *FileStorage) ProjectExists(id string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.projectExistsUnsafe(id)
}

// Close closes the storage (no-op for file storage)
func (fs *FileStorage) Close() error {
	return nil
}

// Internal unsafe methods (must be called with mutex held)
func (fs *FileStorage) getTaskUnsafe(id string) (*models.Task, error) {
	// Try to find by display_id first (format: PROJECT-NUMBER)
	parts := strings.Split(id, "-")
	if len(parts) == 2 {
		// This looks like a display ID (e.g., "PF-1")
		projectID := parts[0]
		return fs.tryLoadTaskFile(filepath.Join(fs.dataDir, "projects", projectID, "tasks"), id+".json")
	}

	// If it's not a display ID, it might be a UUID - search all projects
	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		tasksDir := filepath.Join(fs.dataDir, "projects", project.ID, "tasks")

		// Try both UUID.json and DisplayID.json files
		task, err := fs.tryLoadTaskFile(tasksDir, id+".json")
		if err == nil {
			return task, nil
		}

		// Also check all task files in this project to find one with matching ID
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			continue // Skip this project if we can't read the directory
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				task, err := fs.tryLoadTaskFile(tasksDir, entry.Name())
				if err == nil && task.ID == id {
					return task, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("task not found: %s", id)
}

// tryLoadTaskFile attempts to load a task from a specific file path
func (fs *FileStorage) tryLoadTaskFile(tasksDir, filename string) (*models.Task, error) {
	filePath := filepath.Join(tasksDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (fs *FileStorage) getTaskByDisplayIDUnsafe(displayID string) (*models.Task, error) {
	// Get all projects first
	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Normalize input display ID for case-insensitive comparison
	normalizedDisplayID := strings.ToUpper(displayID)

	// Search for the task in each project's tasks directory
	for _, project := range projects {
		tasksDir := filepath.Join(fs.dataDir, "projects", project.ID, "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Project directory doesn't exist yet, skip it
			}
			return nil, fmt.Errorf("failed to read tasks directory %s: %w", tasksDir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				// Try to load the task file directly
				task, err := fs.tryLoadTaskFile(tasksDir, entry.Name())
				if err == nil && strings.ToUpper(task.DisplayID) == normalizedDisplayID {
					return task, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("task not found with display ID: %s", displayID)
}

func (fs *FileStorage) saveTaskUnsafe(task *models.Task) error {
	// Determine which project directory to save to
	projectID := task.ProjectID
	if projectID == "" {
		// For backward compatibility, assign to default project
		defaultProject, err := fs.getOrCreateDefaultProjectUnsafe()
		if err != nil {
			return fmt.Errorf("failed to get default project: %w", err)
		}
		task.ProjectID = defaultProject.ID
		projectID = defaultProject.ID
	}

	// Ensure the project tasks directory exists
	tasksDir := filepath.Join(fs.dataDir, "projects", projectID, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// Use display_id as filename if available, otherwise fallback to id
	filename := task.ID + ".json"
	if task.DisplayID != "" {
		filename = task.DisplayID + ".json"
	}
	filePath := filepath.Join(tasksDir, filename)

	// For migration support: if we're saving with display_id but old id file exists, remove it
	if task.DisplayID != "" {
		oldFilePath := filepath.Join(tasksDir, task.ID+".json")
		if _, err := os.Stat(oldFilePath); err == nil {
			// Old file exists, remove it after successful save
			defer func() {
				os.Remove(oldFilePath)
			}()
		}
	}

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
	// First, try to get the task to find its display_id and project
	task, err := fs.getTaskUnsafe(id)
	if err != nil {
		// Task not found, this is fine for idempotent delete
		return nil
	}

	// Try to delete by display_id first (new format)
	if task.DisplayID != "" {
		displayFilePath := filepath.Join(fs.dataDir, "projects", task.ProjectID, "tasks", task.DisplayID+".json")
		if err := os.Remove(displayFilePath); err == nil {
			return nil // Successfully deleted by display_id
		}
	}

	// Try to delete by id (old format)
	idFilePath := filepath.Join(fs.dataDir, "projects", task.ProjectID, "tasks", task.ID+".json")
	if err := os.Remove(idFilePath); err == nil {
		return nil // Successfully deleted by id
	}

	// If we still haven't found it, search all projects (fallback)
	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		// Try display_id filename if available
		if task.DisplayID != "" {
			displayFilePath := filepath.Join(fs.dataDir, "projects", project.ID, "tasks", task.DisplayID+".json")
			if err := os.Remove(displayFilePath); err == nil {
				return nil // Successfully deleted
			}
		}

		// Try id filename
		idFilePath := filepath.Join(fs.dataDir, "projects", project.ID, "tasks", task.ID+".json")
		if err := os.Remove(idFilePath); err == nil {
			return nil // Successfully deleted
		}
	}

	// If we get here, task wasn't found in any project
	return nil // Don't error if task doesn't exist (idempotent delete)
}

func (fs *FileStorage) taskExistsUnsafe(id string) bool {
	// Get all projects and check each one
	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return false
	}

	for _, project := range projects {
		tasksDir := filepath.Join(fs.dataDir, "projects", project.ID, "tasks")

		// First check for display_id filename (assume id might be display_id)
		displayFilePath := filepath.Join(tasksDir, id+".json")
		if _, err := os.Stat(displayFilePath); err == nil {
			return true
		}

		// Check all files to see if any task has this id
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			continue // Skip this project if can't read directory
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				if task, err := fs.tryLoadTaskFile(tasksDir, entry.Name()); err == nil {
					if task.ID == id {
						return true
					}
				}
			}
		}
	}

	return false
}

// Internal project methods (must be called with mutex held)

func (fs *FileStorage) getProjectUnsafe(id string) (*models.Project, error) {
	filePath := filepath.Join(fs.dataDir, "projects", id+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read project file: %w", err)
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return &project, nil
}

func (fs *FileStorage) saveProjectUnsafe(project *models.Project) error {
	filePath := filepath.Join(fs.dataDir, "projects", project.ID+".json")
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	return nil
}

func (fs *FileStorage) deleteProjectUnsafe(id string) error {
	filePath := filepath.Join(fs.dataDir, "projects", id+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete project file: %w", err)
	}
	return nil
}

func (fs *FileStorage) listProjectsUnsafe() ([]*models.Project, error) {
	projectsDir := filepath.Join(fs.dataDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projects []*models.Project
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			projectID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
			project, err := fs.getProjectUnsafe(projectID)
			if err == nil {
				projects = append(projects, project)
			}
		}
	}

	return projects, nil
}

func (fs *FileStorage) projectExistsUnsafe(id string) bool {
	filePath := filepath.Join(fs.dataDir, "projects", id+".json")
	_, err := os.Stat(filePath)
	return err == nil
}

// GetNextDisplayID generates and returns the next sequential display ID for a project
func (fs *FileStorage) GetNextDisplayID(projectID string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Get the project to retrieve the display prefix
	project, err := fs.getProjectUnsafe(projectID)
	if err != nil {
		return "", fmt.Errorf("project not found: %w", err)
	}

	// Get the current counter for this project
	counter, err := fs.getProjectCounterUnsafe(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project counter: %w", err)
	}

	// Increment the counter
	counter++

	// Save the updated counter
	if err := fs.saveProjectCounterUnsafe(projectID, counter); err != nil {
		return "", fmt.Errorf("failed to save project counter: %w", err)
	}

	// Format and return the display ID
	return fmt.Sprintf("%s-%d", project.DisplayPrefix, counter), nil
}

// getOrCreateDefaultProjectUnsafe gets or creates a default project for backward compatibility
func (fs *FileStorage) getOrCreateDefaultProjectUnsafe() (*models.Project, error) {
	// Check if default project already exists
	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Look for existing default project
	for _, project := range projects {
		if project.Name == "Default Project" {
			return project, nil
		}
	}

	// Create default project if none exists
	defaultProject := models.NewProject("Default Project", "Default project for tasks without explicit project assignment", "PF")
	defaultProject.ID = uuid.New().String()

	if err := fs.saveProjectUnsafe(defaultProject); err != nil {
		return nil, fmt.Errorf("failed to save default project: %w", err)
	}

	return defaultProject, nil
}

// getNextDisplayIDUnsafe generates the next display ID without locking (internal use)
func (fs *FileStorage) getNextDisplayIDUnsafe(projectID string) (string, error) {
	// Get the project to retrieve the display prefix
	project, err := fs.getProjectUnsafe(projectID)
	if err != nil {
		return "", fmt.Errorf("project not found: %w", err)
	}

	// Get the current counter for this project
	counter, err := fs.getProjectCounterUnsafe(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project counter: %w", err)
	}

	// Increment the counter
	counter++

	// Save the updated counter
	if err := fs.saveProjectCounterUnsafe(projectID, counter); err != nil {
		return "", fmt.Errorf("failed to save project counter: %w", err)
	}

	// Format and return the display ID
	return fmt.Sprintf("%s-%d", project.DisplayPrefix, counter), nil
}

// getProjectCounterUnsafe reads the current counter value for a project
func (fs *FileStorage) getProjectCounterUnsafe(projectID string) (int, error) {
	counterFile := filepath.Join(fs.dataDir, "projects", projectID+".counter")
	data, err := os.ReadFile(counterFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Counter file doesn't exist, start from 0
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read counter file: %w", err)
	}

	var counter int
	if err := json.Unmarshal(data, &counter); err != nil {
		return 0, fmt.Errorf("failed to unmarshal counter: %w", err)
	}

	return counter, nil
}

// saveProjectCounterUnsafe saves the counter value for a project
func (fs *FileStorage) saveProjectCounterUnsafe(projectID string, counter int) error {
	counterFile := filepath.Join(fs.dataDir, "projects", projectID+".counter")
	data, err := json.Marshal(counter)
	if err != nil {
		return fmt.Errorf("failed to marshal counter: %w", err)
	}

	if err := os.WriteFile(counterFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write counter file: %w", err)
	}

	return nil
}

// listAllTasksUnsafe returns all tasks across all projects (must be called with mutex held)
func (fs *FileStorage) listAllTasksUnsafe() ([]*models.Task, error) {
	// Get all projects first
	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var tasks []*models.Task
	seenTasks := make(map[string]bool) // Track task IDs to avoid duplicates

	// Scan each project's tasks directory
	for _, project := range projects {
		tasksDir := filepath.Join(fs.dataDir, "projects", project.ID, "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Project tasks directory doesn't exist yet, skip it
			}
			return nil, fmt.Errorf("failed to read tasks directory %s: %w", tasksDir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				// Load the task directly from the file
				task, err := fs.tryLoadTaskFile(tasksDir, entry.Name())
				if err == nil && !seenTasks[task.ID] {
					tasks = append(tasks, task)
					seenTasks[task.ID] = true
				}
			}
		}
	}

	return tasks, nil
}
