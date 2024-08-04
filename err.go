package cli

type NamespaceNotFoundError struct {
	message      string
	alternatives []string
}

func NamespaceNotFound(message string, alternatives []string) ErrorWithAlternatives {
	if message == "" {
		message = "Namespace not found."
	}

	return &NamespaceNotFoundError{
		message:      message,
		alternatives: alternatives,
	}
}

func (e *NamespaceNotFoundError) Error() string {
	return e.message
}

func (e *NamespaceNotFoundError) Alternatives() []string {
	return e.alternatives
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
