package migrations

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	UpSQL       string
	DownSQL     string
	Description string
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db *sql.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{
		db: db,
	}
}

// InitializeMigrationTable creates the migration tracking table
func (mm *MigrationManager) InitializeMigrationTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		rollback_sql TEXT NOT NULL
	);`

	_, err := mm.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns a list of applied migration versions
func (mm *MigrationManager) GetAppliedMigrations() ([]int, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := mm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// ApplyMigration executes a migration and records it
func (mm *MigrationManager) ApplyMigration(migration Migration) error {
	// Start transaction
	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute the migration
	_, err = tx.Exec(migration.UpSQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration %d (%s): %w", migration.Version, migration.Name, err)
	}

	// Record the migration
	recordSQL := `
	INSERT INTO schema_migrations (version, name, description, rollback_sql) 
	VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(recordSQL, migration.Version, migration.Name, migration.Description, migration.DownSQL)
	if err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	fmt.Printf("✅ Applied migration %d: %s\n", migration.Version, migration.Name)
	return nil
}

// RollbackMigration reverses a migration
func (mm *MigrationManager) RollbackMigration(version int) error {
	// Get rollback SQL
	var rollbackSQL, name string
	query := "SELECT rollback_sql, name FROM schema_migrations WHERE version = $1"
	err := mm.db.QueryRow(query, version).Scan(&rollbackSQL, &name)
	if err != nil {
		return fmt.Errorf("failed to get rollback SQL for migration %d: %w", version, err)
	}

	// Start transaction
	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback
	_, err = tx.Exec(rollbackSQL)
	if err != nil {
		return fmt.Errorf("failed to execute rollback for migration %d (%s): %w", version, name, err)
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = $1", version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", version, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %d: %w", version, err)
	}

	fmt.Printf("⏪ Rolled back migration %d: %s\n", version, name)
	return nil
}

// GetPendingMigrations returns migrations that haven't been applied yet
func (mm *MigrationManager) GetPendingMigrations(allMigrations []Migration) ([]Migration, error) {
	appliedVersions, err := mm.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup
	applied := make(map[int]bool)
	for _, version := range appliedVersions {
		applied[version] = true
	}

	// Find pending migrations
	var pending []Migration
	for _, migration := range allMigrations {
		if !applied[migration.Version] {
			pending = append(pending, migration)
		}
	}

	// Sort by version
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return pending, nil
}

// GetMigrationsToRollback returns migrations that need to be rolled back to reach target version
func (mm *MigrationManager) GetMigrationsToRollback(targetVersion int) ([]int, error) {
	appliedVersions, err := mm.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	var toRollback []int
	for _, version := range appliedVersions {
		if version > targetVersion {
			toRollback = append(toRollback, version)
		}
	}

	// Sort in descending order (rollback newest first)
	sort.Sort(sort.Reverse(sort.IntSlice(toRollback)))

	return toRollback, nil
}

// ValidateMigrations checks that migration versions are unique and sequential
func ValidateMigrations(migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Check for duplicates and gaps
	seen := make(map[int]bool)
	for _, migration := range migrations {
		if seen[migration.Version] {
			return fmt.Errorf("duplicate migration version: %d", migration.Version)
		}
		seen[migration.Version] = true
	}

	return nil
}

// GenerateTimestampVersion generates a timestamp-based version number
func GenerateTimestampVersion() int {
	// Format: YYYYMMDDHHMMSS (14 digits)
	now := time.Now().UTC()
	versionStr := now.Format("20060102150405")
	version, _ := strconv.Atoi(versionStr)
	return version
}

// ParseMigrationVersion extracts version from migration filename
// Expected format: {version}_{name}.sql
func ParseMigrationVersion(filename string) (int, string, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid version number in filename %s: %w", filename, err)
	}

	name := strings.TrimSuffix(parts[1], ".sql")
	return version, name, nil
}
