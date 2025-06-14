package models

import (
	"errors"
	"strings"
	"time"
)

// Project validation errors
var (
	ErrProjectNameRequired   = errors.New("project name is required")
	ErrProjectPrefixRequired = errors.New("project display prefix is required")
	ErrProjectPrefixTooLong  = errors.New("project display prefix cannot exceed 10 characters")
)

// Project represents a project in the system that contains tasks
type Project struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	DisplayPrefix string            `json:"display_prefix"` // e.g., "PF", "WEB" for human-readable task IDs
	Settings      map[string]string `json:"settings"`       // Project-specific configuration
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// NewProject creates a new project with default values
func NewProject(name, description, displayPrefix string) *Project {
	now := time.Now()
	return &Project{
		Name:          name,
		Description:   description,
		DisplayPrefix: strings.ToUpper(displayPrefix),
		Settings:      make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Validate checks if the project has valid data
func (p *Project) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrProjectNameRequired
	}
	if strings.TrimSpace(p.DisplayPrefix) == "" {
		return ErrProjectPrefixRequired
	}
	if len(p.DisplayPrefix) > 10 {
		return ErrProjectPrefixTooLong
	}
	return nil
}

// UpdateTimestamp updates the UpdatedAt field to current time
func (p *Project) UpdateTimestamp() {
	p.UpdatedAt = time.Now()
}

// GetSetting retrieves a setting value by key
func (p *Project) GetSetting(key string) (string, bool) {
	value, exists := p.Settings[key]
	return value, exists
}

// SetSetting sets a setting value
func (p *Project) SetSetting(key, value string) {
	if p.Settings == nil {
		p.Settings = make(map[string]string)
	}
	p.Settings[key] = value
	p.UpdateTimestamp()
}

// RemoveSetting removes a setting by key
func (p *Project) RemoveSetting(key string) {
	if p.Settings != nil {
		delete(p.Settings, key)
		p.UpdateTimestamp()
	}
}
