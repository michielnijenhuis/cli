package cli

import "strings"

type TrimmedBufferOutput struct {
	buffer    strings.Builder
	maxLength int
	*Output
}

func (o *TrimmedBufferOutput) DoWrite(message string, newLine bool) {
	o.buffer.WriteString(message)

	if newLine {
		o.buffer.WriteString(Eol)
	}

	if o.buffer.Len() >= o.maxLength {
		var newBuffer strings.Builder
		newBuffer.WriteString(o.buffer.String()[0:o.maxLength])
	}
}

func (o *TrimmedBufferOutput) Fetch() string {
	content := o.buffer
	o.buffer = strings.Builder{}

	return content.String()
}
