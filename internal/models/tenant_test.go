package models

import (
	"testing"
	"time"
)

func TestNewTenant(t *testing.T) {
	name := "Test Company"
	description := "A test company tenant"
	storageType := StorageTypePostgres
	authProvider := AuthProviderOAuth2

	tenant := NewTenant(name, description, storageType, authProvider)

	if tenant.Name != name {
		t.Errorf("Expected name %s, got %s", name, tenant.Name)
	}
	if tenant.Description != description {
		t.Errorf("Expected description %s, got %s", description, tenant.Description)
	}
	if tenant.Status != TenantStatusPending {
		t.Errorf("Expected status %s, got %s", TenantStatusPending, tenant.Status)
	}
	if tenant.Settings.StorageType != storageType {
		t.Errorf("Expected storage type %s, got %s", storageType, tenant.Settings.StorageType)
	}
	if tenant.Settings.AuthProvider != authProvider {
		t.Errorf("Expected auth provider %s, got %s", authProvider, tenant.Settings.AuthProvider)
	}
	if tenant.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if tenant.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestTenant_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tenant  *Tenant
		wantErr bool
		errType error
	}{
		{
			name: "valid tenant",
			tenant: &Tenant{
				Name:   "Valid Company",
				Status: TenantStatusActive,
				Settings: TenantSettings{
					StorageType:  StorageTypePostgres,
					AuthProvider: AuthProviderOAuth2,
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			tenant: &Tenant{
				Name:   "",
				Status: TenantStatusActive,
				Settings: TenantSettings{
					StorageType:  StorageTypePostgres,
					AuthProvider: AuthProviderOAuth2,
				},
			},
			wantErr: true,
			errType: ErrTenantNameRequired,
		},
		{
			name: "name too long",
			tenant: &Tenant{
				Name:   string(make([]byte, 256)), // 256 characters
				Status: TenantStatusActive,
				Settings: TenantSettings{
					StorageType:  StorageTypePostgres,
					AuthProvider: AuthProviderOAuth2,
				},
			},
			wantErr: true,
			errType: ErrTenantNameTooLong,
		},
		{
			name: "empty storage type",
			tenant: &Tenant{
				Name:   "Valid Company",
				Status: TenantStatusActive,
				Settings: TenantSettings{
					StorageType:  "",
					AuthProvider: AuthProviderOAuth2,
				},
			},
			wantErr: true,
			errType: ErrTenantStorageTypeEmpty,
		},
		{
			name: "invalid auth provider",
			tenant: &Tenant{
				Name:   "Valid Company",
				Status: TenantStatusActive,
				Settings: TenantSettings{
					StorageType:  StorageTypePostgres,
					AuthProvider: "invalid",
				},
			},
			wantErr: true,
			errType: ErrTenantAuthProviderInvalid,
		},
		{
			name: "invalid custom domain",
			tenant: &Tenant{
				Name:   "Valid Company",
				Status: TenantStatusActive,
				Settings: TenantSettings{
					StorageType:  StorageTypePostgres,
					AuthProvider: AuthProviderOAuth2,
					CustomDomain: "invalid..domain",
				},
			},
			wantErr: true,
			errType: ErrTenantDomainInvalid,
		},
		{
			name: "valid custom domain",
			tenant: &Tenant{
				Name:   "Valid Company",
				Status: TenantStatusActive,
				Settings: TenantSettings{
					StorageType:  StorageTypePostgres,
					AuthProvider: AuthProviderOAuth2,
					CustomDomain: "projects.company.com",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tenant.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Tenant.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errType != nil && err != tt.errType {
				t.Errorf("Tenant.Validate() error = %v, expected %v", err, tt.errType)
			}
		})
	}
}

func TestTenant_StatusMethods(t *testing.T) {
	tenant := NewTenant("Test Company", "Description", StorageTypePostgres, AuthProviderLocal)

	// Test initial status
	if tenant.Status != TenantStatusPending {
		t.Errorf("Expected initial status %s, got %s", TenantStatusPending, tenant.Status)
	}

	// Test Activate
	originalTime := tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	tenant.Activate()
	if tenant.Status != TenantStatusActive {
		t.Errorf("Expected status %s after Activate(), got %s", TenantStatusActive, tenant.Status)
	}
	if !tenant.IsActive() {
		t.Error("Expected IsActive() to return true")
	}
	if tenant.UpdatedAt.Equal(originalTime) {
		t.Error("Expected UpdatedAt to be updated after Activate()")
	}

	// Test Suspend
	originalTime = tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	tenant.Suspend()
	if tenant.Status != TenantStatusSuspended {
		t.Errorf("Expected status %s after Suspend(), got %s", TenantStatusSuspended, tenant.Status)
	}
	if !tenant.IsSuspended() {
		t.Error("Expected IsSuspended() to return true")
	}
	if tenant.UpdatedAt.Equal(originalTime) {
		t.Error("Expected UpdatedAt to be updated after Suspend()")
	}

	// Test Delete
	originalTime = tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	tenant.Delete()
	if tenant.Status != TenantStatusDeleted {
		t.Errorf("Expected status %s after Delete(), got %s", TenantStatusDeleted, tenant.Status)
	}
	if !tenant.IsDeleted() {
		t.Error("Expected IsDeleted() to return true")
	}
	if tenant.UpdatedAt.Equal(originalTime) {
		t.Error("Expected UpdatedAt to be updated after Delete()")
	}
}

func TestTenant_FeatureMethods(t *testing.T) {
	tenant := NewTenant("Test Company", "Description", StorageTypePostgres, AuthProviderLocal)

	// Test initial features
	if len(tenant.Settings.Features) != 0 {
		t.Error("Expected no initial features")
	}
	if tenant.HasFeature("test-feature") {
		t.Error("Expected HasFeature to return false for non-existent feature")
	}

	// Test SetFeature
	originalTime := tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	tenant.SetFeature("test-feature")
	if !tenant.HasFeature("test-feature") {
		t.Error("Expected HasFeature to return true after SetFeature")
	}
	if len(tenant.Settings.Features) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(tenant.Settings.Features))
	}
	if tenant.UpdatedAt.Equal(originalTime) {
		t.Error("Expected UpdatedAt to be updated after SetFeature()")
	}

	// Test SetFeature duplicate
	tenant.SetFeature("test-feature") // Should not add duplicate
	if len(tenant.Settings.Features) != 1 {
		t.Errorf("Expected 1 feature after duplicate SetFeature, got %d", len(tenant.Settings.Features))
	}

	// Test RemoveFeature
	originalTime = tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	tenant.RemoveFeature("test-feature")
	if tenant.HasFeature("test-feature") {
		t.Error("Expected HasFeature to return false after RemoveFeature")
	}
	if len(tenant.Settings.Features) != 0 {
		t.Errorf("Expected 0 features after RemoveFeature, got %d", len(tenant.Settings.Features))
	}
	if tenant.UpdatedAt.Equal(originalTime) {
		t.Error("Expected UpdatedAt to be updated after RemoveFeature()")
	}
}

func TestTenant_LimitMethods(t *testing.T) {
	tenant := NewTenant("Test Company", "Description", StorageTypePostgres, AuthProviderLocal)

	// Test initial limits
	if len(tenant.Settings.Limits) != 0 {
		t.Error("Expected no initial limits")
	}
	if _, exists := tenant.GetLimit("max-projects"); exists {
		t.Error("Expected GetLimit to return false for non-existent limit")
	}

	// Test SetLimit
	originalTime := tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	tenant.SetLimit("max-projects", 10)
	if limit, exists := tenant.GetLimit("max-projects"); !exists || limit != 10 {
		t.Errorf("Expected limit 10, got %d (exists: %v)", limit, exists)
	}
	if tenant.UpdatedAt.Equal(originalTime) {
		t.Error("Expected UpdatedAt to be updated after SetLimit()")
	}
}

func TestTenant_PreferenceMethods(t *testing.T) {
	tenant := NewTenant("Test Company", "Description", StorageTypePostgres, AuthProviderLocal)

	// Test initial preferences
	if len(tenant.Settings.Preferences) != 0 {
		t.Error("Expected no initial preferences")
	}
	if _, exists := tenant.GetPreference("theme"); exists {
		t.Error("Expected GetPreference to return false for non-existent preference")
	}

	// Test SetPreference
	originalTime := tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	tenant.SetPreference("theme", "dark")
	if pref, exists := tenant.GetPreference("theme"); !exists || pref != "dark" {
		t.Errorf("Expected preference 'dark', got %v (exists: %v)", pref, exists)
	}
	if tenant.UpdatedAt.Equal(originalTime) {
		t.Error("Expected UpdatedAt to be updated after SetPreference()")
	}
}

func TestIsValidTenantStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"active", true},
		{"suspended", true},
		{"deleted", true},
		{"pending", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsValidTenantStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidTenantStatus(%s) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestIsValidStorageType(t *testing.T) {
	tests := []struct {
		storageType string
		valid       bool
	}{
		{"file", true},
		{"postgres", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.storageType, func(t *testing.T) {
			if got := IsValidStorageType(tt.storageType); got != tt.valid {
				t.Errorf("IsValidStorageType(%s) = %v, want %v", tt.storageType, got, tt.valid)
			}
		})
	}
}

func TestIsValidAuthProvider(t *testing.T) {
	tests := []struct {
		authProvider string
		valid        bool
	}{
		{"local", true},
		{"oauth2", true},
		{"saml", true},
		{"sso", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.authProvider, func(t *testing.T) {
			if got := IsValidAuthProvider(tt.authProvider); got != tt.valid {
				t.Errorf("IsValidAuthProvider(%s) = %v, want %v", tt.authProvider, got, tt.valid)
			}
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		domain string
		valid  bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"projects.company.co.uk", true},
		{"a.b", true},
		{"", false},
		{"..invalid", false},
		{".invalid", false},
		{"invalid.", false},
		{"invalid..domain", false},
		{"inv@lid.com", false},
		{string(make([]byte, 254)), false}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			if got := isValidDomain(tt.domain); got != tt.valid {
				t.Errorf("isValidDomain(%s) = %v, want %v", tt.domain, got, tt.valid)
			}
		})
	}
}
