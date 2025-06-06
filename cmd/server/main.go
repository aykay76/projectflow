package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aykay76/projectflow/internal/handlers"
	"github.com/aykay76/projectflow/internal/storage"
)

func main() {
	// Initialize storage
	storageDir := getEnv("STORAGE_DIR", "./data")
	store, err := storage.NewFileStorage(storageDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize handlers
	handler := handlers.NewHandler(store)

	// Setup routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/tasks", handler.HandleTasks)
	mux.HandleFunc("/api/tasks/", handler.HandleTask)
	mux.HandleFunc("/api/hierarchy", handler.HandleHierarchy)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Web interface
	mux.HandleFunc("/", handler.HandleIndex)

	port := getEnv("PORT", "8080")
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
