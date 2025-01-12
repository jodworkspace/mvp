package domain

import "time"

type Base struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

const (
	ColID        = "id"
	ColCreatedAt = "created_at"
	ColUpdatedAt = "updated_at"
)
