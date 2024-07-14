package output

import (
	"github.com/michielnijenhuis/cli/formatter"
)

type TrimmedBufferOutput struct {
	buffer    string
	maxLength uint
	Output
}

func NewTrimmedBufferOutput(maxLength uint, verbosity uint, decorated bool, formatter formatter.OutputFormatterInferface) *TrimmedBufferOutput {
	output := NewOutput(verbosity, decorated, formatter)
	trimmedBufferOutput := &TrimmedBufferOutput{
		buffer:    "",
		maxLength: maxLength,
		Output:    *output,
	}

	trimmedBufferOutput.outputter = func(message string, newLine bool) {
		trimmedBufferOutput.buffer += message

		if newLine {
			trimmedBufferOutput.buffer += "\n"
		}

		trimmedBufferOutput.buffer = trimmedBufferOutput.buffer[0:trimmedBufferOutput.maxLength]
	}

	return trimmedBufferOutput
}

func (o *TrimmedBufferOutput) Fetch() string {
	content := o.buffer
	o.buffer = ""

	return content
}
