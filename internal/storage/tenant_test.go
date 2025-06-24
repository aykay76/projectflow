package storage

import (
	"context"
	"testing"
	"time"

	"github.com/aykay76/projectflow/internal/models"
)

func TestStorage_TenantCRUD(t *testing.T) {
	// Test with both storage implementations
	storageImpls := []struct {
		name    string
		storage Storage
	}{
		{"FileStorage", createTestFileStorage(t)},
	}

	// Only test PostgreSQL if available
	if postgresStorage := createTestPostgresStorage(t); postgresStorage != nil {
		storageImpls = append(storageImpls, struct {
			name    string
			storage Storage
		}{"PostgresStorage", postgresStorage})
	}

	for _, impl := range storageImpls {
		t.Run(impl.name, func(t *testing.T) {
			testTenantCRUD(t, impl.storage)
		})
	}
}

func testTenantCRUD(t *testing.T, storage Storage) {
	ctx := context.Background()

	// Test CreateTenant
	t.Run("CreateTenant", func(t *testing.T) {
		tenant := models.NewTenant("Test Company", "A test company",
			models.StorageTypeFile, models.AuthProviderLocal)

		err := storage.CreateTenant(ctx, tenant)
		if err != nil {
			t.Fatalf("CreateTenant failed: %v", err)
		}

		// Verify tenant was created
		if tenant.ID == "" {
			t.Error("Expected tenant ID to be set")
		}
	})

	// Test GetTenant
	t.Run("GetTenant", func(t *testing.T) {
		// Create a tenant first
		tenant := models.NewTenant("Get Test Company", "A test company for get",
			models.StorageTypeFile, models.AuthProviderLocal)

		err := storage.CreateTenant(ctx, tenant)
		if err != nil {
			t.Fatalf("CreateTenant failed: %v", err)
		}

		// Get the tenant
		retrieved, err := storage.GetTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenant failed: %v", err)
		}

		// Verify tenant data
		if retrieved.ID != tenant.ID {
			t.Errorf("Expected ID %s, got %s", tenant.ID, retrieved.ID)
		}
		if retrieved.Name != tenant.Name {
			t.Errorf("Expected name %s, got %s", tenant.Name, retrieved.Name)
		}
		if retrieved.Description != tenant.Description {
			t.Errorf("Expected description %s, got %s", tenant.Description, retrieved.Description)
		}
		if retrieved.Status != tenant.Status {
			t.Errorf("Expected status %s, got %s", tenant.Status, retrieved.Status)
		}
	})

	// Test GetTenant with non-existent ID
	t.Run("GetTenant_NotFound", func(t *testing.T) {
		_, err := storage.GetTenant(ctx, "non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent tenant")
		}
	})

	// Test UpdateTenant
	t.Run("UpdateTenant", func(t *testing.T) {
		// Create a tenant first
		tenant := models.NewTenant("Update Test Company", "A test company for update",
			models.StorageTypeFile, models.AuthProviderLocal)

		err := storage.CreateTenant(ctx, tenant)
		if err != nil {
			t.Fatalf("CreateTenant failed: %v", err)
		}

		// Update the tenant
		tenant.Name = "Updated Company Name"
		tenant.Description = "Updated description"
		tenant.Status = models.TenantStatusActive

		err = storage.UpdateTenant(ctx, tenant)
		if err != nil {
			t.Fatalf("UpdateTenant failed: %v", err)
		}

		// Verify the update
		updated, err := storage.GetTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenant failed: %v", err)
		}

		if updated.Name != "Updated Company Name" {
			t.Errorf("Expected name 'Updated Company Name', got %s", updated.Name)
		}
		if updated.Description != "Updated description" {
			t.Errorf("Expected description 'Updated description', got %s", updated.Description)
		}
		if updated.Status != models.TenantStatusActive {
			t.Errorf("Expected status %s, got %s", models.TenantStatusActive, updated.Status)
		}
	})

	// Test UpdateTenant with optimistic locking
	t.Run("UpdateTenant_OptimisticLocking", func(t *testing.T) {
		// Create a tenant first
		tenant := models.NewTenant("Locking Test Company", "A test company for locking",
			models.StorageTypeFile, models.AuthProviderLocal)

		err := storage.CreateTenant(ctx, tenant)
		if err != nil {
			t.Fatalf("CreateTenant failed: %v", err)
		}

		// Get the tenant twice to simulate concurrent access
		tenant1, err := storage.GetTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenant failed: %v", err)
		}

		tenant2, err := storage.GetTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenant failed: %v", err)
		}

		// Update first tenant
		tenant1.Name = "First Update"
		err = storage.UpdateTenant(ctx, tenant1)
		if err != nil {
			t.Fatalf("First UpdateTenant failed: %v", err)
		}

		// Try to update second tenant - should fail due to optimistic locking
		tenant2.Name = "Second Update"
		err = storage.UpdateTenant(ctx, tenant2)
		if err == nil {
			t.Error("Expected optimistic locking error for concurrent update")
		}
	})

	// Test DeleteTenant
	t.Run("DeleteTenant", func(t *testing.T) {
		// Create a tenant first
		tenant := models.NewTenant("Delete Test Company", "A test company for delete",
			models.StorageTypeFile, models.AuthProviderLocal)

		err := storage.CreateTenant(ctx, tenant)
		if err != nil {
			t.Fatalf("CreateTenant failed: %v", err)
		}

		// Delete the tenant
		err = storage.DeleteTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("DeleteTenant failed: %v", err)
		}

		// Verify the tenant is deleted (should not be found)
		_, err = storage.GetTenant(ctx, tenant.ID)
		if err == nil {
			t.Error("Expected error for deleted tenant")
		}

		// Verify TenantExists returns false
		if storage.TenantExists(ctx, tenant.ID) {
			t.Error("TenantExists should return false for deleted tenant")
		}
	})

	// Test DeleteTenant with non-existent ID
	t.Run("DeleteTenant_NotFound", func(t *testing.T) {
		err := storage.DeleteTenant(ctx, "non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent tenant")
		}
	})

	// Test ListTenants
	t.Run("ListTenants", func(t *testing.T) {
		// Create multiple tenants
		tenants := []*models.Tenant{
			models.NewTenant("List Test Company 1", "First company",
				models.StorageTypeFile, models.AuthProviderLocal),
			models.NewTenant("List Test Company 2", "Second company",
				models.StorageTypeFile, models.AuthProviderLocal),
			models.NewTenant("List Test Company 3", "Third company",
				models.StorageTypeFile, models.AuthProviderLocal),
		}

		for _, tenant := range tenants {
			err := storage.CreateTenant(ctx, tenant)
			if err != nil {
				t.Fatalf("CreateTenant failed: %v", err)
			}
		}

		// List all tenants
		result, total, err := storage.ListTenants(ctx, 0, 0)
		if err != nil {
			t.Fatalf("ListTenants failed: %v", err)
		}

		// Verify we got at least our test tenants
		if len(result) < 3 {
			t.Errorf("Expected at least 3 tenants, got %d", len(result))
		}

		if total < 3 {
			t.Errorf("Expected total count at least 3, got %d", total)
		}
	})

	// Test ListTenants with pagination
	t.Run("ListTenants_Pagination", func(t *testing.T) {
		// Create multiple tenants
		tenants := []*models.Tenant{
			models.NewTenant("Pagination Test Company 1", "First company",
				models.StorageTypeFile, models.AuthProviderLocal),
			models.NewTenant("Pagination Test Company 2", "Second company",
				models.StorageTypeFile, models.AuthProviderLocal),
		}

		for _, tenant := range tenants {
			err := storage.CreateTenant(ctx, tenant)
			if err != nil {
				t.Fatalf("CreateTenant failed: %v", err)
			}
		}

		// List with pagination
		result, total, err := storage.ListTenants(ctx, 1, 0)
		if err != nil {
			t.Fatalf("ListTenants failed: %v", err)
		}

		// Verify pagination worked
		if len(result) != 1 {
			t.Errorf("Expected 1 tenant with limit 1, got %d", len(result))
		}

		if total < 2 {
			t.Errorf("Expected total count at least 2, got %d", total)
		}

		// Test offset
		result2, _, err := storage.ListTenants(ctx, 1, 1)
		if err != nil {
			t.Fatalf("ListTenants with offset failed: %v", err)
		}

		if len(result2) != 1 {
			t.Errorf("Expected 1 tenant with limit 1 offset 1, got %d", len(result2))
		}

		// Results should be different
		if len(result) > 0 && len(result2) > 0 && result[0].ID == result2[0].ID {
			t.Error("Expected different results with offset")
		}
	})

	// Test TenantExists
	t.Run("TenantExists", func(t *testing.T) {
		// Create a tenant first
		tenant := models.NewTenant("Exists Test Company", "A test company for exists",
			models.StorageTypeFile, models.AuthProviderLocal)

		err := storage.CreateTenant(ctx, tenant)
		if err != nil {
			t.Fatalf("CreateTenant failed: %v", err)
		}

		// Verify TenantExists returns true
		if !storage.TenantExists(ctx, tenant.ID) {
			t.Error("TenantExists should return true for existing tenant")
		}

		// Verify TenantExists returns false for non-existent tenant
		if storage.TenantExists(ctx, "non-existent-id") {
			t.Error("TenantExists should return false for non-existent tenant")
		}
	})
}

func TestStorage_TenantValidation(t *testing.T) {
	// Test with both storage implementations
	storageImpls := []struct {
		name    string
		storage Storage
	}{
		{"FileStorage", createTestFileStorage(t)},
	}

	// Only test PostgreSQL if available
	if postgresStorage := createTestPostgresStorage(t); postgresStorage != nil {
		storageImpls = append(storageImpls, struct {
			name    string
			storage Storage
		}{"PostgresStorage", postgresStorage})
	}

	for _, impl := range storageImpls {
		t.Run(impl.name, func(t *testing.T) {
			testTenantValidation(t, impl.storage)
		})
	}
}

func testTenantValidation(t *testing.T, storage Storage) {
	ctx := context.Background()

	// Test validation errors
	validationTests := []struct {
		name      string
		tenant    *models.Tenant
		expectErr bool
	}{
		{
			name: "empty name",
			tenant: &models.Tenant{
				Name:   "",
				Status: models.TenantStatusPending,
				Settings: models.TenantSettings{
					StorageType:  models.StorageTypeFile,
					AuthProvider: models.AuthProviderLocal,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			name: "name too long",
			tenant: &models.Tenant{
				Name:   string(make([]byte, 256)), // 256 characters
				Status: models.TenantStatusPending,
				Settings: models.TenantSettings{
					StorageType:  models.StorageTypeFile,
					AuthProvider: models.AuthProviderLocal,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			name: "invalid storage type",
			tenant: &models.Tenant{
				Name:   "Valid Company",
				Status: models.TenantStatusPending,
				Settings: models.TenantSettings{
					StorageType:  models.StorageType("invalid"),
					AuthProvider: models.AuthProviderLocal,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			name: "invalid auth provider",
			tenant: &models.Tenant{
				Name:   "Valid Company",
				Status: models.TenantStatusPending,
				Settings: models.TenantSettings{
					StorageType:  models.StorageTypeFile,
					AuthProvider: models.AuthProvider("invalid"),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			name: "valid tenant",
			tenant: &models.Tenant{
				Name:   "Valid Company",
				Status: models.TenantStatusPending,
				Settings: models.TenantSettings{
					StorageType:  models.StorageTypeFile,
					AuthProvider: models.AuthProviderLocal,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: false,
		},
	}

	for _, test := range validationTests {
		t.Run(test.name, func(t *testing.T) {
			err := storage.CreateTenant(ctx, test.tenant)
			if test.expectErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !test.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestStorage_TenantDuplicateNames(t *testing.T) {
	// Test with both storage implementations
	storageImpls := []struct {
		name    string
		storage Storage
	}{
		{"FileStorage", createTestFileStorage(t)},
	}

	// Only test PostgreSQL if available
	if postgresStorage := createTestPostgresStorage(t); postgresStorage != nil {
		storageImpls = append(storageImpls, struct {
			name    string
			storage Storage
		}{"PostgresStorage", postgresStorage})
	}

	for _, impl := range storageImpls {
		t.Run(impl.name, func(t *testing.T) {
			testTenantDuplicateNames(t, impl.storage)
		})
	}
}

func testTenantDuplicateNames(t *testing.T, storage Storage) {
	ctx := context.Background()

	// Create first tenant
	tenant1 := models.NewTenant("Duplicate Test Company", "First company",
		models.StorageTypeFile, models.AuthProviderLocal)

	err := storage.CreateTenant(ctx, tenant1)
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Try to create second tenant with same name
	tenant2 := models.NewTenant("Duplicate Test Company", "Second company",
		models.StorageTypeFile, models.AuthProviderLocal)

	err = storage.CreateTenant(ctx, tenant2)
	if err == nil {
		t.Error("Expected error for duplicate tenant name")
	}
}

// Helper functions to create test storage instances
func createTestFileStorage(t *testing.T) Storage {
	tempDir := t.TempDir()
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test file storage: %v", err)
	}
	return storage
}

func createTestPostgresStorage(t *testing.T) Storage {
	// Skip if PostgreSQL is not available
	connectionString := getTestPostgresConnectionString()
	if connectionString == "" {
		return nil // Return nil instead of skipping
	}

	storage, err := NewPostgresStorage(connectionString)
	if err != nil {
		return nil // Return nil instead of skipping
	}
	return storage
}

func getTestPostgresConnectionString() string {
	// Return empty string if no test database is configured
	// In a real test environment, this would return a test database connection string
	return ""
}
