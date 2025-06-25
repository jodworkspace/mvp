package domain

import "time"

type User struct {
	ID                string    `json:"id" db:"id"`
	DisplayName       string    `json:"displayName" db:"display_name"`
	Email             string    `json:"email" db:"email"`
	EmailVerified     bool      `json:"emailVerified" db:"email_verified"`
	Password          string    `json:"password" db:"password"`
	PIN               string    `json:"pin" db:"pin"`
	AvatarURL         string    `json:"avatarUrl" db:"avatar_url"`
	PreferredLanguage string    `json:"preferredLanguage" db:"preferred_language"`
	Active            bool      `json:"active" db:"active"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`
}

type Link struct {
	Issuer                string    `json:"issuer" db:"issuer"`          // PK
	ExternalID            string    `json:"externalId" db:"external_id"` // PK
	UserID                string    `json:"userId" db:"user_id"`
	AccessToken           string    `json:"-" db:"access_token"`
	RefreshToken          string    `json:"-" db:"refresh_token"`
	AccessTokenExpiredAt  time.Time `json:"-" db:"access_token_expired_at"`
	RefreshTokenExpiredAt time.Time `json:"-" db:"refresh_token_expired_at"`
	CreatedAt             time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time `json:"updatedAt" db:"updated_at"`
}

const (
	TableUsers           = "users"
	ColDisplayName       = "display_name"
	ColEmail             = "email"
	ColEmailVerified     = "email_verified"
	ColPassword          = "password"
	ColPIN               = "pin"
	ColAvatarURL         = "avatar_url"
	ColPreferredLanguage = "preferred_language"
	ColActive            = "active"

	TableLinks               = "links"
	ColUserID                = "user_id"
	ColIssuer                = "issuer"
	ColExternalID            = "external_id"
	ColAccessToken           = "access_token"
	ColRefreshToken          = "refresh_token"
	ColAccessTokenExpiresAt  = "access_token_expires_at"
	ColRefreshTokenExpiresAt = "refresh_token_expires_at"
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

	LinkAllCols = []string{
		ColUserID,
		ColIssuer,
		ColExternalID,
		ColAccessToken,
		ColRefreshToken,
		ColCreatedAt,
		ColUpdatedAt,
		ColAccessTokenExpiresAt,
		ColRefreshTokenExpiresAt,
	}
)
