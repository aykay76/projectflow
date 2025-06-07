package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aykay76/projectflow/internal/mcp"
	"github.com/aykay76/projectflow/internal/storage"
)

func main() {
	// Initialize storage
	storageDir := getEnv("STORAGE_DIR", "./data")
	store, err := storage.NewFileStorage(storageDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize MCP server
	mcpServer := mcp.NewMCPServer(store)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping MCP server...")
		cancel()
	}()

	// Start the MCP server
	log.Println("Starting ProjectFlow MCP server...")
	if err := mcpServer.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("MCP server error: %v", err)
	}

	log.Println("ProjectFlow MCP server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
