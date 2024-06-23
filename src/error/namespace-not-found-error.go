package error

type NamespaceNotFoundError struct {
	message string
}

func NewNamespaceNotFoundError(message string) NamespaceNotFoundError {
	if message == "" {
		message = "Namespace not found."
	}

	err := NamespaceNotFoundError{
		message: message,
	}

	return err
}

func (e NamespaceNotFoundError) Error() string {
	return e.message
}
