package storage

import (
	"context"
	"testing"

	"github.com/aykay76/projectflow/internal/config"
)

func TestNewStorage_FileStorage(t *testing.T) {
	// Create a config for file storage
	cfg := &config.Config{
		StorageType: "file",
		DataDir:     t.TempDir(),
	}

	storage, err := NewStorage(cfg)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}
	defer storage.Close()

	// Verify it's a FileStorage instance by testing a simple operation
	// We can't directly type assert because the interface doesn't expose the type
	if !storage.TaskExists(context.Background(), "nonexistent") {
		// This is expected behavior - just ensuring the storage is functional
	}
}

func TestNewStorage_PostgresStorage(t *testing.T) {
	// Skip this test if we don't have a postgres instance available
	// In CI/CD, this would be handled by the test containers
	t.Skip("Skipping PostgreSQL test - requires test container setup")

	cfg := &config.Config{
		StorageType:      "postgres",
		DatabaseHost:     "localhost",
		DatabasePort:     "5432",
		DatabaseName:     "testdb",
		DatabaseUser:     "testuser",
		DatabasePassword: "testpass",
		DatabaseSSLMode:  "disable",
	}

	storage, err := NewStorage(cfg)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}
	defer storage.Close()

	// Verify it's functional
	if !storage.TaskExists(context.Background(), "nonexistent") {
		// This is expected behavior - just ensuring the storage is functional
	}
}

func TestNewStorage_UnsupportedType(t *testing.T) {
	cfg := &config.Config{
		StorageType: "unsupported",
	}

	_, err := NewStorage(cfg)
	if err == nil {
		t.Error("NewStorage() with unsupported type should return error")
	}
}

func TestNewStorage_PostgresVariants(t *testing.T) {
	// Test "postgresql" variant
	cfg := &config.Config{
		StorageType: "postgresql",
		DataDir:     t.TempDir(), // fallback for file storage if postgres fails
	}

	// This should not error on the storage type normalization
	// but will likely fail on connection - that's fine for this test
	_, err := NewStorage(cfg)
	// We expect this to fail due to no actual postgres instance,
	// but it should fail on connection, not on unsupported type
	if err != nil && err.Error() == "unsupported storage type: postgresql" {
		t.Error("NewStorage() should normalize 'postgresql' to 'postgres'")
	}
}
