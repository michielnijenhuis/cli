package error

type InvalidOptionError struct {
	message string
}

func NewInvalidOptionError(message string) InvalidOptionError {
	if message == "" {
		message = "Invalid option."
	}

	err := InvalidOptionError{
		message: message,
	}

	return err
}

func (e InvalidOptionError) Error() string {
	return e.message
}
