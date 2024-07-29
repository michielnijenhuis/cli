package descriptor

import (
	"github.com/michielnijenhuis/cli/output"
)

func Write(o output.OutputInterface, content string, decorated bool) {
	if decorated {
		o.Write(content, false, output.OUTPUT_NORMAL)
	} else {
		o.Write(content, false, output.OUTPUT_RAW)
	}
}
