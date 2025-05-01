package domain

type User struct {
	Base
	DisplayName       string `json:"display_name" db:"display_name"`
	Email             string `json:"email" db:"email"`
	EmailVerified     bool   `json:"email_verified" db:"email_verified"`
	Password          string `json:"password" db:"password"`
	PIN               string `json:"pin" db:"pin"`
	AvatarURL         string `json:"avatar_url" db:"avatar_url"`
	PreferredLanguage string `json:"preferred_language" db:"preferred_language"`
	Active            bool   `json:"active" db:"active"`
}

type FederatedUser struct {
	UserID       string `json:"user_id" db:"user_id"`
	Provider     string `json:"provider" db:"provider"`
	ExternalID   string `json:"external_id" db:"external_user_id"`
	AccessToken  string `json:"access_token" db:"access_token"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
}

const (
	TableUser                = "users"
	ColUserDisplayName       = "display_name"
	ColUserEmail             = "email"
	ColUserEmailVerified     = "email_verified"
	ColUserPassword          = "password"
	ColUserPIN               = "pin"
	ColUserAvatarURL         = "avatar_url"
	ColUserPreferredLanguage = "preferred_language"
	ColUserActive            = "active"
)

var (
	UserPublicCol = []string{
		ColID,
		ColUserDisplayName,
		ColUserEmail,
		ColUserEmailVerified,
		ColUserAvatarURL,
		ColUserPreferredLanguage,
		ColUserActive,
		ColCreatedAt,
		ColUpdatedAt,
	}

	UserProtectedCol = []string{
		ColUserPassword,
		ColUserPIN,
	}
)
