package main

import (
	"fmt"
	"log"

	"github.com/aykay76/projectflow/internal/storage"
)

func main() {
	// Initialize file storage
	fileStorage, err := storage.NewFileStorage("./data")
	if err != nil {
		log.Fatal("Failed to initialize file storage:", err)
	}
	defer fileStorage.Close()

	// Test display ID lookup for tasks we created earlier
	displayIDs := []string{"PF-1", "PF-2"}

	for _, displayID := range displayIDs {
		task, err := fileStorage.GetTaskByDisplayID(displayID)
		if err != nil {
			fmt.Printf("Error getting task with display ID %s: %v\n", displayID, err)
			continue
		}

		fmt.Printf("Found task with display ID %s:\n", displayID)
		fmt.Printf("  ID: %s\n", task.ID)
		fmt.Printf("  Title: %s\n", task.Title)
		fmt.Printf("  Project ID: %s\n", task.ProjectID)
		fmt.Printf("  Description: %s\n", task.Description)
		fmt.Println()
	}

	// Test with non-existent display ID
	_, err = fileStorage.GetTaskByDisplayID("PF-999")
	if err != nil {
		fmt.Printf("Expected error for non-existent display ID: %v\n", err)
	}
}
