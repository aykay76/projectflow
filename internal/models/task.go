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
	DueDate     *time.Time   `json:"due_date,omitempty"`
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

// HierarchyTask represents a task with its nested children for hierarchy view
type HierarchyTask struct {
	*Task
	ChildTasks []*HierarchyTask `json:"child_tasks"`
}
