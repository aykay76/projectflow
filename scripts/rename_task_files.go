package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Task represents the structure of task JSON files
type Task struct {
	ID        string `json:"id"`
	DisplayID string `json:"display_id"`
	Title     string `json:"title"`
}

// RenameOperation represents a file rename operation
type RenameOperation struct {
	ProjectID   string
	OldFilename string
	NewFilename string
	OldPath     string
	NewPath     string
	Task        Task
}

func main() {
	var (
		dataDir = flag.String("data-dir", "./data", "Path to the data directory")
		dryRun  = flag.Bool("dry-run", false, "Perform a dry run without actually renaming files")
		verbose = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	if *verbose {
		log.Printf("Starting task file rename process")
		log.Printf("Data directory: %s", *dataDir)
		log.Printf("Dry run mode: %t", *dryRun)
	}

	// Scan for rename operations
	operations, err := scanForRenameOperations(*dataDir, *verbose)
	if err != nil {
		log.Fatalf("Failed to scan for rename operations: %v", err)
	}

	if len(operations) == 0 {
		log.Printf("No files need to be renamed")
		return
	}

	log.Printf("Found %d files to rename", len(operations))

	// Check for conflicts
	conflicts := checkForConflicts(operations)
	if len(conflicts) > 0 {
		log.Printf("Found %d conflicts that need to be resolved:", len(conflicts))
		for _, conflict := range conflicts {
			log.Printf("  - Conflict: %s already exists in project %s", conflict.NewFilename, conflict.ProjectID)
		}
		if !*dryRun {
			log.Fatalf("Cannot proceed with conflicts. Please resolve them manually.")
		}
	}

	// Execute rename operations
	if *dryRun {
		log.Printf("DRY RUN - Operations that would be performed:")
		for _, op := range operations {
			log.Printf("  - Project %s: %s -> %s", op.ProjectID, op.OldFilename, op.NewFilename)
		}
	} else {
		err := executeRenameOperations(operations, *verbose)
		if err != nil {
			log.Fatalf("Failed to execute rename operations: %v", err)
		}
		log.Printf("Successfully renamed %d task files", len(operations))
	}
}

// scanForRenameOperations scans all project directories for task files that need renaming
func scanForRenameOperations(dataDir string, verbose bool) ([]RenameOperation, error) {
	var operations []RenameOperation

	projectsDir := filepath.Join(dataDir, "projects")

	// Read all project directories
	projectEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	for _, projectEntry := range projectEntries {
		if !projectEntry.IsDir() {
			continue
		}

		projectID := projectEntry.Name()
		if verbose {
			log.Printf("Scanning project: %s", projectID)
		}

		tasksDir := filepath.Join(projectsDir, projectID, "tasks")

		// Check if tasks directory exists
		if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
			if verbose {
				log.Printf("  - No tasks directory found for project %s", projectID)
			}
			continue
		}

		// Read all task files in this project
		taskEntries, err := os.ReadDir(tasksDir)
		if err != nil {
			if verbose {
				log.Printf("  - Failed to read tasks directory for project %s: %v", projectID, err)
			}
			continue
		}

		for _, taskEntry := range taskEntries {
			if taskEntry.IsDir() || !strings.HasSuffix(taskEntry.Name(), ".json") {
				continue
			}

			filename := taskEntry.Name()
			taskID := strings.TrimSuffix(filename, ".json")

			// Read task file to get display_id
			taskPath := filepath.Join(tasksDir, filename)
			task, err := readTaskFile(taskPath)
			if err != nil {
				if verbose {
					log.Printf("  - Failed to read task file %s: %v", taskPath, err)
				}
				continue
			}

			// Check if this file needs renaming
			if task.DisplayID != "" && task.DisplayID != taskID {
				newFilename := task.DisplayID + ".json"
				newPath := filepath.Join(tasksDir, newFilename)

				operation := RenameOperation{
					ProjectID:   projectID,
					OldFilename: filename,
					NewFilename: newFilename,
					OldPath:     taskPath,
					NewPath:     newPath,
					Task:        task,
				}
				operations = append(operations, operation)

				if verbose {
					log.Printf("  - Found file to rename: %s -> %s", filename, newFilename)
				}
			}
		}
	}

	return operations, nil
}

// readTaskFile reads and parses a task JSON file
func readTaskFile(path string) (Task, error) {
	var task Task

	data, err := os.ReadFile(path)
	if err != nil {
		return task, fmt.Errorf("failed to read file: %w", err)
	}

	err = json.Unmarshal(data, &task)
	if err != nil {
		return task, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return task, nil
}

// checkForConflicts checks if any rename operations would overwrite existing files
func checkForConflicts(operations []RenameOperation) []RenameOperation {
	var conflicts []RenameOperation

	for _, op := range operations {
		// Check if target file already exists
		if _, err := os.Stat(op.NewPath); err == nil {
			conflicts = append(conflicts, op)
		}
	}

	return conflicts
}

// executeRenameOperations performs the actual file rename operations
func executeRenameOperations(operations []RenameOperation, verbose bool) error {
	successCount := 0

	for _, op := range operations {
		if verbose {
			log.Printf("Renaming: %s -> %s", op.OldPath, op.NewPath)
		}

		err := os.Rename(op.OldPath, op.NewPath)
		if err != nil {
			return fmt.Errorf("failed to rename %s to %s: %w", op.OldPath, op.NewPath, err)
		}

		successCount++
	}

	log.Printf("Successfully renamed %d files", successCount)
	return nil
}
