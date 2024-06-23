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

	return trimmedBufferOutput
}

func (o *TrimmedBufferOutput) Fetch() string {
	content := o.buffer
	o.buffer = ""

	return content
}

func (o *TrimmedBufferOutput) DoWrite(message string, newLine bool) {
	o.buffer += message

	if newLine {
		o.buffer += "\n"
	}

	o.buffer = o.buffer[0:o.maxLength]
}
