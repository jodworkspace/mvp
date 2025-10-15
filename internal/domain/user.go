package domain

import (
	"time"
)

type User struct {
	ID                string    `json:"id"`
	DisplayName       string    `json:"displayName"`
	Email             string    `json:"email"`
	EmailVerified     bool      `json:"emailVerified"`
	Password          string    `json:"password"`
	PIN               string    `json:"pin"`
	AvatarURL         string    `json:"avatarUrl"`
	PreferredLanguage string    `json:"preferredLanguage"`
	Active            bool      `json:"active"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Link struct {
	Issuer                string    `json:"issuer"`     // PK
	ExternalID            string    `json:"externalId"` // PK
	UserID                string    `json:"userId"`
	AccessToken           string    `json:"-"`
	RefreshToken          string    `json:"-"`
	AccessTokenExpiredAt  time.Time `json:"-"`
	RefreshTokenExpiredAt time.Time `json:"-" `
	CreatedAt             time.Time `json:"createdAt" `
	UpdatedAt             time.Time `json:"updatedAt" `
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
