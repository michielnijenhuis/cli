package question

import (
	"fmt"
	"strings"
)

type ChoiceQuestion struct {
	*Question[string]
	prompt       string
	errorMessage string
	choices      map[string]string
}

func NewChoiceQuestion(question string, choices map[string]string, defaultValue string) *ChoiceQuestion {
	q := NewQuestion[string](question, defaultValue)
	if len(choices) == 0 {
		panic("Choice question must have at least 1 choice available.")
	}

	cq := &ChoiceQuestion{
		Question:     q,
		prompt:       " > ",
		errorMessage: `Value "%s" is invalid`,
		choices:      choices,
	}

	cq.SetValidator(cq.defaultValidator())

	return cq
}

func (cq *ChoiceQuestion) Choices() map[string]string {
	return cq.choices
}

func (cq *ChoiceQuestion) Prompt() string {
	return cq.prompt
}

func (cq *ChoiceQuestion) SetPrompt(prompt string) {
	cq.prompt = prompt
}

func (cq *ChoiceQuestion) SetErrorMessage(message string) {
	cq.errorMessage = message
	cq.SetValidator(cq.defaultValidator())
}

func (cq *ChoiceQuestion) defaultValidator() QuestionValidator[string] {
	choices := cq.Choices()
	errorMessage := cq.errorMessage

	return func(selected string) string {
		if cq.IsTrimmable() {
			selected = strings.TrimSpace(selected)
		}

		results := make([]string, 0)
		for key, choice := range choices {
			if choice == selected {
				results = append(results, key)
			}
		}

		if len(results) > 1 {
			// TODO: return error
			panic(fmt.Sprintf("The provided answer is ambiguous. Value should be one of \"%s\".", strings.Join(results, "\" or \"")))
		}

		var result string
		for key, choice := range choices {
			if choice == selected {
				result = key
				break
			}
		}

		if result == "" {
			if _, exists := choices[selected]; exists {
				result = selected
			}
		}

		if result == "" {
			// TODO: return error
			panic(fmt.Sprintf(errorMessage, selected))
		}

		return result
	}
}
