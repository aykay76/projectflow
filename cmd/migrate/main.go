package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/aykay76/projectflow/internal/config"
	"github.com/aykay76/projectflow/internal/migrations"
	_ "github.com/lib/pq" // PostgreSQL driver
)

const usage = `
Migration Tool for ProjectFlow

Usage:
  migrate [command] [options]

Commands:
  init     - Initialize migration tracking table
  up       - Apply all pending migrations
  down     - Rollback to specified version
  status   - Show migration status
  create   - Create a new migration file

Options:
  -version int
        Target version for rollback (used with 'down' command)
  -name string
        Migration name (used with 'create' command)

Examples:
  migrate init
  migrate up
  migrate down -version 20240101120000
  migrate status
  migrate create -name "add_tenant_support"

Configuration is loaded from environment variables.
`

func main() {
	var (
		version = flag.Int("version", 0, "Target version for rollback")
		name    = flag.String("name", "", "Migration name")
	)
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(1)
	}

	command := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.GetDatabaseConnectionString())
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("❌ Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Create migration manager
	manager := migrations.NewMigrationManager(db)

	switch command {
	case "init":
		if err := initCommand(manager); err != nil {
			fmt.Printf("❌ Init failed: %v\n", err)
			os.Exit(1)
		}

	case "up":
		if err := upCommand(manager); err != nil {
			fmt.Printf("❌ Migration failed: %v\n", err)
			os.Exit(1)
		}

	case "down":
		if *version == 0 {
			fmt.Println("❌ Version is required for rollback")
			os.Exit(1)
		}
		if err := downCommand(manager, *version); err != nil {
			fmt.Printf("❌ Rollback failed: %v\n", err)
			os.Exit(1)
		}

	case "status":
		if err := statusCommand(manager); err != nil {
			fmt.Printf("❌ Status failed: %v\n", err)
			os.Exit(1)
		}

	case "create":
		if *name == "" {
			fmt.Println("❌ Migration name is required")
			os.Exit(1)
		}
		if err := createCommand(*name); err != nil {
			fmt.Printf("❌ Create failed: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func initCommand(manager *migrations.MigrationManager) error {
	fmt.Println("🔧 Initializing migration tracking table...")
	if err := manager.InitializeMigrationTable(); err != nil {
		return err
	}
	fmt.Println("✅ Migration tracking table initialized")
	return nil
}

func upCommand(manager *migrations.MigrationManager) error {
	fmt.Println("🚀 Applying pending migrations...")

	// Load all migrations
	allMigrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get pending migrations
	pending, err := manager.GetPendingMigrations(allMigrations)
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		fmt.Println("✅ No pending migrations")
		return nil
	}

	fmt.Printf("📋 Found %d pending migrations\n", len(pending))

	// Apply each migration
	for _, migration := range pending {
		if err := manager.ApplyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	fmt.Printf("✅ Successfully applied %d migrations\n", len(pending))
	return nil
}

func downCommand(manager *migrations.MigrationManager, targetVersion int) error {
	fmt.Printf("⏪ Rolling back to version %d...\n", targetVersion)

	// Get migrations to rollback
	toRollback, err := manager.GetMigrationsToRollback(targetVersion)
	if err != nil {
		return fmt.Errorf("failed to get migrations to rollback: %w", err)
	}

	if len(toRollback) == 0 {
		fmt.Println("✅ Already at target version or newer")
		return nil
	}

	fmt.Printf("📋 Rolling back %d migrations\n", len(toRollback))

	// Rollback migrations in reverse order
	for _, version := range toRollback {
		if err := manager.RollbackMigration(version); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", version, err)
		}
	}

	fmt.Printf("✅ Successfully rolled back to version %d\n", targetVersion)
	return nil
}

func statusCommand(manager *migrations.MigrationManager) error {
	fmt.Println("📊 Migration Status")
	fmt.Println("==================")

	// Get applied migrations
	applied, err := manager.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Load all migrations
	allMigrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get pending migrations
	pending, err := manager.GetPendingMigrations(allMigrations)
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	fmt.Printf("Applied: %d migrations\n", len(applied))
	fmt.Printf("Pending: %d migrations\n", len(pending))

	if len(applied) > 0 {
		fmt.Println("\nApplied Migrations:")
		for _, version := range applied {
			fmt.Printf("  ✅ %d\n", version)
		}
	}

	if len(pending) > 0 {
		fmt.Println("\nPending Migrations:")
		for _, migration := range pending {
			fmt.Printf("  ⏳ %d - %s\n", migration.Version, migration.Name)
		}
	}

	return nil
}

func createCommand(name string) error {
	version := migrations.GenerateTimestampVersion()
	filename := fmt.Sprintf("%d_%s.sql", version, name)
	filepath := fmt.Sprintf("internal/migrations/scripts/%s", filename)

	template := fmt.Sprintf(`-- Migration: %s
-- Version: %d
-- Description: %s

-- +migrate Up
-- Add your up migration SQL here


-- +migrate Down
-- Add your down migration SQL here

`, name, version, name)

	if err := os.WriteFile(filepath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	fmt.Printf("✅ Created migration file: %s\n", filepath)
	return nil
}

// loadMigrations loads all migration files from the scripts directory
// This is a placeholder - in the real implementation, you would parse actual .sql files
func loadMigrations() ([]migrations.Migration, error) {
	// For now, return an empty slice
	// TODO: Implement file loading logic in next iteration
	return []migrations.Migration{}, nil
}
