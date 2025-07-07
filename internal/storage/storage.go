package storage

import (
	"context"

	"github.com/aykay76/projectflow/internal/models"
)

// Storage defines the interface for task and project storage operations
// All methods accept a context.Context for tenant-aware operations
type Storage interface {
	// Task operations
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, id string) (*models.Task, error)
	GetTaskByDisplayID(ctx context.Context, displayID string) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, projectID string) ([]*models.Task, error)

	// Hierarchy operations
	GetTaskChildren(ctx context.Context, parentID string) ([]*models.Task, error)
	GetTaskParent(ctx context.Context, childID string) (*models.Task, error)
	GetTaskHierarchy(ctx context.Context) ([]*models.HierarchyTask, error)
	GetTaskHierarchyByProject(ctx context.Context, projectID string) ([]*models.HierarchyTask, error)

	// Project operations

	// CreateProject creates a new project in storage
	// Returns error if project with same name already exists or validation fails
	CreateProject(ctx context.Context, project *models.Project) error

	// GetProject retrieves a project by its ID
	// Returns error if project not found
	GetProject(ctx context.Context, id string) (*models.Project, error)

	// UpdateProject updates an existing project
	// Returns error if project not found or validation fails
	UpdateProject(ctx context.Context, project *models.Project) error

	// DeleteProject removes a project by ID
	// Returns error if project not found or has associated tasks
	DeleteProject(ctx context.Context, id string) error

	// ListProjects returns all projects in the system
	ListProjects(ctx context.Context) ([]*models.Project, error)

	// GetProjectByName retrieves a project by its name
	// Returns error if project not found
	GetProjectByName(ctx context.Context, name string) (*models.Project, error)

	// GetNextDisplayID generates and returns the next sequential display ID for a project
	// Returns formatted display ID (e.g., "PF-1", "PF-2") or error if project not found
	GetNextDisplayID(ctx context.Context, projectID string) (string, error)

	// GetProjectByDisplayPrefix retrieves a project by its display prefix
	// Returns error if project not found
	GetProjectByDisplayPrefix(ctx context.Context, displayPrefix string) (*models.Project, error)

	// Tenant operations
	// CreateTenant creates a new tenant in storage
	// Returns error if tenant with same name already exists or validation fails
	CreateTenant(ctx context.Context, tenant *models.Tenant) error

	// GetTenant retrieves a tenant by its ID
	// Returns error if tenant not found
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)

	// UpdateTenant updates an existing tenant with optimistic locking
	// Returns error if tenant not found, validation fails, or concurrent modification detected
	UpdateTenant(ctx context.Context, tenant *models.Tenant) error

	// DeleteTenant removes a tenant by ID (soft delete)
	// Returns error if tenant not found
	DeleteTenant(ctx context.Context, id string) error

	// ListTenants returns a paginated list of tenants
	// limit: maximum number of tenants to return (0 for no limit)
	// offset: number of tenants to skip for pagination
	// Returns tenants, total count, and error
	ListTenants(ctx context.Context, limit, offset int) ([]*models.Tenant, int, error)

	// TenantExists checks if a tenant exists by ID
	TenantExists(ctx context.Context, id string) bool

	// Utility operations
	TaskExists(ctx context.Context, id string) bool
	ProjectExists(ctx context.Context, id string) bool
	Close() error
}

// Tenant context keys and helper functions
type contextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey contextKey = "tenant_id"
)

// WithTenant adds tenant ID to context for tenant-aware operations
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID extracts tenant ID from context
// Returns empty string if no tenant ID is set
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// HasTenant checks if context contains tenant information
func HasTenant(ctx context.Context) bool {
	return GetTenantID(ctx) != ""
}
