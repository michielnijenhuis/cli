package error

type CommandNotFoundError struct {
	message      string
	alternatives []string
}

func NewCommandNotFoundError(message string, alternatives []string) CommandNotFoundError {
	if message == "" {
		message = "Command not found."
	}

	return CommandNotFoundError{
		message:      message,
		alternatives: alternatives,
	}
}

func (e CommandNotFoundError) Error() string {
	return e.message
}

func (e CommandNotFoundError) GetAlternatives() []string {
	return e.alternatives
}
