package application

import (
	"github.com/michielnijenhuis/cli/types"
)

type NamespaceNotFoundError struct {
	message      string
	alternatives []string
}

func NamespaceNotFound(message string, alternatives []string) types.ErrorWithAlternatives {
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
