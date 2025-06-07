package models

import (
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
	StatusBlocked    TaskStatus = "blocked"
)

// TaskPriority represents the priority of a task
type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

// TaskType represents the type of task
type TaskType string

const (
	TypeEpic    TaskType = "epic"
	TypeStory   TaskType = "story"
	TypeTask    TaskType = "task"
	TypeSubtask TaskType = "subtask"
)

// Task represents a work item in the system
type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Type        TaskType     `json:"type"`
	ParentID    string       `json:"parent_id,omitempty"`
	Children    []string     `json:"children"`
	StartDate   *time.Time   `json:"start_date,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// NewTask creates a new task with default values
func NewTask(title, description string) *Task {
	now := time.Now()
	return &Task{
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		Priority:    PriorityMedium,
		Type:        TypeTask,
		Children:    []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AddChild adds a child task ID to this task
func (t *Task) AddChild(childID string) {
	for _, child := range t.Children {
		if child == childID {
			return // Already exists
		}
	}
	t.Children = append(t.Children, childID)
	t.UpdatedAt = time.Now()
}

// RemoveChild removes a child task ID from this task
func (t *Task) RemoveChild(childID string) {
	for i, child := range t.Children {
		if child == childID {
			t.Children = append(t.Children[:i], t.Children[i+1:]...)
			t.UpdatedAt = time.Now()
			return
		}
	}
}

// IsValidStatus checks if the given status is valid
func IsValidStatus(status string) bool {
	switch TaskStatus(status) {
	case StatusTodo, StatusInProgress, StatusDone, StatusBlocked:
		return true
	default:
		return false
	}
}

// IsValidPriority checks if the given priority is valid
func IsValidPriority(priority string) bool {
	switch TaskPriority(priority) {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	default:
		return false
	}
}

// IsValidType checks if the given type is valid
func IsValidType(taskType string) bool {
	switch TaskType(taskType) {
	case TypeEpic, TypeStory, TypeTask, TypeSubtask:
		return true
	default:
		return false
	}
}

// IsOverdue checks if the task is overdue
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusDone {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// DaysUntilDue returns the number of days until the task is due
// Returns negative values for overdue tasks
func (t *Task) DaysUntilDue() int {
	if t.DueDate == nil {
		return 0
	}
	duration := time.Until(*t.DueDate)
	return int(duration.Hours() / 24)
}

// SetDueDate sets the due date from a string in YYYY-MM-DD format
func (t *Task) SetDueDate(dateStr string) error {
	if dateStr == "" {
		t.DueDate = nil
		return nil
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err
	}

	t.DueDate = &parsedDate
	t.UpdatedAt = time.Now()
	return nil
}

// GetDueDateString returns the due date as a string in YYYY-MM-DD format
func (t *Task) GetDueDateString() string {
	if t.DueDate == nil {
		return ""
	}
	return t.DueDate.Format("2006-01-02")
}

// SetStartDate sets the start date from a string in RFC3339 format
func (t *Task) SetStartDate(dateStr string) error {
	if dateStr == "" {
		t.StartDate = nil
		return nil
	}

	parsedDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return err
	}

	t.StartDate = &parsedDate
	t.UpdatedAt = time.Now()
	return nil
}

// GetStartDateString returns the start date as a string in RFC3339 format
func (t *Task) GetStartDateString() string {
	if t.StartDate == nil {
		return ""
	}
	return t.StartDate.Format(time.RFC3339)
}

// StartTask sets the task status to in_progress and sets the start date to now if not already set
func (t *Task) StartTask() {
	if t.StartDate == nil {
		now := time.Now()
		t.StartDate = &now
	}
	t.Status = StatusInProgress
	t.UpdatedAt = time.Now()
}

// CompleteTask marks the task as done and sets the completion date
func (t *Task) CompleteTask() {
	now := time.Now()
	t.CompletedAt = &now
	t.Status = StatusDone
	t.UpdatedAt = now
}

// IsCompleted returns true if the task has been completed (has a completion date)
func (t *Task) IsCompleted() bool {
	return t.CompletedAt != nil
}

// GetCompletedDateString returns the completion date as a string in RFC3339 format
func (t *Task) GetCompletedDateString() string {
	if t.CompletedAt == nil {
		return ""
	}
	return t.CompletedAt.Format(time.RFC3339)
}

// SetCompletedDate sets the completion date from a string in RFC3339 format
func (t *Task) SetCompletedDate(dateStr string) error {
	if dateStr == "" {
		t.CompletedAt = nil
		return nil
	}

	parsedDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return err
	}

	t.CompletedAt = &parsedDate
	t.Status = StatusDone
	t.UpdatedAt = time.Now()
	return nil
}

// GetActualDuration returns the duration from start date to completion (or now if not completed)
func (t *Task) GetActualDuration() time.Duration {
	if t.StartDate == nil {
		return 0
	}

	endTime := time.Now()
	if t.CompletedAt != nil {
		endTime = *t.CompletedAt
	} else if t.Status == StatusDone && t.UpdatedAt.After(*t.StartDate) {
		// Backward compatibility: use UpdatedAt for completed tasks without CompletedAt
		endTime = t.UpdatedAt
	}

	return endTime.Sub(*t.StartDate)
}

// GetActualDurationDays returns the actual duration in days
func (t *Task) GetActualDurationDays() int {
	duration := t.GetActualDuration()
	return int(duration.Hours() / 24)
}

// IsStarted returns true if the task has been started (has a start date)
func (t *Task) IsStarted() bool {
	return t.StartDate != nil
}

// IsDeliveredEarly returns true if the task was completed before the due date (different day)
func (t *Task) IsDeliveredEarly() bool {
	if t.CompletedAt == nil || t.DueDate == nil {
		return false
	}
	// Check if completed on a different (earlier) day
	completedDay := t.CompletedAt.Truncate(24 * time.Hour)
	dueDay := t.DueDate.Truncate(24 * time.Hour)
	return completedDay.Before(dueDay)
}

// IsDeliveredLate returns true if the task was completed after the due date (different day)
func (t *Task) IsDeliveredLate() bool {
	if t.CompletedAt == nil || t.DueDate == nil {
		return false
	}
	// Check if completed on a different (later) day
	completedDay := t.CompletedAt.Truncate(24 * time.Hour)
	dueDay := t.DueDate.Truncate(24 * time.Hour)
	return completedDay.After(dueDay)
}

// IsDeliveredOnTime returns true if the task was completed on the due date
func (t *Task) IsDeliveredOnTime() bool {
	if t.CompletedAt == nil || t.DueDate == nil {
		return false
	}
	// Consider same day as on time (truncate to day level)
	completedDay := t.CompletedAt.Truncate(24 * time.Hour)
	dueDay := t.DueDate.Truncate(24 * time.Hour)
	return completedDay.Equal(dueDay)
}

// GetDeliveryVariance returns the duration between completion and due date
// Positive values indicate late delivery, negative values indicate early delivery
func (t *Task) GetDeliveryVariance() time.Duration {
	if t.CompletedAt == nil || t.DueDate == nil {
		return 0
	}
	return t.CompletedAt.Sub(*t.DueDate)
}

// GetDeliveryVarianceDays returns the delivery variance in days
func (t *Task) GetDeliveryVarianceDays() int {
	variance := t.GetDeliveryVariance()
	// Round to nearest day for more intuitive results
	hours := variance.Hours()
	days := hours / 24
	if days >= 0 {
		return int(days + 0.5) // Round up for positive values
	}
	return int(days - 0.5) // Round down for negative values
}

// HierarchyTask represents a task with its nested children for hierarchy view
type HierarchyTask struct {
	*Task
	ChildTasks []*HierarchyTask `json:"child_tasks"`
}
