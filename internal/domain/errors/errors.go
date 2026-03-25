package errors

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrAppNotFound        = errors.New("app not found")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrTokenExpired = errors.New("token is expired")
	ErrInvalidToken = errors.New("invalid token")

	ErrInvalidCode     = errors.New("invalid reset code")
	ErrCodeExpired     = errors.New("reset code has expired")
	ErrCodeAlreadyUsed = errors.New("reset code was already used")
)
