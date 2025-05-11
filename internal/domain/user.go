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
	Base
	UserID       string `json:"user_id" db:"user_id"`
	Issuer       string `json:"issuer" db:"issuer"`
	ExternalID   string `json:"external_id" db:"external_id"`
	AccessToken  string `json:"access_token" db:"access_token"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
}

const (
	TableUser            = "users"
	ColDisplayName       = "display_name"
	ColEmail             = "email"
	ColEmailVerified     = "email_verified"
	ColPassword          = "password"
	ColPIN               = "pin"
	ColAvatarURL         = "avatar_url"
	ColPreferredLanguage = "preferred_language"
	ColActive            = "active"

	TableFederatedUser = "federated_users"
	ColUserID          = "user_id"
	ColIssuer          = "issuer"
	ColExternalID      = "external_id"
	ColAccessToken     = "access_token"
	ColRefreshToken    = "refresh_token"
)

var (
	UserPublicCols = []string{
		ColID,
		ColDisplayName,
		ColEmail,
		ColEmailVerified,
		ColAvatarURL,
		ColPreferredLanguage,
		ColActive,
		ColCreatedAt,
		ColUpdatedAt,
	}

	UserProtectedCols = []string{
		ColPassword,
		ColPIN,
	}

	FederatedUserAllCols = []string{
		ColID,
		ColUserID,
		ColIssuer,
		ColExternalID,
		ColAccessToken,
		ColRefreshToken,
		ColCreatedAt,
		ColUpdatedAt,
	}
)
