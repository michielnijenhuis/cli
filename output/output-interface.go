package output

import (
	"github.com/michielnijenhuis/cli/formatter"
)

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

type OutputInterface interface {
	SetFormatter(formatter formatter.OutputFormatterInferface)
	Formatter() formatter.OutputFormatterInferface
	SetDecorated(decorated bool)
	IsDecorated() bool
	SetVerbosity(level uint)
	Verbosity() uint
	IsQuiet() bool
	IsVerbose() bool
	IsVeryVerbose() bool
	IsDebug() bool
	Writeln(message string, options uint)
	Writelns(messages []string, options uint)
	Write(message string, newLine bool, options uint)
	WriteMany(messages []string, newLine bool, options uint)
}
