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
	DataDir string `env:"DATA_DIR" default:"./data"`

	// Logging configuration
	LogLevel  string `env:"LOG_LEVEL" default:"INFO"`
	LogFormat string `env:"LOG_FORMAT" default:"text"`
}

// Load loads configuration from environment variables with validation
func Load() (*Config, error) {
	config := &Config{
		Port:            getEnv("PORT", "8080"),
		ShutdownTimeout: getEnvInt("SHUTDOWN_TIMEOUT", 30),
		DataDir:         getEnv("DATA_DIR", "./data"),
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
		"data_dir", c.DataDir,
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
