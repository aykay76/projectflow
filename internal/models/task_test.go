package models

import (
	"testing"
	"time"
)

func TestTask_SetDueDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{
			name:    "valid date",
			dateStr: "2025-06-07",
			wantErr: false,
		},
		{
			name:    "empty date",
			dateStr: "",
			wantErr: false,
		},
		{
			name:    "invalid date format",
			dateStr: "06/07/2025",
			wantErr: true,
		},
		{
			name:    "invalid date",
			dateStr: "2025-13-32",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{}
			err := task.SetDueDate(tt.dateStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("Task.SetDueDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.dateStr != "" {
				expected, _ := time.Parse("2006-01-02", tt.dateStr)
				if task.DueDate == nil || !task.DueDate.Equal(expected) {
					t.Errorf("Task.SetDueDate() = %v, want %v", task.DueDate, expected)
				}
			}

			if tt.dateStr == "" && task.DueDate != nil {
				t.Errorf("Task.SetDueDate() with empty string should set DueDate to nil")
			}
		})
	}
}

func TestTask_GetDueDateString(t *testing.T) {
	tests := []struct {
		name     string
		dueDate  *time.Time
		expected string
	}{
		{
			name:     "nil due date",
			dueDate:  nil,
			expected: "",
		},
		{
			name:     "valid due date",
			dueDate:  &time.Time{},
			expected: "0001-01-01",
		},
	}

	// Set up a specific date for testing
	testDate := time.Date(2025, 6, 7, 0, 0, 0, 0, time.UTC)
	tests[1].dueDate = &testDate
	tests[1].expected = "2025-06-07"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{DueDate: tt.dueDate}
			result := task.GetDueDateString()

			if result != tt.expected {
				t.Errorf("Task.GetDueDateString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTask_IsOverdue(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		dueDate  *time.Time
		status   TaskStatus
		expected bool
	}{
		{
			name:     "nil due date",
			dueDate:  nil,
			status:   StatusTodo,
			expected: false,
		},
		{
			name:     "due date in future",
			dueDate:  &tomorrow,
			status:   StatusTodo,
			expected: false,
		},
		{
			name:     "due date in past",
			dueDate:  &yesterday,
			status:   StatusTodo,
			expected: true,
		},
		{
			name:     "overdue but done",
			dueDate:  &yesterday,
			status:   StatusDone,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				DueDate: tt.dueDate,
				Status:  tt.status,
			}
			result := task.IsOverdue()

			if result != tt.expected {
				t.Errorf("Task.IsOverdue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTask_DaysUntilDue(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		dueDate  *time.Time
		expected int
	}{
		{
			name:     "nil due date",
			dueDate:  nil,
			expected: 0,
		},
		{
			name:     "due tomorrow",
			dueDate:  &tomorrow,
			expected: 1,
		},
		{
			name:     "overdue by one day",
			dueDate:  &yesterday,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{DueDate: tt.dueDate}
			result := task.DaysUntilDue()

			// Allow for some variance due to timing
			if tt.dueDate != nil && (result < tt.expected-1 || result > tt.expected+1) {
				t.Errorf("Task.DaysUntilDue() = %v, want approximately %v", result, tt.expected)
			} else if tt.dueDate == nil && result != tt.expected {
				t.Errorf("Task.DaysUntilDue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	title := "Test Task"
	description := "Test Description"

	task := NewTask(title, description)

	if task.Title != title {
		t.Errorf("NewTask() title = %v, want %v", task.Title, title)
	}

	if task.Description != description {
		t.Errorf("NewTask() description = %v, want %v", task.Description, description)
	}

	if task.Status != StatusTodo {
		t.Errorf("NewTask() status = %v, want %v", task.Status, StatusTodo)
	}

	if task.Priority != PriorityMedium {
		t.Errorf("NewTask() priority = %v, want %v", task.Priority, PriorityMedium)
	}

	if task.Type != TypeTask {
		t.Errorf("NewTask() type = %v, want %v", task.Type, TypeTask)
	}

	if task.Children == nil {
		t.Errorf("NewTask() children should be initialized")
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"todo", true},
		{"in_progress", true},
		{"done", true},
		{"blocked", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			if result != tt.valid {
				t.Errorf("IsValidStatus(%v) = %v, want %v", tt.status, result, tt.valid)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		valid    bool
	}{
		{"low", true},
		{"medium", true},
		{"high", true},
		{"critical", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			result := IsValidPriority(tt.priority)
			if result != tt.valid {
				t.Errorf("IsValidPriority(%v) = %v, want %v", tt.priority, result, tt.valid)
			}
		})
	}
}

func TestIsValidType(t *testing.T) {
	tests := []struct {
		taskType string
		valid    bool
	}{
		{"epic", true},
		{"story", true},
		{"task", true},
		{"subtask", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			result := IsValidType(tt.taskType)
			if result != tt.valid {
				t.Errorf("IsValidType(%v) = %v, want %v", tt.taskType, result, tt.valid)
			}
		})
	}
}
