package error

type InvalidArgumentError struct {
	message string
}

func NewInvalidArgumentError(message string) InvalidArgumentError {
	if message == "" {
		message = "Invalid argument."
	}

	err := InvalidArgumentError{
		message: message,
	}

	return err
}

func (e InvalidArgumentError) Error() string {
	return e.message
}
