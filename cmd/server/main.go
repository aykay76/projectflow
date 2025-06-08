package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aykay76/projectflow/internal/handlers"
	"github.com/aykay76/projectflow/internal/health"
	"github.com/aykay76/projectflow/internal/logger"
	"github.com/aykay76/projectflow/internal/storage"
)

func main() {
	// Initialize logging first
	logger.Setup()
	
	// Initialize storage
	storageDir := getEnv("STORAGE_DIR", "./data")
	store, err := storage.NewFileStorage(storageDir)
	if err != nil {
		slog.Error("Failed to initialize storage", "error", err, "storage_dir", storageDir)
		os.Exit(1)
	}
	
	slog.Info("Storage initialized", "storage_dir", storageDir)

	// Initialize handlers
	handler := handlers.NewHandler(store)
	
	// Initialize health checker
	healthChecker := health.NewHealthChecker(store, "1.0.0")

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/health", healthChecker.HandleHealth)
	mux.HandleFunc("/ready", healthChecker.HandleReady)

	// API routes
	mux.HandleFunc("/api/tasks", handler.HandleTasks)
	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		parts := strings.Split(path, "/")

		if len(parts) >= 2 && parts[1] == "children" {
			if len(parts) == 2 {
				// /api/tasks/{id}/children
				handler.HandleTaskChildren(w, r)
			} else if len(parts) == 3 {
				// /api/tasks/{id}/children/{child_id}
				handler.HandleTaskChildRelation(w, r)
			} else {
				http.Error(w, "Invalid URL path", http.StatusBadRequest)
			}
		} else if len(parts) >= 2 && parts[1] == "move" {
			// /api/tasks/{id}/move
			handler.HandleTaskMove(w, r)
		} else if len(parts) == 1 {
			// /api/tasks/{id}
			handler.HandleTask(w, r)
		} else {
			http.Error(w, "Invalid URL path", http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/api/hierarchy", handler.HandleHierarchy)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Web interface
	mux.HandleFunc("/", handler.HandleIndex)

	// Configure server
	port := getEnv("PORT", "8080")
	shutdownTimeoutStr := getEnv("SHUTDOWN_TIMEOUT", "30")
	shutdownTimeout, err := strconv.Atoi(shutdownTimeoutStr)
	if err != nil {
		slog.Warn("Invalid SHUTDOWN_TIMEOUT value, using default", "value", shutdownTimeoutStr, "default", 30)
		shutdownTimeout = 30
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	
	// Register for shutdown signals
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		slog.Info("Server starting", "port", port, "shutdown_timeout_seconds", shutdownTimeout)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err, "port", port)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-quit
	slog.Info("Shutdown signal received, starting graceful shutdown")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeout)*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed, forcing shutdown", "error", err, "timeout_seconds", shutdownTimeout)
		
		// Close storage before exit
		if closeErr := store.Close(); closeErr != nil {
			slog.Error("Failed to close storage during forced shutdown", "error", closeErr)
		}
		
		os.Exit(1)
	}

	// Close storage connections
	if err := store.Close(); err != nil {
		slog.Error("Failed to close storage during shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server shutdown completed successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
