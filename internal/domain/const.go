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
	KeyAccessToken    = "access_token"
	KeyRefreshToken   = "refresh_token"
	SessionCookieName = "sid"

	FileTypeFolder = "folder"
	FileTypeFile   = "file"
	MimeTypeFolder = "application/vnd.google-apps.folder"
	MimeTypeFile   = "application/vnd.google-apps.file"
)

var (
	InternalServerError = errors.New("internal server error")
	InvalidSessionError = errors.New("invalid session")
)
