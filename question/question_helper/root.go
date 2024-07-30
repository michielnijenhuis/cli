package question_helper

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/michielnijenhuis/cli/formatter"
	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/input"
	"github.com/michielnijenhuis/cli/output"
	"github.com/michielnijenhuis/cli/question"
	"github.com/michielnijenhuis/cli/terminal"
	"github.com/michielnijenhuis/cli/types"
)

func Ask[T any](i input.InputInterface, o output.OutputInterface, question types.QuestionInterface[T]) (T, error) {
	consoleOutput, isConsoleOutput := o.(output.ConsoleOutputInterface)
	if isConsoleOutput {
		o = consoleOutput.ErrorOutput()
	}

	if !i.IsInteractive() {
		return cast[T](defaultAnswer[T](question).(T)), nil
	}

	value, err := doAsk(o, question)
	if err != nil {
		i.SetInteractive(false)
		fallbackOutput := defaultAnswer[T](question)

		str, ok := any(fallbackOutput).(string)
		if ok {
			if str == "" {
				return cast[T](str), err
			}

			return cast[T](str), err
		}

		if fallbackOutput == nil {
			var empty T
			return empty, err
		}

		return cast[T](fallbackOutput), err
	} else {
		return value, nil
	}
}

func cast[T any](value any) T {
	casted, matches := value.(T)
	if matches {
		return casted
	}

	var empty T
	return empty
}

func doAsk[T any](o output.OutputInterface, question types.QuestionInterface[T]) (T, error) {
	writePrompt[T](o, question)

	var input string
	var ret T
	var err error

	if terminal.IsInteractive() {
		attempts := question.MaxAttempts()

		for input == "" && attempts > 0 {
			attempts--
			fmt.Println("> ")

			_, err = fmt.Scanln(&input)
			if err != nil {
				writeError(o, err)
				err = nil
			}
		}

		if input == "" {
			var empty T
			return empty, errors.New("missing input")
		}
	}

	if err != nil {
		var empty T
		return empty, err
	}

	consoleSectionOutput, ok := o.(*output.ConsoleSectionOutput)
	if ok {
		consoleSectionOutput.AddContent("", true)
		consoleSectionOutput.AddContent(input, true)
	}

	if question.IsTrimmable() {
		input = strings.TrimSpace(input)
	}

	if len(input) == 0 {
		ret = question.Default()
	}

	normalizer := question.Normalizer()
	if normalizer != nil {
		ret = normalizer(input)
	}

	validator := question.Validator()
	if validator != nil {
		validated, err := validator(ret)
		if err != nil {
			var empty T
			return empty, err
		}
		ret = validated
	}

	return ret, nil
}

func defaultAnswer[T any](qs interface{}) any {
	q, ok := qs.(types.QuestionInterface[T])
	if !ok {
		var empty T
		return empty
	}

	defaultValue := q.Default()

	validator := q.Validator()
	if validator != nil {
		df, _ := validator(defaultValue)
		return df
	}

	choiceQuestion, ok := qs.(*question.ChoiceQuestion)
	if !ok {
		return defaultValue
	}

	choices := choiceQuestion.Choices()
	str, ok := any(defaultValue).(string)
	if ok {
		return choices[str]
	}

	return defaultValue
}

func writePrompt[T any](output output.OutputInterface, qs interface{}) {
	q, ok := qs.(types.QuestionInterface[T])
	if !ok {
		return
	}

	text := formatter.EscapeTrailingBackslash(q.GetQuestion())

	if q.IsMultiline() {
		text += fmt.Sprintf(" (press %s to continue)", eofShortcut())
	}

	if str, ok := any(q.Default()).(string); ok && str == "" {
		text = fmt.Sprintf(" <info>%s</info>", text)
	} else if cq, ok := qs.(*question.ConfirmationQuestion); ok {
		highlight := "yes"
		if !cq.Default() {
			highlight = "no"
		}

		text = fmt.Sprintf(" <info>%s (yes/no)</info> [<highlight>%s</highlight>]", text, highlight)
	} else if cq, ok := qs.(*question.ChoiceQuestion); ok {
		choices := cq.Choices()
		str, isStr := any(q.Default()).(string)
		comment := str
		if isStr {
			val, exists := choices[str]
			if exists {
				comment = val
			}
		}
		text = fmt.Sprintf(" <info>%s</info> [<comment>%s</comment>]", text, comment)
	}

	output.Writeln(text, 0)

	prompt := " > "

	choice, ok := qs.(*question.ChoiceQuestion)
	if ok {
		output.Writelns(formatChoiceQuestionChoices(choice, "comment"), 0)
		prompt = choice.Prompt()
	}

	output.Write(prompt, false, 0)
}

func formatChoiceQuestionChoices(question *question.ChoiceQuestion, tag string) []string {
	messages := make([]string, 0)
	choices := question.Choices()

	var maxWidth int
	for key := range choices {
		maxWidth = max(maxWidth, helper.Width(key))
	}

	for key, value := range choices {
		padding := strings.Repeat(" ", maxWidth-helper.Width(key))
		message := fmt.Sprintf("  [<%s>%s%s</%s>] %s", tag, key, padding, tag, value)
		messages = append(messages, message)
	}

	return messages
}

func writeError(output output.OutputInterface, err error) {
	style, ok := output.(types.StyleInterface)
	if ok {
		style.NewLine(1)
		style.Err([]string{err.Error()})
		return
	}

	message := formatter.FormatBlock([]string{err.Error()}, "error", false)

	output.Writeln(message, 0)
}

func eofShortcut() string {
	if runtime.GOOS == "windows" {
		return "<comment>Ctrl+Z</comment> then <comment>Enter</comment>"
	}

	return "<comment>Ctrl+D</comment>"
}
