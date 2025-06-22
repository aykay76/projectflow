package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aykay76/projectflow/internal/config"
	"github.com/aykay76/projectflow/internal/handlers"
	"github.com/aykay76/projectflow/internal/health"
	"github.com/aykay76/projectflow/internal/llm"
	"github.com/aykay76/projectflow/internal/logger"
	"github.com/aykay76/projectflow/internal/metrics"
	"github.com/aykay76/projectflow/internal/middleware"
	"github.com/aykay76/projectflow/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		// Use basic logging before logger is configured
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logging with configuration
	logger.SetupWithConfig(cfg.GetLogLevel(), cfg.GetLogFormat())

	// Log configuration on startup
	cfg.LogConfiguration()

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		slog.Error("Failed to initialize storage", "error", err, "storage_type", cfg.GetStorageType())
		os.Exit(1)
	}

	slog.Info("Storage initialized", "type", cfg.GetStorageType())

	// Initialize LLM service
	llmService, err := llm.NewService(cfg, slog.Default())
	if err != nil {
		slog.Error("Failed to initialize LLM service", "error", err)
		os.Exit(1)
	}

	slog.Info("LLM service initialized", "enabled", llmService.IsEnabled())

	// Initialize metrics
	appMetrics := metrics.NewMetrics()
	slog.Info("Metrics initialized")

	// Initialize handlers
	handler := handlers.NewHandler(store)
	
	// Initialize chat handler
	chatHandler := handlers.NewChatHandler(store, llmService, slog.Default())

	// Initialize health checker
	healthChecker := health.NewHealthChecker(store, "1.0.0")
	healthChecker.SetLLMService(llmService)

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/health", healthChecker.HandleHealth)
	mux.HandleFunc("/ready", healthChecker.HandleReady)

	// Metrics endpoint for Prometheus scraping
	mux.Handle("/metrics", promhttp.Handler())

	// API routes
	mux.HandleFunc("/api/tasks", handler.HandleTasks)
	mux.HandleFunc("/api/tasks/by-display-id/", handler.HandleTaskByDisplayID)
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

	// Project API routes
	mux.HandleFunc("/api/projects", handler.HandleProjects)
	mux.HandleFunc("/api/projects/", handler.HandleProject)

	// Chat API routes
	mux.HandleFunc("/api/chat", chatHandler.HandleChat)
	mux.HandleFunc("/api/chat/history", chatHandler.HandleChatHistory)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Web interface
	mux.HandleFunc("/", handler.HandleIndex)

	// Apply middleware: Request ID and Metrics collection
	handlerWithMiddleware := middleware.RequestIDMiddleware(
		middleware.MetricsMiddleware(appMetrics)(mux),
	)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handlerWithMiddleware,
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)

	// Register for shutdown signals
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		slog.Info("Server starting", "port", cfg.Port, "shutdown_timeout_seconds", cfg.ShutdownTimeout)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err, "port", cfg.Port)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-quit
	slog.Info("Shutdown signal received, starting graceful shutdown")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed, forcing shutdown", "error", err, "timeout_seconds", cfg.ShutdownTimeout)

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
