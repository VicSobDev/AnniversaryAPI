package auth

import "errors"

var (
	errInvalidPassword = errors.New("invalid password")
	errUserNotFound    = errors.New("user not found")
)
