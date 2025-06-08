package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Setup configures the global slog logger based on environment variables
func Setup() {
	logLevel := getLogLevelFromEnv()

	// Create JSON handler for structured logging
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	// Use JSON handler for production, text handler for development
	var handler slog.Handler
	logFormat := os.Getenv("LOG_FORMAT")
	if strings.ToLower(logFormat) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// Default to JSON for production
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log the configuration
	slog.Info("Logger configured",
		"level", logLevel.String(),
		"format", getLogFormat(),
	)
}

// getLogLevelFromEnv parses the LOG_LEVEL environment variable
func getLogLevelFromEnv() slog.Level {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getLogFormat returns the current log format
func getLogFormat() string {
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "" {
		return "json"
	}
	return strings.ToLower(logFormat)
}

// Convenience functions that wrap slog for easier usage

// Debug logs a debug message with optional attributes
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// Info logs an info message with optional attributes
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Warn logs a warning message with optional attributes
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error logs an error message with optional attributes
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// WithGroup returns a Logger that starts a group
func WithGroup(name string) *slog.Logger {
	return slog.Default().WithGroup(name)
}

// With returns a Logger that includes the given attributes
func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}
