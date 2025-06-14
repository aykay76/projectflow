package storage

import (
	"github.com/aykay76/projectflow/internal/models"
)

// Storage defines the interface for task and project storage operations
type Storage interface {
	// Task operations
	CreateTask(task *models.Task) error
	GetTask(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	ListTasks() ([]*models.Task, error)

	// Hierarchy operations
	GetTaskChildren(parentID string) ([]*models.Task, error)
	GetTaskParent(childID string) (*models.Task, error)
	GetTaskHierarchy() ([]*models.HierarchyTask, error)

	// Project operations

	// CreateProject creates a new project in storage
	// Returns error if project with same name already exists or validation fails
	CreateProject(project *models.Project) error

	// GetProject retrieves a project by its ID
	// Returns error if project not found
	GetProject(id string) (*models.Project, error)

	// UpdateProject updates an existing project
	// Returns error if project not found or validation fails
	UpdateProject(project *models.Project) error

	// DeleteProject removes a project by ID
	// Returns error if project not found or has associated tasks
	DeleteProject(id string) error

	// ListProjects returns all projects in the system
	ListProjects() ([]*models.Project, error)

	// GetProjectByName retrieves a project by its name
	// Returns error if project not found
	GetProjectByName(name string) (*models.Project, error)

	// Utility operations
	TaskExists(id string) bool
	ProjectExists(id string) bool
	Close() error
}
