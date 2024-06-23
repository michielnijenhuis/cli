package error

type RuntimeError struct {
	message string
}

func NewRuntimeError(message string) RuntimeError {
	if message == "" {
		message = "Unknown error."
	}

	err := RuntimeError{
		message: message,
	}

	return err
}

func (e RuntimeError) Error() string {
	return e.message
}
