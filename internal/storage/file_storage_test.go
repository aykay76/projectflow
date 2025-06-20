package storage

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/aykay76/projectflow/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_CreateTask(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	task := models.NewTask("Test Task", "Test Description")

	err = storage.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if task.ID == "" {
		t.Error("CreateTask() should set task ID")
	}

	// Verify file was created in the correct project directory
	// Since no project was specified, task should be assigned to default project
	if task.ProjectID == "" {
		t.Error("CreateTask() should assign task to default project")
	}

	// Get the project to find its display prefix
	project, err := storage.GetProject(task.ProjectID)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	// Task should be saved with display ID as filename if available
	filename := task.ID + ".json"
	if task.DisplayID != "" {
		filename = task.DisplayID + ".json"
	}
	filePath := filepath.Join(tempDir, "projects", project.DisplayPrefix, "tasks", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("CreateTask() should create file on disk")
	}
}

func TestFileStorage_GetTask(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a task first
	originalTask := models.NewTask("Test Task", "Test Description")
	err = storage.CreateTask(originalTask)
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
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	_, err = storage.GetTask("nonexistent-id")
	if err == nil {
		t.Error("GetTask() with nonexistent ID should return error")
	}
}

func TestFileStorage_GetTaskByDisplayID(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()

	// Create a task (this should auto-generate a display ID)
	task := models.NewTask("Test Display ID Task", "Testing display ID lookup")
	err = storage.CreateTask(task)
	require.NoError(t, err)
	require.NotEmpty(t, task.DisplayID, "Task should have a display ID")

	// Test retrieving task by display ID
	retrievedTask, err := storage.GetTaskByDisplayID(task.DisplayID)
	require.NoError(t, err)

	assert.Equal(t, task.ID, retrievedTask.ID)
	assert.Equal(t, task.DisplayID, retrievedTask.DisplayID)
	assert.Equal(t, task.Title, retrievedTask.Title)
	assert.Equal(t, task.Description, retrievedTask.Description)
	assert.Equal(t, task.ProjectID, retrievedTask.ProjectID)
}

func TestFileStorage_GetTaskByDisplayID_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()

	// Test with non-existent display ID
	_, err = storage.GetTaskByDisplayID("NONEXISTENT-999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found with display ID")
}

func TestFileStorage_GetTaskByDisplayID_CaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()

	// Create a task (this should auto-generate a display ID like "PF-1")
	task := models.NewTask("Test Case Insensitive Task", "Testing case insensitive lookup")
	err = storage.CreateTask(task)
	require.NoError(t, err)
	require.NotEmpty(t, task.DisplayID, "Task should have a display ID")

	// Test retrieving task by display ID with different cases
	testCases := []string{
		task.DisplayID,                                 // exact case
		strings.ToLower(task.DisplayID),                // all lowercase
		strings.ToUpper(task.DisplayID),                // all uppercase
		strings.Title(strings.ToLower(task.DisplayID)), // mixed case
	}

	for _, displayID := range testCases {
		retrievedTask, err := storage.GetTaskByDisplayID(displayID)
		require.NoError(t, err, "Should find task with display ID: %s", displayID)
		assert.Equal(t, task.ID, retrievedTask.ID, "Should return same task for display ID: %s", displayID)
		assert.Equal(t, task.DisplayID, retrievedTask.DisplayID, "Original display ID should be preserved")
	}
}

func TestFileStorage_UpdateTask(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a task first
	task := models.NewTask("Original Title", "Original Description")
	err = storage.CreateTask(task)
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
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a task first
	task := models.NewTask("Test Task", "Test Description")
	err = storage.CreateTask(task)
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
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create multiple tasks
	task1 := models.NewTask("Task 1", "Description 1")
	task2 := models.NewTask("Task 2", "Description 2")

	err = storage.CreateTask(task1)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	err = storage.CreateTask(task2)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// List tasks from the default project
	// Tasks created without project ID are assigned to the default project
	defaultProject, err := storage.getOrCreateDefaultProjectUnsafe()
	if err != nil {
		t.Fatalf("Failed to get default project: %v", err)
	}

	tasks, err := storage.ListTasks(defaultProject.ID)
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
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create parent task
	parentTask := models.NewTask("Parent Task", "Parent Description")
	parentTask.Type = models.TypeEpic
	err = storage.CreateTask(parentTask)
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

// Project tests

func TestFileStorage_CreateProject(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	project := models.NewProject("Test Project", "Test Description", "TEST")

	err = storage.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if project.ID == "" {
		t.Error("CreateProject() should set project ID")
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "projects", project.DisplayPrefix+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("CreateProject() should create file on disk")
	}
}

func TestFileStorage_GetProject(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a project first
	originalProject := models.NewProject("Test Project", "Test Description", "TEST")
	err = storage.CreateProject(originalProject)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test getting the project
	retrievedProject, err := storage.GetProject(originalProject.ID)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}

	if retrievedProject.Name != originalProject.Name {
		t.Errorf("GetProject() name = %v, want %v", retrievedProject.Name, originalProject.Name)
	}

	if retrievedProject.Description != originalProject.Description {
		t.Errorf("GetProject() description = %v, want %v", retrievedProject.Description, originalProject.Description)
	}

	if retrievedProject.DisplayPrefix != originalProject.DisplayPrefix {
		t.Errorf("GetProject() DisplayPrefix = %v, want %v", retrievedProject.DisplayPrefix, originalProject.DisplayPrefix)
	}
}

func TestFileStorage_GetProject_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	_, err = storage.GetProject("nonexistent-id")
	if err == nil {
		t.Error("GetProject() with nonexistent ID should return error")
	}
}

func TestFileStorage_UpdateProject(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a project first
	project := models.NewProject("Original Name", "Original Description", "ORIG")
	err = storage.CreateProject(project)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update the project
	project.Name = "Updated Name"
	project.Description = "Updated Description"
	project.SetSetting("testKey", "testValue")
	err = storage.UpdateProject(project)
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	// Verify the update
	retrievedProject, err := storage.GetProject(project.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated project: %v", err)
	}

	if retrievedProject.Name != "Updated Name" {
		t.Errorf("UpdateProject() name = %v, want %v", retrievedProject.Name, "Updated Name")
	}

	if retrievedProject.Description != "Updated Description" {
		t.Errorf("UpdateProject() description = %v, want %v", retrievedProject.Description, "Updated Description")
	}

	value, exists := retrievedProject.GetSetting("testKey")
	if !exists || value != "testValue" {
		t.Errorf("UpdateProject() setting = %v (exists: %v), want testValue (true)", value, exists)
	}
}

func TestFileStorage_DeleteProject(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a project first
	project := models.NewProject("Test Project", "Test Description", "TEST")
	err = storage.CreateProject(project)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete the project
	err = storage.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// Verify it's gone
	_, err = storage.GetProject(project.ID)
	if err == nil {
		t.Error("DeleteProject() should remove project")
	}

	// Verify file was deleted
	filePath := filepath.Join(tempDir, "projects", project.DisplayPrefix+".json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("DeleteProject() should remove file from disk")
	}
}

func TestFileStorage_ListProjects(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create multiple projects
	project1 := models.NewProject("Project 1", "Description 1", "P1")
	project2 := models.NewProject("Project 2", "Description 2", "P2")

	err = storage.CreateProject(project1)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	err = storage.CreateProject(project2)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// List projects
	projects, err := storage.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("ListProjects() returned %d projects, want 2", len(projects))
	}

	// Check that both projects are present
	found1, found2 := false, false
	for _, project := range projects {
		if project.ID == project1.ID {
			found1 = true
		}
		if project.ID == project2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("ListProjects() should include project1")
	}
	if !found2 {
		t.Error("ListProjects() should include project2")
	}
}

func TestFileStorage_GetProjectByName(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a project
	originalProject := models.NewProject("Unique Project Name", "Test Description", "UPN")
	err = storage.CreateProject(originalProject)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test getting by name
	retrievedProject, err := storage.GetProjectByName("Unique Project Name")
	if err != nil {
		t.Fatalf("GetProjectByName() error = %v", err)
	}

	if retrievedProject.ID != originalProject.ID {
		t.Errorf("GetProjectByName() ID = %v, want %v", retrievedProject.ID, originalProject.ID)
	}

	// Test non-existent name
	_, err = storage.GetProjectByName("Non-existent Project")
	if err == nil {
		t.Error("GetProjectByName() with non-existent name should return error")
	}
}

func TestFileStorage_ProjectExists(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)

	// Create a test project
	project := models.NewProject("Test Project", "A test project", "TEST")
	err = storage.CreateProject(project)
	require.NoError(t, err)

	// Test exists
	exists := storage.ProjectExists(project.ID)
	assert.True(t, exists)

	// Test non-existent project
	exists = storage.ProjectExists("non-existent")
	assert.False(t, exists)
}

func TestFileStorage_GetNextDisplayID(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)

	// Create a test project
	project := models.NewProject("Test Project", "A test project", "TEST")
	err = storage.CreateProject(project)
	require.NoError(t, err)

	// Test getting sequential display IDs
	displayID1, err := storage.GetNextDisplayID(project.ID)
	require.NoError(t, err)
	assert.Equal(t, "TEST-1", displayID1)

	displayID2, err := storage.GetNextDisplayID(project.ID)
	require.NoError(t, err)
	assert.Equal(t, "TEST-2", displayID2)

	displayID3, err := storage.GetNextDisplayID(project.ID)
	require.NoError(t, err)
	assert.Equal(t, "TEST-3", displayID3)

	// Test with non-existent project
	_, err = storage.GetNextDisplayID("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestFileStorage_GetNextDisplayID_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)

	// Create a test project
	project := models.NewProject("Test Project", "A test project", "CONC")
	err = storage.CreateProject(project)
	require.NoError(t, err)

	// Test concurrent access to display ID generation
	const numGoroutines = 10
	displayIDs := make([]string, numGoroutines)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			displayID, err := storage.GetNextDisplayID(project.ID)
			require.NoError(t, err)

			mu.Lock()
			displayIDs[index] = displayID
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all display IDs are unique and in expected format
	seen := make(map[string]bool)
	for _, displayID := range displayIDs {
		assert.False(t, seen[displayID], "Duplicate display ID: %s", displayID)
		seen[displayID] = true
		assert.True(t, strings.HasPrefix(displayID, "CONC-"), "Invalid display ID format: %s", displayID)
	}

	// Should have generated exactly numGoroutines unique IDs
	assert.Len(t, seen, numGoroutines)
}
