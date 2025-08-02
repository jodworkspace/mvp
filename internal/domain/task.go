package domain

import "time"

type Task struct {
	ID            string    `json:"id" db:"id"`
	Title         string    `json:"title"`
	Details       string    `json:"details"`
	IsCompleted   bool      `json:"isCompleted"`
	PriorityLevel int       `json:"priorityLevel"`
	StartDate     time.Time `json:"startDate"`
	DueDate       time.Time `json:"dueDate"`
	OwnerUserID   string    `json:"ownerUserId"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

const (
	TableTask            = "tasks"
	ColTaskTitle         = "title"
	ColTaskDetails       = "details"
	ColTaskPriorityLevel = "priority_level"
	ColTaskIsCompleted   = "is_completed"
	ColTaskStartDate     = "start_date"
	ColTaskDueDate       = "due_date"
	ColTaskOwnerID       = "owner_id"
)

var (
	TaskAllColumns = []string{
		ColID,
		ColTaskTitle,
		ColTaskDetails,
		ColTaskPriorityLevel,
		ColTaskIsCompleted,
		ColTaskStartDate,
		ColTaskDueDate,
		ColTaskOwnerID,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
