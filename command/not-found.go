package command

import (
	"github.com/michielnijenhuis/cli/types"
)

type CommandNotFoundError struct {
	message      string
	alternatives []string
}

func NotFound(message string, alternatives []string) types.ErrorWithAlternatives {
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
