package output

import (
	"os"
	"regexp"

	Formatter "github.com/michielnijenhuis/cli/formatter"
)

type StreamOutput struct {
	stream *os.File
	Output
}

func NewStreamOutput(stream *os.File, verbosity uint, decorated bool, formatter Formatter.OutputFormatterInferface) *StreamOutput {
	so := &StreamOutput{
		stream: stream,
		Output: *NewOutput(verbosity, decorated, formatter),
	}

	so.outputter = func(message string, newLine bool) {
		if newLine {
			message += "\n"
		}
	
		so.stream.WriteString(message)
	}

	return so
}

func (o *StreamOutput) GetStream() *os.File {
	return o.stream
}

func HasColorSupport() bool {
	_, envSet := os.LookupEnv("NO_COLOR")
	if envSet {
		return false
	}

	if os.Getenv("TERM_PROGRAM") == "Hyper" ||
		os.Getenv("COLORTERM") != "false" ||
		os.Getenv("ANSICON") != "false" ||
		os.Getenv("ConEmuANSI") == "ON" {
		return true
	}

	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}

	re := regexp.MustCompile("/^((screen|xterm|vt100|vt220|putty|rxvt|ansi|cygwin|linux).*)|(.*-256(color)?(-bce)?)/")
	return re.MatchString(term)
}
