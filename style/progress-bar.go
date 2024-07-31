package style

import (
	"github.com/michielnijenhuis/cli/output"
	terminal "github.com/michielnijenhuis/cli/terminal/cursor"
)

const (
	FORMAT_VERBOSE            = "verbose"
	FORMAT_VERY_VERBOSE       = "very_verbose"
	FORMAT_DEBUG              = "debug"
	FORMAT_NORMAL             = "normal"
	FORMAT_VERBOSE_NOMAX      = "verbose_nomax"
	FORMAT_VERY_VERBOSE_NOMAX = "very_verbose_nomax"
	FORMAT_DEBUG_NOMAX        = "debug_nomax"
	FORMAT_NORMAL_NOMAX       = "normal_nomax"
)

type Formatter func(self *ProgressBar, o output.OutputInterface) string

var formatters map[string]Formatter
var formats map[string]string

type ProgressBar struct {
	barWidth                 int
	barChar                  byte
	emptyBarChar             byte
	progressChar             byte
	format                   string
	internalFormat           string
	redrawFreq               int
	writeCount               int
	lastWriteTime            int
	minSecondsBetweenRedraws int
	maxSecondsBetweenRedraws int
	output                   output.OutputInterface
	step                     int
	startingStep             int
	max                      int
	startTime                int
	stepWidth                int
	percent                  float32
	messages                 map[string]string
	overwrite                bool
	previousMessage          string
	cursor                   terminal.Cursor
	placeholders             map[string]Formatter
}

// TODO: implement
