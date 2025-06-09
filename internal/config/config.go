package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port            string `env:"PORT" default:"8080"`
	ShutdownTimeout int    `env:"SHUTDOWN_TIMEOUT" default:"30"`

	// Storage configuration
	StorageType string `env:"STORAGE_TYPE" default:"file"`
	DataDir     string `env:"DATA_DIR" default:"./data"`

	// Database configuration (PostgreSQL)
	DatabaseHost     string `env:"DB_HOST" default:"localhost"`
	DatabasePort     string `env:"DB_PORT" default:"5432"`
	DatabaseName     string `env:"DB_NAME" default:"projectflow"`
	DatabaseUser     string `env:"DB_USER" default:"projectflow"`
	DatabasePassword string `env:"DB_PASSWORD" default:""`
	DatabaseSSLMode  string `env:"DB_SSL_MODE" default:"prefer"`

	// Logging configuration
	LogLevel  string `env:"LOG_LEVEL" default:"INFO"`
	LogFormat string `env:"LOG_FORMAT" default:"text"`
}

// Load loads configuration from environment variables with validation
func Load() (*Config, error) {
	config := &Config{
		Port:            getEnv("PORT", "8080"),
		ShutdownTimeout: getEnvInt("SHUTDOWN_TIMEOUT", 30),
		StorageType:     getEnv("STORAGE_TYPE", "file"),
		DataDir:         getEnv("DATA_DIR", "./data"),
		DatabaseHost:    getEnv("DB_HOST", "localhost"),
		DatabasePort:    getEnv("DB_PORT", "5432"),
		DatabaseName:    getEnv("DB_NAME", "projectflow"),
		DatabaseUser:    getEnv("DB_USER", "projectflow"),
		DatabasePassword: getEnv("DB_PASSWORD", ""),
		DatabaseSSLMode: getEnv("DB_SSL_MODE", "prefer"),
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
		LogFormat:       getEnv("LOG_FORMAT", "text"),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Validate port
	if c.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be a valid port number (1-65535), got: %s", c.Port)
	}

	// Validate shutdown timeout
	if c.ShutdownTimeout < 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be non-negative, got: %d", c.ShutdownTimeout)
	}
	if c.ShutdownTimeout > 300 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT should not exceed 300 seconds, got: %d", c.ShutdownTimeout)
	}

	// Validate data directory
	if c.DataDir == "" {
		return fmt.Errorf("DATA_DIR cannot be empty")
	}

	// Validate storage type
	validStorageTypes := []string{"file", "postgres", "postgresql"}
	storageType := strings.ToLower(c.StorageType)
	var validStorage bool
	for _, sType := range validStorageTypes {
		if storageType == sType {
			validStorage = true
			break
		}
	}
	if !validStorage {
		return fmt.Errorf("STORAGE_TYPE must be one of %v, got: %s", validStorageTypes, c.StorageType)
	}

	// Validate database configuration if using database storage
	if storageType == "postgres" || storageType == "postgresql" {
		if c.DatabaseHost == "" {
			return fmt.Errorf("DB_HOST cannot be empty when using database storage")
		}
		if c.DatabasePort == "" {
			return fmt.Errorf("DB_PORT cannot be empty when using database storage")
		}
		if port, err := strconv.Atoi(c.DatabasePort); err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("DB_PORT must be a valid port number (1-65535), got: %s", c.DatabasePort)
		}
		if c.DatabaseName == "" {
			return fmt.Errorf("DB_NAME cannot be empty when using database storage")
		}
		if c.DatabaseUser == "" {
			return fmt.Errorf("DB_USER cannot be empty when using database storage")
		}
		// Password can be empty for trust authentication or .pgpass
		validSSLModes := []string{"disable", "require", "verify-ca", "verify-full", "prefer", "allow"}
		var validSSL bool
		for _, mode := range validSSLModes {
			if c.DatabaseSSLMode == mode {
				validSSL = true
				break
			}
		}
		if !validSSL {
			return fmt.Errorf("DB_SSL_MODE must be one of %v, got: %s", validSSLModes, c.DatabaseSSLMode)
		}
	}

	// Validate log level
	validLogLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	logLevel := strings.ToUpper(c.LogLevel)
	var validLevel bool
	for _, level := range validLogLevels {
		if logLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("LOG_LEVEL must be one of %v, got: %s", validLogLevels, c.LogLevel)
	}

	// Validate log format
	validLogFormats := []string{"json", "text"}
	logFormat := strings.ToLower(c.LogFormat)
	var validFormat bool
	for _, format := range validLogFormats {
		if logFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("LOG_FORMAT must be one of %v, got: %s", validLogFormats, c.LogFormat)
	}

	return nil
}

// LogConfiguration logs the current configuration (without sensitive data)
func (c *Config) LogConfiguration() {
	slog.Info("Application configuration loaded",
		"port", c.Port,
		"shutdown_timeout_seconds", c.ShutdownTimeout,
		"storage_type", c.StorageType,
		"data_dir", c.DataDir,
		"db_host", c.DatabaseHost,
		"db_port", c.DatabasePort,
		"db_name", c.DatabaseName,
		"db_user", c.DatabaseUser,
		"db_ssl_mode", c.DatabaseSSLMode,
		"log_level", c.LogLevel,
		"log_format", c.LogFormat,
	)
}

// GetPort returns the port as an integer
func (c *Config) GetPort() int {
	port, _ := strconv.Atoi(c.Port) // Already validated in Validate()
	return port
}

// GetLogLevel returns the log level as slog.Level
func (c *Config) GetLogLevel() slog.Level {
	switch strings.ToUpper(c.LogLevel) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default fallback
	}
}

// GetLogFormat returns whether to use JSON format
func (c *Config) GetLogFormat() string {
	return strings.ToLower(c.LogFormat)
}

// GetStorageType returns the normalized storage type
func (c *Config) GetStorageType() string {
	storageType := strings.ToLower(c.StorageType)
	if storageType == "postgresql" {
		return "postgres"
	}
	return storageType
}

// GetDatabasePort returns the database port as an integer
func (c *Config) GetDatabasePort() int {
	port, _ := strconv.Atoi(c.DatabasePort) // Already validated in Validate()
	return port
}

// GetDatabaseConnectionString builds a PostgreSQL connection string
func (c *Config) GetDatabaseConnectionString() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("host=%s", c.DatabaseHost))
	parts = append(parts, fmt.Sprintf("port=%s", c.DatabasePort))
	parts = append(parts, fmt.Sprintf("dbname=%s", c.DatabaseName))
	parts = append(parts, fmt.Sprintf("user=%s", c.DatabaseUser))
	
	if c.DatabasePassword != "" {
		parts = append(parts, fmt.Sprintf("password=%s", c.DatabasePassword))
	}
	
	parts = append(parts, fmt.Sprintf("sslmode=%s", c.DatabaseSSLMode))
	
	return strings.Join(parts, " ")
}

// getEnv returns environment variable value or default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns environment variable as integer or default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
