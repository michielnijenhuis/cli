package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

const maxLineLength = 120

type Output struct {
	Stream         *os.File
	Stderr         *Output
	verbosity      uint
	decorated      bool
	formatter      *OutputFormatter
	lineLength     int
	bufferedOutput *TrimmedBufferOutput
	input          *Input
	// progressBar    *ProgressBar
}

const (
	VerbosityQuiet       uint = 16
	VerbosityNormal      uint = 32
	VerbosityVerbose     uint = 64
	VerbosityVeryVerbose uint = 128
	VerbosityDebug       uint = 256
)

const (
	OutputNormal uint = 1
	OutputRaw    uint = 2
	OutputPlain  uint = 4
)

func setupNewOutput(input *Input, stream *os.File, formatter *OutputFormatter) *Output {
	o := &Output{
		Stream:     stream,
		verbosity:  VerbosityNormal,
		decorated:  HasColorSupport(),
		formatter:  formatter,
		lineLength: maxLineLength,
		input:      input,
		bufferedOutput: &TrimmedBufferOutput{
			Output: &Output{
				Stream:     stream,
				verbosity:  VerbosityNormal,
				decorated:  HasColorSupport(),
				formatter:  formatter,
				lineLength: maxLineLength,
				input:      input,
			},
		},
	}

	if formatter != nil {
		formatter.Decorated = o.decorated
	}

	width, _ := TerminalWidth()
	if width == 0 {
		width = maxLineLength
	}

	o.lineLength = min(width, maxLineLength)

	return o
}

func NewOutput(input *Input) *Output {
	f := &OutputFormatter{
		Styles: DefaultOutputTheme,
	}
	o := setupNewOutput(input, os.Stdout, f)
	o.Stderr = setupNewOutput(input, os.Stderr, f)
	return o
}

func (o *Output) Formatter() *OutputFormatter {
	checkPtr(o.formatter, "output formatter")
	return o.formatter
}

func (o *Output) IsDecorated() bool {
	return o.Formatter().Decorated
}

func (o *Output) Verbosity() uint {
	return o.verbosity
}

func (o *Output) IsQuiet() bool {
	return o.verbosity == VerbosityQuiet
}

func (o *Output) IsVerbose() bool {
	return o.verbosity == VerbosityVerbose
}

func (o *Output) IsVeryVerbose() bool {
	return o.verbosity == VerbosityVeryVerbose
}

func (o *Output) IsDebug() bool {
	return o.verbosity == VerbosityDebug
}

func (o *Output) Writeln(s string, options uint) {
	o.Writelns([]string{s}, options)
}

func (o *Output) Writelns(s []string, options uint) {
	o.WriteMany(s, true, options)
}

func (o *Output) Write(message string, newLine bool, options uint) {
	o.WriteMany([]string{message}, newLine, options)
}

func (o *Output) WriteMany(messages []string, newLine bool, options uint) {
	types := OutputNormal | OutputRaw | OutputPlain

	t := types & options
	if t == 0 {
		t = OutputNormal
	}

	verbosities := VerbosityQuiet | VerbosityNormal | VerbosityVerbose | VerbosityVeryVerbose | VerbosityDebug
	verbosity := verbosities & options
	if verbosity == 0 {
		verbosity = VerbosityNormal
	}

	if verbosity > o.Verbosity() {
		return
	}

	re := regexp.MustCompile("<[^>]+>")

	var message string
	for _, m := range messages {
		message = m
		switch t {
		case OutputNormal:
			message = o.Formatter().Format(message)
		case OutputPlain:
			message = re.ReplaceAllString(o.Formatter().Format(message), "")
		}

		o.DoWrite(message, newLine)
	}
}

func (o *Output) DoWrite(message string, newLine bool) {
	if newLine {
		message += "\n"
	}

	_, err := o.Stream.WriteString(message)
	if err != nil {
		panic(err)
	}
}

func (o *Output) SetDecorated(decorated bool) {
	doSetDecorated(o, decorated)
	doSetDecorated(o.Stderr, decorated)
}

func doSetDecorated(o *Output, decorated bool) {
	if o != nil {
		o.decorated = decorated
	}
}

func (o *Output) SetFormatter(formatter *OutputFormatter) {
	doSetFormatter(o, formatter)
	doSetFormatter(o.Stderr, formatter)
}

func doSetFormatter(o *Output, formatter *OutputFormatter) {
	if o != nil {
		o.formatter = formatter
	}
}

func (o *Output) SetVerbosity(verbose uint) {
	doSetVerbosity(o, verbose)
	doSetVerbosity(o.Stderr, verbose)
}

func doSetVerbosity(o *Output, verbose uint) {
	if o != nil {
		o.verbosity = verbose
	}
}

func (o *Output) Title(message string) {
	o.autoPrependBlock()
	messages := []string{
		fmt.Sprintf("<comment>%s</>", EscapeTrailingBackslash(message)),
		fmt.Sprintf("<comment>%s</>", strings.Repeat("=", helper.Width(o.Formatter().RemoveDecoration(message)))),
	}
	o.Writelns(messages, 0)
	o.NewLine(1)
}

func (o *Output) Section(message string) {
	o.autoPrependBlock()
	messages := []string{
		fmt.Sprintf("<comment>%s</>", EscapeTrailingBackslash(message)),
		fmt.Sprintf("<comment>%s</>", strings.Repeat("-", helper.Width(o.Formatter().RemoveDecoration(message)))),
	}
	o.Writelns(messages, 0)
	o.NewLine(1)
}

func (o *Output) Listing(elements []string, prefix string) {
	o.autoPrependBlock()
	els := make([]string, len(elements))
	for i, element := range elements {
		els[i] = fmt.Sprintf(" %s %s", prefix, element)
	}
	o.Writelns(els, 0)
	o.NewLine(1)
}

func (o *Output) Block(messages []string, tag string, style string, prefix string, padding bool, escape bool) {
	o.autoPrependBlock()
	o.Writelns(o.createBlock(messages, tag, style, prefix, padding, escape), 0)
	o.NewLine(1)
}

func (o *Output) createBlock(messages []string, tag string, style string, prefix string, padding bool, escape bool) []string {
	indentLength := 0
	prefixLength := helper.Width(o.Formatter().RemoveDecoration(prefix))
	lines := make([]string, 0, len(messages))
	lineIndentation := ""

	if tag != "" {
		tag = "[" + tag + "]"
		indentLength = helper.Width(tag)
		lineIndentation = strings.Repeat(" ", indentLength)
	}

	for i, message := range messages {
		if escape {
			message = Escape(message)
		}

		wrapped := helper.Wrap(message, o.lineLength-prefixLength-indentLength, "\n", false)
		lines = append(lines, strings.Split(wrapped, "\n")...)

		if len(messages) > 1 && i < len(messages)-1 {
			lines = append(lines, "")
		}
	}

	firstLineIndex := 0
	if padding && o.IsDecorated() {
		firstLineIndex = 1
		helper.Unshift(&lines, "")
		lines = append(lines, "")
	}

	for i, line := range lines {
		if tag != "" {
			if firstLineIndex == i {
				line = tag + line
			} else {
				line = lineIndentation + line
			}
		}

		line = prefix + line

		line += strings.Repeat(" ", max(o.lineLength-helper.Width(o.Formatter().RemoveDecoration(line)), 0))

		if style != "" {
			line = fmt.Sprintf("<%s>%s</>", style, line)
		}

		lines[i] = line
	}

	return lines
}

func (o *Output) Text(messages []string) {
	o.autoPrependBlock()
	for _, m := range messages {
		o.Writeln(" "+m, 0)
	}
}

func (o *Output) Success(messages []string) {
	o.Block(messages, "OK", "fg=black;bg=green", " ", true, true)
}

func (o *Output) Err(messages []string) {
	o.Block(messages, "ERROR", "fg=white;bg=red", " ", true, true)
}

func (o *Output) Warning(messages []string) {
	o.Block(messages, "WARNING", "fg=black;bg=yellow", " ", true, true)
}

func (o *Output) Info(messages []string) {
	o.Block(messages, "INFO", "fg=white;bg=blue", " ", true, true)
}

func (o *Output) Note(messages []string) {
	o.Block(messages, "NOTE", "fg=yellow", " ! ", true, true)
}

func (o *Output) Caution(messages []string) {
	o.Block(messages, "CAUTION", "fg=yellow;bg=red", " ! ", true, true)
}

func (o *Output) Comment(messages []string) {
	o.Block(messages, "", "", "<fg=default;bg=default> // </>", false, false)
}

// TODO: implement
func (o *Output) Table(headers []string, rows map[string]string) {}

func (o *Output) Ask(question string, defaultValue string, validator func(string) (string, error)) (string, error) {
	q := &Question[string]{
		Query:        question,
		DefaultValue: defaultValue,
		Validator:    validator,
	}

	return askQuestion[string](q, o.input, o)
}

func (o *Output) AskHidden(question string, validator func(string) (string, error)) (string, error) {
	q := &Question[string]{
		Query:     question,
		Hidden:    true,
		Validator: validator,
	}

	return askQuestion[string](q, o.input, o)
}

func (o *Output) Confirm(q string, defaultValue bool) (bool, error) {
	cq := &ConfirmationQuestion{
		Question: &Question[bool]{
			Query:        q,
			DefaultValue: defaultValue,
		},
	}

	return askQuestion[bool](cq, o.input, o)
}

func (o *Output) Choice(question string, choices map[string]string, defaultValue string) (string, error) {
	if defaultValue != "" {
		values := make(map[string]string)
		for k, v := range choices {
			values[v] = k
		}

		dv, ok := values[defaultValue]
		if ok {
			defaultValue = dv
		}
	}

	q := &ChoiceQuestion{
		Question: &Question[string]{
			Query:        question,
			DefaultValue: defaultValue,
		},
		Choices: choices,
	}

	return askQuestion[string](q, o.input, o)
}

func (o *Output) NewLine(count int) {
	for count > 0 {
		o.Writeln("", 0)
		count--
	}
}

// TODO: implement
func (o *Output) ProgressStart(max uint) {}

// TODO: implement
func (o *Output) ProgressAdvance(step uint) {}

// TODO: implement
func (o *Output) ProgressFinish() {}

// TODO: implement
func (o *Output) Box(title string, body string, footer string, color string, info string) {}

func askQuestion[T any](qi QuestionInterface, i *Input, o *Output) (T, error) {
	if i.IsInteractive() {
		o.autoPrependBlock()
	}

	answer, err := Ask[T](i, o, qi)
	if err != nil {
		var empty T
		return empty, nil
	}

	if i.IsInteractive() {
		o.NewLine(1)
		checkPtr(o.bufferedOutput, "output bufferedOutput")
		o.bufferedOutput.Write("\n", false, 0)
	}

	return answer, nil
}

func (o *Output) autoPrependBlock() {
	checkPtr(o.bufferedOutput, "output bufferedOutput")
	chars := o.bufferedOutput.Fetch()

	if len(chars) > 2 {
		chars = chars[:len(chars)-2]
	}

	if chars == "" {
		o.NewLine(1) // empty history, so we should start with a new line.
		return
	}

	var lineBreakCount int
	for i := 0; i < len(chars); i++ {
		char := chars[i]
		if char == '\n' {
			lineBreakCount++
		}
	}

	// Prepend new line for each non LF chars (This means no blank line was output before)
	o.NewLine(2 - lineBreakCount)
}
