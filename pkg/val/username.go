package val

import "errors"

var (
	ErrInvalidUsername = errors.New(
		"must be between 3 and 20 characters long and can only contain letters, numbers, underscores, and hyphens.",
	)
)

func ValidateUsername(username string) error {
	if !isValidUsername(username) {
		return ErrInvalidUsername
	}

	return nil
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}

	return true
}
