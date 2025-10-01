package domain

import "errors"

const (
	ColID        = "id"
	ColCreatedAt = "created_at"
	ColUpdatedAt = "updated_at"

	ProviderGoogle = "google" // Google Drive Storage
	ProviderGitHub = "github" // GitHub Repository Storage

	KeyPagination     = "pagination"
	KeyUserID         = "user_id"
	KeyIssuer         = "issuer"
	KeyDriverID       = "driver_id"
	KeyAccessToken    = "access_token"
	KeyRefreshToken   = "refresh_token"
	SessionCookieName = "sid"

	FileTypeFolder = "application/vnd.google-apps.folder"
)

var (
	InternalServerError = errors.New("internal server error")
)
