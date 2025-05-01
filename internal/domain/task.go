package domain

import "time"

type Task struct {
	Base
	Title             string        `json:"title"`
	Details           string        `json:"details"`
	IsCompleted       bool          `json:"is_completed"`
	PriorityLevel     int           `json:"priority_level"`
	StartDate         time.Time     `json:"start_date"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	DueDate           time.Time     `json:"due_date"`
	OwnerUserID       string        `json:"owner_user_id"`
}

const (
	TableTask                = "tasks"
	ColTaskTitle             = "title"
	ColTaskDetails           = "details"
	ColTaskPriorityLevel     = "priority_level"
	ColTaskIsCompleted       = "is_completed"
	ColTaskStartDate         = "start_date"
	ColTaskEstimatedDuration = "estimated_duration"
	ColTaskDueDate           = "due_date"
	ColTaskOwnerID           = "owner_id"
)

var (
	TaskAllColumns = []string{
		ColID,
		ColTaskTitle,
		ColTaskDetails,
		ColTaskPriorityLevel,
		ColTaskIsCompleted,
		ColTaskStartDate,
		ColTaskEstimatedDuration,
		ColTaskDueDate,
		ColTaskOwnerID,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
