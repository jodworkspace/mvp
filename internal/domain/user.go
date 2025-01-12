package domain

type User struct {
	UserID            string `json:"user_id" db:"id"`
	Username          string `json:"username" db:"username"`
	FullName          string `json:"full_name" db:"full_name"`
	Email             string `json:"email" db:"email"`
	EmailVerified     bool   `json:"email_verified" db:"email_verified"`
	PIN               string `json:"pin" db:"pin"`
	AvatarURL         string `json:"avatar_url" db:"avatar_url"`
	PreferredLanguage string `json:"preferred_language" db:"preferred_language"`
	Active            bool   `json:"active" db:"active"`
	CreatedAt         string `json:"created_at" db:"created_at"`
	UpdatedAt         string `json:"updated_at" db:"updated_at"`
}
