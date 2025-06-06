package storage

import (
	"github.com/aykay76/projectflow/internal/models"
)

// Storage defines the interface for task storage operations
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

	// Utility operations
	TaskExists(id string) bool
	Close() error
}
