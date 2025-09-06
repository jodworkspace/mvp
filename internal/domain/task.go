package domain

import "time"

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Details     string     `json:"details"`
	Priority    int        `json:"priority"`
	IsCompleted bool       `json:"isCompleted" db:"is_completed"`
	StartDate   *time.Time `json:"startDate" db:"start_date"`
	DueDate     *time.Time `json:"dueDate" db:"due_date"`
	OwnerID     string     `json:"ownerID" db:"owner_id"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

const (
	TableTask          = "tasks"
	ColTaskTitle       = "title"
	ColTaskDetails     = "details"
	ColTaskPriority    = "priority"
	ColTaskIsCompleted = "is_completed"
	ColTaskStartDate   = "start_date"
	ColTaskDueDate     = "due_date"
	ColTaskOwnerID     = "owner_id"
)

var (
	TaskAllColumns = []string{
		ColID,
		ColTaskTitle,
		ColTaskDetails,
		ColTaskPriority,
		ColTaskIsCompleted,
		ColTaskStartDate,
		ColTaskDueDate,
		ColTaskOwnerID,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
