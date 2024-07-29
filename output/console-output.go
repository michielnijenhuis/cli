package output

import (
	"os"

	"github.com/michielnijenhuis/cli/formatter"
)

type ConsoleOutput struct {
	stderr                OutputInterface
	consoleSectionOutputs []ConsoleSectionOutput
	StreamOutput
}

func NewConsoleOutput(verbosity uint, decorated bool, formatter formatter.OutputFormatterInferface) *ConsoleOutput {
	streamOutput := NewStreamOutput(OpenOutputStream(), verbosity, decorated, formatter)

	var stderr OutputInterface
	if formatter == nil {
		stderr = NewStreamOutput(OpenErrorStream(), verbosity, decorated, formatter)
	} else {
		stderr = NewStreamOutput(OpenErrorStream(), verbosity, decorated, streamOutput.Formatter())
		streamOutput.SetDecorated(decorated)
		stderr.SetDecorated(decorated)
	}

	return &ConsoleOutput{
		stderr:                stderr,
		consoleSectionOutputs: make([]ConsoleSectionOutput, 0),
		StreamOutput:          *streamOutput,
	}
}

func (o *ConsoleOutput) Section() *ConsoleSectionOutput {
	return nil
}

func (o *ConsoleOutput) SetDecorated(decorated bool) {
	o.StreamOutput.SetDecorated(decorated)
	o.stderr.SetDecorated(decorated)
}

func (o *ConsoleOutput) SetFormatter(formatter formatter.OutputFormatterInferface) {
	o.StreamOutput.SetFormatter(formatter)
	o.stderr.SetFormatter(formatter)
}

func (o *ConsoleOutput) SetVerbosity(verbose uint) {
	o.StreamOutput.SetVerbosity(verbose)
	o.stderr.SetVerbosity(verbose)
}

func (o *ConsoleOutput) ErrorOutput() OutputInterface {
	return o.stderr
}

func (o *ConsoleOutput) SetErrorOutput(output OutputInterface) {
	o.stderr = output
}

func OpenOutputStream() *os.File {
	return os.Stdout
}

func OpenErrorStream() *os.File {
	return os.Stderr
}
