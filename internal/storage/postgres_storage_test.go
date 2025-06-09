package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aykay76/projectflow/internal/models"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupPostgresTestContainer(t *testing.T) (*PostgresStorage, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Create storage instance
	storage, err := NewPostgresStorage(connStr)
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to create PostgreSQL storage: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		if err := storage.Close(); err != nil {
			t.Logf("Failed to close storage: %v", err)
		}
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return storage, cleanup
}

func TestPostgresStorage_CreateTask(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	task := models.NewTask("Test Task", "Test Description")

	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if task.ID == "" {
		t.Error("CreateTask() should set task ID")
	}

	// Verify task exists in database
	if !storage.TaskExists(task.ID) {
		t.Error("CreateTask() should create task in database")
	}
}

func TestPostgresStorage_GetTask(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create a task first
	originalTask := models.NewTask("Test Task", "Test Description")
	originalTask.Status = models.StatusInProgress
	originalTask.Priority = models.PriorityHigh
	originalTask.Type = models.TypeStory

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

	if retrievedTask.Status != originalTask.Status {
		t.Errorf("GetTask() status = %v, want %v", retrievedTask.Status, originalTask.Status)
	}

	if retrievedTask.Priority != originalTask.Priority {
		t.Errorf("GetTask() priority = %v, want %v", retrievedTask.Priority, originalTask.Priority)
	}

	if retrievedTask.Type != originalTask.Type {
		t.Errorf("GetTask() type = %v, want %v", retrievedTask.Type, originalTask.Type)
	}
}

func TestPostgresStorage_GetTask_NotFound(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	_, err := storage.GetTask("nonexistent-id")
	if err == nil {
		t.Error("GetTask() with nonexistent ID should return error")
	}
}

func TestPostgresStorage_UpdateTask(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create a task first
	task := models.NewTask("Original Title", "Original Description")
	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update the task
	task.Title = "Updated Title"
	task.Description = "Updated Description"
	task.Status = models.StatusDone
	task.Priority = models.PriorityCritical
	task.UpdatedAt = time.Now()

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
	if retrievedTask.Status != models.StatusDone {
		t.Errorf("UpdateTask() status = %v, want %v", retrievedTask.Status, models.StatusDone)
	}
	if retrievedTask.Priority != models.PriorityCritical {
		t.Errorf("UpdateTask() priority = %v, want %v", retrievedTask.Priority, models.PriorityCritical)
	}
}

func TestPostgresStorage_UpdateTask_NotFound(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	task := models.NewTask("Test Task", "Test Description")
	task.ID = "nonexistent-id"

	err := storage.UpdateTask(task)
	if err == nil {
		t.Error("UpdateTask() with nonexistent ID should return error")
	}
}

func TestPostgresStorage_DeleteTask(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

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

	// Verify TaskExists returns false
	if storage.TaskExists(task.ID) {
		t.Error("DeleteTask() should make TaskExists return false")
	}
}

func TestPostgresStorage_DeleteTask_WithChildren(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

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

	// Verify parent has child
	updatedParent, err := storage.GetTask(parentTask.ID)
	if err != nil {
		t.Fatalf("Failed to get parent: %v", err)
	}
	if len(updatedParent.Children) != 1 || updatedParent.Children[0] != childTask.ID {
		t.Fatalf("Parent should have child task")
	}

	// Delete parent task (should cascade delete children)
	err = storage.DeleteTask(parentTask.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify both parent and child are gone
	_, err = storage.GetTask(parentTask.ID)
	if err == nil {
		t.Error("Parent task should be deleted")
	}

	_, err = storage.GetTask(childTask.ID)
	if err == nil {
		t.Error("Child task should be deleted when parent is deleted")
	}
}

func TestPostgresStorage_ListTasks(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

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

func TestPostgresStorage_GetTaskChildren(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create parent task
	parentTask := models.NewTask("Parent Task", "Parent Description")
	parentTask.Type = models.TypeEpic
	err := storage.CreateTask(parentTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Create child tasks
	child1 := models.NewTask("Child 1", "Child 1 Description")
	child1.ParentID = parentTask.ID
	child1.Type = models.TypeStory
	err = storage.CreateTask(child1)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	child2 := models.NewTask("Child 2", "Child 2 Description")
	child2.ParentID = parentTask.ID
	child2.Type = models.TypeTask
	err = storage.CreateTask(child2)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Get children
	children, err := storage.GetTaskChildren(parentTask.ID)
	if err != nil {
		t.Fatalf("GetTaskChildren() error = %v", err)
	}

	if len(children) != 2 {
		t.Errorf("GetTaskChildren() returned %d children, want 2", len(children))
	}

	// Verify children are correct
	foundChild1, foundChild2 := false, false
	for _, child := range children {
		if child.ID == child1.ID {
			foundChild1 = true
		}
		if child.ID == child2.ID {
			foundChild2 = true
		}
		if child.ParentID != parentTask.ID {
			t.Errorf("Child %s has wrong parent ID: %s", child.ID, child.ParentID)
		}
	}

	if !foundChild1 {
		t.Error("GetTaskChildren() should include child1")
	}
	if !foundChild2 {
		t.Error("GetTaskChildren() should include child2")
	}
}

func TestPostgresStorage_GetTaskParent(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create parent task
	parentTask := models.NewTask("Parent Task", "Parent Description")
	parentTask.Type = models.TypeEpic
	err := storage.CreateTask(parentTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Create child task
	childTask := models.NewTask("Child Task", "Child Description")
	childTask.ParentID = parentTask.ID
	childTask.Type = models.TypeStory
	err = storage.CreateTask(childTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Get parent
	retrievedParent, err := storage.GetTaskParent(childTask.ID)
	if err != nil {
		t.Fatalf("GetTaskParent() error = %v", err)
	}

	if retrievedParent.ID != parentTask.ID {
		t.Errorf("GetTaskParent() returned ID %s, want %s", retrievedParent.ID, parentTask.ID)
	}
}

func TestPostgresStorage_GetTaskParent_NoParent(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create task without parent
	task := models.NewTask("Orphan Task", "Description")
	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Try to get parent
	_, err = storage.GetTaskParent(task.ID)
	if err == nil {
		t.Error("GetTaskParent() should return error for task without parent")
	}
}

func TestPostgresStorage_GetTaskHierarchy(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

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

func TestPostgresStorage_TaskExists(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create a task
	task := models.NewTask("Test Task", "Test Description")
	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test TaskExists for existing task
	if !storage.TaskExists(task.ID) {
		t.Error("TaskExists() should return true for existing task")
	}

	// Test TaskExists for non-existing task
	if storage.TaskExists("nonexistent-id") {
		t.Error("TaskExists() should return false for non-existing task")
	}
}

func TestPostgresStorage_ParentChildRelationship(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

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

	// Verify parent-child relationship is established
	updatedParent, err := storage.GetTask(parentTask.ID)
	if err != nil {
		t.Fatalf("Failed to get updated parent: %v", err)
	}

	if len(updatedParent.Children) != 1 {
		t.Errorf("Parent should have 1 child, got %d", len(updatedParent.Children))
	}

	if updatedParent.Children[0] != childTask.ID {
		t.Errorf("Parent's child ID = %s, want %s", updatedParent.Children[0], childTask.ID)
	}

	// Verify child has correct parent
	retrievedChild, err := storage.GetTask(childTask.ID)
	if err != nil {
		t.Fatalf("Failed to get child: %v", err)
	}

	if retrievedChild.ParentID != parentTask.ID {
		t.Errorf("Child's parent ID = %s, want %s", retrievedChild.ParentID, parentTask.ID)
	}
}

func TestPostgresStorage_DateFields(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create task with dates
	task := models.NewTask("Test Task", "Test Description")

	// Set some dates
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	startDate := now.Add(-1 * time.Hour)

	task.StartedAt = &startDate
	task.DueDate = &dueDate

	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Retrieve and verify dates
	retrievedTask, err := storage.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrievedTask.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	} else if !retrievedTask.StartedAt.Truncate(time.Second).Equal(startDate.Truncate(time.Second)) {
		t.Errorf("StartedAt = %v, want %v", retrievedTask.StartedAt, startDate)
	}

	if retrievedTask.DueDate == nil {
		t.Error("DueDate should not be nil")
	} else if !retrievedTask.DueDate.Truncate(time.Second).Equal(dueDate.Truncate(time.Second)) {
		t.Errorf("DueDate = %v, want %v", retrievedTask.DueDate, dueDate)
	}
}

func TestPostgresStorage_ConnectionError(t *testing.T) {
	// Test with invalid connection string
	_, err := NewPostgresStorage("invalid connection string")
	if err == nil {
		t.Error("NewPostgresStorage() with invalid connection string should return error")
	}
}

func TestPostgresStorage_TransactionRollback(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Create a task that will be used as parent
	parentTask := models.NewTask("Parent Task", "Parent Description")
	err := storage.CreateTask(parentTask)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Create a task with a non-existent parent to trigger rollback
	childTask := models.NewTask("Child Task", "Child Description")
	childTask.ParentID = "nonexistent-parent-id"

	err = storage.CreateTask(childTask)
	if err == nil {
		t.Error("CreateTask() with invalid parent should return error")
	}

	// Verify the task was not created (transaction rolled back)
	tasks, err := storage.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	// Should only have the parent task, not the failed child
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task after failed create, got %d", len(tasks))
	}
}

func TestPostgresStorage_ConcurrentAccess(t *testing.T) {
	storage, cleanup := setupPostgresTestContainer(t)
	defer cleanup()

	// Test concurrent task creation
	const numGoroutines = 10
	const tasksPerGoroutine = 5

	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(prefix int) {
			for j := 0; j < tasksPerGoroutine; j++ {
				task := models.NewTask(
					fmt.Sprintf("Task %d-%d", prefix, j),
					fmt.Sprintf("Description %d-%d", prefix, j),
				)
				if err := storage.CreateTask(task); err != nil {
					results <- err
					return
				}
			}
			results <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent access error: %v", err)
		}
	}

	// Verify all tasks were created
	tasks, err := storage.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	expectedCount := numGoroutines * tasksPerGoroutine
	if len(tasks) != expectedCount {
		t.Errorf("Expected %d tasks, got %d", expectedCount, len(tasks))
	}
}
