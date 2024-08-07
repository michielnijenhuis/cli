package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

func getQuestion[T any](val any) *Question[T] {
	rv := reflect.ValueOf(val)

	if rv.Kind() != reflect.Ptr {
		return nil
	}

	rv = rv.Elem()

	if rv.Kind() != reflect.Struct {
		return nil
	}

	if rv.Type() == reflect.TypeOf(Question[T]{}) {
		return val.(*Question[T])
	}

	q := Question[T]{}
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)

		if field.Kind() != reflect.Ptr {
			continue
		}

		field = field.Elem()

		if field.Type() == reflect.TypeOf(q) {
			addr := field.Addr()
			iface := addr.Interface()
			ptr := iface.(*Question[T])
			return ptr
		}
	}

	return nil
}

type QuestionInterface interface {
	IsQuestion() bool
}

func Ask[T any](i *Input, o *Output, question QuestionInterface) (T, error) {
	checkPtr(i, "question input")
	checkPtr(o, "question output")
	o = o.Stderr
	checkPtr(o, "output stderr")

	if !i.IsInteractive() {
		q := getQuestion[T](question)
		checkPtr(q, "question")
		return q.DefaultValue, nil
	}

	inputStream := i.Stream

	value, err := doAsk[T](o, question, inputStream)
	if err != nil {
		i.SetInteractive(false)
		fallbackOutput := defaultAnswer[T](question)

		return fallbackOutput, err
	} else {
		return value, nil
	}
}

func normalizeAnswer(answer string, question QuestionInterface) any {
	if cq, ok := question.(*ConfirmationQuestion); ok {
		n := cq.Normalizer
		if n == nil {
			n = cq.DefaultNormalizer()
		}
		if n != nil {
			return n(answer)
		}
	} else if chq, ok := question.(*ChoiceQuestion); ok {
		n := chq.DefaultNormalizer()
		if n != nil {
			return n(answer)
		}
	}

	return answer
}

func validateAnswer[T any](answer T, question QuestionInterface) (T, error) {
	if cq, ok := question.(*ConfirmationQuestion); ok {
		v := cq.Validator
		if v == nil {
			v = cq.DefaultValidator()
		}

		if v != nil {
			typed, matches := any(answer).(bool)
			if matches {
				validated, err := v(typed)
				typed := any(validated).(T)
				return typed, err
			}
		}

		return answer, nil
	} else if chq, ok := question.(*ChoiceQuestion); ok {
		v := chq.Validator
		if v == nil {
			v = chq.DefaultValidator()
		}

		if v != nil {
			typed, matches := any(answer).(string)
			if matches {
				validated, err := v(typed)
				typed := any(validated).(T)
				return typed, err
			}
		}

		return answer, nil
	}

	return answer, nil
}

func doAsk[T any](o *Output, question QuestionInterface, inputStream *os.File) (T, error) {
	q := getQuestion[T](question)
	checkPtr(q, "question")

	writePrompt[T](o, question)

	var input string
	var ret T
	var err error

	if TerminalIsInteractive() {
		if inputStream == nil {
			inputStream = os.Stdin
		}

		reader := bufio.NewReader(inputStream)
		attempts := q.Attempts
		hasMaxAttempts := attempts > 0

		for input == "" && (!hasMaxAttempts || attempts > 0) {
			attempts--

			input, err = reader.ReadString('\n')
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

	if !q.PreventTrimming {
		input = strings.TrimSpace(input)
	}

	if len(input) == 0 {
		ret = q.DefaultValue
	}

	normalized := normalizeAnswer(input, question)
	if normalized != input {
		typed, ok := normalized.(T)
		if ok {
			ret = typed
		}
	}

	validated, err := validateAnswer(ret, question)
	if err != nil {
		var empty T
		return empty, err
	}

	return validated, nil
}

func defaultAnswer[T any](qs QuestionInterface) T {
	q := getQuestion[T](qs)
	checkPtr(q, "question")

	defaultValue := q.DefaultValue

	validator := q.Validator
	if validator == nil {
		validator = q.DefaultValidator()
	}

	if validator != nil {
		validated, _ := validator(defaultValue)
		return validated
	}

	return defaultValue
}

func writePrompt[T any](output *Output, qs any) {
	var query string
	var multiline bool
	var defaultValue any

	if q, ok := qs.(*Question[T]); ok {
		query = q.Query
		multiline = q.Multiline
		defaultValue = q.DefaultValue
	} else if q, ok := qs.(*ConfirmationQuestion); ok {
		query = q.Query
		multiline = q.Multiline
		defaultValue = q.DefaultValue
	} else if q, ok := qs.(*ChoiceQuestion); ok {
		query = q.Query
		multiline = q.Multiline
		defaultValue = q.DefaultValue
	} else {
		panic("unknown question struct")
	}

	text := EscapeTrailingBackslash(query)

	if multiline {
		text += fmt.Sprintf(" (press %s to continue)", eofShortcut())
	}

	if str, ok := defaultValue.(string); ok && str == "" {
		text = fmt.Sprintf(" <info>%s</info>", text)
	} else if cq, ok := qs.(*ConfirmationQuestion); ok {
		highlight := "yes"
		if !cq.DefaultValue {
			highlight = "no"
		}

		text = fmt.Sprintf(" %s (yes/no) [<accent>%s</accent>]", text, highlight)
	} else if cq, ok := qs.(*ChoiceQuestion); ok {
		str, isStr := defaultValue.(string)
		comment := str
		if isStr {
			val, exists := cq.Choices[str]
			if exists {
				comment = val
			}
		}
		text = fmt.Sprintf(" <info>%s</info> [<comment>%s</comment>]", text, comment)
	}

	output.Writeln(text, 0)

	prompt := " > "

	choice, ok := qs.(*ChoiceQuestion)
	if ok {
		output.Writelns(formatChoiceQuestionChoices(choice, "comment"), 0)
		choicePrompt := choice.Prompt
		if choicePrompt == "" {
			choicePrompt = ChoiceQuestionDefaultPrompt
		}
		prompt = choicePrompt
	}

	output.Write(prompt, false, 0)
}

func formatChoiceQuestionChoices(question *ChoiceQuestion, tag string) []string {
	messages := make([]string, 0)

	var maxWidth int
	for key := range question.Choices {
		maxWidth = max(maxWidth, helper.Width(key))
	}

	for key, value := range question.Choices {
		padding := strings.Repeat(" ", maxWidth-helper.Width(key))
		message := fmt.Sprintf("  [<%s>%s%s</%s>] %s", tag, key, padding, tag, value)
		messages = append(messages, message)
	}

	return messages
}

func writeError(output *Output, err error) {
	output.NewLine(1)
	output.Err(err.Error())
}

func eofShortcut() string {
	if runtime.GOOS == "windows" {
		return "<comment>Ctrl+Z</comment> then <comment>Enter</comment>"
	}

	return "<comment>Ctrl+D</comment>"
}
