package cli

type ErrorWithAlternatives interface {
	error
	Alternatives() []string
}

type CommandNotFoundError struct {
	message      string
	alternatives []string
}

func CommandNotFound(message string, alternatives []string) ErrorWithAlternatives {
	if message == "" {
		message = "Command not found."
	}

	return &CommandNotFoundError{
		message:      message,
		alternatives: alternatives,
	}
}

func (e *CommandNotFoundError) Error() string {
	return e.message
}

func (e *CommandNotFoundError) Alternatives() []string {
	return e.alternatives
}
