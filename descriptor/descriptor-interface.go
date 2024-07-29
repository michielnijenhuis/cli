package descriptor

import (
	"github.com/michielnijenhuis/cli/output"
)

type DescriptorOptions struct {
	namespace  string
	rawText    bool
	short      bool
	totalWidth int
}

type DescriptorInterface interface {
	Describe(output output.OutputInterface, options DescriptorOptions)
}

func NewDescriptorOptions(namespace string, rawText bool, short bool, totalWidth int) *DescriptorOptions {
	return &DescriptorOptions{namespace: namespace, rawText: rawText, short: rawText, totalWidth: totalWidth}
}
