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
	// Create tasks table with proper JSON support for children array
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(36) PRIMARY KEY,
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
		FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE SET NULL
	);`

	if _, err := ps.db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_type ON tasks(type);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);",
	}

	for _, indexSQL := range indexes {
		if _, err := ps.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// CreateTask creates a new task and assigns it an ID
func (ps *PostgresStorage) CreateTask(task *models.Task) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Generate UUID for new task
	task.ID = uuid.New().String()

	// Begin transaction
	tx, err := ps.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Serialize children array to JSON
	childrenJSON, err := json.Marshal(task.Children)
	if err != nil {
		return fmt.Errorf("failed to marshal children array: %w", err)
	}

	// Insert the task
	insertSQL := `
		INSERT INTO tasks (id, title, description, status, priority, type, parent_id, children, started_at, due_date, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err = tx.Exec(insertSQL,
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
		task.CreatedAt,
		task.UpdatedAt,
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

// ListTasks returns all tasks
func (ps *PostgresStorage) ListTasks() ([]*models.Task, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

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
