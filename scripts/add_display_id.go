// DEPRECATED: This script was used to migrate tasks to use display_prefix instead of UUIDs.
// The migration has been completed successfully. All tasks now use display_prefix format (e.g., PF-1, PF-2).
// This file is kept for historical reference only.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Project represents the project structure
type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DisplayPrefix string    `json:"display_prefix"`
	Settings      map[string]interface{} `json:"settings"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Task represents the task structure with flexible display_id handling
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Type        string    `json:"type"`
	ParentID    string    `json:"parent_id,omitempty"`
	Children    []string  `json:"children"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DisplayID   string    `json:"display_id,omitempty"` // This is what we're adding (string with prefix)
	ProjectID   string    `json:"project_id,omitempty"`
}

// TaskWithPath holds a task and its file path for processing
type TaskWithPath struct {
	Task Task
	Path string
}

func main() {
	projectID := "b997c175-bab7-48d6-8158-c714fc2d32fa"
	projectFile := "/Users/vanilla/git/aykay76/projectflow/data/projects/" + projectID + ".json"
	projectDir := "/Users/vanilla/git/aykay76/projectflow/data/projects/" + projectID + "/tasks"
	
	// Read the project file to get the display_prefix
	projectData, err := ioutil.ReadFile(projectFile)
	if err != nil {
		log.Fatalf("Failed to read project file: %v", err)
	}
	
	var project Project
	if err := json.Unmarshal(projectData, &project); err != nil {
		log.Fatalf("Failed to unmarshal project: %v", err)
	}
	
	fmt.Printf("Project: %s, Display Prefix: %s\n", project.Name, project.DisplayPrefix)
	
	// Read all task files
	files, err := ioutil.ReadDir(projectDir)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	var tasks []TaskWithPath
	
	// Load all tasks
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		filePath := filepath.Join(projectDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filePath, err)
			continue
		}
		
		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			log.Printf("Failed to unmarshal task from %s: %v", filePath, err)
			continue
		}
		
		tasks = append(tasks, TaskWithPath{
			Task: task,
			Path: filePath,
		})
	}
	
	// Sort tasks by created_at timestamp
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Task.CreatedAt.Before(tasks[j].Task.CreatedAt)
	})
	
	// Add display_id field starting from 1 with project prefix
	for i := range tasks {
		tasks[i].Task.DisplayID = project.DisplayPrefix + "-" + strconv.Itoa(i+1)
		// Update the updated_at timestamp
		tasks[i].Task.UpdatedAt = time.Now()
	}
	
	// Write all tasks back to files
	for _, taskWithPath := range tasks {
		data, err := json.MarshalIndent(taskWithPath.Task, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal task %s: %v", taskWithPath.Task.ID, err)
			continue
		}
		
		if err := ioutil.WriteFile(taskWithPath.Path, data, 0644); err != nil {
			log.Printf("Failed to write file %s: %v", taskWithPath.Path, err)
			continue
		}
		
		fmt.Printf("Updated task %s (display_id: %s) - %s\n", 
			taskWithPath.Task.ID, 
			taskWithPath.Task.DisplayID, 
			taskWithPath.Task.Title)
	}
	
	fmt.Printf("\nSuccessfully updated %d tasks with display_id field\n", len(tasks))
}
