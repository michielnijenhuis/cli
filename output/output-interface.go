package output

import (
	"github.com/michielnijenhuis/cli/formatter"
)

const (
	VERBOSITY_QUIET        uint = 16
	VERBOSITY_NORMAL       uint = 32
	VERBOSITY_VERBOSE      uint = 64
	VERBOSITY_VERY_VERBOSE uint = 128
	VERBOSITY_DEBUG        uint = 256
)

const (
	OUTPUT_NORMAL uint = 1
	OUTPUT_RAW    uint = 2
	OUTPUT_PLAIN  uint = 4
)

type OutputInterface interface {
	SetFormatter(formatter formatter.OutputFormatterInferface)
	GetFormatter() formatter.OutputFormatterInferface
	SetDecorated(decorated bool)
	IsDecorated() bool
	SetVerbosity(level uint)
	GetVerbosity() uint
	IsQuiet() bool
	IsVerbose() bool
	IsVeryVerbose() bool
	IsDebug() bool
	Writeln(message string, options uint)
	Writelns(messages []string, options uint)
	Write(message string, newLine bool, options uint)
	WriteMany(messages []string, newLine bool, options uint)
}
