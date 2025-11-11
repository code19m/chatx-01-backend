package val

import (
	"errors"
	"regexp"
)

const (
	emailRegex = `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
)

var (
	ErrInvalidEmail = errors.New("must be a valid email address.")
)

func ValidateEmail(email string) error {
	if regexp.MustCompile(emailRegex).MatchString(email) {
		return nil
	}

	return ErrInvalidEmail
}
