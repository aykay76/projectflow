package storage

import (
	"context"
	"testing"

	"github.com/aykay76/projectflow/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTenantContextHelpers tests the tenant context helper functions
func TestTenantContextHelpers(t *testing.T) {
	t.Run("WithTenant and GetTenantID", func(t *testing.T) {
		ctx := WithTenant(context.Background(), "test-tenant")
		tenantID := GetTenantID(ctx)
		assert.Equal(t, "test-tenant", tenantID)
	})

	t.Run("GetTenantID with default context", func(t *testing.T) {
		ctx := context.Background()
		tenantID := GetTenantID(ctx)
		assert.Equal(t, "default", tenantID)
	})

	t.Run("HasTenant with tenant context", func(t *testing.T) {
		ctx := WithTenant(context.Background(), "test-tenant")
		assert.True(t, HasTenant(ctx))
	})

	t.Run("HasTenant with default context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, HasTenant(ctx))
	})

	t.Run("Nested tenant contexts", func(t *testing.T) {
		ctx1 := WithTenant(context.Background(), "tenant1")
		ctx2 := WithTenant(ctx1, "tenant2")

		// The most recent tenant should be used
		tenantID := GetTenantID(ctx2)
		assert.Equal(t, "tenant2", tenantID)
	})

	t.Run("empty tenant ID falls back to default", func(t *testing.T) {
		ctx := WithTenant(context.Background(), "")
		tenantID := GetTenantID(ctx)
		assert.Equal(t, "default", tenantID)
	})

	t.Run("whitespace-only tenant ID falls back to default", func(t *testing.T) {
		ctx := WithTenant(context.Background(), "   ")
		tenantID := GetTenantID(ctx)
		assert.Equal(t, "default", tenantID)
	})

	t.Run("special characters in tenant ID", func(t *testing.T) {
		specialTenant := "tenant-123_ABC.test"
		ctx := WithTenant(context.Background(), specialTenant)
		tenantID := GetTenantID(ctx)
		assert.Equal(t, specialTenant, tenantID)
	})

	t.Run("case sensitivity in tenant IDs", func(t *testing.T) {
		ctx1 := WithTenant(context.Background(), "TestTenant")
		ctx2 := WithTenant(context.Background(), "testtenant")

		// Both should be preserved as-is
		tenantID1 := GetTenantID(ctx1)
		tenantID2 := GetTenantID(ctx2)
		assert.Equal(t, "TestTenant", tenantID1)
		assert.Equal(t, "testtenant", tenantID2)
		assert.NotEqual(t, tenantID1, tenantID2)
	})
}

// TestBackwardCompatibility tests that operations with context work correctly
// Note: This test validates that the context-aware interface works, but tenant isolation
// is not yet implemented in the current storage implementation.
func TestBackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()

	// Operations with default context should work as before
	defaultCtx := context.Background()

	// Create a task with default context
	task := models.NewTask("Default Task", "Task for default tenant")
	err = storage.CreateTask(defaultCtx, task)
	require.NoError(t, err)

	// All operations should work with default context
	t.Run("all operations work with default context", func(t *testing.T) {
		// Get task
		retrievedTask, err := storage.GetTask(defaultCtx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.Title, retrievedTask.Title)

		// Task exists
		exists := storage.TaskExists(defaultCtx, task.ID)
		assert.True(t, exists)

		// Get by display ID
		retrievedTask, err = storage.GetTaskByDisplayID(defaultCtx, task.DisplayID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, retrievedTask.ID)

		// Update task
		task.Title = "Updated Default Task"
		err = storage.UpdateTask(defaultCtx, task)
		assert.NoError(t, err)

		// List tasks should include this task
		project, err := storage.GetProject(defaultCtx, task.ProjectID)
		require.NoError(t, err)

		tasks, err := storage.ListTasks(defaultCtx, project.ID)
		require.NoError(t, err)

		var foundTask bool
		for _, t := range tasks {
			if t.ID == task.ID {
				foundTask = true
				break
			}
		}
		assert.True(t, foundTask)
	})

	t.Run("tenant context operations work with current implementation", func(t *testing.T) {
		// Create a tenant context
		tenantCtx := WithTenant(context.Background(), "test-tenant")

		// Create a task with tenant context
		tenantTask := models.NewTask("Tenant Task", "Task created with tenant context")
		err := storage.CreateTask(tenantCtx, tenantTask)
		require.NoError(t, err)

		// Should be able to retrieve it (since isolation is not yet implemented)
		retrievedTask, err := storage.GetTask(tenantCtx, tenantTask.ID)
		require.NoError(t, err)
		assert.Equal(t, tenantTask.Title, retrievedTask.Title)

		// Should also be retrievable from default context (since isolation is not yet implemented)
		retrievedTask, err = storage.GetTask(defaultCtx, tenantTask.ID)
		require.NoError(t, err)
		assert.Equal(t, tenantTask.Title, retrievedTask.Title)
	})
}

// TestContextPropagation validates that context is properly passed through all storage operations
func TestContextPropagation(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()

	// Test that all methods accept context without error
	ctx := WithTenant(context.Background(), "test-tenant")

	// Test project operations
	project := models.NewProject("Test Project", "Test Description", "TP")
	err = storage.CreateProject(ctx, project)
	require.NoError(t, err)

	_, err = storage.GetProject(ctx, project.ID)
	require.NoError(t, err)

	err = storage.UpdateProject(ctx, project)
	require.NoError(t, err)

	_, err = storage.ListProjects(ctx)
	require.NoError(t, err)

	_, err = storage.GetProjectByName(ctx, project.Name)
	require.NoError(t, err)

	_, err = storage.GetProjectByDisplayPrefix(ctx, project.DisplayPrefix)
	require.NoError(t, err)

	exists := storage.ProjectExists(ctx, project.ID)
	assert.True(t, exists)

	// Test task operations
	task := models.NewTask("Test Task", "Test Description")
	err = storage.CreateTask(ctx, task)
	require.NoError(t, err)

	_, err = storage.GetTask(ctx, task.ID)
	require.NoError(t, err)

	_, err = storage.GetTaskByDisplayID(ctx, task.DisplayID)
	require.NoError(t, err)

	err = storage.UpdateTask(ctx, task)
	require.NoError(t, err)

	_, err = storage.ListTasks(ctx, project.ID)
	require.NoError(t, err)

	_, err = storage.GetTaskChildren(ctx, task.ID)
	require.NoError(t, err)

	_, err = storage.GetTaskHierarchy(ctx)
	require.NoError(t, err)

	_, err = storage.GetNextDisplayID(ctx, project.ID)
	require.NoError(t, err)

	exists = storage.TaskExists(ctx, task.ID)
	assert.True(t, exists)

	// Test deletion operations
	err = storage.DeleteTask(ctx, task.ID)
	require.NoError(t, err)

	err = storage.DeleteProject(ctx, project.ID)
	require.NoError(t, err)
}

// TestTenantIsolationPlaceholder - This test documents the expected behavior for tenant isolation
// TODO: This test should be updated once tenant isolation is fully implemented
func TestTenantIsolationPlaceholder(t *testing.T) {
	t.Skip("Tenant isolation is not yet implemented. This test documents expected behavior.")

	// When tenant isolation is implemented, this is the expected behavior:
	// 1. Each tenant should have isolated data storage
	// 2. Tasks and projects created in one tenant context should not be visible in another
	// 3. The default tenant should be isolated from named tenants
	// 4. All CRUD operations should respect tenant boundaries
	// 5. Hierarchy operations should only show tasks within the same tenant
	// 6. Display ID generation should be per-tenant
}

// TestErrorHandling tests error handling in context-aware operations
func TestErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	require.NoError(t, err)
	defer storage.Close()

	ctx := WithTenant(context.Background(), "test-tenant")

	// Test getting non-existent task
	_, err = storage.GetTask(ctx, "non-existent-id")
	assert.Error(t, err)

	// Test getting task by non-existent display ID
	_, err = storage.GetTaskByDisplayID(ctx, "NON-EXISTENT-1")
	assert.Error(t, err)

	// Test getting non-existent project
	_, err = storage.GetProject(ctx, "non-existent-project-id")
	assert.Error(t, err)

	// Test getting project by non-existent name
	_, err = storage.GetProjectByName(ctx, "Non-existent Project")
	assert.Error(t, err)

	// Test updating non-existent task
	task := models.NewTask("Test Task", "Test Description")
	task.ID = "non-existent-id"
	err = storage.UpdateTask(ctx, task)
	assert.Error(t, err)

	// Test deleting non-existent task
	err = storage.DeleteTask(ctx, "non-existent-id")
	assert.Error(t, err)
}
