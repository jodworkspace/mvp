package domain

import "errors"

var (
	InternalErr       = errors.New("internal error")
	UserNotFoundErr   = errors.New("user not found")
	TokenExpiredErr   = errors.New("token is expired")
	TokenMalformedErr = errors.New("token is malformed")
)

const (
	TableUser   = "users"
	ColUserID   = "user_id"
	ColUsername = "username"
	ColFullName = "full_name"
	ColEmail    = "email"

	ColCreatedAt = "created_at"
	ColUpdatedAt = "updated_at"
)

var (
	UserPublicCol = []string{
		ColUserID,
		ColUsername,
		ColFullName,
		ColEmail,
	}
)
