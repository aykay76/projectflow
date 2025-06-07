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

func TestTask_SetStartedAt(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{
			name:    "valid RFC3339 date with timezone",
			dateStr: "2025-06-07T15:04:05+07:00",
			wantErr: false,
		},
		{
			name:    "valid UTC date",
			dateStr: "2025-06-07T15:04:05Z",
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
			name:    "invalid RFC3339 date",
			dateStr: "2025-13-32T25:00:00Z",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{}
			err := task.SetStartedAt(tt.dateStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("Task.SetStartedAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.dateStr != "" {
				expected, _ := time.Parse(time.RFC3339, tt.dateStr)
				if task.StartedAt == nil || !task.StartedAt.Equal(expected) {
					t.Errorf("Task.SetStartedAt() = %v, want %v", task.StartedAt, expected)
				}
			}

			if !tt.wantErr && tt.dateStr == "" {
				if task.StartedAt != nil {
					t.Errorf("Task.SetStartedAt() with empty string should set StartedAt to nil")
				}
			}
		})
	}
}

func TestTask_GetStartedAtString(t *testing.T) {
	task := &Task{}

	// Test with nil start date
	if got := task.GetStartedAtString(); got != "" {
		t.Errorf("Task.GetStartedAtString() with nil start date = %v, want empty string", got)
	}

	// Test with actual start date
	startDate := time.Date(2025, 6, 7, 15, 4, 5, 0, time.UTC)
	task.StartedAt = &startDate
	expected := startDate.Format(time.RFC3339)
	if got := task.GetStartedAtString(); got != expected {
		t.Errorf("Task.GetStartedAtString() = %v, want %v", got, expected)
	}
}

func TestTask_StartTask(t *testing.T) {
	task := &Task{
		Status: StatusTodo,
	}

	// Test starting a task that hasn't been started
	task.StartTask()

	if task.Status != StatusInProgress {
		t.Errorf("Task.StartTask() status = %v, want %v", task.Status, StatusInProgress)
	}

	if task.StartedAt == nil {
		t.Errorf("Task.StartTask() should set StartedAt when not already set")
	}

	// Test starting a task that already has a start date
	originalStartDate := *task.StartedAt
	time.Sleep(time.Millisecond) // Ensure time difference
	task.StartTask()

	if !task.StartedAt.Equal(originalStartDate) {
		t.Errorf("Task.StartTask() should not change existing StartedAt")
	}
}

func TestTask_GetActualDuration(t *testing.T) {
	// Test with no start date
	task := &Task{}
	if duration := task.GetActualDuration(); duration != 0 {
		t.Errorf("Task.GetActualDuration() with no start date = %v, want 0", duration)
	}

	// Test with start date but not completed
	startTime := time.Now().Add(-2 * time.Hour)
	task.StartedAt = &startTime
	task.Status = StatusInProgress

	duration := task.GetActualDuration()
	if duration < time.Hour || duration > 3*time.Hour {
		t.Errorf("Task.GetActualDuration() for in-progress task = %v, expected around 2 hours", duration)
	}

	// Test with completed task
	task.Status = StatusDone
	task.UpdatedAt = startTime.Add(time.Hour)

	duration = task.GetActualDuration()
	expectedDuration := time.Hour
	if duration != expectedDuration {
		t.Errorf("Task.GetActualDuration() for completed task = %v, want %v", duration, expectedDuration)
	}
}

func TestTask_GetActualDurationDays(t *testing.T) {
	task := &Task{}

	// Test with no start date
	if days := task.GetActualDurationDays(); days != 0 {
		t.Errorf("Task.GetActualDurationDays() with no start date = %v, want 0", days)
	}

	// Test with 2-day duration
	startTime := time.Now().Add(-48 * time.Hour)
	task.StartedAt = &startTime
	task.Status = StatusDone
	task.UpdatedAt = startTime.Add(48 * time.Hour)

	days := task.GetActualDurationDays()
	if days != 2 {
		t.Errorf("Task.GetActualDurationDays() = %v, want 2", days)
	}
}

func TestTask_IsStarted(t *testing.T) {
	task := &Task{}

	// Test with no start date
	if task.IsStarted() {
		t.Errorf("Task.IsStarted() with no start date should return false")
	}

	// Test with start date
	now := time.Now()
	task.StartedAt = &now
	if !task.IsStarted() {
		t.Errorf("Task.IsStarted() with start date should return true")
	}
}

func TestTask_CompleteTask(t *testing.T) {
	task := NewTask("Test Task", "Test Description")

	// Task should not be completed initially
	if task.IsCompleted() {
		t.Errorf("Task.IsCompleted() should return false for new task")
	}

	if task.Status == StatusDone {
		t.Errorf("Task.Status should not be StatusDone for new task")
	}

	// Complete the task
	beforeComplete := time.Now()
	task.CompleteTask()
	afterComplete := time.Now()

	// Check task is marked as completed
	if !task.IsCompleted() {
		t.Errorf("Task.IsCompleted() should return true after CompleteTask()")
	}

	if task.Status != StatusDone {
		t.Errorf("Task.Status should be StatusDone after CompleteTask(), got %v", task.Status)
	}

	// Check completion date is set correctly
	if task.CompletedAt == nil {
		t.Errorf("Task.CompletedAt should not be nil after CompleteTask()")
	} else {
		if task.CompletedAt.Before(beforeComplete) || task.CompletedAt.After(afterComplete) {
			t.Errorf("Task.CompletedAt should be between %v and %v, got %v", beforeComplete, afterComplete, task.CompletedAt)
		}
	}
}

func TestTask_SetCompletedDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{
			name:    "valid RFC3339 date",
			dateStr: "2025-06-07T15:04:05Z",
			wantErr: false,
		},
		{
			name:    "valid RFC3339 date with timezone",
			dateStr: "2025-06-07T15:04:05+02:00",
			wantErr: false,
		},
		{
			name:    "empty date",
			dateStr: "",
			wantErr: false,
		},
		{
			name:    "invalid date format",
			dateStr: "2025-06-07",
			wantErr: true,
		},
		{
			name:    "invalid date",
			dateStr: "invalid-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("Test Task", "Test Description")
			err := task.SetCompletedDate(tt.dateStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("Task.SetCompletedDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.dateStr == "" {
					if task.CompletedAt != nil {
						t.Errorf("Task.CompletedAt should be nil for empty date string")
					}
				} else {
					expected, _ := time.Parse(time.RFC3339, tt.dateStr)
					if task.CompletedAt == nil || !task.CompletedAt.Equal(expected) {
						t.Errorf("Task.CompletedAt = %v, want %v", task.CompletedAt, expected)
					}
					if task.Status != StatusDone {
						t.Errorf("Task.Status should be StatusDone after setting completion date, got %v", task.Status)
					}
				}
			}
		})
	}
}

func TestTask_GetCompletedDateString(t *testing.T) {
	task := NewTask("Test Task", "Test Description")

	// Test with no completion date
	if got := task.GetCompletedDateString(); got != "" {
		t.Errorf("Task.GetCompletedDateString() with no completion date = %v, want empty string", got)
	}

	// Test with completion date
	testDate := time.Date(2025, 6, 7, 15, 4, 5, 0, time.UTC)
	task.CompletedAt = &testDate
	expected := "2025-06-07T15:04:05Z"
	if got := task.GetCompletedDateString(); got != expected {
		t.Errorf("Task.GetCompletedDateString() = %v, want %v", got, expected)
	}
}

func TestTask_GetActualDurationWithCompletion(t *testing.T) {
	task := NewTask("Test Task", "Test Description")

	// Test with no start date
	if duration := task.GetActualDuration(); duration != 0 {
		t.Errorf("Task.GetActualDuration() with no start date should return 0, got %v", duration)
	}

	// Set start date
	startDate := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	task.StartedAt = &startDate

	// Test with start date but no completion (should use current time)
	duration := task.GetActualDuration()
	if duration <= 0 {
		t.Errorf("Task.GetActualDuration() with start date should be positive, got %v", duration)
	}

	// Set completion date
	completionDate := time.Date(2025, 6, 5, 16, 30, 0, 0, time.UTC)
	task.CompletedAt = &completionDate

	// Test with both start and completion dates
	expectedDuration := completionDate.Sub(startDate)
	actualDuration := task.GetActualDuration()
	if actualDuration != expectedDuration {
		t.Errorf("Task.GetActualDuration() = %v, want %v", actualDuration, expectedDuration)
	}
}

func TestTask_DeliveryPerformance(t *testing.T) {
	// Create a task with due date
	task := NewTask("Test Task", "Test Description")
	dueDate := time.Date(2025, 6, 10, 17, 0, 0, 0, time.UTC)
	task.DueDate = &dueDate

	// Test without completion date
	if task.IsDeliveredEarly() || task.IsDeliveredLate() || task.IsDeliveredOnTime() {
		t.Errorf("Delivery methods should return false when task is not completed")
	}

	// Test early delivery
	earlyCompletion := time.Date(2025, 6, 8, 15, 0, 0, 0, time.UTC)
	task.CompletedAt = &earlyCompletion

	if !task.IsDeliveredEarly() {
		t.Errorf("Task.IsDeliveredEarly() should return true for early completion")
	}
	if task.IsDeliveredLate() || task.IsDeliveredOnTime() {
		t.Errorf("Only IsDeliveredEarly() should be true for early completion")
	}

	expectedVariance := earlyCompletion.Sub(dueDate)
	if variance := task.GetDeliveryVariance(); variance != expectedVariance {
		t.Errorf("Task.GetDeliveryVariance() = %v, want %v", variance, expectedVariance)
	}

	// Test late delivery
	lateCompletion := time.Date(2025, 6, 12, 10, 0, 0, 0, time.UTC)
	task.CompletedAt = &lateCompletion

	if !task.IsDeliveredLate() {
		t.Errorf("Task.IsDeliveredLate() should return true for late completion")
	}
	if task.IsDeliveredEarly() || task.IsDeliveredOnTime() {
		t.Errorf("Only IsDeliveredLate() should be true for late completion")
	}

	// Test on-time delivery (same day)
	onTimeCompletion := time.Date(2025, 6, 10, 14, 0, 0, 0, time.UTC)
	task.CompletedAt = &onTimeCompletion

	if !task.IsDeliveredOnTime() {
		t.Errorf("Task.IsDeliveredOnTime() should return true for same-day completion")
	}
	if task.IsDeliveredEarly() || task.IsDeliveredLate() {
		t.Errorf("Only IsDeliveredOnTime() should be true for same-day completion")
	}

	// Test delivery variance in days
	task.CompletedAt = &lateCompletion
	expectedDays := 2 // 2 days late
	if days := task.GetDeliveryVarianceDays(); days != expectedDays {
		t.Errorf("Task.GetDeliveryVarianceDays() = %v, want %v", days, expectedDays)
	}
}
