package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aykay76/projectflow/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresStorage implements the Storage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewPostgresStorage creates a new PostgreSQL-based storage instance
func NewPostgresStorage(connectionString string) (*PostgresStorage, error) {
	// Open database connection
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{
		db: db,
	}

	// Initialize database schema
	if err := storage.initializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return storage, nil
}

// initializeSchema creates the necessary database tables and indexes
func (ps *PostgresStorage) initializeSchema() error {
	// Step 1: Create tenants table first (required for foreign key relationships)
	createTenantsTableSQL := `
	CREATE TABLE IF NOT EXISTS tenants (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		settings JSONB DEFAULT '{}'::jsonb,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT check_tenant_status CHECK (status IN ('active', 'inactive', 'suspended'))
	);`

	if _, err := ps.db.Exec(createTenantsTableSQL); err != nil {
		return fmt.Errorf("failed to create tenants table: %w", err)
	}

	// Step 2: Create base tables without tenant foreign keys (for backward compatibility)
	createTasksTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(36) PRIMARY KEY,
		display_id VARCHAR(50),
		project_id VARCHAR(36),
		title VARCHAR(255) NOT NULL,
		description TEXT,
		status VARCHAR(20) NOT NULL DEFAULT 'todo',
		priority VARCHAR(20) NOT NULL DEFAULT 'medium',
		type VARCHAR(20) NOT NULL DEFAULT 'task',
		parent_id VARCHAR(36),
		children JSONB DEFAULT '[]'::jsonb,
		started_at TIMESTAMPTZ,
		due_date TIMESTAMPTZ,
		completed_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(display_id)
	);`

	if _, err := ps.db.Exec(createTasksTableSQL); err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Create projects table without tenant foreign key initially
	createProjectsTableSQL := `
	CREATE TABLE IF NOT EXISTS projects (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		description TEXT,
		display_prefix VARCHAR(10) NOT NULL,
		task_counter INTEGER NOT NULL DEFAULT 0,
		settings JSONB DEFAULT '{}'::jsonb,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	if _, err := ps.db.Exec(createProjectsTableSQL); err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}

	// Create users table (new table, can include all constraints from start)
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		tenant_id VARCHAR(36) NOT NULL,
		username VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		is_active BOOLEAN NOT NULL DEFAULT true,
		last_login TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, username),
		UNIQUE(tenant_id, email),
		CONSTRAINT check_user_role CHECK (role IN ('admin', 'user', 'viewer'))
	);`

	if _, err := ps.db.Exec(createUsersTableSQL); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Step 3: Add missing columns to existing tables (for backward compatibility)

	// Add task_counter column to existing projects table if it doesn't exist
	alterTableSQL := `
	DO $$ 
	BEGIN 
		IF NOT EXISTS (
			SELECT column_name 
			FROM information_schema.columns 
			WHERE table_name='projects' AND column_name='task_counter'
		) THEN
			ALTER TABLE projects ADD COLUMN task_counter INTEGER NOT NULL DEFAULT 0;
		END IF;
	END $$;`

	if _, err := ps.db.Exec(alterTableSQL); err != nil {
		return fmt.Errorf("failed to add task_counter column: %w", err)
	}

	// Add tenant_id column to existing projects table if it doesn't exist
	alterProjectsSQL := `
	DO $$ 
	BEGIN 
		IF NOT EXISTS (
			SELECT column_name 
			FROM information_schema.columns 
			WHERE table_name='projects' AND column_name='tenant_id'
		) THEN
			ALTER TABLE projects ADD COLUMN tenant_id VARCHAR(36);
		END IF;
	END $$;`

	if _, err := ps.db.Exec(alterProjectsSQL); err != nil {
		return fmt.Errorf("failed to add tenant_id column to projects: %w", err)
	}

	// Add display_id, project_id, and tenant_id columns to existing tasks table if they don't exist
	alterTasksSQL := `
	DO $$ 
	BEGIN 
		-- Add display_id column
		IF NOT EXISTS (
			SELECT column_name 
			FROM information_schema.columns 
			WHERE table_name='tasks' AND column_name='display_id'
		) THEN
			ALTER TABLE tasks ADD COLUMN display_id VARCHAR(50) UNIQUE;
		END IF;
		
		-- Add project_id column
		IF NOT EXISTS (
			SELECT column_name 
			FROM information_schema.columns 
			WHERE table_name='tasks' AND column_name='project_id'
		) THEN
			ALTER TABLE tasks ADD COLUMN project_id VARCHAR(36);
		END IF;
		
		-- Add tenant_id column
		IF NOT EXISTS (
			SELECT column_name 
			FROM information_schema.columns 
			WHERE table_name='tasks' AND column_name='tenant_id'
		) THEN
			ALTER TABLE tasks ADD COLUMN tenant_id VARCHAR(36);
		END IF;
	END $$;`

	if _, err := ps.db.Exec(alterTasksSQL); err != nil {
		return fmt.Errorf("failed to add display_id, project_id, and tenant_id columns: %w", err)
	}

	// Step 4: Add foreign key constraints (after all columns exist)

	// Add foreign key constraints for projects table
	addProjectsConstraintsSQL := `
	DO $$
	BEGIN
		-- Add foreign key constraint for tenant_id if it doesn't exist
		IF NOT EXISTS (
			SELECT constraint_name 
			FROM information_schema.table_constraints 
			WHERE table_name='projects' AND constraint_name='fk_projects_tenant_id'
		) THEN
			ALTER TABLE projects 
			ADD CONSTRAINT fk_projects_tenant_id 
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
		END IF;
		
		-- Update unique constraint to include tenant_id if needed
		-- Note: This requires dropping and recreating the constraint
		IF EXISTS (
			SELECT constraint_name 
			FROM information_schema.table_constraints 
			WHERE table_name='projects' AND constraint_name='projects_name_key'
		) THEN
			ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_name_key;
			ALTER TABLE projects ADD CONSTRAINT projects_tenant_name_unique UNIQUE(tenant_id, name);
		END IF;
	END $$;`

	if _, err := ps.db.Exec(addProjectsConstraintsSQL); err != nil {
		return fmt.Errorf("failed to add projects constraints: %w", err)
	}

	// Add foreign key constraints for tasks table
	addTasksConstraintsSQL := `
	DO $$
	BEGIN
		-- Add parent_id foreign key if it doesn't exist
		IF NOT EXISTS (
			SELECT constraint_name 
			FROM information_schema.table_constraints 
			WHERE table_name='tasks' AND constraint_name='tasks_parent_id_fkey'
		) THEN
			ALTER TABLE tasks 
			ADD CONSTRAINT tasks_parent_id_fkey 
			FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE SET NULL;
		END IF;
		
		-- Add project_id foreign key if it doesn't exist
		IF NOT EXISTS (
			SELECT constraint_name 
			FROM information_schema.table_constraints 
			WHERE table_name='tasks' AND constraint_name='tasks_project_id_fkey'
		) THEN
			ALTER TABLE tasks 
			ADD CONSTRAINT tasks_project_id_fkey 
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
		END IF;
		
		-- Add tenant_id foreign key if it doesn't exist
		IF NOT EXISTS (
			SELECT constraint_name 
			FROM information_schema.table_constraints 
			WHERE table_name='tasks' AND constraint_name='fk_tasks_tenant_id'
		) THEN
			ALTER TABLE tasks 
			ADD CONSTRAINT fk_tasks_tenant_id 
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
		END IF;
	END $$;`

	if _, err := ps.db.Exec(addTasksConstraintsSQL); err != nil {
		return fmt.Errorf("failed to add tasks constraints: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		// Tenants table indexes
		"CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);",
		"CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);",
		// Tasks table indexes
		"CREATE INDEX IF NOT EXISTS idx_tasks_tenant_id ON tasks(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_tenant_status ON tasks(tenant_id, status);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_tenant_priority ON tasks(tenant_id, priority);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_type ON tasks(type);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);",
		// Projects table indexes
		"CREATE INDEX IF NOT EXISTS idx_projects_tenant_id ON projects(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_projects_tenant_name ON projects(tenant_id, name);",
		"CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);",
		"CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);",
		// Users table indexes
		"CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);",
		"CREATE INDEX IF NOT EXISTS idx_users_tenant_username ON users(tenant_id, username);",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);",
		"CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);",
		"CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login);",
	}

	for _, indexSQL := range indexes {
		if _, err := ps.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Step 5: Initialize Row-Level Security (RLS)
	if err := ps.initializeRLS(); err != nil {
		return fmt.Errorf("failed to initialize RLS: %w", err)
	}

	// Step 6: Migrate existing data to default tenant
	if err := ps.migrateExistingDataToDefaultTenant(); err != nil {
		return fmt.Errorf("failed to migrate existing data: %w", err)
	}

	return nil
}

// CreateTask creates a new task and assigns it an ID
func (ps *PostgresStorage) CreateTask(task *models.Task) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Ensure tenant context is set for this operation
	if err := ps.ensureTenantContext(); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Generate UUID for new task
	task.ID = uuid.New().String()

	// Begin transaction for task creation and display ID generation
	tx, err := ps.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for the transaction
	defaultTenantID, err := ps.getOrCreateDefaultTenant()
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %w", err)
	}
	if err := ps.setTenantContext(tx, defaultTenantID); err != nil {
		return fmt.Errorf("failed to set tenant context for transaction: %w", err)
	}

	// Handle project association and display ID generation
	if task.ProjectID == "" {
		// For backward compatibility, assign to default project
		defaultProject, err := ps.getOrCreateDefaultProjectTx(tx)
		if err != nil {
			return fmt.Errorf("failed to get default project: %w", err)
		}
		task.ProjectID = defaultProject.ID
	}

	// Generate display ID if project exists
	if task.ProjectID != "" {
		displayID, err := ps.getNextDisplayIDTx(tx, task.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to generate display ID: %w", err)
		}
		task.DisplayID = displayID
	}

	// Serialize children array to JSON
	childrenJSON, err := json.Marshal(task.Children)
	err = nil
	// Insert the task with new fields including tenant_id
	insertSQL := `
		INSERT INTO tasks (id, display_id, project_id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at, tenant_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err = tx.Exec(insertSQL,
		task.ID,
		nullString(task.DisplayID),
		nullString(task.ProjectID),
		task.Title,
		task.Description,
		string(task.Status),
		string(task.Priority),
		string(task.Type),
		nullString(task.ParentID),
		childrenJSON,
		task.StartedAt,
		task.DueDate,
		task.CompletedAt,
		task.CreatedAt,
		task.UpdatedAt,
		defaultTenantID, // Add tenant_id to the insert
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	// If this task has a parent, add it to parent's children
	if task.ParentID != "" {
		if err := ps.addChildToParentTx(tx, task.ParentID, task.ID); err != nil {
			return fmt.Errorf("failed to update parent task: %w", err)
		}
	}

	return tx.Commit()
}

// GetTask retrieves a task by ID
func (ps *PostgresStorage) GetTask(id string) (*models.Task, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Ensure tenant context is set for this operation
	if err := ps.ensureTenantContext(); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	querySQL := `
		SELECT id, display_id, project_id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE id = $1`

	row := ps.db.QueryRow(querySQL, id)

	task, err := ps.scanTaskWithDisplayID(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// GetTaskByDisplayID retrieves a task by its display ID (e.g., "PF-1", "PF-2")
func (ps *PostgresStorage) GetTaskByDisplayID(displayID string) (*models.Task, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Ensure tenant context is set for this operation
	if err := ps.ensureTenantContext(); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	querySQL := `
		SELECT id, display_id, project_id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE UPPER(display_id) = UPPER($1)`

	row := ps.db.QueryRow(querySQL, displayID)

	task, err := ps.scanTaskWithDisplayID(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found with display ID: %s", displayID)
		}
		return nil, fmt.Errorf("failed to get task by display ID: %w", err)
	}

	return task, nil
}

// UpdateTask updates an existing task
func (ps *PostgresStorage) UpdateTask(task *models.Task) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Check if task exists
	if !ps.taskExistsUnsafe(task.ID) {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Serialize children array to JSON
	childrenJSON, err := json.Marshal(task.Children)
	if err != nil {
		return fmt.Errorf("failed to marshal children array: %w", err)
	}

	updateSQL := `
		UPDATE tasks SET 
			title = $2,
			description = $3,
			status = $4,
			priority = $5,
			type = $6,
			parent_id = $7,
			children = $8,
			started_at = $9,
			due_date = $10,
			completed_at = $11,
			updated_at = $12
		WHERE id = $1`

	_, err = ps.db.Exec(updateSQL,
		task.ID,
		task.Title,
		task.Description,
		string(task.Status),
		string(task.Priority),
		string(task.Type),
		nullString(task.ParentID),
		childrenJSON,
		task.StartedAt,
		task.DueDate,
		task.CompletedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// DeleteTask deletes a task and removes it from parent's children
func (ps *PostgresStorage) DeleteTask(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Begin transaction
	tx, err := ps.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the task first to find its parent and children
	task, err := ps.getTaskUnsafeTx(tx, id)
	if err != nil {
		return err
	}

	// Remove from parent's children if it has a parent
	if task.ParentID != "" {
		if err := ps.removeChildFromParentTx(tx, task.ParentID, id); err != nil {
			return fmt.Errorf("failed to update parent task: %w", err)
		}
	}

	// Delete all children recursively
	for _, childID := range task.Children {
		if err := ps.deleteTaskUnsafeTx(tx, childID); err != nil {
			return fmt.Errorf("failed to delete child task %s: %w", childID, err)
		}
	}

	// Delete the task itself
	if err := ps.deleteTaskUnsafeTx(tx, id); err != nil {
		return err
	}

	return tx.Commit()
}

// ListTasks returns tasks for a specific project
func (ps *PostgresStorage) ListTasks(projectID string) ([]*models.Task, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Ensure tenant context is set for this operation
	if err := ps.ensureTenantContext(); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// If no projectID specified, return empty list
	if projectID == "" {
		return []*models.Task{}, nil
	}

	querySQL := `
		SELECT id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE project_id = $1 ORDER BY created_at DESC`

	rows, err := ps.db.Query(querySQL, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := ps.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return tasks, nil
}

// GetTaskChildren returns all direct children of a task
func (ps *PostgresStorage) GetTaskChildren(parentID string) ([]*models.Task, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	querySQL := `
		SELECT id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE parent_id = $1 ORDER BY created_at`

	rows, err := ps.db.Query(querySQL, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query child tasks: %w", err)
	}
	defer rows.Close()

	var children []*models.Task
	for rows.Next() {
		task, err := ps.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child task: %w", err)
		}
		children = append(children, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over child rows: %w", err)
	}

	return children, nil
}

// GetTaskParent returns the parent task of a given task
func (ps *PostgresStorage) GetTaskParent(childID string) (*models.Task, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// First get the child task to find its parent ID
	child, err := ps.getTaskUnsafe(childID)
	if err != nil {
		return nil, err
	}

	if child.ParentID == "" {
		return nil, fmt.Errorf("task has no parent: %s", childID)
	}

	return ps.getTaskUnsafe(child.ParentID)
}

// GetTaskHierarchy returns all tasks organized in hierarchical structure
func (ps *PostgresStorage) GetTaskHierarchy() ([]*models.HierarchyTask, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Get all tasks
	allTasks, err := ps.listTasksUnsafe()
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
			hierarchyTask := ps.buildHierarchyTask(task, taskMap)
			rootTasks = append(rootTasks, hierarchyTask)
		}
	}

	return rootTasks, nil
}

// TaskExists checks if a task exists
func (ps *PostgresStorage) TaskExists(id string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.taskExistsUnsafe(id)
}

// Project CRUD methods

// CreateProject creates a new project and assigns it an ID
func (ps *PostgresStorage) CreateProject(project *models.Project) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Ensure tenant context is set for this operation
	if err := ps.ensureTenantContext(); err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Get default tenant ID for the project
	defaultTenantID, err := ps.getOrCreateDefaultTenant()
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %w", err)
	}

	// Generate UUID for new project
	project.ID = uuid.New().String()

	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(project.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Insert the project with tenant_id
	insertSQL := `
		INSERT INTO projects (id, name, description, display_prefix, task_counter, settings, created_at, updated_at, tenant_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = ps.db.Exec(insertSQL,
		project.ID,
		project.Name,
		project.Description,
		project.DisplayPrefix,
		0, // Initialize task_counter to 0
		settingsJSON,
		project.CreatedAt,
		project.UpdatedAt,
		defaultTenantID, // Add tenant_id
	)

	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}

	return nil
}

// GetProject retrieves a project by ID
func (ps *PostgresStorage) GetProject(id string) (*models.Project, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Ensure tenant context is set for this operation
	if err := ps.ensureTenantContext(); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	selectSQL := `
		SELECT id, name, description, display_prefix, settings, created_at, updated_at
		FROM projects WHERE id = $1`

	row := ps.db.QueryRow(selectSQL, id)

	var project models.Project
	var settingsJSON []byte

	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.DisplayPrefix,
		&settingsJSON,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	// Unmarshal settings JSON
	if err := json.Unmarshal(settingsJSON, &project.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &project, nil
}

// UpdateProject updates an existing project
func (ps *PostgresStorage) UpdateProject(project *models.Project) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(project.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	updateSQL := `
		UPDATE projects 
		SET name = $2, description = $3, display_prefix = $4, settings = $5, updated_at = $6
		WHERE id = $1`

	result, err := ps.db.Exec(updateSQL,
		project.ID,
		project.Name,
		project.Description,
		project.DisplayPrefix,
		settingsJSON,
		project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found: %s", project.ID)
	}

	return nil
}

// DeleteProject deletes a project
func (ps *PostgresStorage) DeleteProject(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	deleteSQL := "DELETE FROM projects WHERE id = $1"
	result, err := ps.db.Exec(deleteSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found: %s", id)
	}

	return nil
}

// ListProjects returns all projects
func (ps *PostgresStorage) ListProjects() ([]*models.Project, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Ensure tenant context is set for this operation
	if err := ps.ensureTenantContext(); err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	selectSQL := `
		SELECT id, name, description, display_prefix, settings, created_at, updated_at
		FROM projects ORDER BY created_at ASC`

	rows, err := ps.db.Query(selectSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		var project models.Project
		var settingsJSON []byte

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.DisplayPrefix,
			&settingsJSON,
			&project.CreatedAt,
			&project.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}

		// Unmarshal settings JSON
		if err := json.Unmarshal(settingsJSON, &project.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		projects = append(projects, &project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate project rows: %w", err)
	}

	return projects, nil
}

// GetProjectByName retrieves a project by name
func (ps *PostgresStorage) GetProjectByName(name string) (*models.Project, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	selectSQL := `
		SELECT id, name, description, display_prefix, settings, created_at, updated_at
		FROM projects WHERE name = $1`

	row := ps.db.QueryRow(selectSQL, name)

	var project models.Project
	var settingsJSON []byte

	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.DisplayPrefix,
		&settingsJSON,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", name)
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	// Unmarshal settings JSON
	if err := json.Unmarshal(settingsJSON, &project.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &project, nil
}

// GetProjectByDisplayPrefix retrieves a project by display prefix
func (ps *PostgresStorage) GetProjectByDisplayPrefix(displayPrefix string) (*models.Project, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	selectSQL := `
		SELECT id, name, description, display_prefix, settings, created_at, updated_at
		FROM projects WHERE display_prefix = $1`

	row := ps.db.QueryRow(selectSQL, displayPrefix)

	var project models.Project
	var settingsJSON []byte

	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.DisplayPrefix,
		&settingsJSON,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", displayPrefix)
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	// Unmarshal settings JSON
	if err := json.Unmarshal(settingsJSON, &project.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &project, nil
}

// ProjectExists checks if a project exists
func (ps *PostgresStorage) ProjectExists(id string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var count int
	selectSQL := "SELECT COUNT(*) FROM projects WHERE id = $1"
	err := ps.db.QueryRow(selectSQL, id).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// Close closes the database connection
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}

// Helper methods

// scanTask scans a database row into a Task struct
func (ps *PostgresStorage) scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Task, error) {
	var task models.Task
	var parentID sql.NullString
	var childrenJSON []byte
	var startedAt, dueDate, completedAt sql.NullTime

	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.Type,
		&parentID,
		&childrenJSON,
		&startedAt,
		&dueDate,
		&completedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if parentID.Valid {
		task.ParentID = parentID.String
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	// Unmarshal children JSON
	if err := json.Unmarshal(childrenJSON, &task.Children); err != nil {
		return nil, fmt.Errorf("failed to unmarshal children: %w", err)
	}

	return &task, nil
}

// scanTaskWithDisplayID scans a database row into a Task struct (with display ID)
func (ps *PostgresStorage) scanTaskWithDisplayID(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Task, error) {
	var task models.Task
	var parentID sql.NullString
	var childrenJSON []byte
	var startedAt, dueDate, completedAt sql.NullTime

	err := scanner.Scan(
		&task.ID,
		&task.DisplayID,
		&task.ProjectID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.Type,
		&parentID,
		&childrenJSON,
		&startedAt,
		&dueDate,
		&completedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if parentID.Valid {
		task.ParentID = parentID.String
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	// Unmarshal children JSON
	if err := json.Unmarshal(childrenJSON, &task.Children); err != nil {
		return nil, fmt.Errorf("failed to unmarshal children: %w", err)
	}

	return &task, nil
}

// buildHierarchyTask recursively builds a HierarchyTask with its children
func (ps *PostgresStorage) buildHierarchyTask(task *models.Task, taskMap map[string]*models.Task) *models.HierarchyTask {
	hierarchyTask := &models.HierarchyTask{
		Task:       task,
		ChildTasks: []*models.HierarchyTask{},
	}

	// Recursively build children
	for _, childID := range task.Children {
		if childTask, exists := taskMap[childID]; exists {
			childHierarchyTask := ps.buildHierarchyTask(childTask, taskMap)
			hierarchyTask.ChildTasks = append(hierarchyTask.ChildTasks, childHierarchyTask)
		}
	}

	return hierarchyTask
}

// Unsafe methods (must be called with mutex held)

func (ps *PostgresStorage) getTaskUnsafe(id string) (*models.Task, error) {
	querySQL := `
		SELECT id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE id = $1`

	row := ps.db.QueryRow(querySQL, id)

	task, err := ps.scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

func (ps *PostgresStorage) getTaskUnsafeTx(tx *sql.Tx, id string) (*models.Task, error) {
	querySQL := `
		SELECT id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at
		FROM tasks WHERE id = $1`

	row := tx.QueryRow(querySQL, id)

	task, err := ps.scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

func (ps *PostgresStorage) listTasksUnsafe() ([]*models.Task, error) {
	querySQL := `
		SELECT id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at
		FROM tasks ORDER BY created_at DESC`

	rows, err := ps.db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := ps.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return tasks, nil
}

func (ps *PostgresStorage) taskExistsUnsafe(id string) bool {
	var exists bool
	querySQL := "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)"
	err := ps.db.QueryRow(querySQL, id).Scan(&exists)
	return err == nil && exists
}

func (ps *PostgresStorage) deleteTaskUnsafeTx(tx *sql.Tx, id string) error {
	deleteSQL := "DELETE FROM tasks WHERE id = $1"
	result, err := tx.Exec(deleteSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", id)
	}

	return nil
}

func (ps *PostgresStorage) addChildToParentTx(tx *sql.Tx, parentID, childID string) error {
	// Get current children array
	var childrenJSON []byte
	querySQL := "SELECT children FROM tasks WHERE id = $1"
	err := tx.QueryRow(querySQL, parentID).Scan(&childrenJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parent task not found: %s", parentID)
		}
		return fmt.Errorf("failed to get parent task: %w", err)
	}

	// Unmarshal current children
	var children []string
	if err := json.Unmarshal(childrenJSON, &children); err != nil {
		return fmt.Errorf("failed to unmarshal children: %w", err)
	}

	// Check if child already exists
	for _, child := range children {
		if child == childID {
			return nil // Already exists
		}
	}

	// Add the new child
	children = append(children, childID)

	// Marshal back to JSON
	newChildrenJSON, err := json.Marshal(children)
	if err != nil {
		return fmt.Errorf("failed to marshal children: %w", err)
	}

	// Update the parent
	updateSQL := "UPDATE tasks SET children = $1, updated_at = NOW() WHERE id = $2"
	_, err = tx.Exec(updateSQL, newChildrenJSON, parentID)
	if err != nil {
		return fmt.Errorf("failed to update parent task: %w", err)
	}

	return nil
}

func (ps *PostgresStorage) removeChildFromParentTx(tx *sql.Tx, parentID, childID string) error {
	// Get current children array
	var childrenJSON []byte
	querySQL := "SELECT children FROM tasks WHERE id = $1"
	err := tx.QueryRow(querySQL, parentID).Scan(&childrenJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Parent doesn't exist, nothing to do
		}
		return fmt.Errorf("failed to get parent task: %w", err)
	}

	// Unmarshal current children
	var children []string
	if err := json.Unmarshal(childrenJSON, &children); err != nil {
		return fmt.Errorf("failed to unmarshal children: %w", err)
	}

	// Remove the child
	for i, child := range children {
		if child == childID {
			children = append(children[:i], children[i+1:]...)
			break
		}
	}

	// Marshal back to JSON
	newChildrenJSON, err := json.Marshal(children)
	if err != nil {
		return fmt.Errorf("failed to marshal children: %w", err)
	}

	// Update the parent
	updateSQL := "UPDATE tasks SET children = $1, updated_at = NOW() WHERE id = $2"
	_, err = tx.Exec(updateSQL, newChildrenJSON, parentID)
	if err != nil {
		return fmt.Errorf("failed to update parent task: %w", err)
	}

	return nil
}

// nullString converts a string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// GetNextDisplayID generates and returns the next sequential display ID for a project
func (ps *PostgresStorage) GetNextDisplayID(projectID string) (string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Begin transaction for atomic counter increment
	tx, err := ps.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get project and increment counter in one atomic operation
	var displayPrefix string
	var counter int

	updateSQL := `
		UPDATE projects 
		SET task_counter = task_counter + 1, updated_at = NOW()
		WHERE id = $1
		RETURNING display_prefix, task_counter`

	err = tx.QueryRow(updateSQL, projectID).Scan(&displayPrefix, &counter)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("project not found: %s", projectID)
		}
		return "", fmt.Errorf("failed to increment counter: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Format and return the display ID
	return fmt.Sprintf("%s-%d", displayPrefix, counter), nil
}

// getOrCreateDefaultProjectTx gets or creates a default project within a transaction
func (ps *PostgresStorage) getOrCreateDefaultProjectTx(tx *sql.Tx) (*models.Project, error) {
	// Check if default project already exists
	selectSQL := `SELECT id, name, description, display_prefix, settings, created_at, updated_at FROM projects WHERE name = $1`
	row := tx.QueryRow(selectSQL, "Default Project")

	var project models.Project
	var settingsJSON []byte

	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.DisplayPrefix,
		&settingsJSON,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == nil {
		// Project exists, unmarshal settings and return
		if err := json.Unmarshal(settingsJSON, &project.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
		return &project, nil
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query default project: %w", err)
	}

	// Default project doesn't exist, create it
	defaultProject := models.NewProject("Default Project", "Default project for tasks without explicit project assignment", "PF")
	defaultProject.ID = uuid.New().String()

	// Serialize settings to JSON
	settingsJSON, err = json.Marshal(defaultProject.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Insert the default project
	insertSQL := `
		INSERT INTO projects (id, name, description, display_prefix, task_counter, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(insertSQL,
		defaultProject.ID,
		defaultProject.Name,
		defaultProject.Description,
		defaultProject.DisplayPrefix,
		0, // Initialize task_counter to 0
		settingsJSON,
		defaultProject.CreatedAt,
		defaultProject.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert default project: %w", err)
	}

	return defaultProject, nil
}

// getNextDisplayIDTx generates the next display ID within a transaction
func (ps *PostgresStorage) getNextDisplayIDTx(tx *sql.Tx, projectID string) (string, error) {
	// Get project and increment counter in one atomic operation
	var displayPrefix string
	var counter int

	updateSQL := `
		UPDATE projects 
		SET task_counter = task_counter + 1, updated_at = NOW()
		WHERE id = $1
		RETURNING display_prefix, task_counter`

	err := tx.QueryRow(updateSQL, projectID).Scan(&displayPrefix, &counter)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("project not found: %s", projectID)
		}
		return "", fmt.Errorf("failed to increment counter: %w", err)
	}

	// Format and return the display ID
	return fmt.Sprintf("%s-%d", displayPrefix, counter), nil
}

// initializeRLS sets up Row-Level Security policies and functions
func (ps *PostgresStorage) initializeRLS() error {
	// Read and execute RLS migration SQL
	rlsStatements := []string{
		// Enable RLS on all tenant-aware tables
		"ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;",
		"ALTER TABLE projects ENABLE ROW LEVEL SECURITY;",
		"ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;",
		"ALTER TABLE users ENABLE ROW LEVEL SECURITY;",

		// Create tenant context management functions
		`CREATE OR REPLACE FUNCTION get_current_tenant_id() RETURNS VARCHAR(36) AS $$
		BEGIN
			RETURN current_setting('app.current_tenant_id', true);
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,

		`CREATE OR REPLACE FUNCTION set_current_tenant_id(tenant_id VARCHAR(36)) RETURNS VOID AS $$
		BEGIN
			PERFORM set_config('app.current_tenant_id', tenant_id, true);
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,

		`CREATE OR REPLACE FUNCTION is_admin_user() RETURNS BOOLEAN AS $$
		BEGIN
			RETURN current_setting('app.is_admin', true)::BOOLEAN;
		EXCEPTION
			WHEN OTHERS THEN
				RETURN FALSE;
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,

		`CREATE OR REPLACE FUNCTION set_admin_context(is_admin BOOLEAN) RETURNS VOID AS $$
		BEGIN
			PERFORM set_config('app.is_admin', is_admin::TEXT, true);
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,

		`CREATE OR REPLACE FUNCTION init_tenant_context(tenant_id VARCHAR(36), is_admin BOOLEAN DEFAULT FALSE) RETURNS VOID AS $$
		BEGIN
			PERFORM set_current_tenant_id(tenant_id);
			IF is_admin THEN
				PERFORM set_admin_context(TRUE);
			ELSE
				PERFORM set_admin_context(FALSE);
			END IF;
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,

		`CREATE OR REPLACE FUNCTION clear_tenant_context() RETURNS VOID AS $$
		BEGIN
			PERFORM set_config('app.current_tenant_id', '', true);
			PERFORM set_config('app.is_admin', 'false', true);
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;`,
	}

	// Execute RLS setup statements
	for _, stmt := range rlsStatements {
		if _, err := ps.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute RLS statement: %w", err)
		}
	}

	// Create RLS policies (drop existing policies first to avoid conflicts)
	policies := []string{
		// Drop existing policies if they exist
		"DROP POLICY IF EXISTS tenant_isolation_policy ON tenants;",
		"DROP POLICY IF EXISTS project_tenant_isolation_policy ON projects;",
		"DROP POLICY IF EXISTS task_tenant_isolation_policy ON tasks;",
		"DROP POLICY IF EXISTS user_tenant_isolation_policy ON users;",

		// Create new policies
		`CREATE POLICY tenant_isolation_policy ON tenants
			FOR ALL TO PUBLIC
			USING (is_admin_user() OR id = get_current_tenant_id());`,

		`CREATE POLICY project_tenant_isolation_policy ON projects
			FOR ALL TO PUBLIC  
			USING (is_admin_user() OR tenant_id = get_current_tenant_id());`,

		`CREATE POLICY task_tenant_isolation_policy ON tasks
			FOR ALL TO PUBLIC
			USING (is_admin_user() OR tenant_id = get_current_tenant_id());`,

		`CREATE POLICY user_tenant_isolation_policy ON users
			FOR ALL TO PUBLIC
			USING (is_admin_user() OR tenant_id = get_current_tenant_id());`,
	}

	// Execute RLS policies
	for _, policy := range policies {
		if _, err := ps.db.Exec(policy); err != nil {
			return fmt.Errorf("failed to create RLS policy: %w", err)
		}
	}

	return nil
}

// setTenantContext sets the tenant context for the current database transaction
func (ps *PostgresStorage) setTenantContext(tx *sql.Tx, tenantID string) error {
	_, err := tx.Exec("SELECT init_tenant_context($1, false)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}

// setTenantContextDB sets the tenant context for the current database connection
func (ps *PostgresStorage) setTenantContextDB(tenantID string) error {
	_, err := ps.db.Exec("SELECT init_tenant_context($1, false)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}

// setAdminContext sets admin context to bypass RLS policies
func (ps *PostgresStorage) setAdminContext(tx *sql.Tx) error {
	_, err := tx.Exec("SELECT set_admin_context(true)")
	if err != nil {
		return fmt.Errorf("failed to set admin context: %w", err)
	}
	return nil
}

// setAdminContextDB sets admin context to bypass RLS policies for DB connection
func (ps *PostgresStorage) setAdminContextDB() error {
	_, err := ps.db.Exec("SELECT set_admin_context(true)")
	if err != nil {
		return fmt.Errorf("failed to set admin context: %w", err)
	}
	return nil
}

// clearTenantContext clears the tenant context
func (ps *PostgresStorage) clearTenantContext(tx *sql.Tx) error {
	_, err := tx.Exec("SELECT clear_tenant_context()")
	if err != nil {
		return fmt.Errorf("failed to clear tenant context: %w", err)
	}
	return nil
}

// clearTenantContextDB clears the tenant context for DB connection
func (ps *PostgresStorage) clearTenantContextDB() error {
	_, err := ps.db.Exec("SELECT clear_tenant_context()")
	if err != nil {
		return fmt.Errorf("failed to clear tenant context: %w", err)
	}
	return nil
}

// getOrCreateDefaultTenant ensures a default tenant exists and returns its ID
func (ps *PostgresStorage) getOrCreateDefaultTenant() (string, error) {
	const defaultTenantID = "default-tenant"
	const defaultTenantName = "Default Tenant"

	// Check if default tenant exists
	var existingID string
	querySQL := "SELECT id FROM tenants WHERE id = $1"
	err := ps.db.QueryRow(querySQL, defaultTenantID).Scan(&existingID)

	if err == nil {
		// Default tenant exists
		return existingID, nil
	}

	if err != sql.ErrNoRows {
		// Unexpected error
		return "", fmt.Errorf("failed to check for default tenant: %w", err)
	}

	// Create default tenant
	createSQL := `
		INSERT INTO tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(), NOW())`

	_, err = ps.db.Exec(createSQL, defaultTenantID, defaultTenantName)
	if err != nil {
		return "", fmt.Errorf("failed to create default tenant: %w", err)
	}

	return defaultTenantID, nil
}

// ensureTenantContext sets the tenant context for current operations
// For now, this uses a default tenant to maintain backward compatibility
func (ps *PostgresStorage) ensureTenantContext() error {
	defaultTenantID, err := ps.getOrCreateDefaultTenant()
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %w", err)
	}

	return ps.setTenantContextDB(defaultTenantID)
}

// migrateExistingDataToDefaultTenant ensures all existing data is assigned to the default tenant
func (ps *PostgresStorage) migrateExistingDataToDefaultTenant() error {
	defaultTenantID, err := ps.getOrCreateDefaultTenant()
	if err != nil {
		return fmt.Errorf("failed to get default tenant: %w", err)
	}

	// Update existing projects without tenant_id
	updateProjectsSQL := `
		UPDATE projects 
		SET tenant_id = $1 
		WHERE tenant_id IS NULL`

	_, err = ps.db.Exec(updateProjectsSQL, defaultTenantID)
	if err != nil {
		return fmt.Errorf("failed to update projects with tenant_id: %w", err)
	}

	// Update existing tasks without tenant_id
	updateTasksSQL := `
		UPDATE tasks 
		SET tenant_id = $1 
		WHERE tenant_id IS NULL`

	_, err = ps.db.Exec(updateTasksSQL, defaultTenantID)
	if err != nil {
		return fmt.Errorf("failed to update tasks with tenant_id: %w", err)
	}

	return nil
}
