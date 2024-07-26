package descriptor

import (
	Output "github.com/michielnijenhuis/cli/output"
)

func Write(output Output.OutputInterface, content string, decorated bool) {
	if decorated {
		output.Write(content, false, Output.OUTPUT_NORMAL)
	} else {
		output.Write(content, false, Output.OUTPUT_RAW)
	}
}
