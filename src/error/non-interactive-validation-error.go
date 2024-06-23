package error

type NonInteractiveValidationError struct {
	message string
}

func NewNonInteractiveValidationError(message string) NonInteractiveValidationError {
	if message == "" {
		message = "Validation error."
	}

	err := NonInteractiveValidationError{
		message: message,
	}

	return err
}

func (e NonInteractiveValidationError) Error() string {
	return e.message
}
