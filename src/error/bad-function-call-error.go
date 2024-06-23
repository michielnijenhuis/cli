package error

type BadFunctionCallError struct {
	message string
}

func NewBadFunctionCallError(message string) BadFunctionCallError {
	if message == "" {
		message = "Bad function call."
	}

	err := BadFunctionCallError{
		message: message,
	}

	return err
}

func (e BadFunctionCallError) Error() string {
	return e.message
}
