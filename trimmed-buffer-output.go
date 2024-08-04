package cli

type TrimmedBufferOutput struct {
	buffer    string
	maxLength uint
	*Output
}

func (o *TrimmedBufferOutput) DoWrite(message string, newLine bool) {
	o.buffer += message

	if newLine {
		o.buffer += "\n"
	}

	if len(o.buffer) >= int(o.maxLength) {
		o.buffer = o.buffer[0:o.maxLength]
	}
}

func (o *TrimmedBufferOutput) Fetch() string {
	content := o.buffer
	o.buffer = ""

	return content
}
