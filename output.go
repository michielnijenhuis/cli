package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/internal/helper"
	"github.com/michielnijenhuis/cli/internal/terminal"
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

	width := terminal.Columns()
	if width == 0 {
		width = maxLineLength
	}

	o.lineLength = min(width, maxLineLength)

	return o
}

func NewOutput(input *Input) *Output {
	f := &OutputFormatter{}
	o := setupNewOutput(input, os.Stdout, f)
	o.Stderr = setupNewOutput(input, os.Stderr, f)
	return o
}

func (o *Output) Formatter() *OutputFormatter {
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

func (o *Output) Writelnf(f string, options uint, args ...any) {
	o.Writelns([]string{fmt.Sprintf(f, args...)}, options)
}

func (o *Output) Writelns(s []string, options uint) {
	o.WriteMany(s, true, options)
}

func (o *Output) Write(message string, newLine bool, options uint) {
	o.WriteMany([]string{message}, newLine, options)
}

func (o *Output) Format(message string) string {
	return o.Formatter().Format(message)
}

func (o *Output) Formatf(format string, args ...any) string {
	return o.Formatter().Format(fmt.Sprintf(format, args...))
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

	errs := make([]error, 0)
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

	if len(errs) > 0 && o.IsDebug() {
		log.Fatalln(errors.Join(errs...))
	}
}

func (o *Output) DoWrite(message string, newLine bool) {
	if newLine {
		message += Eol
	}

	_, err := o.Stream.WriteString(message)
	if err != nil && o.IsDebug() {
		log.Fatalln(err)
	}
}

func (o *Output) SetDecorated(decorated bool) {
	doSetDecorated(o, decorated)
	doSetDecorated(o.Stderr, decorated)

	if o.Formatter() != nil {
		o.Formatter().Decorated = decorated
	}
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

func (o *Output) Block(messages []string, tag string, escape bool) {
	theme, _ := GetTheme(tag)

	if theme.Padding {
		o.autoPrependBlock()
	}

	o.Writelns(o.CreateBlock(messages, tag, theme, escape), 0)

	if theme.Padding {
		o.NewLine(1)
	}
}

func (o *Output) CreateBlock(messages []string, tag string, theme *Theme, escape bool) []string {
	prefix := ""
	if theme != nil && theme.Label != "" {
		prefix += theme.Label
	}

	indentLength := 0
	prefixLength := helper.Width(o.Formatter().RemoveDecoration(prefix))
	lines := make([]string, 0, len(messages))
	lineIndentation := ""

	if prefix != "" {
		indentLength = helper.Width(prefix) + 2
		lineIndentation = strings.Repeat(" ", indentLength)
	}

	for i, message := range messages {
		if escape {
			message = Escape(message)
		}

		wrapped := helper.Wrap(message, o.lineLength-prefixLength-indentLength, Eol, false)
		lines = append(lines, strings.Split(wrapped, Eol)...)

		if len(messages) > 1 && i < len(messages)-1 {
			lines = append(lines, "")
		}
	}

	firstLineIndex := 0
	if theme != nil && theme.Padding && o.IsDecorated() {
		firstLineIndex = 1
		helper.Unshift(&lines, "")
		lines = append(lines, "")
	}

	for i, line := range lines {
		if prefix != "" {
			if firstLineIndex == i {
				if tag != "" && theme != nil && !theme.FullyColored {
					prefix = fmt.Sprintf("<%s>%s</>", tag, prefix)
				}

				line = prefix + line
			} else {
				line = lineIndentation + line
			}
		}

		line += strings.Repeat(" ", max(o.lineLength-helper.Width(o.Formatter().RemoveDecoration(line)), 0))

		if tag != "" && theme.FullyColored {
			line = fmt.Sprintf("<%s>%s</>", tag, line)
		}

		lines[i] = line
	}

	return lines
}

func (o *Output) NewLine(count int) {
	for count > 0 {
		o.Writeln("", 0)
		count--
	}
}

func (o *Output) autoPrependBlock() {
	chars := o.bufferedOutput.Fetch()

	if len(chars) > 2 {
		chars = chars[:len(chars)-2]
	}

	if chars == "" {
		o.NewLine(1)
		return
	}

	var lineBreakCount int
	for i := 0; i < len(chars); i++ {
		char := chars[i]
		if char == '\n' {
			lineBreakCount++
		}
	}

	o.NewLine(2 - lineBreakCount)
}

func (o *Output) PrintByTheme(themeName string, v any, escape bool) {
	theme, err := GetTheme(themeName)
	if err != nil {
		switch v := v.(type) {
		case error:
			o.Writelnf("%s", OutputNormal, v.Error())
		case string:
			o.Writelnf("%s", OutputNormal, v)
		case []string:
			o.Writelns(v, OutputNormal)
		case nil:
			return
		default:
			o.Writelnf("%v", OutputNormal, v)
		}

		return
	}

	o.autoPrependBlock()

	prefix := theme.GetStyle().Apply(theme.Label) + " "
	indentLength := 0
	prefixLength := helper.Width(theme.Label + " ")
	lines := make([]string, 0)
	lineIndentation := ""

	if prefix != "" {
		indentLength = helper.Width(theme.Label + " ")
		lineIndentation = strings.Repeat(" ", indentLength)
	}

	messages := make([]string, 0)
	switch v := v.(type) {
	case error:
		messages = append(messages, v.Error())
	case string:
		messages = append(messages, v)
	case []string:
		messages = v
	default:
		messages = append(messages, fmt.Sprintf("%v", v))
	}

	for i, message := range messages {
		if escape {
			message = Escape(message)
		}

		width := o.lineLength - prefixLength - indentLength
		wrapped := helper.Wrap(message, width, Eol, false)
		lines = append(lines, strings.Split(wrapped, Eol)...)

		if len(messages) > 1 && i < len(messages)-1 {
			lines = append(lines, "")
		}
	}

	for i, line := range lines {
		if prefix != "" {
			if i == 0 {
				line = prefix + line
			} else {
				line = lineIndentation + line
			}
		}

		spaces := max(o.lineLength-helper.Width(o.Formatter().RemoveDecoration(line)), 0)
		line += strings.Repeat(" ", spaces)
		lines[i] = line
	}

	o.Writelns(lines, OutputNormal)
}
