package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProject_NewProject(t *testing.T) {
	project := NewProject("Test Project", "A test project", "TEST")

	if project.Name != "Test Project" {
		t.Errorf("NewProject() name = %v, want %v", project.Name, "Test Project")
	}

	if project.Description != "A test project" {
		t.Errorf("NewProject() description = %v, want %v", project.Description, "A test project")
	}

	if project.ID != "" {
		t.Errorf("NewProject() should not set ID automatically")
	}

	if project.DisplayPrefix != "TEST" {
		t.Errorf("NewProject() DisplayPrefix = %v, want %v", project.DisplayPrefix, "TEST")
	}

	if project.Settings == nil {
		t.Errorf("NewProject() should initialize Settings map")
	}

	if project.CreatedAt.IsZero() {
		t.Errorf("NewProject() should set CreatedAt")
	}

	if project.UpdatedAt.IsZero() {
		t.Errorf("NewProject() should set UpdatedAt")
	}
}

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project *Project
		wantErr bool
	}{
		{
			name:    "valid project",
			project: NewProject("Valid Project", "Valid description", "VALID"),
			wantErr: false,
		},
		{
			name: "empty name",
			project: &Project{
				Name:          "",
				Description:   "Valid description",
				DisplayPrefix: "VALID",
			},
			wantErr: true,
		},
		{
			name: "empty display prefix",
			project: &Project{
				Name:          "Valid Project",
				Description:   "Valid description",
				DisplayPrefix: "",
			},
			wantErr: true,
		},
		{
			name: "display prefix too long",
			project: &Project{
				Name:          "Valid Project",
				Description:   "Valid description",
				DisplayPrefix: "TOOLONGPREFIX", // More than 10 characters
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Project.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProject_UpdateTimestamp(t *testing.T) {
	project := NewProject("Test Project", "Test description", "TEST")
	originalTime := project.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	project.UpdateTimestamp()

	if !project.UpdatedAt.After(originalTime) {
		t.Error("UpdateTimestamp() should update the UpdatedAt field")
	}
}

func TestProject_SetSetting(t *testing.T) {
	project := NewProject("Test Project", "Test description", "TEST")

	project.SetSetting("testKey", "testValue")

	value, exists := project.GetSetting("testKey")
	if !exists {
		t.Error("SetSetting() should add the setting")
	}

	if value != "testValue" {
		t.Errorf("SetSetting() value = %v, want %v", value, "testValue")
	}
}

func TestProject_GetSetting(t *testing.T) {
	project := NewProject("Test Project", "Test description", "TEST")

	// Test non-existent setting
	_, exists := project.GetSetting("nonExistent")
	if exists {
		t.Error("GetSetting() should return false for non-existent setting")
	}

	// Test existing setting
	project.SetSetting("existingKey", "existingValue")
	value, exists := project.GetSetting("existingKey")
	if !exists {
		t.Error("GetSetting() should return true for existing setting")
	}

	if value != "existingValue" {
		t.Errorf("GetSetting() value = %v, want %v", value, "existingValue")
	}
}

func TestProject_RemoveSetting(t *testing.T) {
	project := NewProject("Test Project", "Test description", "TEST")

	// Add a setting first
	project.SetSetting("testKey", "testValue")

	// Remove it
	project.RemoveSetting("testKey")

	// Verify it's gone
	_, exists := project.GetSetting("testKey")
	if exists {
		t.Error("RemoveSetting() should remove the setting")
	}
}

func TestProject_JSONSerialization(t *testing.T) {
	original := NewProject("Test Project", "Test description", "TEST")
	original.ID = "test-id-123"
	original.SetSetting("key1", "value1")
	original.SetSetting("key2", "42") // Settings are strings

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal project: %v", err)
	}

	// Unmarshal from JSON
	var restored Project
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal project: %v", err)
	}

	// Verify all fields are preserved
	if restored.ID != original.ID {
		t.Errorf("JSON serialization ID = %v, want %v", restored.ID, original.ID)
	}

	if restored.Name != original.Name {
		t.Errorf("JSON serialization Name = %v, want %v", restored.Name, original.Name)
	}

	if restored.Description != original.Description {
		t.Errorf("JSON serialization Description = %v, want %v", restored.Description, original.Description)
	}

	if restored.DisplayPrefix != original.DisplayPrefix {
		t.Errorf("JSON serialization DisplayPrefix = %v, want %v", restored.DisplayPrefix, original.DisplayPrefix)
	}

	// Verify settings
	value1, exists1 := restored.GetSetting("key1")
	if !exists1 || value1 != "value1" {
		t.Errorf("JSON serialization settings key1 = %v (exists: %v), want value1 (true)", value1, exists1)
	}

	value2, exists2 := restored.GetSetting("key2")
	if !exists2 || value2 != "42" {
		t.Errorf("JSON serialization settings key2 = %v (exists: %v), want 42 (true)", value2, exists2)
	}

	// Verify timestamps are preserved (approximately)
	if restored.CreatedAt.Unix() != original.CreatedAt.Unix() {
		t.Errorf("JSON serialization CreatedAt mismatch")
	}

	if restored.UpdatedAt.Unix() != original.UpdatedAt.Unix() {
		t.Errorf("JSON serialization UpdatedAt mismatch")
	}
}

func BenchmarkProject_NewProject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewProject("Benchmark Project", "Benchmark description", "BENCH")
	}
}

func BenchmarkProject_SetSetting(b *testing.B) {
	project := NewProject("Benchmark Project", "Benchmark description", "BENCH")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		project.SetSetting("key", "value")
	}
}

func BenchmarkProject_GetSetting(b *testing.B) {
	project := NewProject("Benchmark Project", "Benchmark description", "BENCH")
	project.SetSetting("key", "value")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		project.GetSetting("key")
	}
}

func BenchmarkProject_JSONMarshal(b *testing.B) {
	project := NewProject("Benchmark Project", "Benchmark description", "BENCH")
	project.SetSetting("key1", "value1")
	project.SetSetting("key2", "42")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		json.Marshal(project)
	}
}

func BenchmarkProject_JSONUnmarshal(b *testing.B) {
	project := NewProject("Benchmark Project", "Benchmark description", "BENCH")
	project.SetSetting("key1", "value1")
	data, _ := json.Marshal(project)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var p Project
		json.Unmarshal(data, &p)
	}
}
