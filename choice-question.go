package cli

import (
	"fmt"
	"strings"
)

type ChoiceQuestion struct {
	*Question[string]
	Prompt       string
	errorMessage string
	Choices      map[string]string
}

const ChoiceQuestionDefaultPrompt = " > "
const ChoiceQuestionDefaultErrorMessage = `Value "%s" is invalid`

func (cq *ChoiceQuestion) SetErrorMessage(message string) {
	cq.errorMessage = message
	cq.Validator = cq.DefaultValidator()
}

func (cq *ChoiceQuestion) DefaultValidator() QuestionValidator[string] {
	choices := cq.Choices
	checkPtr(choices, "choice question choices")
	errorMessage := cq.errorMessage

	return func(selected string) (string, error) {
		if !cq.PreventTrimming {
			selected = strings.TrimSpace(selected)
		}

		results := make([]string, 0)
		for key, choice := range choices {
			if choice == selected {
				results = append(results, key)
			}
		}

		if len(results) > 1 {
			return "", fmt.Errorf("the provided answer is ambiguous. Value should be one of \"%s\"", strings.Join(results, "\" or \""))
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
			return "", fmt.Errorf(errorMessage, selected)
		}

		return result, nil
	}
}
