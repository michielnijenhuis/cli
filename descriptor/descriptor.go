package descriptor

import (
	"github.com/michielnijenhuis/cli/output"
)

func Write(o output.OutputInterface, content string, decorated bool) {
	if decorated {
		o.Write(content, false, output.OutputNormal)
	} else {
		o.Write(content, false, output.OutputRaw)
	}
}
