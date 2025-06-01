package errorx

import "errors"

var (
	ErrInternalServer = errors.New("internal server error")

	ErrExpiredToken   = errors.New("expired token")
	ErrMalformedToken = errors.New("malformed token")
	ErrInvalidClaims  = errors.New("invalid claims")

	ErrInvalidProvider = errors.New("invalid provider")

	ErrUserNotFound = errors.New("user not found")
)

func handleHTTPError(err error) {}
