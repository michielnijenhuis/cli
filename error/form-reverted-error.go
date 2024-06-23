package error

type FormRevertedError struct {
	message string
}

func NewFormRevertedError(message string) FormRevertedError {
	if message == "" {
		message = "Form reverted error."
	}

	err := FormRevertedError{
		message: message,
	}

	return err
}

func (e FormRevertedError) Error() string {
	return e.message
}
