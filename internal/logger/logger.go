package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Setup configures the global slog logger based on environment variables
func Setup() {
	logLevel := getLogLevelFromEnv()
	logFormat := getLogFormat()

	// Create JSON handler for structured logging
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	// Use text handler for development, JSON handler for production
	var handler slog.Handler
	if logFormat == "text" {
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
		"format", logFormat,
	)
}

// SetupWithConfig configures the global slog logger using provided configuration
func SetupWithConfig(level slog.Level, format string) {
	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Use text handler for development, JSON handler for production
	var handler slog.Handler
	if strings.ToLower(format) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// Default to JSON for production
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log the configuration
	slog.Info("Logger configured",
		"level", level.String(),
		"format", strings.ToLower(format),
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

// Context-aware logging functions for request correlation

// DebugContext logs a debug message with context and optional attributes
func DebugContext(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context and optional attributes
func InfoContext(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context and optional attributes
func WarnContext(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context and optional attributes
func ErrorContext(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}
