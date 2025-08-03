package domain

import "time"

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Details     string     `json:"details"`
	IsCompleted bool       `json:"isCompleted"`
	Priority    int        `json:"priority"`
	StartDate   *time.Time `json:"startDate"`
	DueDate     *time.Time `json:"dueDate"`
	OwnerID     string     `json:"ownerID"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
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
