package output

import (
	"regexp"

	Formatter "github.com/michielnijenhuis/cli/formatter"
)

type Outputter func(message string, newLine bool)

type Output struct {
	verbosity uint
	decorated bool
	formatter Formatter.OutputFormatterInferface
	outputter Outputter
}

func NewOutput(verbosity uint, decorated bool, formatter Formatter.OutputFormatterInferface) *Output {
	if verbosity == 0 {
		verbosity = VERBOSITY_NORMAL
	}

	if formatter == nil {
		formatter = Formatter.NewOutputFormatter(false, nil)
	}

	formatter.SetDecorated(decorated)

	output := &Output{
		verbosity: verbosity,
		decorated: decorated,
		formatter: formatter,
		outputter: nil,
	}

	return output
}

func (o *Output) SetFormatter(formatter Formatter.OutputFormatterInferface) {
	o.formatter = formatter
}

func (o *Output) GetFormatter() Formatter.OutputFormatterInferface {
	return o.formatter
}

func (o *Output) SetDecorated(decorated bool) {
	o.formatter.SetDecorated(decorated)
}

func (o *Output) IsDecorated() bool {
	return o.formatter.IsDecorated()
}

func (o *Output) SetVerbosity(verbosity uint) {
	o.verbosity = verbosity
}

func (o *Output) GetVerbosity() uint {
	return o.verbosity
}

func (o *Output) IsQuiet() bool {
	return o.verbosity == VERBOSITY_QUIET
}

func (o *Output) IsVerbose() bool {
	return o.verbosity == VERBOSITY_VERBOSE
}

func (o *Output) IsVeryVerbose() bool {
	return o.verbosity == VERBOSITY_VERY_VERBOSE
}

func (o *Output) IsDebug() bool {
	return o.verbosity == VERBOSITY_DEBUG
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
	types := OUTPUT_NORMAL | OUTPUT_RAW | OUTPUT_PLAIN

	t := types & options
	if t == 0 {
		t = OUTPUT_NORMAL
	}

	verbosities := VERBOSITY_QUIET | VERBOSITY_NORMAL | VERBOSITY_VERBOSE | VERBOSITY_VERY_VERBOSE | VERBOSITY_DEBUG
	verbosity := verbosities & options
	if verbosity == 0 {
		verbosity = VERBOSITY_NORMAL
	}

	if verbosity > o.GetVerbosity() {
		return
	}

	re := regexp.MustCompile("<[^>]+>")

	var message string
	for _, m := range messages {
		message = m
		switch t {
		case OUTPUT_NORMAL:
			message = o.formatter.Format(message)
		case OUTPUT_PLAIN:
			message = re.ReplaceAllString(o.formatter.Format(message), "")
		}

		o.DoWrite(message, newLine)
	}
}

func (o *Output) DoWrite(message string, newLine bool) {
	if o.outputter == nil {
		panic("Outputter not found")
	}

	o.outputter(message, newLine)
}
