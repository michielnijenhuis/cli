package style

import (
	"github.com/michielnijenhuis/cli/formatter"
	"github.com/michielnijenhuis/cli/input"
	"github.com/michielnijenhuis/cli/output"
	"github.com/michielnijenhuis/cli/terminal"
)

const MAX_LINE_LENGTH = 120

type Style struct {
	progressBar    interface{} // TODO
	lineLength     int
	bufferedOutput *output.TrimmedBufferOutput
	input          input.InputInterface
	output         output.OutputInterface
}

func NewStyle(i input.InputInterface, o output.OutputInterface) *Style {
	s := &Style{
		progressBar:    nil,
		lineLength:     -1,
		bufferedOutput: output.NewTrimmedBufferOutput(2, o.GetVerbosity(), o.IsDecorated(), o.GetFormatter().Clone()),
		input:          i,
		output:         o,
	}

	width, _ := terminal.GetWidth()
	if width == 0 {
		width = MAX_LINE_LENGTH
	}

	s.lineLength = min(width, MAX_LINE_LENGTH)

	return s
}

func (s *Style) SetFormatter(formatter formatter.OutputFormatterInferface) {
	s.output.SetFormatter(formatter)
}

func (s *Style) GetFormatter() formatter.OutputFormatterInferface {
	return s.output.GetFormatter()
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

func (s *Style) GetVerbosity() uint {
	return s.output.GetVerbosity()
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
func (s *Style) Confirm(question string, defaultValue bool) bool {
	return false
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
