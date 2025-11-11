package domain

import "errors"

// Domain-specific errors for auth module.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrIncorrectPassword  = errors.New("incorrect password")
)
