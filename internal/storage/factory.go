package storage

import (
	"fmt"
	"log/slog"

	"github.com/aykay76/projectflow/internal/config"
)

// NewStorage creates a new storage instance based on the configuration
func NewStorage(cfg *config.Config) (Storage, error) {
	storageType := cfg.GetStorageType()

	slog.Info("Initializing storage", "type", storageType)

	switch storageType {
	case "postgres":
		connectionString := cfg.GetDatabaseConnectionString()
		slog.Debug("Creating PostgreSQL storage", "connection_string", "[REDACTED]")
		return NewPostgresStorage(connectionString)

	case "file":
		slog.Debug("Creating file storage", "data_dir", cfg.DataDir)
		return NewFileStorage(cfg.DataDir)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
