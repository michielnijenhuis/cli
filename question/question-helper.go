package question

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	Err "github.com/michielnijenhuis/cli/error"
	Formatter "github.com/michielnijenhuis/cli/formatter"
	Helper "github.com/michielnijenhuis/cli/helper"
	Input "github.com/michielnijenhuis/cli/input"
	Output "github.com/michielnijenhuis/cli/output"
	Style "github.com/michielnijenhuis/cli/style"
	Terminal "github.com/michielnijenhuis/cli/terminal"
)

func Ask[T any](input Input.InputInterface, output Output.OutputInterface, question QuestionInterface[T]) (T, error) {
	consoleOutput, isConsoleOutput := output.(Output.ConsoleOutputInterface)
	if isConsoleOutput {
		output = consoleOutput.GetErrorOutput()
	}

	if !input.IsInteractive() {
		return cast[T](defaultAnswer[T](question).(T)), nil
	}

	var inputStream *os.File
	streamableInput, ok := input.(Input.StreamableInputInterface)
	if ok {
		stream := streamableInput.GetStream()
		if stream != nil {
			inputStream = stream
		}
	}

	value, err := doAsk(output, question, inputStream)
	if err != nil {
		_, ok := err.(Err.MissingInputError)
		if !ok {
			return value, err
		}

		input.SetInteractive(false)
		fallbackOutput := defaultAnswer[T](question)

		str, ok := any(fallbackOutput).(string)
		if ok {
			if str == "" {
				return cast[T](str), err
			}

			return cast[T](str), nil
		}

		if fallbackOutput == nil {
			var empty T
			return empty, err
		}

		return cast[T](fallbackOutput), nil
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

type QuestionInterface[T any] interface {
	Default() T
	Normalizer() QuestionNormalizer[T]
}

// TODO: implement
func doAsk[T any](output Output.OutputInterface, question QuestionInterface[T], inputStream *os.File) (T, error) {
	writePrompt[T](output, question)

	if inputStream == nil {
		inputStream = os.Stdin
	}

	var ret T

	if Terminal.IsInteractive() {
		// outputStream := os.Stdout
		// streamOutput, ok := output.(*Output.StreamOutput)
		// if ok {
		// 	outputStream = streamOutput.GetStream()
		// }

		// TODO: do ask
	}

	str := any(ret).(string)
	consoleSectionOutput, ok := output.(*Output.ConsoleSectionOutput)
	if ok {
		consoleSectionOutput.AddContent("", true)
		consoleSectionOutput.AddContent(str, true)
	}

	if len(str) == 0 {
		ret = question.Default()
	}

	normalizer := question.Normalizer()
	if normalizer != nil {
		ret = normalizer(str)
	}

	return ret, nil
}

func defaultAnswer[T any](question interface{}) any {
	q, ok := question.(*Question[T])
	if !ok {
		var empty T
		return empty
	}

	defaultValue := q.Default()

	validator := q.Validator()
	if validator != nil {
		return validator(defaultValue)
	}

	choiceQuestion, ok := question.(*ChoiceQuestion)
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

func writePrompt[T any](output Output.OutputInterface, question interface{}) {
	q, ok := question.(*Question[T])
	if !ok {
		return
	}

	text := Formatter.EscapeTrailingBackslash(q.Question())

	if q.IsMultiline() {
		text += fmt.Sprintf(" (press %s to continue)", getEofShortcut())
	}

	if str, ok := any(q.Default()).(string); ok && str == "" {
		text = fmt.Sprintf(" <info>%s</info>", text)
	} else if cq, ok := question.(*ConfirmationQuestion); ok {
		highlight := "yes"
		if !cq.Default() {
			highlight = "no"
		}

		text = fmt.Sprintf(" <info>%s (yes/no)</info> [<highlight>%s</highlight>]", text, highlight)
	} else if cq, ok := question.(*ChoiceQuestion); ok {
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

	choice, ok := question.(*ChoiceQuestion)
	if ok {
		output.Writelns(formatChoiceQuestionChoices(choice, "comment"), 0)
		prompt = choice.Prompt()
	}

	output.Write(prompt, false, 0)
}

func formatChoiceQuestionChoices(question *ChoiceQuestion, tag string) []string {
	messages := make([]string, 0)
	choices := question.Choices()

	var maxWidth int
	for key := range choices {
		maxWidth = max(maxWidth, Helper.Width(key))
	}

	for key, value := range choices {
		padding := strings.Repeat(" ", maxWidth-Helper.Width(key))
		message := fmt.Sprintf("  [<%s>%s%s</%s>] %s", tag, key, padding, tag, value)
		messages = append(messages, message)
	}

	return messages
}

func writeError(output Output.OutputInterface, err error) {
	style, ok := output.(*Style.Style)
	if ok {
		style.NewLine(1)
		style.Err([]string{err.Error()})
		return
	}

	message := Formatter.FormatBlock([]string{err.Error()}, "error", false)

	output.Writeln(message, 0)
}

func getEofShortcut() string {
	if runtime.GOOS == "windows" {
		return "<comment>Ctrl+Z</comment> then <comment>Enter</comment>"
	}

	return "<comment>Ctrl+D</comment>"
}
