package errs

type ValidationError struct {
	Message string
	Fields  map[string]string // nil for non-field errors
}

func NewValidationError(message string) error {
	return ValidationError{
		Message: message,
	}
}

// Error returns the reason for the validation error.
func (v ValidationError) Error() string {
	return v.Message
}

func AddFieldError(err error, field string, message string) error {
	validationError, ok := err.(ValidationError)
	if !ok {
		validationError = ValidationError{
			Message: err.Error(),
			Fields:  make(map[string]string),
		}
	}

	if validationError.Fields == nil {
		validationError.Fields = make(map[string]string)
	}

	validationError.Fields[field] = message
	return validationError
}
