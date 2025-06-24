package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

	// Create tenants subdirectory
	tenantsDir := filepath.Join(dataDir, "tenants")
	if err := os.MkdirAll(tenantsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tenants directory: %w", err)
	}

	fs := &FileStorage{
		dataDir: dataDir,
	}

	// Run migration to convert UUID-based project files to display prefix-based
	if err := fs.MigrateProjectFilesToDisplayPrefix(); err != nil {
		return nil, fmt.Errorf("failed to migrate project files: %w", err)
	}

	return fs, nil
}

// CreateTask creates a new task and assigns it an ID
func (fs *FileStorage) CreateTask(ctx context.Context, task *models.Task) error {
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
func (fs *FileStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getTaskUnsafe(id)
}

// GetTaskByDisplayID retrieves a task by its display ID (e.g., "PF-1", "PF-2")
func (fs *FileStorage) GetTaskByDisplayID(ctx context.Context, displayID string) (*models.Task, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getTaskByDisplayIDUnsafe(displayID)
}

// UpdateTask updates an existing task
func (fs *FileStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.taskExistsUnsafe(task.ID) {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	return fs.saveTaskUnsafe(task)
}

// DeleteTask deletes a task and removes it from parent's children
func (fs *FileStorage) DeleteTask(ctx context.Context, id string) error {
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
func (fs *FileStorage) ListTasks(ctx context.Context, projectID string) ([]*models.Task, error) {
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

	// Get the project's display prefix for the directory structure
	displayPrefix, err := fs.getProjectDisplayPrefixUnsafe(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project display prefix: %w", err)
	}

	// Scan the specific project's tasks directory
	projectDir := filepath.Join(fs.dataDir, "projects", displayPrefix, "tasks")
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return tasks, nil // Project tasks directory doesn't exist yet, return empty list
		}
		return nil, fmt.Errorf("failed to read project tasks directory %s: %w", projectDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Load the task directly from the file
			task, err := fs.tryLoadTaskFile(projectDir, entry.Name())
			if err == nil {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

// GetTaskChildren returns all direct children of a task
func (fs *FileStorage) GetTaskChildren(ctx context.Context, parentID string) ([]*models.Task, error) {
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
func (fs *FileStorage) GetTaskParent(ctx context.Context, childID string) (*models.Task, error) {
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
func (fs *FileStorage) GetTaskHierarchy(ctx context.Context) ([]*models.HierarchyTask, error) {
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
func (fs *FileStorage) TaskExists(ctx context.Context, id string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.taskExistsUnsafe(id)
}

// Project CRUD methods

// CreateProject creates a new project and assigns it an ID
func (fs *FileStorage) CreateProject(ctx context.Context, project *models.Project) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Generate UUID for new project
	project.ID = uuid.New().String()

	return fs.saveProjectUnsafe(project)
}

// GetProject retrieves a project by ID
func (fs *FileStorage) GetProject(ctx context.Context, id string) (*models.Project, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getProjectUnsafe(id)
}

// GetProjectByDisplayPrefix retrieves a project by its display prefix
func (fs *FileStorage) GetProjectByDisplayPrefix(ctx context.Context, displayPrefix string) (*models.Project, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getProjectUnsafe(displayPrefix)
}

// UpdateProject updates an existing project
func (fs *FileStorage) UpdateProject(ctx context.Context, project *models.Project) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.projectExistsUnsafe(project.ID) {
		return fmt.Errorf("project not found: %s", project.ID)
	}

	return fs.saveProjectUnsafe(project)
}

// DeleteProject deletes a project
func (fs *FileStorage) DeleteProject(ctx context.Context, id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if !fs.projectExistsUnsafe(id) {
		return fmt.Errorf("project not found: %s", id)
	}

	return fs.deleteProjectUnsafe(id)
}

// ListProjects returns all projects
func (fs *FileStorage) ListProjects(ctx context.Context) ([]*models.Project, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.listProjectsUnsafe()
}

// GetProjectByName retrieves a project by name
func (fs *FileStorage) GetProjectByName(ctx context.Context, name string) (*models.Project, error) {
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
func (fs *FileStorage) ProjectExists(ctx context.Context, id string) bool {
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
		displayPrefix := parts[0]
		return fs.tryLoadTaskFile(filepath.Join(fs.dataDir, "projects", displayPrefix, "tasks"), id+".json")
	}

	// If it's not a display ID, it might be a UUID - search all projects
	projects, err := fs.listProjectsUnsafe()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		projectDir := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix, "tasks")

		// Check all task files in this project to find one with matching ID
		entries, err := os.ReadDir(projectDir)
		if err != nil {
			continue // Skip this project if we can't read the directory
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				task, err := fs.tryLoadTaskFile(projectDir, entry.Name())
				if err == nil && task.ID == id {
					return task, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("task not found: %s", id)
}

// tryLoadTaskFile attempts to load a task from a specific file path
func (fs *FileStorage) tryLoadTaskFile(projectDir, filename string) (*models.Task, error) {
	filePath := filepath.Join(projectDir, filename)
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

	// Search for the task in each project's directory
	for _, project := range projects {
		projectDir := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix, "tasks")
		entries, err := os.ReadDir(projectDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Project tasks directory doesn't exist yet, skip it
			}
			return nil, fmt.Errorf("failed to read project tasks directory %s: %w", projectDir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				// Try to load the task file directly
				task, err := fs.tryLoadTaskFile(projectDir, entry.Name())
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

	// Get the project's display prefix for the directory structure
	displayPrefix, err := fs.getProjectDisplayPrefixUnsafe(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project display prefix: %w", err)
	}

	// Ensure the project tasks directory exists (using display prefix)
	projectDir := filepath.Join(fs.dataDir, "projects", displayPrefix, "tasks")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project tasks directory: %w", err)
	}

	// Use display_id as filename if available, otherwise fallback to id
	filename := task.ID + ".json"
	if task.DisplayID != "" {
		filename = task.DisplayID + ".json"
	}
	filePath := filepath.Join(projectDir, filename)

	// For migration support: if we're saving with display_id but old id file exists, remove it
	if task.DisplayID != "" {
		oldFilePath := filepath.Join(projectDir, task.ID+".json")
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

	// Get the project's display prefix for the directory structure
	displayPrefix, err := fs.getProjectDisplayPrefixUnsafe(task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project display prefix: %w", err)
	}

	// Try to delete by display_id first (new format)
	if task.DisplayID != "" {
		displayFilePath := filepath.Join(fs.dataDir, "projects", displayPrefix, "tasks", task.DisplayID+".json")
		if err := os.Remove(displayFilePath); err == nil {
			return nil // Successfully deleted by display_id
		}
	}

	// Try to delete by id (old format)
	idFilePath := filepath.Join(fs.dataDir, "projects", displayPrefix, "tasks", task.ID+".json")
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
			displayFilePath := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix, "tasks", task.DisplayID+".json")
			if err := os.Remove(displayFilePath); err == nil {
				return nil // Successfully deleted
			}
		}

		// Try id filename
		idFilePath := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix, "tasks", task.ID+".json")
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
		projectDir := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix, "tasks")

		// First check for display_id filename (assume id might be display_id)
		displayFilePath := filepath.Join(projectDir, id+".json")
		if _, err := os.Stat(displayFilePath); err == nil {
			return true
		}

		// Check all files to see if any task has this id
		entries, err := os.ReadDir(projectDir)
		if err != nil {
			continue // Skip this project if can't read directory
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				if task, err := fs.tryLoadTaskFile(projectDir, entry.Name()); err == nil {
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

// getProjectDisplayPrefixUnsafe retrieves the display prefix for a project by ID
func (fs *FileStorage) getProjectDisplayPrefixUnsafe(projectID string) (string, error) {
	// If the projectID looks like a display prefix (short string without hyphens indicating UUID),
	// we can return it directly. Otherwise, look up the project.
	if len(projectID) < 10 && !strings.Contains(projectID, "-") {
		// This looks like a display prefix, verify it exists
		project, err := fs.getProjectUnsafe(projectID)
		if err != nil {
			return "", err
		}
		return project.DisplayPrefix, nil
	}

	// This might be a UUID, look up the project to get display prefix
	project, err := fs.getProjectUnsafe(projectID)
	if err != nil {
		return "", err
	}
	return project.DisplayPrefix, nil
}

func (fs *FileStorage) getProjectUnsafe(id string) (*models.Project, error) {
	// First try to find by display prefix (new format)
	filePath := filepath.Join(fs.dataDir, "projects", id+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If not found by display prefix, try to find by UUID (legacy format)
			return fs.getProjectByUUIDUnsafe(id)
		}
		return nil, fmt.Errorf("failed to read project file: %w", err)
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project: %w", err)
	}

	return &project, nil
}

// getProjectByUUIDUnsafe looks for a project by UUID (legacy format)
func (fs *FileStorage) getProjectByUUIDUnsafe(uuid string) (*models.Project, error) {
	projectsDir := filepath.Join(fs.dataDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(projectsDir, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var project models.Project
			if err := json.Unmarshal(data, &project); err != nil {
				continue
			}

			if project.ID == uuid {
				return &project, nil
			}
		}
	}

	return nil, fmt.Errorf("project not found: %s", uuid)
}

func (fs *FileStorage) saveProjectUnsafe(project *models.Project) error {
	// Use display prefix for file naming instead of UUID
	filePath := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix+".json")
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
	// First get the project to determine its display prefix
	project, err := fs.getProjectUnsafe(id)
	if err != nil {
		return fmt.Errorf("project not found: %s", id)
	}

	// Delete the project file (stored by display prefix)
	filePath := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete project file: %w", err)
	}

	// Also delete counter file (also stored by display prefix)
	counterPath := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix+".counter")
	if err := os.Remove(counterPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete project counter file: %w", err)
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
	_, err := fs.getProjectUnsafe(id)
	return err == nil
}

// GetNextDisplayID generates and returns the next sequential display ID for a project
func (fs *FileStorage) GetNextDisplayID(ctx context.Context, projectID string) (string, error) {
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
	// First try to find the project to get its display prefix
	project, err := fs.getProjectUnsafe(projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to find project: %w", err)
	}

	counterFile := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix+".counter")
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
	// First try to find the project to get its display prefix
	project, err := fs.getProjectUnsafe(projectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	counterFile := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix+".counter")
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
		projectDir := filepath.Join(fs.dataDir, "projects", project.DisplayPrefix, "tasks")
		entries, err := os.ReadDir(projectDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Project tasks directory doesn't exist yet, skip it
			}
			return nil, fmt.Errorf("failed to read project tasks directory %s: %w", projectDir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				// Load the task directly from the file
				task, err := fs.tryLoadTaskFile(projectDir, entry.Name())
				if err == nil && !seenTasks[task.ID] {
					tasks = append(tasks, task)
					seenTasks[task.ID] = true
				}
			}
		}
	}

	return tasks, nil
}

// MigrateProjectFilesToDisplayPrefix migrates existing UUID-based project files to display prefix naming
func (fs *FileStorage) MigrateProjectFilesToDisplayPrefix() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	projectsDir := filepath.Join(fs.dataDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return fmt.Errorf("failed to read projects directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			fileName := strings.TrimSuffix(entry.Name(), ".json")

			// Check if this is a UUID-named file (36 chars with hyphens)
			if len(fileName) == 36 && strings.Count(fileName, "-") == 4 {
				// This is a UUID-based file, migrate it
				oldPath := filepath.Join(projectsDir, entry.Name())

				// Read the project to get its display prefix
				data, err := os.ReadFile(oldPath)
				if err != nil {
					continue // Skip files we can't read
				}

				var project models.Project
				if err := json.Unmarshal(data, &project); err != nil {
					continue // Skip files we can't parse
				}

				// Create new file with display prefix name
				newPath := filepath.Join(projectsDir, project.DisplayPrefix+".json")

				// Only migrate if the new file doesn't already exist
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					if err := os.Rename(oldPath, newPath); err != nil {
						return fmt.Errorf("failed to rename project file %s to %s: %w", oldPath, newPath, err)
					}

					// Also migrate the counter file if it exists
					oldCounterPath := filepath.Join(projectsDir, fileName+".counter")
					newCounterPath := filepath.Join(projectsDir, project.DisplayPrefix+".counter")

					if _, err := os.Stat(oldCounterPath); err == nil {
						if err := os.Rename(oldCounterPath, newCounterPath); err != nil {
							// Log warning but don't fail the migration
							fmt.Printf("Warning: failed to rename counter file %s to %s: %v\n", oldCounterPath, newCounterPath, err)
						}
					}
				}
			}
		}
	}

	return nil
}

// Tenant CRUD Operations for File Storage

// CreateTenant creates a new tenant in file storage
func (fs *FileStorage) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Validate tenant data
	if err := tenant.Validate(); err != nil {
		return fmt.Errorf("tenant validation failed: %w", err)
	}

	// Generate UUID for new tenant
	tenant.ID = uuid.New().String()

	// Create tenants directory if it doesn't exist
	tenantsDir := filepath.Join(fs.dataDir, "tenants")
	if err := os.MkdirAll(tenantsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tenants directory: %w", err)
	}

	// Check if tenant with same name already exists
	if exists, err := fs.tenantNameExists(tenant.Name); err != nil {
		return fmt.Errorf("failed to check tenant name existence: %w", err)
	} else if exists {
		return fmt.Errorf("tenant with name '%s' already exists", tenant.Name)
	}

	// Save tenant to file
	tenantPath := filepath.Join(tenantsDir, tenant.ID+".json")
	data, err := json.MarshalIndent(tenant, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tenant data: %w", err)
	}

	if err := os.WriteFile(tenantPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tenant file: %w", err)
	}

	return nil
}

// GetTenant retrieves a tenant by its ID from file storage
func (fs *FileStorage) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	tenantsDir := filepath.Join(fs.dataDir, "tenants")
	tenantPath := filepath.Join(tenantsDir, id+".json")

	data, err := os.ReadFile(tenantPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tenant with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to read tenant file: %w", err)
	}

	var tenant models.Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant data: %w", err)
	}

	// Don't return deleted tenants
	if tenant.IsDeleted() {
		return nil, fmt.Errorf("tenant with id %s not found", id)
	}

	return &tenant, nil
}

// UpdateTenant updates an existing tenant in file storage with optimistic locking
func (fs *FileStorage) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Validate tenant data
	if err := tenant.Validate(); err != nil {
		return fmt.Errorf("tenant validation failed: %w", err)
	}

	tenantsDir := filepath.Join(fs.dataDir, "tenants")
	tenantPath := filepath.Join(tenantsDir, tenant.ID+".json")

	// Check if tenant exists and get current data for optimistic locking
	currentData, err := os.ReadFile(tenantPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tenant with id %s not found", tenant.ID)
		}
		return fmt.Errorf("failed to read current tenant file: %w", err)
	}

	var currentTenant models.Tenant
	if err := json.Unmarshal(currentData, &currentTenant); err != nil {
		return fmt.Errorf("failed to unmarshal current tenant data: %w", err)
	}

	// Optimistic locking check - compare timestamps
	if !currentTenant.UpdatedAt.Equal(tenant.UpdatedAt) {
		return fmt.Errorf("tenant has been modified by another process: expected updated_at %v, got %v", tenant.UpdatedAt, currentTenant.UpdatedAt)
	}

	// Check if tenant with same name already exists (excluding current tenant)
	if exists, existingID, err := fs.tenantNameExistsExcluding(tenant.Name, tenant.ID); err != nil {
		return fmt.Errorf("failed to check tenant name existence: %w", err)
	} else if exists {
		return fmt.Errorf("tenant with name '%s' already exists (ID: %s)", tenant.Name, existingID)
	}

	// Update timestamp
	tenant.UpdatedAt = time.Now()

	// Save updated tenant to file
	data, err := json.MarshalIndent(tenant, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tenant data: %w", err)
	}

	if err := os.WriteFile(tenantPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tenant file: %w", err)
	}

	return nil
}

// DeleteTenant removes a tenant by ID (soft delete) in file storage
func (fs *FileStorage) DeleteTenant(ctx context.Context, id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	tenantsDir := filepath.Join(fs.dataDir, "tenants")
	tenantPath := filepath.Join(tenantsDir, id+".json")

	// Check if tenant exists
	data, err := os.ReadFile(tenantPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tenant with id %s not found", id)
		}
		return fmt.Errorf("failed to read tenant file: %w", err)
	}

	var tenant models.Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return fmt.Errorf("failed to unmarshal tenant data: %w", err)
	}

	// Check if already deleted
	if tenant.IsDeleted() {
		return fmt.Errorf("tenant with id %s not found or already deleted", id)
	}

	// Soft delete: update status to 'deleted'
	tenant.Delete() // This sets status to 'deleted' and updates timestamp

	// Save updated tenant to file
	updatedData, err := json.MarshalIndent(tenant, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated tenant data: %w", err)
	}

	if err := os.WriteFile(tenantPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated tenant file: %w", err)
	}

	return nil
}

// ListTenants returns a paginated list of tenants from file storage
func (fs *FileStorage) ListTenants(ctx context.Context, limit, offset int) ([]*models.Tenant, int, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	tenantsDir := filepath.Join(fs.dataDir, "tenants")

	// Read all tenant files
	files, err := os.ReadDir(tenantsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Tenant{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to read tenants directory: %w", err)
	}

	var allTenants []*models.Tenant
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			tenantPath := filepath.Join(tenantsDir, file.Name())
			data, err := os.ReadFile(tenantPath)
			if err != nil {
				continue // Skip files we can't read
			}

			var tenant models.Tenant
			if err := json.Unmarshal(data, &tenant); err != nil {
				continue // Skip files we can't parse
			}

			// Only include non-deleted tenants
			if !tenant.IsDeleted() {
				allTenants = append(allTenants, &tenant)
			}
		}
	}

	totalCount := len(allTenants)

	// Apply pagination
	var result []*models.Tenant
	if offset < totalCount {
		end := offset + limit
		if limit <= 0 || end > totalCount {
			end = totalCount
		}
		result = allTenants[offset:end]
	}

	return result, totalCount, nil
}

// TenantExists checks if a tenant exists by ID in file storage
func (fs *FileStorage) TenantExists(ctx context.Context, id string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	tenantsDir := filepath.Join(fs.dataDir, "tenants")
	tenantPath := filepath.Join(tenantsDir, id+".json")

	data, err := os.ReadFile(tenantPath)
	if err != nil {
		return false
	}

	var tenant models.Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return false
	}

	// Only return true for non-deleted tenants
	return !tenant.IsDeleted()
}

// Helper functions for file storage tenant operations

// tenantNameExists checks if a tenant with the given name already exists
func (fs *FileStorage) tenantNameExists(name string) (bool, error) {
	tenantsDir := filepath.Join(fs.dataDir, "tenants")

	files, err := os.ReadDir(tenantsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read tenants directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			tenantPath := filepath.Join(tenantsDir, file.Name())
			data, err := os.ReadFile(tenantPath)
			if err != nil {
				continue // Skip files we can't read
			}

			var tenant models.Tenant
			if err := json.Unmarshal(data, &tenant); err != nil {
				continue // Skip files we can't parse
			}

			// Check name match for non-deleted tenants
			if !tenant.IsDeleted() && tenant.Name == name {
				return true, nil
			}
		}
	}

	return false, nil
}

// tenantNameExistsExcluding checks if a tenant with the given name exists, excluding a specific tenant ID
func (fs *FileStorage) tenantNameExistsExcluding(name, excludeID string) (bool, string, error) {
	tenantsDir := filepath.Join(fs.dataDir, "tenants")

	files, err := os.ReadDir(tenantsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to read tenants directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			tenantPath := filepath.Join(tenantsDir, file.Name())
			data, err := os.ReadFile(tenantPath)
			if err != nil {
				continue // Skip files we can't read
			}

			var tenant models.Tenant
			if err := json.Unmarshal(data, &tenant); err != nil {
				continue // Skip files we can't parse
			}

			// Check name match for non-deleted tenants, excluding the specified ID
			if !tenant.IsDeleted() && tenant.Name == name && tenant.ID != excludeID {
				return true, tenant.ID, nil
			}
		}
	}

	return false, "", nil
}
