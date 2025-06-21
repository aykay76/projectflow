package main

import (
	"fmt"
	"log"

	"github.com/aykay76/projectflow/internal/storage"
)

func main() {
	store, err := storage.NewFileStorage("./data")
	if err != nil {
		log.Fatal(err)
	}

	tasks, err := store.ListTasks("PF")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d tasks in PF project\n", len(tasks))
	if len(tasks) > 0 {
		fmt.Printf("First task: %s - %s\n", tasks[0].DisplayID, tasks[0].Title)
	}

	// Test with empty project ID
	emptyTasks, err := store.ListTasks("")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d tasks with empty project ID (should be 0)\n", len(emptyTasks))

	// Test with non-existent project
	nonExistentTasks, err := store.ListTasks("NONEXISTENT")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d tasks in non-existent project (should be 0)\n", len(nonExistentTasks))
}
