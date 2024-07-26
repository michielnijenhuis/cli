package descriptor

import (
	output "github.com/michielnijenhuis/cli/output"
)

type DescriptorOptions struct {
	format     string
	namespace  string
	rawText    bool
	rawOutput  bool
	short      bool
	totalWidth int
}

type DescriptorInterface interface {
	Describe(output output.OutputInterface, options DescriptorOptions)
}
