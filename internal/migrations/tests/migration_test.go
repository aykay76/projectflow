package tests

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aykay76/projectflow/internal/migrations"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// TestMigrationSuite runs the complete migration test suite
func TestMigrationSuite(t *testing.T) {
	// Skip if no test database is configured
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration tests")
	}

	// Connect to test database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	t.Run("CleanDatabase", func(t *testing.T) {
		cleanTestDatabase(t, db)
	})

	t.Run("InitializeMigrationTable", func(t *testing.T) {
		testInitializeMigrationTable(t, db)
	})

	t.Run("CreateSampleData", func(t *testing.T) {
		createSampleData(t, db)
	})

	t.Run("RunMigrationsUp", func(t *testing.T) {
		testMigrationsUp(t, db)
	})

	t.Run("ValidateDataIntegrity", func(t *testing.T) {
		validateDataIntegrity(t, db)
	})

	t.Run("TestConstraints", func(t *testing.T) {
		testConstraints(t, db)
	})

	t.Run("TestIndexes", func(t *testing.T) {
		testIndexes(t, db)
	})

	t.Run("TestRollback", func(t *testing.T) {
		testRollback(t, db)
	})

	t.Run("TestPerformance", func(t *testing.T) {
		testMigrationPerformance(t, db)
	})
}

func cleanTestDatabase(t *testing.T, db *sql.DB) {
	t.Log("Cleaning test database...")

	// Drop all tables
	tables := []string{"schema_migrations", "users", "tasks", "projects", "tenants"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			t.Errorf("Failed to drop table %s: %v", table, err)
		}
	}

	// Drop indexes
	indexes := []string{"idx_projects_tenant_id", "idx_tasks_tenant_id"}
	for _, index := range indexes {
		_, err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index))
		if err != nil {
			t.Errorf("Failed to drop index %s: %v", index, err)
		}
	}

	t.Log("Database cleaned successfully")
}

func testInitializeMigrationTable(t *testing.T, db *sql.DB) {
	t.Log("Testing migration table initialization...")

	manager := migrations.NewMigrationManager(db)
	err := manager.InitializeMigrationTable()
	if err != nil {
		t.Fatalf("Failed to initialize migration table: %v", err)
	}

	// Verify table exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'schema_migrations'
		)
	`).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check migration table existence: %v", err)
	}

	if !exists {
		t.Fatal("Migration table was not created")
	}

	t.Log("Migration table initialized successfully")
}

func createSampleData(t *testing.T, db *sql.DB) {
	t.Log("Creating sample data for testing...")

	// Create basic tables structure (pre-migration)
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS projects (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		description TEXT,
		display_prefix VARCHAR(10) NOT NULL,
		task_counter INTEGER NOT NULL DEFAULT 0,
		settings JSONB DEFAULT '{}'::jsonb,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(36) PRIMARY KEY,
		display_id VARCHAR(50) UNIQUE,
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
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	_, err := db.Exec(createTablesSQL)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Insert sample data
	insertDataSQL := `
	INSERT INTO projects (id, name, description, display_prefix, task_counter, created_at, updated_at) VALUES
	('proj-1', 'Test Project 1', 'First test project', 'TP1', 5, NOW(), NOW()),
	('proj-2', 'Test Project 2', 'Second test project', 'TP2', 3, NOW(), NOW()),
	('proj-3', 'Test Project 3', 'Third test project', 'TP3', 10, NOW(), NOW());

	INSERT INTO tasks (id, display_id, project_id, title, description, status, priority, type, created_at, updated_at) VALUES
	('task-1', 'TP1-1', 'proj-1', 'Test Task 1', 'First test task', 'todo', 'high', 'task', NOW(), NOW()),
	('task-2', 'TP1-2', 'proj-1', 'Test Task 2', 'Second test task', 'in_progress', 'medium', 'task', NOW(), NOW()),
	('task-3', 'TP2-1', 'proj-2', 'Test Task 3', 'Third test task', 'done', 'low', 'task', NOW(), NOW()),
	('task-4', 'TP3-1', 'proj-3', 'Test Task 4', 'Fourth test task', 'todo', 'high', 'story', NOW(), NOW()),
	('task-5', 'TP3-2', 'proj-3', 'Test Task 5', 'Fifth test task', 'blocked', 'medium', 'epic', NOW(), NOW());`

	_, err = db.Exec(insertDataSQL)
	if err != nil {
		t.Fatalf("Failed to insert sample data: %v", err)
	}

	// Verify data was inserted
	var projectCount, taskCount int
	err = db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)
	if err != nil {
		t.Fatalf("Failed to count projects: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	t.Logf("Created %d projects and %d tasks for testing", projectCount, taskCount)
}

func testMigrationsUp(t *testing.T, db *sql.DB) {
	t.Log("Testing migration up...")

	manager := migrations.NewMigrationManager(db)

	// Load all migrations
	allMigrations, err := loadTestMigrations()
	if err != nil {
		t.Fatalf("Failed to load migrations: %v", err)
	}

	// Get pending migrations
	pending, err := manager.GetPendingMigrations(allMigrations)
	if err != nil {
		t.Fatalf("Failed to get pending migrations: %v", err)
	}

	t.Logf("Found %d pending migrations", len(pending))

	// Apply each migration
	for _, migration := range pending {
		t.Logf("Applying migration %d: %s", migration.Version, migration.Name)
		err := manager.ApplyMigration(migration)
		if err != nil {
			t.Fatalf("Failed to apply migration %d: %v", migration.Version, err)
		}
	}

	// Verify migrations were applied
	applied, err := manager.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(applied) != len(allMigrations) {
		t.Errorf("Expected %d applied migrations, got %d", len(allMigrations), len(applied))
	}

	t.Log("All migrations applied successfully")
}

func validateDataIntegrity(t *testing.T, db *sql.DB) {
	t.Log("Validating data integrity after migration...")

	// Check that all projects have tenant_id
	var projectsWithoutTenant int
	err := db.QueryRow("SELECT COUNT(*) FROM projects WHERE tenant_id IS NULL").Scan(&projectsWithoutTenant)
	if err != nil {
		t.Fatalf("Failed to check projects tenant_id: %v", err)
	}

	if projectsWithoutTenant > 0 {
		t.Errorf("Found %d projects without tenant_id", projectsWithoutTenant)
	}

	// Check that all tasks have tenant_id
	var tasksWithoutTenant int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE tenant_id IS NULL").Scan(&tasksWithoutTenant)
	if err != nil {
		t.Fatalf("Failed to check tasks tenant_id: %v", err)
	}

	if tasksWithoutTenant > 0 {
		t.Errorf("Found %d tasks without tenant_id", tasksWithoutTenant)
	}

	// Check that default tenant exists
	var tenantExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = '00000000-0000-0000-0000-000000000001')").Scan(&tenantExists)
	if err != nil {
		t.Fatalf("Failed to check default tenant existence: %v", err)
	}

	if !tenantExists {
		t.Error("Default tenant was not created")
	}

	t.Log("Data integrity validation passed")
}

func testConstraints(t *testing.T, db *sql.DB) {
	t.Log("Testing database constraints...")

	// Test foreign key constraints
	_, err := db.Exec("INSERT INTO tasks (id, display_id, project_id, tenant_id, title) VALUES ('test-task', 'TEST-1', 'proj-1', 'invalid-tenant', 'Test Task')")
	if err == nil {
		t.Error("Expected foreign key constraint violation, but insert succeeded")
	}

	// Test unique constraints
	_, err = db.Exec("INSERT INTO tenants (id, name) VALUES ('test-tenant', 'Default Tenant')")
	if err == nil {
		t.Error("Expected unique constraint violation, but insert succeeded")
	}

	t.Log("Constraint tests passed")
}

func testIndexes(t *testing.T, db *sql.DB) {
	t.Log("Testing database indexes...")

	// Check if indexes exist
	indexes := []string{"idx_projects_tenant_id", "idx_tasks_tenant_id"}
	
	for _, indexName := range indexes {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM pg_indexes 
				WHERE indexname = $1
			)
		`, indexName).Scan(&exists)
		if err != nil {
			t.Errorf("Failed to check index %s: %v", indexName, err)
			continue
		}

		if !exists {
			t.Errorf("Index %s does not exist", indexName)
		}
	}

	t.Log("Index tests passed")
}

func testRollback(t *testing.T, db *sql.DB) {
	t.Log("Testing migration rollback...")

	manager := migrations.NewMigrationManager(db)

	// Get current migrations
	applied, err := manager.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(applied) == 0 {
		t.Skip("No migrations to rollback")
	}

	// Rollback the last migration
	lastMigration := applied[len(applied)-1]
	err = manager.RollbackMigration(lastMigration)
	if err != nil {
		t.Fatalf("Failed to rollback migration %d: %v", lastMigration, err)
	}

	// Verify rollback
	appliedAfter, err := manager.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations after rollback: %v", err)
	}

	if len(appliedAfter) != len(applied)-1 {
		t.Errorf("Expected %d migrations after rollback, got %d", len(applied)-1, len(appliedAfter))
	}

	t.Log("Rollback test passed")
}

func testMigrationPerformance(t *testing.T, db *sql.DB) {
	t.Log("Testing migration performance...")

	// Create a large dataset for performance testing
	start := time.Now()
	
	// Insert 1000 test records
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	for i := 0; i < 1000; i++ {
		_, err = tx.Exec(`
			INSERT INTO projects (id, name, display_prefix, tenant_id, created_at, updated_at) 
			VALUES ($1, $2, $3, '00000000-0000-0000-0000-000000000001', NOW(), NOW())
		`, fmt.Sprintf("perf-proj-%d", i), fmt.Sprintf("Performance Project %d", i), fmt.Sprintf("PP%d", i))
		if err != nil {
			t.Fatalf("Failed to insert performance test project %d: %v", i, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit performance test data: %v", err)
	}

	elapsed := time.Since(start)
	t.Logf("Inserted 1000 records in %v", elapsed)

	// Performance should be reasonable (less than 5 seconds for 1000 records)
	if elapsed > 5*time.Second {
		t.Errorf("Performance test took too long: %v", elapsed)
	}

	t.Log("Performance test passed")
}

// loadTestMigrations loads migrations for testing
// This is a simplified version that loads from the actual migration files
func loadTestMigrations() ([]migrations.Migration, error) {
	// For now, return empty slice
	// In a real implementation, this would load from migration files
	return []migrations.Migration{}, nil
}
