package exception

import "errors"

var (
	ErrInternalServer = errors.New("internal server error")

	ErrExpiredToken   = errors.New("expired token")
	ErrMalformedToken = errors.New("malformed token")
	ErrUserNotFound   = errors.New("user not found")
)

func handleHTTPError(err error) {}
