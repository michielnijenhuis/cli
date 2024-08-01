package style

import (
	"github.com/michielnijenhuis/cli/formatter"
	"github.com/michielnijenhuis/cli/input"
	"github.com/michielnijenhuis/cli/output"
	"github.com/michielnijenhuis/cli/question"
	"github.com/michielnijenhuis/cli/question/questionHelper"
	"github.com/michielnijenhuis/cli/terminal"
	"github.com/michielnijenhuis/cli/types"
)

const maxLineLength = 120

type Style struct {
	progressBar    *ProgressBar
	lineLength     int
	bufferedOutput *output.TrimmedBufferOutput
	input          input.InputInterface
	output         output.OutputInterface
}

func NewStyle(i input.InputInterface, o output.OutputInterface) *Style {
	s := &Style{
		progressBar:    nil,
		lineLength:     -1,
		bufferedOutput: output.NewTrimmedBufferOutput(2, o.Verbosity(), o.IsDecorated(), o.Formatter().Clone()),
		input:          i,
		output:         o,
	}

	width, _ := terminal.Width()
	if width == 0 {
		width = maxLineLength
	}

	s.lineLength = min(width, maxLineLength)

	return s
}

func (s *Style) SetFormatter(formatter formatter.OutputFormatterInferface) {
	s.output.SetFormatter(formatter)
}

func (s *Style) Formatter() formatter.OutputFormatterInferface {
	return s.output.Formatter()
}

func (s *Style) SetDecorated(decorated bool) {
	s.output.SetDecorated(decorated)
}

func (s *Style) IsDecorated() bool {
	return s.output.IsDecorated()
}

func (s *Style) SetVerbosity(level uint) {
	s.output.SetVerbosity(level)
}

func (s *Style) Verbosity() uint {
	return s.output.Verbosity()
}

func (s *Style) IsQuiet() bool {
	return s.output.IsQuiet()
}

func (s *Style) IsVerbose() bool {
	return s.output.IsVerbose()
}

func (s *Style) IsVeryVerbose() bool {
	return s.output.IsVeryVerbose()
}

func (s *Style) IsDebug() bool {
	return s.output.IsDebug()
}

func (s *Style) Writeln(message string, options uint) {
	s.output.Writeln(message, options)
}

func (s *Style) Writelns(messages []string, options uint) {
	s.output.Writelns(messages, options)
}

func (s *Style) Write(message string, newLine bool, options uint) {
	s.output.Write(message, newLine, options)
}

func (s *Style) WriteMany(messages []string, newLine bool, options uint) {
	s.output.WriteMany(messages, newLine, options)
}

// TODO: implement
func (s *Style) Title(message string) {}

// TODO: implement
func (s *Style) Section(message string) {}

// TODO: implement
func (s *Style) Listing(elements []string, prefix string) {}

// TODO: implement
func (s *Style) Text(messages []string) {}

// TODO: implement
func (s *Style) Success(messages []string) {}

// TODO: implement
func (s *Style) Err(messages []string) {}

// TODO: implement
func (s *Style) Warning(messages []string) {}

// TODO: implement
func (s *Style) Info(messages []string) {}

// TODO: implement
func (s *Style) Note(messages []string) {}

// TODO: implement
func (s *Style) Caution(messages []string) {}

// TODO: implement
func (s *Style) Table(headers []string, rows map[string]string) {}

// TODO: implement
func (s *Style) Ask(questions string, defaultValue string, validator func(string) bool) string {
	return defaultValue
}

// TODO: implement
func (s *Style) AskHidden(question string, validator func(string) bool) string {
	return ""
}

// TODO: implement
func (s *Style) Confirm(q string, defaultValue bool) (bool, error) {
	cq := question.NewConfirmationQuestion(q, defaultValue, nil)
	return askQuestion(s, cq, s.input, s.output)
}

// TODO: implement
func (s *Style) Choice(question string, choices map[string]string, defaultValue string) string {
	return defaultValue
}

// TODO: implement
func (s *Style) NewLine(count int) {}

// TODO: implement
func (s *Style) ProgressStart(max uint) {}

// TODO: implement
func (s *Style) ProgressAdvance(step uint) {}

// TODO: implement
func (s *Style) ProgressFinish() {}

// TODO: implement
func (s *Style) Box(title string, body string, footer string, color string, info string) {}

func askQuestion[T any](s *Style, q types.QuestionInterface[T], i input.InputInterface, o output.OutputInterface) (T, error) {
	if i.IsInteractive() {
		s.autoPrependBlock()
	}

	answer, err := questionHelper.Ask(i, o, q)

	if err != nil {
		var empty T
		return empty, nil
	}

	if i.IsInteractive() {
		consoleSectionOutput, ok := o.(*output.ConsoleSectionOutput)
		if ok {
			// add the new line of the `return` to submit the input to ConsoleSectionOutput, because ConsoleSectionOutput is holding all it's lines.
			// this is relevant when a `ConsoleSectionOutput.clear` is called.
			consoleSectionOutput.AddNewLineOfInputSubmit()
		}
		s.NewLine(1)
		s.bufferedOutput.Write("\n", false, 0)
	}

	return answer, nil
}

func (s *Style) autoPrependBlock() {
	chars := s.bufferedOutput.Fetch()

	if len(chars) > 2 {
		chars = chars[:len(chars)-2]
	}

	if chars == "" {
		s.NewLine(1) // empty history, so we should start with a new line.
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
	s.NewLine(2 - lineBreakCount)
}
