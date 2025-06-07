package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aykay76/projectflow/internal/models"
)

func TestFileStorage_CreateTask(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	storage := NewFileStorage(tempDir)

	task := models.NewTask("Test Task", "Test Description")

	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if task.ID == "" {
		t.Error("CreateTask() should set task ID")
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "tasks", task.ID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("CreateTask() should create file on disk")
	}
}

func TestFileStorage_GetTask(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create a task first
	originalTask := models.NewTask("Test Task", "Test Description")
	err := storage.CreateTask(originalTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test getting the task
	retrievedTask, err := storage.GetTask(originalTask.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrievedTask.Title != originalTask.Title {
		t.Errorf("GetTask() title = %v, want %v", retrievedTask.Title, originalTask.Title)
	}

	if retrievedTask.Description != originalTask.Description {
		t.Errorf("GetTask() description = %v, want %v", retrievedTask.Description, originalTask.Description)
	}
}

func TestFileStorage_GetTask_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	_, err := storage.GetTask("nonexistent-id")
	if err == nil {
		t.Error("GetTask() with nonexistent ID should return error")
	}
}

func TestFileStorage_UpdateTask(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create a task first
	task := models.NewTask("Original Title", "Original Description")
	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update the task
	task.Title = "Updated Title"
	task.Description = "Updated Description"
	err = storage.UpdateTask(task)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	// Verify the update
	retrievedTask, err := storage.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated task: %v", err)
	}

	if retrievedTask.Title != "Updated Title" {
		t.Errorf("UpdateTask() title = %v, want %v", retrievedTask.Title, "Updated Title")
	}

	if retrievedTask.Description != "Updated Description" {
		t.Errorf("UpdateTask() description = %v, want %v", retrievedTask.Description, "Updated Description")
	}
}

func TestFileStorage_DeleteTask(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create a task first
	task := models.NewTask("Test Task", "Test Description")
	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete the task
	err = storage.DeleteTask(task.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify the task is gone
	_, err = storage.GetTask(task.ID)
	if err == nil {
		t.Error("DeleteTask() should make task no longer retrievable")
	}

	// Verify file was deleted
	filePath := filepath.Join(tempDir, "tasks", task.ID+".json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("DeleteTask() should remove file from disk")
	}
}

func TestFileStorage_ListTasks(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create multiple tasks
	task1 := models.NewTask("Task 1", "Description 1")
	task2 := models.NewTask("Task 2", "Description 2")

	err := storage.CreateTask(task1)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	err = storage.CreateTask(task2)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// List tasks
	tasks, err := storage.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("ListTasks() returned %d tasks, want 2", len(tasks))
	}

	// Check that both tasks are present
	found1, found2 := false, false
	for _, task := range tasks {
		if task.ID == task1.ID {
			found1 = true
		}
		if task.ID == task2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("ListTasks() should include task1")
	}
	if !found2 {
		t.Error("ListTasks() should include task2")
	}
}

func TestFileStorage_GetTaskHierarchy(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create parent task
	parentTask := models.NewTask("Parent Task", "Parent Description")
	parentTask.Type = models.TypeEpic
	err := storage.CreateTask(parentTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Create child task
	childTask := models.NewTask("Child Task", "Child Description")
	childTask.Type = models.TypeStory
	childTask.ParentID = parentTask.ID
	err = storage.CreateTask(childTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update parent to include child
	parentTask.AddChild(childTask.ID)
	err = storage.UpdateTask(parentTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test hierarchy
	hierarchy, err := storage.GetTaskHierarchy()
	if err != nil {
		t.Fatalf("GetTaskHierarchy() error = %v", err)
	}

	// Should have one top-level item (the epic)
	if len(hierarchy) != 1 {
		t.Errorf("GetTaskHierarchy() returned %d top-level items, want 1", len(hierarchy))
		return
	}

	topLevel := hierarchy[0]
	if topLevel.ID != parentTask.ID {
		t.Errorf("GetTaskHierarchy() top-level task ID = %v, want %v", topLevel.ID, parentTask.ID)
	}

	if len(topLevel.ChildTasks) != 1 {
		t.Errorf("GetTaskHierarchy() parent has %d children, want 1", len(topLevel.ChildTasks))
		return
	}

	child := topLevel.ChildTasks[0]
	if child.ID != childTask.ID {
		t.Errorf("GetTaskHierarchy() child task ID = %v, want %v", child.ID, childTask.ID)
	}
}
