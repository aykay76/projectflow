package models

import (
	"errors"
	"strings"
	"time"
)

// Tenant validation errors
var (
	ErrTenantNameRequired        = errors.New("tenant name is required")
	ErrTenantNameTooLong         = errors.New("tenant name cannot exceed 255 characters")
	ErrTenantDomainInvalid       = errors.New("tenant domain is invalid")
	ErrTenantStorageTypeEmpty    = errors.New("tenant storage type is required")
	ErrTenantAuthProviderInvalid = errors.New("tenant auth provider is invalid")
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
	TenantStatusPending   TenantStatus = "pending"
)

// StorageType represents the storage backend type for a tenant
type StorageType string

const (
	StorageTypeFile     StorageType = "file"
	StorageTypePostgres StorageType = "postgres"
)

// AuthProvider represents the authentication provider for a tenant
type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderOAuth2 AuthProvider = "oauth2"
	AuthProviderSAML   AuthProvider = "saml"
	AuthProviderSSO    AuthProvider = "sso"
)

// TenantSettings contains tenant-specific configuration
type TenantSettings struct {
	StorageType    StorageType            `json:"storage_type"`
	DatabaseConfig map[string]string      `json:"database_config,omitempty"`
	AuthProvider   AuthProvider           `json:"auth_provider"`
	AuthConfig     map[string]string      `json:"auth_config,omitempty"`
	CustomDomain   string                 `json:"custom_domain,omitempty"`
	Branding       map[string]string      `json:"branding,omitempty"`
	Features       []string               `json:"features,omitempty"`
	Limits         map[string]int         `json:"limits,omitempty"`
	Preferences    map[string]interface{} `json:"preferences,omitempty"`
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Status      TenantStatus   `json:"status"`
	Settings    TenantSettings `json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// NewTenant creates a new tenant with default values
func NewTenant(name, description string, storageType StorageType, authProvider AuthProvider) *Tenant {
	now := time.Now()
	return &Tenant{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Status:      TenantStatusPending,
		Settings: TenantSettings{
			StorageType:    storageType,
			AuthProvider:   authProvider,
			DatabaseConfig: make(map[string]string),
			AuthConfig:     make(map[string]string),
			Branding:       make(map[string]string),
			Features:       []string{},
			Limits:         make(map[string]int),
			Preferences:    make(map[string]interface{}),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate checks if the tenant has valid data
func (t *Tenant) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrTenantNameRequired
	}
	if len(t.Name) > 255 {
		return ErrTenantNameTooLong
	}
	if t.Settings.StorageType == "" {
		return ErrTenantStorageTypeEmpty
	}
	if !IsValidStorageType(string(t.Settings.StorageType)) {
		return ErrTenantStorageTypeEmpty
	}
	if !IsValidAuthProvider(string(t.Settings.AuthProvider)) {
		return ErrTenantAuthProviderInvalid
	}
	if t.Settings.CustomDomain != "" && !isValidDomain(t.Settings.CustomDomain) {
		return ErrTenantDomainInvalid
	}
	return nil
}

// Activate changes the tenant status to active
func (t *Tenant) Activate() {
	t.Status = TenantStatusActive
	t.UpdatedAt = time.Now()
}

// Suspend changes the tenant status to suspended
func (t *Tenant) Suspend() {
	t.Status = TenantStatusSuspended
	t.UpdatedAt = time.Now()
}

// Delete changes the tenant status to deleted (soft delete)
func (t *Tenant) Delete() {
	t.Status = TenantStatusDeleted
	t.UpdatedAt = time.Now()
}

// IsActive returns true if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// IsSuspended returns true if the tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == TenantStatusSuspended
}

// IsDeleted returns true if the tenant is deleted
func (t *Tenant) IsDeleted() bool {
	return t.Status == TenantStatusDeleted
}

// SetFeature adds or updates a feature in the tenant settings
func (t *Tenant) SetFeature(feature string) {
	for _, f := range t.Settings.Features {
		if f == feature {
			return // Already exists
		}
	}
	t.Settings.Features = append(t.Settings.Features, feature)
	t.UpdatedAt = time.Now()
}

// RemoveFeature removes a feature from the tenant settings
func (t *Tenant) RemoveFeature(feature string) {
	for i, f := range t.Settings.Features {
		if f == feature {
			t.Settings.Features = append(t.Settings.Features[:i], t.Settings.Features[i+1:]...)
			t.UpdatedAt = time.Now()
			return
		}
	}
}

// HasFeature checks if the tenant has a specific feature enabled
func (t *Tenant) HasFeature(feature string) bool {
	for _, f := range t.Settings.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// SetLimit sets a usage limit for the tenant
func (t *Tenant) SetLimit(limitType string, value int) {
	t.Settings.Limits[limitType] = value
	t.UpdatedAt = time.Now()
}

// GetLimit returns the usage limit for a specific type
func (t *Tenant) GetLimit(limitType string) (int, bool) {
	limit, exists := t.Settings.Limits[limitType]
	return limit, exists
}

// SetPreference sets a preference value for the tenant
func (t *Tenant) SetPreference(key string, value interface{}) {
	t.Settings.Preferences[key] = value
	t.UpdatedAt = time.Now()
}

// GetPreference returns a preference value for the tenant
func (t *Tenant) GetPreference(key string) (interface{}, bool) {
	value, exists := t.Settings.Preferences[key]
	return value, exists
}

// IsValidTenantStatus checks if the given tenant status is valid
func IsValidTenantStatus(status string) bool {
	switch TenantStatus(status) {
	case TenantStatusActive, TenantStatusSuspended, TenantStatusDeleted, TenantStatusPending:
		return true
	default:
		return false
	}
}

// IsValidStorageType checks if the given storage type is valid
func IsValidStorageType(storageType string) bool {
	switch StorageType(storageType) {
	case StorageTypeFile, StorageTypePostgres:
		return true
	default:
		return false
	}
}

// IsValidAuthProvider checks if the given auth provider is valid
func IsValidAuthProvider(authProvider string) bool {
	switch AuthProvider(authProvider) {
	case AuthProviderLocal, AuthProviderOAuth2, AuthProviderSAML, AuthProviderSSO:
		return true
	default:
		return false
	}
}

// isValidDomain performs basic domain validation
func isValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Basic domain validation - just check for basic format
	// In production, you might want more sophisticated validation
	if strings.Contains(domain, "..") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	// Check for valid characters (simplified)
	for _, char := range domain {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '.' || char == '-') {
			return false
		}
	}

	return true
}
