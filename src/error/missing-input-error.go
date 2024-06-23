package error

type MissingInputError struct {
	message string
}

func NewMissingInputError(message string) MissingInputError {
	if message == "" {
		message = "Missing input."
	}

	err := MissingInputError{
		message: message,
	}

	return err
}

func (e MissingInputError) Error() string {
	return e.message
}
