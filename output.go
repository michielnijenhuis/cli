package cli

import (
	"os"
	"regexp"
)

const maxLineLength = 120

type Output struct {
	Stream         *os.File
	stderr         *Output
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
		Stream:         stream,
		verbosity:      VerbosityNormal,
		decorated:      HasColorSupport(),
		formatter:      formatter,
		lineLength:     maxLineLength,
		bufferedOutput: &TrimmedBufferOutput{},
		input:          input,
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
	o.stderr = setupNewOutput(input, os.Stderr, f)
	return o
}

func (o *Output) Formatter() *OutputFormatter {
	return o.formatter
}

func (o *Output) IsDecorated() bool {
	return o.formatter.Decorated
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
			message = o.formatter.Format(message)
		case OutputPlain:
			message = re.ReplaceAllString(o.formatter.Format(message), "")
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
	doSetDecorated(o.stderr, decorated)
}

func doSetDecorated(o *Output, decorated bool) {
	if o != nil {
		o.decorated = decorated
	}
}

func (o *Output) SetFormatter(formatter *OutputFormatter) {
	doSetFormatter(o, formatter)
	doSetFormatter(o.stderr, formatter)
}

func doSetFormatter(o *Output, formatter *OutputFormatter) {
	if o != nil {
		o.formatter = formatter
	}
}

func (o *Output) SetVerbosity(verbose uint) {
	doSetVerbosity(o, verbose)
	doSetVerbosity(o.stderr, verbose)
}

func doSetVerbosity(o *Output, verbose uint) {
	if o != nil {
		o.verbosity = verbose
	}
}

func (o *Output) ErrorOutput() *Output {
	return o.stderr
}

func (o *Output) SetErrorOutput(output *Output) {
	o.stderr = output
}

// TODO: implement
func (o *Output) Title(message string) {}

// TODO: implement
func (o *Output) Section(message string) {}

// TODO: implement
func (o *Output) Listing(elements []string, prefix string) {}

// TODO: implement
func (o *Output) Text(messages []string) {}

// TODO: implement
func (o *Output) Success(messages []string) {}

// TODO: implement
func (o *Output) Err(messages []string) {}

// TODO: implement
func (o *Output) Warning(messages []string) {}

// TODO: implement
func (o *Output) Info(messages []string) {}

// TODO: implement
func (o *Output) Note(messages []string) {}

// TODO: implement
func (o *Output) Caution(messages []string) {}

// TODO: implement
func (o *Output) Table(headers []string, rows map[string]string) {}

// TODO: implement
func (o *Output) Ask(questions string, defaultValue string, validator func(string) bool) string {
	return defaultValue
}

// TODO: implement
func (o *Output) AskHidden(question string, validator func(string) bool) string {
	return ""
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

// TODO: implement
func (o *Output) Choice(question string, choices map[string]string, defaultValue string) string {
	return defaultValue
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
		o.bufferedOutput.Write("\n", false, 0)
	}

	return answer, nil
}

func (o *Output) autoPrependBlock() {
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
