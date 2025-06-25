package domain

import "time"

type Document struct {
	ID            string    `json:"id"`
	FileName      string    `json:"fileName" db:"file_name"`
	FileExtension string    `json:"fileExtension" db:"file_extension"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}
