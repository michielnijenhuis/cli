package error

type ManuallyFailedError struct {
	message string
}

func NewManuallyFailedError(message string) ManuallyFailedError {
	if message == "" {
		message = "Manually failed command."
	}

	err := ManuallyFailedError{
		message: message,
	}

	return err
}

func (e ManuallyFailedError) Error() string {
	return e.message
}
